package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"category-service/internal/category/model"
	"category-service/internal/category/repository"
	"category-service/internal/category/validator"
	"category-service/internal/shared/apperror"
	"category-service/internal/shared/pagination"
	"category-service/internal/shared/validation"

	"github.com/google/uuid"
)

type CategoryService interface {
	Count(ctx context.Context, categoryFilter model.CategoryFilter) (int64, error)
	List(ctx context.Context, params pagination.Params, categoryFilter model.CategoryFilter) ([]*model.Category, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Category, error)
	Create(ctx context.Context, req *model.CreateCategoryRequest) (*model.Category, error)
	Update(ctx context.Context, id uuid.UUID, req *model.UpdateCategoryRequest) (*model.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

var errCategoryIDRequired = fmt.Errorf("category id is required: %w", apperror.ErrBadRequest)

type categoryService struct {
	repository repository.CategoryRepository
}

func NewCategoryService(
	repository repository.CategoryRepository,
) CategoryService {
	return &categoryService{
		repository: repository,
	}
}

func (s *categoryService) Count(ctx context.Context, categoryFilter model.CategoryFilter) (int64, error) {
	count, err := s.repository.Count(ctx, categoryFilter)
	if err != nil {
		return 0, fmt.Errorf("count categories: %w", err)
	}
	return count, nil
}

func (s *categoryService) List(
	ctx context.Context,
	params pagination.Params,
	categoryFilter model.CategoryFilter,
) ([]*model.Category, error) {
	categories, err := s.repository.List(ctx, params, categoryFilter)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}

	return categories, nil
}

func (s *categoryService) Get(
	ctx context.Context,
	id uuid.UUID,
) (*model.Category, error) {

	if id == uuid.Nil {
		return nil, errCategoryIDRequired
	}

	category, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get category: %w", err)
	}
	if category == nil {
		return nil, validation.Errors{{Field: model.Category_ID, Code: validator.NotFound, Params: []any{id}}}
	}

	return category, nil
}

func (s *categoryService) Create(ctx context.Context, req *model.CreateCategoryRequest) (*model.Category, error) {

	if err := validator.Create(ctx, req, s.repository); err != nil {
		return nil, err
	}

	now := time.Now()

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate category id: %w", err)
	}

	category := &model.Category{
		ID:          id,
		Code:        strings.TrimSpace(req.Code),
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		Status:      req.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}

	if err := s.repository.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	return category, nil
}

func (s *categoryService) Update(ctx context.Context, id uuid.UUID, req *model.UpdateCategoryRequest) (*model.Category, error) {

	if id == uuid.Nil {
		return nil, errCategoryIDRequired
	}

	if err := validator.Update(ctx, id, req, s.repository); err != nil {
		return nil, err
	}

	category, err := s.repository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get category before update: %w", err)
	}
	if category == nil {
		return nil, validation.Errors{{Field: model.Category_ID, Code: validator.NotFound, Params: []any{id}}}
	}

	category.Code = strings.TrimSpace(req.Code)
	category.Name = strings.TrimSpace(req.Name)
	category.Description = req.Description
	category.Status = req.Status
	category.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}

	return category, nil
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID) error {

	if id == uuid.Nil {
		return errCategoryIDRequired
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	return nil
}
