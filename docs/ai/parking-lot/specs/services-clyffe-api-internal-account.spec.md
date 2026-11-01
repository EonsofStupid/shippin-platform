---
clyffy_touchpoint:
  version: 2
  workspace_id: wardenclyffe.intelligence
  project_key: services-clyffe-api-internal-account
  persona: clyffy-operator
  kind: buildspec
  owner: docs/ai/parking-lot/specs/services-clyffe-api-internal-account.spec.md
  module: null
  focus_feature: "Customer tenant isolation"
  boundary: services/clyffe-api/internal/account
  source_skill: docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md
  mesh_registry: wardenclyffe/registry/context-mesh.yaml
  sync:
    qdrant: true
    surreal: true
---

# Build Spec — services-clyffe-api-internal-account

Focus: **Customer tenant isolation**. Boundary: `services/clyffe-api/internal/account`. Emitted 2026-11-01 from agenda
`docs/ai/parking-lot/agenda/services-clyffe-api-internal-account.md` (3 locked, 0 parked).
Contract: `docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md` §Build Spec Contract.

## 1. Overview
The original customer handlers listed all tenants and orders without verifying
an identity. They were disabled until the session, membership and restricted-role
checks described here could be implemented and tested.

**Non-goals:** A new login provider, complete customer onboarding, managed
projects, or a production-ready release. Existing sessions remain owned by
`shippin-auth`; these account reads do not establish authorization everywhere.

## 2. Golden patterns
Obeys: `docs/WARDENCLYFFE_NAMING_CONVENTIONS.md`,
`docs/HYPERMODULAR_DDD_FOLDER_STRUCTURE.md`,
`docs/ai/WARDENCLYFFE_STRUCTURE_STANDARD.md`. Reuse the existing account handler;
the boundary review found that context already exists. Leave health separate.

## 3. Locked decisions
- **[feature] fail closed customer account endpoints** (locked 2026-11-01): Return service unavailable for the existing customer home, accounts and orders routes until verified customer identity and tenant authorization exist; keep health available.
- **[feature] tenant scoped customer access**: Resolve the existing session,
  verify the customer and membership, and scope each order read to one tenant.
- **[feature] restricted customer data role**: Reject a privileged or
  table-owning customer DB login and use transaction-local row-security context.

Agenda items locked:
- **D1** (`services/clyffe-api/internal/account`): Deny customer account endpoints until verified identity and tenant scoped access exist
- **D2** (`services/clyffe-api/internal/account`): Reuse verified sessions and scope account and order reads to membership
- **D3** (`services/clyffe-api/internal/account`): Require a restricted non-owner customer database role

## 4. Minimal requirements (10x foresight)
Per item: the smallest thing that ships *complete*, plus what to configure now
because it is 10x costlier to retrofit.
- **D1** (`services/clyffe-api/internal/account`): Deny customer account endpoints until verified identity and tenant scoped access exist

All three registered customer routes must deny requests without a valid cookie.
Caller-supplied identity headers and tenant IDs are never proof of membership.
The API must fail startup on an unsafe DB role and fail closed on DB errors.

## 5. Established patterns / formulas
`existing session cookie -> verified principal -> authorized tenant -> scoped
read-only transaction with transaction-local RLS context -> customer response`.

## 6. File-creation guidance (notes vs. code)
Which deliverables are **touchpoints/Markdown** (routing, decisions, runbooks)
and which are **code** (boundary, language, colocated companions). What must
never be created (defer duplicates to boundary-guard).
Code and tests remain in `services/clyffe-api/internal/account/`; the existing
identity tables own sessions and membership, and migration 0009 owns the
restricted database role. The release roadmap records remaining limitations.
This touchpoint is not an identity store or API contract.

## 7. LLM dynamic-flow within control-layer boundaries
How Clyffy and the coding agents operate against this spec at runtime: the
designation rules (connector vs plugin, ai-identity), what they may read, and
what must never enter an agent's context (the brokered secret boundary).
Agents can review the source and extend the tenant-aware implementation. This
change creates no runtime agent, connector or secrets interface. Do not record customer
data, sessions or credentials in plans, journals or touchpoints.

## 8. Verify
Per locked item, the explicit check that proves it done.
- **D1** (`services/clyffe-api/internal/account`): Deny customer account endpoints until verified identity and tenant scoped access exist
- **D2** (`services/clyffe-api/internal/account`): Reuse verified sessions and scope account and order reads to membership
- **D3** (`services/clyffe-api/internal/account`): Require a restricted non-owner customer database role

Run the chi router with a nil store to verify it fails closed. On a disposable
PostgreSQL 18 database, run the two-tenant integration test with a restricted
login and actual `shippin-auth` login/logout, revoked/expired/inactive sessions,
forged headers, membership removal and pooled-context checks. Run Go tests and
vet. Other customer routes, provisioning and production deployment still need
their own end-to-end checks.

## Parked / deferred
_none_
