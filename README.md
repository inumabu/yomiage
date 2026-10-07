# 🗣️ Yomiage Keiryou v5.3.2

DiscordのメッセージをVOICEVOXで読み上げる`yomiage`を、低メモリ環境向けに調整した配布・運用セットです。
TTS音声をディスクキャッシュへ保存し、合成の同時実行を直列化することで、VPS・Docker Desktop・WSL2・Raspberry Pi構成での安定運用を目指します。

## 含まれるもの

- `build/`: 固定した上流コミットへキャッシュ機能とDAVE対応を適用し、Linux amd64をビルド
- `patches/`: 上流`inumabu/yomiage`への再現可能なパッチ
- `windows/docker/`: Windows 10/11 + Docker Desktop + WSL2向け構成
- `windows/wsl/`: WSLから実行するLinuxバイナリの例
- `modes/`: 常駐、スケジュール、Pi、VPSの運用モード
- `pc-schedule/`: systemd timerによるPC起動・停止スケジュール
- `bot/systemd/`、`voicevox/systemd/`: Linuxサービス定義、バックアップ、キャッシュGC

## 🚀 最短手順: Windows + Docker Desktop

```powershell
cd windows/docker
.\setup-env.ps1
.\up.ps1
.\logs.ps1
```

🔐 `setup-env.ps1`でTokenを非表示入力できます。`.\up.ps1`は`.env`が未作成またはToken未設定の場合、自動的に対話設定を開始します。

停止は以下です。

```powershell
.\down.ps1
```

PowerShellの実行ポリシーでブロックされる場合は、同じディレクトリの`up.cmd`、`down.cmd`を使えます。

## Linuxビルド

必要環境: Bash、Git、Go 1.24以上、CMake、C++ツールチェーン、`patch`。

```bash
TARGET_ARCH=amd64 bash ./build/build.sh
```

DAVE対応のlibdaveは現在Linux amd64を正式対象としています。arm64はlibdaveのクロスビルド対応後に追加予定です。

生成物:

```text
build/dist/yomiage-keiryou-amd64
```

Makeを使う場合:

```bash
make build-amd64
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
- [🔐 対話型env設定](docs/env-setup.md)
- [PCスケジュール](pc-schedule/README.md)

## CI

GitHub Actionsでは、DAVE対応Linux amd64のテスト・ビルド、WSL amd64用バイナリ生成、Dockerイメージのビルドを行います。タグPush時のみGHCRへ公開します。

## 注意事項

- Discordトークンや`.env`、バックアップZIPはGitへコミットしないでください。
- VOICEVOX EngineとBotは同じホストで常駐させる場合、メモリ上限を確認してください。
- `patches/0001-keiryou-cache.patch`は指定した上流コミット専用です。上流コミットを変更する場合は、差分の再生成とテストが必要です。

## 📨 Discord表示の使い分け

- `/status`、`/queue`、`/help`、設定系コマンドの結果は、実行者だけに見える`Ephemeral Embed`です。
- ボタン操作の結果も、押した本人だけに表示されます。
- 通常メッセージコマンドとBotの障害通知は、チャンネル全体に表示される公開Embedです。
- `✅`成功、`⚠️`注意、`❌`失敗を色と絵文字で統一しています。

## 🔐 DAVE対応

DiscordのE2EE/DAVE必須化に対応するため、`github.com/aleph-garden/discordgo`とDiscord公式`libdave`を使用します。Docker/CIのビルドにはGo 1.24、CGO、CMake、OpenSSL 3、C++ツールチェーンが必要です。現在の正式ビルド対象はLinux amd64です。
