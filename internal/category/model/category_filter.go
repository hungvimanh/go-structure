package model

import "category-service/internal/shared/filter"

type CategoryFilter struct {
	Code   *filter.StringFilter `json:"code"`
	Name   *filter.StringFilter `json:"name"`
	Status *filter.IntFilter    `json:"status"`

	Search *string `json:"-"`
}

type CategoryListRequest struct {
	Skip   *int           `json:"skip"`
	Take   *int           `json:"take"`
	Filter CategoryFilter `json:"filter"`
}
