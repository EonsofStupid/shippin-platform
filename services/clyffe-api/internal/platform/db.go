package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a pgx connection pool and verifies connectivity.
func NewPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("CLYFFE_DB_URL is required")
	}
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}

// ValidateCustomerRole rejects owner, elevated or unexpected database roles.
// It does not replace query-level authorization or audit arbitrary direct grants.
func ValidateCustomerRole(ctx context.Context, pool *pgxpool.Pool) error {
	var safe bool
	err := pool.QueryRow(ctx, `
		SELECT current_user = session_user
		  AND COALESCE(pg_has_role(current_user, to_regrole('shippin_customer_api'), 'USAGE'), false)
		  AND NOT EXISTS (
		    SELECT 1 FROM pg_roles r
		    WHERE pg_has_role(current_user, r.oid, 'MEMBER')
		      AND (r.rolsuper OR r.rolbypassrls OR r.rolcreaterole OR r.rolcreatedb OR r.rolreplication
		           OR (r.rolname <> current_user AND r.rolname <> 'shippin_customer_api'))
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		    WHERE (n.nspname, c.relname) IN (('identity','sessions'), ('shippin_core','subjects'),
		      ('shippin_core','subject_tenant'), ('shippin_core','tenants'), ('clyffe_core','orders'))
		      AND pg_has_role(current_user, c.relowner, 'MEMBER')
		  )
		  AND COALESCE((SELECT c.relrowsecurity FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		    WHERE n.nspname = 'clyffe_core' AND c.relname = 'orders'), false)
		  AND NOT COALESCE(has_any_column_privilege(current_user, to_regclass('identity.credentials'), 'SELECT'), true)
		  AND NOT EXISTS (
		    SELECT 1 FROM pg_proc p JOIN pg_namespace n ON n.oid = p.pronamespace
		    WHERE n.nspname = 'identity' AND has_function_privilege(current_user, p.oid, 'EXECUTE')
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		    WHERE (n.nspname, c.relname) IN (('identity','sessions'), ('shippin_core','subjects'),
		      ('shippin_core','subject_tenant'), ('shippin_core','tenants'), ('clyffe_core','orders'))
		      AND (has_table_privilege(current_user, c.oid, 'INSERT') OR has_any_column_privilege(current_user, c.oid, 'INSERT')
		        OR has_table_privilege(current_user, c.oid, 'UPDATE') OR has_any_column_privilege(current_user, c.oid, 'UPDATE')
		        OR has_table_privilege(current_user, c.oid, 'DELETE') OR has_table_privilege(current_user, c.oid, 'TRUNCATE'))
		  )`).Scan(&safe)
	if err != nil {
		return fmt.Errorf("validate customer database role: %w", err)
	}
	if !safe {
		return fmt.Errorf("customer database role is missing, elevated, owns data, or lacks row security")
	}
	return nil
}
