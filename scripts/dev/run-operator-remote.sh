#!/usr/bin/env bash
# Operator console + API on the LIVE remote origin: http://<this-host>:8081
# Vite stays on loopback :4176; shippin-api reverse-proxies pages from :8081.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PIDDIR="${SHIPPIN_PID_DIR:-$ROOT/.dev-pids}"
mkdir -p "$PIDDIR"

export SHIPPIN_REPO_ROOT="${SHIPPIN_REPO_ROOT:-$ROOT}"
export SHIPPIN_DB_URL="${SHIPPIN_DB_URL:-postgres://shippin:shippin_dev_local@127.0.0.1:5432/shippin_mesh?sslmode=disable}"
export SHIPPIN_API_ADDR="${SHIPPIN_API_ADDR:-:8081}"
export SHIPPIN_CONSOLE_PROXY="${SHIPPIN_CONSOLE_PROXY:-http://127.0.0.1:4176}"
export SHIPPIN_CONSOLE_HOST="${SHIPPIN_CONSOLE_HOST:-127.0.0.1}"
export SHIPPIN_CONSOLE_PORT="${SHIPPIN_CONSOLE_PORT:-4176}"
export SHIPPIN_HMR_CLIENT_PORT="${SHIPPIN_HMR_CLIENT_PORT:-8081}"
export SHIPPIN_OPERATOR_USER="${SHIPPIN_OPERATOR_USER:-operator}"
export SHIPPIN_OPERATOR_PASS="${SHIPPIN_OPERATOR_PASS:-warden-dev}"

if [ -f "$ROOT/secrets/proxmox.env" ]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/secrets/proxmox.env" 2>/dev/null || true
  set +a
fi

echo "[operator] building shippin-api..."
( cd "$ROOT/services/shippin-api" && go build -o "$PIDDIR/shippin-api" ./cmd/shippin-api )

echo "[operator] starting console on ${SHIPPIN_CONSOLE_HOST}:${SHIPPIN_CONSOLE_PORT}..."
if command -v fuser >/dev/null 2>&1; then
  fuser -k "${SHIPPIN_CONSOLE_PORT}/tcp" 2>/dev/null || true
fi
(
  cd "$ROOT"
  exec npx vite --host "$SHIPPIN_CONSOLE_HOST" --port "$SHIPPIN_CONSOLE_PORT"
) >"$PIDDIR/console.log" 2>&1 &
echo $! >"$PIDDIR/console.pid"

for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
  if curl -sf -m 1 "http://127.0.0.1:${SHIPPIN_CONSOLE_PORT}/login" >/dev/null 2>&1; then
    echo "[operator] console ready"
    break
  fi
  sleep 0.4
done

echo "[operator] shippin-api on ${SHIPPIN_API_ADDR} (proxies UI from ${SHIPPIN_CONSOLE_PROXY})"
exec "$PIDDIR/shippin-api"
