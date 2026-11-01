// Package proxmox is the Warden client for Proxmox VE (substrate).
// Proxmox powers VMs/LXCs; this package is how Warden talks to that OS.
package proxmox

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
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
	CAFile    string
}

// LoadConfig reads PROXMOX_* env vars. Also loads a dotenv-style file if set:
// SHIPPIN_PROXMOX_ENV, or secrets/proxmox.env under SHIPPIN_REPO_ROOT when present.
// Never logs file contents.
func LoadConfig() Config {
	loadProxmoxEnvFile()
	port := 8006
	if p := os.Getenv("PROXMOX_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			port = 0
		} else {
			port = n
		}
	}
	verify := true
	if setting := strings.ToLower(strings.TrimSpace(os.Getenv("PROXMOX_VERIFY_TLS"))); setting != "" {
		verify = setting == "1" || setting == "true" || setting == "yes"
	}
	return Config{
		Host:      os.Getenv("PROXMOX_HOST"),
		Port:      port,
		Node:      os.Getenv("PROXMOX_NODE"),
		TokenID:   os.Getenv("PROXMOX_TOKEN_ID"),
		TokenSec:  os.Getenv("PROXMOX_TOKEN_SECRET"),
		VerifyTLS: verify,
		CAFile:    os.Getenv("PROXMOX_CA_CERT"),
	}
}

func loadProxmoxEnvFile() {
	candidates := []string{}
	if p := os.Getenv("SHIPPIN_PROXMOX_ENV"); p != "" {
		candidates = append(candidates, p)
	}
	root := os.Getenv("SHIPPIN_REPO_ROOT")
	if root != "" {
		candidates = append(candidates, root+"/secrets/proxmox.env")
	}
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

// Configured reports whether credentials are present (not whether API works).
func (c Config) Configured() bool {
	return c.Host != "" && c.TokenID != "" && c.TokenSec != ""
}

// Client talks to Proxmox REST API.
type Client struct {
	cfg  Config
	http *http.Client
	base string
	auth string
}

func NewClient(cfg Config) (*Client, error) {
	if !cfg.VerifyTLS {
		return nil, fmt.Errorf("proxmox: TLS verification must be enabled; configure PROXMOX_CA_CERT for a private CA")
	}
	if cfg.Host == "" || strings.ContainsAny(cfg.Host, "/\\?#@ \t\r\n") ||
		(strings.Contains(cfg.Host, ":") && net.ParseIP(cfg.Host) == nil) || cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("proxmox: invalid host or port")
	}
	var roots *x509.CertPool
	if cfg.CAFile != "" {
		pem, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("proxmox: cannot read CA certificate: %w", err)
		}
		roots, err = x509.SystemCertPool()
		if err != nil || roots == nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("proxmox: CA certificate file contains no valid certificates")
		}
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots},
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 30 * time.Second, Transport: tr},
		base: "https://" + net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)) + "/api2/json",
		auth: fmt.Sprintf("PVEAPIToken=%s=%s", cfg.TokenID, cfg.TokenSec),
	}, nil
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
	Node string `json:"node"`
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
	Status     string `json:"status"` // running | stopped
	ExitStatus string `json:"exitstatus,omitempty"`
	UPID       string `json:"upid,omitempty"`
	Type       string `json:"type,omitempty"`
	User       string `json:"user,omitempty"`
}

// WaitTask polls task until not running or timeout.
func (c *Client) WaitTask(ctx context.Context, node, upid string, timeout time.Duration) (*TaskStatus, error) {
	deadline := time.Now().Add(timeout)
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
		case <-time.After(800 * time.Millisecond):
		}
	}
}
