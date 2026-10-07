# 🍓 Raspberry Pi

v5 の標準ターゲットは **64-bit Raspberry Pi OS / ARM64** です。

```bash
TARGET_ARCH=arm64 ./build/build.sh
```

Bot は Go 単体バイナリ + ffmpeg + systemd で動作します。VOICEVOX は PC 側へ逃がす構成を推奨します。
