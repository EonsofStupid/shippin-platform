// Command shippin-api is the WardenClyffe operator/server-control API.
// It is the only service that holds infrastructure authority (Proxmox tokens,
// fleet truth) and serves the UI-driven, Supabase-inspired data console.
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

	"github.com/shippin/shippin-api/internal/audit"
	"github.com/shippin/shippin-api/internal/automation"
	"github.com/shippin/shippin-api/internal/clyffy"
	"github.com/shippin/shippin-api/internal/connect"
	"github.com/shippin/shippin-api/internal/data"
	"github.com/shippin/shippin-api/internal/edge"
	"github.com/shippin/shippin-api/internal/fleet"
	"github.com/shippin/shippin-api/internal/identity"
	"github.com/shippin/shippin-api/internal/mesh"
	"github.com/shippin/shippin-api/internal/platform"
	"github.com/shippin/shippin-api/internal/proxmox"
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
	log.Printf("connected to postgres")

	// Contexts (bounded contexts wired here).
	auditSink := audit.New(pool)
	fleetStore := fleet.NewStore(pool)
	automationSvc := automation.NewService(pool, fleetStore, auditSink)
	dataStore := data.NewStore(pool)

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
		platform.JSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "shippin-api"})
	})

	idStore := identity.NewStore(cfg.OperatorUser, cfg.OperatorPass)
	// operatorAuthz gates mutating mesh routes: a valid bearer with operator role.
	operatorAuthz := func(token string) bool {
		op, ok := idStore.Validate(token)
		return ok && op.Role == "operator"
	}

	identity.NewHandler(idStore, identity.NewAccounts(pool)).Routes(r)
	fleet.NewHandler(fleetStore, operatorAuthz).Routes(r)
	automation.NewHandler(automationSvc, operatorAuthz).Routes(r)
	data.NewHandler(dataStore, operatorAuthz).Routes(r)
	mesh.NewHandler(mesh.NewStore(cfg.RepoRoot, cfg.SyncPlanPath, cfg.SyncBin), operatorAuthz).Routes(r)
	edge.NewHandler(edge.NewStore(pool), operatorAuthz).Routes(r)
	connect.NewHandler(connect.NewStore(cfg.RepoRoot), operatorAuthz).Routes(r)
	clyffy.NewHandler(clyffy.NewStore(pool)).Routes(r)
	// Slice 0: Proxmox substrate inventory + task-true start/stop.
	proxmox.NewHandler(proxmox.NewService(proxmox.LoadConfig(), pool, auditSink), operatorAuthz).Routes(r)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("shippin-api listening on %s", cfg.Addr)
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
