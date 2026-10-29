#!/usr/bin/env bash
# Keep the operator Vite console alive. TanStack Start can abort the process
# when an SSR stream exceeds 120s; this loop brings it back.
set -u
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PIDDIR="${SHIPPIN_PID_DIR:-$ROOT/.dev-pids}"
mkdir -p "$PIDDIR"
echo $$ >"$PIDDIR/console-supervisor.pid"

export SHIPPIN_CONSOLE_HOST="${SHIPPIN_CONSOLE_HOST:-127.0.0.1}"
export SHIPPIN_CONSOLE_PORT="${SHIPPIN_CONSOLE_PORT:-4176}"
export SHIPPIN_HMR_CLIENT_PORT="${SHIPPIN_HMR_CLIENT_PORT:-8081}"
export SHIPPIN_API_PROXY="${SHIPPIN_API_PROXY:-http://127.0.0.1:8081}"

cd "$ROOT"
while true; do
  echo "[console] starting vite ${SHIPPIN_CONSOLE_HOST}:${SHIPPIN_CONSOLE_PORT} $(date -Is)" >>"$PIDDIR/console.log"
  npx vite --host "$SHIPPIN_CONSOLE_HOST" --port "$SHIPPIN_CONSOLE_PORT" >>"$PIDDIR/console.log" 2>&1 &
  vite_pid=$!
  echo "$vite_pid" >"$PIDDIR/console.pid"
  wait "$vite_pid"
  code=$?
  echo "[console] vite exited ${code} $(date -Is); restarting in 1s" >>"$PIDDIR/console.log"
  sleep 1
done
