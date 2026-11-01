-- Shippin canonical schema — migration 0006 (identity & tenancy)
-- Spec: docs/specs/IDENTITY_TENANCY_SPEC.md ("the hades list").
-- EXTENDS the existing shippin_core.subjects from 0001 (2nd-pass find) — does
-- not green-field. Requires PostgreSQL >= 18 (native uuidv7()).
-- Apply: warden-migrate up        (or psql -f, idempotent)

BEGIN;

CREATE EXTENSION IF NOT EXISTS citext;

-- ---------------------------------------------------------------------------
-- subjects: add verification + status; email -> citext, unique where present
-- ---------------------------------------------------------------------------
DO $$ BEGIN
    CREATE TYPE shippin_core.subject_status AS ENUM ('pending','active','suspended');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

ALTER TABLE shippin_core.subjects
    ADD COLUMN IF NOT EXISTS email_verified_at timestamptz,
    ADD COLUMN IF NOT EXISTS status shippin_core.subject_status NOT NULL DEFAULT 'pending';
ALTER TABLE shippin_core.subjects
    ALTER COLUMN email TYPE citext USING email::citext,
    ALTER COLUMN id SET DEFAULT uuidv7();
CREATE UNIQUE INDEX IF NOT EXISTS subjects_email_unique
    ON shippin_core.subjects(email) WHERE email IS NOT NULL;

-- ---------------------------------------------------------------------------
-- subject_tenant: roles as rows (dual-role operator: customer + super_admin)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS shippin_core.subject_tenant (
    subject_id  uuid NOT NULL REFERENCES shippin_core.subjects(id),
    tenant_id   uuid NOT NULL REFERENCES shippin_core.tenants(id),
    role        text NOT NULL,           -- customer | super_admin | viewer | ...
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (subject_id, tenant_id, role)
);

-- ---------------------------------------------------------------------------
-- identity schema: email verification (hashed single-use tokens)
-- ---------------------------------------------------------------------------
CREATE SCHEMA IF NOT EXISTS identity;

