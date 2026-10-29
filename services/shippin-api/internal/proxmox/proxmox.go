// Package proxmox is the Warden client for Proxmox VE (substrate).
// Proxmox powers VMs/LXCs; this package is how Warden talks to that OS.
package proxmox

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is loaded from environment (secrets/proxmox.env or process env).
// Never log TokenID or TokenSecret.
type Config struct {
	Host      string
	Port      int
	Node      string // default node hint
	TokenID   string // user@realm!tokenid
	TokenSec  string
	VerifyTLS bool
}

// LoadConfig reads PROXMOX_* env vars. Also loads a dotenv-style file if set:
// SHIPPIN_PROXMOX_ENV, or secrets/proxmox.env under SHIPPIN_REPO_ROOT when present.
// Never logs file contents.
func LoadConfig() Config {
	loadProxmoxEnvFile()
	port := 8006
	if p := os.Getenv("PROXMOX_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	verify := false
	switch strings.ToLower(os.Getenv("PROXMOX_VERIFY_TLS")) {
	case "1", "true", "yes":
		verify = true
	}
	return Config{
		Host:      env("PROXMOX_HOST", "10.0.0.1"),
		Port:      port,
		Node:      env("PROXMOX_NODE", "server1"),
		TokenID:   os.Getenv("PROXMOX_TOKEN_ID"),
		TokenSec:  os.Getenv("PROXMOX_TOKEN_SECRET"),
		VerifyTLS: verify,
	}
}

func loadProxmoxEnvFile() {
	candidates := []string{}
	if p := os.Getenv("SHIPPIN_PROXMOX_ENV"); p != "" {
		candidates = append(candidates, p)
	}
	root := os.Getenv("SHIPPIN_REPO_ROOT")
	if root == "" {
		root = "/workspace/WardenClyffe-latest"
	}
	candidates = append(candidates, root+"/secrets/proxmox.env")
	for _, path := range candidates {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(b), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			i := strings.IndexByte(line, '=')
			if i < 1 {
				continue
			}
			k := strings.TrimSpace(line[:i])
			v := strings.TrimSpace(line[i+1:])
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
		return
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// Configured reports whether credentials are present (not whether API works).
func (c Config) Configured() bool {
	return c.Host != "" && c.TokenID != "" && c.TokenSec != ""
}

// Client talks to Proxmox REST API.
type Client struct {
	cfg    Config
	http   *http.Client
	base   string
	auth   string
}

func NewClient(cfg Config) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !cfg.VerifyTLS}, //nolint:gosec // operator-controlled lab default
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 30 * time.Second, Transport: tr},
		base: fmt.Sprintf("https://%s:%d/api2/json", cfg.Host, cfg.Port),
		auth: fmt.Sprintf("PVEAPIToken=%s=%s", cfg.TokenID, cfg.TokenSec),
	}
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.auth)
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return fmt.Errorf("proxmox GET %s: %s: %s", path, res.Status, truncate(string(body), 200))
	}
	var wrap struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return err
	}
	return json.Unmarshal(wrap.Data, out)
}

