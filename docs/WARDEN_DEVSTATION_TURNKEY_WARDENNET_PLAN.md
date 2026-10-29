---
wardenclyffe_touchpoint:
  version: 1
  kind: devstation-turnkey-wardennet-plan
  namespace: wardenclyffe.warden.turnkey
  owner: docs/WARDEN_DEVSTATION_TURNKEY_WARDENNET_PLAN.md
  module: module-01-warden
  mesh_registry: wardenclyffe/registry/context-mesh.yaml
  reads:
    - docs/WARDEN_CONNECTOR_PLUGIN_DESIGNATION.md
    - docs/WARDEN_DEVSTATION_AND_CLYFFE_CODE.md
    - docs/CLYFFE_CODE_TURNKEY_SERVICE_SPEC.md
    - docs/CLYFFY_DYNAMIC_UI_SPEC.md
    - docs/FOUNDATION_SERVICE_MATRIX.md
    - docs/MASTER_CLYFFY_ROLLOUT_PLAN.md
    - docs/WARDEN_ESTABLISHMENT_POAM.md
---

# Warden VM · Devstation Turnkey · WardenNet — Establishment Plan

Senior-dev establishment plan for: (1) the Warden VM build + public IP migration
`104 → 204`, (2) **WardenNet** — our branded Tailscale (Headscale) overlay with
OPNsense WireGuard failsafe, (3) the **Devstation** as a secure operational space
for the Clyffy AI orchestrator, and (4) the **DevForge** turnkey workspace product
(hosted VSCode + seamless local-tool connect for Gemini / Cursor / Antigravity /
Codex / Claude). Operated as a Warden-managed service by **Hades**.

> Status: PLAN. Items marked **(confirm)** are load-bearing assumptions awaiting
> your confirmation. Items marked **(built)** already exist in this repo.

## 0. Open confirmations (load-bearing)

| # | Assumption | Needs |
|---|---|---|
| C1 | The `104 → 204` migration is the **public ingress IP** (`104.176.44.101` today → a new `204.x.x.x`). | exact target `204.x.x.x`; provider; cutover window |
| C2 | "Branded copy of Tailscale managed by us" = **Headscale** (self-hosted Tailscale control plane) as WardenNet, with OPNsense WireGuard as failsafe. | confirm Headscale vs alternative |
| C3 | **DevSpace** uses **Coder** (paid/waived, included). DevForge remains the Code-OSS editor on that workspace. Not an evaluation. Not later. | none — see `../shippin/docs/platform/PRIVATE-CLOUD-COMMUNITY.md` |
| C4 | Local tools (Gemini CLI, Cursor, Antigravity, Codex, Claude) connect over **WardenNet + Remote-SSH**, not public exposure. | confirm |

## 1. Roles & ownership

- **Warden** = infrastructure/control authority (Proxmox, hosts, routes, DNS,
  certs, inventory, audit). Owns lifecycle of every workspace and the IP cutover.
- **Clyffy** = AI orchestrator that operates *on* the devstation, reading Warden's
  captured model (built: `internal/clyffy`, `/api/clyffy/*`). Designation rules
  apply (built: `docs/WARDEN_CONNECTOR_PLUGIN_DESIGNATION.md`).
- **Clyffe** = customer surface ("my services").
- **Hades** = operator/owner who runs Warden-as-a-service for the estate
  (`mediastack/estate.toml` precedent). Warden is *managed by* Warden tooling,
  *operated by* Hades.

## 2. Two surfaces, one console (login-gated)

A single console (`apps/console`, built) presents two audiences, gated by
Authentik group (built blueprint plan in the designation + auth work):

| Surface | Audience | Shows |
|---|---|---|
| **Customer** (`/clyffe`) | customer | "My services" — assigned workspaces (DevForge, devstations), tier, status, billing refs, support. NO Proxmox, NO operator data. |
| **Admin** (`/warden`) | operator (Hades) | Warden infrastructure: every **host** with its **kind clearly labelled** (proxmox / coolify / baremetal / …), platforms, services, designation (connector/plugin), intelligence plane, audit, route intent, the IP-migration state. |

Host/platform-kind visibility is **built** (`shippin_infra.platforms.kind`,
`shippin_infra.services.role/designation`); the admin Foundation view already
renders it. This plan extends it with the customer surface + login gating.

## 3. WardenNet — branded Tailscale (Headscale) + OPNsense failsafe

The overlay that makes the public IP migration safe and the devstation reachable
without public exposure.

```
Primary overlay : WardenNet  = Headscale control plane (our branded Tailscale)
                  - tailscale clients on devstations, operator laptops, DevForge
                  - MagicDNS: devstation.wardennet, devforge-<id>.wardennet
                  - ACLs map to Authentik groups (operator vs customer vs ai)
Failsafe path   : OPNsense WireGuard (LXC 111)  - independent tunnel if Headscale
                  control plane is down or mid-migration
Public edge     : Caddy LXC 115 (clyffy-edge) terminates TLS; only the edge and
                  jump host are ever public (no workspace IP in public DNS)
```

Why this shape: connectivity to the devstation/services rides WardenNet, so it is
**independent of the public ingress IP**. That is what makes the `104 → 204`
cutover non-disruptive — clients keep their WardenNet address while the public A
records move.

New captured services (to add): `service.wardennet` (role `network`, connector),
`service.opnsense-wg` failsafe (already `service.edge-opnsense`).

## 4. Warden VM build + `104 → 204` IP migration runbook

**Build:** Warden VM (VMID 102 today, `warden.rrflow.ai`) absorbed into the Go
warden-api built here; reconciled from the repo by `scripts/dev/warden-reconcile.sh`
(built) on a systemd timer.

