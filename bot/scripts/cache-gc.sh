#!/usr/bin/env bash
set -euo pipefail

ENV_FILE=/etc/yomiage-keiryou/yomiage.env
[[ -r "$ENV_FILE" ]] && . "$ENV_FILE"

CACHE_DIR="${YOMIAGE_CACHE_DIR:-/var/cache/yomiage-keiryou/tts}"
TTL="${YOMIAGE_CACHE_TTL:-168h}"
MAX_BYTES="${YOMIAGE_CACHE_MAX_BYTES:-268435456}"

mkdir -p "$CACHE_DIR"
chmod 0750 "$CACHE_DIR" || true

# GNU coreutils: remove files older than TTL. TTL=0 disables age-based deletion.
case "$TTL" in
  0) TTL_MIN=0;;
  *d) value=${TTL%d}; [[ "$value" =~ ^[0-9]+$ ]] || { echo "YOMIAGE_CACHE_TTLが不正です: $TTL" >&2; exit 2; }; TTL_MIN=$(( value * 1440 ));;
  *h) value=${TTL%h}; [[ "$value" =~ ^[0-9]+$ ]] || { echo "YOMIAGE_CACHE_TTLが不正です: $TTL" >&2; exit 2; }; TTL_MIN=$(( value * 60 ));;
  *m) value=${TTL%m}; [[ "$value" =~ ^[0-9]+$ ]] || { echo "YOMIAGE_CACHE_TTLが不正です: $TTL" >&2; exit 2; }; TTL_MIN=$value;;
  *) echo "YOMIAGE_CACHE_TTLが不正です: $TTL" >&2; exit 2;;
esac
if (( TTL_MIN > 0 )); then
  find "$CACHE_DIR" -maxdepth 1 -type f -name '*.wav' -mmin "+$TTL_MIN" -delete
fi

# Evict oldest entries until under the byte cap.
[[ "$MAX_BYTES" =~ ^[0-9]+$ ]] || { echo "YOMIAGE_CACHE_MAX_BYTESが不正です: $MAX_BYTES" >&2; exit 2; }
if (( MAX_BYTES > 0 )); then
  while :; do
    total=$(du -sb "$CACHE_DIR" 2>/dev/null | awk '{print $1}')
    [[ -z "$total" ]] && total=0
    (( total <= MAX_BYTES )) && break

    oldest=$(find "$CACHE_DIR" -maxdepth 1 -type f -name '*.wav' -printf '%T@\t%p\n' 2>/dev/null | sort -n | head -n 1 | cut -f2-)
    [[ -n "$oldest" ]] || break
    rm -f -- "$oldest"
  done
fi
