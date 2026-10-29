---
wardenclyffe_touchpoint:
  version: 1
  kind: subtree-agent-contract
  namespace: wardenclyffe.docs
  reads:
    - ../AGENTS.md
    - ai/WARDENCLYFFE_BASE_SKILL.md
---

# Docs Agent Contract

The repo-root `AGENTS.md` applies here.

Docs should clarify that Shippin owns customer and infrastructure management,
while Wardenclyffe is the Zuul zero-mesh product. They may retain Warden and
Clyffe as implementation/module compatibility names. Docs also cover Proxmox
substrate strategy, MCP mesh touchpoints, intelligence-layer contracts, and
the legacy Module 1 Warden / Module 2 Clyffe planning split.

Use `docs/ai/TOUCHPOINT_TEMPLATE.md` when adding new agent-facing touchpoints.

Before adding or renaming product docs, check:

- `WARDENCLYFFE_MODULE_MAP.md`
- `WARDENCLYFFE_NAMING_CONVENTIONS.md`
- `HYPERMODULAR_DDD_FOLDER_STRUCTURE.md`
- `HOST_FLEET_AND_ONBOARDING.md`
- `PUBLIC_IP_HOMEBASE_FOUNDATION.md`
- `FOUNDATION_SERVICE_MATRIX.md`
- `WARDEN_ESTABLISHMENT_POAM.md`
- `WARDEN_OPERATOR_CAPSULE.md`
- `WARDEN_DEVSTATION_AND_CLYFFE_CODE.md`
- `MASTER_CLYFFY_ROLLOUT_PLAN.md`
- `FOUNDATION_APP_RESEARCH_2026_05.md`
- `WARDENCLYFFE_BACKEND_OPTIONS_2026_05.md`
- `PROXMOX_FREE_CHEATSHEET.md`
