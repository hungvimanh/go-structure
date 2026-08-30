package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"category-service/internal/category/model"
	"category-service/internal/category/service"
	"category-service/internal/category/validator"
	"category-service/internal/i18n"
	"category-service/internal/shared/apperror"
	"category-service/internal/shared/httpresponse"
	"category-service/internal/shared/pagination"
	"category-service/internal/shared/validation"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	service service.CategoryService
	catalog *i18n.Catalog
}

func NewCategoryHandler(
	service service.CategoryService,
	catalog *i18n.Catalog,
) *CategoryHandler {
	return &CategoryHandler{
		service: service,
		catalog: catalog,
	}
}

func (h *CategoryHandler) respondError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	if verrs, ok := errors.AsType[validation.Errors](err); ok {
		locale := i18n.ParseAcceptLanguage(r.Header.Get("Accept-Language"))

		status := http.StatusBadRequest
		if len(verrs) == 1 && verrs[0].Code == validator.NotFound {
			status = http.StatusNotFound
		}

		httpresponse.WriteJSON(
			w,
			status,
			map[string]any{
				"errors": h.catalog.TranslateAll(locale, verrs),
			},
		)
		return
	}

	status, message := classifyError(err)
	if status == http.StatusInternalServerError {
		log.Printf("internal error: %v", err)
	}

	httpresponse.WriteError(w, status, message)
}

func classifyError(err error) (int, string) {
	switch {
	case errors.Is(err, apperror.ErrBadRequest):
		return http.StatusBadRequest, "bad request"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

// Sample tra ve du lieu tinh, minh hoa cho endpoint public (khong yeu cau access token)
// nam trong cung 1 handler voi cac endpoint con lai dang bi bao ve boi auth middleware.
func (h *CategoryHandler) Sample(
	w http.ResponseWriter,
	r *http.Request,
) {
	now := time.Now()

	sample := &model.Category{
		ID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Code:      "SAMPLE",
		Name:      "Sample Category",
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		sample,
	)
}

func (h *CategoryHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	params, err := pagination.Parse(r)
	if err != nil {
		writeBindingError(w, "invalid pagination params")
		return
	}

	categories, err := h.service.List(r.Context(), params)
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		categories,
	)
}

func (h *CategoryHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		writeBindingError(w, "invalid category id")
		return
	}

	category, err := h.service.Get(
		r.Context(),
		id,
	)
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		category,
	)
}

func (h *CategoryHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req model.CreateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBindingError(w, "invalid request body")
		return
	}

	category, err := h.service.Create(
		r.Context(),
		&req,
	)
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	httpresponse.WriteJSON(
		w,
		http.StatusCreated,
		category,
	)
}

func (h *CategoryHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		writeBindingError(w, "invalid category id")
		return
	}

	var req model.UpdateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBindingError(w, "invalid request body")
		return
	}

	category, err := h.service.Update(
		r.Context(),
		id,
		&req,
	)
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		category,
	)
}

func (h *CategoryHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		writeBindingError(w, "invalid category id")
		return
	}

	if err := h.service.Delete(
		r.Context(),
		id,
	); err != nil {
		h.respondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeBindingError(
	w http.ResponseWriter,
	message string,
) {
	httpresponse.WriteError(w, http.StatusMisdirectedRequest, message)
}
