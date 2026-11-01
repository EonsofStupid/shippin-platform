package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/shippin/shippin-api/internal/automation"
	"github.com/shippin/shippin-api/internal/clyffy"
	"github.com/shippin/shippin-api/internal/data"
	"github.com/shippin/shippin-api/internal/fleet"
	"github.com/shippin/shippin-api/internal/identity"
)

func TestOperatorDataAndFleetRoutes(t *testing.T) {
	identityStore := identity.NewStore("test-operator", "test-password")
	token, _, err := identityStore.Login("test-operator", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	authorize := func(token string) bool {
		operator, ok := identityStore.Validate(token)
		return ok && operator.Role == "operator"
	}
	router := chi.NewRouter()
	fleet.NewHandler(fleet.NewStore(nil), authorize).Routes(router)
	automation.NewHandler(automation.NewService(nil, nil, nil), authorize).Routes(router)
	data.NewHandler(data.NewStore(nil), authorize).Routes(router)
	clyffy.NewHandler(clyffy.NewStore(nil), authorize).Routes(router)
	routes := []struct{ method, path string }{
		{http.MethodGet, "/api/warden/workspaces"},
		{http.MethodGet, "/api/warden/workspaces/example"},
		{http.MethodPost, "/api/warden/provision"},
		{http.MethodGet, "/api/warden/data/schemas"},
		{http.MethodGet, "/api/warden/data/schemas/shippin_core/tables"},
		{http.MethodGet, "/api/warden/data/tables/shippin_core/tenants"},
		{http.MethodGet, "/api/warden/data/tables/shippin_core/tenants/rows"},
		{http.MethodPost, "/api/warden/data/query"},
		{http.MethodGet, "/api/clyffy/home"},
		{http.MethodGet, "/api/clyffy/services"},
		{http.MethodGet, "/api/clyffy/platforms"},
		{http.MethodGet, "/api/clyffy/intelligence"},
		{http.MethodGet, "/api/clyffy/plugins"},
		{http.MethodGet, "/api/clyffy/connectors"},
	}
	checkDenied := func(authorization string) {
		t.Helper()
		for _, route := range routes {
			req := httptest.NewRequest(route.method, route.path, strings.NewReader(`{}`))
			req.Header.Set("Authorization", authorization)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != http.StatusUnauthorized {
				t.Errorf("%s %s: got %d, want 401", route.method, route.path, response.Code)
			}
		}
	}
	checkDenied("")
	checkDenied("Bearer invalid")
	// The real handlers reject invalid inputs before touching the nil test database.
	for _, path := range []string{"/api/warden/provision", "/api/warden/data/query"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{`))
		req.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != http.StatusBadRequest {
			t.Errorf("valid operator %s: got %d, want 400 from handler", path, response.Code)
		}
	}
	identityStore.Logout(token)
	checkDenied("Bearer " + token)
}
