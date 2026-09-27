package repository

import (
	"category-service/internal/category/model"
	"category-service/internal/shared/filter"
	"category-service/internal/shared/query"
)

func categoryFilterExpression(value model.CategoryFilter) query.Expression {
	return query.And(
		stringFilterExpression(categoryCode, value.Code),
		stringFilterExpression(categoryName, value.Name),
		intFilterExpression(categoryStatus, value.Status),
	)
}

func categoryListOptions(skip, take int, search *string) query.Options {
	options := query.Options{
		Offset: skip,
		Limit:  take,
	}
	if search != nil {
		options.Search = *search
	}
	return options
}

func stringFilterExpression(column query.Column, value *filter.StringFilter) query.Expression {
	if value == nil {
		return nil
	}
	return query.And(
		stringPredicate(column, query.StringEqual, value.Eq),
		stringPredicate(column, query.StringNotEqual, value.Neq),
		stringPredicate(column, query.StringContains, value.Contains),
		stringPredicate(column, query.StringNotContains, value.NotContains),
		stringPredicate(column, query.StringReverseContains, value.ReverseContains),
		stringPredicate(column, query.StringReverseNotContains, value.ReverseNotContains),
		stringPredicate(column, query.StringCombineContains, value.CombineContains),
		stringPredicate(column, query.StringStartsWith, value.StartsWith),
		stringPredicate(column, query.StringNotStartsWith, value.NotStartsWith),
		stringPredicate(column, query.StringReverseStartsWith, value.ReverseStartsWith),
		stringPredicate(column, query.StringReverseNotStartsWith, value.ReverseNotStartsWith),
		stringPredicate(column, query.StringCombineStartsWith, value.CombineStartsWith),
		stringPredicate(column, query.StringEndsWith, value.EndsWith),
		stringPredicate(column, query.StringNotEndsWith, value.NotEndsWith),
		stringPredicate(column, query.StringReverseEndsWith, value.ReverseEndsWith),
		stringPredicate(column, query.StringReverseNotEndsWith, value.ReverseNotEndsWith),
		stringPredicate(column, query.StringCombineEndsWith, value.CombineEndsWith),
		query.StringIn(column, value.In),
		query.StringNotIn(column, value.NotIn),
	)
}

func stringPredicate(column query.Column, operator query.StringOperator, value *string) query.Expression {
	if value == nil {
		return nil
	}
	return query.String(column, operator, *value)
}

func intFilterExpression(column query.Column, value *filter.IntFilter) query.Expression {
	if value == nil {
		return nil
	}
	return query.And(
		intPredicate(column, query.Equal, value.Eq),
		intPredicate(column, query.NotEqual, value.Neq),
		intPredicate(column, query.GreaterThan, value.Gt),
		intPredicate(column, query.GreaterThanOrEqual, value.Gte),
		intPredicate(column, query.LessThan, value.Lt),
		intPredicate(column, query.LessThanOrEqual, value.Lte),
		query.In(column, value.In),
		query.NotIn(column, value.NotIn),
	)
}

func intPredicate(column query.Column, operator query.ComparisonOperator, value *int) query.Expression {
	if value == nil {
		return nil
	}
	return query.Compare(column, operator, *value)
}
