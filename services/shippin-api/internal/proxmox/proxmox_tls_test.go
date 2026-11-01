package proxmox

import (
	"context"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
)

func TestClientRequiresTrustedTLS(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/api2/json/version" || r.Header.Get("Authorization") != "PVEAPIToken=test@pve!token=test-secret" {
			t.Errorf("unexpected request path or authorization")
		}
		fmt.Fprint(w, `{"data":{"version":"test"}}`)
	}))
	defer server.Close()
	host, portText, err := net.SplitHostPort(server.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	cfg := Config{Host: host, Port: port, TokenID: "test@pve!token", TokenSec: "test-secret", VerifyTLS: true}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Version(context.Background()); err == nil {
		t.Fatal("untrusted server certificate was accepted")
	}
	if requests.Load() != 0 {
		t.Fatal("untrusted endpoint received a request")
	}

	caFile := filepath.Join(t.TempDir(), "ca.pem")
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := os.WriteFile(caFile, cert, 0600); err != nil {
		t.Fatal(err)
	}
	cfg.CAFile = caFile
	client, err = NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	version, err := client.Version(context.Background())
	if err != nil || version["version"] != "test" || requests.Load() != 1 {
		t.Fatalf("trusted request failed: version=%v requests=%d err=%v", version, requests.Load(), err)
	}
	wrongHost := cfg
	wrongHost.Host = "wrong.invalid"
	client, err = NewClient(wrongHost)
	if err != nil {
		t.Fatal(err)
	}
	client.http.Transport.(*http.Transport).DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	if _, err := client.Version(context.Background()); err == nil || requests.Load() != 1 {
		t.Fatal("hostname mismatch must fail before a request reaches the server")
	}

	cfg.VerifyTLS = false
	if _, err := NewClient(cfg); err == nil {
		t.Fatal("disabling verification must fail")
	}
	cfg.VerifyTLS = true
	cfg.CAFile = filepath.Join(t.TempDir(), "missing.pem")
	if _, err := NewClient(cfg); err == nil {
		t.Fatal("missing CA must fail")
	}
	badCA := filepath.Join(t.TempDir(), "invalid.pem")
	if err := os.WriteFile(badCA, []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg.CAFile = badCA
	if _, err := NewClient(cfg); err == nil {
		t.Fatal("invalid CA must fail")
	}
}

func TestConfigDoesNotAssumeHostOrDisableTLS(t *testing.T) {
	t.Setenv("SHIPPIN_REPO_ROOT", t.TempDir())
	for _, key := range []string{"SHIPPIN_PROXMOX_ENV", "PROXMOX_HOST", "PROXMOX_NODE", "PROXMOX_PORT", "PROXMOX_TOKEN_ID", "PROXMOX_TOKEN_SECRET", "PROXMOX_VERIFY_TLS", "PROXMOX_CA_CERT"} {
		t.Setenv(key, "")
	}
	cfg := LoadConfig()
	if cfg.Host != "" || cfg.Node != "" || !cfg.VerifyTLS || cfg.Configured() {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	t.Setenv("PROXMOX_HOST", "example.test")
	t.Setenv("PROXMOX_TOKEN_ID", "test-token")
	t.Setenv("PROXMOX_TOKEN_SECRET", "test-secret")
	t.Setenv("PROXMOX_VERIFY_TLS", "false")
	service := NewService(LoadConfig(), nil, nil)
	status := service.Status(context.Background())
	if !status.Configured || status.Reachable || status.Error == "" {
		t.Fatalf("invalid configuration must remain unreachable: %+v", status)
	}
	if _, err := service.StartGuest(context.Background(), "node", "qemu", 101); err == nil {
		t.Fatal("invalid client configuration must fail without panic")
	}
}
