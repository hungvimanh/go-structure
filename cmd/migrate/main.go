package main

import (
	"context"
	"fmt"
	"log"

	"category-service/internal/config"
	"category-service/internal/database"
	"category-service/internal/database/migration"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool, err := database.NewPostgresPool(context.Background(), cfg)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	if err := migration.Run(context.Background(), pool); err != nil {
		log.Fatal(fmt.Errorf("apply migrations: %w", err))
	}

	log.Print("database migrations applied successfully")
}
