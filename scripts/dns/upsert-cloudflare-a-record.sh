#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  upsert-cloudflare-a-record.sh (--zone-id ZONE_ID | --zone-name ZONE) --name FQDN --target IP [--proxied] [--apply]

Environment:
  CLOUDFLARE_API_TOKEN may contain a Cloudflare token with DNS edit access.
  If unset, this script tries Infisical:
    project: 4a897376-3cbd-4aeb-8550-c7d3ed7ad113
    env:     dev
    path:    /
    secret:  WARDEN_CLOUDFLARE_DNS_ADMIN

Examples:
  scripts/dns/upsert-cloudflare-a-record.sh \
    --zone-name clyffy.ai \
    --name ssh.clyffy.ai \
    --target 104.176.44.101

  scripts/dns/upsert-cloudflare-a-record.sh \
    --zone-name clyffy.ai \
    --name master.clyffy.ai \
    --target 104.176.44.101 \
    --apply
USAGE
}

zone_id=""
zone_name=""
record_name=""
target_ip=""
proxied=false
apply=false

infisical_project_id="${INFISICAL_CLYFFY_PROJECT_ID:-4a897376-3cbd-4aeb-8550-c7d3ed7ad113}"
infisical_env="${INFISICAL_ENV:-dev}"
infisical_path="${INFISICAL_SECRET_PATH:-/}"
infisical_secret_name="${INFISICAL_CLOUDFLARE_SECRET_NAME:-WARDEN_CLOUDFLARE_DNS_ADMIN}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --zone-id)
      zone_id="${2:-}"
      shift 2
      ;;
    --zone-name)
      zone_name="${2:-}"
      shift 2
      ;;
    --name)
      record_name="${2:-}"
      shift 2
      ;;
    --target)
      target_ip="${2:-}"
      shift 2
      ;;
    --proxied)
      proxied=true
      shift
      ;;
    --apply)
      apply=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$zone_id" && -z "$zone_name" ]]; then
  usage >&2
  exit 2
fi

if [[ -n "$zone_id" && -n "$zone_name" ]]; then
  echo "Use either --zone-id or --zone-name, not both." >&2
  exit 2
fi

if [[ -z "$record_name" || -z "$target_ip" ]]; then
  usage >&2
  exit 2
fi

token="${CLOUDFLARE_API_TOKEN:-}"
token_source="env:CLOUDFLARE_API_TOKEN"

if [[ -z "$token" && -x "$(command -v infisical || true)" ]]; then
  token="$(
    infisical secrets get "$infisical_secret_name" \
      --projectId "$infisical_project_id" \
      --env "$infisical_env" \
      --path "$infisical_path" \
      --plain 2>/dev/null || true
  )"
  token_source="infisical:${infisical_project_id}:${infisical_env}:${infisical_path}:${infisical_secret_name}"
fi

if [[ -n "$zone_name" && -z "$token" ]]; then
  echo "Missing Cloudflare token. Set CLOUDFLARE_API_TOKEN or authenticate Infisical for $infisical_secret_name." >&2
  exit 1
fi

if [[ -n "$zone_name" ]]; then
  encoded_zone="$(python3 - <<PY
from urllib.parse import quote
print(quote("$zone_name", safe=""))
PY
)"
  zones="$(
    curl -fsS \
      -H "Authorization: Bearer ${token}" \
      -H "Content-Type: application/json" \
      "https://api.cloudflare.com/client/v4/zones?name=${encoded_zone}&per_page=5"
  )"
  zone_id="$(jq -r '.result[0].id // empty' <<<"$zones")"
  if [[ -z "$zone_id" ]]; then
    echo "Cloudflare zone not found: $zone_name" >&2
    exit 1
  fi
fi

jq -n \
  --arg action "$([[ "$apply" == true ]] && echo apply || echo dry-run)" \
  --arg zone_id "$zone_id" \
  --arg zone_name "$zone_name" \
  --arg record_name "$record_name" \
  --arg target_ip "$target_ip" \
  --arg token_source "$token_source" \
  --argjson proxied "$proxied" \
  --argjson token_present "$([[ -n "$token" ]] && echo true || echo false)" \
  '{
    action: $action,
    zone_id: $zone_id,
    zone_name: $zone_name,
    record_name: $record_name,
    target_ip: $target_ip,
    proxied: $proxied,
    token_source: $token_source,
    token_present: $token_present
  }'

if [[ "$apply" != true ]]; then
  echo "Dry run only. Re-run with --apply after route readiness is verified."
  exit 0
fi

if [[ -z "$token" ]]; then
  echo "Missing Cloudflare token. Set CLOUDFLARE_API_TOKEN or authenticate Infisical for $infisical_secret_name." >&2
  exit 1
fi

payload="$(
  jq -n \
    --arg name "$record_name" \
    --arg content "$target_ip" \
    --argjson proxied "$proxied" \
    '{
      type: "A",
      name: $name,
      content: $content,
      ttl: 1,
      proxied: $proxied,
      comment: "Shippin managed DNS record"
    }'
)"

encoded_name="$(python3 - <<PY
from urllib.parse import quote
print(quote("$record_name", safe=""))
PY
)"

base_url="https://api.cloudflare.com/client/v4/zones/${zone_id}/dns_records"
existing="$(
  curl -fsS \
    -H "Authorization: Bearer ${token}" \
    -H "Content-Type: application/json" \
    "${base_url}?type=A&name=${encoded_name}"
)"

record_id="$(jq -r '.result[0].id // empty' <<<"$existing")"

if [[ -n "$record_id" ]]; then
  result="$(
    curl -fsS -X PUT \
      -H "Authorization: Bearer ${token}" \
      -H "Content-Type: application/json" \
      --data "$payload" \
      "${base_url}/${record_id}"
  )"
  verb="updated"
else
  result="$(
    curl -fsS -X POST \
      -H "Authorization: Bearer ${token}" \
      -H "Content-Type: application/json" \
      --data "$payload" \
      "$base_url"
  )"
  verb="created"
fi

jq --arg verb "$verb" '{
  result: $verb,
  success,
  name: .result.name,
  type: .result.type,
  content: .result.content,
  proxied: .result.proxied,
  id: .result.id
}' <<<"$result"
