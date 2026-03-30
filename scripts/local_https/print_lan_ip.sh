#!/usr/bin/env bash
set -euo pipefail

HOSTNAME="xpressgo.home.arpa"
URL="https://$HOSTNAME"

detect_lan_ip() {
  if command -v ip >/dev/null 2>&1; then
    local ip_addr
    ip_addr="$(ip route get 1.1.1.1 2>/dev/null | awk '{for (i = 1; i <= NF; i++) if ($i == "src") { print $(i + 1); exit }}')"
    if [[ -n "$ip_addr" ]]; then
      printf '%s\n' "$ip_addr"
      return 0
    fi
  fi

  if command -v hostname >/dev/null 2>&1; then
    local ip_addr
    ip_addr="$(hostname -I 2>/dev/null | awk '{print $1}')"
    if [[ -n "$ip_addr" ]]; then
      printf '%s\n' "$ip_addr"
      return 0
    fi
  fi

  return 1
}

printf '%s\n' "Local HTTPS Mini App URL: $URL"

if lan_ip="$(detect_lan_ip)"; then
  printf '%s\n' "Current laptop LAN IP: $lan_ip"
  printf '%s\n' "Map $HOSTNAME to $lan_ip on your phone before opening Telegram on a new Wi-Fi network."
else
  printf '%s\n' "Could not detect the current LAN IP automatically."
  printf '%s\n' "Map $HOSTNAME to this laptop's current Wi-Fi IP on your phone before opening Telegram."
fi
