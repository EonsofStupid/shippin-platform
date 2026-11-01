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
- Lefthook and Biome cover a limited, reviewed set of frontend files, including
  Proxmox and staff data/fleet callers. Zod validates the Proxmox feature's
  existing API responses; it is not a provider-neutral server contract.
- The existing API still exposes other routes without a complete authorization
  model. Development configuration includes shared defaults. The current
  application is not ready to expose on a server or offer as a release.
- Customer home, accounts and orders use the existing customer session and
  membership records with a restricted database role. Any membership role can
  currently read its organization's order metadata; finer billing permissions
  and the broader customer identity and enrollment system require release review.
- Internal assistant inventory and probe routes require the existing operator
  bearer session. This does not establish a separate AI service identity.
- Installation-specific network, account and public-IP seeds have been removed
  from fresh migrations 0004, 0006 and 0007. Previously applied data is not
  removed; review existing installations before release.
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
allowance. The owner requested a **configurable default of five**, with the
association configurable and a project as its default. It remains unclear whether
five means instances or GB; do not implement that value until clarified. This is
product direction, not an implemented limit. **Policy decisions remain open:**
what constitutes an instance, who may change each limit or association, whether
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

### Guided connections for external applications

**Planned; no connection endpoint, network-policy reconciler, customer guide or
provider integration is implemented yet.** After an authorized customer creates
an instance, its project should offer a **Connect an application** flow for
Lovable, Replit, Bolt and a generic client. Show the selected organization,
workspace, project, environment and instance before changing access. Let users
change the association only where their role and the owning resources permit it;
the server must validate every relationship. Show actual service readiness and
explain when a step is unavailable, rather than presenting placeholders as live.

Provide these routes only as their underlying services become available:

| Connection path | Intended guide | Evidence and limits |
| --- | --- | --- |
| Server-side PostgreSQL | Give an authorized user an instance-scoped connection endpoint and a separately managed credential for a server or serverless function. Guide Replit users to project and published-app secrets, as applicable; support Prisma, Drizzle and ordinary drivers when verified. A browser-only app needs a backend or API. | Replit documents [secrets](https://docs.replit.com/core-concepts/project-editor/app-setup/secrets) for connection strings and separate published-app settings. Do not overwrite an existing `DATABASE_URL` or infer that a provider's built-in database dashboard will manage an external database. |
| HTTPS project API | Where Shippin implements a tenant-safe HTTPS application API, explain how to use it from a frontend or server and how authentication, permissions and CORS differ. Guide Lovable users through a supported custom connector or direct API integration. | Lovable documents [API integrations](https://docs.lovable.dev/integrations/any-api) and [custom connectors](https://docs.lovable.dev/integrations/create-connector). These require a reachable API; they do not make a bare PostgreSQL database an HTTP API. |
| Bolt | Offer the generic HTTPS or server-side path when the particular app runtime supports it; provide exact instructions only after testing an actual Bolt project. | Bolt documents a [Supabase-specific integration](https://support.bolt.new/bolt-cloud/database/advanced). PostgreSQL compatibility alone does not establish that Shippin works with that native integration, which has its own plan and migration requirements. |
| Private network | Offer private networking or a supported tunnel when the application environment can use it. Explain the connection requirements and limitations. | Availability depends on the installation, host adapter and external runtime; do not assume Lovable's public API connector can reach a private-only endpoint. |

**Access and firewall ownership:** start with no external database access. The
installation operator configures available network capabilities; an authorized
customer requests the access mode and allowed source networks for their own
instance. The control plane owns policy intent and audit records; a trusted host
agent/provider adapter applies and reconciles only the rules it owns. Report
pending, applied, failed and stale observations separately. Rules must not flush
or overwrite unrelated host, cluster or customer firewall rules. Define IPv4,
IPv6, NAT, DNS, port and provider support explicitly. Use the specific published
source ranges only for the traffic path to which a provider says they apply;
do not bake them into code or assume a runtime has a stable address. Lovable
documents [gateway connector ranges](https://docs.lovable.dev/integrations/security#ip-allowlisting),
which do not automatically describe other Lovable traffic. When a provider has
no suitable fixed range, explain supported alternatives such as an authenticated
HTTPS API or private tunnel; never silently broaden a firewall rule.

For Incus, choose the enforcement point from the actual network topology. Its
[bridge ACL limitations](https://linuxcontainers.org/incus/docs/main/howto/network_acls/#bridge-limitations)
mean a network-level bridge ACL alone does not provide intra-bridge isolation;
NIC-level ACLs and OVN have different behavior. A
[proxy device](https://linuxcontainers.org/incus/docs/main/reference/devices_proxy/)
can also change which client address the database sees, depending on its mode.
Test source restrictions at the point that sees the real source, including the
edge and any NAT. The older console's candidate host firewall code uses an
IPv4-only edge and workspace-tagged ACL rules; review its partial-apply and
recovery behavior before adopting it for tenant databases or IPv6.

**Credentials and transport:** issue per-instance, least-privilege, revocable
credentials with separate application and migration permissions; do not expose a
database password or privileged service key in frontend code, chat prompts,
repositories, logs, analytics or error messages. Show a secret only to an
authorized user when needed and provide rotation and revocation. Require encrypted
connections, with server certificate and hostname verification supported by the
selected client; document private CA distribution when applicable. PostgreSQL's
[`hostssl` rules](https://www.postgresql.org/docs/current/auth-pg-hba-conf.html)
and [client TLS verification](https://www.postgresql.org/docs/current/libpq-ssl.html)
are separate from host firewall rules. A browser-safe API key is not a database
credential and is safe to publish only when the API actually enforces end-user
authorization and data isolation.

**Implementation and release gates:**

1. Establish the customer identity, project/instance ownership and entitlements
   in the preceding steps. Define one versioned server contract for access grants,
   endpoint readiness and connection policies before adding UI schemas or routes.
2. Implement instance-scoped credentials, TLS, network policy reconciliation and
   audited lifecycle actions using the Incus host agent first; add a separate
   Proxmox adapter later. A failed policy update must not be reported as active.
3. Build the guided flow against real authorized APIs, with provider-specific
   instructions, masked secrets, source-network entry and clear error states.
   Distinguish a successful probe from Shippin from a successful connection made
   by the external application. Diagnostic endpoints must not accept arbitrary
   customer-supplied target URLs or expose internal network services.
4. On disposable infrastructure, verify allowed and denied IPv4/IPv6 paths,
   private and public modes, TLS hostname failures, credential rotation and
   revocation, cross-tenant denial, concurrent policy changes, agent restart,
   failed apply and rollback without disturbing unrelated rules. Test real
   development and published applications separately. Record the provider,
   runtime, date and supported features for each actual Lovable, Replit and Bolt
   test; label any untested path unavailable or experimental.
