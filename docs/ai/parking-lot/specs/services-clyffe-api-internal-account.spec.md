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
`docs/ai/parking-lot/agenda/services-clyffe-api-internal-account.md` (1 locked, 0 parked).
Contract: `docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md` §Build Spec Contract.

## 1. Overview
The current customer handlers can list all tenants and orders without verifying
an identity. Return 503 before accessing those stores until authorization exists.

**Non-goals:** Implementing customer login, scoped authorization, or managed
projects. The 503 response is a temporary restriction, not tenant isolation.

## 2. Golden patterns
Obeys: `docs/WARDENCLYFFE_NAMING_CONVENTIONS.md`,
`docs/HYPERMODULAR_DDD_FOLDER_STRUCTURE.md`,
`docs/ai/WARDENCLYFFE_STRUCTURE_STANDARD.md`. Reuse the existing account handler;
the boundary review found that context already exists. Leave health separate.

## 3. Locked decisions
- **[feature] fail closed customer account endpoints** (locked 2026-11-01): Return service unavailable for the existing customer home, accounts and orders routes until verified customer identity and tenant authorization exist; keep health available.

Agenda items locked:
- **D1** (`services/clyffe-api/internal/account`): Deny customer account endpoints until verified identity and tenant scoped access exist

## 4. Minimal requirements (10x foresight)
Per item: the smallest thing that ships *complete*, plus what to configure now
because it is 10x costlier to retrofit.
- **D1** (`services/clyffe-api/internal/account`): Deny customer account endpoints until verified identity and tenant scoped access exist

All three registered customer routes must stop before reading the database for
every request, including a caller-supplied tenant ID or arbitrary bearer token.
Reenable only after a verified principal and membership are applied server-side
to every query, with two-tenant negative tests and appropriate database roles.

## 5. Established patterns / formulas
`customer route -> unavailable response`; future replacement:
`verified principal -> authorized tenant -> scoped store query`.

## 6. File-creation guidance (notes vs. code)
Which deliverables are **touchpoints/Markdown** (routing, decisions, runbooks)
and which are **code** (boundary, language, colocated companions). What must
never be created (defer duplicates to boundary-guard).
Code and tests remain in `services/clyffe-api/internal/account/`; the release
roadmap describes the restriction. This touchpoint is not an identity store or
API contract. Do not add an unreviewed parallel tenant authority.

## 7. LLM dynamic-flow within control-layer boundaries
How Clyffy and the coding agents operate against this spec at runtime: the
designation rules (connector vs plugin, ai-identity), what they may read, and
what must never enter an agent's context (the brokered secret boundary).
Agents can review the source and propose a tenant-aware replacement. This change
creates no runtime agent, connector or secrets interface. Do not record customer
data, sessions or credentials in plans, journals or touchpoints.

## 8. Verify
Per locked item, the explicit check that proves it done.
- **D1** (`services/clyffe-api/internal/account`): Deny customer account endpoints until verified identity and tenant scoped access exist

Run the real chi router with a nil store: every registered customer route must
return 503 for absent and arbitrary tokens and for a supplied tenant ID without
reaching the store. Run the module tests and vet. Future release requires real
cross-tenant integration tests; this restriction alone does not provide them.

## Parked / deferred
_none_
