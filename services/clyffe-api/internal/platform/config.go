package platform

import "os"

// Config holds runtime configuration for clyffe-api. Secrets come from the
// environment (populated from Infisical / /run/shippin-secrets in production).
//
// Boundary: clyffe-api is the customer-facing plane. In production its DBURL
// should use a least-privilege role scoped to the customer-safe schemas
// (shippin_core.tenants, clyffe_*), never the operator infra schemas.
type Config struct {
	Addr     string // listen address, e.g. ":8082"
	DBURL    string // postgres connection string
	LogLevel string
}

// LoadConfig reads configuration from the environment with sane local defaults
// matching the devstation Postgres provisioned for development.
func LoadConfig() Config {
	return Config{
		Addr:     env("CLYFFE_API_ADDR", ":8082"),
		DBURL:    env("CLYFFE_DB_URL", "postgres://shippin:shippin_dev_local@127.0.0.1:5432/shippin_mesh?sslmode=disable"),
		LogLevel: env("CLYFFE_LOG_LEVEL", "info"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
