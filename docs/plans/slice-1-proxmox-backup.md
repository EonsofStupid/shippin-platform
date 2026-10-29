---
clyffy_touchpoint:
  version: 2
  workspace_id: wardenclyffe.infra
  project_key: slice-1-proxmox-backup
  persona: clyffy-operator
  kind: plan
  owner: docs/plans/slice-1-proxmox-backup.md
  module: module-01-warden
  focus_feature: "Building WardenClyffe"
  boundary: services/shippin-api/internal/proxmox
  sync:
    qdrant: true
    surreal: true
---

# Slice 1 — Backup tab + S3 connector + shippin-secrets

**Status:** implementing 2026-08-21.  
**Parking-lot:** `docs/ai/parking-lot/agenda/services-shippin-api-internal-proxmox.md`

## Concept

WardenClyffe is the **internal** name for Proxmox-as-Shippin layers. The runtime
keyring is **Shippin**. The operator saves an S3-compatible connector on the
existing Proxmox page and runs open-source backups from there.

## In scope

1. Canonical `/run/shippin-secrets` with `/run/warden-secrets` as a symlink.
2. Backup tab on `/admin/proxmox` (design-system Tabs). No new service.
3. Operator-gated save/test of S3 connector (`secrets/s3.env` + runtime file).
4. `restic` workspace backup to that connector (excludes cache/scratch/target).
5. `vzdump` guest-image backup via the existing Proxmox client (UPID, no 90s wait).

## Out of scope (parked)

- PBS S3 datastore
- Pulling `.vma` files into restic
- Renaming `/etc/warden` or `/api/warden/*`

## API

| Method | Path | Auth | Purpose |
|---|---|---|---|
| GET | `/api/warden/proxmox/backup/status` | open | redacted connector + restic + jobs |
| POST | `/api/warden/proxmox/backup/s3` | operator | save connector (never echoed) |
| POST | `/api/warden/proxmox/backup/s3/test` | operator | HeadBucket |
| POST | `/api/warden/proxmox/backup/run` | operator | `workspace` or `guest-image` |

## Verify

1. `GET /api/warden/proxmox/backup/status` → 200, `s3.configured` bool, no secret values.
2. Save connector from the Backup tab; status shows suffix only.
3. Test connection against a real bucket (operator provides keys).
4. Workspace backup creates `backup_workspace` action_request.
5. Guest image backup returns a Proxmox UPID when PVE is configured.
