package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"category-service/internal/category/model"
	"category-service/internal/category/service"
	"category-service/internal/shared/httpresponse"
	"category-service/internal/shared/pagination"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// Sample tra ve du lieu tinh, minh hoa cho endpoint public (khong yeu cau access token)
// nam trong cung 1 handler voi cac endpoint con lai dang bi bao ve boi auth middleware.
func (h *CategoryHandler) Sample(w http.ResponseWriter, r *http.Request) error {
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
	return nil
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) error {
	params, err := pagination.Parse(r)
	if err != nil {
		return httpresponse.NewBindingError("invalid pagination params", err)
	}

	categories, err := h.service.List(r.Context(), params)
	if err != nil {
		return err
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		categories,
	)
	return nil
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) error {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return httpresponse.NewBindingError("invalid category id", err)
	}

	category, err := h.service.Get(r.Context(), id)
	if err != nil {
		return err
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		category,
	)
	return nil
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var req model.CreateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpresponse.NewBindingError("invalid request body", err)
	}

	category, err := h.service.Create(r.Context(), &req)
	if err != nil {
		return err
	}

	httpresponse.WriteJSON(
		w,
		http.StatusCreated,
		category,
	)
	return nil
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) error {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return httpresponse.NewBindingError("invalid category id", err)
	}

	var req model.UpdateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpresponse.NewBindingError("invalid request body", err)
	}

	category, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		return err
	}

	httpresponse.WriteJSON(
		w,
		http.StatusOK,
		category,
	)
	return nil
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		return httpresponse.NewBindingError("invalid category id", err)
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
