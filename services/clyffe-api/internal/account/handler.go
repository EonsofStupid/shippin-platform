package account

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/shippin/clyffe-api/internal/platform"
)

// Handler exposes the customer portal under /api/clyffe.
type Handler struct {
	store      *Store
	cookieName string
}

func NewHandler(store *Store, cookieName string) *Handler {
	return &Handler{store: store, cookieName: cookieName}
}

func (h *Handler) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.requireCustomer)
		r.Get("/api/clyffe/home", h.home)
		r.Get("/api/clyffe/accounts", h.accounts)
		r.Get("/api/clyffe/orders", h.orders)
	})
}

type subjectContextKey struct{}

func (h *Handler) requireCustomer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if h.store == nil || h.store.db == nil {
			platform.Error(w, http.StatusServiceUnavailable, "customer account access is not available")
			return
		}
		cookie, err := r.Cookie(h.cookieName)
		if err != nil {
			platform.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		subject, err := h.store.SessionSubject(r.Context(), cookie.Value)
		if errors.Is(err, ErrUnauthenticated) {
			platform.Error(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if err != nil {
			platform.Error(w, http.StatusServiceUnavailable, "customer account access is not available")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), subjectContextKey{}, subject)))
	})
}

func subject(r *http.Request) string {
	value, _ := r.Context().Value(subjectContextKey{}).(string)
	return value
}

func (h *Handler) accounts(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListAccounts(r.Context(), subject(r))
	if err != nil {
		platform.Error(w, http.StatusServiceUnavailable, "accounts are unavailable")
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

func (h *Handler) orders(w http.ResponseWriter, r *http.Request) {
	tenant := r.URL.Query().Get("tenant_id")
	if tenant == "" {
		platform.Error(w, http.StatusBadRequest, "organization is required")
		return
	}
	items, err := h.store.ListOrders(r.Context(), subject(r), tenant)
	if errors.Is(err, ErrTenantNotFound) {
		platform.Error(w, http.StatusNotFound, "organization not found")
		return
	}
	if err != nil {
		platform.Error(w, http.StatusServiceUnavailable, "orders are unavailable")
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

// home is the customer portal overview: the account(s) and a roll-up of orders.
func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accounts, err := h.store.ListAccounts(ctx, subject(r))
	if err != nil {
		platform.Error(w, http.StatusServiceUnavailable, "accounts are unavailable")
		return
	}
	byStatus := map[string]int{}
	count := 0
	for _, account := range accounts {
		orders, err := h.store.ListOrders(ctx, subject(r), account.ID)
		if err != nil {
			platform.Error(w, http.StatusServiceUnavailable, "orders are unavailable")
			return
		}
		for _, o := range orders {
			byStatus[o.Status]++
			count++
		}
	}
	platform.JSON(w, http.StatusOK, map[string]any{
		"portal":  "clyffe",
		"persona": "clyffe-customer",
		"summary": map[string]any{
			"accounts":         len(accounts),
			"orders":           count,
			"orders_by_status": byStatus,
		},
		"accounts": accounts,
		"note":     "Clyffe is the customer plane. It reads customer-safe data only; operator infra is reached through sanitized Warden endpoints.",
	})
}
