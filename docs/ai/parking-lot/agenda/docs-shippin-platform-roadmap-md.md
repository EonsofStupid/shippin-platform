---
clyffy_touchpoint:
  version: 2
  workspace_id: wardenclyffe.intelligence
  project_key: docs-shippin-platform-roadmap-md
  persona: clyffy-operator
  kind: agenda
  owner: docs/ai/parking-lot/agenda/docs-shippin-platform-roadmap-md.md
  module: null
  focus_feature: "Shippin managed database projects"
  boundary: docs/shippin-platform-roadmap.md
  domain_status: existing
  source_skill: docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md
  mesh_registry: wardenclyffe/registry/context-mesh.yaml
  sync:
    qdrant: true
    surreal: true
---

# Parking-Lot Agenda — docs-shippin-platform-roadmap-md

Focus: **Shippin managed database projects**. Boundary: `docs/shippin-platform-roadmap.md`. Opened 2026-11-01.

Ordered decision items. State ∈ open|locked|parked; `rounds` capped at 12. Managed by `scripts/foundation/parking-lot-agenda.py` per `docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md`. Decisions are captured by `parking-lot-capture.py`; the spec is emitted by `parking-lot-spec.py`.

<!-- agenda:start -->
- [1] locked rounds=1 || q=Protect existing operator data, fleet and provisioning endpoints and preserve authenticated UI access || boundary=services/shippin-api
- [2] locked rounds=1 || q=Record managed-project lifecycle, ownership and verifiable delivery milestones || boundary=docs/shippin-platform-roadmap.md
- [3] parked rounds=0 || q=Determine two-instance and three-GB allowance scope || boundary=docs/shippin-platform-roadmap.md || reason=Awaiting owner decisions on allowance scope and storage allocation.
<!-- agenda:end -->
