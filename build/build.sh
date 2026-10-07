#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UPSTREAM_URL="https://github.com/inumabu/yomiage.git"
UPSTREAM_COMMIT="3b810b45153f200d8a6ec8e190dee7a52d090d6e"
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
    echo "unsupported target: $TARGET_OS/$TARGET_ARCH (use linux/amd64)" >&2
    exit 2
    ;;
esac

rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

echo "[1/8] cloning pinned upstream"
git clone --filter=blob:none --no-checkout "$UPSTREAM_URL" "$WORK_DIR/yomiage" >/dev/null 2>&1
cd "$WORK_DIR/yomiage"
git fetch --depth 1 origin "$UPSTREAM_COMMIT" >/dev/null 2>&1
git checkout --detach "$UPSTREAM_COMMIT" >/dev/null 2>&1
CURRENT="$(git rev-parse HEAD)"
if [[ "$CURRENT" != "$UPSTREAM_COMMIT" ]]; then
  echo "ERROR: failed to checkout pinned upstream commit: $CURRENT" >&2
  exit 1
fi

echo "[2/8] checking keiryou patch"
git apply --check "$ROOT_DIR/patches/0001-keiryou-cache.patch"
git apply --check "$ROOT_DIR/patches/0002-dave.patch"

echo "[3/8] applying patches"
git apply "$ROOT_DIR/patches/0001-keiryou-cache.patch"
git apply "$ROOT_DIR/patches/0002-dave.patch"

echo "[4/8] cloning DAVE DiscordGo and libdave"
git clone --depth 1 --branch "$DAVE_REF" --recurse-submodules "$DAVE_URL" "$WORK_DIR/discordgo" >/dev/null 2>&1
DAVE_CPP="$WORK_DIR/discordgo/dave/libdave/cpp"

echo "[5/8] building libdave"
cd "$DAVE_CPP"
./vcpkg/bootstrap-vcpkg.sh -disableMetrics >/dev/null
make BUILD_TYPE=Release

LIBDAVE_BUILD="$DAVE_CPP/build"
VCPKG_LIBDIR="$(find "$LIBDAVE_BUILD/vcpkg_installed" -maxdepth 2 -type d -name lib | head -1)"
if [[ -z "$VCPKG_LIBDIR" || ! -f "$LIBDAVE_BUILD/libdave.a" ]]; then
  echo "ERROR: libdave build output was not found" >&2
  exit 1
fi
VCPKG_LIBS="$(find "$VCPKG_LIBDIR" -maxdepth 1 -name '*.a' | sort | tr '\n' ' ')"
export CGO_ENABLED=1
export CGO_CFLAGS="-I$DAVE_CPP/includes"
export CGO_LDFLAGS="-L$VCPKG_LIBDIR $LIBDAVE_BUILD/libdave.a -Wl,--start-group $VCPKG_LIBS -Wl,--end-group -lstdc++ -lm -ldl -lpthread"

cd "$WORK_DIR/yomiage"
echo "[6/8] testing DAVE-enabled build"
go test -vet=off ./...

echo "[7/8] building $TARGET_OS/$TARGET_ARCH"
GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -trimpath -ldflags='-s -w' -o "$DIST_DIR/yomiage-keiryou-$TARGET_ARCH" .

echo "[8/8] result"
ls -lh "$DIST_DIR/yomiage-keiryou-$TARGET_ARCH"
