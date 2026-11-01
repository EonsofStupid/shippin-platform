package platform

import "os"

// Config holds runtime configuration for clyffe-api. Secrets come from the
// environment (populated from Infisical / /run/warden-secrets in production).
//
// Boundary: clyffe-api is the customer-facing plane. In production its DBURL
// should use a least-privilege role scoped to the customer-safe schemas
// (shippin_core.tenants, clyffe_*), never the operator infra schemas.
type Config struct {
	Addr       string // listen address, e.g. ":8082"
	DBURL      string // postgres connection string
	LogLevel   string
	CookieName string
}

// LoadConfig reads configuration from the environment with sane local defaults
// matching the devstation Postgres provisioned for development.
func LoadConfig() Config {
	return Config{
		Addr:       env("CLYFFE_API_ADDR", ":8082"),
		DBURL:      os.Getenv("CLYFFE_DB_URL"),
		LogLevel:   env("CLYFFE_LOG_LEVEL", "info"),
		CookieName: env("SHIPPIN_AUTH_COOKIE", "shippin_session"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
