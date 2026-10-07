#!/usr/bin/env bash
set -euo pipefail
[[ $EUID -eq 0 ]] || { echo 'run as root' >&2; exit 1; }

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

install -d -m 0750 /etc/yomiage-keiryou /var/lib/yomiage-keiryou /var/cache/yomiage-keiryou/tts /var/cache/yomiage-keiryou/tmp /var/backups/yomiage-keiryou /usr/local/libexec/yomiage-keiryou

if ! id -u yomiage-keiryou >/dev/null 2>&1; then
  useradd --system --home /var/lib/yomiage-keiryou --shell /usr/sbin/nologin yomiage-keiryou
fi
chown -R yomiage-keiryou:yomiage-keiryou /var/lib/yomiage-keiryou /var/cache/yomiage-keiryou
chmod 0700 /var/lib/yomiage-keiryou /var/cache/yomiage-keiryou/tmp
chmod 0750 /var/cache/yomiage-keiryou/tts

[[ -x /usr/local/bin/yomiage-keiryou ]] || { echo '/usr/local/bin/yomiage-keiryou is missing' >&2; exit 1; }
install -m 0755 "$ROOT_DIR/bot/scripts/cache-gc.sh" /usr/local/libexec/yomiage-keiryou/cache-gc.sh
install -m 0755 "$ROOT_DIR/bot/scripts/backup-yomiage.sh" /usr/local/libexec/yomiage-keiryou/backup-yomiage.sh
install -m 0755 "$ROOT_DIR/bot/scripts/restore-yomiage.sh" /usr/local/libexec/yomiage-keiryou/restore.sh

if [[ ! -f /etc/yomiage-keiryou/yomiage.env ]]; then
  install -m 0600 "$ROOT_DIR/bot/yomiage.env.example" /etc/yomiage-keiryou/yomiage.env
fi
if [[ ! -f /etc/yomiage-keiryou/backup.env ]]; then
  install -m 0600 "$ROOT_DIR/bot/backup.env.example" /etc/yomiage-keiryou/backup.env
fi
if [[ ! -f /etc/yomiage-keiryou/network.env ]]; then
  install -m 0644 "$ROOT_DIR/bot/network.env.example" /etc/yomiage-keiryou/network.env
fi

install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou.service" /etc/systemd/system/yomiage-keiryou.service
install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou-cache-gc.service" /etc/systemd/system/yomiage-keiryou-cache-gc.service
install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou-cache-gc.timer" /etc/systemd/system/yomiage-keiryou-cache-gc.timer
install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou-backup.service" /etc/systemd/system/yomiage-keiryou-backup.service
install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou-backup.timer" /etc/systemd/system/yomiage-keiryou-backup.timer
install -d -m 0755 /etc/systemd/journald.conf.d
install -m 0644 "$ROOT_DIR/bot/systemd/yomiage-keiryou.conf" /etc/systemd/journald.conf.d/99-yomiage-keiryou.conf
systemctl restart systemd-journald

systemctl daemon-reload
echo 'Installed. Edit /etc/yomiage-keiryou/yomiage.env and backup.env, then:'
echo '  systemctl enable --now yomiage-keiryou.service'
echo '  systemctl enable --now yomiage-keiryou-cache-gc.timer'
echo '  systemctl enable --now yomiage-keiryou-backup.timer'
