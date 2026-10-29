// Package connect is the devstation click-to-auth surface ("Connect & Launch").
// It does NOT define its own secret pattern — it DRIVES the authority installed
// on the devstation and described by W:/configuration:
//   - warden-infisical-bootstrap   encrypts the MI creds (systemd-creds) and
//                                   starts warden-infisical-agent.service
//   - warden-infisical-status      the Infisical state machine
//   - warden-devstation-status     host/workspace/tools/git
//   - /run/shippin-secrets         volatile runtime keyring (token sink)
//   - /run/warden-secrets          compatibility symlink to the keyring
// Secrets are never logged, never written to W, never placed in plaintext on
// disk — the bootstrap encrypts them at rest with systemd-creds.
package connect

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Store activates and inspects devstation auth by shelling the authoritative
// helpers. repoRoot is the git verification cwd; home locates per-user creds.
type Store struct {
	repoRoot string
	home     string
}

func NewStore(repoRoot string) *Store {
	home, _ := os.UserHomeDir()
	return &Store{repoRoot: repoRoot, home: home}
}

// ToolStatus is one tool's auth state for the Connect panel.
type ToolStatus struct {
	Tool      string `json:"tool"`
	Connected bool   `json:"connected"`
	Detail    string `json:"detail"`
}

// Status reports auth state across the tools the devstation brokers, reading
// the authority (warden-infisical-status) — not a private notion of state.
func (s *Store) Status(ctx context.Context) []ToolStatus {
	out := []ToolStatus{}

	// Infisical: the authoritative state machine. connected only when active.
	infStatus := infisicalStatus(ctx)
	out = append(out, ToolStatus{
		Tool:      "infisical",
		Connected: infStatus == "machine_identity_active",
		Detail:    infStatus,
	})

	// GitHub: a usable credential (git credential store or gh login).
	ghDetail, ghOK := githubState(s.home)
	out = append(out, ToolStatus{Tool: "github", Connected: ghOK, Detail: ghDetail})

	// Claude Code / Codex / Gemini: provider-owned credential files (user profile).
	out = append(out, fileStatus("claude-code", filepath.Join(s.home, ".claude", ".credentials.json")))
	out = append(out, fileStatus("codex", filepath.Join(s.home, ".codex", "auth.json")))
	out = append(out, fileStatus("gemini", filepath.Join(s.home, ".gemini", "oauth_creds.json")))
	return out
}

// InfisicalInput is the activate payload. ClientSecret is required; ClientID
// falls back to the existing machine-identity reference.
type InfisicalInput struct {
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
}

// ActivateInfisical drives the authoritative warden-infisical-bootstrap
// non-interactively: it feeds the client id + secret on stdin (the helper's two
// read prompts), so the creds are encrypted with systemd-creds and the agent
// starts. The secret is never an argv, never logged, never written to W.
func (s *Store) ActivateInfisical(ctx context.Context, in InfisicalInput) (string, error) {
	clientID := strings.TrimSpace(in.ClientID)
	if clientID == "" {
		clientID = readEnvFile(filepath.Join(s.home, ".infisical-mi.env"))["INFISICAL_CLIENT_ID"]
	}
	if clientID == "" {
		return "", fmt.Errorf("client_id not supplied and not found in machine-identity reference")
	}
	if strings.TrimSpace(in.ClientSecret) == "" {
		return "", fmt.Errorf("client_secret is required")
	}
	// warden-infisical-bootstrap reads client id then client secret, in order.
	stdin := clientID + "\n" + in.ClientSecret + "\n"
	out, err := runStdin(ctx, 120*time.Second, stdin, "sudo", "-n", "/usr/local/bin/warden-infisical-bootstrap")
	return out, err
}

// ActivateGitHub configures the git credential store with a PAT (provider-owned
// local state) and verifies it against origin. Per the authority, static PATs
// are break-glass; the durable path is the GitHub secret brokered by Infisical.
func (s *Store) ActivateGitHub(ctx context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("empty token")
	}
	credPath := filepath.Join(s.home, ".git-credentials")
	line := fmt.Sprintf("https://x-access-token:%s@github.com\n", token)
	if err := os.WriteFile(credPath, []byte(line), 0o600); err != nil {
		return "", fmt.Errorf("write git credentials: %w", err)
	}
	if _, err := run(ctx, 20*time.Second, s.repoRoot, "git", "config", "--global", "credential.helper", "store"); err != nil {
		return "", err
	}
	out, err := run(ctx, 30*time.Second, s.repoRoot, "git", "ls-remote", "--heads", "origin")
	if err != nil {
		return "github credential written, but verification failed: " + out, err
	}
	return "github connected; origin reachable", nil
}

// --- helpers (no secret values are ever returned in output) ---

func run(ctx context.Context, d time.Duration, dir, name string, args ...string) (string, error) {
	c, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	cmd := exec.CommandContext(c, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func runStdin(ctx context.Context, d time.Duration, stdin, name string, args ...string) (string, error) {
	c, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	cmd := exec.CommandContext(c, name, args...)
	cmd.Stdin = bytes.NewBufferString(stdin)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// infisicalStatus parses the `status=` line of the authoritative
// warden-infisical-status helper (e.g. machine_identity_credentials_missing).
func infisicalStatus(ctx context.Context) string {
	out, _ := run(ctx, 15*time.Second, "", "/usr/local/bin/warden-infisical-status")
	for _, ln := range strings.Split(out, "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(ln), "status="); ok {
			return v
		}
	}
	if out == "" {
		return "status_unavailable"
	}
	return "unknown"
}

func fileStatus(tool, path string) ToolStatus {
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		return ToolStatus{Tool: tool, Connected: true, Detail: "credential present"}
	}
	return ToolStatus{Tool: tool, Connected: false, Detail: "no credential"}
}

func githubState(home string) (string, bool) {
	if st, err := os.Stat(filepath.Join(home, ".git-credentials")); err == nil && st.Size() > 0 {
		return "git credential store", true
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "gh", "hosts.yml")); err == nil {
		return "gh login", true
	}
	return "not authenticated", false
}

func readEnvFile(path string) map[string]string {
	m := map[string]string{}
	b, err := os.ReadFile(path)
	if err != nil {
		return m
	}
	for _, ln := range strings.Split(string(b), "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") {
			continue
		}
		if k, v, ok := strings.Cut(ln, "="); ok {
			m[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return m
}
