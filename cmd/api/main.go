package main

import (
	"context"
	"log"
	"net/http"

	"category-service/internal/auth"
	"category-service/internal/category/handler"
	"category-service/internal/category/repository"
	"category-service/internal/category/service"
	"category-service/internal/config"
	"category-service/internal/database"
	"category-service/internal/i18n"
	"category-service/internal/shared/httpresponse"

	"github.com/go-chi/chi/v5"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	catalog, err := i18n.LoadCatalog(cfg.I18nDir)
	if err != nil {
		log.Fatalf("load i18n catalog: %v", err)
	}

	jwtPublicKey, err := auth.LoadPublicKey(cfg.JWTPublicKey)
	if err != nil {
		log.Fatalf("load jwt public key: %v", err)
	}

	log.Printf("PostgreSQL connected successfully")

	errorResponder := httpresponse.NewErrorResponder(catalog, cfg.DevMode)

	categoryRepository := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepository)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	router := chi.NewRouter()

	// Public — khong yeu cau access token, minh hoa cach public 1 endpoint
	// trong cung 1 handler voi cac endpoint con lai dang bi bao ve boi auth.
	router.Get("/categories/sample", errorResponder.Wrap(categoryHandler.Sample))

	// Protected — group rieng, chi middleware nay ap dung trong group,
	// khong lan ra route public o tren.
	router.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtPublicKey))

		r.Route("/categories", func(r chi.Router) {
			r.Get("/", errorResponder.Wrap(categoryHandler.List))
			r.Post("/", errorResponder.Wrap(categoryHandler.Create))
			r.Get("/{id}", errorResponder.Wrap(categoryHandler.Get))
			r.Put("/{id}", errorResponder.Wrap(categoryHandler.Update))
			r.Delete("/{id}", errorResponder.Wrap(categoryHandler.Delete))
		})
	})

	log.Printf("Server will run on port %s", cfg.AppPort)
	if err := http.ListenAndServe(":"+cfg.AppPort, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
