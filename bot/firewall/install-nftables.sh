#!/usr/bin/env bash
set -euo pipefail
[[ $EUID -eq 0 ]] || { echo 'root権限で実行してください' >&2; exit 1; }

ENV=/etc/yomiage-keiryou/network.env
[[ -r "$ENV" ]] || { echo "$ENV が見つかりません" >&2; exit 1; }
. "$ENV"
: "${SSH_PORT:=22}"
[[ "$SSH_PORT" =~ ^[0-9]+$ ]] && (( SSH_PORT >= 1 && SSH_PORT <= 65535 )) || { echo 'SSH_PORTは1〜65535で指定してください' >&2; exit 2; }

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
echo 'Bot用Firewallを設定しました。'
