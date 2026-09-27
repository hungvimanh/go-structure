package repository

import (
	"errors"

	"category-service/internal/shared/query"
)

var errInvalidSpec = errors.New("invalid repository entity specification")

// Projection binds a registered query projection to the exact scanner required
// for its column sequence.
type Projection[T any] struct {
	Scan func(RowScanner) (T, error)
}

// SoftDeleteSpec defines an explicit timestamp column for logical deletion.
type SoftDeleteSpec struct {
	Column       query.Column
	TouchColumns []query.Column
}

// EntitySpec is static entity metadata. Modules must provide identifiers,
// projections/scanners, write extractors, and any domain-specific predicates.
type EntitySpec[T any, ID any] struct {
	Table        query.Table
	Key          query.Column
	Query        query.Registry
	Projections  map[query.ProjectionKey]Projection[T]
	InsertColumns []query.Column
	InsertValues  func(T) []any
	UpdateColumns []query.Column
	UpdateValues  func(T) []any
	SoftDelete    *SoftDeleteSpec
}

func (value EntitySpec[T, ID]) validate() error {
	if value.Table.SQL() == "" || value.Key.SQL() == "" {
		return errInvalidSpec
	}
	if len(value.Query.Projections) == 0 || len(value.Projections) == 0 || len(value.Query.DefaultSort) == 0 {
		return errInvalidSpec
	}
	if _, found := value.Query.Projections[value.Query.DefaultProjection]; !found {
		return errInvalidSpec
	}

	for key, projection := range value.Query.Projections {
		if key == (query.ProjectionKey{}) || !validColumns(projection.Columns) {
			return errInvalidSpec
		}
		mapper, found := value.Projections[key]
		if !found || mapper.Scan == nil {
			return errInvalidSpec
		}
	}
	for key, mapper := range value.Projections {
		if key == (query.ProjectionKey{}) || mapper.Scan == nil {
			return errInvalidSpec
		}
		if _, found := value.Query.Projections[key]; !found {
			return errInvalidSpec
		}
	}
	for key, column := range value.Query.SearchFields {
		if key == (query.SearchKey{}) || column.SQL() == "" {
			return errInvalidSpec
		}
	}
	for key, column := range value.Query.SortFields {
		if key == (query.SortKey{}) || column.SQL() == "" {
			return errInvalidSpec
		}
	}
	for _, term := range value.Query.DefaultSort {
		if !validSortTerm(value.Query, term) {
			return errInvalidSpec
		}
	}
	if !validSortTerm(value.Query, value.Query.TieBreaker) {
		return errInvalidSpec
	}

	if !validWriteMetadata(value.InsertColumns, value.InsertValues) || !validWriteMetadata(value.UpdateColumns, value.UpdateValues) {
		return errInvalidSpec
	}
	if value.SoftDelete != nil {
		if value.SoftDelete.Column.SQL() == "" || (len(value.SoftDelete.TouchColumns) > 0 && !validColumns(value.SoftDelete.TouchColumns)) {
			return errInvalidSpec
		}
	}
	return nil
}

func validColumns(columns []query.Column) bool {
	if len(columns) == 0 {
		return false
	}
	for _, column := range columns {
		if column.SQL() == "" {
			return false
		}
	}
	return true
}

func validWriteMetadata[T any](columns []query.Column, values func(T) []any) bool {
	if len(columns) == 0 {
		return values == nil
	}
	return values != nil && validColumns(columns)
}

func validSortTerm(registry query.Registry, term query.SortTerm) bool {
	if term.Key == (query.SortKey{}) || (term.Direction != query.Ascending && term.Direction != query.Descending) {
		return false
	}
	_, found := registry.SortFields[term.Key]
	return found
}

// DeleteMode selects an explicit hard or soft delete operation.
type DeleteMode uint8

const (
	SoftDelete DeleteMode = iota
	HardDelete
)
