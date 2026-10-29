# Secrets (never commit real values)

This directory is gitignored except `README.md` and `*.example` files.

## Infisical (preferred on this host)

CLI is expected to be logged in (`infisical user`). Project id defaults to the
Clyffy/Warden Infisical project used by storage client.

```bash
make sync-secrets
# writes secrets/proxmox.env from Infisical keys (never commits)
```

**Required Infisical keys** (env `dev`, path `/`):

| Infisical key | Maps to |
|---------------|---------|
| `PROXMOX_TOKEN_ID` | `PROXMOX_TOKEN_ID` |
| `PROXMOX_TOKEN_SECRET` | `PROXMOX_TOKEN_SECRET` |
| `PROXMOX_HOST` (optional) | else `SERVER1_TAILNET_IP` or `10.0.0.1` |
| `PROXMOX_NODE` (optional) | else `server1` |
| `PROXMOX_PORT` (optional) | else `8006` |
| `PROXMOX_VERIFY_TLS` (optional) | else `false` |

If tokens are missing in Infisical, create them once (PVE API token UI → set in Infisical), then re-run `make sync-secrets`.

## Proxmox (Slice 0 — required for live inventory)

```bash
make sync-secrets
# or manual:
cp secrets/proxmox.env.example secrets/proxmox.env
# Edit secrets/proxmox.env — fill token from Infisical / PVE token UI
```

| Variable | Safe to log? | Notes |
|----------|--------------|--------|
| `PROXMOX_HOST` | yes | e.g. `10.0.0.1` |
| `PROXMOX_PORT` | yes | `8006` |
| `PROXMOX_NODE` | yes | e.g. `server1` |
| `PROXMOX_VERIFY_TLS` | yes | lab often `false` |
| `PROXMOX_TOKEN_ID` | **no** | `user@realm!tokenid` |
| `PROXMOX_TOKEN_SECRET` | **no** | secret |

Create token in Proxmox: Datacenter → Permissions → API Tokens.  
Privilege Separation often off for early automation; tighten later.

warden-api loads `secrets/proxmox.env` automatically when  
`WARDEN_REPO_ROOT` points at the repo (default on devstation).

## Operator bootstrap (console login)

Dev defaults (override in env for real deploys):

```text
WARDEN_OPERATOR_USER=operator
WARDEN_OPERATOR_PASS=warden-dev
```

## Storage client (optional)

`warden-storage-client.env` — see example if present.

## Production

Infisical → keyring `/run/shippin-secrets` (`/run/warden-secrets` is a compatibility symlink). Never paste secrets into chat or git.

## S3 backup connector (Slice 1)

Save from `/admin/proxmox` → **Backup**, or:

```bash
cp secrets/s3.env.example secrets/s3.env
# fill S3_* + RESTIC_PASSWORD; never commit
```

R2 / MinIO / AWS / B2 all work (S3 API). The connector is also mirrored to `/run/shippin-secrets/s3.env`.
