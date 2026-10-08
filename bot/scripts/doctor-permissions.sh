#!/usr/bin/env bash
set -euo pipefail

APP_USER=yomiage-keiryou
for path in /etc/yomiage-keiryou /var/lib/yomiage-keiryou /var/cache/yomiage-keiryou /var/cache/yomiage-keiryou/tts /var/cache/yomiage-keiryou/tmp /var/backups/yomiage-keiryou; do
  if [[ -e "$path" ]]; then
    printf '%-42s ' "$path"
    stat -c '所有者=%U:%G 権限=%a' "$path"
  else
    printf '%-42s 未作成\n' "$path"
  fi
done

if id -u "$APP_USER" >/dev/null 2>&1; then
  printf '\nユーザー: '; id "$APP_USER"
else
  printf '\nユーザー: 未作成\n'
fi

if [[ -f /usr/local/bin/yomiage-keiryou ]]; then
  printf 'バイナリ: '; stat -c '所有者=%U:%G 権限=%a %n' /usr/local/bin/yomiage-keiryou
else
  echo 'バイナリ: 未作成'
fi
