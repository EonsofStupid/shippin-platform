-- Shippin canonical schema — migration 0007
-- Public IP INVENTORY: the addresses Shippin owns/routes, so the operator
-- can add and update them and assign them to hosts. The existing
-- shippin_infra.ip_migrations (0004) stays as the TRANSITION log between
-- addresses; this table is the inventory those transitions reference.
-- Apply: psql -h 127.0.0.1 -U shippin -d shippin_mesh -f data/schema/sql/0007_public_ips.sql
-- Idempotent.

BEGIN;

DO $$ BEGIN
    CREATE TYPE shippin_infra.public_ip_role AS ENUM
        ('ingress','egress','exit','reserved');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE shippin_infra.public_ip_status AS ENUM
        ('active','reserved','released');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS shippin_infra.public_ips (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    address         inet NOT NULL UNIQUE,
    provider        text,                                  -- who announces it (DC/ISP/cloud)
    host_id         uuid REFERENCES shippin_infra.hosts(id),-- assigned host, if any
    role            shippin_infra.public_ip_role   NOT NULL DEFAULT 'reserved',
    status          shippin_infra.public_ip_status NOT NULL DEFAULT 'reserved',
    label           text,                                  -- human name, e.g. 'homebase-ingress'
    note            text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS public_ips_host_idx   ON shippin_infra.public_ips(host_id);
CREATE INDEX IF NOT EXISTS public_ips_status_idx ON shippin_infra.public_ips(status);

DROP TRIGGER IF EXISTS set_updated_at ON shippin_infra.public_ips;
CREATE TRIGGER set_updated_at BEFORE UPDATE ON shippin_infra.public_ips
  FOR EACH ROW EXECUTE FUNCTION shippin_core.set_updated_at();

INSERT INTO shippin_core.schema_migrations(version) VALUES ('0007_public_ips')
  ON CONFLICT (version) DO NOTHING;

COMMIT;
