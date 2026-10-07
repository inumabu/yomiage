# 🍓 Raspberry Pi

Raspberry Pi は **64-bit Raspberry Pi OS / ARM64** 上でBotを実行できますが、現在のDAVE対応ビルドはLinux amd64のみを正式対象としています。
そのため、Pi向けのARM64バイナリは現時点ではこのリポジトリの `build/build.sh` から生成できません。
GitHub Actionsで生成した正式対応のamd64バイナリはPiでは利用できないため、ARM64対応が追加されるまでPi構成は準備中として扱ってください。

Bot は Go 単体バイナリ + ffmpeg + systemd で動作し、VOICEVOX はPC側へ逃がす構成を推奨します。
