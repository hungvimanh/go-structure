package main

import (
	"context"
	"log"

	"category-service/internal/config"
	"category-service/internal/database"
	"category-service/internal/i18n"
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

	if _, err := i18n.LoadCatalog(cfg.I18nDir); err != nil {
		log.Fatalf("load i18n catalog: %v", err)
	}

	log.Printf("PostgreSQL connected successfully")
	log.Printf("Server will run on port %s", cfg.AppPort)
}
