---
wardenclyffe_touchpoint:
  version: 1
  kind: base-skill
  namespace: wardenclyffe.skill.base
  wraps:
    - AGENTS.md
    - CLAUDE.md
    - .cursor/rules/wardenclyffe-intelligence.mdc
    - .cursor/skills/wardenclyffe-base/SKILL.md
    - .agents/skills/wardenclyffe-base/SKILL.md
  mesh_registry: wardenclyffe/registry/context-mesh.yaml
---

# Shippin Platform Base Skill (legacy WardenClyffe wrapper name)

This is the source skill that other AI wrappers should defer to.

## Operating Frame

- Shippin is the customer and infrastructure management platform.
- Warden is a legacy implementation/module name for Shippin operator and
  server-control code.
- Clyffe is a legacy implementation/module name for Shippin customer portal,
  knowledge base, tickets, CRM, and customer-safe service-panel code.
- Wardenclyffe is specifically the Zuul zero-mesh product. It is not the name
  of this Warden/Clyffe implementation split.
- Module 1 is Warden. Module 2 is Clyffe. Module numbers are planning labels,
  not durable code names.
- Proxmox is the current substrate for the internal two-server pilot.
- The nested `wardenclyffe/` repo name is a compatibility identifier. It
  contains the active Go Warden scaffold and current Context Mesh registry,
  but is not Wardenclyffe/Zuul product authority.

## How Agents Should Work

1. Read the nearest `AGENTS.md` before editing.
2. Use Shippin for product ownership; retain Warden/Clyffe only where code or
   migration compatibility requires it.
3. Keep customer-facing actions scoped through Shippin APIs, including legacy
   Warden implementation endpoints during migration.
4. Treat MCP mesh definitions as registry-owned, not prose-owned.
5. Add or update Markdown touchpoints when a new domain, subsystem, or
   intelligence route is introduced.
6. Never commit secrets or live credential material.
7. Paid/waived engines are included. Coder is a DevSpace. Nomen is the
   IdP. Do not re-open Authentik-vs-Zitadel or AGPL-as-blocker.
   `../shippin/docs/platform/PRIVATE-CLOUD-COMMUNITY.md`.

## Source-Of-Truth Order

Use this order when sources disagree:

1. Live code and configs in the relevant subtree.
2. Nearest `AGENTS.md`.
2a. `../shippin/docs/platform/PRIVATE-CLOUD-COMMUNITY.md` for included engines.
2b. `../shippin/docs/platform/NAMING-AND-SSOT.md` and
    `../shippin/docs/platform/PRODUCT-ARCHITECTURE.md` for product ownership.
3. `docs/WARDENCLYFFE_MODULE_MAP.md`.
4. `docs/WARDENCLYFFE_NAMING_CONVENTIONS.md`.
5. `docs/HYPERMODULAR_DDD_FOLDER_STRUCTURE.md`.
6. `docs/HOST_FLEET_AND_ONBOARDING.md`.
7. `docs/PUBLIC_IP_HOMEBASE_FOUNDATION.md`.
8. `docs/FOUNDATION_SERVICE_MATRIX.md`.
9. `docs/WARDEN_ESTABLISHMENT_POAM.md`.
10. `docs/WARDEN_OPERATOR_CAPSULE.md`.
11. `docs/WARDEN_DEVSTATION_AND_CLYFFE_CODE.md`.
12. `docs/WARDEN_REMOTE_AGENT_STREAMS.md`.
13. `docs/WARDEN_CLOUDFLARE_DNS_FOUNDATION.md`.
14. `docs/MASTER_CLYFFY_ROLLOUT_PLAN.md`.
15. `docs/CLYFFY_MCP_ORCHESTRATOR.md`.
16. `docs/CLYFFY_DYNAMIC_UI_SPEC.md`.
17. `docs/CLYFFY_DYNAMIC_UI_POAM.md`.
18. `docs/CLYFFE_CODE_TURNKEY_SERVICE_SPEC.md`.
19. `docs/FOUNDATION_APP_RESEARCH_2026_05.md`.
20. `docs/WARDEN_CLYFFE_ARCHITECTURE.md`.
21. `docs/WARDEN_CLYFFE_PILOT_ROADMAP.md`.
22. `wardenclyffe/REGISTRY.md` and `wardenclyffe/registry/context-mesh.yaml`
   for the Go-side Warden/intelligence mesh.
23. `docs/ai/INTELLIGENCE_LAYER_MODERNIZATION.md` for the Markdown-as-touchpoint,
   generated-stores-as-memory rule.
24. `docs/ai/SURREALDB_INTELLIGENCE_PROJECTION_V2.md` for the dynamic
   SurrealDB projection and context-pack plan.
25. Wrapper files such as `CLAUDE.md` and `.cursor/rules/*.mdc`.

Wrappers are pointers. They are not authorities.

## When To Update This Skill

Update this file when:

- the Warden/Clyffe product boundary changes,
- the MCP mesh source of truth moves,
- a new agent family needs a wrapper,
- the intelligence layer gets a new memory or routing contract,
- a new repo-wide safety rule is needed.
