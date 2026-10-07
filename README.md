# 🗣️ Yomiage Keiryou v5.3

DiscordのメッセージをVOICEVOXで読み上げる`yomiage`を、低メモリ環境向けに調整した配布・運用セットです。
TTS音声をディスクキャッシュへ保存し、合成の同時実行を直列化することで、VPS・Docker Desktop・WSL2・Raspberry Pi構成での安定運用を目指します。

## 含まれるもの

- `build/`: 固定した上流コミットへキャッシュ機能を適用し、Linux amd64/arm64をビルド
- `patches/`: 上流`inumabu/yomiage`への再現可能なパッチ
- `windows/docker/`: Windows 10/11 + Docker Desktop + WSL2向け構成
- `windows/wsl/`: WSLから実行するLinuxバイナリの例
- `modes/`: 常駐、スケジュール、Pi、VPSの運用モード
- `pc-schedule/`: systemd timerによるPC起動・停止スケジュール
- `bot/systemd/`、`voicevox/systemd/`: Linuxサービス定義、バックアップ、キャッシュGC

## 最短手順: Windows + Docker Desktop

```powershell
cd windows/docker
Copy-Item .env.example .env
notepad .env
.\up.ps1
.\logs.ps1
```

`.env`へ`DISCORD_TOKEN`を設定してください。停止は以下です。

```powershell
.\down.ps1
```

PowerShellの実行ポリシーでブロックされる場合は、同じディレクトリの`up.cmd`、`down.cmd`を使えます。

## Linuxビルド

必要環境: Bash、Git、Go 1.22以上、`patch`。

```bash
TARGET_ARCH=amd64 bash ./build/build.sh
TARGET_ARCH=arm64 bash ./build/build.sh
```

生成物:

```text
build/dist/yomiage-keiryou-amd64
build/dist/yomiage-keiryou-arm64
```

Makeを使う場合:

```bash
make build-amd64
make build-arm64
make archive-amd64
```

ビルドスクリプトは`UPSTREAM_COMMIT`に固定された上流ソースを取得し、パッチ検証、`go test ./...`、クロスビルドを順番に行います。

## 主な環境変数

| 変数 | 既定値 | 用途 |
| --- | --- | --- |
| `DISCORD_TOKEN` | なし | Discord Botトークン |
| `VOICEVOX_URL` | `http://localhost:50021` | VOICEVOX EngineのURL |
| `VOICEVOX_SPEAKER` | `3` | 話者ID |
| `YOMIAGE_SETTINGS_FILE` | 上流実装の既定値 | サーバー設定ファイル |
| `YOMIAGE_CACHE_DIR` | `./cache/tts` | WAVキャッシュディレクトリ |
| `YOMIAGE_CACHE_TTL` | `168h` | キャッシュの有効期間 |
| `YOMIAGE_CACHE_MAX_BYTES` | `268435456` | キャッシュ上限 |
| `YOMIAGE_MAX_AUDIO_BYTES` | `33554432` | 1音声の最大サイズ |

## キャッシュ仕様

- キャッシュキーは本文・話者・音量・速度からSHA-256で生成します。
- 同じ設定の音声は再利用し、VOICEVOXへの不要な再合成を避けます。
- 合成は一度に1件へ制限し、低メモリ環境でのピーク使用量を抑えます。
- 一時ファイルを同じキャッシュディレクトリに作成し、完成後に`rename`するため、途中のWAVを再利用しません。
- TTL超過または総容量超過時は古いキャッシュから削除します。

## 運用ドキュメント

- [モード選択](docs/mode-selection.md)
- [移行手順](docs/transition.md)
- [権限トラブルシューティング](docs/permission-troubleshooting.md)
- [Windows + Docker](windows/docker/README.md)
- [WSL](windows/wsl/README.md)
- [VOICEVOX](voicevox/README.md)
- [PCスケジュール](pc-schedule/README.md)

## CI

GitHub Actionsでは、Linux amd64/arm64のテスト・ビルド、WSL用バイナリ生成、Dockerイメージのビルドを行います。タグPush時のみGHCRへ公開します。

## 注意事項

- Discordトークンや`.env`、バックアップZIPはGitへコミットしないでください。
- VOICEVOX EngineとBotは同じホストで常駐させる場合、メモリ上限を確認してください。
- `patches/0001-keiryou-cache.patch`は指定した上流コミット専用です。上流コミットを変更する場合は、差分の再生成とテストが必要です。
