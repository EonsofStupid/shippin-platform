package proxmox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/shippin/shippin-api/internal/audit"
)

// Service owns inventory + task-true guest actions (control plane).
type Service struct {
	cfg     Config
	cli     *Client
	db      *pgxpool.Pool
	audit   *audit.Sink
	initErr error
}

func NewService(cfg Config, db *pgxpool.Pool, au *audit.Sink) *Service {
	var cli *Client
	var initErr error
	if cfg.Configured() {
		cli, initErr = NewClient(cfg)
	}
	return &Service{cfg: cfg, cli: cli, db: db, audit: au, initErr: initErr}
}

// Status is operator-facing connectivity/config state (no secrets).
type Status struct {
	Configured bool           `json:"configured"`
	Host       string         `json:"host"`
	Port       int            `json:"port"`
	Node       string         `json:"node"`
	Reachable  bool           `json:"reachable"`
	Version    map[string]any `json:"version,omitempty"`
	Error      string         `json:"error,omitempty"`
	Message    string         `json:"message"`
}

func (s *Service) Status(ctx context.Context) Status {
	st := Status{
		Configured: s.cfg.Configured(),
		Host:       s.cfg.Host,
		Port:       s.cfg.Port,
		Node:       s.cfg.Node,
	}
	if !st.Configured {
		st.Message = "Configure the Proxmox host and API token (see secrets/proxmox.env.example)."
		return st
	}
	if s.initErr != nil {
		st.Error = s.initErr.Error()
		st.Message = "Proxmox configuration is invalid."
		return st
	}
	ver, err := s.cli.Version(ctx)
	if err != nil {
		st.Error = err.Error()
		st.Message = "Credentials present but Proxmox API not reachable."
		return st
	}
	st.Reachable = true
	st.Version = ver
	st.Message = "Proxmox substrate online. Warden can inventory and act."
	return st
}

func (s *Service) requireClient() error {
	if s.initErr != nil {
		return s.initErr
	}
	if s.cli == nil || !s.cfg.Configured() {
		return fmt.Errorf("proxmox not configured")
	}
	return nil
}

// ListGuests returns live inventory from Proxmox.
func (s *Service) ListGuests(ctx context.Context) ([]Guest, error) {
	if err := s.requireClient(); err != nil {
		return nil, err
	}
	return s.cli.ListGuests(ctx)
}

// ActionResult is returned after start/stop.
type ActionResult struct {
	ActionRequestID string `json:"action_request_id"`
	Kind            string `json:"kind"` // guest_start | guest_stop
	Node            string `json:"node"`
	GuestKind       string `json:"guest_kind"`
	VMID            int    `json:"vmid"`
	UPID            string `json:"upid,omitempty"`
	Status          string `json:"status"` // action_request status
	ExitStatus      string `json:"exit_status,omitempty"`
	Error           string `json:"error,omitempty"`
}

// StartGuest runs task-true start.
func (s *Service) StartGuest(ctx context.Context, node, guestKind string, vmid int) (*ActionResult, error) {
	if err := s.requireClient(); err != nil {
		return nil, err
	}
	return s.runGuestAction(ctx, "guest_start", node, guestKind, vmid, s.cli.StartGuest)
}

// StopGuest runs task-true stop.
func (s *Service) StopGuest(ctx context.Context, node, guestKind string, vmid int) (*ActionResult, error) {
	if err := s.requireClient(); err != nil {
		return nil, err
	}
	return s.runGuestAction(ctx, "guest_stop", node, guestKind, vmid, s.cli.StopGuest)
}

type guestActor func(ctx context.Context, node, kind string, vmid int) (string, error)