func (c *Client) post(ctx context.Context, path string, form url.Values) (string, error) {
	var reader io.Reader
	if form != nil {
		reader = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, reader)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", c.auth)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("proxmox POST %s: %s: %s", path, res.Status, truncate(string(body), 200))
	}
	var wrap struct {
		Data any `json:"data"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		return "", err
	}
	switch v := wrap.Data.(type) {
	case string:
		return v, nil
	default:
		b, _ := json.Marshal(v)
		return string(b), nil
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Version hits /version (connectivity probe).
func (c *Client) Version(ctx context.Context) (map[string]any, error) {
	var m map[string]any
	if err := c.get(ctx, "/version", &m); err != nil {
		return nil, err
	}
	return m, nil
}

// NodeStatus is a subset of node status fields.
type NodeStatus struct {
	Node string  `json:"node"`
	// CPU/mem from resources listing when available
}

// Guest is a qemu or lxc guest on a node.
type Guest struct {
	Node        string  `json:"node"`
	Kind        string  `json:"kind"` // qemu | lxc
	VMID        int     `json:"vmid"`
	Name        string  `json:"name"`
	Status      string  `json:"status"` // running | stopped | ...
	CPUs        float64 `json:"cpus,omitempty"`
	MaxMem      int64   `json:"maxmem,omitempty"`
	MaxDisk     int64   `json:"maxdisk,omitempty"`
	Uptime      int64   `json:"uptime,omitempty"`
	Template    int     `json:"template,omitempty"`
	StableLabel string  `json:"stable_label"` // e.g. proxmox.server1.qemu.116
}

// ListGuests returns VMs and LXCs on the configured default node, or all nodes if node empty.
func (c *Client) ListGuests(ctx context.Context) ([]Guest, error) {
	nodes := []string{c.cfg.Node}
	if c.cfg.Node == "" || c.cfg.Node == "*" {
		var nl []struct {
			Node string `json:"node"`
		}
		if err := c.get(ctx, "/nodes", &nl); err != nil {
			return nil, err
		}
		nodes = nodes[:0]
		for _, n := range nl {
			nodes = append(nodes, n.Node)
		}
	}
	var out []Guest
	for _, node := range nodes {
		gs, err := c.listNodeGuests(ctx, node)
		if err != nil {
			return nil, err
		}
		out = append(out, gs...)
	}
	return out, nil
}

func (c *Client) listNodeGuests(ctx context.Context, node string) ([]Guest, error) {
	var out []Guest
	// qemu
	var qemus []struct {
		VMID     int     `json:"vmid"`
		Name     string  `json:"name"`
		Status   string  `json:"status"`
		CPUs     float64 `json:"cpus"`
		MaxMem   int64   `json:"maxmem"`
		MaxDisk  int64   `json:"maxdisk"`
		Uptime   int64   `json:"uptime"`
		Template int     `json:"template"`
	}
	if err := c.get(ctx, "/nodes/"+url.PathEscape(node)+"/qemu", &qemus); err != nil {
		return nil, err
	}
	for _, q := range qemus {
		out = append(out, Guest{
			Node: node, Kind: "qemu", VMID: q.VMID, Name: q.Name, Status: q.Status,
			CPUs: q.CPUs, MaxMem: q.MaxMem, MaxDisk: q.MaxDisk, Uptime: q.Uptime, Template: q.Template,
			StableLabel: fmt.Sprintf("proxmox.%s.qemu.%d", node, q.VMID),
		})
	}
	// lxc
	var lxcs []struct {
		VMID     int     `json:"vmid"`
		Name     string  `json:"name"`
		Status   string  `json:"status"`
		CPUs     float64 `json:"cpus"`
		MaxMem   int64   `json:"maxmem"`
		MaxDisk  int64   `json:"maxdisk"`
		Uptime   int64   `json:"uptime"`
		Template int     `json:"template"`
	}
	if err := c.get(ctx, "/nodes/"+url.PathEscape(node)+"/lxc", &lxcs); err != nil {
		return nil, err
	}
	for _, x := range lxcs {
		out = append(out, Guest{
			Node: node, Kind: "lxc", VMID: x.VMID, Name: x.Name, Status: x.Status,
			CPUs: x.CPUs, MaxMem: x.MaxMem, MaxDisk: x.MaxDisk, Uptime: x.Uptime, Template: x.Template,
			StableLabel: fmt.Sprintf("proxmox.%s.lxc.%d", node, x.VMID),
		})
	}
	return out, nil
}

// Storage is a Proxmox storage record (subset).
type Storage struct {
	Storage string `json:"storage"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Active  int    `json:"active,omitempty"`
	Enabled int    `json:"enabled,omitempty"`
}

// ListStorage returns datacenter storage (used to pick vzdump targets).
func (c *Client) ListStorage(ctx context.Context) ([]Storage, error) {
	var items []Storage
	if err := c.get(ctx, "/storage", &items); err != nil {
		return nil, err
	}
	return items, nil
}

// Vzdump starts a guest image backup; returns a UPID.
func (c *Client) Vzdump(ctx context.Context, node string, form url.Values) (string, error) {
	path := "/nodes/" + url.PathEscape(node) + "/vzdump"
	return c.post(ctx, path, form)
}

// StartGuest posts start; returns UPID string.
func (c *Client) StartGuest(ctx context.Context, node, kind string, vmid int) (string, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d/status/start", url.PathEscape(node), kind, vmid)
	return c.post(ctx, path, nil)
}

// StopGuest posts stop (not shutdown — hard stop for slice 0 simplicity).
func (c *Client) StopGuest(ctx context.Context, node, kind string, vmid int) (string, error) {
	path := fmt.Sprintf("/nodes/%s/%s/%d/status/stop", url.PathEscape(node), kind, vmid)
	return c.post(ctx, path, nil)
}

// TaskStatus is a Proxmox task (UPID) status.
type TaskStatus struct {
	Status    string `json:"status"`    // running | stopped
	ExitStatus string `json:"exitstatus,omitempty"`
	UPID      string `json:"upid,omitempty"`
	Type      string `json:"type,omitempty"`
	User      string `json:"user,omitempty"`
}

// WaitTask polls task until not running or timeout.
func (c *Client) WaitTask(ctx context.Context, node, upid string, timeout time.Duration) (*TaskStatus, error) {
	deadline := time.Now().Add(timeout)
	interval := 800 * time.Millisecond
	if timeout > 2*time.Minute {
		interval = 5 * time.Second
	}
	path := fmt.Sprintf("/nodes/%s/tasks/%s/status", url.PathEscape(node), url.PathEscape(upid))
	for {
		var st TaskStatus
		if err := c.get(ctx, path, &st); err != nil {
			return nil, err
		}
		if st.Status != "running" {
			return &st, nil
		}
		if time.Now().After(deadline) {
			return &st, fmt.Errorf("task still running after %s", timeout)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}
