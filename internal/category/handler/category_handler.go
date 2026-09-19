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

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type CategoryHandler struct {
	service service.CategoryService
}

func NewCategoryHandler(service service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// Sample tra ve du lieu tinh, minh hoa cho endpoint public (khong yeu cau access token)
// nam trong cung 1 handler voi cac endpoint con lai dang bi bao ve boi auth middleware.
func (h *CategoryHandler) Sample(c *echo.Context) error {
	now := time.Now()

	sample := &model.Category{
		ID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Code:      "SAMPLE",
		Name:      "Sample Category",
		Status:    1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return c.JSON(http.StatusOK, sample)
}

func (h *CategoryHandler) List(c *echo.Context) error {
	req := c.Request()
	query := req.URL.Query()
	for key := range query {
		switch key {
		case "skip", "take", "search":
		default:
			return httpresponse.NewBindingError("invalid category list params", pagination.ErrInvalid)
		}
	}

	params, err := pagination.Parse(req)
	if err != nil {
		return httpresponse.NewBindingError("invalid category list params", err)
	}

	categoryFilter := model.CategoryFilter{}
	if search := strings.TrimSpace(query.Get("search")); search != "" {
		categoryFilter.Search = &search
	}

	categories, err := h.service.List(req.Context(), params, categoryFilter)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, categories)
}

// Search accepts typed Category filters while preserving the list response shape.
func (h *CategoryHandler) Search(c *echo.Context) error {
	req := c.Request()
	var body model.CategoryListRequest

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		return httpresponse.NewBindingError("invalid category search request", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return httpresponse.NewBindingError("invalid category search request", err)
	}

	params, err := pagination.FromValues(body.Skip, body.Take)
	if err != nil {
		return httpresponse.NewBindingError("invalid category search request", err)
	}

	categories, err := h.service.List(req.Context(), params, body.Filter)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, categories)
}

func (h *CategoryHandler) Get(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return httpresponse.NewBindingError("invalid category id", err)
	}

	category, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) Create(c *echo.Context) error {
	var req model.CreateCategoryRequest

	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httpresponse.NewBindingError("invalid request body", err)
	}

	category, err := h.service.Create(c.Request().Context(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) Update(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return httpresponse.NewBindingError("invalid category id", err)
	}

	var req model.UpdateCategoryRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return httpresponse.NewBindingError("invalid request body", err)
	}

	category, err := h.service.Update(c.Request().Context(), id, &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) Delete(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return httpresponse.NewBindingError("invalid category id", err)
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
