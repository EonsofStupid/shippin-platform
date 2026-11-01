-- Customer portal access is a read-only group role. Operators create a
-- separate login and grant it this role; no credentials are stored here.
-- Existing accounts and memberships are never deleted or changed.
BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'shippin_customer_api') THEN
        CREATE ROLE shippin_customer_api NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'shippin_customer_api'
               AND (rolcanlogin OR rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls)) THEN
        RAISE EXCEPTION 'shippin_customer_api must be a restricted non-login role';
    END IF;
END $$;

-- PostgreSQL functions otherwise default to PUBLIC execution. Customer reads
-- must not gain access to identity administration or password verification.
REVOKE EXECUTE ON ALL FUNCTIONS IN SCHEMA identity FROM PUBLIC;

GRANT USAGE ON SCHEMA identity, shippin_core, clyffe_core, app TO shippin_customer_api;
GRANT SELECT (token_hash, subject_id, expires_at) ON identity.sessions TO shippin_customer_api;
GRANT SELECT (id, kind, status, is_active, email_verified_at) ON shippin_core.subjects TO shippin_customer_api;
GRANT SELECT (subject_id, tenant_id) ON shippin_core.subject_tenant TO shippin_customer_api;
GRANT SELECT (id, slug, name, created_at) ON shippin_core.tenants TO shippin_customer_api;
GRANT SELECT (id, tenant_id, product, tier, repo, region, status, created_at)
    ON clyffe_core.orders TO shippin_customer_api;
GRANT EXECUTE ON FUNCTION app.current_tenant() TO shippin_customer_api;

INSERT INTO shippin_core.schema_migrations(version) VALUES ('0009_customer_access')
    ON CONFLICT (version) DO NOTHING;
COMMIT;
