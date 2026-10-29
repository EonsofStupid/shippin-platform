#!/usr/bin/env bash
set -euo pipefail

PROFILE="${1:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLIENT_SCRIPT="${SCRIPT_DIR}/warden-storage-client.sh"
COMMAND_PATH="${WARDEN_STORAGE_COMMAND_PATH:-/usr/local/bin/warden-storage}"
CONFIG_DIR="${WARDEN_STORAGE_CONFIG_DIR:-/etc/warden}"
CONFIG_FILE="${CONFIG_DIR}/storage-client.env"
SERVICE_FILE="/etc/systemd/system/warden-storage.service"

usage() {
  cat <<'USAGE'
Usage:
  sudo scripts/storage/install-warden-storage-client.sh local-wsl
  sudo scripts/storage/install-warden-storage-client.sh devstation

Installs the Warden shared-storage client command and a non-secret client
config. The SMB password remains brokered from the operator capsule/Infisical.

Profiles:
  local-wsl    /mnt/warden/storage with ~/warden-storage symlink
  devstation   /workspace/warden-storage with systemd boot mount
USAGE
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "error: run with sudo/root" >&2
    exit 1
  fi
}

detect_operator_user() {
  if [[ -n "${WARDEN_STORAGE_OPERATOR_USER:-}" ]]; then
    printf '%s\n' "${WARDEN_STORAGE_OPERATOR_USER}"
  elif [[ -n "${SUDO_USER:-}" && "${SUDO_USER}" != "root" ]]; then
    printf '%s\n' "${SUDO_USER}"
  elif id wardenop >/dev/null 2>&1; then
    printf '%s\n' "wardenop"
  else
    echo "error: set WARDEN_STORAGE_OPERATOR_USER" >&2
    exit 1
  fi
}

write_config() {
  local mount_path symlink_path broker_host broker_mode operator_user
  operator_user="$(detect_operator_user)"

  case "${PROFILE}" in
    local-wsl)
      mount_path="/mnt/warden/storage"
      symlink_path="~/warden-storage"
      broker_host="warden-capsule"
      broker_mode="infisical-ssh"
      ;;
    devstation)
      mount_path="/workspace/warden-storage"
      symlink_path=""
      broker_host="warden-storage-broker"
      broker_mode="ssh-forced-command"
      ;;
    help|-h|--help|"")
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac

  install -d -m 0755 "${CONFIG_DIR}"
  cat > "${CONFIG_FILE}" <<EOF
# Warden shared-storage client configuration.
# This file intentionally contains no secrets.
WARDEN_STORAGE_SERVICE_NAME='warden-shared-storage-01'
WARDEN_STORAGE_OPERATOR_USER='${operator_user}'
WARDEN_STORAGE_SMB_HOST='10.0.0.117'
WARDEN_STORAGE_SMB_SHARE='warden-storage'
WARDEN_STORAGE_MOUNT_PATH='${mount_path}'
WARDEN_STORAGE_SYMLINK_PATH='${symlink_path}'
WARDEN_STORAGE_BROKER_MODE='${broker_mode}'
WARDEN_STORAGE_BROKER_HOST='${broker_host}'
WARDEN_INFISICAL_PROJECT_ID='4a897376-3cbd-4aeb-8550-c7d3ed7ad113'
WARDEN_INFISICAL_ENV='dev'
WARDEN_SHARED_STORAGE_INFISICAL_PATH='/warden/shared-storage/01'
WARDEN_SHARED_STORAGE_SECRET_NAME='WARDEN_SHARED_STORAGE_01_SMB_PASSWORD'
WARDEN_STORAGE_CRED_FILE='/run/shippin-secrets/warden-shared-storage-01.smb'
EOF
  chmod 0644 "${CONFIG_FILE}"
}

install_command() {
  install -m 0755 "${CLIENT_SCRIPT}" "${COMMAND_PATH}"
}

install_systemd_service() {
  if [[ "${PROFILE}" != "devstation" ]]; then
    return
  fi
  if ! command -v systemctl >/dev/null 2>&1; then
    return
  fi

  cat > "${SERVICE_FILE}" <<'EOF'
[Unit]
Description=Warden shared storage client
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
EnvironmentFile=-/etc/warden/storage-client.env
ExecStart=/usr/local/bin/warden-storage mount
ExecStop=/usr/local/bin/warden-storage unmount
RemainAfterExit=yes
TimeoutStartSec=90
TimeoutStopSec=30

[Install]
WantedBy=multi-user.target
EOF

  systemctl daemon-reload
  systemctl enable warden-storage.service
}

main() {
  require_root
  if [[ ! -f "${CLIENT_SCRIPT}" ]]; then
    echo "error: missing ${CLIENT_SCRIPT}" >&2
    exit 1
  fi

  install_command
  write_config
  install_systemd_service
  "${COMMAND_PATH}" status
}

main "$@"
