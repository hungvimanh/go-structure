package repository

import (
	"context"

	"category-service/internal/category/model"
	"category-service/internal/shared/pagination"
	"category-service/internal/shared/query"
	sharedrepository "category-service/internal/shared/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository interface {
	Count(ctx context.Context, categoryFilter model.CategoryFilter) (int64, error)
	List(ctx context.Context, params pagination.Params, categoryFilter model.CategoryFilter) ([]*model.Category, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Category, error)
	ExistsByCode(ctx context.Context, code string, excludeID uuid.UUID) (bool, error)
	Create(ctx context.Context, category *model.Category) error
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryRepository struct {
	engine sharedrepository.Repository[*model.Category, uuid.UUID]
}

// NewCategoryRepository creates the Category adapter over its static schema
// descriptor. Descriptor validation failures are programmer errors and cannot
// be recovered at request time.
func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	engine, err := sharedrepository.New(db, categorySpec())
	if err != nil {
		panic(err)
	}
	return &categoryRepository{engine: engine}
}

func (value *categoryRepository) Count(ctx context.Context, categoryFilter model.CategoryFilter) (int64, error) {
	count, err := value.engine.Count(ctx, categoryFilterExpression(categoryFilter))
	return int64(count), err
}

func (value *categoryRepository) List(ctx context.Context, params pagination.Params, categoryFilter model.CategoryFilter) ([]*model.Category, error) {
	return value.engine.List(ctx, categoryFilterExpression(categoryFilter), categoryListOptions(params.Skip, params.Take, categoryFilter.Search))
}

func (value *categoryRepository) Get(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	category, found, err := value.engine.Get(ctx, id)
	if err != nil || !found {
		return nil, err
	}
	return category, nil
}

func (value *categoryRepository) ExistsByCode(ctx context.Context, code string, excludeID uuid.UUID) (bool, error) {
	predicate := query.And(
		query.String(categoryCode, query.StringEqual, code),
		query.Not(query.Compare(categoryID, query.Equal, excludeID)),
	)
	return value.engine.Exists(ctx, predicate)
}

func (value *categoryRepository) Create(ctx context.Context, category *model.Category) error {
	_, err := value.engine.Create(ctx, category)
	return err
}

func (value *categoryRepository) Update(ctx context.Context, category *model.Category) error {
	_, err := value.engine.Update(ctx, category.ID, category)
	return err
}

func (value *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := value.engine.Delete(ctx, id, sharedrepository.SoftDelete)
	return err
}
