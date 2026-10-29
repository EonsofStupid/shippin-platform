#!/usr/bin/env bash
# Prove Slice 0 endpoints are up. Does not perform start/stop (operator chooses guest).
set -euo pipefail
API="${SHIPPIN_API_URL:-http://127.0.0.1:8081}"
CONSOLE="${SHIPPIN_CONSOLE_URL:-http://127.0.0.1:5173}"

echo "== verify-slice0 =="
echo "API=$API"

fail=0
check() {
  local name="$1" url="$2"
  code=$(curl -sS -m 5 -o /tmp/wc-verify.json -w '%{http_code}' "$url" || echo fail)
  if [ "$code" = "200" ] || [ "$code" = "503" ]; then
    echo "  OK  $name ($code)"
    head -c 200 /tmp/wc-verify.json 2>/dev/null; echo
  else
    echo "  FAIL $name ($code)"
    fail=1
  fi
}

check "healthz" "$API/healthz"
check "proxmox status" "$API/api/warden/proxmox/status"
check "backup status" "$API/api/warden/proxmox/backup/status"

if curl -sS -m 2 -o /dev/null -w '' "$API/api/warden/proxmox/status" 2>/dev/null; then
  configured=$(python3 -c "import json; print(json.load(open('/tmp/wc-verify.json')).get('configured',False))" 2>/dev/null || echo false)
  if [ "$configured" = "True" ] || [ "$configured" = "true" ]; then
    check "proxmox guests" "$API/api/warden/proxmox/guests"
  else
    echo "  SKIP guests (Proxmox not configured — fill secrets/proxmox.env)"
  fi
fi

code=$(curl -sS -m 3 -o /dev/null -w '%{http_code}' "$CONSOLE/login" || echo fail)
if [ "$code" = "200" ]; then
  echo "  OK  console /login ($code)"
else
  echo "  WARN console /login ($code) — is make stack / npm run dev up?"
fi

if [ "$fail" -ne 0 ]; then
  echo "VERIFY FAILED"
  exit 1
fi
echo "VERIFY OK (configure Proxmox token for full inventory)"
