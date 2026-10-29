#!/usr/bin/env bash
# warden-reconcile — GitOps-lite reconciler for the Warden VM.
# Pulls the repo, applies DB migrations, rebuilds shippin-api + console, and
# re-probes captured services. Intended to run on a systemd timer on the Warden
# VM so the VM "regularly loads changes via the repo".
#
# Spec: docs/SHIPPIN_CONNECTOR_PLUGIN_DESIGNATION.md (§ Repo-driven config)
# Install as a timer:  see the heredoc at the bottom (--install-timer).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BRANCH="${SHIPPIN_RECONCILE_BRANCH:-$(git -C "$ROOT" rev-parse --abbrev-ref HEAD)}"
DBURL="${SHIPPIN_DB_URL:-postgres://shippin:shippin_dev_local@127.0.0.1:5432/shippin_mesh?sslmode=disable}"
PSQL_DSN="${SHIPPIN_PSQL_DSN:-host=127.0.0.1 user=shippin dbname=shippin_mesh}"

log() { printf '[reconcile %s] %s\n' "$(date -u +%H:%M:%S)" "$*"; }

if [[ "${1:-}" == "--install-timer" ]]; then
  cat <<UNIT
# /etc/systemd/system/warden-reconcile.service
[Unit]
Description=WardenClyffe repo reconciler
After=network-online.target postgresql.service
[Service]
Type=oneshot
User=wardenop
WorkingDirectory=${ROOT}
ExecStart=${ROOT}/scripts/dev/warden-reconcile.sh
# /etc/systemd/system/warden-reconcile.timer
[Unit]
Description=Run WardenClyffe reconciler every 5 minutes
[Timer]
OnBootSec=2min
OnUnitActiveSec=5min
[Install]
WantedBy=timers.target
UNIT
  exit 0
fi

cd "$ROOT"

log "fetch + fast-forward ${BRANCH}"
git fetch --quiet origin "$BRANCH" || { log "git fetch failed (offline?) — continuing with local"; }
BEFORE="$(git rev-parse HEAD)"
git merge --ff-only "origin/${BRANCH}" 2>/dev/null || log "no fast-forward (local ahead or detached) — continuing"
AFTER="$(git rev-parse HEAD)"
[[ "$BEFORE" == "$AFTER" ]] && log "repo unchanged ($AFTER)" || log "advanced ${BEFORE:0:8} -> ${AFTER:0:8}"

log "apply DB migrations (idempotent)"
for f in data/schema/sql/*.sql; do
  log "  schema: $f"
  PGPASSWORD="${SHIPPIN_DB_PASSWORD:-shippin_dev_local}" psql "$PSQL_DSN" -v ON_ERROR_STOP=1 -q -f "$f"
done
for f in data/seed/*.sql; do
  log "  seed: $f"
  PGPASSWORD="${SHIPPIN_DB_PASSWORD:-shippin_dev_local}" psql "$PSQL_DSN" -v ON_ERROR_STOP=1 -q -f "$f"
done

log "rebuild shippin-api"
( cd services/shippin-api && go build -o /tmp/shippin-api ./cmd/shippin-api )

log "rebuild console"
( cd apps/console && [ -d node_modules ] || npm install --no-audit --no-fund; npm run build >/dev/null )

log "reconcile complete (restart services / re-probe handled by supervisor)"
