---
clyffy_touchpoint:
  version: 2
  workspace_id: wardenclyffe.intelligence
  project_key: services-shippin-api-internal-proxmox
  persona: clyffy-operator
  kind: agenda
  owner: docs/ai/parking-lot/agenda/services-shippin-api-internal-proxmox.md
  module: null
  focus_feature: "Building WardenClyffe"
  boundary: services/shippin-api/internal/proxmox
  domain_status: existing
  source_skill: docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md
  mesh_registry: wardenclyffe/registry/context-mesh.yaml
  sync:
    qdrant: true
    surreal: true
---

# Parking-Lot Agenda — services-shippin-api-internal-proxmox

Focus: **Building WardenClyffe**. Boundary: `services/shippin-api/internal/proxmox`. Opened 2026-08-21.

Ordered decision items. State ∈ open|locked|parked; `rounds` capped at 12. Managed by `scripts/foundation/parking-lot-agenda.py` per `docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md`. Decisions are captured by `parking-lot-capture.py`; the spec is emitted by `parking-lot-spec.py`.

<!-- agenda:start -->
- [1] locked rounds=1 || q=Canonical runtime keyring is /run/shippin-secrets; WardenClyffe stays an internal layer name; warden-secrets is compatibility only || boundary=modules/warden/infrastructure/devstation/turnkey
- [2] locked rounds=1 || q=Backup UI is a tab on existing /admin/proxmox, not a new service or route tree || boundary=src/domains/warden/proxmox
- [3] locked rounds=1 || q=S3 connector saved operator-gated to secrets/s3.env + runtime keyring; never logged or echoed || boundary=services/shippin-api/internal/proxmox
- [4] locked rounds=1 || q=Open-source backup engines: restic to S3-compatible object storage for workspace; Proxmox vzdump for guest images || boundary=services/shippin-api/internal/proxmox
- [5] parked rounds=1 || q=PBS S3 datastore and pulling .vma into restic is later; not this slice || boundary=services/shippin-api/internal/proxmox || reason=PBS S3 datastore and vma-to-restic ingest need a live PVE storage probe; slice 1 ships restic+vzdump first.
<!-- agenda:end -->
