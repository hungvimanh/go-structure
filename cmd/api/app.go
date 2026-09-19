package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"category-service/internal/auth"
	"category-service/internal/config"
	"category-service/internal/database"
	"category-service/internal/i18n"
	"category-service/internal/shared/httpresponse"
	"category-service/internal/shared/reqtimeout"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/fx"
)

func providePostgresPool(lc fx.Lifecycle, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := database.NewPostgresPool(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	log.Printf("PostgreSQL connected successfully")

	// db.Close() chi chay khi Fx dung app (OnStop chay theo thu tu nguoc voi
	// luc duoc Append) — xem registerHTTPServer, hook cua no duoc append sau
	// hook nay nen se OnStop truoc, dam bao request dang chay xong xuoi truoc
	// khi dong connection pool.
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}

func provideCatalog(cfg *config.Config) (*i18n.Catalog, error) {
	catalog, err := i18n.LoadCatalog(cfg.I18nDir)
	if err != nil {
		return nil, fmt.Errorf("load i18n catalog: %w", err)
	}
	return catalog, nil
}

func provideJWTPublicKey(cfg *config.Config) (*rsa.PublicKey, error) {
	key, err := auth.LoadPublicKey(cfg.JWTPublicKey)
	if err != nil {
		return nil, fmt.Errorf("load jwt public key: %w", err)
	}
	return key, nil
}

func provideAuthMiddleware(jwtPublicKey *rsa.PublicKey) echo.MiddlewareFunc {
	return auth.Middleware(jwtPublicKey)
}

func provideErrorResponder(cfg *config.Config, catalog *i18n.Catalog) *httpresponse.ErrorResponder {
	return httpresponse.NewErrorResponder(catalog, cfg.DevMode)
}

func provideRouter(cfg *config.Config, errorResponder *httpresponse.ErrorResponder) *echo.Echo {
	router := echo.New()
	router.HTTPErrorHandler = errorResponder.HTTPErrorHandler

	// Middleware ngoai cung — phai dang ky truoc de bat duoc panic tu bat ky
	// middleware/handler nao ben trong (ke ca reqtimeout).
	router.Use(middleware.Recover())

	// Ap dung cho moi route (public lan protected).
	router.Use(reqtimeout.Middleware(cfg.RequestTimeout))

	return router
}

// registerHTTPServer dang ky http.Server vao Fx lifecycle: OnStart bind port
// va serve trong goroutine rieng, OnStop shutdown graceful. Duoc invoke sau
// category.Module (xem main.go) de hook OnStop cua no chay truoc hook dong
// postgres pool.
func registerHTTPServer(lc fx.Lifecycle, cfg *config.Config, router *echo.Echo) {
	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		// WriteTimeout phai > cfg.RequestTimeout, khong thi http.Server tu cat
		// connection truoc khi reqtimeout/ErrorResponder kip ghi xong response 504.
		WriteTimeout: cfg.RequestTimeout + serverTimeoutBuffer,
		IdleTimeout:  idleTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return fmt.Errorf("listen on %s: %w", server.Addr, err)
			}

			log.Printf("Server will run on port %s", cfg.AppPort)
			go func() {
				if err := server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatalf("server error: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Printf("shutdown signal received, draining in-flight requests")

			if err := server.Shutdown(ctx); err != nil {
				log.Printf("graceful shutdown did not finish cleanly: %v", err)
				return err
			}

			log.Printf("server shut down gracefully")
			return nil
		},
	})
}
