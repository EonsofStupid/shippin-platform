package automation

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/shippin/shippin-api/internal/platform"
)

type Handler struct {
	svc   *Service
	authz platform.Authorizer
}

func NewHandler(svc *Service, authz platform.Authorizer) *Handler {
	return &Handler{svc: svc, authz: authz}
}

func (h *Handler) Routes(r chi.Router) {
	r.With(platform.RequireOperator(h.authz)).Post("/api/warden/provision", h.provision)
}

func (h *Handler) provision(w http.ResponseWriter, r *http.Request) {
	var in ProvisionInput
	if err := platform.DecodeJSON(r, &in); err != nil {
		platform.Error(w, http.StatusBadRequest, "invalid body: "+err.Error())
		return
	}
	res, err := h.svc.Provision(r.Context(), in)
	if err != nil {
		platform.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	platform.JSON(w, http.StatusCreated, res)
}
