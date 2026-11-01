package clyffy

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/shippin/shippin-api/internal/platform"
)

// Handler exposes the Clyffy orchestrator under /api/clyffy.
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
		r.Get("/api/clyffy/home", h.home)
		r.Get("/api/clyffy/services", h.services)
		r.Get("/api/clyffy/platforms", h.platforms)
		r.Get("/api/clyffy/intelligence", h.intelligence)
		r.Get("/api/clyffy/plugins", h.plugins)
		r.Get("/api/clyffy/connectors", h.connectors)
	})
}

func (h *Handler) plugins(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListByDesignation(r.Context(), "plugin")
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{
		"designation": "plugin", "audience": "ai-only", "items": items, "count": len(items)})
}

func (h *Handler) connectors(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListByDesignation(r.Context(), "connector")
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{
		"designation": "connector", "audience": "operator+ai (shared)", "items": items, "count": len(items)})
}

func (h *Handler) services(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListServices(r.Context())
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{"items": items, "count": len(items)})
}

func (h *Handler) platforms(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListPlatforms(r.Context())
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) intelligence(w http.ResponseWriter, r *http.Request) {
	probes, err := h.store.IntelligenceStatus(r.Context())
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{"plane": "clyffy", "probes": probes})
}

// home is the orchestrator overview: what Warden has captured + intelligence
// plane health + what configuration is owed before customer service.
func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	services, err := h.store.ListServices(ctx)
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platforms, err := h.store.ListPlatforms(ctx)
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	probes, _ := h.store.IntelligenceStatus(ctx)

	byRole := map[string]int{}
	byDesignation := map[string]int{}
	owedTotal := 0
	credentialGated := 0
	for _, s := range services {
		byRole[s.Role]++
		byDesignation[s.Designation]++
		owedTotal += len(s.Owed)
		if s.AuthRequired && s.CredentialRef != nil {
			credentialGated++
		}
	}

	platform.JSON(w, http.StatusOK, map[string]any{
		"orchestrator": "warden-clyffy",
		"persona":      "clyffy-master",
		"summary": map[string]any{
			"platforms":                 len(platforms),
			"services":                  len(services),
			"services_by_role":          byRole,
			"services_by_designation":   byDesignation,
			"config_items_owed":         owedTotal,
			"credential_gated_services": credentialGated,
		},
		"platforms":    platforms,
		"intelligence": probes,
		"note":         "Clyffy reads Warden's captured foundation model. Live managed-service connections require Infisical credentials referenced per service.",
	})
}
