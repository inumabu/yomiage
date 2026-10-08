#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if ! command -v node >/dev/null 2>&1; then
  echo "Node.jsが見つかりません。導入を試みます。"
  if command -v apt-get >/dev/null 2>&1; then
    if command -v sudo >/dev/null 2>&1; then
      sudo apt-get update && sudo apt-get install -y --no-install-recommends nodejs npm
    else
      apt-get update && apt-get install -y --no-install-recommends nodejs npm
    fi
  else
    echo "apt-getが見つからないためNode.jsを自動導入できません。Node.js 22以上を導入して再実行してください。" >&2
    exit 1
  fi
fi
if ! command -v node >/dev/null 2>&1; then
  echo "Node.jsの導入後もnodeコマンドが見つかりません。PATHを確認してください。" >&2
  exit 1
fi
exec node "$ROOT_DIR/scripts/setup.mjs" "$@"
