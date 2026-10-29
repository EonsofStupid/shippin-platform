package platform

import (
	"os"
)

// Config holds runtime configuration for shippin-api. Secrets come from the
// environment (populated from Infisical / /run/shippin-secrets in production).
type Config struct {
	Addr     string // listen address, e.g. ":8081"
	DBURL    string // postgres connection string
	LogLevel string

	// Operator bootstrap credential for the identity context. Dev defaults;
	// in production these come from Infisical and OIDC supersedes them.
	OperatorUser string
	OperatorPass string

	// RepoRoot locates repo-authored sources (registry, validator) for mesh.
	RepoRoot string

	// SyncPlanPath is the intelligence-sync projection plan on W (read by the
	// mesh projection endpoint). SyncBin is the deployed intelligence-sync
	// binary the sync-run endpoint executes — both default to the W layout.
	SyncPlanPath string
	SyncBin      string
}

// LoadConfig reads configuration from the environment with sane local defaults
// matching the devstation Postgres provisioned for development.
func LoadConfig() Config {
	return Config{
		Addr:         env("SHIPPIN_API_ADDR", ":8081"),
		DBURL:        env("SHIPPIN_DB_URL", "postgres://shippin:shippin_dev_local@127.0.0.1:5432/shippin_mesh?sslmode=disable"),
		LogLevel:     env("SHIPPIN_LOG_LEVEL", "info"),
		OperatorUser: env("SHIPPIN_OPERATOR_USER", "operator"),
		OperatorPass: env("SHIPPIN_OPERATOR_PASS", "warden-dev"),
		RepoRoot:     env("SHIPPIN_REPO_ROOT", "/workspace/warden-storage/projects/shippin-platform"),
		SyncPlanPath: env("SYNC_PLAN_PATH", "/workspace/warden-storage/registry/projection-plan.json"),
		SyncBin:      env("SHIPPIN_SYNC_BIN", "/workspace/warden-storage/plugins/bin/intelligence-sync"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
