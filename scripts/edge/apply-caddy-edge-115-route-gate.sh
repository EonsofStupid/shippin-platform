#!/usr/bin/env bash
set -euo pipefail

WAN_IF="${WAN_IF:-vmbr0}"
LAN_IF="${LAN_IF:-vmbr1}"
EDGE_IP="${WARDEN_EDGE_IP:-10.0.0.115}"
LEGACY_EDGE_IP="${WARDEN_LEGACY_EDGE_IP:-10.0.0.100}"
OLD_DUP_EDGE_IP="${WARDEN_OLD_DUP_EDGE_IP:-10.1.1.100}"

remove_rule() {
  local table="$1" chain="$2"
  shift 2
  while iptables -t "$table" -D "$chain" "$@" 2>/dev/null; do :; done
}

ensure_rule() {
  local table="$1" chain="$2"
  shift 2
  if ! iptables -t "$table" -C "$chain" "$@" 2>/dev/null; then
    iptables -t "$table" -A "$chain" "$@"
  fi
}

remove_rule nat PREROUTING -i "$WAN_IF" -p tcp -m tcp --dport 80 -j DNAT --to-destination "$LEGACY_EDGE_IP:80"
remove_rule nat PREROUTING -i "$WAN_IF" -p tcp -m tcp --dport 443 -j DNAT --to-destination "$LEGACY_EDGE_IP:443"
remove_rule nat PREROUTING -i "$WAN_IF" -p tcp -m tcp --dport 80 -j DNAT --to-destination "$OLD_DUP_EDGE_IP:80"
remove_rule nat PREROUTING -i "$WAN_IF" -p tcp -m tcp --dport 443 -j DNAT --to-destination "$OLD_DUP_EDGE_IP:443"
remove_rule nat PREROUTING -i "$WAN_IF" -p tcp -m tcp --dport 5432 -j DNAT --to-destination "$LEGACY_EDGE_IP:5432"
remove_rule nat POSTROUTING -d "$LEGACY_EDGE_IP/32" -p tcp -m multiport --dports 80,443 -j MASQUERADE
remove_rule nat POSTROUTING -d "$LEGACY_EDGE_IP/32" -p tcp -m tcp --dport 5432 -j MASQUERADE
remove_rule nat POSTROUTING -d "$EDGE_IP/32" -p tcp -m multiport --dports 80,443 -j MASQUERADE

ensure_rule nat PREROUTING -i "$WAN_IF" -p tcp -m tcp --dport 80 -j DNAT --to-destination "$EDGE_IP:80"
ensure_rule nat PREROUTING -i "$WAN_IF" -p tcp -m tcp --dport 443 -j DNAT --to-destination "$EDGE_IP:443"
ensure_rule nat POSTROUTING -d "$EDGE_IP/32" -p tcp -m multiport --dports 80,443 -j MASQUERADE

remove_rule filter FORWARD -d "$LEGACY_EDGE_IP/32" -i "$WAN_IF" -o "$LAN_IF" -p tcp -m tcp --dport 80 -j ACCEPT
remove_rule filter FORWARD -d "$LEGACY_EDGE_IP/32" -i "$WAN_IF" -o "$LAN_IF" -p tcp -m tcp --dport 443 -j ACCEPT
remove_rule filter FORWARD -d "$LEGACY_EDGE_IP/32" -p tcp -m tcp --dport 80 -j ACCEPT
remove_rule filter FORWARD -d "$LEGACY_EDGE_IP/32" -p tcp -m tcp --dport 443 -j ACCEPT
remove_rule filter FORWARD -d "$OLD_DUP_EDGE_IP/32" -i "$WAN_IF" -o "$LAN_IF" -p tcp -m tcp --dport 80 -j ACCEPT
remove_rule filter FORWARD -d "$OLD_DUP_EDGE_IP/32" -i "$WAN_IF" -o "$LAN_IF" -p tcp -m tcp --dport 443 -j ACCEPT
remove_rule filter FORWARD -d "$LEGACY_EDGE_IP/32" -p tcp -m tcp --dport 5432 -j ACCEPT
remove_rule filter FORWARD -d "$EDGE_IP/32" -i "$WAN_IF" -o "$LAN_IF" -p tcp -m tcp --dport 80 -j ACCEPT
remove_rule filter FORWARD -d "$EDGE_IP/32" -i "$WAN_IF" -o "$LAN_IF" -p tcp -m tcp --dport 443 -j ACCEPT

ensure_rule filter FORWARD -d "$EDGE_IP/32" -i "$WAN_IF" -o "$LAN_IF" -p tcp -m tcp --dport 80 -j ACCEPT
ensure_rule filter FORWARD -d "$EDGE_IP/32" -i "$WAN_IF" -o "$LAN_IF" -p tcp -m tcp --dport 443 -j ACCEPT

iptables -t nat -S
iptables -S FORWARD | grep -E "10\\.0\\.0\\.(100|115)|10\\.1\\.1\\.100|5432" || true
