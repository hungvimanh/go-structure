package query

import "strings"

// Expression is a closed, descriptor-owned predicate tree.
type Expression interface {
	compile(*compiler) (string, error)
	measure(*metrics, int) error
}

// ComparisonOperator is a SQL comparison represented without caller-provided SQL.
type ComparisonOperator uint8

const (
	Equal ComparisonOperator = iota
	NotEqual
	GreaterThan
	GreaterThanOrEqual
	LessThan
	LessThanOrEqual
)

// StringOperator describes a case-insensitive literal string predicate.
type StringOperator uint8

const (
	StringEqual StringOperator = iota
	StringNotEqual
	StringContains
	StringNotContains
	StringReverseContains
	StringReverseNotContains
	StringCombineContains
	StringStartsWith
	StringNotStartsWith
	StringReverseStartsWith
	StringReverseNotStartsWith
	StringCombineStartsWith
	StringEndsWith
	StringNotEndsWith
	StringReverseEndsWith
	StringReverseNotEndsWith
	StringCombineEndsWith
)

type comparisonExpression struct {
	column   Column
	operator ComparisonOperator
	value    any
}

type stringExpression struct {
	column   Column
	operator StringOperator
	value    string
}

type setExpression struct {
	column   Column
	values   any
	rawCount int
	not      bool
	fold     bool
	empty    bool
}

type nullExpression struct {
	column Column
	not    bool
}

type groupExpression struct {
	operator string
	values   []Expression
}

type notExpression struct{ value Expression }

// Compare creates a parameterized scalar comparison.
func Compare(column Column, operator ComparisonOperator, value any) Expression {
	return comparisonExpression{column: column, operator: operator, value: value}
}

// String creates a parameterized case-insensitive literal string predicate.
// Whitespace-only values are a no-op.
func String(column Column, operator StringOperator, value string) Expression {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return stringExpression{column: column, operator: operator, value: value}
}

// In creates a PostgreSQL array-membership predicate. A nil slice is omitted;
// a non-nil empty slice is false. The supplied values are deduplicated before binding.
func In[T comparable](column Column, values []T) Expression {
	return newSetExpression(column, values, false, false)
}

// NotIn creates a PostgreSQL array non-membership predicate. A nil or non-nil
// empty slice is a no-op. The supplied values are deduplicated before binding.
func NotIn[T comparable](column Column, values []T) Expression {
	return newSetExpression(column, values, true, false)
}

// StringIn creates a case-insensitive string set predicate.
func StringIn(column Column, values []string) Expression {
	return newSetExpression(column, normalizeStrings(values), false, true)
}

// StringNotIn creates a case-insensitive string non-membership predicate.
func StringNotIn(column Column, values []string) Expression {
	return newSetExpression(column, normalizeStrings(values), true, true)
}

func newSetExpression[T comparable](column Column, values []T, not, fold bool) Expression {
	if values == nil {
		return nil
	}
	if len(values) == 0 {
		return setExpression{column: column, not: not, empty: true}
	}

	seen := make(map[T]struct{}, len(values))
	deduplicated := make([]T, 0, len(values))
	for _, value := range values {
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		deduplicated = append(deduplicated, value)
	}
	return setExpression{column: column, values: deduplicated, rawCount: len(values), not: not, fold: fold}
}

// IsNull creates an IS NULL predicate.
func IsNull(column Column) Expression {
	return nullExpression{column: column}
}

// IsNotNull creates an IS NOT NULL predicate.
func IsNotNull(column Column) Expression {
	return nullExpression{column: column, not: true}
}

// And combines non-nil predicates with AND. No predicates is a no-op.
func And(values ...Expression) Expression {
	return group("AND", values)
}

// Or combines non-nil predicates with OR. No predicates is a no-op.
func Or(values ...Expression) Expression {
	return group("OR", values)
}

func group(operator string, values []Expression) Expression {
	filtered := make([]Expression, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}
		if child, ok := value.(groupExpression); ok && child.operator == operator {
			filtered = append(filtered, child.values...)
			continue
		}
		filtered = append(filtered, value)
	}
	if len(filtered) == 0 {
		return nil
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return groupExpression{operator: operator, values: filtered}
}

// Not negates a predicate. A nil predicate remains a no-op.
func Not(value Expression) Expression {
	if value == nil {
		return nil
	}
	return notExpression{value: value}
}

func normalizeStrings(values []string) []string {
	if values == nil {
		return nil
	}
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = strings.ToLower(value)
	}
	return result
}

func (value comparisonExpression) compile(compiler *compiler) (string, error) {
	operator, ok := map[ComparisonOperator]string{
		Equal:              "=",
		NotEqual:           "<>",
		GreaterThan:        ">",
		GreaterThanOrEqual: ">=",
		LessThan:           "<",
		LessThanOrEqual:    "<=",
	}[value.operator]
	if !ok {
		return "", newError(InvalidIdentifier, "unknown comparison operator")
	}
	return value.column.sql() + " " + operator + " " + compiler.bind(value.value), nil
}

