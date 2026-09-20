package httpresponse

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"category-service/internal/i18n"
	"category-service/internal/shared/apperror"
	"category-service/internal/shared/validation"
	"github.com/labstack/echo/v5"
)

// BindingError la loi cu phap request (id/JSON/query khong parse duoc).
// Handler tra ve loi nay de ErrorResponder.RespondError xu ly tap trung ve 421.
type BindingError struct {
	Message string
	Cause   error
}

func NewBindingError(message string, cause error) *BindingError {
	return &BindingError{Message: message, Cause: cause}
}

func (e *BindingError) Error() string { return e.Message }

func (e *BindingError) Unwrap() error { return e.Cause }

// AuthenticationError giu cause token verification de log server-side, trong khi
// ErrorResponder luon tra response 401 chung chung cho client.
type AuthenticationError struct {
	Cause error
}

func NewAuthenticationError(cause error) *AuthenticationError {
	return &AuthenticationError{Cause: cause}
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed: %v", e.Cause)
}

func (e *AuthenticationError) Unwrap() error { return e.Cause }

// ErrorResponder chuan hoa cach tra loi khi co error, dung chung cho moi module.
type ErrorResponder struct {
	catalog *i18n.Catalog
	// devMode: true khi DEV_MODE=1. Chi anh huong loi technical (khong phai
	// validation.Errors) — tra thang err.Error() thay vi message chung chung,
	// de debug local/staging. KHONG duoc bat o production (lo chi tiet loi ha tang).
	devMode bool
}

func NewErrorResponder(catalog *i18n.Catalog, devMode bool) *ErrorResponder {
	return &ErrorResponder{catalog: catalog, devMode: devMode}
}

// HTTPErrorHandler la mot diem vao duy nhat cho error tu Echo handler,
// middleware va panic recovery. Route 404/405 duoc Echo xu ly nhu fallback.
func (r *ErrorResponder) HTTPErrorHandler(c *echo.Context, err error) {
	response, unwrapErr := echo.UnwrapResponse(c.Response())
	if unwrapErr == nil && response.Committed {
		return
	}

	switch echo.StatusCode(err) {
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		echo.DefaultHTTPErrorHandler(false)(c, err)
	default:
		r.RespondError(c, err)
	}
}

// RespondError phan loai application error theo thu tu validation, binding,
// authentication, roi den technical error. Authentication khong bao gio lo cause
// qua DEV_MODE.
func (r *ErrorResponder) RespondError(c *echo.Context, err error) {
	if verrs, ok := errors.AsType[validation.Errors](err); ok {
		locale := i18n.ParseAcceptLanguage(c.Request().Header.Get("Accept-Language"))
		r.respondJSON(c, http.StatusBadRequest, map[string]any{
			"errors": r.catalog.TranslateAll(locale, verrs),
		})
		return
	}

	var bindingErr *BindingError
	if errors.As(err, &bindingErr) {
		r.respondJSON(c, http.StatusMisdirectedRequest, map[string]string{"error": bindingErr.Message})
		return
	}

	var authenticationErr *AuthenticationError
	if errors.As(err, &authenticationErr) {
		r.logError(err)
		r.respondJSON(c, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	status, message := classifyError(err)
	if status == http.StatusInternalServerError || status == http.StatusGatewayTimeout {
		r.logError(err)
	}

	if r.devMode {
		message = err.Error()
	}

	r.respondJSON(c, status, map[string]string{"error": message})
}

func (r *ErrorResponder) respondJSON(c *echo.Context, status int, data any) {
	if err := c.JSON(status, data); err != nil {
		r.logError(err)
	}
}

func (r *ErrorResponder) logError(err error) {
	log.Printf("internal error: %v", err)
}

func classifyError(err error) (int, string) {
	switch {
	case errors.Is(err, apperror.ErrBadRequest):
		return http.StatusBadRequest, "bad request"
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "request timeout"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
