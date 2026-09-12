package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
	query := r.URL.Query()
	for key := range query {
		switch key {
		case "skip", "take", "search":
		default:
			return httpresponse.NewBindingError("invalid category list params", pagination.ErrInvalid)
		}
	}

	params, err := pagination.Parse(r)
	if err != nil {
		return httpresponse.NewBindingError("invalid category list params", err)
	}

	categoryFilter := model.CategoryFilter{}
	if search := strings.TrimSpace(query.Get("search")); search != "" {
		categoryFilter.Search = &search
	}

	categories, err := h.service.List(r.Context(), params, categoryFilter)
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

// Search accepts typed Category filters while preserving the list response shape.
func (h *CategoryHandler) Search(w http.ResponseWriter, r *http.Request) error {
	var req model.CategoryListRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return httpresponse.NewBindingError("invalid category search request", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return httpresponse.NewBindingError("invalid category search request", err)
	}

	params, err := pagination.FromValues(req.Skip, req.Take)
	if err != nil {
		return httpresponse.NewBindingError("invalid category search request", err)
	}

	categories, err := h.service.List(r.Context(), params, req.Filter)
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
