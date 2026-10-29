#!/usr/bin/env bash
set -euo pipefail

COMMAND="${1:-status}"
CONFIG_FILE="${WARDEN_STORAGE_CLIENT_CONFIG:-/etc/warden/storage-client.env}"

if [[ -r "${CONFIG_FILE}" ]]; then
  # shellcheck source=/dev/null
  source "${CONFIG_FILE}"
fi

: "${WARDEN_STORAGE_SERVICE_NAME:=warden-shared-storage-01}"
: "${WARDEN_STORAGE_SMB_HOST:=10.0.0.117}"
: "${WARDEN_STORAGE_SMB_SHARE:=warden-storage}"
: "${WARDEN_STORAGE_MOUNT_PATH:=/mnt/warden/storage}"
if [[ -z "${WARDEN_STORAGE_SYMLINK_PATH+x}" ]]; then
  WARDEN_STORAGE_SYMLINK_PATH="~/warden-storage"
fi
: "${WARDEN_STORAGE_BROKER_MODE:=infisical-ssh}"
: "${WARDEN_STORAGE_BROKER_HOST:=warden-capsule}"
: "${WARDEN_INFISICAL_PROJECT_ID:=4a897376-3cbd-4aeb-8550-c7d3ed7ad113}"
: "${WARDEN_INFISICAL_ENV:=dev}"
: "${WARDEN_SHARED_STORAGE_INFISICAL_PATH:=/warden/shared-storage/01}"
: "${WARDEN_SHARED_STORAGE_SECRET_NAME:=WARDEN_SHARED_STORAGE_01_SMB_PASSWORD}"
: "${WARDEN_STORAGE_CRED_FILE:=/run/shippin-secrets/warden-shared-storage-01.smb}"

SERVICE_NAME="${WARDEN_STORAGE_SERVICE_NAME}"
SMB_HOST="${WARDEN_STORAGE_SMB_HOST}"
SMB_SHARE="${WARDEN_STORAGE_SMB_SHARE}"
MOUNT_PATH="${WARDEN_STORAGE_MOUNT_PATH}"
SYMLINK_PATH="${WARDEN_STORAGE_SYMLINK_PATH}"
BROKER_MODE="${WARDEN_STORAGE_BROKER_MODE}"
BROKER_HOST="${WARDEN_STORAGE_BROKER_HOST}"
INFISICAL_PROJECT_ID="${WARDEN_INFISICAL_PROJECT_ID}"
INFISICAL_ENV="${WARDEN_INFISICAL_ENV}"
INFISICAL_PATH="${WARDEN_SHARED_STORAGE_INFISICAL_PATH}"
SECRET_NAME="${WARDEN_SHARED_STORAGE_SECRET_NAME}"
CRED_FILE="${WARDEN_STORAGE_CRED_FILE}"

usage() {
  cat <<'USAGE'
Usage:
  warden-storage status
  warden-storage mount
  warden-storage unmount
  warden-storage path

Purpose:
  Mount Warden shared storage on a Linux/WSL client without printing secrets.

Defaults:
  Config:     /etc/warden/storage-client.env
  SMB host:   10.0.0.117
  SMB share:  warden-storage
  Local path: /mnt/warden/storage or /workspace/warden-storage

The mount command brokers the SMB password through warden-capsule/Infisical and
writes a temporary CIFS credentials file under /run/shippin-secrets.
USAGE
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    exec sudo -E "$0" "$@"
  fi
}

operator_user() {
  if [[ -n "${WARDEN_STORAGE_OPERATOR_USER:-}" ]]; then
    printf '%s\n' "${WARDEN_STORAGE_OPERATOR_USER}"
    return
  fi
  if [[ -n "${SUDO_USER:-}" && "${SUDO_USER}" != "root" ]]; then
    printf '%s\n' "${SUDO_USER}"
  elif id wardenop >/dev/null 2>&1; then
    printf '%s\n' "wardenop"
  else
    id -un
  fi
}

run_as_operator() {
  local user
  user="$(operator_user)"
  if [[ "${EUID}" -eq 0 && "${user}" != "root" ]]; then
    runuser -u "${user}" -- "$@"
  else
    "$@"
  fi
}

