package query

import "fmt"

// ErrorCode classifies request/query validation failures before database access.
type ErrorCode string

const (
	InvalidIdentifier   ErrorCode = "invalid_identifier"
	InvalidSortKey      ErrorCode = "invalid_sort_key"
	InvalidSearchKey    ErrorCode = "invalid_search_key"
	InvalidProjectionKey ErrorCode = "invalid_projection_key"
	InvalidDirection    ErrorCode = "invalid_direction"
	InvalidPagination   ErrorCode = "invalid_pagination"
	ComplexityExceeded  ErrorCode = "complexity_exceeded"
)

// Error reports an invalid descriptor or query request.
type Error struct {
	Code    ErrorCode
	Message string
}

func (value *Error) Error() string {
	return fmt.Sprintf("query %s: %s", value.Code, value.Message)
}

func newError(code ErrorCode, message string) error {
	return &Error{Code: code, Message: message}
}
