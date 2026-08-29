package validation

import (
	"errors"
	"strings"
)

var Err = errors.New("validation error")

type FieldError struct {
	Field  string
	Code   string
	Params []any
}

type Errors []FieldError

func (e Errors) Error() string {
	codes := make([]string, len(e))
	for i, fe := range e {
		codes[i] = fe.Code
	}
	return "validation failed: " + strings.Join(codes, ", ")
}

func (e Errors) Unwrap() error {
	if len(e) == 0 {
		return nil
	}
	return Err
}
