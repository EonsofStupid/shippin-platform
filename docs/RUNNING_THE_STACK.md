# Running the WardenClyffe stack (devstation)

> **Development reference, not a release installer.** Some historical names and
> paths below predate the Shippin migration. The customer API now requires a
> dedicated restricted PostgreSQL login; it will not use the operator DB URL.

The operator API requires explicitly configured `SHIPPIN_OPERATOR_USER` and
`SHIPPIN_OPERATOR_PASS` environment variables. Use installation-specific
credentials and supply them through your secret-management process; there is no
default operator login. This development bootstrap is not a production
first-owner enrollment or a complete password policy.

Apply the migrations through `0009_customer_access.sql` on PostgreSQL 18 as an
appropriately privileged migration owner. Migration 0009 creates the non-login
`shippin_customer_api` group role and grants only the columns used by the
customer account API. Provision a separate login using your installation's
secret-management process, grant that login the group role, and set
`CLYFFE_DB_URL` to its connection string. The login must not own the tables, be
a superuser, bypass row security, create roles/databases, replicate, or inherit
other roles. The API rejects unsafe role configuration at startup.

For example, a database administrator can create the login and assign a password
interactively in `psql`; use a different login name if your installation needs
one. The password is not part of the migration or repository:

```sql
CREATE ROLE shippin_customer_login LOGIN;
GRANT shippin_customer_api TO shippin_customer_login;
\password shippin_customer_login
```

The customer API reads the `SHIPPIN_AUTH_COOKIE` cookie name used by
`shippin-auth` (default `shippin_session`). Serve both through a deliberately
configured same-origin HTTPS gateway; the current development CORS middleware
does not implement cross-origin credentialed sessions. Account and order routes
require an active verified customer session. `/api/clyffe/orders` requires a
`tenant_id` belonging to that subject. At present any membership role can read
that organization's order metadata; granular billing permissions are not yet
implemented. The existing `shippin-auth` login and other API routes still require
production security review.

Migration 0009 revokes default PUBLIC execution of identity functions. Any
separate identity-service database login must receive explicit EXECUTE grants
for the functions it uses; existing data and memberships are not removed. The
installation-specific network, account and IP seeds were removed from fresh
migrations 0004, 0006 and 0007; previously applied data is left untouched.

This is the live foundation: **Warden server (Go) + Clyffe portal API (Go) +
self-hosted Postgres + React console** with a Supabase-inspired, Go-backed data
management layer. Warden is the operator plane; Clyffe is the customer plane.

## Components built

```
data/schema/sql/0001_init.sql      canonical Postgres schema (8 schemas)
services/warden-api/               Go server (chi + pgx), port :8081 — OPERATOR plane
  internal/platform                config, pgx pool, http helpers, CORS
  internal/fleet                   shippin_infra.resources (workspaces)
  internal/automation              plan -> action_request -> provision
  internal/audit                   append-only shippin_audit.events
  internal/dbadmin                 Supabase-style data API (our own, in Go)
  internal/clyffy                  Clyffy orchestrator (reads captured foundation)
services/clyffe-api/               Go server (chi + pgx), port :8082 — CUSTOMER plane
  internal/platform                config, pgx pool, http helpers, CORS
  internal/portal                  customer-safe reads: accounts + orders
apps/console/                      React + Vite + React-Aria-Components, port :5173
  src/styles/tokens.css            OKLCH color · rem space · clamp() type
  src/ds                           design-system primitives over RAC
  src/views                        Workspaces · DataBrowser · OrderDevstation
```

Boundary: `clyffe-api` holds no infrastructure authority and reads only
customer-safe data (`shippin_core.tenants`, `clyffe_*`); it never touches
`shippin_infra` / `shippin_audit`.

## Prerequisites (already provisioned on warden-devstation-01)

- PostgreSQL 18 local, role `warden`, db `wardenclyffe`.
- Schema applied: `psql -h 127.0.0.1 -U warden -d wardenclyffe -f data/schema/sql/0001_init.sql`
- Go 1.26, Node 24.

## Run

```bash
bash scripts/dev/run-stack.sh
```

Starts warden-api on `:8081`, clyffe-api on `:8082`, and the console on `:5173`
(the console proxies `/api` to warden-api). Override ports/DB with
`WARDEN_API_ADDR`, `CLYFFE_API_ADDR`, `WARDEN_DB_URL`, `CLYFFE_DB_URL`.

Clyffe-api surface (customer plane):

| Method | Path | Purpose |
|---|---|---|
| GET | `/healthz` | liveness + db ping |
| GET | `/api/clyffe/home` | portal overview: accounts + order roll-up |
| GET | `/api/clyffe/accounts` | customer accounts (tenants) |
| GET | `/api/clyffe/orders?tenant_id=` | orders for one authorized tenant |

## View it from your machine

The devstation has no public route. SSH local-forward the console port:

```bash
ssh -L 5173:127.0.0.1:5173 warden-devstation
# then open http://127.0.0.1:5173
```

## API surface (warden-api)

| Method | Path | Purpose |
|---|---|---|
| GET  | `/healthz` | liveness + db ping |
| GET  | `/api/warden/workspaces` | fleet resources |
| GET  | `/api/warden/workspaces/{id}` | one resource |
| POST | `/api/warden/provision` | order/provision a workspace (action_request + resource + audit) |
| GET  | `/api/warden/data/schemas` | managed schemas + table counts |
| GET  | `/api/warden/data/schemas/{schema}/tables` | tables + columns/rows/size |
| GET  | `/api/warden/data/tables/{schema}/{table}` | columns (type, nullable, pk) |
| GET  | `/api/warden/data/tables/{schema}/{table}/rows?limit=&offset=` | paginated rows |
| POST | `/api/warden/data/query` | **read-only** SQL (read-only tx + statement timeout) |

Safety: the data layer only exposes the 8 WardenClyffe schemas; mutating SQL
fails in the read-only transaction; non-managed schemas return 403.
