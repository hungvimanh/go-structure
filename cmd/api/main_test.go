package main

import (
	"testing"
	"time"

	"category-service/internal/config"

	"go.uber.org/fx"
)

// TestAppOptions_ValidateApp kiem tra dependency graph (moi provider deu co
// nguoi cung cap, khong thieu/thua type) ma khong that su Start app — Fx
// ValidateApp khong invoke constructor nao, nen khong can postgres/port that.
func TestAppOptions_ValidateApp(t *testing.T) {
	cfg := &config.Config{
		AppPort:        "8080",
		DBHost:         "localhost",
		DBPort:         "5432",
		DBUser:         "user",
		DBPassword:     "password",
		DBName:         "db",
		DBSSLMode:      "disable",
		I18nDir:        "i18n",
		JWTPublicKey:   "jwt_public.pem",
		JWTPrivateKey:  "jwt_private.pem",
		DevMode:        true,
		RequestTimeout: 10 * time.Second,
	}

	if err := fx.ValidateApp(appOptions(cfg)); err != nil {
		t.Fatalf("appOptions dependency graph invalid: %v", err)
	}
}
