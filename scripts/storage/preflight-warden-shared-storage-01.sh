#!/usr/bin/env bash
set -euo pipefail

HOST="${1:-server1}"
VMID="${WARDEN_STORAGE_VMID:-117}"
REQUIRED_GIB="${WARDEN_STORAGE_REQUIRED_GIB:-500}"

ssh "$HOST" bash -s -- "$VMID" "$REQUIRED_GIB" <<'REMOTE'
set -euo pipefail

VMID="$1"
REQUIRED_GIB="$2"

echo "warden shared storage preflight"
echo "host=$(hostname)"
echo "candidate_vmid=$VMID"
echo

echo "proxmox storage status"
pvesm status
echo

AVAILABLE_KIB="$(pvesm status --storage local-lvm | awk 'NR==2 { print $6 }')"
AVAILABLE_GIB="$((AVAILABLE_KIB / 1024 / 1024))"
echo "local_lvm_available_gib=$AVAILABLE_GIB"
echo "required_available_gib=$REQUIRED_GIB"
if [ "$AVAILABLE_GIB" -lt "$REQUIRED_GIB" ]; then
  echo "ERROR: local-lvm does not have enough available space for the planned 400 GiB carve plus safety margin." >&2
  exit 1
fi
echo

if pct status "$VMID" >/dev/null 2>&1 || qm status "$VMID" >/dev/null 2>&1; then
  echo "ERROR: VMID $VMID is already allocated." >&2
  exit 1
fi
echo "vmid_available=true"
echo

echo "current lxcs"
pct list
echo

echo "current vms"
qm list
echo

echo "thin pool detail"
lvs --units g --nosuffix --separator ' ' -o vg_name,lv_name,lv_size,data_percent,metadata_percent
echo

echo "preflight_result=pass"
REMOTE
