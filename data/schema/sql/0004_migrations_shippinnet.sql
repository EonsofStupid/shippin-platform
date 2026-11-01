-- Shippin canonical schema — migration 0004
-- Make the public IP migration (104 -> 204) first-class + auditable, and add the
-- WardenNet (Headscale) platform. Spec: docs/WARDEN_DEVSTATION_TURNKEY_WARDENNET_PLAN.md
-- Apply: psql -h 127.0.0.1 -U shippin -d shippin_mesh -f data/schema/sql/0004_migrations_shippinnet.sql
-- Idempotent.

BEGIN;

DO $$ BEGIN
    CREATE TYPE shippin_infra.migration_state AS ENUM
        ('planned','dual_homed','cutover','draining','complete','rolled_back');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS shippin_infra.ip_migrations (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope       text NOT NULL,                 -- 'public-ingress' | host stable_id
    from_ip     text NOT NULL,
    to_ip       text NOT NULL,
    state       shippin_infra.migration_state NOT NULL DEFAULT 'planned',
    started_at  timestamptz,
    cutover_at  timestamptz,
    note        text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

DROP TRIGGER IF EXISTS set_updated_at ON shippin_infra.ip_migrations;
CREATE TRIGGER set_updated_at BEFORE UPDATE ON shippin_infra.ip_migrations
  FOR EACH ROW EXECUTE FUNCTION shippin_core.set_updated_at();

INSERT INTO shippin_core.schema_migrations(version) VALUES ('0004_migrations_shippinnet')
  ON CONFLICT (version) DO NOTHING;

COMMIT;