broker_password() {
  local remote_cmd
  if [[ "${WARDEN_STORAGE_ALLOW_ENV_SECRET:-0}" == "1" && -n "${WARDEN_SHARED_STORAGE_01_SMB_PASSWORD:-}" ]]; then
    printf '%s\n' "${WARDEN_SHARED_STORAGE_01_SMB_PASSWORD}"
    return
  fi
  if [[ -n "${WARDEN_SHARED_STORAGE_01_SMB_PASSWORD:-}" ]]; then
    echo "error: WARDEN_SHARED_STORAGE_01_SMB_PASSWORD is set but env-secret mode is not allowed" >&2
    exit 2
  fi

  remote_cmd="$(printf 'infisical secrets get %q --projectId %q --env %q --path %q --output json' \
    "${SECRET_NAME}" "${INFISICAL_PROJECT_ID}" "${INFISICAL_ENV}" "${INFISICAL_PATH}")"
  remote_cmd="${remote_cmd} | python3 -c 'import json,sys; data=json.load(sys.stdin); item=data[0] if isinstance(data,list) else data; value=item.get(\"secretValue\") or item.get(\"value\"); assert value; print(value)'"

  case "${BROKER_MODE}" in
    infisical-ssh)
      run_as_operator ssh -T "${BROKER_HOST}" "${remote_cmd}"
      ;;
    ssh-forced-command)
      run_as_operator ssh -T "${BROKER_HOST}" warden-storage-read-secret
      ;;
    infisical-local)
      eval "${remote_cmd}"
      ;;
    *)
      echo "error: unsupported WARDEN_STORAGE_BROKER_MODE=${BROKER_MODE}" >&2
      exit 2
      ;;
  esac
}

write_credentials() {
  local password
  password="$(broker_password)"
  if [[ -z "${password}" ]]; then
    echo "error: empty SMB password from broker" >&2
    exit 1
  fi

  install -d -m 0700 /run/shippin-secrets
  ln -sfn shippin-secrets /run/warden-secrets 2>/dev/null || true
  {
    printf 'username=%s\n' 'warden-share'
    printf 'password=%s\n' "${password}"
  } > "${CRED_FILE}"
  unset password
  chmod 0600 "${CRED_FILE}"
}

cleanup_credentials() {
  rm -f "${CRED_FILE}"
}

ensure_link() {
  local user home_dir link_path
  if [[ -z "${SYMLINK_PATH}" ]]; then
    return
  fi
  user="$(operator_user)"
  home_dir="$(getent passwd "${user}" | cut -d: -f6)"
  link_path="${SYMLINK_PATH/#\~/${home_dir}}"

  if [[ "${link_path}" != "${MOUNT_PATH}" ]]; then
    if [[ -L "${link_path}" || ! -e "${link_path}" ]]; then
      ln -sfn "${MOUNT_PATH}" "${link_path}"
      chown -h "${user}:${user}" "${link_path}" 2>/dev/null || true
    fi
  fi
}

cmd_mount() {
  if mountpoint -q "${MOUNT_PATH}"; then
    echo "already_mounted=1"
    ensure_link
    cmd_status
    return
  fi

  require_root mount
  install -d -m 0770 "${MOUNT_PATH}"

  local user uid gid
  user="$(operator_user)"
  uid="$(id -u "${user}")"
  gid="$(id -g "${user}")"
  chown "${uid}:${gid}" "${MOUNT_PATH}"

  write_credentials
  trap cleanup_credentials EXIT
  mount -t cifs "//${SMB_HOST}/${SMB_SHARE}" "${MOUNT_PATH}" \
    -o "credentials=${CRED_FILE},vers=3.1.1,seal,uid=${uid},gid=${gid},file_mode=0660,dir_mode=0770,noserverino"
  cleanup_credentials
  trap - EXIT
  echo "mounted=1"

  ensure_link
  cmd_status
}

cmd_unmount() {
  require_root unmount
  if mountpoint -q "${MOUNT_PATH}"; then
    umount "${MOUNT_PATH}"
    echo "unmounted=1"
  else
    echo "mounted=0"
  fi
  cleanup_credentials
}

cmd_status() {
  echo "service=${SERVICE_NAME}"
  echo "mount_path=${MOUNT_PATH}"
  echo "smb=//${SMB_HOST}/${SMB_SHARE}"
  if [[ -n "${SYMLINK_PATH}" ]]; then
    local user home_dir link_path
    user="$(operator_user)"
    home_dir="$(getent passwd "${user}" | cut -d: -f6)"
    link_path="${SYMLINK_PATH/#\~/${home_dir}}"
    echo "symlink_path=${link_path}"
  fi
  if mountpoint -q "${MOUNT_PATH}"; then
    local source size used available pct target
    read -r source size used available pct target < <(df -hP "${MOUNT_PATH}" | awk 'NR==2 {print $1, $2, $3, $4, $5, $6}')
    echo "mounted=1"
    echo "source=${source}"
    echo "size=${size}"
    echo "used=${used}"
    echo "available=${available}"
    echo "use_percent=${pct}"
    echo "path=${target}"
  else
    echo "mounted=0"
  fi
}

case "${COMMAND}" in
  mount) cmd_mount ;;
  unmount) cmd_unmount ;;
  status) cmd_status ;;
  path) printf '%s\n' "${MOUNT_PATH}" ;;
  help|-h|--help) usage ;;
  *)
    usage >&2
    exit 2
    ;;
esac
