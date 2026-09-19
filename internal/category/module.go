package category

import (
	"category-service/internal/category/handler"
	"category-service/internal/category/repository"
	"category-service/internal/category/service"

	"github.com/labstack/echo/v5"
	"go.uber.org/fx"
)

// Module gom toan bo Fx wiring cua category (provide repository -> service ->
// handler) va dang ky route qua invoke. Them 1 module khac (vd "product") chi
// can them 1 fx.Module tuong tu vao danh sach options trong main.go, khong
// can biet chi tiet khoi tao ben trong.
var Module = fx.Module("category",
	fx.Provide(
		repository.NewCategoryRepository,
		service.NewCategoryService,
		handler.NewCategoryHandler,
	),
	fx.Invoke(registerRoutes),
)

// registerRoutes dang ky ca route public (sample, khong qua authMiddleware) va
// route protected (CRUD, qua authMiddleware) vao router.
func registerRoutes(
	router *echo.Echo,
	h *handler.CategoryHandler,
	authMiddleware echo.MiddlewareFunc,
) {
	router.GET("/categories/sample", h.Sample)

	categories := router.Group("/categories", authMiddleware)
	categories.GET("/", h.List)
	categories.POST("/search", h.Search)
	categories.POST("/", h.Create)
	categories.GET("/:id", h.Get)
	categories.PUT("/:id", h.Update)
	categories.DELETE("/:id", h.Delete)
}
