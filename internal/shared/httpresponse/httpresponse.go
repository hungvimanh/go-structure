package httpresponse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"category-service/internal/i18n"
	"category-service/internal/shared/apperror"
	"category-service/internal/shared/validation"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(
			w,
			`{"error":"failed to encode response"}`,
			http.StatusInternalServerError,
		)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// WriteBindingError dung cho loi cu phap request (id khong parse duoc, JSON
// body decode loi, query param sai dinh dang, ...) — ngang hang moi module,
// khong phai business error nen khong di qua ErrorResponder.
func WriteBindingError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusMisdirectedRequest, message)
}

// BindingError la loi cu phap request (id/JSON/query khong parse duoc).
// Handler tra ve loi nay (thay vi tu ghi response) de ErrorResponder.RespondError
// xu ly tap trung ve 421. Cause giu lai error goc (uuid.Parse, json.Decode, ...)
// de sau nay co log storage thi van log/trace duoc nguyen nhan that.
type BindingError struct {
	Message string
	Cause   error
}

func NewBindingError(message string, cause error) *BindingError {
	return &BindingError{Message: message, Cause: cause}
}

func (e *BindingError) Error() string { return e.Message }
func (e *BindingError) Unwrap() error { return e.Cause }

// HandlerFunc la chu ky handler nghiep vu: chi mo ta cong vao (w, r) va cong ra
// (response da ghi, hoac error). Khong tu quyet dinh status/message khi loi —
// viec do la cua ErrorResponder. Dung chung cho moi module qua Wrap.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Wrap bien 1 HandlerFunc thanh http.HandlerFunc de dang ky voi router (chi/net-http).
// Day la diem duy nhat goi RespondError cho toan bo handler da wrap — them
// binding error moi, hay sau nay can log error vao DB/queue, chi sua trong
// RespondError (hoac logError), khong phai sua tung handler.
func (r *ErrorResponder) Wrap(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := h(w, req); err != nil {
			r.RespondError(w, req, err)
		}
	}
}

// Recoverer la middleware chi-compatible (func(http.Handler) http.Handler) —
// dang ky qua router.Use(errorResponder.Recoverer), phai la middleware ngoai
// cung (dang ky truoc reqtimeout va moi middleware khac) de bat duoc panic tu
// bat ky dau trong chain, khong rieng gi handler nghiep vu.
// Panic duoc doi thanh error kem stack trace, roi di qua RespondError nhu moi
// loi technical khac — nghia la van 500, van duoc logError, va chi lo stack
// trace ra client khi devMode=true (giong moi loi technical khac, khong dac cach).
func (r *ErrorResponder) Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Errorf("panic: %v\n%s", rec, debug.Stack())
				r.RespondError(w, req, err)
			}
		}()

		next.ServeHTTP(w, req)
	})
}

// ErrorResponder chuan hoa cach tra loi khi co error, dung chung cho moi module
// (khong rieng gi category). Moi module chi can 1 instance, tao 1 lan trong main
// va truyen vao qua constructor cua tung handler.
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

// RespondError la diem duy nhat trong toan bo app quyet dinh status/message
// tu 1 error, va la diem duy nhat log error technical. Phan loai theo thu tu:
//   - validation.Errors -> 400, kem danh sach loi da dich theo Accept-Language.
//     Khong bi anh huong boi devMode vi day khong phai loi technical.
//   - *BindingError -> 421 qua WriteBindingError (message do handler dat san).
//   - error khac (technical) -> theo classifyError (apperror.ErrBadRequest -> 400,
//     context.DeadlineExceeded -> 504 khi reqtimeout middleware cancel request,
//     con lai -> 500), 500/504 duoc log qua logError. Neu devMode=true, message
//     tra ve la err.Error() day du thay vi message chung chung.
func (r *ErrorResponder) RespondError(w http.ResponseWriter, req *http.Request, err error) {
	if verrs, ok := errors.AsType[validation.Errors](err); ok {
		locale := i18n.ParseAcceptLanguage(req.Header.Get("Accept-Language"))

		WriteJSON(
			w,
			http.StatusBadRequest,
			map[string]any{
				"errors": r.catalog.TranslateAll(locale, verrs),
			},
		)
		return
	}

	var bindingErr *BindingError
	if errors.As(err, &bindingErr) {
		WriteBindingError(w, bindingErr.Message)
		return
	}

	status, message := classifyError(err)
	if status == http.StatusInternalServerError || status == http.StatusGatewayTimeout {
		r.logError(err)
	}

	if r.devMode {
		message = err.Error()
	}

	WriteError(w, status, message)
}

// logError la diem duy nhat ghi log cho loi technical (500). Sau nay can luu
// vao DB hoac ban len queue, chi sua trong ham nay.
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
