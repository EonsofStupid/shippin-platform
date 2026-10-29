---
clyffy_touchpoint:
  version: 2
  workspace_id: wardenclyffe.infra
  project_key: identity-tenancy-spec
  persona: clyffy-operator
  kind: doc
  owner: docs/specs/IDENTITY_TENANCY_SPEC.md
  module: module-01-warden
  mesh_registry: wardenclyffe/registry/context-mesh.yaml
  sync:
    qdrant: true
    surreal: true
---

# Identity & Tenancy Spec (the "hades" list)

**Supersession (2026-08-29):** The IdP is **Nomen** (`projects/tessera`).
Do not implement a second identity authority here. Workspace tenancy
tables may still be Warden/Clyffe product facts; login and seats are
Nomen. See `../shippin/docs/platform/PRIVATE-CLOUD-COMMUNITY.md`.

Build spec for making a customer real. Production-only baseline. Verified live:
**PostgreSQL 18.4**, native `uuidv7()`, `pgcrypto`. Owns: warden `identity`
context + clyffe `account`. Status today: **nothing exists yet** — this is the
list to PR into existence.

## Ownership chain (target)

```
Warden (control plane) → provisions/manages → devstation (service)
  → attaches → W-drive volume (per-service, managed)   → owned by → customer "hades"
```

## Identity model (locked decisions)

- **PK = UUIDv7** on every entity (`default uuidv7()`, PG18 native — time-ordered,
  index-friendly). No app-side IDs.
- **username == email**, stored `citext` (case-insensitive) `UNIQUE`.
- **display_name** = separate text, the human label shown in UI.
- **email second-verification required** before active (token flow below).
- **email hidden by default** in UI — masked, reveal on eye-click + blur (a
  coldlight `RevealableEmail` component).

## Tables (DDL sketch)

```sql
-- a human identity (operator or customer). hades is a row here.
shippin_core.subjects (
  id            uuid primary key default uuidv7(),
  email         citext unique not null,         -- == username
  display_name  text   not null,
  email_verified_at timestamptz,
  status        text not null default 'pending', -- pending|active|suspended
  created_at    timestamptz not null default now(),
  updated_at    timestamptz not null default now()
)
shippin_core.tenants ( id uuid pk default uuidv7(), slug citext unique, name text, ... )  -- exists
shippin_core.subject_tenant ( subject_id uuid, tenant_id uuid, role text, primary key (subject_id, tenant_id) )
identity.email_verification (
  id uuid pk default uuidv7(), subject_id uuid not null,
  token_hash bytea not null,        -- sha256(token); raw token returned once
  expires_at timestamptz not null, used_at timestamptz,
  created_at timestamptz not null default now()
)
```

## Functions (the full list)

1. `core.set_updated_at()` — BEFORE UPDATE: `NEW.updated_at = now()`.
2. `core.touch_audit()` — AFTER INSERT/UPDATE/DELETE → append row to
   `shippin_audit.events` (actor, action, entity, before/after jsonb).
3. `identity.issue_email_verification(subject)` — token = `encode(gen_random_bytes(32),'hex')`
   returned ONCE; store `digest(token,'sha256')`, `expires_at = now()+'24h'`, single-use.
4. `identity.verify_email(token)` — match by hash, check unused+unexpired →
   set `subjects.email_verified_at = now()`, `status='active'`, mark token used.
5. `identity.on_email_change_reverify()` — BEFORE UPDATE: if email changed set
   `email_verified_at = null, status='pending'`; AFTER: re-issue verification.
6. `identity.create_customer(p_email, p_display_name)` — SECURITY DEFINER:
   insert subject (uuidv7) + tenant + ownership + issue verification + audit;
   returns `(subject_id, tenant_id, verification_token)`. **This is the automated
   tenant-with-UUID creation.**
7. `app.current_subject()` / `app.current_tenant()` — read `current_setting('app.*',true)::uuid`.

## Triggers (bindings)

- `trg_set_updated_at` BEFORE UPDATE on every mutable table.
- `trg_audit` AFTER INSERT/UPDATE/DELETE on subjects, tenants, services, volumes.
- `trg_email_reverify` BEFORE+AFTER UPDATE on subjects.
- **RLS**: `ENABLE ROW LEVEL SECURITY` on tenant-scoped tables; policy
  `USING (tenant_id = app.current_tenant())`.

## Email verification flow

1. `create_customer` → returns one-time token (emailed; never stored raw).
2. Customer clicks link → `verify_email(token)` → `status='active'`.
3. Email change → reverify required before re-activation.

Real-world references captured: Supabase auth (confirmation tokens, updated_at
triggers, RLS), Authentik (identity/flows). Methods are PG18-native (uuidv7,
pgcrypto, citext, RLS) — no extensions beyond pgcrypto/citext.

## PR plan

`identity` context PR: schema migration (tables + functions + triggers + RLS) →
Go `internal/identity` create/verify endpoints → React `RevealableEmail` +
account view. Then bind ownership (hades → devstation → volume).
