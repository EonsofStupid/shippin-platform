#!/usr/bin/env bash
set -euo pipefail

RUNTIME="${1:-codex}"
PROFILE="${2:-devstation}"
WORKDIR="${3:-}"

case "${RUNTIME}" in
  codex|claude) ;;
  *)
    echo "error: unsupported runtime '${RUNTIME}'" >&2
    exit 2
    ;;
esac

if [[ -z "${WORKDIR}" ]]; then
  case "${PROFILE}" in
    devstation)
      WORKDIR="/workspace/warden-storage/projects/shippin-platform"
      ;;
    capsule)
      WORKDIR="/workspace/WardenClyffe-latest"
      ;;
    *)
      WORKDIR="${PWD}"
      ;;
  esac
fi

if [[ ! -d "${WORKDIR}" ]]; then
  echo "error: workdir not found: ${WORKDIR}" >&2
  exit 1
fi

if ! command -v "${RUNTIME}" >/dev/null 2>&1; then
  echo "error: ${RUNTIME} is not installed on this host" >&2
  exit 1
fi

SESSION="${WARDEN_AGENT_STREAM_NAME:-warden-${PROFILE}-${RUNTIME}}"

if [[ "${WARDEN_AGENT_NO_TMUX:-0}" == "1" || -n "${TMUX:-}" || ! -t 0 || ! -t 1 ]] || ! command -v tmux >/dev/null 2>&1; then
  cd "${WORKDIR}"
  exec "${RUNTIME}"
fi

exec tmux new-session -A -s "${SESSION}" "${SHELL:-/bin/bash}" -lc "cd $(printf '%q' "${WORKDIR}") && exec $(printf '%q' "${RUNTIME}")"
