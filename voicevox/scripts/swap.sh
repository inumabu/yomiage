#!/usr/bin/env bash
set -euo pipefail
[[ $EUID -eq 0 ]] || { echo 'root権限で実行してください' >&2; exit 1; }

SIZE=${SWAP_SIZE:-2G}
FILE=/swapfile

if swapon --show=NAME --noheadings | grep -qx "$FILE"; then
  echo "$FILE は既に有効です"
  exit 0
fi
if [[ ! -f "$FILE" ]]; then
  fallocate -l "$SIZE" "$FILE"
  chmod 600 "$FILE"
  mkswap "$FILE"
fi
swapon "$FILE"
grep -qF "$FILE none swap sw 0 0" /etc/fstab || echo "$FILE none swap sw 0 0" >> /etc/fstab
sysctl vm.swappiness=10
cat >/etc/sysctl.d/99-yomiage-keiryou-swap.conf <<SYSCTL
vm.swappiness=10
SYSCTL
