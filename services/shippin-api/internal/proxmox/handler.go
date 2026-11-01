package proxmox

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/shippin/shippin-api/internal/platform"
)

// Handler exposes Proxmox inventory + task-true actions.
type Handler struct {
	svc   *Service
	authz platform.Authorizer
}

func NewHandler(svc *Service, authz platform.Authorizer) *Handler {
	return &Handler{svc: svc, authz: authz}
}

// Routes mounts under /api/warden/proxmox.
func (h *Handler) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(platform.RequireOperator(h.authz))
		r.Get("/api/warden/proxmox/status", h.status)
		r.Get("/api/warden/proxmox/guests", h.guests)
		r.Post("/api/warden/proxmox/guests/{node}/{kind}/{vmid}/start", h.start)
		r.Post("/api/warden/proxmox/guests/{node}/{kind}/{vmid}/stop", h.stop)
		r.Get("/api/warden/proxmox/actions/{id}", h.action)
	})
}

func (h *Handler) status(w http.ResponseWriter, r *http.Request) {
	platform.JSON(w, http.StatusOK, h.svc.Status(r.Context()))
}

func (h *Handler) guests(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListGuests(r.Context())
	if err != nil {
		if err.Error() == "proxmox not configured" {
			platform.Error(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		platform.Error(w, http.StatusBadGateway, err.Error())
		return
	}
	platform.JSON(w, http.StatusOK, map[string]any{"count": len(list), "items": list})
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	h.act(w, r, true)
}

func (h *Handler) stop(w http.ResponseWriter, r *http.Request) {
	h.act(w, r, false)
}

func (h *Handler) act(w http.ResponseWriter, r *http.Request, start bool) {
	node := chi.URLParam(r, "node")
	kind := chi.URLParam(r, "kind")
	vmid, err := strconv.Atoi(chi.URLParam(r, "vmid"))
	if err != nil {
		platform.Error(w, http.StatusBadRequest, "invalid vmid")
		return
	}
	var res *ActionResult
	if start {
		res, err = h.svc.StartGuest(r.Context(), node, kind, vmid)
	} else {
		res, err = h.svc.StopGuest(r.Context(), node, kind, vmid)
	}
	if err != nil {
		if err.Error() == "proxmox not configured" {
			platform.Error(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		platform.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	code := http.StatusOK
	if res.Status == "failed" {
		code = http.StatusConflict
	}
	platform.JSON(w, code, res)
}

func (h *Handler) action(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	out, err := h.svc.GetAction(r.Context(), id)
	if err != nil {
		platform.Error(w, http.StatusNotFound, "action not found")
		return
	}
	platform.JSON(w, http.StatusOK, out)
}
