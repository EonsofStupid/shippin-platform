#!/usr/bin/env bash
# Stop processes started by run-stack / manual dev binaries on this host.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PIDDIR="${SHIPPIN_PID_DIR:-$ROOT/.dev-pids}"

stop_pidfile() {
  local f="$1"
  if [ -f "$f" ]; then
    local pid
    pid="$(cat "$f" 2>/dev/null || true)"
    if [ -n "${pid:-}" ] && kill -0 "$pid" 2>/dev/null; then
      echo "stopping pid $pid ($f)"
      kill "$pid" 2>/dev/null || true
      sleep 0.5
      kill -9 "$pid" 2>/dev/null || true
    fi
    rm -f "$f"
  fi
}

if [ -d "$PIDDIR" ]; then
  stop_pidfile "$PIDDIR/shippin-api.pid"
  stop_pidfile "$PIDDIR/clyffe-api.pid"
  stop_pidfile "$PIDDIR/console.pid"
fi

# Best-effort by port (devstation only)
for port in 8081 8082 5173; do
  if command -v fuser >/dev/null 2>&1; then
    fuser -k "${port}/tcp" 2>/dev/null || true
  fi
done

echo "stack stop attempted."
