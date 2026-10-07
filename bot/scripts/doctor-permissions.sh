#!/usr/bin/env bash
set -euo pipefail

APP_USER=yomiage-keiryou
for path in /etc/yomiage-keiryou /var/lib/yomiage-keiryou /var/cache/yomiage-keiryou /var/cache/yomiage-keiryou/tts /var/cache/yomiage-keiryou/tmp /var/backups/yomiage-keiryou; do
  if [[ -e "$path" ]]; then
    printf '%-42s ' "$path"
    stat -c 'owner=%U:%G mode=%a' "$path"
  else
    printf '%-42s MISSING\n' "$path"
  fi
done

if id -u "$APP_USER" >/dev/null 2>&1; then
  printf '\nuser: '; id "$APP_USER"
else
  echo '\nuser: MISSING'
fi

if [[ -f /usr/local/bin/yomiage-keiryou ]]; then
  printf 'binary: '; stat -c 'owner=%U:%G mode=%a %n' /usr/local/bin/yomiage-keiryou
else
  echo 'binary: MISSING'
fi
