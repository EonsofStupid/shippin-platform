package proxmox

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/shippin/shippin-api/internal/identity"
)

func TestRoutesRequireOperator(t *testing.T) {
	store := identity.NewStore("test-operator", "test-password")
	token, _, err := store.Login("test-operator", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	authorize := func(token string) bool {
		op, ok := store.Validate(token)
		return ok && op.Role == "operator"
	}
	router := chi.NewRouter()
	NewHandler(NewService(Config{}, nil, nil), authorize).Routes(router)
	routes := []struct{ method, path string }{
		{http.MethodGet, "/api/warden/proxmox/status"},
		{http.MethodGet, "/api/warden/proxmox/guests"},
		{http.MethodPost, "/api/warden/proxmox/guests/node/qemu/101/start"},
		{http.MethodPost, "/api/warden/proxmox/guests/node/qemu/101/stop"},
		{http.MethodGet, "/api/warden/proxmox/actions/id"},
	}
	for _, route := range routes {
		for _, credential := range []string{"", "Bearer invalid"} {
			t.Run(route.method+" "+route.path+" "+credential, func(t *testing.T) {
				req := httptest.NewRequest(route.method, route.path, nil)
				req.Header.Set("Authorization", credential)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)
				if rec.Code != http.StatusUnauthorized {
					t.Fatalf("got %d, want 401", rec.Code)
				}
			})
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/api/warden/proxmox/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid operator: got %d, want 200", rec.Code)
	}
	store.Logout(token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("revoked token: got %d, want 401", rec.Code)
	}
}
