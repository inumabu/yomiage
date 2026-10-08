#!/usr/bin/env bash
set -euo pipefail

[[ $EUID -eq 0 ]] || { echo 'root権限で実行してください' >&2; exit 1; }
BINARY_SRC="${1:-}"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

if [[ -z "$BINARY_SRC" ]]; then
  echo "使い方: sudo bash ./windows/wsl/install-bot.sh /path/to/yomiage-keiryou-amd64" >&2
  exit 2
fi
[[ -f "$BINARY_SRC" ]] || { echo "バイナリが見つかりません: $BINARY_SRC" >&2; exit 1; }

install -d -m 0750 /etc/yomiage-keiryou /etc/systemd/journald.conf.d /var/lib/yomiage-keiryou /var/cache/yomiage-keiryou/tts /var/cache/yomiage-keiryou/tmp
if ! id -u yomiage-keiryou >/dev/null 2>&1; then
  useradd --system --home /var/lib/yomiage-keiryou --shell /usr/sbin/nologin yomiage-keiryou
fi
install -m 0755 "$BINARY_SRC" /usr/local/bin/yomiage-keiryou
chown -R yomiage-keiryou:yomiage-keiryou /var/lib/yomiage-keiryou /var/cache/yomiage-keiryou
chmod 0700 /var/cache/yomiage-keiryou/tmp
chmod 0750 /var/cache/yomiage-keiryou/tts

install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou.service" /etc/systemd/system/yomiage-keiryou.service
install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou-cache-gc.service" /etc/systemd/system/yomiage-keiryou-cache-gc.service
install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou-cache-gc.timer" /etc/systemd/system/yomiage-keiryou-cache-gc.timer
install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou.conf" /etc/systemd/journald.conf.d/99-yomiage-keiryou.conf

if [[ ! -f /etc/yomiage-keiryou/yomiage.env ]]; then
  install -m 0600 "$ROOT_DIR/modes/01-local-pc-always/yomiage.env.example" /etc/yomiage-keiryou/yomiage.env
fi

systemctl daemon-reload
echo 'WSL用Botをインストールしました。/etc/yomiage-keiryou/yomiage.envを編集してから、次を実行してください:'
echo '  systemctl enable --now yomiage-keiryou.service'
echo '  systemctl enable --now yomiage-keiryou-cache-gc.timer'
