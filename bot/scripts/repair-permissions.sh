#!/usr/bin/env bash
set -euo pipefail
[[ $EUID -eq 0 ]] || { echo 'root権限で実行してください: sudo bash ./bot/scripts/repair-permissions.sh' >&2; exit 1; }

if ! id -u yomiage-keiryou >/dev/null 2>&1; then
  useradd --system --home /var/lib/yomiage-keiryou --shell /usr/sbin/nologin yomiage-keiryou
fi

install -d -m 0750 /etc/yomiage-keiryou
install -d -m 0750 /var/lib/yomiage-keiryou
install -d -m 0750 /var/cache/yomiage-keiryou/tts
install -d -m 0700 /var/cache/yomiage-keiryou/tmp
install -d -m 0700 /var/backups/yomiage-keiryou

chown -R yomiage-keiryou:yomiage-keiryou /var/lib/yomiage-keiryou /var/cache/yomiage-keiryou
chmod 0700 /var/cache/yomiage-keiryou/tmp
chmod 0750 /var/cache/yomiage-keiryou/tts

if [[ -f /var/lib/yomiage-keiryou/settings.json ]]; then
  chown yomiage-keiryou:yomiage-keiryou /var/lib/yomiage-keiryou/settings.json
  chmod 0600 /var/lib/yomiage-keiryou/settings.json
fi

if [[ -f /usr/local/bin/yomiage-keiryou ]]; then
  chown root:root /usr/local/bin/yomiage-keiryou
  chmod 0755 /usr/local/bin/yomiage-keiryou
fi

echo '権限を修復しました。'
