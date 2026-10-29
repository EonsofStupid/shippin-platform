#!/usr/bin/env bash
set -euo pipefail

HOST="${1:-server1}"
CTID="${WARDEN_EDGE_CTID:-115}"
EDGE_IP="${WARDEN_EDGE_IP:-10.0.0.115}"
LEGACY_EDGE_VM="${WARDEN_LEGACY_EDGE_VM:-501}"
LEGACY_EDGE_IP="${WARDEN_LEGACY_EDGE_IP:-10.0.0.100}"

ssh "$HOST" bash -s -- "$CTID" "$EDGE_IP" "$LEGACY_EDGE_VM" "$LEGACY_EDGE_IP" <<'REMOTE'
set -euo pipefail

CTID="$1"
EDGE_IP="$2"
LEGACY_EDGE_VM="$3"
LEGACY_EDGE_IP="$4"

echo "warden caddy edge preflight"
echo "host=$(hostname)"
echo "candidate_ctid=$CTID"
echo "candidate_ip=$EDGE_IP"
echo "legacy_edge_vm=$LEGACY_EDGE_VM"
echo "legacy_edge_ip=$LEGACY_EDGE_IP"
echo

if pct status "$CTID" >/dev/null 2>&1 || qm status "$CTID" >/dev/null 2>&1; then
  echo "ctid_available=false"
  exit 1
fi
echo "ctid_available=true"

if ping -c 1 -W 1 "$EDGE_IP" >/dev/null 2>&1; then
  echo "candidate_ip_free=false"
  exit 1
fi
echo "candidate_ip_free=true"
echo

echo "available_templates"
find /var/lib/vz/template/cache -maxdepth 1 -type f -printf '%f\n' 2>/dev/null | sort | grep -E 'debian|ubuntu' || true
echo

echo "current_public_nat"
iptables-save 2>/dev/null | grep -E "dport (80|443|5432)|--dport (80|443|5432)|${LEGACY_EDGE_IP}|${EDGE_IP}" || true
echo

echo "legacy_caddy_status"
if qm status "$LEGACY_EDGE_VM" >/dev/null 2>&1; then
  qm guest exec "$LEGACY_EDGE_VM" -- /bin/bash -lc 'export HOME=/root; cd /opt/wardenclyffe-caddy-edge && docker compose ps && docker compose exec -T caddy-edge caddy validate --config /etc/caddy/Caddyfile' || true
else
  echo "legacy_edge_vm_missing=true"
fi
echo

echo "preflight_result=pass"
REMOTE
