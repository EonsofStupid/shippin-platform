---
clyffy_touchpoint:
  version: 2
  workspace_id: wardenclyffe.infra
  project_key: turnkey-vm-capsule-research
  persona: clyffy-operator
  kind: research-brief
  owner: docs/WARDEN_TURNKEY_VM_CAPSULE_RESEARCH_2026_06.md
  module: module-01-warden
  mesh_registry: wardenclyffe/registry/context-mesh.yaml
  reads:
    - docs/WARDEN_DEVSTATION_AND_CLYFFE_CODE.md
    - docs/WARDEN_OPERATOR_CAPSULE.md
    - docs/CLYFFE_CODE_TURNKEY_SERVICE_SPEC.md
    - docs/specs/WORKSTATION_BRIDGE_TEMPLATE.md
  sync:
    qdrant: false
    surreal: false
---

# Deep Research Task — Turnkey AI Devstation + Capsule Template & Deterministic Buildplan Baseline

Written in COSTAR shape (Context · Objective · Style · Tone · Audience · Response)
so it also serves as a best-practice prompt template. Core constraint repeated
throughout: **deterministic scripts + explicit if/then checks, not an AI agent
improvising.** AI is a consumer of the workspace, never the thing that builds or
secures it.

## (C) Context

We operate **WardenClyffe**: Warden (operator/infra control plane, Go) and Clyffe
(customer portal). We are productizing a **turnkey, subscription hosted coding
workspace** ("Clyffe Code") proven first by our private operator VM
`warden-devstation-01` (Proxmox, VMID 116, 8 vCPU / 16 GiB / 160 GiB, private
`vmbr1`, no public route). The product gives an **indie dev or AI a safe remote
workspace they drive from their local machine** using **Claude Code, Codex CLI,
Cursor (Remote-SSH), and Gemini CLI** — files, builds, terminals, agents, and
language servers run on the remote VM; the local app is just the editor shell.

Hard environment facts to design within:
- **Two surfaces:** a customer **Devstation** VM and a secret-sensitive **Operator Capsule**.
- **Secrets** are brokered (Infisical machine identity → `/run/warden-secrets` tmpfs;
  SOPS/age; **never** in repo or process env). No API key may leak to the workspace
  user or to an AI agent's context.
- **Networking:** private only, reached over **WardenNet** (Headscale/Tailscale
  overlay) + OPNsense WireGuard failsafe; public DNS points only at a Warden-owned
  edge/jump. Browser IDE (`code-server`) binds `127.0.0.1` and is tunneled.
- **Storage:** a per-service **W-drive** (today SMB `//10.0.0.117/warden-storage`;
  product direction is per-tenant volumes).
- **Hardware:** dedicated **NVIDIA Blackwell** systems are being added soon to power a
  **hybrid** model (local inference + cloud), where our assistant **Clyffy** optimizes
  prompts (COSTAR / best-practice) before they reach cloud models like Claude.
- **Governance:** everything is captured as versioned Markdown **touchpoints** that
  project into a **SurrealDB + Qdrant intelligence layer** and share with **Postgres**
  product truth. This research output must itself become a reusable **buildplan
  baseline** for every AI we work with.

## (O) Objective

Produce a **single authoritative, evidence-backed buildplan** for a **cutting-edge,
safe, turnkey AI-developer VM + Capsule template**, plus the **reusable deterministic
buildplan methodology** (layers → passes → requirement gates) that produced it. The
provisioning and lifecycle must be **auto-scripted, idempotent, and driven by explicit
if/then check logic — not by an AI agent making judgment calls at runtime.** AI is a
*consumer* of the workspace, never the thing that builds or secures it.

Answer these research dimensions, each with primary-source evidence and a
recommendation with rationale:

**Layer 1 — Substrate & VM template**
1. Best-practice for a **golden VM template** on Proxmox (cloud-init + Packer vs.
   Proxmox template clone vs. NixOS image): reproducibility, drift control,
   snapshot/rollback. Compare and recommend.
2. Per-tenant **resource isolation & quotas** (cgroups v2, vCPU/RAM/disk caps,
   idle/suspend policy) and how to enforce them deterministically.

**Layer 2 — Capsule & secret safety (the "safe space" core)**
3. Patterns for an **AI-safe secret boundary**: brokered runtime secrets
   (Infisical/OpenBao), tmpfs, short-lived tokens, and a **secrets-exclusion guard**
   that prevents any secret from entering an AI agent's file/context window. Survey
   real tools/policies (gitleaks/trufflehog pre-read scanning, OS keyring, FS ACLs,
   seccomp/AppArmor, microVMs like Firecracker/Kata for agent sandboxing).
