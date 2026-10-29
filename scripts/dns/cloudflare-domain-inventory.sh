#!/usr/bin/env bash
set -euo pipefail

infisical_project_id="${INFISICAL_CLYFFY_PROJECT_ID:-4a897376-3cbd-4aeb-8550-c7d3ed7ad113}"
infisical_env="${INFISICAL_ENV:-dev}"
infisical_path="${INFISICAL_SECRET_PATH:-/}"
infisical_secret_name="${INFISICAL_CLOUDFLARE_SECRET_NAME:-WARDEN_CLOUDFLARE_DNS_ADMIN}"

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

if [[ -z "$token" ]]; then
  echo "Missing Cloudflare token. Set CLOUDFLARE_API_TOKEN or authenticate Infisical for $infisical_secret_name." >&2
  exit 1
fi

zones_json="$(mktemp)"
records_json=""
trap 'rm -f "$zones_json" "$records_json"' EXIT

curl -fsS \
  -H "Authorization: Bearer ${token}" \
  -H "Content-Type: application/json" \
  "https://api.cloudflare.com/client/v4/zones?per_page=100" >"$zones_json"

jq -n \
  --arg token_source "$token_source" \
  --argjson zone_count "$(jq '.result | length' "$zones_json")" \
  '{token_source: $token_source, token_present: true, zone_count: $zone_count}'

jq -r '.result[] | [.name, .id, .status, .account.name] | @tsv' "$zones_json" |
while IFS=$'\t' read -r zone zone_id status account; do
  records_json="$(mktemp)"
  code="$(
    curl -sS -o "$records_json" -w '%{http_code}' \
      -H "Authorization: Bearer ${token}" \
      -H "Content-Type: application/json" \
      "https://api.cloudflare.com/client/v4/zones/${zone_id}/dns_records?per_page=200" || true
  )"

  if [[ "$code" != "200" ]]; then
    jq -n \
      --arg zone "$zone" \
      --arg zone_id "$zone_id" \
      --arg status "$status" \
      --arg account "$account" \
      --arg http "$code" \
      --argjson errors "$(jq '.errors // []' "$records_json" 2>/dev/null || echo '[]')" \
      '{zone: $zone, zone_id: $zone_id, status: $status, account: $account, dns_records_http: $http, errors: $errors}'
    rm -f "$records_json"
    continue
  fi

  jq -n \
    --arg zone "$zone" \
    --arg zone_id "$zone_id" \
    --arg status "$status" \
    --arg account "$account" \
    --arg http "$code" \
    --argjson record_count "$(jq '.result | length' "$records_json")" \
    --argjson public_records "$(jq '[.result[] | select(.type=="A" or .type=="AAAA" or .type=="CNAME") | {type, name, content, proxied, ttl}]' "$records_json")" \
    '{
      zone: $zone,
      zone_id: $zone_id,
      status: $status,
      account: $account,
      dns_records_http: $http,
      record_count: $record_count,
      public_records: $public_records
    }'
  rm -f "$records_json"
done
