package repository

import (
	"category-service/internal/category/model"
	"category-service/internal/shared/query"
	sharedrepository "category-service/internal/shared/repository"

	"github.com/google/uuid"
)

var (
	categoryTable       = query.MustTable("category")
	categoryID          = query.MustColumn("id")
	categoryCode        = query.MustColumn("code")
	categoryName        = query.MustColumn("name")
	categoryDescription = query.MustColumn("description")
	categoryStatus      = query.MustColumn("status")
	categoryCreatedAt   = query.MustColumn("created_at")
	categoryUpdatedAt   = query.MustColumn("updated_at")
	categoryDeletedAt   = query.MustColumn("deleted_at")

	categoryFullProjection = query.MustProjectionKey("full")
	categoryCodeSearch     = query.MustSearchKey("code")
	categoryNameSearch     = query.MustSearchKey("name")
	categoryCreatedSort    = query.MustSortKey("created_at")
	categoryIDSort         = query.MustSortKey("id")
)

func categorySpec() sharedrepository.EntitySpec[*model.Category, uuid.UUID] {
	columns := []query.Column{
		categoryID,
		categoryCode,
		categoryName,
		categoryDescription,
		categoryStatus,
		categoryCreatedAt,
		categoryUpdatedAt,
		categoryDeletedAt,
	}
	return sharedrepository.EntitySpec[*model.Category, uuid.UUID]{
		Table: categoryTable,
		Key:   categoryID,
		Query: query.Registry{
			SearchFields: map[query.SearchKey]query.Column{
				categoryCodeSearch: categoryCode,
				categoryNameSearch: categoryName,
			},
			SortFields: map[query.SortKey]query.Column{
				categoryCreatedSort: categoryCreatedAt,
				categoryIDSort:      categoryID,
			},
			Projections: map[query.ProjectionKey]query.Projection{
				categoryFullProjection: {Columns: columns},
			},
			DefaultProjection: categoryFullProjection,
			DefaultSort: []query.SortTerm{
				{Key: categoryCreatedSort, Direction: query.Descending},
			},
			TieBreaker: query.SortTerm{Key: categoryIDSort, Direction: query.Descending},
		},
		Projections: map[query.ProjectionKey]sharedrepository.Projection[*model.Category]{
			categoryFullProjection: {Scan: scanCategory},
		},
		InsertColumns: []query.Column{
			categoryID,
			categoryCode,
			categoryName,
			categoryDescription,
			categoryStatus,
			categoryCreatedAt,
			categoryUpdatedAt,
		},
		InsertValues: func(category *model.Category) []any {
			return []any{
				category.ID,
				category.Code,
				category.Name,
				category.Description,
				category.Status,
				category.CreatedAt,
				category.UpdatedAt,
			}
		},
		UpdateColumns: []query.Column{
			categoryCode,
			categoryName,
			categoryDescription,
			categoryStatus,
			categoryUpdatedAt,
		},
		UpdateValues: func(category *model.Category) []any {
			return []any{
				category.Code,
				category.Name,
				category.Description,
				category.Status,
				category.UpdatedAt,
			}
		},
		SoftDelete: &sharedrepository.SoftDeleteSpec{
			Column:       categoryDeletedAt,
			TouchColumns: []query.Column{categoryUpdatedAt},
		},
	}
}

func scanCategory(row sharedrepository.RowScanner) (*model.Category, error) {
	category := &model.Category{}
	if err := row.Scan(
		&category.ID,
		&category.Code,
		&category.Name,
		&category.Description,
		&category.Status,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	); err != nil {
		return nil, err
	}
	return category, nil
}
