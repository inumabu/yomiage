#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DAVE_URL="https://github.com/aleph-garden/discordgo.git"
DAVE_REF="v0.29.1-dave.26"
TARGET_OS="${TARGET_OS:-linux}"
TARGET_ARCH="${TARGET_ARCH:-amd64}"
WORK_DIR="$(mktemp -d)"
DIST_DIR="$ROOT_DIR/build/dist"
trap 'rm -rf "$WORK_DIR"' EXIT

case "$TARGET_OS/$TARGET_ARCH" in
  linux/amd64) ;;
  linux/arm64)
    echo "DAVE対応ビルドは現在linux/amd64のみ対応しています。arm64対応はlibdaveのクロスビルド整備後に追加します。" >&2
    exit 2
    ;;
  *)
    echo "未対応の対象です: $TARGET_OS/$TARGET_ARCH（linux/amd64を指定してください）" >&2
    exit 2
    ;;
esac

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# 配布物は常にこのリポジトリの検証済みHEADから作る。古い上流コミットを
# 再取得する方式では、READMEや本体の修正がリリースバイナリへ反映されない。
echo "[1/7] 現在のリポジトリソースを準備"
mkdir -p "$WORK_DIR/yomiage"
# archiveの展開先を明示し、ビルド作業領域をリポジトリ本体と分離する。
git -C "$ROOT_DIR" archive --format=tar HEAD | tar -x -C "$WORK_DIR/yomiage"

cd "$WORK_DIR/yomiage"
echo "[2/7] パッチを確認"
git apply --check "$ROOT_DIR/patches/0001-keiryou-cache.patch"
git apply --check "$ROOT_DIR/patches/0002-dave.patch"

echo "[3/7] パッチを適用"
git apply "$ROOT_DIR/patches/0001-keiryou-cache.patch"
git apply "$ROOT_DIR/patches/0002-dave.patch"

echo "[4/7] DAVE対応DiscordGoとlibdaveを取得"
git clone --depth 1 --branch "$DAVE_REF" --recurse-submodules "$DAVE_URL" "$WORK_DIR/discordgo" >/dev/null 2>&1
DAVE_CPP="$WORK_DIR/discordgo/dave/libdave/cpp"

echo "[5/7] libdaveをビルド"
cd "$DAVE_CPP"
./vcpkg/bootstrap-vcpkg.sh -disableMetrics >/dev/null
make BUILD_TYPE=Release

LIBDAVE_BUILD="$DAVE_CPP/build"
VCPKG_LIBDIR="$(find "$LIBDAVE_BUILD/vcpkg_installed" -maxdepth 2 -type d -name lib | head -1)"
if [[ -z "$VCPKG_LIBDIR" || ! -f "$LIBDAVE_BUILD/libdave.a" ]]; then
  echo "エラー: libdaveのビルド成果物が見つかりません" >&2
  exit 1
fi
VCPKG_LIBS="$(find "$VCPKG_LIBDIR" -maxdepth 1 -name '*.a' | sort | tr '\n' ' ')"
export CGO_ENABLED=1
export CGO_CFLAGS="-I$DAVE_CPP/includes"
export CGO_LDFLAGS="-L$VCPKG_LIBDIR $LIBDAVE_BUILD/libdave.a -Wl,--start-group $VCPKG_LIBS -Wl,--end-group -lstdc++ -lm -ldl -lpthread"

cd "$WORK_DIR/yomiage"
echo "[6/7] $TARGET_OS/$TARGET_ARCHのテストとビルド"
go test -vet=off ./...
GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -trimpath -ldflags='-s -w' -o "$DIST_DIR/yomiage-keiryou-$TARGET_ARCH" .

echo "[7/7] 成果物"
ls -lh "$DIST_DIR/yomiage-keiryou-$TARGET_ARCH"
