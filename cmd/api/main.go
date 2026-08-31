package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"category-service/internal/auth"
	"category-service/internal/category"
	"category-service/internal/config"
	"category-service/internal/database"
	"category-service/internal/i18n"
	"category-service/internal/shared/httpresponse"
	"category-service/internal/shared/reqtimeout"

	"github.com/go-chi/chi/v5"
)

// Cac hang so tuning o tang http.Server — hardcode giong cach posgres.go hardcode
// MaxConns/MinConns (khong can chinh theo tung moi truong nhu REQUEST_TIMEOUT_SECONDS).
const (
	// readHeaderTimeout: thoi gian toi da doc xong request header — chan slowloris
	// (client gui header nho giot de giu connection).
	readHeaderTimeout = 5 * time.Second
	// readTimeout: thoi gian toi da doc xong ca header + body.
	readTimeout = 15 * time.Second
	// idleTimeout: thoi gian toi da 1 keep-alive connection duoc giu khi khong co request.
	idleTimeout = 90 * time.Second
	// serverTimeoutBuffer: cong them vao cfg.RequestTimeout de tinh WriteTimeout va
	// shutdownTimeout — phai > 0 de http.Server khong cat ngang response 504 ma
	// reqtimeout/ErrorResponder dang co gang ghi ra khi context vua het han, va de
	// Shutdown co du thoi gian cho 1 request dang chay (toi da cfg.RequestTimeout)
	// hoan tat truoc khi bi force-close.
	serverTimeoutBuffer = 5 * time.Second
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

	// Moi module tu chiu trach nhiem wiring (repository -> service -> handler)
	// va route cua chinh no — xem internal/category/module.go. Them module moi
	// chi can them 1 dong o day + 1 dong RegisterRoutes ben duoi.
	categoryModule := category.New(db)

	router := chi.NewRouter()

	// Middleware ngoai cung — phai dang ky truoc de bat duoc panic tu bat ky
	// middleware/handler nao ben trong (ke ca reqtimeout). Panic duoc doi thanh
	// error 500 qua RespondError thay vi lam sap ca server; xem httpresponse.go.
	router.Use(errorResponder.Recoverer)

	// Ap dung cho moi route (public lan protected) — phai khai bao truoc moi
	// route dang ky tren router nay (yeu cau cua chi). Khong tu ghi response
	// khi timeout, chi cancel context; xem internal/shared/reqtimeout.
	router.Use(reqtimeout.Middleware(cfg.RequestTimeout))

	categoryModule.RegisterRoutes(router, errorResponder, auth.Middleware(jwtPublicKey))

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

	serverErrs := make(chan error, 1)
	go func() {
		log.Printf("Server will run on port %s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrs <- err
			return
		}
		close(serverErrs)
	}()

	shutdownSignalCtx, stopNotify := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopNotify()

	select {
	case err, ok := <-serverErrs:
		if ok {
			log.Fatalf("server error: %v", err)
		}
	case <-shutdownSignalCtx.Done():
		stopNotify()
		log.Printf("shutdown signal received, draining in-flight requests")

		// Cung khoang thoi gian voi WriteTimeout: 1 request dang chay khi shutdown
		// bat dau co the mat toi da cfg.RequestTimeout de tu ket thuc (qua reqtimeout).
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout+serverTimeoutBuffer)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown did not finish cleanly: %v", err)
		} else {
			log.Printf("server shut down gracefully")
		}
	}

	// db.Close() (deferred o tren) chi chay sau khi Shutdown tra ve — dam bao
	// request dang chay xong xuoi truoc khi dong connection pool.
}
