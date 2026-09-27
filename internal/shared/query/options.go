package query

import (
	"sort"
	"strconv"
	"strings"
)

// SortDirection is a closed PostgreSQL sort direction.
type SortDirection uint8

const (
	Ascending SortDirection = iota
	Descending
)

// SortTerm selects a configured sortable field and its direction.
type SortTerm struct {
	Key       SortKey
	Direction SortDirection
}

// Projection defines a registered selection of descriptor-owned columns.
type Projection struct {
	Key     ProjectionKey
	Columns []Column
}

// Registry is static module metadata for query options. All identifiers and
// keys must be defined by the module, never derived from a request.
type Registry struct {
	SearchFields      map[SearchKey]Column
	SortFields        map[SortKey]Column
	Projections       map[ProjectionKey]Projection
	DefaultProjection ProjectionKey
	DefaultSort       []SortTerm
	TieBreaker        SortTerm
}

// Options captures transport-neutral search, sort, projection, and offset
// pagination request values.
type Options struct {
	Search       string
	SearchFields []SearchKey
	Sort         []SortTerm
	Projection   *ProjectionKey
	Offset       int
	Limit        int
}

// ListQuery is the fully compiled list statement.
type ListQuery struct {
	SQL        string
	Args       []any
	Projection Projection
}

// CompileList compiles a SELECT statement from descriptor-owned metadata.
func CompileList(table Table, predicate Expression, registry Registry, options Options) (ListQuery, error) {
	if options.Offset < 0 || options.Limit < 0 {
		return ListQuery{}, newError(InvalidPagination, "offset and limit must not be negative")
	}

	projection, err := registry.resolveProjection(options.Projection)
	if err != nil {
		return ListQuery{}, err
	}
	if len(projection.Columns) == 0 {
		return ListQuery{}, newError(InvalidProjectionKey, "projection has no columns")
	}

	search, err := registry.searchExpression(options.Search, options.SearchFields)
	if err != nil {
		return ListQuery{}, err
	}
	compiled, err := Compile(And(predicate, search))
	if err != nil {
		return ListQuery{}, err
	}

	order, err := registry.resolveSort(options.Sort)
	if err != nil {
		return ListQuery{}, err
	}

	columns := make([]string, len(projection.Columns))
	for index, column := range projection.Columns {
		columns[index] = column.sql()
	}

	builder := strings.Builder{}
	builder.WriteString("SELECT ")
	builder.WriteString(strings.Join(columns, ", "))
	builder.WriteString(" FROM ")
	builder.WriteString(table.sql())
	if compiled.SQL != "" {
		builder.WriteString(" WHERE ")
		builder.WriteString(compiled.SQL)
	}
	builder.WriteString(" ORDER BY ")
	builder.WriteString(order)

	args := append([]any(nil), compiled.Args...)
	args = append(args, options.Limit)
	builder.WriteString(" LIMIT $")
	builder.WriteString(integerString(len(args)))
	if options.Offset > 0 {
		args = append(args, options.Offset)
		builder.WriteString(" OFFSET $")
		builder.WriteString(integerString(len(args)))
	}

	return ListQuery{SQL: builder.String(), Args: args, Projection: projection}, nil
}

func (value Registry) resolveProjection(requested *ProjectionKey) (Projection, error) {
	key := value.DefaultProjection
	if requested != nil {
		key = *requested
	}
	projection, found := value.Projections[key]
	if !found {
		return Projection{}, newError(InvalidProjectionKey, "unknown projection key")
	}
	projection.Key = key
	return projection, nil
}

func (value Registry) searchExpression(term string, selected []SearchKey) (Expression, error) {
	if strings.TrimSpace(term) == "" {
		return nil, nil
	}
	fields := selected
	if len(fields) == 0 {
		fields = make([]SearchKey, 0, len(value.SearchFields))
		for key := range value.SearchFields {
			fields = append(fields, key)
		}
		sort.Slice(fields, func(left, right int) bool {
			return fields[left].value < fields[right].value
		})
	}
	predicates := make([]Expression, 0, len(fields))
	for _, key := range fields {
		column, found := value.SearchFields[key]
		if !found {
			return nil, newError(InvalidSearchKey, "unknown search key")
		}
		predicates = append(predicates, String(column, StringContains, term))
	}
	return Or(predicates...), nil
}

func (value Registry) resolveSort(requested []SortTerm) (string, error) {
	terms := requested
	if len(terms) == 0 {
		terms = value.DefaultSort
	}
	if len(terms) == 0 {
		return "", newError(InvalidSortKey, "no default sort configured")
	}

	ordered := make([]string, 0, len(terms)+1)
	tieBreakerIncluded := false
	for _, term := range terms {
		column, found := value.SortFields[term.Key]
		if !found {
			return "", newError(InvalidSortKey, "unknown sort key")
		}
		direction, err := term.Direction.sql()
		if err != nil {
			return "", err
		}
		ordered = append(ordered, column.sql()+" "+direction)
		if term.Key == value.TieBreaker.Key {
			tieBreakerIncluded = true
		}
	}
	if !tieBreakerIncluded {
		column, found := value.SortFields[value.TieBreaker.Key]
		if !found {
			return "", newError(InvalidSortKey, "tie-breaker is not configured")
		}
		direction, err := value.TieBreaker.Direction.sql()
		if err != nil {
			return "", err
		}
		ordered = append(ordered, column.sql()+" "+direction)
	}
	return strings.Join(ordered, ", "), nil
}

func (value SortDirection) sql() (string, error) {
	switch value {
	case Ascending:
		return "ASC", nil
	case Descending:
		return "DESC", nil
	default:
		return "", newError(InvalidDirection, "unknown sort direction")
	}
}

func integerString(value int) string {
	return strconv.Itoa(value)
}
