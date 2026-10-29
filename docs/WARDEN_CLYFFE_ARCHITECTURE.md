# Shippin management architecture (legacy Warden/Clyffe implementation)

This repo implements two Shippin management surfaces using legacy module
names:

- **Warden** is the Shippin operator/server-control implementation.
- **Clyffe** is the Shippin customer portal, customer knowledge base, and tenant-safe service implementation.

They are two Shippin surfaces, not a product named WardenClyffe:

- Shippin owns infrastructure authority through the current Warden implementation.
- Shippin Estate/Clyffe consumes scoped Shippin APIs on behalf of customers.

**Wardenclyffe is specifically the Zuul zero-mesh product.** It owns private
connectivity, peer/device enrollment, mesh policy, reachability, relay, mesh
health, and mesh-service discovery. It does not own Shippin customer or
infrastructure management.

Canonical module labels:

- **Module 1: Warden** - Proxmox UI manager, operator control plane, Warden API,
  infrastructure graph, and MCP mesh observatory.
- **Module 2: Clyffe** - customer portal, customer knowledge base, tickets,
  CRM, and customer-safe service panel.

Use module labels in planning only. Durable code, APIs, schemas, and UI copy
should use Warden and Clyffe.

## Internal-First Goal

The first target is a turnkey Shippin management deployment for our own two servers.

For this phase, Proxmox VE is treated as the local infrastructure substrate:

- Proxmox provides KVM, LXC, storage, networking, backups, tasks, consoles, and cluster primitives.
- The Warden implementation provides Shippin's modern operator experience, automation layer, AI assistant, and API.
- The Clyffe implementation provides Shippin Estate's customer portal and knowledge base using tenant-safe Shippin APIs.

The long-term direction is that Shippin progressively absorbs the workflows we
need from Proxmox and modernizes them under Shippin product ownership. We are
not starting with a hard fork. We are starting with a Proxmox-native
compatibility adapter and control-plane implementation that can later become
deeper and more self-contained where it makes sense.

## Product Boundary

### Warden

Warden is for trusted operators and server owners.

Responsibilities:

- Discover and manage Proxmox nodes and clusters.
- Manage VMs, LXC containers, templates, backups, snapshots, networks, storage, and firewalls.
- Manage public IP ingress, domains, TLS, edge routes, and route health.
- Provide fleet health, logs, metrics, alerts, status pages, and issue scans.
- Run approved automation and provisioning workflows.
- Host the infrastructure AI assistant and MCP meshnode.
- Own credentials, audit logs, tenant boundaries, and infrastructure inventory.

Warden may talk directly to Proxmox APIs, local Linux tools, legacy-named
WardenClyffeNet and WardenClyffeDisk compatibility services, Qdrant,
PostgreSQL, and other internal services.

Warden also owns private operator workspaces:

- **Warden Operator Capsule** is the secret-sensitive shell for bootstrap and
  host/operator work.
- **Warden Devstation** is the private hosted VS Code/Cursor/Codex/Claude
  workstation for daily development, including a localhost-bound code-server
  browser IDE reached through SSH forwarding.

These are not customer workspaces.

### Clyffy MCP Orchestrator

Clyffy is the main MCP orchestrator and assistant runtime for Shippin. It
does not replace Warden authority or Clyffe customer boundaries. It organizes
the tool mesh, workspace memory, agent handoffs, Qdrant retrieval, SurrealDB
graph projections, and Clyffy Master assistant workflows.

The dedicated Clyffy Master runtime is the future `clyffy-master` service
tracked in `docs/CLYFFY_MCP_ORCHESTRATOR.md` and
`docs/MASTER_CLYFFY_ROLLOUT_PLAN.md`.

The boundary is:

```text
OPNsense = network truth
Authentik = human identity truth
Better Auth = app-local auth/session/API layer
Warden = infrastructure/control truth
Bifrost = controlled bridge/orchestration gateway
Clyffy = MCP orchestrator and assistant runtime
Postgres = product truth
Qdrant = retrieval memory
SurrealDB = graph/reasoning projection
```

### Clyffe

Clyffe is for customers.

Responsibilities:

- Show only customer-owned or assigned VMs, containers, services, tickets, invoices, and docs.
- Provide safe actions such as start, stop, restart, console access, rebuild from approved template, backup restore request, and support request.
- Provide a customer knowledge base and AI-assisted help surface.
- Provide customer tickets, account history, service notes, contact records, and CRM-style relationship tracking.
- Later provide **Clyffe Code**, a tenant-safe hosted coding workspace service
  derived from the Warden Devstation pattern but created from clean templates
  with no operator auth state or secrets.
- Track customer service ownership across VMs, containers, domains, backups, support plans, and managed operations.
- Never expose Proxmox directly.
- Never bypass Warden's API, permissions, audit log, or rate limits.

### Clyffe Customer Panel

The Clyffe panel should be built around customer operations, not infrastructure jargon.

Core customer views:

- Dashboard: active services, health, resource usage, open tickets, recent changes.
- Services: assigned VMs, containers, databases, domains, storage, backups, and status.
- Console: controlled console access only for assigned guests.
- Backups: available restore points and restore requests.
- Tickets: support requests, incidents, approvals, notes, attachments, and status.
- Knowledge base: Shippin docs, service guides, runbooks, and customer-safe AI answers.
- Account/CRM: contacts, organizations, service notes, plan details, invoices or billing-provider links.

Customer actions must be permission-scoped, rate-limited, and audit-logged. Any action that can destroy data, alter networking, change billing, or affect other tenants should become a request or approval workflow rather than direct execution.

### Warden Operator Panel

The Warden panel should be the command center for our servers.

Core operator views:

- Fleet: Proxmox nodes, native Warden nodes, health, metrics, alerts, and inventory.
- Edge/DNS: public IPs, forwards, domains, TLS certificates, edge routes, and route probes.
- Guests: VMs, LXC containers, templates, snapshots, backups, consoles, and ownership.
- Customers: tenants, contacts, assigned services, support state, notes, and account health.
- Tickets/CRM: queues, incidents, tasks, customer notes, internal notes, and escalation.
- Provisioning: templates, products, plans, resource limits, hooks, and approval flows.
- Automation: scheduled jobs, workflows, AI-proposed fixes, and run history.
- Knowledge: docs, runbooks, embeddings status, Qdrant collections, and AI sources.
- Audit: operator actions, customer actions, Proxmox tasks, API calls, and approval trails.

## Proxmox Strategy

The immediate strategy is to study Proxmox deeply and wrap it properly:

- Use the Proxmox REST API as the primary integration path.
- Use local `pvesh`, `qm`, and `pct` only where needed as host-local helpers or fallbacks.
- Model Proxmox tasks as first-class Warden jobs.
- Translate Proxmox users, pools, guests, storage, nodes, and permissions into Warden-owned records.
- Keep customer-facing actions scoped through Clyffe and Warden's tenant API.

The first useful integration surfaces are:

1. Node discovery and health.
2. VM and LXC inventory.
3. Guest lifecycle: start, stop, restart, shutdown.
4. Task polling and audit logging.
5. Template inventory and customer-safe provisioning.
6. Console access with tenant guardrails.
7. Backups, snapshots, and restores.

## Data Layer Options Under Review

This section is not a locked decision. These are evaluation tracks for the
Warden/Clyffe control plane and should be revisited after a direct fit check
against the Rust code, Go Warden work, Proxmox model, customer portal needs,
and the two-server pilot.

Current candidates:

- **MariaDB/MySQL-compatible stack** for database-management product surfaces,
  managed customer databases, and places where MySQL compatibility matters.
- **PostgreSQL-compatible stack** for relational control-plane records such as
  tenants, users, RBAC, inventory, Proxmox resources, audit logs, support
  records, tickets, CRM, billing links, and workflow state.
- **SurrealDB** for graph/document/agent-memory experiments and projections
  where its model genuinely helps.
- **Qdrant** for vector retrieval, project embeddings, semantic knowledge-base
  search, and Clyffy assistant retrieval.
- **Object storage** for screenshots, generated reports, backup metadata
  exports, docs assets, and customer artifacts.
- **Dragonfly/Redis-compatible cache or NATS** for sessions, queues, events,
  and long-running job coordination if those workloads need a separate layer.

WardenClyffeScale remains a legacy-named, separate database replication
product or managed service offering. It is not part of Wardenclyffe/Zuul and
does not decide Shippin's primary storage backend by itself.

## Auth Options Under Review

This section is also not a locked decision. The likely direction is
OIDC-first identity instead of custom auth inside the server manager, but the
provider choice remains open.

Candidate tracks:

- ZITADEL for modern self-hosted identity, organizations, projects, OIDC, and
  customer/operator separation.
- Authentik because the Go Warden repo already has Authentik work and
  templates.
- Keycloak as a heavier enterprise alternative.
- Ory as a composable alternative if we want to own more assembly.

For the internal two-server phase, Warden can start with a small local admin bootstrap, but the architecture should assume OIDC and tenant-aware permissions.

## AI Direction

Warden's AI system is an operator assistant first:

- Understand server inventory, Proxmox state, Shippin infrastructure docs, logs, incidents, and projects.
- Propose actions with risk levels.
- Require explicit approval before any write or destructive operation.
- Store embeddings in Qdrant.
- Keep personal/project memory separate from customer-visible Clyffe knowledge.

Clyffe's AI system is customer-safe:

- Answer docs and service questions.
- Explain assigned VM/container status.
- Open support tickets.
- Suggest safe actions.
- Escalate to Warden/operator when needed.

## Branding Rule

The customer and infrastructure management platform is presented as
**Shippin**:

- Shippin Operator for trusted server/operator control.
- Shippin Estate for the customer portal and knowledge base.
- Wardenclyffe only for the Zuul zero-mesh product.
- Warden/Clyffe filenames, schemas, and CLI names may remain during migration
  where compatibility requires them; they are not current product ownership.