**Migration (104 → 204), Warden-owned, zero-downtime via WardenNet:**

1. **(confirm C1)** Acquire `204.x.x.x`; record both IPs in `shippin_infra` as a
   tracked migration (route intent owned by Warden).
2. Bring up the new edge on `204.x.x.x` (Caddy LXC 115) **in parallel** with the
   `104.176.44.101` edge. Both front the same internal services over WardenNet.
3. Lower TLS/DNS TTLs (PowerDNS LXC 109 + Cloudflare) to 60s ahead of cutover.
4. Issue certs for the new edge (step-ca LXC 108 / Cloudflare).
5. Flip public A records (`warden.*`, `master.clyffy.ai`, customer domains) from
   `104` → `204`; keep `104` serving until drained.
6. Verify via WardenNet (unaffected) + external probes; Warden records each step
   as an audit event.
7. Decommission the `104` edge after the drain window; keep as rollback for N days.

During all of this, devstation/DevForge sessions stay up because they ride
WardenNet, not the public IP.

## 5. Devstation as a secure AI operational space (Clyffy)

Make the devstation a true operational space where Clyffy (empowered by Codex /
Cursor / Claude / Gemini / Antigravity) can work, with clean privilege separation:

- **Identity:** `shippin_core.subjects.subject_kind = ai` for Clyffy/agents (built);
  AI auth via `ai_bridge.identity_grants` (radius / client_cert / password_failsafe,
  IP-bound) — built schema, blueprints pending Authentik.
- **Secrets:** brokered to `/run/warden-secrets` tmpfs via the `warden-secret-*`
  helpers (built on the VM) + Infisical; never in repo (enforced `.gitignore`).
- **Network:** WardenNet only; no public bind; code-server bound `127.0.0.1`.
- **Designation enforcement:** Clyffy reaches **plugins** (Qdrant/SurrealDB/Harrier/
  RRD) as ai-only and **connectors** only through Warden policy + audit (built,
  DB-enforced).
- **Operational loop:** repo reconciler keeps the VM current; Clyffy reads the
  captured model via `/api/clyffy/*`.

## 6. DevForge — the turnkey workspace product

DevForge is the productized hosted workspace (the Clyffe Code promise, named).
One Warden-provisioned workspace, reachable two ways:

```
                       WardenNet (Headscale) overlay
   local tools  ───────────────────────────────────────────►  devstation/DevForge VM
   (Gemini CLI, Cursor, Antigravity,                            - files, terminals, builds,
    Codex, Claude) via Remote-SSH                                 LSP, agents run here
                                                                - W-drive (SMB) mounted
   browser  ──────────►  DevForge (hosted VSCode: code-server)  - per-workspace Proxmox disk
                         OIDC-gated via Authentik through edge
```

**Turnkey flow (dynamic / automated via the frontend):** customer orders on the
console → Warden `automation` plan→approve (built) → `provisioner` clones the
DevForge template on Proxmox → attaches W-drive + per-workspace disk → joins
WardenNet → issues OIDC access → workspace appears in the customer surface. No
SSH keys, ports, DNS, or secrets managed by the customer (per
`CLYFFE_CODE_TURNKEY_SERVICE_SPEC.md`).

**Tool matrix (all "work locally, connect seamlessly"):**

| Tool | Connect model |
|---|---|
| Cursor / VS Code / Codex / Claude | Remote-SSH over WardenNet to the workspace |
| Gemini CLI / Antigravity | run in-workspace, or local CLI authenticated to the workspace over WardenNet |
| DevForge (browser) | `code-server` over the OIDC-gated edge; fallback when no local install |

New captured services (to add): `service.devforge-template` (role `devstation`,
core), and per-customer `service.devforge.<id>` rows on provision.

## 7. Data-model additions (capture these in Warden)

- `shippin_infra.platforms`: ensure `headscale`/`wardennet` represented (network platform).
- `shippin_infra.services`: add `service.wardennet` (connector, network),
  `service.devforge-template` (core), `service.rrd` (built).
- A small `shippin_infra.ip_migrations` table (from_ip, to_ip, state, started_at,
  cutover_at) so the `104→204` migration is first-class and auditable.
- `shippin_audit.events`: every migration + provision step (built sink).

## 8. Sequenced delivery

1. **(needs creds)** Infisical auth on devstation → wire warden-api to managed
   Postgres 110; create Qdrant collection; connect SurrealDB clyffy plane.
2. WardenNet: stand up Headscale (capture as service), enroll devstation +
   operator; OPNsense WireGuard failsafe.
3. Authentik blueprints: operator (Hades) vs customer groups; AI-identity
   (radius/cert/password-failsafe) — gate the two console surfaces.
4. Customer surface (`/clyffe` "my services") + admin host/kind view polish.
5. DevForge template + provisioner clone path; tool-connect docs per tool.
6. `104 → 204` migration runbook execution (Warden-owned, WardenNet-safe).

## 9. What I need from you

- Paste the **Infisical machine identity** (client id/secret + reachable URL) — see
  chat. Unblocks step 1 and everything live.
- Confirm **C1–C4** above.

## References
- `docs/WARDEN_CONNECTOR_PLUGIN_DESIGNATION.md` (designation + privilege model — built)
- `docs/WARDEN_DEVSTATION_AND_CLYFFE_CODE.md`, `docs/CLYFFE_CODE_TURNKEY_SERVICE_SPEC.md`
- `docs/CLYFFY_DYNAMIC_UI_SPEC.md` (Boundaries), `docs/FOUNDATION_SERVICE_MATRIX.md`
- `docs/RUNNING_THE_STACK.md`, `scripts/dev/warden-reconcile.sh`
