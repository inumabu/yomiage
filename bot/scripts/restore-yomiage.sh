#!/usr/bin/env bash
set -euo pipefail
umask 077

ENV_FILE=/etc/yomiage-keiryou/yomiage.env
[[ -r "$ENV_FILE" ]] && . "$ENV_FILE"
SETTINGS_FILE="${YOMIAGE_SETTINGS_FILE:-/var/lib/yomiage-keiryou/settings.json}"
ARCHIVE="${1:-}"

[[ -n "$ARCHIVE" ]] || { echo "使い方: restore-yomiage.sh バックアップ.tar.gz" >&2; exit 2; }
[[ -f "$ARCHIVE" ]] || { echo "バックアップが見つかりません: $ARCHIVE" >&2; exit 1; }

install -d -m 0750 "$(dirname "$SETTINGS_FILE")"
if [[ -f "$SETTINGS_FILE" ]]; then
  cp -a "$SETTINGS_FILE" "$SETTINGS_FILE.before-restore.$(date -u +%Y%m%d-%H%M%S)"
fi

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT
tar -xzf "$ARCHIVE" -C "$tmpdir" --no-same-owner --no-same-permissions -- settings.json
found="$tmpdir/settings.json"
[[ -f "$found" ]] || { echo "アーカイブ内にsettings.jsonがありません" >&2; exit 1; }
install -m 0640 "$found" "$SETTINGS_FILE"
echo "復元しました: $SETTINGS_FILE"
