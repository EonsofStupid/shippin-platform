package data

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/shippin/shippin-api/internal/platform"
)

// Handler exposes the Supabase-inspired data API under /api/warden/data.
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
		r.Get("/api/warden/data/schemas", h.schemas)
		r.Get("/api/warden/data/schemas/{schema}/tables", h.tables)
		r.Get("/api/warden/data/tables/{schema}/{table}", h.columns)
		r.Get("/api/warden/data/tables/{schema}/{table}/rows", h.rows)
		r.Post("/api/warden/data/query", h.query)
	})
}

func (h *Handler) schemas(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListSchemas(r.Context())
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) tables(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListTables(r.Context(), chi.URLParam(r, "schema"))
	if h.writeErr(w, err) {
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) columns(w http.ResponseWriter, r *http.Request) {
	cols, err := h.store.GetColumns(r.Context(), chi.URLParam(r, "schema"), chi.URLParam(r, "table"))
	if h.writeErr(w, err) {
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{
		"schema": chi.URLParam(r, "schema"), "table": chi.URLParam(r, "table"), "columns": cols})
}

func (h *Handler) rows(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	page, err := h.store.ListRows(r.Context(), chi.URLParam(r, "schema"), chi.URLParam(r, "table"), limit, offset)
	if h.writeErr(w, err) {
		return
	}
	platform.JSON(w, http.StatusOK, page)
}

type queryRequest struct {
	SQL string `json:"sql"`
}

func (h *Handler) query(w http.ResponseWriter, r *http.Request) {
	var req queryRequest
	if err := platform.DecodeJSON(r, &req); err != nil {
		platform.Error(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	if req.SQL == "" {
		platform.Error(w, http.StatusBadRequest, "sql is required")
		return
	}
	res, err := h.store.RunReadOnly(r.Context(), req.SQL)
	if err != nil {
		// SQL errors are user-facing here (read-only console), surface the message.
		platform.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, res)
}

// writeErr maps store errors to HTTP responses. Returns true if it wrote one.
func (h *Handler) writeErr(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, ErrForbiddenSchema):
		platform.Error(w, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrNotFound):
		platform.Error(w, http.StatusNotFound, err.Error())
	default:
		platform.Error(w, http.StatusInternalServerError, err.Error())
	}
	return true
}
