#!/usr/bin/env bash
set -euo pipefail
umask 077

ENV_FILE=/etc/yomiage-keiryou/yomiage.env
[[ -r "$ENV_FILE" ]] && . "$ENV_FILE"
SETTINGS_FILE="${YOMIAGE_SETTINGS_FILE:-/var/lib/yomiage-keiryou/settings.json}"
ARCHIVE="${1:-}"

[[ -n "$ARCHIVE" ]] || { echo "usage: restore-yomiage.sh BACKUP.tar.gz" >&2; exit 2; }
[[ -f "$ARCHIVE" ]] || { echo "backup not found: $ARCHIVE" >&2; exit 1; }

install -d -m 0750 "$(dirname "$SETTINGS_FILE")"
if [[ -f "$SETTINGS_FILE" ]]; then
  cp -a "$SETTINGS_FILE" "$SETTINGS_FILE.before-restore.$(date -u +%Y%m%d-%H%M%S)"
fi

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT
tar -xzf "$ARCHIVE" -C "$tmpdir"
found=$(find "$tmpdir" -maxdepth 1 -type f -name 'settings.json' -print -quit)
[[ -n "$found" ]] || { echo "settings.json not found in archive" >&2; exit 1; }
install -m 0640 "$found" "$SETTINGS_FILE"
echo "restored: $SETTINGS_FILE"
