package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shippin/shippin-api/internal/audit"
	"github.com/shippin/shippin-api/internal/automation"
	"github.com/shippin/shippin-api/internal/data"
	"github.com/shippin/shippin-api/internal/fleet"
	"github.com/shippin/shippin-api/internal/identity"
)

// Requires a disposable database initialized with migration 0001. It creates
// records and must never target a shared or production database.
func TestOperatorRoutesDatabaseIntegration(t *testing.T) {
	url := os.Getenv("SHIPPIN_TEST_DISPOSABLE_DATABASE_URL")
	if url == "" {
		t.Skip("disposable database not configured")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	identities := identity.NewStore("integration-operator", "integration-password")
	authorize := func(token string) bool { op, ok := identities.Validate(token); return ok && op.Role == "operator" }
	fleetStore := fleet.NewStore(pool)
	router := chi.NewRouter()
	identity.NewHandler(identities, nil).Routes(router)
	fleet.NewHandler(fleetStore, authorize).Routes(router)
	automation.NewHandler(automation.NewService(pool, fleetStore, audit.New(pool)), authorize).Routes(router)
	data.NewHandler(data.NewStore(pool), authorize).Routes(router)
	server := httptest.NewServer(router)
	defer server.Close()
	request := func(method, path, token, body string, want int) []byte {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+path, bytes.NewBufferString(body))
		if err != nil {
			t.Fatal(err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		result, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != want {
			t.Fatalf("%s %s: got %d, want %d: %s", method, path, response.StatusCode, want, result)
		}
		return result
	}
	request(http.MethodPost, "/api/warden/provision", "", `{"tier":"test"}`, http.StatusUnauthorized)
	var before int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM shippin_infra.resources`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if before != 0 {
		t.Fatal("integration database must start with no resources")
	}
	login := request(http.MethodPost, "/api/warden/identity/login", "", `{"username":"integration-operator","password":"integration-password"}`, http.StatusOK)
	var session struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(login, &session); err != nil || session.Token == "" {
		t.Fatal("login did not issue a session")
	}
	request(http.MethodGet, "/api/warden/data/schemas", session.Token, "", http.StatusOK)
	request(http.MethodGet, "/api/warden/data/schemas/shippin_core/tables", session.Token, "", http.StatusOK)
	request(http.MethodGet, "/api/warden/data/tables/shippin_core/tenants", session.Token, "", http.StatusOK)
	request(http.MethodGet, "/api/warden/data/tables/shippin_core/tenants/rows", session.Token, "", http.StatusOK)
	request(http.MethodPost, "/api/warden/data/query", session.Token, `{"sql":"SELECT 1 AS value"}`, http.StatusOK)
	created := request(http.MethodPost, "/api/warden/provision", session.Token, `{"tier":"test"}`, http.StatusCreated)
	var result struct {
		Resource struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(created, &result); err != nil || result.Resource.ID == "" || result.Resource.State != "requested" {
		t.Fatalf("unexpected request result: %s", created)
	}
	request(http.MethodGet, "/api/warden/workspaces", session.Token, "", http.StatusOK)
	request(http.MethodGet, "/api/warden/workspaces/"+result.Resource.ID, session.Token, "", http.StatusOK)
	var count int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM shippin_infra.resources`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("resource count = %d, error = %v", count, err)
	}
	request(http.MethodPost, "/api/warden/identity/logout", session.Token, "", http.StatusNoContent)
	request(http.MethodGet, "/api/warden/workspaces", session.Token, "", http.StatusUnauthorized)
	request(http.MethodPost, "/api/warden/data/query", session.Token, `{"sql":"SELECT 1"}`, http.StatusUnauthorized)
}
