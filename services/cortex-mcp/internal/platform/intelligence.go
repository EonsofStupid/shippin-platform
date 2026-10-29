package platform

import (
	"os"
	"path/filepath"
	"strings"
)

// IntelligenceConfig holds the self-hosted intelligence-plane endpoints.
// Defaults follow the captured foundation map (Qdrant LXC 106, SurrealDB LXC
// 104 — the self-hosted target of the Surreal-Cloud migration). Credentials
// are NEVER stored here or in any plugin package: Secret() resolves refs from
// the environment or the Infisical-materialized /run/shippin-secrets tmpfs.
type IntelligenceConfig struct {
	QdrantURL  string
	SurrealURL string
	SurrealNS  string
	SurrealDB  string
}

// LoadIntelligenceConfig reads endpoints from the environment with
// foundation-map defaults.
func LoadIntelligenceConfig() IntelligenceConfig {
	return IntelligenceConfig{
		QdrantURL:  envOr("CORTEX_QDRANT_URL", "http://10.0.0.106:6333"),
		SurrealURL: envOr("CORTEX_SURREAL_URL", "http://10.0.0.104:8000"),
		SurrealNS:  envOr("CORTEX_SURREAL_NS", "clyffy"),
		SurrealDB:  envOr("CORTEX_SURREAL_DB", "intelligence"),
	}
}

// Secret resolves a named credential: env var first, then the brokered
// runtime file /run/warden-secrets/<lowercase-name>. Returns "" when absent;
// callers degrade to unauthenticated requests and surface the gate, never the
// value.
func Secret(name string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	for _, dir := range []string{"/run/shippin-secrets", "/run/warden-secrets"} {
		b, err := os.ReadFile(filepath.Join(dir, strings.ToLower(name)))
		if err == nil && len(b) > 0 {
			return strings.TrimSpace(string(b))
		}
	}
	return ""
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
