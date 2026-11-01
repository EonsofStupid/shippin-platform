# Shippin platform release roadmap

Shippin's intended first release is a self-hosted primary panel that an operator
can set up on a server, then use to enroll additional Debian hosts or VMs and
manage Incus workloads. Proxmox is a distinct provider. Existing code and
separately maintained services are candidates for review; names or Git ancestry
alone do not establish compatibility, security or deployability.

## Current state

- `src/` contains the React/TanStack application; `services/` contains Go
  services. The Proxmox API routes require an operator bearer session and its
  client requires verified TLS, with an optional private CA.
- Lefthook and Biome currently cover the reviewed Proxmox frontend files only.
  Zod validates that feature's existing API responses; it is not a replacement
  for a provider-neutral server contract.
- The existing API still exposes other routes without a complete authorization
  model. Development configuration includes shared defaults. The current
  application is not ready to expose on a server or offer as a release.
- Feature branches, local changes, Warden sources, earlier API contracts, panel
  code and installer work require individual review and reconciliation. Preserve
  the existing work until the relevant behavior and history are accounted for.

## Delivery sequence

1. **Reconcile sources and contracts.** Inventory branches, working trees and
   candidate components; compare their behavior, configuration, dependencies
   and tests. Choose one owner for each service, schema and UI feature. Migrate
   deliberately, preserving compatibility and a rollback path. Update imports,
   callers and checks alongside any rename or move.
2. **Secure the primary.** Establish server-side authentication and authorization
   for every API route, including data access and provisioning. Remove shared
   factory credentials and installation-specific defaults. Add validated operator
   configuration, TLS, audit, durable state and recoverable first-owner setup.
   Test denial, revoked sessions and tenant/resource boundaries.
3. **Enroll infrastructure.** Generate a short-lived, single-node enrollment
   from the running primary. Bind a node-generated identity to it, support
   revocation and reconnect, and show host/VM roles and operator-defined tags.
   Existing disks, networks and cluster membership must not be modified silently.
4. **Operate providers.** Implement the host agent and Incus adapter first.
   Prove create, inspect, stop and remove against a disposable real host,
   including authorization, isolation, failure and recovery. Add and verify the
   Proxmox adapter separately, including task completion and certificate trust.
5. **Package and update.** Produce a reproducible Debian primary installation
   and versioned node artifacts. Define configuration and secret ownership,
   persistent data, backup/restore, artifact verification and rollback. Publish
   a tested pull/deploy procedure that reports the installed version and health
   after an update; verify from a clean checkout and a disposable server.
6. **Expand quality gates.** Extend Biome, schema validation and tests as each
   context is reviewed. CI must exercise the same checks and meaningful behavior;
   a hook, typecheck or mocked provider alone is not release evidence.

The primary installer, secondary enrollment, tenant isolation, live provider
operations and upgrade/recovery flow remain goals until their checks are recorded.
No production infrastructure should be changed as a side effect of these plans.

## Managed application and database projects

The intended customer experience is to select an organization and workspace,
name a project, add a managed PostgreSQL database, choose Prisma or Drizzle when
appropriate, and start from a template in a new GitHub repository or connect an
existing repository. The project can attach object storage buckets, functions,
an AI gateway, realtime and Redis or Dragonfly where those capabilities are
offered. Customers need actual project status and usage; staff need an authorized
control-plane view of capacity, lifecycle, configuration and service health.
These are product goals, not a list of services already implemented.

Every tier must provide a baseline of at least two instances and a 3 GB storage
allowance. **Policy decisions remain open:** what constitutes an instance, whether
the allowance belongs to an organization, workspace or user, whether 3 GB applies
per instance or is shared, and whether it covers database storage, buckets or
both. Resolve these before publishing a price, writing entitlement defaults or
implementing enforcement; do not substitute the unrelated historical workspace
and storage-volume tier numbers. Define GB versus GiB explicitly. Paid tiers may
grant more; the minimum is not automatically a maximum. Storage, database disk,
function invocations, realtime usage, AI tokens and cache usage need separate,
clearly described limits and cost policies.

Implement the capability in this order, checking the current code and contracts
again at each step:

1. **Identity and ownership.** Establish durable customer sessions and verified
   organization membership and roles; define the project/workspace relationship
   without assuming either is identical to an infrastructure provider's project.
   The existing operator-only data browser accesses platform records and is not
   a tenant database console. Test cross-organization reads, writes, logs,
   metrics, jobs and secret access with two real test tenants.
2. **Tier and entitlement authority.** Choose one control-plane owner for plan
   versions, subscription assignment, overrides, limits and usage. Expose
   authorized staff configuration and a customer-safe view of the effective
   allowance. Reserve capacity transactionally before provisioning, count pending
   operations, make retries idempotent, and test concurrent requests, failed
   provisioning, tier changes and recovery. Do not delete customer data on a
   downgrade. The answer to the open policy questions controls this step.
3. **Database lifecycle.** Add provider-neutral desired and observed states and
   a durable job/audit trail; use an Incus adapter first and review Proxmox as a
   separate adapter. Create isolated test databases with scoped credentials,
   configured resource limits, backup and tested restore. Show provisioning,
   failure, recovery and stale or unavailable metrics honestly. Test actual
   connectivity, isolation and restore against a disposable host.
4. **Project source and deployment.** Review the candidate console project and
   GitHub App flows before selective adoption, with explicit migration and no
   second source of truth. Support an installation-scoped GitHub App with the
   minimum required permissions, signed webhooks, revocation and short-lived
   credentials. Templates are versioned and reviewed. Existing repositories
   retain their existing ORM and migration history unless the user explicitly
   selects a change; never run a destructive reset as an import shortcut.
   See [GitHub permissions](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/choosing-permissions-for-a-github-app),
   [Prisma migrations](https://www.prisma.io/docs/orm/migrations/applying-a-migration)
   and [Drizzle migrations](https://orm.drizzle.team/docs/migrations).
5. **Attachable services and metering.** Implement storage, functions, realtime,
   AI gateway and cache as separate capabilities with authorized project
   attachments, service-specific credentials, health, actual usage and quota
   enforcement. Storage enforcement must cover concurrent uploads, multipart
   uploads, deletes and reconciliation. AI providers and endpoints are
   operator-configured; do not place provider secrets in the browser. Redis and
   Dragonfly require compatibility and durability checks for the selected use.
6. **End-to-end release proof.** On a clean disposable installation, create two
   allowed instances through the customer panel, deny a request beyond the
   configured entitlement, deploy a template and an existing repository, connect
   the database and enabled services, observe real usage, and verify isolation,
   failed jobs, restart recovery, backup/restore and upgrades. Include operator
   and customer authorization denials. Record actual artifacts and versions;
   mocked providers and green unit tests alone do not establish release readiness.
