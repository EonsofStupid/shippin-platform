---
clyffy_touchpoint:
  version: 2
  workspace_id: wardenclyffe.intelligence
  project_key: docs-shippin-platform-roadmap-md
  persona: clyffy-operator
  kind: buildspec
  owner: docs/ai/parking-lot/specs/docs-shippin-platform-roadmap-md.spec.md
  module: null
  focus_feature: "Shippin managed database projects"
  boundary: docs/shippin-platform-roadmap.md
  source_skill: docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md
  mesh_registry: wardenclyffe/registry/context-mesh.yaml
  sync:
    qdrant: true
    surreal: true
---

# Build Spec — docs-shippin-platform-roadmap-md

Focus: **Shippin managed database projects**. Boundary: `docs/shippin-platform-roadmap.md`. Emitted 2026-11-01 from agenda
`docs/ai/parking-lot/agenda/docs-shippin-platform-roadmap-md.md` (2 locked, 1 parked).
Contract: `docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md` §Build Spec Contract.

## 1. Overview
Protect the existing operator data browser, fleet and provisioning routes and
record the path to managed application/database projects in the release roadmap.

**Non-goals:** A customer tenant API, an implemented managed database service,
production identity, quotas, host operations, GitHub provisioning and deployment.
All require the separate work and checks in the roadmap.

## 2. Golden patterns
Obeys: `docs/WARDENCLYFFE_NAMING_CONVENTIONS.md`,
`docs/HYPERMODULAR_DDD_FOLDER_STRUCTURE.md`,
`docs/ai/WARDENCLYFFE_STRUCTURE_STANDARD.md`. Extend the existing Go contexts,
existing React callers and existing release roadmap; do not create a competing
data or identity authority. The older Warden/Clyffe naming reflects current
source and will be migrated deliberately under the user's Shippin direction.

## 3. Locked decisions
- **[feature] managed project release path** (locked 2026-11-01): Record explicit tenant isolation, provisioning, metering, source integration and recovery gates in the existing platform roadmap; unimplemented capabilities remain planned.
- **[feature] operator route protection**: Existing data, fleet and provision
  endpoints require an operator session. A separate, tenant-aware interface is
  needed for customer resources. Leaving them public would expose platform data.

Agenda items locked:
- **D1** (`services/shippin-api`): Protect existing operator data, fleet and provisioning endpoints and preserve authenticated UI access
- **D2** (`docs/shippin-platform-roadmap.md`): Record managed-project lifecycle, ownership and verifiable delivery milestones

## 4. Minimal requirements (10x foresight)
Per item: the smallest thing that ships *complete*, plus what to configure now
because it is 10x costlier to retrofit.
- **D1** (`services/shippin-api`): Protect existing operator data, fleet and provisioning endpoints and preserve authenticated UI access
- **D2** (`docs/shippin-platform-roadmap.md`): Record managed-project lifecycle, ownership and verifiable delivery milestones

The server must validate every listed route, not rely on React navigation. The
existing UI must send the existing session and report errors. Future work must
keep tenant resources, database credentials, quota reservations and customer
data scoped to their verified owner and prevent cross-tenant access.

## 5. Established patterns / formulas
`handler -> RequireOperator -> domain operation` for these legacy staff routes;
`identity session -> authenticated client -> handler` for existing views.
Future customer actions require their own verified tenant and object permission;
an operator bearer is not customer authorization.

## 6. File-creation guidance (notes vs. code)
Which deliverables are **touchpoints/Markdown** (routing, decisions, runbooks)
and which are **code** (boundary, language, colocated companions). What must
never be created (defer duplicates to boundary-guard).
Code stays in `services/shippin-api/internal/{data,fleet,automation}` and their
current wiring; callers stay in existing `src/domains/warden/` views and
`src/lib/api.ts`. This touchpoint routes agents to the roadmap, which owns
release direction; neither file is an API schema, database or service registry.
The boundary guard identified existing contexts, which are reused.

## 7. LLM dynamic-flow within control-layer boundaries
How Clyffy and the coding agents operate against this spec at runtime: the
designation rules (connector vs plugin, ai-identity), what they may read, and
what must never enter an agent's context (the brokered secret boundary).
Agents may read the roadmap and source and propose scoped changes. They must not
place session tokens, provider credentials, customer data or environment files in
touchpoints, plans or journals. No new agent tool, connector, AI identity or
secrets broker is implemented by this work; the eventual runtime must enforce
its own least-privilege policy.

## 8. Verify
Per locked item, the explicit check that proves it done.
- **D1** (`services/shippin-api`): Protect existing operator data, fleet and provisioning endpoints and preserve authenticated UI access
- **D2** (`docs/shippin-platform-roadmap.md`): Record managed-project lifecycle, ownership and verifiable delivery milestones

Run Go route tests with absent, invalid and revoked sessions for every route;
use a disposable Postgres instance to verify actual login, data queries,
provision request, fleet reads and logout. Verify the frontend sends the token
and rejects a missing token; run build and typecheck. The roadmap lists further
end-to-end checks, which remain outstanding until real services exist.

## Parked / deferred
- **D3** (`docs/shippin-platform-roadmap.md`): Determine two-instance and three-GB allowance scope — _parked: Awaiting owner decisions on allowance scope and storage allocation._
