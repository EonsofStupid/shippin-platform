#!/usr/bin/env bash
# Bootstrap this checkout for pre-release work. Safe to re-run.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

echo "== WardenClyffe pre-release setup =="
echo "ROOT=$ROOT"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "ERROR: missing required tool: $1" >&2
    exit 1
  fi
  local ver
  case "$1" in
    go) ver="$(go version 2>&1)" ;;
    *) ver="$("$1" --version 2>&1 | head -1)" ;;
  esac
  echo "  ok $1 ($ver)"
}

echo "-- tools --"
need go
need node
need npm
need psql
need curl
command -v git >/dev/null && echo "  ok git ($(git --version 2>&1))" || true

echo "-- secrets templates --"
mkdir -p "$ROOT/secrets"
if [ ! -f "$ROOT/secrets/proxmox.env" ] && [ -f "$ROOT/secrets/proxmox.env.example" ]; then
  cp "$ROOT/secrets/proxmox.env.example" "$ROOT/secrets/proxmox.env"
  echo "  created secrets/proxmox.env (FILL TOKEN — not committed)"
else
  echo "  secrets/proxmox.env already present or no example"
fi

echo "-- node modules --"
if [ ! -d node_modules ]; then
  npm install --no-audit --no-fund
else
  echo "  node_modules present"
fi

echo "-- go modules --"
( cd services/shippin-api && go mod download )
( cd services/clyffe-api && go mod download )
echo "  go mods downloaded"

echo "-- postgres --"
if pg_isready -h 127.0.0.1 -p 5432 >/dev/null 2>&1; then
  echo "  postgres accepting connections"
  if psql "postgres://shippin:shippin_dev_local@127.0.0.1:5432/shippin_mesh?sslmode=disable" -c 'select 1' >/dev/null 2>&1; then
    echo "  shippin_mesh db reachable as shippin"
  else
    echo "  WARN: cannot query shippin_mesh as shippin (check role/db/password)"
  fi
else
  echo "  WARN: postgres not ready on 127.0.0.1:5432"
fi

echo "-- build binaries once --"
( cd services/shippin-api && go build -o /tmp/shippin-api ./cmd/shippin-api )
( cd services/clyffe-api && go build -o /tmp/clyffe-api ./cmd/clyffe-api )
echo "  /tmp/shippin-api /tmp/clyffe-api"

echo
echo "Next:"
echo "  1. Edit secrets/proxmox.env with a real Proxmox API token"
echo "  2. make stack          # start APIs + console"
echo "  3. make verify-slice0  # health + proxmox status"
echo "  4. Login UI: http://127.0.0.1:5173/login  (operator / warden-dev)"
echo "  5. Slice 0:  http://127.0.0.1:5173/admin/proxmox"
echo
echo "Read docs/STATE.md for living pre-release status."
echo "OK setup complete."
