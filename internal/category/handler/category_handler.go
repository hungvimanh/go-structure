package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"category-service/internal/category/model"
	"category-service/internal/category/service"
	"category-service/internal/category/validator"
	"category-service/internal/i18n"
	"category-service/internal/shared/apperror"
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

		writeJSON(
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

	writeJSON(
		w,
		status,
		map[string]string{
			"error": message,
		},
	)
}

func classifyError(err error) (int, string) {
	switch {
	case errors.Is(err, apperror.ErrBadRequest):
		return http.StatusBadRequest, "bad request"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func (h *CategoryHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	categories, err := h.service.List(r.Context())
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	writeJSON(
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

	writeJSON(
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

	writeJSON(
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

	writeJSON(
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
	writeJSON(
		w,
		http.StatusMisdirectedRequest,
		map[string]string{
			"error": message,
		},
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(
			w,
			`{"error":"failed to encode response"}`,
			http.StatusInternalServerError,
		)
	}
}
