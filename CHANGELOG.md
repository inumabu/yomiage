# 📝 Changelog

## v5.3.0

- 🛠️ CIを実行ビットに依存しない方式へ変更
  - `bash ./build/build.sh` でビルド
  - CI開始時に `.sh` の実行属性を正規化
- 🔧 Makefileも `bash` 経由でビルドする方式に変更
- 🪟 GitHub/ZIP経由でファイルモードが `0644` になっても `Permission denied` になりにくい構成へ変更

## v5.1

- 🧪 GitHub Actions CIを追加
  - Linux amd64 / arm64 のテスト・ビルド
  - WSL向けLinuxバイナリの生成
  - Dockerイメージのビルド
  - Git tag時のGHCR push
- 🪟 Windows系PC対応を追加
  - 🐳 Docker Desktop + WSL2 backend
  - 🐧 WSL2 / Debian native
  - PowerShellによる起動・停止・ログ・バックアップ
- 💾 Windows Dockerでも `settings.json` とTTSキャッシュをローカル永続化
- 🔒 VOICEVOX 50021はDocker内部ネットワークのみ公開

## v5

- 🧩 4運用モードを同梱
  - 🏠 自宅PC 1台・常時起動
  - ⏰ 自宅PC 1台・時間帯起動
  - 🍓 Raspberry Pi = Bot / 🖥️ PC = VOICEVOX
  - ☁️ ConoHa VPS
- 💾 TTS WAV のディスクストリーミング + ローカルキャッシュを共通化
- 🧹 キャッシュ TTL / 容量上限を共通化
- 🔨 amd64 / arm64 のビルドを選択可能化
- 🔄 systemd / timer の自己ホスト運用手順を整理
- ☁️ `settings.json` の R2 バックアップをオプション化
- 💰 ConoHa の費用比較と段階的増強ルートをREADMEへ整理


## 5.2.0
- 🧯 Linux/WSL permission doctor + repair scripts
- 🪟 Windows `up.cmd` for PowerShell execution-policy issues
- 🐳 Docker named-volume ownership initializer/repair helper
- 🐧 Fixed WSL installer directory creation order
