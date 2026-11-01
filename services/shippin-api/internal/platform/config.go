package platform

import (
	"fmt"
	"os"
	"strings"
)

// Config holds runtime configuration for shippin-api. Secrets come from the
// environment (populated from Infisical / /run/warden-secrets in production).
type Config struct {
	Addr     string // listen address, e.g. ":8081"
	DBURL    string // postgres connection string
	LogLevel string

	// Explicit bootstrap credential for the existing development identity context.
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
		OperatorUser: os.Getenv("SHIPPIN_OPERATOR_USER"),
		OperatorPass: os.Getenv("SHIPPIN_OPERATOR_PASS"),
		RepoRoot:     env("SHIPPIN_REPO_ROOT", "/workspace/WardenClyffe-latest"),
		SyncPlanPath: env("SYNC_PLAN_PATH", "/workspace/warden-storage/registry/projection-plan.json"),
		SyncBin:      env("SHIPPIN_SYNC_BIN", "/workspace/warden-storage/plugins/bin/intelligence-sync"),
	}
}

// ValidateOperatorCredentials rejects an unconfigured development operator.
// This is not a production first-owner enrollment or password policy.
func (c Config) ValidateOperatorCredentials() error {
	if strings.TrimSpace(c.OperatorUser) == "" || strings.TrimSpace(c.OperatorPass) == "" {
		return fmt.Errorf("SHIPPIN_OPERATOR_USER and SHIPPIN_OPERATOR_PASS must be configured")
	}
	return nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
