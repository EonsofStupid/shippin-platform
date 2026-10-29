---
wardenclyffe_touchpoint:
  version: 1
  kind: root-agent-contract
  namespace: wardenclyffe.root
  base_skill: docs/ai/WARDENCLYFFE_BASE_SKILL.md
  mesh_touchpoints: docs/ai/MCP_MESH_TOUCHPOINTS.md
  current_mesh_registry: wardenclyffe/registry/context-mesh.yaml
  intelligence_contract: docs/ai/INTELLIGENCE_TOUCHPOINTS.md
---

# Shippin Platform Agent Contract

This is the root project context for Codex, Claude, Cursor, Gemini, and
other coding agents working in this repository.

Read this before making changes. Subtree `AGENTS.md` files add local rules
and nearest file wins for that subtree.

## Paid engines — no penalty box

Umbrella law: `../shippin/docs/platform/PRIVATE-CLOUD-COMMUNITY.md`.

Coder on a workspace is a Shippin **DevSpace**. It is included in the
private-cloud community, same class as Chatmaton. Nomen is the IdP
(`../tessera`). Do not treat AGPL, “evaluate Coder later,” or
“customers must never see the engine” as a reason to disable what was
paid for. That sentence is void.

## What This Repo Is

This repository is active implementation and migration evidence for the
**Shippin** customer and infrastructure management platform:

- **Shippin** owns customer, entitlement, workspace, host, infrastructure
  lifecycle, backup/recovery, Operator, and Estate truth.
- **Warden** is a legacy implementation/module name for operator and
  server-control code under Shippin.
- **Clyffe** is a legacy implementation/module name for customer-safe portal,
  knowledge, ticket, CRM, and service-panel code under Shippin.
- **Wardenclyffe is specifically the Zuul zero-mesh product**, not the name of
  the Warden/Clyffe split and not the Shippin infrastructure manager.
- **Module 1** and **Module 2** remain planning labels only.
- Proxmox is one current Shippin infrastructure substrate for the internal
  two-server pilot.

The nested `wardenclyffe/` repo name is a compatibility identifier. It
contains active Go Warden implementation and Context Mesh/MCP/intelligence
scaffolding, but it is not Wardenclyffe/Zuul product authority. Treat it as an
important source of current Shippin operational evidence, and do not mutate it
unless the task is explicitly about that Go implementation repo.

## Read Next

Use this order when you need architecture or agent context:

1. `docs/ai/WARDENCLYFFE_BASE_SKILL.md`
1a. `docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md` — FORCED planning/decision-capture
    mode (auto-entered on any planning turn; locks deps/features/capabilities into
    the SurrealDB intelligence layer via decision touchpoints)
1b. `docs/ai/WARDENCLYFFE_BOUNDARY_GUARD_SKILL.md` — FORCED pre-creation review
    (file-tree watch + duplicate check before any new boundary/folder)
2. `docs/ai/MCP_MESH_TOUCHPOINTS.md`
3. `docs/ai/INTELLIGENCE_TOUCHPOINTS.md`
3a. `docs/ai/INTELLIGENCE_LAYER_MODERNIZATION.md` for the rule that Markdown
    is a touchpoint layer, not the intelligence store
3b. `docs/ai/SURREALDB_INTELLIGENCE_PROJECTION_V2.md` for the dynamic
    SurrealDB projection and context-pack plan
3c. `docs/SURREALDB_PUBLIC_SELF_HOSTING_PLAN.md` for the public-safe
    self-hosted SurrealDB route, auth, backup, and Warden proxy posture
3d. `docs/SURREALDB_SELF_HOSTED_RUNBOOK.md` for the live LXC `104` service,
    backup, restore, and cloud export/import gate
