#!/usr/bin/env bash
set -euo pipefail

: "${WARDEN_INFISICAL_PROJECT_ID:=4a897376-3cbd-4aeb-8550-c7d3ed7ad113}"
: "${WARDEN_INFISICAL_ENV:=dev}"
: "${WARDEN_SHARED_STORAGE_INFISICAL_PATH:=/warden/shared-storage/01}"
: "${WARDEN_SHARED_STORAGE_SECRET_NAME:=WARDEN_SHARED_STORAGE_01_SMB_PASSWORD}"
: "${WARDEN_STORAGE_BROKER_ORIGINAL_COMMAND:=warden-storage-read-secret}"

if [[ "${SSH_ORIGINAL_COMMAND:-}" != "${WARDEN_STORAGE_BROKER_ORIGINAL_COMMAND}" ]]; then
  echo "error: unauthorized storage broker command" >&2
  exit 2
fi

infisical secrets get "${WARDEN_SHARED_STORAGE_SECRET_NAME}" \
  --projectId "${WARDEN_INFISICAL_PROJECT_ID}" \
  --env "${WARDEN_INFISICAL_ENV}" \
  --path "${WARDEN_SHARED_STORAGE_INFISICAL_PATH}" \
  --output json |
  python3 -c 'import json,sys; data=json.load(sys.stdin); item=data[0] if isinstance(data,list) else data; value=item.get("secretValue") or item.get("value"); assert value; print(value)'
