package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"category-service/internal/category/model"
	"category-service/internal/shared/filter"
	"category-service/internal/shared/pagination"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
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
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &categoryRepository{
		db: db,
	}
}

func (r *categoryRepository) Count(
	ctx context.Context,
	categoryFilter model.CategoryFilter,
) (int64, error) {
	whereClause, args, _ := buildFilter(categoryFilter, 1)
	query := `
		SELECT COUNT(*)
		FROM Category
		WHERE DeletedAt IS NULL` + whereClause

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *categoryRepository) List(
	ctx context.Context,
	params pagination.Params,
	categoryFilter model.CategoryFilter,
) ([]*model.Category, error) {
	whereClause, args, nextArg := buildFilter(categoryFilter, 1)
	query := fmt.Sprintf(`
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
		WHERE DeletedAt IS NULL%s
		ORDER BY CreatedAt DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, nextArg, nextArg+1)

	args = append(args, params.Take, params.Skip)
	rows, err := r.db.Query(ctx, query, args...)
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

func buildFilter(categoryFilter model.CategoryFilter, startArg int) (string, []any, int) {
	var predicates []string
	var args []any
	nextArg := startArg

	if categoryFilter.Search != nil {
		placeholder := fmt.Sprintf("$%d", nextArg)
		appendFilterPredicate(
			&predicates,
			&args,
			&nextArg,
			fmt.Sprintf("(Code ILIKE %s ESCAPE '\\' OR Name ILIKE %s ESCAPE '\\')", placeholder, placeholder),
			"%"+escapeLikePattern(*categoryFilter.Search)+"%",
		)
	}

	appendStringFilter(&predicates, &args, &nextArg, "Code", categoryFilter.Code)
	appendStringFilter(&predicates, &args, &nextArg, "Name", categoryFilter.Name)
	appendIntFilter(&predicates, &args, &nextArg, "Status", categoryFilter.Status)

	if len(predicates) == 0 {
		return "", args, nextArg
	}

	return " AND " + strings.Join(predicates, " AND "), args, nextArg
}

func appendStringFilter(
	predicates *[]string,
	args *[]any,
	nextArg *int,
	column string,
	stringFilter *filter.StringFilter,
) {
	if stringFilter == nil {
		return
	}

	if stringFilter.Eq != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s = $%d", column, *nextArg), *stringFilter.Eq)
	}
	if stringFilter.Neq != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s <> $%d", column, *nextArg), *stringFilter.Neq)
	}
	if stringFilter.Contains != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s ILIKE $%d ESCAPE '\\'", column, *nextArg), "%"+escapeLikePattern(*stringFilter.Contains)+"%")
	}
	if stringFilter.StartsWith != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s ILIKE $%d ESCAPE '\\'", column, *nextArg), escapeLikePattern(*stringFilter.StartsWith)+"%")
	}
	if stringFilter.EndsWith != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s ILIKE $%d ESCAPE '\\'", column, *nextArg), "%"+escapeLikePattern(*stringFilter.EndsWith))
	}
	if len(stringFilter.In) > 0 {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s = ANY($%d)", column, *nextArg), stringFilter.In)
	}
	if len(stringFilter.NotIn) > 0 {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("NOT (%s = ANY($%d))", column, *nextArg), stringFilter.NotIn)
	}
}

func appendIntFilter(
	predicates *[]string,
	args *[]any,
	nextArg *int,
	column string,
	intFilter *filter.IntFilter,
) {
	if intFilter == nil {
		return
	}

	if intFilter.Eq != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s = $%d", column, *nextArg), *intFilter.Eq)
	}
	if intFilter.Neq != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s <> $%d", column, *nextArg), *intFilter.Neq)
	}
	if intFilter.Gt != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s > $%d", column, *nextArg), *intFilter.Gt)
	}
	if intFilter.Gte != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s >= $%d", column, *nextArg), *intFilter.Gte)
	}
	if intFilter.Lt != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s < $%d", column, *nextArg), *intFilter.Lt)
	}
	if intFilter.Lte != nil {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s <= $%d", column, *nextArg), *intFilter.Lte)
	}
	if len(intFilter.In) > 0 {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("%s = ANY($%d)", column, *nextArg), intFilter.In)
	}
	if len(intFilter.NotIn) > 0 {
		appendFilterPredicate(predicates, args, nextArg, fmt.Sprintf("NOT (%s = ANY($%d))", column, *nextArg), intFilter.NotIn)
	}
}

func appendFilterPredicate(predicates *[]string, args *[]any, nextArg *int, predicate string, arg any) {
	*predicates = append(*predicates, predicate)
	*args = append(*args, arg)
	*nextArg++
}

func escapeLikePattern(value string) string {
	return strings.NewReplacer(
		"\\", "\\\\",
		"%", "\\%",
		"_", "\\_",
	).Replace(value)
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