func (value comparisonExpression) measure(metrics *metrics, depth int) error {
	return metrics.addPredicate(depth)
}

func (value stringExpression) compile(compiler *compiler) (string, error) {
	placeholder := compiler.bind(value.value)
	column := value.column.sql()
	positive, negative, err := compileStringPredicate(column, placeholder, value.operator)
	if err != nil {
		return "", err
	}
	if negative {
		return "(" + column + " IS NOT NULL AND NOT (" + positive + "))", nil
	}
	return positive, nil
}

func (value stringExpression) measure(metrics *metrics, depth int) error {
	return metrics.addPredicate(depth)
}

func (value setExpression) compile(compiler *compiler) (string, error) {
	if value.empty {
		if value.not {
			return "", nil
		}
		return "FALSE", nil
	}
	column := value.column.sql()
	if value.fold {
		column = "LOWER(" + column + ")"
	}
	if value.not {
		return column + " <> ALL(" + compiler.bind(value.values) + ")", nil
	}
	return column + " = ANY(" + compiler.bind(value.values) + ")", nil
}

func (value setExpression) measure(metrics *metrics, depth int) error {
	if value.empty {
		return metrics.addPredicate(depth)
	}
	return metrics.addSet(depth, value.rawCount)
}

func (value nullExpression) compile(_ *compiler) (string, error) {
	if value.not {
		return value.column.sql() + " IS NOT NULL", nil
	}
	return value.column.sql() + " IS NULL", nil
}

func (value nullExpression) measure(metrics *metrics, depth int) error {
	return metrics.addPredicate(depth)
}

func (value groupExpression) compile(compiler *compiler) (string, error) {
	parts := make([]string, 0, len(value.values))
	for _, child := range value.values {
		part, err := child.compile(compiler)
		if err != nil {
			return "", err
		}
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	return "(" + strings.Join(parts, " "+value.operator+" ") + ")", nil
}

func (value groupExpression) measure(metrics *metrics, depth int) error {
	for _, child := range value.values {
		if err := child.measure(metrics, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func (value notExpression) compile(compiler *compiler) (string, error) {
	part, err := value.value.compile(compiler)
	if err != nil || part == "" {
		return part, err
	}
	return "NOT (" + part + ")", nil
}

func (value notExpression) measure(metrics *metrics, depth int) error {
	return value.value.measure(metrics, depth+1)
}

func compileStringPredicate(column, placeholder string, operator StringOperator) (string, bool, error) {
	normalizedColumn := "LOWER(" + column + ")"
	normalizedValue := "LOWER(" + placeholder + ")"
	normal := func(mode string) string {
		switch mode {
		case "equal":
			return normalizedColumn + " = " + normalizedValue
		case "contains":
			return "STRPOS(" + normalizedColumn + ", " + normalizedValue + ") > 0"
		case "starts":
			return "LEFT(" + normalizedColumn + ", LENGTH(" + normalizedValue + ")) = " + normalizedValue
		default:
			return "RIGHT(" + normalizedColumn + ", LENGTH(" + normalizedValue + ")) = " + normalizedValue
		}
	}
	reverse := func(mode string) string {
		switch mode {
		case "contains":
			return "STRPOS(" + normalizedValue + ", " + normalizedColumn + ") > 0"
		case "starts":
			return "LEFT(" + normalizedValue + ", LENGTH(" + normalizedColumn + ")) = " + normalizedColumn
		default:
			return "RIGHT(" + normalizedValue + ", LENGTH(" + normalizedColumn + ")) = " + normalizedColumn
		}
	}

	switch operator {
	case StringEqual:
		return normal("equal"), false, nil
	case StringNotEqual:
		return normal("equal"), true, nil
	case StringContains:
		return normal("contains"), false, nil
	case StringNotContains:
		return normal("contains"), true, nil
	case StringReverseContains:
		return reverse("contains"), false, nil
	case StringReverseNotContains:
		return reverse("contains"), true, nil
	case StringCombineContains:
		return "(" + normal("contains") + " OR " + reverse("contains") + ")", false, nil
	case StringStartsWith:
		return normal("starts"), false, nil
	case StringNotStartsWith:
		return normal("starts"), true, nil
	case StringReverseStartsWith:
		return reverse("starts"), false, nil
	case StringReverseNotStartsWith:
		return reverse("starts"), true, nil
	case StringCombineStartsWith:
		return "(" + normal("starts") + " OR " + reverse("starts") + ")", false, nil
	case StringEndsWith:
		return normal("ends"), false, nil
	case StringNotEndsWith:
		return normal("ends"), true, nil
	case StringReverseEndsWith:
		return reverse("ends"), false, nil
	case StringReverseNotEndsWith:
		return reverse("ends"), true, nil
	case StringCombineEndsWith:
		return "(" + normal("ends") + " OR " + reverse("ends") + ")", false, nil
	default:
		return "", false, newError(InvalidIdentifier, "unknown string operator")
	}
}
