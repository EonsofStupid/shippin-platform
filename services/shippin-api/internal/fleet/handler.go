package fleet

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/shippin/shippin-api/internal/platform"
)

// Handler exposes fleet read endpoints under /api/warden/workspaces.
type Handler struct {
	store *Store
	authz platform.Authorizer
}

func NewHandler(store *Store, authz platform.Authorizer) *Handler {
	return &Handler{store: store, authz: authz}
}

func (h *Handler) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(platform.RequireOperator(h.authz))
		r.Get("/api/warden/workspaces", h.list)
		r.Get("/api/warden/workspaces/{id}", h.get)
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.List(r.Context())
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	res, err := h.store.Get(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, pgx.ErrNoRows) {
		platform.Error(w, http.StatusNotFound, "workspace not found")
		return
	}
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, res)
}
