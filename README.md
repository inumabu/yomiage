# 🗣️ Yomiage Keiryou v5.3.4

DiscordのメッセージをVOICEVOXで読み上げる `yomiage` の配布・運用セットです。低メモリ環境での安定運用を目的に、TTS音声のディスクキャッシュ、VOICEVOX合成の直列化、systemd / Docker Desktop / WSL2向けの運用ファイルを同梱しています。

> **現在の正式ビルド対象は Linux amd64 のみです。** DiscordのDAVE/E2EE対応に必要な `libdave` のクロスビルドが整備されるまで、ARM64向けビルドは提供していません。

## 目次

- [構成の選び方](#構成の選び方)
- [最短で始める](#最短で始める)
- [必要条件](#必要条件)
- [ビルド](#ビルド)
- [環境変数](#環境変数)
- [運用モード](#運用モード)
- [Linuxでsystemd運用](#linuxでsystemd運用)
- [WindowsでDocker Desktop運用](#windowsでdocker-desktop運用)
- [WindowsのWSL2運用](#windowsのwsl2運用)
- [TTSキャッシュ](#ttsキャッシュ)
- [バックアップと復元](#バックアップと復元)
- [Discord表示とコマンド](#discord表示とコマンド)
- [トラブルシューティング](#トラブルシューティング)
- [CIとリリース](#ciとリリース)
- [ディレクトリ構成](#ディレクトリ構成)

## 構成の選び方

| 条件 | 推奨モード | 概要 |
|---|---|---|
| まず簡単に始めたい | [`modes/01-local-pc-always`](modes/01-local-pc-always/README.md) | BotとVOICEVOXを同じPCで常時起動 |
| PCを使う時間だけ起動したい | [`modes/02-local-pc-scheduled`](modes/02-local-pc-scheduled/README.md) | systemd timerでサービスを時間帯起動 |
| Piを常時起動し、VOICEVOXをPCへ分離したい | [`modes/03-pi-bot-pc-voicevox`](modes/03-pi-bot-pc-voicevox/README.md) | ただし現行のDAVE対応ARM64ビルドは準備中 |
| 外出先からも安定運用したい | [`modes/04-vps`](modes/04-vps/README.md) | VPS上でBotとVOICEVOXを同居または分離 |
| Windowsですぐ試したい | [`windows/docker/README.md`](windows/docker/README.md) | Docker DesktopでBotとVOICEVOXを管理 |
| Windows上でLinux運用したい | [`windows/wsl/README.md`](windows/wsl/README.md) | WSL2 Debian + systemd |

すでにPCがある場合は `01-local-pc-always`、Windowsで設定を簡単に済ませたい場合は Docker Desktop を推奨します。構成を後から変更する場合は [`docs/transition.md`](docs/transition.md) を参照してください。

## 最短で始める

### 🛠️ 構築自動化スクリプト

Node.jsを共通の実行入口として、Windows、Linux、WSLから同じ構築処理を実行できます。詳細は[`docs/automation.md`](docs/automation.md)を参照してください。

Windows PowerShell:

```powershell
.\setup.ps1
```

PowerShellが使えない場合:

```cmd
setup.cmd
```

Linux / WSL:

```bash
chmod +x setup.sh
./setup.sh
```

主なモード:

```bash
./setup.sh --mode docker       # Docker Desktop / Docker Composeで構築
./setup.sh --mode build        # Linux amd64向けDAVE対応ビルド
./setup.sh --mode verify       # スクリプトと構成の検証
./setup.sh --mode docker --force-env
```

### Windows + Docker Desktop

Docker DesktopのLinux containers / WSL2 backendを起動した状態で、リポジトリのルートから構築自動化スクリプトを実行します。

```powershell
.\setup.ps1
```

`setup.ps1`がDocker確認、対話型のToken・利用ルール同意、`.env`作成、イメージビルド、コンテナ起動、状態確認まで実行します。

停止:

```powershell
docker compose -f windows/docker/compose.yml down
```

PowerShellの実行ポリシーによりブロックされる場合は、ルートの`setup.cmd`を使用してください。設定項目の詳細は [`docs/env-setup.md`](docs/env-setup.md) を参照してください。

### Linux + systemd

BotとVOICEVOXを同じホストで動かす最小構成です。まずルートの自動化スクリプトでLinux amd64バイナリを作成し、その後systemdへインストールします。

```bash
./setup.sh --mode build
sudo install -m 0755 build/dist/yomiage-keiryou-amd64 /usr/local/bin/yomiage-keiryou
sudo bash ./voicevox/scripts/install.sh
sudo bash ./bot/scripts/install.sh
```

インストール後に環境ファイルを編集します。

```bash
sudoedit /etc/voicevox-engine.env
sudoedit /etc/yomiage-keiryou/yomiage.env
sudo systemctl enable --now voicevox-engine.service
sudo systemctl enable --now yomiage-keiryou.service
```

Bot単体の構成やVPS分離構成では、[`bot/README.md`](bot/README.md)、[`voicevox/README.md`](voicevox/README.md)、各モードのREADMEを先に確認してください。

## 必要条件

### DAVE対応ビルド

- Linux amd64
- Bash、Git、`patch`
- Go 1.24以上
- CMake、Make、C++ツールチェーン
- OpenSSL 3開発パッケージ
- `ffmpeg`
- `libdave`のビルドに必要なネットワークアクセス

Docker Desktop構成では、必要なビルドツールを [`windows/docker/Dockerfile`](windows/docker/Dockerfile) 内でインストールします。

### 実行環境

- Discord Bot Token
- VOICEVOX Engine
- BotとVOICEVOX間のHTTP接続
- 音声再生時の `ffmpeg`
- Linux systemd構成では、Bot実行ユーザーが設定ファイル・キャッシュディレクトリへアクセスできること

VOICEVOXのAPIポート `50021` は、同一ホスト・Docker内部ネットワーク・許可したLANホスト以外へ公開しないでください。

## ビルド

ビルドスクリプトは、上流ソースを `build/build.sh` 内の `UPSTREAM_COMMIT` に固定して取得し、次の順番で処理します。

通常は、次の自動化入口を使用してください。

```bash
./setup.sh --mode build
# 生成物: build/dist/yomiage-keiryou-amd64
```

1. 上流コミットをチェックアウト
2. キャッシュ機能とDAVE対応パッチを検証・適用
3. `libdave`をビルド
4. `go test ./...`を実行
5. Linux amd64バイナリを生成

直接`build/build.sh`を実行する方法は、上級者向けの低レベル手順です。通常の構築では`setup.sh --mode build`を使用してください。

Makeを使う場合:

```bash
make build-amd64
make archive-amd64
```

アーカイブは `yomiage-keiryou-linux-amd64.tar.gz` として作成されます。

> `TARGET_ARCH=arm64` は現在サポートされていません。Pi向けの案内は [`pi/README.md`](pi/README.md) を参照してください。

## 環境変数

### Botの基本設定

| 変数 | 既定値 | 用途 |
|---|---:|---|
| `DISCORD_TOKEN` | なし | Discord Developer Portalで発行したBot Token |
| `DISCORD_GUILD_ID` | 空欄 | 開発中にスラッシュコマンドを即時反映するGuild ID |
| `VOICEVOX_URL` | `http://localhost:50021` | VOICEVOX EngineのURL |
| `VOICEVOX_SPEAKER` | `3` | 初期話者ID |
| `YOMIAGE_SETTINGS_FILE` | `./settings.json` | Guildごとの設定保存先 |

### キャッシュとメモリ

| 変数 | 既定値 | 用途 |
|---|---:|---|
| `YOMIAGE_CACHE_DIR` | `./cache/tts` | WAVキャッシュディレクトリ |
| `YOMIAGE_CACHE_TTL` | `168h` | キャッシュの有効期間 |
| `YOMIAGE_CACHE_MAX_BYTES` | `268435456` | キャッシュ総容量の上限（バイト） |
| `YOMIAGE_MAX_AUDIO_BYTES` | `33554432` | 1音声ファイルの上限（バイト） |
| `GOMEMLIMIT` | 構成依存 | Goランタイムのメモリ上限 |
| `GOGC` | 構成依存 | Go GC比率 |

`YOMIAGE_CACHE_MAX_BYTES`と`YOMIAGE_MAX_AUDIO_BYTES`は、単位付き文字列ではなく整数のバイト数で指定します。構成別の例は各 `*.env.example` を使用してください。

## 運用モード

### 同じPCで常時起動

[`modes/01-local-pc-always/yomiage.env.example`](modes/01-local-pc-always/yomiage.env.example) を基に設定します。VOICEVOXは `127.0.0.1:50021` で待ち受けさせ、外部へ公開しません。

### 同じPCで時間帯起動

通常のBot・VOICEVOXサービスをインストールした後、[`pc-schedule/README.md`](pc-schedule/README.md) のtimerを追加します。標準設定は18:00に起動し、02:00に停止します。これはOSの電源を入れたり切ったりする設定ではありません。

### Pi + PC分離

PiからPCのVOICEVOXへ接続し、FirewallでPiのIPだけを許可します。設定例は [`modes/03-pi-bot-pc-voicevox`](modes/03-pi-bot-pc-voicevox/README.md) にあります。ただし、現行のDAVE対応バイナリはamd64のみのため、ARM64 Piでの実運用はARM64対応後に利用してください。

### VPS

1GB VPSでBotとVOICEVOXを同居させる場合は `modes/04-vps/yomiage-onehost.env.example`、分離する場合は `bot-512.env.example` と `voicevox-1gb.env.example` を使用します。メモリ使用量を実測し、必要に応じてBotとVOICEVOXを分離してください。

## Linuxでsystemd運用

主なインストール先は次のとおりです。

| パス | 内容 |
|---|---|
| `/usr/local/bin/yomiage-keiryou` | Botバイナリ |
| `/etc/yomiage-keiryou/yomiage.env` | Bot設定 |
| `/etc/voicevox-engine.env` | VOICEVOX設定 |
| `/var/lib/yomiage-keiryou` | `settings.json` |
| `/var/cache/yomiage-keiryou/tts` | TTSキャッシュ |
| `/var/backups/yomiage-keiryou` | ローカルバックアップ |

サービス確認:

```bash
systemctl status yomiage-keiryou.service voicevox-engine.service
journalctl -u yomiage-keiryou.service -f
journalctl -u voicevox-engine.service -f
```

キャッシュGCと設定バックアップのtimerを有効にする場合:

```bash
sudo systemctl enable --now yomiage-keiryou-cache-gc.timer
sudo systemctl enable --now yomiage-keiryou-backup.timer
```

## WindowsでDocker Desktop運用

詳細は [`windows/docker/README.md`](windows/docker/README.md) を参照してください。

- BotとVOICEVOXはComposeで管理します。
- `50021/tcp`はホストへ公開せず、Compose内部からのみ接続します。
- `settings.json`とTTSキャッシュは `yomiage-keiryou-data` named volumeに保存されます。
- TTSキャッシュは再生成可能なため、バックアップ対象は `settings.json`のみです。
- Volume権限の問題は `repair-volume.ps1` で修復できます。

バックアップ:

```powershell
cd windows\docker
.\backup.ps1
.\restore.ps1 .\backup\settings-YYYYMMDD-HHMMSS.zip
cd ..\..
```

## WindowsのWSL2運用

WSL2 Debianでsystemdを有効化し、GitHub Actionsの `yomiage-keiryou-wsl-binaries` artifactからamd64バイナリを取得します。手順は [`windows/wsl/README.md`](windows/wsl/README.md) にあります。

```bash
sudo bash ./windows/wsl/install-bot.sh ./yomiage-keiryou-amd64
sudo systemctl enable --now yomiage-keiryou.service
```

WSL内で実行するスクリプトは、Windowsマウント配下の実行ビット問題を避けるため、必要に応じて `bash ./script.sh` の形式で起動してください。

## TTSキャッシュ

- キャッシュキーは本文・話者・音量・速度からSHA-256で生成します。
- 同じ設定の音声は再利用し、VOICEVOXへの再合成を減らします。
- 合成は全Guild共通で1件ずつ実行し、低メモリ環境でのピーク使用量を抑えます。
- 一時ファイルをキャッシュディレクトリ内に作成し、完成後にrenameするため、不完全なWAVは再利用しません。
- TTL超過または総容量超過時は古いWAVから削除します。
- キャッシュは削除しても再生成できるため、設定バックアップには含めません。

手動でGCを実行する場合:

```bash
sudo systemctl start yomiage-keiryou-cache-gc.service
```

## バックアップと復元

Linuxでは `settings.json` のバックアップを毎日実行できます。ローカルバックアップの保持期間は `bot/backup.env.example` の `LOCAL_RETENTION_DAYS` で指定します。Cloudflare R2へのアップロードは、必要な資格情報を設定したうえで `R2_ENABLED=1` にした場合のみ実行されます。

```bash
sudo systemctl start yomiage-keiryou-backup.service
sudo /usr/local/libexec/yomiage-keiryou/restore.sh /var/backups/yomiage-keiryou/yomiage-settings-YYYYMMDD-HHMMSS.tar.gz
```

Discord Token、`.env`、R2のアクセスキー、バックアップアーカイブはGitへコミットしないでください。Tokenやアクセスキーをログ・Issue・スクリーンショットへ貼らないでください。

## Discord表示とコマンド

- `/status`、`/queue`、`/help`、設定系スラッシュコマンドの結果は、実行者だけに見えるEphemeral Embedです。
- ボタン操作の結果も、操作した本人だけに表示されます。
- 通常メッセージコマンドとBotの障害通知は、チャンネル全体に表示される公開Embedです。
- 成功は `✅`、注意は `⚠️`、失敗は `❌` で表示します。
- 管理者権限が必要な設定変更があります。Botに必要なDiscord権限を付与してください。

## 📜 利用ルール・ライセンス

このプロジェクトは、OSI承認のオープンソースライセンスではない、作者`inumabu`の**カスタムライセンス**で公開しています。詳細は[`LICENSE`](LICENSE)を確認してください。

### ✅ 許可される利用

- 個人利用、学習、研究、改造
- 営利・非営利を問わない利用
- 本コードを利用したソフトウェア、サービス、作品の作成
- 友人・知人、所属コミュニティ、開発チーム内での共有
- 本リポジトリへのリンクを使った紹介・共有

### ⚠️ 禁止・制限される利用

- 本コードや改変物を自分が作成したものだと主張すること
- コードを不特定多数が取得できる形で再配布・転載すること
- 作者や出典が分からない状態でコードを配布すること
- マルウェア、不正アクセス、情報窃取、詐欺など第三者へ損害を与える目的での利用

作者名やリポジトリへのリンクを添えた紹介を歓迎します。利用前の連絡は必要ありませんが、再配布・改変時は`LICENSE`の条件を守ってください。

## トラブルシューティング

### `VOICEVOX_URL`へ接続できない

1. VOICEVOX Engineが起動しているか確認します。
2. `VOICEVOX_URL`のホスト名・ポートを確認します。
3. Docker構成では `http://voicevox:50021`、同一ホストのsystemd構成では `http://127.0.0.1:50021` を使用します。
4. 分離構成ではFirewallがBotホストからのTCP `50021`を許可しているか確認します。

### `Permission denied` が出る

systemdの実行ユーザー、`settings.json`、キャッシュディレクトリの所有者を確認してください。

```bash
sudo bash ./bot/scripts/doctor-permissions.sh
sudo bash ./bot/scripts/repair-permissions.sh
```

詳細は [`docs/permission-troubleshooting.md`](docs/permission-troubleshooting.md) を参照してください。

### DAVE / libdaveのビルドに失敗する

- Goが1.24以上か確認します。
- CMake、C++コンパイラ、OpenSSL 3、`patch`が入っているか確認します。
- `build/build.sh`はLinux amd64以外を受け付けません。
- 上流コミットやパッチを変更した場合は、パッチの再生成と検証が必要です。

### Docker named volumeの権限エラー

Windows Docker構成では、次を実行してからBotを再起動します。

```powershell
cd windows\docker
.\repair-volume.ps1
cd ..\..
.\setup.ps1 -NoBuild
```

この修復処理はVolume内の所有者だけを変更し、`settings.json`やキャッシュを削除しません。

## CIとリリース

GitHub Actionsは次を実行します。

- DAVE対応Linux amd64のテスト・ビルド
- WSL用amd64バイナリのartifact生成
- `windows/docker/Dockerfile`のDockerイメージビルド
- Git tag push時のGHCR公開

CI定義は [`.github/workflows/ci.yml`](.github/workflows/ci.yml) にあります。リリース番号は [`VERSION`](VERSION) とCHANGELOGの先頭エントリを揃えて更新してください。

## ディレクトリ構成

```text
.
├── build/                 # 上流取得・パッチ適用・DAVEビルド
├── patches/               # 再現可能な差分
├── bot/                   # Linux Botの設定、systemd、バックアップ
├── voicevox/              # VOICEVOXのsystemd、Firewall、swap
├── modes/                 # 運用モード別のenv例と説明
├── pc-schedule/           # systemd timerによる時間帯起動
├── windows/docker/        # Docker Desktop用Composeと補助スクリプト
├── windows/wsl/            # WSL2用インストール補助
├── docs/                  # モード選択、移行、設定、権限トラブル
└── .github/workflows/     # CI / Docker / artifact
```

## ライセンス

ライセンスの詳細は [`LICENSE`](LICENSE) を参照してください。
