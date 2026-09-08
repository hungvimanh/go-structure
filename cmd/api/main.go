package main

import (
	"log"
	"time"

	"category-service/internal/category"
	"category-service/internal/config"

	"go.uber.org/fx"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	newApp(cfg).Run()
}

// newApp lap rap toan bo Fx app tu Config. Tach rieng khoi main() de
// main_test.go co the fx.ValidateApp() cung 1 bo option ma khong can that su
// Start (khong can postgres/port that).
func newApp(cfg *config.Config) *fx.App {
	return fx.New(appOptions(cfg))
}

// appOptions gom toan bo Fx option cua app, tach rieng khoi newApp de
// main_test.go co the fx.ValidateApp(appOptions(cfg)) — kiem tra dependency
// graph (thieu provider, sai type, ...) ma khong can goi that constructor
// (ValidateApp khong invoke input function), nen khong can postgres/port that.
func appOptions(cfg *config.Config) fx.Option {
	// Fx dung timeout nay cho ca Start (bind port) lan Stop (drain + shutdown
	// http server) — cung khoang thoi gian voi WriteTimeout ben duoi, du de 1
	// request dang chay khi shutdown bat dau tu ket thuc qua reqtimeout.
	lifecycleTimeout := cfg.RequestTimeout + serverTimeoutBuffer

	return fx.Options(
		fx.Supply(cfg),
		fx.StartTimeout(lifecycleTimeout),
		fx.StopTimeout(lifecycleTimeout),

		fx.Provide(
			providePostgresPool,
			provideCatalog,
			provideJWTPublicKey,
			provideAuthMiddleware,
			provideErrorResponder,
			provideRouter,
		),

		// Moi module tu chiu trach nhiem provide (repository -> service ->
		// handler) va route cua chinh no — xem internal/category/module.go.
		// Them module moi chi can them 1 dong o day.
		category.Module,

		// Phai invoke sau category.Module de hook OnStop cua http server
		// duoc append sau hook dong postgres pool (xem providePostgresPool),
		// dam bao thu tu shutdown dung.
		fx.Invoke(registerHTTPServer),
	)
}
