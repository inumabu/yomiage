#!/usr/bin/env bash
set -euo pipefail
[[ $EUID -eq 0 ]] || { echo 'run as root' >&2; exit 1; }

ENV=/etc/yomiage-keiryou/network.env
[[ -r "$ENV" ]] || { echo "$ENV is missing" >&2; exit 1; }
. "$ENV"
: "${SSH_PORT:=22}"

cat > /etc/nftables.conf <<RULES
flush ruleset

table inet yomiage_keiryou {
  chain input {
    type filter hook input priority 0; policy drop;
    iifname "lo" accept
    ct state established,related accept
    ip protocol icmp accept
    tcp dport ${SSH_PORT} ct state new accept
  }

  chain forward {
    type filter hook forward priority 0; policy drop;
  }

  chain output {
    type filter hook output priority 0; policy accept;
  }
}
RULES

systemctl enable nftables
nft -f /etc/nftables.conf
systemctl restart nftables
echo 'Bot firewall installed.'
