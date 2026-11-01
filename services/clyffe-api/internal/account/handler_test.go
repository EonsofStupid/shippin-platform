package account

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCustomerRoutesFailClosed(t *testing.T) {
	router := chi.NewRouter()
	// A nil store would panic if any route reached its database handler.
	NewHandler(nil, "shippin_session").Routes(router)
	for _, path := range []string{
		"/api/clyffe/home",
		"/api/clyffe/accounts",
		"/api/clyffe/orders",
		"/api/clyffe/orders?tenant_id=another-tenant",
	} {
		for _, authorization := range []string{"", "Bearer arbitrary-token"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("Authorization", authorization)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			if response.Code != http.StatusServiceUnavailable {
				t.Errorf("GET %s with %q: got %d, want 503", path, authorization, response.Code)
			}
		}
	}
}