4. `docs/WARDENCLYFFE_MODULE_MAP.md`
5. `docs/WARDENCLYFFE_NAMING_CONVENTIONS.md`
6. `docs/HYPERMODULAR_DDD_FOLDER_STRUCTURE.md`
7. `docs/HOST_FLEET_AND_ONBOARDING.md` for host onboarding and fleet work
8. `docs/PUBLIC_IP_HOMEBASE_FOUNDATION.md` for public IP, domain, edge, and Tailscale-exit work
9. `docs/FOUNDATION_SERVICE_MATRIX.md` for VM/LXC app choices and configuration gates
10. `docs/WARDEN_ESTABLISHMENT_POAM.md` for done/needed POA&M status and next live milestones
11. `docs/WARDEN_OPERATOR_CAPSULE.md` for the Linux-first operator workspace and secret-handling capsule
12. `docs/WARDEN_DEVSTATION_AND_CLYFFE_CODE.md` for the private VS Code VM and future hosted coding service
13. `docs/WARDEN_REMOTE_AGENT_STREAMS.md` for Codex/Claude remote stream launch and future app-server gates
14. `docs/WARDEN_SHARED_STORAGE_PLAN.md` for the server1 400 GB hot-tier storage boundary and server2 migration direction
15. `docs/MASTER_CLYFFY_ROLLOUT_PLAN.md` for master.clyffy.ai, Postgres update, and Clyffy route work
13a. `docs/CLYFFY_MCP_ORCHESTRATOR.md` for the main Clyffy MCP orchestrator and foundation service boundaries
13b. `docs/CLYFFY_DYNAMIC_UI_SPEC.md` for dynamic Clyffy/Clyffe UI work
13c. `docs/CLYFFY_DYNAMIC_UI_POAM.md` for Clyffy UI milestones and sprints
13d. `docs/CLYFFE_CODE_TURNKEY_SERVICE_SPEC.md` for Clyffe Code managed workspace product work
14. `docs/FOUNDATION_APP_RESEARCH_2026_05.md` for primary-source app-stack research
15. `docs/WARDEN_CLYFFE_ARCHITECTURE.md`
16. `docs/PROXMOX_FREE_CHEATSHEET.md` for Proxmox management work
17. `docs/WARDENCLYFFE_BACKEND_OPTIONS_2026_05.md` for backend evaluation
18. `docs/GO_WARDEN_ABSORPTION_PLAN.md` when folding Go Warden into the root direction
19. `docs/WARDEN_CLYFFE_PILOT_ROADMAP.md`
20. `modules/README.md` for the root product module scaffold
21. `wardenclyffe/AGENTS.md` when working with or borrowing from the Go Warden repo
22. `wardenclyffe/registry/context-mesh.yaml` for current Context Mesh/MCP registry shape (v3+ per ADR 0030)
23. `wardenclyffe/docs/runbooks/mcp-2026-alignment-checkpoint.md` for the MCP 2026 alignment record, gap matrix, 10 resolved decisions, and the phased Phase 1–9 execution plan
24. `wardenclyffe/docs/specs/14-mcp-federation-and-workspace.md` for the L0+L1+L2 MCP federation model (workspace publish-upward, gateway, leaf contracts)
25. `wardenclyffe/docs/decisions/0030-mcp-2026-baseline.md` for the May 2026 protocol capability bar every new MCP server must meet
26. `wardenclyffe/docs/decisions/0031-workspace-identity.md` for the workspace identity model and the SurrealDB-ns partial-equivalence rule
27. `wardenclyffe/docs/decisions/0033-touchpoint-protocol.md` for the v2 `clyffy_touchpoint` frontmatter shape (replaces v1 `namespace_id` with `workspace_id`)

## Agent Wrapper Rule

This file is the cross-agent wrapper. Thin wrappers may exist for specific
tools:

- `CLAUDE.md`
- `.cursor/rules/wardenclyffe-intelligence.mdc`
- `.cursor/skills/wardenclyffe-base/SKILL.md`
- `.agents/skills/wardenclyffe-base/SKILL.md`

The **parking-lot** and **boundary-guard** skills are authored once under
`docs/ai/` and wrapped for every agent family (Claude, Codex, Cursor, generic
`.agents`):

- `docs/ai/WARDENCLYFFE_PARKING_LOT_SKILL.md` (source) →
  `.claude/skills/parking-lot/`, `.agents/skills/parking-lot/`,
  `.cursor/skills/parking-lot/`, `.codex/skills/parking-lot/`,
  `.cursor/rules/parking-lot.mdc`
- `docs/ai/WARDENCLYFFE_BOUNDARY_GUARD_SKILL.md` (source) →
  `.claude/skills/boundary-guard/`, `.agents/skills/boundary-guard/`,
  `.cursor/skills/boundary-guard/`, `.codex/skills/boundary-guard/`,
  `.cursor/rules/boundary-guard.mdc`

Forcing is wired in `.claude/settings.json`: a `UserPromptSubmit` hook
(`scripts/foundation/parking-lot-hook.py`) auto-enters parking-lot mode on
planning/decision language, and a `PreToolUse` hook
(`scripts/foundation/boundary-guard-hook.py`) runs the duplicate check before a
new folder is created. Cursor uses the always-apply rules; Codex/other agents
use the `description` triggers.

Do not let wrappers become new authorities. Update the base files under
`docs/ai/` first, then adjust wrappers only if their pointer text changes.

## Safety Rules

- Never commit secrets, token values, passwords, cookies, private keys, or
  live API responses that expose credentials.
- Use repo-relative paths in committed docs.
- Treat customer actions as tenant-scoped and audit-logged.
- Treat destructive infrastructure actions as approval workflows unless a
  human explicitly asks for execution.
- Preserve Warden/Clyffe identifiers where code compatibility requires them,
  but use Shippin for product ownership. Wardenclyffe names the Zuul zero-mesh
  product only.

## MCP And Intelligence Rule

Markdown files with `clyffy_touchpoint` frontmatter (v2 shape per
`wardenclyffe/docs/decisions/0033-touchpoint-protocol.md`) are intentional
touchpoints for agents and the MCP mesh tooling. They are not the intelligence
database. Product truth belongs in Postgres/Warden APIs, retrieval belongs in
Qdrant, graph projection belongs in SurrealDB, and run history belongs in
Warden task/audit/trace records. The v1 `wardenclyffe_touchpoint`
shape is deprecated; the v1 → v2 migration window is documented in ADR 0033 §10.
Touchpoints should describe:

- what the file owns,
- which registry or source of truth it defers to,
- which agents should read it,
- what must not be changed from that file,
- which workspace (`workspace_id`) and project_key it routes intelligence writes to.

Do not hardcode a second tool registry in prose. The current registry source
is `wardenclyffe/registry/context-mesh.yaml` (v3+ per ADR 0030) until a root
registry is promoted. The MCP capability bar for new servers is **ADR 0030**;
the federation model is **spec 14 + ADR 0032**; the workspace identity model
(and the partial-equivalence rule that prevents conflating SurrealDB `ns` with
workspace slug) is **ADR 0031**.
