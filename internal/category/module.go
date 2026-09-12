package category

import (
	"net/http"

	"category-service/internal/category/handler"
	"category-service/internal/category/repository"
	"category-service/internal/category/service"
	"category-service/internal/shared/httpresponse"

	"github.com/go-chi/chi/v5"
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
	router *chi.Mux,
	h *handler.CategoryHandler,
	errorResponder *httpresponse.ErrorResponder,
	authMiddleware func(http.Handler) http.Handler,
) {

	router.Get("/categories/sample", errorResponder.Wrap(h.Sample))

	router.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Route("/categories", func(r chi.Router) {
			r.Get("/", errorResponder.Wrap(h.List))
			r.Post("/search", errorResponder.Wrap(h.Search))
			r.Post("/", errorResponder.Wrap(h.Create))
			r.Get("/{id}", errorResponder.Wrap(h.Get))
			r.Put("/{id}", errorResponder.Wrap(h.Update))
			r.Delete("/{id}", errorResponder.Wrap(h.Delete))
		})
	})
}
