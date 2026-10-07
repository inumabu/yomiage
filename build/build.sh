#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UPSTREAM_URL="https://github.com/inumabu/yomiage.git"
UPSTREAM_COMMIT="3b810b45153f200d8a6ec8e190dee7a52d090d6e"
TARGET_OS="${TARGET_OS:-linux}"
TARGET_ARCH="${TARGET_ARCH:-amd64}"
WORK_DIR="$(mktemp -d)"
DIST_DIR="$ROOT_DIR/build/dist"
trap 'rm -rf "$WORK_DIR"' EXIT

case "$TARGET_OS/$TARGET_ARCH" in
  linux/amd64|linux/arm64) ;;
  *) echo "unsupported target: $TARGET_OS/$TARGET_ARCH (use linux/amd64 or linux/arm64)" >&2; exit 2 ;;
esac

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

echo "[1/6] cloning upstream"
git clone --filter=blob:none --no-checkout "$UPSTREAM_URL" "$WORK_DIR/yomiage" >/dev/null 2>&1
cd "$WORK_DIR/yomiage"
git fetch --depth 1 origin "$UPSTREAM_COMMIT" >/dev/null 2>&1
git checkout --detach "$UPSTREAM_COMMIT" >/dev/null 2>&1
CURRENT="$(git rev-parse HEAD)"
if [[ "$CURRENT" != "$UPSTREAM_COMMIT" ]]; then
  echo "ERROR: failed to checkout pinned upstream commit: $CURRENT" >&2
  exit 1
fi

echo "[2/6] checking patch"
git apply --check "$ROOT_DIR/patches/0001-keiryou-cache.patch"

echo "[3/6] applying patch"
git apply "$ROOT_DIR/patches/0001-keiryou-cache.patch"

echo "[4/6] testing"
go test ./...

echo "[5/6] building $TARGET_OS/$TARGET_ARCH"
CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -trimpath -ldflags='-s -w' -o "$DIST_DIR/yomiage-keiryou-$TARGET_ARCH" .

echo "[6/6] result"
ls -lh "$DIST_DIR/yomiage-keiryou-$TARGET_ARCH"
