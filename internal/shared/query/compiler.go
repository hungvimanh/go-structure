package query

import "fmt"

const (
	// DefaultMaxDepth is the maximum nested boolean-expression depth.
	DefaultMaxDepth = 4
	// DefaultMaxPredicates is the maximum predicates accepted per query.
	DefaultMaxPredicates = 64
	// DefaultMaxSetValues is the maximum raw set values accepted across one expression tree.
	DefaultMaxSetValues = 1000
)

// Limits bounds expression complexity before the database is contacted. A zero
// field uses the corresponding package default.
type Limits struct {
	MaxDepth      int
	MaxPredicates int
	MaxSetValues  int
}

// Compiled contains parameterized SQL and its ordered PostgreSQL arguments.
type Compiled struct {
	SQL  string
	Args []any
}

type compiler struct {
	args   []any
	offset int
}

type metrics struct {
	limits    Limits
	predicates int
	setValues int
}

// Compile compiles an expression without a WHERE prefix. A nil expression is a
// no-op and produces an empty SQL string and no arguments.
func Compile(expression Expression) (Compiled, error) {
	return CompileWithLimits(expression, Limits{})
}

// CompileWithLimits compiles an expression after validating its complexity.
func CompileWithLimits(expression Expression, limits Limits) (Compiled, error) {
	return CompileWithOffsetAndLimits(expression, 0, limits)
}

// CompileWithOffset compiles an expression whose arguments follow offset
// arguments already present in the enclosing statement.
func CompileWithOffset(expression Expression, offset int) (Compiled, error) {
	return CompileWithOffsetAndLimits(expression, offset, Limits{})
}

// CompileWithOffsetAndLimits compiles an expression with an initial placeholder
// offset after validating the expression complexity.
func CompileWithOffsetAndLimits(expression Expression, offset int, limits Limits) (Compiled, error) {
	if offset < 0 {
		return Compiled{}, newError(InvalidPagination, "placeholder offset must not be negative")
	}
	if expression == nil {
		return Compiled{}, nil
	}
	metrics := metrics{limits: normalizeLimits(limits)}
	if err := expression.measure(&metrics, 1); err != nil {
		return Compiled{}, err
	}
	compiler := compiler{offset: offset}
	sql, err := expression.compile(&compiler)
	if err != nil {
		return Compiled{}, err
	}
	return Compiled{SQL: sql, Args: compiler.args}, nil
}

func (value *compiler) bind(argument any) string {
	value.args = append(value.args, argument)
	return fmt.Sprintf("$%d", value.offset+len(value.args))
}

func (value *metrics) addPredicate(depth int) error {
	if depth > value.limits.MaxDepth {
		return newError(ComplexityExceeded, "maximum expression depth exceeded")
	}
	value.predicates++
	if value.predicates > value.limits.MaxPredicates {
		return newError(ComplexityExceeded, "maximum predicate count exceeded")
	}
	return nil
}

func (value *metrics) addSet(depth, count int) error {
	if err := value.addPredicate(depth); err != nil {
		return err
	}
	value.setValues += count
	if value.setValues > value.limits.MaxSetValues {
		return newError(ComplexityExceeded, "maximum total set value count exceeded")
	}
	return nil
}

func normalizeLimits(value Limits) Limits {
	if value.MaxDepth == 0 {
		value.MaxDepth = DefaultMaxDepth
	}
	if value.MaxPredicates == 0 {
		value.MaxPredicates = DefaultMaxPredicates
	}
	if value.MaxSetValues == 0 {
		value.MaxSetValues = DefaultMaxSetValues
	}
	return value
}
