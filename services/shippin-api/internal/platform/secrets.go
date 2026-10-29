package platform

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	// CanonicalSecretsDir is the Shippin volatile runtime keyring.
	// WardenClyffe is an internal layer name only; do not use it in runtime paths.
	CanonicalSecretsDir     = "/run/shippin-secrets"
	CompatibilitySecretsDir = "/run/warden-secrets"
)

// SecretsDir returns the writable runtime keyring. SHIPPIN_SECRETS_DIR wins
// (tests). Prefer /run/shippin-secrets; fall back to the warden-secrets
// compatibility path when that is all that exists.
func SecretsDir() string {
	if d := strings.TrimSpace(os.Getenv("SHIPPIN_SECRETS_DIR")); d != "" {
		return d
	}
	if isDir(CanonicalSecretsDir) {
		return CanonicalSecretsDir
	}
	if isDir(CompatibilitySecretsDir) {
		return CompatibilitySecretsDir
	}
	return CanonicalSecretsDir
}

// ReadSecret resolves a named credential: env var first, then
// $SecretsDir/<lowercase-name>. Returns "" when absent. Never logs the value.
func ReadSecret(name string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	for _, dir := range []string{SecretsDir(), CanonicalSecretsDir, CompatibilitySecretsDir} {
		b, err := os.ReadFile(filepath.Join(dir, strings.ToLower(name)))
		if err == nil && len(b) > 0 {
			return strings.TrimSpace(string(b))
		}
	}
	return ""
}

// WriteFile0600 writes p with mode 0600, creating the parent at 0700.
func WriteFile0600(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o600)
}

func isDir(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}