CREATE TABLE IF NOT EXISTS identity.email_verification (
    id          uuid PRIMARY KEY DEFAULT uuidv7(),
    subject_id  uuid NOT NULL REFERENCES shippin_core.subjects(id),
    token_hash  bytea NOT NULL,          -- sha256(token); raw token returned once
    expires_at  timestamptz NOT NULL,
    used_at     timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS email_verification_subject_idx
    ON identity.email_verification(subject_id);

-- identity.issue_email_verification: raw token returned ONCE; hash stored.
CREATE OR REPLACE FUNCTION identity.issue_email_verification(p_subject uuid)
RETURNS text LANGUAGE plpgsql AS $$
DECLARE v_token text;
BEGIN
    v_token := encode(gen_random_bytes(32), 'hex');
    INSERT INTO identity.email_verification (subject_id, token_hash, expires_at)
    VALUES (p_subject, digest(v_token, 'sha256'), now() + interval '24 hours');
    RETURN v_token;
END $$;

-- identity.verify_email: match by hash, single-use, unexpired -> activate.
CREATE OR REPLACE FUNCTION identity.verify_email(p_token text)
RETURNS boolean LANGUAGE plpgsql AS $$
DECLARE v_row identity.email_verification%ROWTYPE;
BEGIN
    SELECT * INTO v_row FROM identity.email_verification
    WHERE token_hash = digest(p_token, 'sha256')
      AND used_at IS NULL AND expires_at > now()
    LIMIT 1;
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE identity.email_verification SET used_at = now() WHERE id = v_row.id;
    UPDATE shippin_core.subjects
       SET email_verified_at = now(), status = 'active'
     WHERE id = v_row.subject_id;
    RETURN true;
END $$;

-- email change forces re-verification.
CREATE OR REPLACE FUNCTION identity.on_email_change_reverify()
RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.email IS DISTINCT FROM OLD.email THEN
        NEW.email_verified_at := NULL;
        NEW.status := 'pending';
    END IF;
    RETURN NEW;
END $$;
DROP TRIGGER IF EXISTS trg_email_reverify ON shippin_core.subjects;
CREATE TRIGGER trg_email_reverify BEFORE UPDATE ON shippin_core.subjects
    FOR EACH ROW EXECUTE FUNCTION identity.on_email_change_reverify();

-- identity.create_customer: the automated tenant-with-UUID creation.
CREATE OR REPLACE FUNCTION identity.create_customer(p_email citext, p_display_name text)
RETURNS TABLE (subject_id uuid, tenant_id uuid, verification_token text)
LANGUAGE plpgsql SECURITY DEFINER AS $$
DECLARE v_subject uuid; v_tenant uuid; v_token text; v_slug text;
BEGIN
    v_slug := lower(regexp_replace(p_display_name, '[^a-zA-Z0-9]+', '-', 'g'));
    INSERT INTO shippin_core.subjects (kind, display_name, email, status)
    VALUES ('customer', p_display_name, p_email, 'pending')
    RETURNING id INTO v_subject;
    INSERT INTO shippin_core.tenants (slug, name)
    VALUES (v_slug, p_display_name)
    RETURNING id INTO v_tenant;
    INSERT INTO shippin_core.subject_tenant (subject_id, tenant_id, role)
    VALUES (v_subject, v_tenant, 'customer');
    v_token := identity.issue_email_verification(v_subject);
    INSERT INTO shippin_audit.events (actor_id, tenant_id, action, target_kind, target_id, data)
    VALUES (v_subject, v_tenant, 'identity.customer_created', 'subject', v_subject::text, '{}'::jsonb);
    RETURN QUERY SELECT v_subject, v_tenant, v_token;
END $$;

-- ---------------------------------------------------------------------------
-- app context + RLS (tenant-scoped tables); enforcement bites via shippin_app
-- (non-owner role). Table owner stays unaffected until services adopt it.
-- ---------------------------------------------------------------------------
CREATE SCHEMA IF NOT EXISTS app;

CREATE OR REPLACE FUNCTION app.current_subject() RETURNS uuid
LANGUAGE sql STABLE AS
$$ SELECT nullif(current_setting('app.current_subject', true), '')::uuid $$;

CREATE OR REPLACE FUNCTION app.current_tenant() RETURNS uuid
LANGUAGE sql STABLE AS
$$ SELECT nullif(current_setting('app.current_tenant', true), '')::uuid $$;

ALTER TABLE shippin_infra.resources ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON shippin_infra.resources;
CREATE POLICY tenant_isolation ON shippin_infra.resources
    USING (tenant_id = app.current_tenant());

ALTER TABLE clyffe_core.orders ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON clyffe_core.orders;
CREATE POLICY tenant_isolation ON clyffe_core.orders
    USING (tenant_id = app.current_tenant());

-- Role creation needs CREATEROLE, which the app role deliberately lacks.
-- Provision `shippin_app` out-of-band (operator/provisioner):
--   CREATE ROLE shippin_app NOLOGIN;
-- This block is tolerant: it creates+grants when privileged, otherwise it
-- raises a NOTICE and the grants are applied on the next privileged run.
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'shippin_app') THEN
        BEGIN
            CREATE ROLE shippin_app NOLOGIN;
        EXCEPTION WHEN insufficient_privilege THEN
            RAISE NOTICE 'shippin_app not created (needs CREATEROLE); provision it out-of-band';
        END;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'shippin_app') THEN
        GRANT USAGE ON SCHEMA shippin_core, shippin_infra, shippin_audit, clyffe_core, identity, app TO shippin_app;
        GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA shippin_core, shippin_infra, clyffe_core TO shippin_app;
        GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA shippin_audit TO shippin_app;
        GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA identity, app TO shippin_app;
    END IF;
END $$;

INSERT INTO shippin_core.schema_migrations(version) VALUES ('0006_identity')
  ON CONFLICT (version) DO NOTHING;

COMMIT;