func (s *Service) runGuestAction(ctx context.Context, kind, node, guestKind string, vmid int, act guestActor) (*ActionResult, error) {
	if err := s.requireClient(); err != nil {
		return nil, err
	}
	if guestKind != "qemu" && guestKind != "lxc" {
		return nil, fmt.Errorf("kind must be qemu or lxc")
	}
	if node == "" {
		node = s.cfg.Node
	}

	payload, _ := json.Marshal(map[string]any{
		"node": node, "guest_kind": guestKind, "vmid": vmid, "action": kind,
	})

	var actionID string
	err := s.db.QueryRow(ctx,
		`INSERT INTO shippin_core.action_requests (kind, status, payload)
		 VALUES ($1, 'planned', $2) RETURNING id::text`, kind, payload).Scan(&actionID)
	if err != nil {
		return nil, fmt.Errorf("create action_request: %w", err)
	}

	// planned → approved → executing (slice 0 auto-approves operator actions)
	_, _ = s.db.Exec(ctx,
		`UPDATE shippin_core.action_requests SET status='approved', updated_at=now() WHERE id=$1`, actionID)
	_, _ = s.db.Exec(ctx,
		`UPDATE shippin_core.action_requests SET status='executing', updated_at=now() WHERE id=$1`, actionID)

	_ = s.audit.Append(ctx, audit.Event{
		Action:     kind + ".executing",
		TargetKind: "proxmox_guest",
		TargetID:   fmt.Sprintf("%s/%s/%d", node, guestKind, vmid),
		Data:       map[string]any{"action_request_id": actionID},
	})

	upid, err := act(ctx, node, guestKind, vmid)
	if err != nil {
		_, _ = s.db.Exec(ctx,
			`UPDATE shippin_core.action_requests SET status='failed', error=$2, updated_at=now() WHERE id=$1`,
			actionID, err.Error())
		_ = s.audit.Append(ctx, audit.Event{
			Action: kind + ".failed", TargetKind: "proxmox_guest",
			TargetID: fmt.Sprintf("%s/%s/%d", node, guestKind, vmid),
			Data:     map[string]any{"action_request_id": actionID, "error": err.Error()},
		})
		return &ActionResult{
			ActionRequestID: actionID, Kind: kind, Node: node, GuestKind: guestKind, VMID: vmid,
			Status: "failed", Error: err.Error(),
		}, nil
	}

	st, waitErr := s.cli.WaitTask(ctx, node, upid, 90*time.Second)
	exit := ""
	if st != nil {
		exit = st.ExitStatus
	}
	final := "succeeded"
	errMsg := ""
	if waitErr != nil {
		final = "failed"
		errMsg = waitErr.Error()
	} else if exit != "" && exit != "OK" {
		final = "failed"
		errMsg = "proxmox exitstatus=" + exit
	}

	resultJSON, _ := json.Marshal(map[string]any{"upid": upid, "exitstatus": exit})
	_, _ = s.db.Exec(ctx,
		`UPDATE shippin_core.action_requests
		 SET status=$2, result=$3, error=NULLIF($4,''), updated_at=now() WHERE id=$1`,
		actionID, final, resultJSON, errMsg)

	_ = s.audit.Append(ctx, audit.Event{
		Action:     kind + "." + final,
		TargetKind: "proxmox_guest",
		TargetID:   fmt.Sprintf("%s/%s/%d", node, guestKind, vmid),
		Data:       map[string]any{"action_request_id": actionID, "upid": upid, "exitstatus": exit},
	})

	return &ActionResult{
		ActionRequestID: actionID, Kind: kind, Node: node, GuestKind: guestKind, VMID: vmid,
		UPID: upid, Status: final, ExitStatus: exit, Error: errMsg,
	}, nil
}

// GetAction returns action_request status by id.
func (s *Service) GetAction(ctx context.Context, id string) (map[string]any, error) {
	var kind, status string
	var payload, result []byte
	var errText *string
	var created, updated time.Time
	err := s.db.QueryRow(ctx,
		`SELECT kind, status::text, payload, result, error, created_at, updated_at
		 FROM shippin_core.action_requests WHERE id=$1`, id).
		Scan(&kind, &status, &payload, &result, &errText, &created, &updated)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"id": id, "kind": kind, "status": status,
		"created_at": created, "updated_at": updated,
	}
	if len(payload) > 0 {
		var p any
		_ = json.Unmarshal(payload, &p)
		out["payload"] = p
	}
	if len(result) > 0 {
		var r any
		_ = json.Unmarshal(result, &r)
		out["result"] = r
	}
	if errText != nil {
		out["error"] = *errText
	}
	return out, nil
}
