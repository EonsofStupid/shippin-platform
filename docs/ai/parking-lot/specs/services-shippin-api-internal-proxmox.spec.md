---
clyffy_touchpoint:
  version: 2
  workspace_id: wardenclyffe.intelligence
  project_key: services-shippin-api-internal-proxmox
  persona: clyffy-operator
  kind: buildspec
  owner: docs/ai/parking-lot/specs/services-shippin-api-internal-proxmox.spec.md
  module: null
  focus_feature: "Building WardenClyffe"
  boundary: services/shippin-api/internal/proxmox
  source_skill: docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md
  mesh_registry: wardenclyffe/registry/context-mesh.yaml
  sync:
    qdrant: true
    surreal: true
---

# Build Spec — services-shippin-api-internal-proxmox

Focus: **Building WardenClyffe**. Boundary: `services/shippin-api/internal/proxmox`. Emitted 2026-08-21 from agenda
`docs/ai/parking-lot/agenda/services-shippin-api-internal-proxmox.md` (4 locked, 1 parked).
Scaffolded by `scripts/foundation/parking-lot-spec.py`; fill each section from the
deliberation. Contract: `docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md` §Build Spec Contract.

## 1. Overview
Slice 1 extends the existing Proxmox operator page with a Backup tab: save an
S3-compatible connector into `/run/shippin-secrets` (WardenClyffe is an internal
layer name only), test it, then run restic (workspace) or vzdump (guest image).
Reuse `services/shippin-api/internal/proxmox` and `src/domains/warden/proxmox`.

**Non-goals (deliberate, until a later trigger):** PBS S3 datastore; ingesting
`.vma` into restic; renaming `/etc/warden` or `/api/warden/*`; new backup
microservice.

## 2. Golden patterns
Obeys: `docs/WARDENCLYFFE_NAMING_CONVENTIONS.md`,
`docs/HYPERMODULAR_DDD_FOLDER_STRUCTURE.md`,
`docs/ai/WARDENCLYFFE_STRUCTURE_STANDARD.md`.
Connect & Launch is the credential-intake pattern (operator-gated POST, never
log/echo secrets). Slice 0 Proxmox inventory is the UI/API spine. restic is
already on the host; no new AWS SDK.

## 3. Locked decisions
- **[feature] s3-backup-connector** (locked 2026-08-21): Operator-gated POST /api/warden/proxmox/backup/s3 saves an S3-compatible connector (R2/MinIO/AWS/B2) to gitignored secrets/s3.env and /run/shippin-secrets/s3.env. Status is redacted. Secrets never logged or returned. Test is HeadBucket with SigV4 and no new AWS SDK dependency.
- **[dependency] restic-and-vzdump** (locked 2026-08-21): Workspace/offsite backup uses restic (already on the host) against the saved S3 connector. Guest image backup uses Proxmox vzdump via the existing API client (task-true UPID, no 90s wait). PBS-as-S3-datastore is parked.

Agenda items locked:
- **D1** (`modules/warden/infrastructure/devstation/turnkey`): Canonical runtime keyring is /run/shippin-secrets; WardenClyffe stays an internal layer name; warden-secrets is compatibility only
- **D2** (`src/domains/warden/proxmox`): Backup UI is a tab on existing /admin/proxmox, not a new service or route tree
- **D3** (`services/shippin-api/internal/proxmox`): S3 connector saved operator-gated to secrets/s3.env + runtime keyring; never logged or echoed
- **D4** (`services/shippin-api/internal/proxmox`): Open-source backup engines: restic to S3-compatible object storage for workspace; Proxmox vzdump for guest images

## 4. Minimal requirements (10x foresight)
Per item: the smallest thing that ships *complete*, plus what to configure now
because it is 10x costlier to retrofit.
- **D1** (`modules/warden/infrastructure/devstation/turnkey`): Canonical runtime keyring is /run/shippin-secrets; WardenClyffe stays an internal layer name; warden-secrets is compatibility only
- **D2** (`src/domains/warden/proxmox`): Backup UI is a tab on existing /admin/proxmox, not a new service or route tree
- **D3** (`services/shippin-api/internal/proxmox`): S3 connector saved operator-gated to secrets/s3.env + runtime keyring; never logged or echoed
- **D4** (`services/shippin-api/internal/proxmox`): Open-source backup engines: restic to S3-compatible object storage for workspace; Proxmox vzdump for guest images

Configure now because retrofit is expensive: `/run/shippin-secrets` as the
canonical dir with a `warden-secrets` symlink; SigV4 without a vendor SDK;
action_requests for backup jobs (same table as start/stop).

## 5. Established patterns / formulas
- Operator gate: `platform.RequireOperator` like Connect & Launch.
- Env file format: dotenv `KEY=value`, mode 0600, gitignored `secrets/*.env`.
- Redaction: access key last 4 only; never return secret_key or restic password.
- Restic repo URL: `s3:<endpoint>/<bucket>/<prefix>`.

## 6. File-creation guidance (notes vs. code)
**Code:** `internal/proxmox/{s3,backup,s3_sign}.go`, `BackupPanel.tsx`, turnkey
`shippin-secret-*` helpers, tmpfiles unit.
**Touchpoints:** this spec, `docs/plans/slice-1-proxmox-backup.md`, parking-lot
decisions.
**Never create:** `scripts/secrets/` (boundary-guard duplicate), a new
`services/backup-api`, `/admin/backup` route.

## 7. LLM dynamic-flow within control-layer boundaries
Agents may read redacted `/backup/status` and the Backup tab copy. Agents must
never print S3 secret keys, restic passwords, or files under
`/run/shippin-secrets`. Connector values come from the operator on the tab, not
from chat.

## 8. Verify
- **D1:** `shippin-secrets-dir` prints `/run/shippin-secrets`; `readlink`
  `/run/warden-secrets` → `shippin-secrets`.
- **D2:** `/admin/proxmox` serves Tabs with Fleet + Backup; no `/admin/backup`.
- **D3:** unauthenticated POST `/backup/s3` → 401; authenticated save returns
  suffix only; GET status blob contains no secret values.
- **D4:** `restic_available` true on this host; `POST /backup/run` kind=
  `workspace` or `guest-image` creates `action_requests`.

## Parked / deferred
- **D5** (`services/shippin-api/internal/proxmox`): PBS S3 datastore and pulling .vma into restic is later; not this slice — _parked: PBS S3 datastore and vma-to-restic ingest need a live PVE storage probe; slice 1 ships restic+vzdump first._
