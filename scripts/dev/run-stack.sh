#!/usr/bin/env bash
# Run the WardenClyffe pre-release stack on the devstation:
#   shippin-api :8081 · clyffe-api :8082 · console :5173
#
# Loads secrets/proxmox.env when present (never printed).
# PID files under .dev-pids/ — stop with scripts/dev/stop-stack.sh or make stop
#
# Laptop view (only when working on this host's console):
#   ssh -N -L 5173:127.0.0.1:5173 warden-devstation
#   http://127.0.0.1:5173/login
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PIDDIR="${SHIPPIN_PID_DIR:-$ROOT/.dev-pids}"
mkdir -p "$PIDDIR"

export SHIPPIN_REPO_ROOT="${SHIPPIN_REPO_ROOT:-$ROOT}"
export SHIPPIN_DB_URL="${SHIPPIN_DB_URL:-postgres://shippin:shippin_dev_local@127.0.0.1:5432/shippin_mesh?sslmode=disable}"
export SHIPPIN_API_ADDR="${SHIPPIN_API_ADDR:-:8081}"
: "${CLYFFE_DB_URL:?Set CLYFFE_DB_URL to a restricted customer API database login; see docs/RUNNING_THE_STACK.md}"
export CLYFFE_DB_URL
export CLYFFE_API_ADDR="${CLYFFE_API_ADDR:-:8082}"

# Load proxmox env into process without echoing
if [ -f "$ROOT/secrets/proxmox.env" ]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/secrets/proxmox.env" 2>/dev/null || true
  set +a
  echo "[dev] loaded secrets/proxmox.env (values not printed)"
else
  echo "[dev] WARN: no secrets/proxmox.env — Slice 0 inventory will be unconfigured"
fi

: "${SHIPPIN_OPERATOR_USER:?Set SHIPPIN_OPERATOR_USER to your chosen development operator name}"
: "${SHIPPIN_OPERATOR_PASS:?Set SHIPPIN_OPERATOR_PASS to your chosen development operator password}"
export SHIPPIN_OPERATOR_USER SHIPPIN_OPERATOR_PASS

echo "[dev] building shippin-api..."
( cd "$ROOT/services/shippin-api" && go build -o "$PIDDIR/shippin-api" ./cmd/shippin-api )
echo "[dev] building clyffe-api..."
( cd "$ROOT/services/clyffe-api" && go build -o "$PIDDIR/clyffe-api" ./cmd/clyffe-api )

# Free ports if leftovers
if command -v fuser >/dev/null 2>&1; then
  fuser -k 8081/tcp 2>/dev/null || true
  fuser -k 8082/tcp 2>/dev/null || true
fi
sleep 0.5

echo "[dev] starting shippin-api on $SHIPPIN_API_ADDR ..."
"$PIDDIR/shippin-api" >"$PIDDIR/shippin-api.log" 2>&1 &
echo $! >"$PIDDIR/shippin-api.pid"

echo "[dev] starting clyffe-api on $CLYFFE_API_ADDR ..."
"$PIDDIR/clyffe-api" >"$PIDDIR/clyffe-api.log" 2>&1 &
echo $! >"$PIDDIR/clyffe-api.pid"

cleanup() {
  if [ -f "$PIDDIR/shippin-api.pid" ]; then kill "$(cat "$PIDDIR/shippin-api.pid")" 2>/dev/null || true; fi
  if [ -f "$PIDDIR/clyffe-api.pid" ]; then kill "$(cat "$PIDDIR/clyffe-api.pid")" 2>/dev/null || true; fi
}
trap cleanup EXIT

echo "[dev] waiting for APIs..."
for i in 1 2 3 4 5 6 7 8 9 10; do
  if curl -sS -m 1 "http://127.0.0.1:${SHIPPIN_API_ADDR#:}/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 0.3
done
curl -sS "http://127.0.0.1:${SHIPPIN_API_ADDR#:}/healthz" || true
echo
curl -sS "http://127.0.0.1:${SHIPPIN_API_ADDR#:}/api/warden/proxmox/status" || true
echo

echo "[dev] starting console on :5173 ..."
echo "[dev] Slice 0 UI: http://127.0.0.1:5173/admin/proxmox (after login)"
echo "[dev] Sign in with your configured operator credentials."
cd "$ROOT"
[ -d node_modules ] || npm install --no-audit --no-fund
exec npm run dev
