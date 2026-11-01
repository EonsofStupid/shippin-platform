// Command clyffe-api is the WardenClyffe customer-facing portal API.
//
// Unlike shippin-api, it holds NO infrastructure authority. It serves the
// customer portal, knowledge base, tickets, and CRM surfaces, reading only
// customer-safe data. Operator/infra data is reached through sanitized Warden
// endpoints, never directly.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/shippin/clyffe-api/internal/account"
	"github.com/shippin/clyffe-api/internal/platform"
)

func main() {
	cfg := platform.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := platform.NewPool(ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	if err := platform.ValidateCustomerRole(ctx, pool); err != nil {
		log.Fatalf("database role: %v", err)
	}
	log.Printf("connected to postgres")

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(platform.CORS)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			platform.Error(w, http.StatusServiceUnavailable, "db unreachable")
			return
		}
		platform.JSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "clyffe-api"})
	})

	account.NewHandler(account.NewStore(pool), cfg.CookieName).Routes(r)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("clyffe-api listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
}
