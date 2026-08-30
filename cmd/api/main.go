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

	categoryRepository := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepository)
	categoryHandler := handler.NewCategoryHandler(categoryService, catalog)

	router := chi.NewRouter()

	// Public — khong yeu cau access token, minh hoa cach public 1 endpoint
	// trong cung 1 handler voi cac endpoint con lai dang bi bao ve boi auth.
	router.Get("/categories/sample", categoryHandler.Sample)

	// Protected — group rieng, chi middleware nay ap dung trong group,
	// khong lan ra route public o tren.
	router.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtPublicKey))

		r.Route("/categories", func(r chi.Router) {
			r.Get("/", categoryHandler.List)
			r.Post("/", categoryHandler.Create)
			r.Get("/{id}", categoryHandler.Get)
			r.Put("/{id}", categoryHandler.Update)
			r.Delete("/{id}", categoryHandler.Delete)
		})
	})

	log.Printf("Server will run on port %s", cfg.AppPort)
	if err := http.ListenAndServe(":"+cfg.AppPort, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
