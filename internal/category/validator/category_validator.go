package validator

import (
	"context"
	"strings"

	"category-service/internal/category/model"
	"category-service/internal/shared/validation"

	"github.com/google/uuid"
)

const (
	codeMaxLength        = 50
	nameMaxLength        = 200
	descriptionMaxLength = 500
	stringEmpty          = ""
)

const (
	CodeRequired       string = "CATEGORY_CODE_REQUIRED"
	CodeTooLong        string = "CATEGORY_CODE_TOO_LONG"
	CodeDuplicate      string = "CATEGORY_CODE_DUPLICATE"
	NameRequired       string = "CATEGORY_NAME_REQUIRED"
	NameTooLong        string = "CATEGORY_NAME_TOO_LONG"
	DescriptionTooLong string = "CATEGORY_DESCRIPTION_TOO_LONG"
	NotFound           string = "CATEGORY_NOT_FOUND"
)

type CodeChecker interface {
	ExistsByCode(ctx context.Context, code string, excludeID uuid.UUID) (bool, error)
}

func Create(ctx context.Context, req *model.CreateCategoryRequest, checker CodeChecker) error {
	if req == nil {
		return validation.Err
	}

	var errs validation.Errors

	codeErrs, dbErr := validateCode(ctx, checker, req.Code, uuid.Nil)
	errs = append(errs, codeErrs...)
	errs = append(errs, validateName(req.Name)...)
	errs = append(errs, validateDescription(req.Description)...)

	if errs != nil {
		return errs
	}
	if dbErr != nil {
		return dbErr
	}
	return nil
}

func Update(ctx context.Context, id uuid.UUID, req *model.UpdateCategoryRequest, checker CodeChecker) error {
	if req == nil {
		return validation.Err
	}

	var errs validation.Errors

	codeErrs, dbErr := validateCode(ctx, checker, req.Code, id)
	errs = append(errs, codeErrs...)
	errs = append(errs, validateName(req.Name)...)
	errs = append(errs, validateDescription(req.Description)...)

	if errs != nil {
		return errs
	}
	if dbErr != nil {
		return dbErr
	}
	return nil
}

func validateCode(ctx context.Context, checker CodeChecker, code string, excludeID uuid.UUID) (validation.Errors, error) {
	code = strings.TrimSpace(code)

	if code == stringEmpty {
		return validation.Errors{{Field: model.Category_Code, Code: CodeRequired}}, nil
	}
	if len([]rune(code)) > codeMaxLength {
		return validation.Errors{{Field: model.Category_Code, Code: CodeTooLong, Params: []any{codeMaxLength}}}, nil
	}

	exists, err := checker.ExistsByCode(ctx, code, excludeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return validation.Errors{{Field: model.Category_Code, Code: CodeDuplicate}}, nil
	}

	return nil, nil
}

func validateName(name string) validation.Errors {
	name = strings.TrimSpace(name)

	if name == stringEmpty {
		return validation.Errors{{Field: model.Category_Name, Code: NameRequired}}
	}
	if len([]rune(name)) > nameMaxLength {
		return validation.Errors{{Field: model.Category_Name, Code: NameTooLong, Params: []any{nameMaxLength}}}
	}
	return nil
}

func validateDescription(description *string) validation.Errors {
	if description == nil {
		return nil
	}
	if len([]rune(*description)) > descriptionMaxLength {
		return validation.Errors{{Field: model.Category_Description, Code: DescriptionTooLong, Params: []any{descriptionMaxLength}}}
	}
	return nil
}