4. **Egress control** for a box running autonomous coding agents (allowlist DNS/IP,
   per-agent network namespaces) without breaking legitimate package/model traffic.

**Layer 3 — Remote-local AI tooling bridge**
5. Authoritative connection models for **Claude Code, Codex CLI, Cursor, Gemini CLI**
   into a remote Linux VM (Remote-SSH vs. tunnels vs. browser IDE). **Coder is
   the included DevSpace engine** (paid/waived; not an evaluation). See
   `../shippin/docs/platform/PRIVATE-CLOUD-COMMUNITY.md`. Which
   combination gives the truest "local feel"?
6. How each tool **authenticates and stores credentials** remotely, and how to keep
   those credentials inside the capsule boundary (per #3).

**Layer 4 — Deterministic provisioning & lifecycle engine (the heart of the ask)**
7. Survey **declarative + if/then orchestration**: cloud-init, Ansible, NixOS, systemd
   units/`.mount`/`.automount`, Terraform/Proxmox provider, and **idempotent
   reconcilers**. Recommend a stack where **provision → health-gate → next step** is
   explicit, auditable, and re-runnable.
8. Patterns for a **preflight/health-check DSL** (the if/then logic): how mature systems
   express "assert X, else remediate Y, else fail closed" (serverspec/goss/`testinfra`,
   systemd readiness gates, health probes). Produce a recommended **check-gate model**.

**Layer 5 — Blackwell / hybrid GPU**
9. Current best-practice (2025–2026) for **Blackwell GPU passthrough/MIG** on a
   KVM/Proxmox host, driver/CUDA baseline, and partitioning a card across
   tiers/tenants safely.
10. **Hybrid inference routing**: local model serving (vLLM/TGI/Ollama-class) + cloud
    fallback, and where a prompt-optimizer ("Clyffy") sits to apply **COSTAR /
    prompt best-practices** before cloud calls — with evidence on what measurably
    improves output quality/cost.

**Layer 6 — The reusable buildplan methodology (becomes our baseline)**
11. From the above, define a **layered, multi-pass buildplan format**: the ordered
    **passes** (discover → spec → gate → provision → verify → capture), the
    **requirement gates** each pass must clear, and the **minimal-memory documentation
    contract** (what must be written down at each pass so no AI ever "works blind").
    This must extend our existing parking-lot/touchpoint capture.

**Layer 7 — Intelligence-layer capture**
12. How to project the buildplan + per-run evidence into **SurrealDB (graph) + Qdrant
    (retrieval) sharing with Postgres (truth)** so the master-Clyffy intelligence layer
    can answer "how do we build a workspace" from captured reality, not prose.

## (S) Style & (R) Response format

Deliver a **cited report** with:
- An **executive recommendation** per layer (one chosen approach + why + rejected
  alternatives).
- **Decision matrices** (option × criteria: safety, determinism, reproducibility, cost,
  maturity, fit).
- A **concrete reference template**: the VM+Capsule config series as **declarative
  manifests + an explicit if/then check-gate sequence** (pseudo-code/flow acceptable;
  must be scriptable, not narrative).
- The **buildplan baseline** (Layer 6) as a reusable spec with named passes and gate
  checklists.
- A **risk/threat section** focused on secret leakage and agent blast-radius, with
  mitigations.
- Every non-obvious claim backed by a **primary source** (vendor docs, RFCs, maintained
  OSS, dated benchmarks). Flag opinion or unverified items.

## (T) Tone & (A) Audience

Senior-infra precise, boring-by-design, decision-oriented. Audience = the operator
(Hades) + the AI agents that will execute the buildplan. Prefer deterministic,
auditable mechanisms over anything that needs runtime AI judgment.

## Constraints & non-goals

- **Deterministic over AI**: provisioning/lifecycle must be scripted if/then logic; AI
  is a workspace consumer only.
- **No secret ever reaches the user or an AI agent's context.** Fail closed.
- **2026 production bar**: prefer current, maintained, primary-sourced approaches; note
  version/date sensitivity (Blackwell drivers, MCP/OAuth2.1, code-server).
- **Out of scope**: billing implementation, customer UI styling (ColdLight), and the MCP
  wire protocol itself — except where they impose a requirement on the VM/Capsule template.
