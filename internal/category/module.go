package category

import (
	"net/http"

	"category-service/internal/category/handler"
	"category-service/internal/category/repository"
	"category-service/internal/category/service"
	"category-service/internal/shared/httpresponse"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Module gom toan bo wiring cua category (repository -> service -> handler) va
// cach dang ky route. main.go chi can goi New() + RegisterRoutes(), khong can
// biet chi tiet khoi tao ben trong — them 1 module khac (vd "product") chi them
// 2 dong trong main.go thay vi ca khoi khoi tao + dinh nghia route nhu truoc,
// giup giam conflict khi nhieu nguoi cung sua main.go de them module moi.
type Module struct {
	handler *handler.CategoryHandler
}

func New(db *pgxpool.Pool) *Module {
	categoryRepository := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepository)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	return &Module{handler: categoryHandler}
}

// RegisterRoutes dang ky ca route public (sample, khong qua authMiddleware) va
// route protected (CRUD, qua authMiddleware) vao router truyen vao.
func (m *Module) RegisterRoutes(
	router chi.Router,
	errorResponder *httpresponse.ErrorResponder,
	authMiddleware func(http.Handler) http.Handler,
) {
	router.Get("/categories/sample", errorResponder.Wrap(m.handler.Sample))

	router.Group(func(r chi.Router) {
		r.Use(authMiddleware)

		r.Route("/categories", func(r chi.Router) {
			r.Get("/", errorResponder.Wrap(m.handler.List))
			r.Post("/", errorResponder.Wrap(m.handler.Create))
			r.Get("/{id}", errorResponder.Wrap(m.handler.Get))
			r.Put("/{id}", errorResponder.Wrap(m.handler.Update))
			r.Delete("/{id}", errorResponder.Wrap(m.handler.Delete))
		})
	})
}
