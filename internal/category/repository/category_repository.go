package repository

import (
	"context"
	"errors"

	"category-service/internal/category/model"
	"category-service/internal/shared/pagination"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository interface {
	Count(ctx context.Context) (int64, error)

	List(ctx context.Context, params pagination.Params) ([]*model.Category, error)

	Get(ctx context.Context, id uuid.UUID) (*model.Category, error)

	ExistsByCode(ctx context.Context, code string, excludeID uuid.UUID) (bool, error)

	Create(ctx context.Context, category *model.Category) error

	Update(ctx context.Context, category *model.Category) error

	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

func (r *categoryRepository) Count(
	ctx context.Context,
) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM Category
		WHERE DeletedAt IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *categoryRepository) List(
	ctx context.Context,
	params pagination.Params,
) ([]*model.Category, error) {
	query := `
		SELECT
			Id,
			Code,
			Name,
			Description,
			Status,
			CreatedAt,
			UpdatedAt,
			DeletedAt
		FROM Category
		WHERE DeletedAt IS NULL
		ORDER BY CreatedAt DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, params.Take, params.Skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*model.Category

	for rows.Next() {
		category := &model.Category{}

		err := rows.Scan(
			&category.ID,
			&category.Code,
			&category.Name,
			&category.Description,
			&category.Status,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.DeletedAt,
		)

		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) Get(
	ctx context.Context,
	id uuid.UUID,
) (*model.Category, error) {
	query := `
		SELECT
			Id,
			Code,
			Name,
			Description,
			Status,
			CreatedAt,
			UpdatedAt,
			DeletedAt
		FROM Category
		WHERE Id = $1
		  AND DeletedAt IS NULL
	`

	category := &model.Category{}

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&category.ID,
		&category.Code,
		&category.Name,
		&category.Description,
		&category.Status,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return category, nil
}

func (r *categoryRepository) ExistsByCode(
	ctx context.Context,
	code string,
	excludeID uuid.UUID,
) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM Category
			WHERE Code = $1
			  AND Id != $2
			  AND DeletedAt IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(
		ctx,
		query,
		code,
		excludeID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *categoryRepository) Create(
	ctx context.Context,
	category *model.Category,
) error {
	query := `
		INSERT INTO Category (
			Id,
			Code,
			Name,
			Description,
			Status,
			CreatedAt,
			UpdatedAt
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		category.ID,
		category.Code,
		category.Name,
		category.Description,
		category.Status,
		category.CreatedAt,
		category.UpdatedAt,
	)

	return err
}

func (r *categoryRepository) Update(
	ctx context.Context,
	category *model.Category,
) error {
	query := `
		UPDATE Category
		SET Code = $1,
			Name = $2,
			Description = $3,
			Status = $4,
			UpdatedAt = $5
		WHERE Id = $6
		  AND DeletedAt IS NULL
	`
	_, err := r.db.Exec(
		ctx,
		query,
		category.Code,
		category.Name,
		category.Description,
		category.Status,
		category.UpdatedAt,
		category.ID,
	)

	return err
}

func (r *categoryRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	query := `
		UPDATE Category
		SET DeletedAt = NOW(),
			UpdatedAt = NOW()
		WHERE Id = $1
		  AND DeletedAt IS NULL
	`

	_, err := r.db.Exec(
		ctx,
		query,
		id,
	)

	return err
}
