package repository

import (
	"context"
	"fmt"
	"strings"

	"category-service/internal/shared/query"
)

// Repository executes generic operations against one explicit EntitySpec.
type Repository[T any, ID any] struct {
	binding Binding
	spec    EntitySpec[T, ID]
}

// New creates a generic repository after validating descriptor consistency.
func New[T any, ID any](db DBTX, spec EntitySpec[T, ID]) (Repository[T, ID], error) {
	if db == nil {
		return Repository[T, ID]{}, fmt.Errorf("%w: database is nil", errInvalidSpec)
	}
	if err := spec.validate(); err != nil {
		return Repository[T, ID]{}, err
	}
	return Repository[T, ID]{binding: NewBinding(db), spec: spec}, nil
}

// WithDB returns a copy bound to a pool or transaction without changing the
// original repository.
func (value Repository[T, ID]) WithDB(db DBTX) Repository[T, ID] {
	value.binding = value.binding.WithDB(db)
	return value
}

// Count returns the active rows that match predicate.
func (value Repository[T, ID]) Count(ctx context.Context, predicate query.Expression) (int, error) {
	compiled, err := query.Compile(value.scoped(predicate))
	if err != nil {
		return 0, err
	}
	statement := "SELECT COUNT(*) FROM " + value.spec.Table.SQL()
	if compiled.SQL != "" {
		statement += " WHERE " + compiled.SQL
	}

	var count int
	if err := value.binding.DB().QueryRow(ctx, statement, compiled.Args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Exists reports whether an active row matches predicate.
func (value Repository[T, ID]) Exists(ctx context.Context, predicate query.Expression) (bool, error) {
	compiled, err := query.Compile(value.scoped(predicate))
	if err != nil {
		return false, err
	}
	statement := "SELECT EXISTS(SELECT 1 FROM " + value.spec.Table.SQL()
	if compiled.SQL != "" {
		statement += " WHERE " + compiled.SQL
	}
	statement += ")"

	var exists bool
	if err := value.binding.DB().QueryRow(ctx, statement, compiled.Args...).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// List returns active rows using the registered projection selected by options.
func (value Repository[T, ID]) List(ctx context.Context, predicate query.Expression, options query.Options) ([]T, error) {
	compiled, err := query.CompileList(value.spec.Table, value.scoped(predicate), value.spec.Query, options)
	if err != nil {
		return nil, err
	}
	projection, found := value.spec.Projections[compiled.Projection.Key]
	if !found {
		return nil, errInvalidSpec
	}

	rows, err := value.binding.DB().Query(ctx, compiled.SQL, compiled.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []T
	for rows.Next() {
		item, err := projection.Scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// Get finds one active row by its key. found is false when no active row exists.
func (value Repository[T, ID]) Get(ctx context.Context, id ID) (item T, found bool, err error) {
	items, err := value.List(ctx, query.Compare(value.spec.Key, query.Equal, id), query.Options{Limit: 1})
	if err != nil {
		return item, false, err
	}
	if len(items) == 0 {
		return item, false, nil
	}
	return items[0], true, nil
}

// Create inserts entity values using descriptor-owned columns.
func (value Repository[T, ID]) Create(ctx context.Context, item T) (CommandResult, error) {
	values, err := value.writeValues(value.spec.InsertColumns, value.spec.InsertValues, item)
	if err != nil {
		return CommandResult{}, err
	}
	columns := sqlColumns(value.spec.InsertColumns)
	statement := "INSERT INTO " + value.spec.Table.SQL() + " (" + strings.Join(columns, ", ") + ") VALUES (" + placeholders(1, len(values)) + ")"
	tag, err := value.binding.DB().Exec(ctx, statement, values...)
	if err != nil {
		return CommandResult{}, err
	}
	return commandResult(tag), nil
}

// Update updates an active row by id and reports the affected-row count.
func (value Repository[T, ID]) Update(ctx context.Context, id ID, item T) (CommandResult, error) {
	values, err := value.writeValues(value.spec.UpdateColumns, value.spec.UpdateValues, item)
	if err != nil {
		return CommandResult{}, err
	}
	assignments := make([]string, len(value.spec.UpdateColumns))
	for index, column := range value.spec.UpdateColumns {
		assignments[index] = column.SQL() + " = $" + integerString(index+1)
	}
	predicate := value.scoped(query.Compare(value.spec.Key, query.Equal, id))
	compiled, err := query.CompileWithOffset(predicate, len(values))
	if err != nil {
		return CommandResult{}, err
	}
	values = append(values, compiled.Args...)
	statement := "UPDATE " + value.spec.Table.SQL() + " SET " + strings.Join(assignments, ", ") + " WHERE " + compiled.SQL
	tag, err := value.binding.DB().Exec(ctx, statement, values...)
	if err != nil {
		return CommandResult{}, err
	}
	return commandResult(tag), nil
}

// Delete deletes an active row by id. SoftDelete requires explicit soft-delete
// metadata; HardDelete always performs a physical delete.
func (value Repository[T, ID]) Delete(ctx context.Context, id ID, mode DeleteMode) (CommandResult, error) {
	predicate := value.scoped(query.Compare(value.spec.Key, query.Equal, id))
	if mode == HardDelete {
		compiled, err := query.Compile(predicate)
		if err != nil {
			return CommandResult{}, err
		}
		statement := "DELETE FROM " + value.spec.Table.SQL() + " WHERE " + compiled.SQL
		tag, err := value.binding.DB().Exec(ctx, statement, compiled.Args...)
		if err != nil {
			return CommandResult{}, err
		}
		return commandResult(tag), nil
	}
	if mode != SoftDelete || value.spec.SoftDelete == nil {
		return CommandResult{}, fmt.Errorf("%w: soft-delete metadata is required", errInvalidSpec)
	}
	compiled, err := query.Compile(predicate)
	if err != nil {
		return CommandResult{}, err
	}
	assignments := make([]string, 0, len(value.spec.SoftDelete.TouchColumns)+1)
	assignments = append(assignments, value.spec.SoftDelete.Column.SQL()+" = CURRENT_TIMESTAMP")
	for _, column := range value.spec.SoftDelete.TouchColumns {
		assignments = append(assignments, column.SQL()+" = CURRENT_TIMESTAMP")
	}
	statement := "UPDATE " + value.spec.Table.SQL() + " SET " + strings.Join(assignments, ", ") + " WHERE " + compiled.SQL
	tag, err := value.binding.DB().Exec(ctx, statement, compiled.Args...)
	if err != nil {
		return CommandResult{}, err
	}
	return commandResult(tag), nil
}

func (value Repository[T, ID]) scoped(predicate query.Expression) query.Expression {
	if value.spec.SoftDelete == nil {
		return predicate
	}
	return query.And(predicate, query.IsNull(value.spec.SoftDelete.Column))
}

func (value Repository[T, ID]) writeValues(columns []query.Column, extract func(T) []any, item T) ([]any, error) {
	if len(columns) == 0 || extract == nil {
		return nil, fmt.Errorf("%w: write metadata is missing", errInvalidSpec)
	}
	values := extract(item)
	if len(values) != len(columns) {
		return nil, fmt.Errorf("%w: value count does not match column count", errInvalidSpec)
	}
	return values, nil
}

func sqlColumns(columns []query.Column) []string {
	values := make([]string, len(columns))
	for index, column := range columns {
		values[index] = column.SQL()
	}
	return values
}

func placeholders(start, count int) string {
	values := make([]string, count)
	for index := range values {
		values[index] = "$" + integerString(start+index)
	}
	return strings.Join(values, ", ")
}

func integerString(value int) string {
	return fmt.Sprintf("%d", value)
}
