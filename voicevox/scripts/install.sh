#!/usr/bin/env bash
set -euo pipefail
[[ $EUID -eq 0 ]] || { echo 'root権限で実行してください' >&2; exit 1; }
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

install -d -m 0755 /usr/local/libexec
[[ -x /opt/voicevox/voicevox_engine ]] || { echo '/opt/voicevox/voicevox_engine が見つかりません' >&2; exit 1; }

if ! id -u voicevox >/dev/null 2>&1; then
  useradd --system --home /opt/voicevox --shell /usr/sbin/nologin voicevox
fi
chown -R voicevox:voicevox /opt/voicevox
install -m 0755 "$ROOT_DIR/voicevox/scripts/voicevox-wrapper" /usr/local/libexec/voicevox-keiryou-wrapper
install -m 0644 "$ROOT_DIR/voicevox/systemd/voicevox-engine.service" /etc/systemd/system/voicevox-engine.service
if [[ ! -f /etc/voicevox-engine.env ]]; then
  install -m 0600 "$ROOT_DIR/voicevox/voicevox.env.example" /etc/voicevox-engine.env
fi
systemctl daemon-reload
echo 'インストールが完了しました。/etc/voicevox-engine.env を編集してから voicevox-engine.service を有効化してください。'
