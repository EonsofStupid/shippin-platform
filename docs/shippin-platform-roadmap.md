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
