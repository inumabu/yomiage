#!/usr/bin/env bash
set -euo pipefail
umask 077

ENV_FILE=/etc/yomiage-keiryou/yomiage.env
BACKUP_ENV=/etc/yomiage-keiryou/backup.env
[[ -r "$ENV_FILE" ]] && . "$ENV_FILE"
[[ -r "$BACKUP_ENV" ]] && . "$BACKUP_ENV"

SETTINGS_FILE="${YOMIAGE_SETTINGS_FILE:-/var/lib/yomiage-keiryou/settings.json}"
BACKUP_DIR=/var/backups/yomiage-keiryou
RETENTION="${LOCAL_RETENTION_DAYS:-14}"
[[ "$RETENTION" =~ ^[0-9]+$ ]] || { echo "LOCAL_RETENTION_DAYSは0以上の整数で指定してください" >&2; exit 2; }
STAMP=$(date -u +%Y%m%d-%H%M%S)
ARCHIVE="$BACKUP_DIR/yomiage-settings-$STAMP.tar.gz"

install -d -m 0700 "$BACKUP_DIR"
install -d -m 0750 "$(dirname "$SETTINGS_FILE")"

if [[ ! -f "$SETTINGS_FILE" ]]; then
  install -m 0600 /dev/null "$SETTINGS_FILE"
  printf '{ }\n' > "$SETTINGS_FILE"
fi

tar -C "$(dirname "$SETTINGS_FILE")" -czf "$ARCHIVE" "$(basename "$SETTINGS_FILE")"

find "$BACKUP_DIR" -type f -name 'yomiage-settings-*.tar.gz' -mtime "+$RETENTION" -delete

if [[ "${R2_ENABLED:-0}" == "1" ]]; then
  command -v rclone >/dev/null 2>&1 || { echo "R2バックアップにはrcloneが必要です" >&2; exit 1; }
  : "${R2_ACCOUNT_ID:?R2_ACCOUNT_IDが必要です}"
  : "${R2_BUCKET:?R2_BUCKETが必要です}"
  : "${R2_ACCESS_KEY_ID:?R2_ACCESS_KEY_IDが必要です}"
  : "${R2_SECRET_ACCESS_KEY:?R2_SECRET_ACCESS_KEYが必要です}"

  endpoint="https://${R2_ACCOUNT_ID}.r2.cloudflarestorage.com"
  key="${R2_PREFIX:-yomiage-keiryou}/$(hostname)/$(basename "$ARCHIVE")"

  rclone copyto "$ARCHIVE" ":s3:${R2_BUCKET}/${key}" \
    --s3-provider Cloudflare \
    --s3-endpoint "$endpoint" \
    --s3-access-key-id "$R2_ACCESS_KEY_ID" \
    --s3-secret-access-key "$R2_SECRET_ACCESS_KEY"
fi

echo "バックアップ: $ARCHIVE"
