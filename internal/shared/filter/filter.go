// Package filter defines transport-neutral typed filter values.
//
// A nil filter is a no-op. Empty or whitespace-only scalar string values are
// also no-ops. A nil In or NotIn slice is omitted and produces no predicate;
// a non-nil empty In slice represents an explicit false predicate, while a
// non-nil empty NotIn slice is a no-op. Multiple active filter fields are composed
// with AND, except Combine string operators, which form an internal OR group.
// SQL compilation and case-insensitive string behavior belong to the query
// package so this package remains independent of persistence concerns.
package filter

import (
	"time"

	"github.com/google/uuid"
)

// StringFilter describes case-insensitive string predicates. Empty or
// whitespace-only scalar values are no-ops when the filter is compiled.
type StringFilter struct {
	Eq                   *string  `json:"eq"`
	Neq                  *string  `json:"neq"`
	Contains             *string  `json:"contains"`
	NotContains          *string  `json:"notContains"`
	ReverseContains      *string  `json:"reverseContains"`
	ReverseNotContains   *string  `json:"reverseNotContains"`
	CombineContains      *string  `json:"combineContains"`
	StartsWith           *string  `json:"startsWith"`
	NotStartsWith        *string  `json:"notStartsWith"`
	ReverseStartsWith    *string  `json:"reverseStartsWith"`
	ReverseNotStartsWith *string  `json:"reverseNotStartsWith"`
	CombineStartsWith    *string  `json:"combineStartsWith"`
	EndsWith             *string  `json:"endsWith"`
	NotEndsWith          *string  `json:"notEndsWith"`
	ReverseEndsWith      *string  `json:"reverseEndsWith"`
	ReverseNotEndsWith   *string  `json:"reverseNotEndsWith"`
	CombineEndsWith      *string  `json:"combineEndsWith"`
	In                   []string `json:"in"`
	NotIn                []string `json:"notIn"`
}

// ScalarFilter describes equality and set predicates for a scalar value.
type ScalarFilter[T any] struct {
	Eq    *T  `json:"eq"`
	Neq   *T  `json:"neq"`
	In    []T `json:"in"`
	NotIn []T `json:"notIn"`
}

// OrderedFilter describes equality, set, and ordered range predicates for a
// scalar value with an order defined by its module and database mapping.
type OrderedFilter[T any] struct {
	Eq    *T  `json:"eq"`
	Neq   *T  `json:"neq"`
	Gt    *T  `json:"gt"`
	Gte   *T  `json:"gte"`
	Lt    *T  `json:"lt"`
	Lte   *T  `json:"lte"`
	In    []T `json:"in"`
	NotIn []T `json:"notIn"`
}

// IntFilter is retained for source compatibility with existing modules.
type IntFilter struct {
	Eq    *int  `json:"eq"`
	Neq   *int  `json:"neq"`
	Gt    *int  `json:"gt"`
	Gte   *int  `json:"gte"`
	Lt    *int  `json:"lt"`
	Lte   *int  `json:"lte"`
	In    []int `json:"in"`
	NotIn []int `json:"notIn"`
}

// Int64Filter is the ordered filter shape for int64 values.
type Int64Filter = OrderedFilter[int64]

// TimeFilter is the ordered filter shape for time values.
type TimeFilter = OrderedFilter[time.Time]

// UUIDFilter is the scalar filter shape for UUID values.
type UUIDFilter = ScalarFilter[uuid.UUID]

// BoolFilter is the scalar filter shape for bool-compatible values.
type BoolFilter = ScalarFilter[bool]

// DecimalFilter is the ordered filter shape for a module-defined
// decimal-compatible value type.
type DecimalFilter[T any] struct {
	Eq    *T  `json:"eq"`
	Neq   *T  `json:"neq"`
	Gt    *T  `json:"gt"`
	Gte   *T  `json:"gte"`
	Lt    *T  `json:"lt"`
	Lte   *T  `json:"lte"`
	In    []T `json:"in"`
	NotIn []T `json:"notIn"`
}
