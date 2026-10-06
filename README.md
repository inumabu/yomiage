# 🔊 Discord 読み上げ BOT

discordgo と VOICEVOX Engine を使い、テキストチャンネルの投稿をボイスチャンネルで読み上げます。

## 📋 必要なもの

- Go 1.22 以降、または Docker
- ffmpeg
- 起動中の VOICEVOX Engine（標準URL: `http://localhost:50021`）
- DiscordアプリのBotトークン

Discord Developer Portal で **Message Content Intent** を有効にしてください。Botにはメッセージ閲覧とボイスチャンネル接続・発話の権限が必要です。

## 🚀 起動

```sh
cp .env.example .env
export DISCORD_TOKEN="Botトークン"
export DISCORD_GUILD_ID="123456789012345678" # 開発中は指定推奨
go run .
```

PowerShell では次のように設定します。

```powershell
Copy-Item .env.example .env
$env:DISCORD_TOKEN = "Botトークン"
$env:DISCORD_GUILD_ID = "123456789012345678"
go run .
```

設定は `settings.json` にサーバー単位で保存されます。保存先は `YOMIAGE_SETTINGS_FILE` で変更できます。`VOICEVOX_URL`、`VOICEVOX_SPEAKER` も環境変数で変更できます。

## 🧭 コマンド

| 機能 | メッセージコマンド | スラッシュコマンド | 権限 |
|---|---|---|---|
| 接続・切断 | `!join` / `!leave` | `/join` / `/leave` | 全員 |
| 任意文の読み上げ | `!say 文章` | `/say text:文章` | 全員 |
| 話者変更 | `!speaker 3` | `/speaker id:3` | 管理者 |
| 音量変更 | `!volume 1.2` | `/volume value:1.2` | 管理者 |
| 速度変更 | `!speed 1.1` | `/speed value:1.1` | 管理者 |
| 状態・キュー確認 | `!status` / `!queue` | `/status` / `/queue` | 全員 |
| キュー指定削除 | `!remove 2` | `/remove index:2` | 管理者 |
| 待機キュー消去 | `!clear` | `/clear` | 管理者 |
| 現在の文をスキップ | `!skip` | `/skip` | 管理者 |
| 次の文を一時停止・再開 | `!pause` / `!resume` | `/pause` / `/resume` | 管理者 |
| 話者一覧 | `!speakers` | `/speakers` | 全員 |
| ヘルプ | `!help` | `/help` | 全員 |

音量は `0.0〜2.0`、速度は `0.5〜2.0` です。設定は再起動後も保持されます。`!queue` は待機文の一覧を表示し、`!remove 2` のように指定番号を削除できます。`/skip` は再生中の文だけをキャンセルし、次のキューへ進みます。`/pause` は次の読み上げを停止し、`/resume` で再開します。

## 🛡️ 読み上げ対象の制御（管理者）

```text
!channel allow チャンネルID   # allowを1つでも設定すると、そのチャンネルだけ対象
!channel block チャンネルID   # チャンネルを除外
!channel remove チャンネルID  # チャンネル設定を解除
!user block ユーザーID         # ユーザーを除外
!user remove ユーザーID        # ユーザー除外を解除
```

スラッシュコマンドでは `/channel action:allow id:...`、`/user action:block id:...` の形式です。除外設定も `settings.json` に保存されます。

## ✅ 仕様・安定性

- 投稿はサーバーごとのキューで順番に処理します。1投稿120文字超は読み上げません。
- キュー容量は30件です。満杯時もDiscordイベント処理をブロックしません。
- VOICEVOXの音声生成に一時的な失敗があった場合は1回自動再試行します。
- 接続中の音声処理は終了時にキャンセルされ、一時ファイルも削除されます。
- メッセージ応答には接続・終了・状態・キュー確認・全消去のボタンが表示されます。

## 🐳 Docker

VOICEVOX Engineと同じネットワークで起動してください。

```sh
docker build -t yomiage .
docker run --rm --env-file .env -v "$PWD/data:/data" yomiage
```

Dockerでは既定のVOICEVOX URLが `http://voicevox:50021`、設定保存先が `/data/settings.json` です。

## 📌 スラッシュコマンドの登録

`DISCORD_GUILD_ID` を指定すると対象サーバーへ登録します。省略時はグローバル登録を試みますが、反映に時間がかかる場合があります。

## 🧪 開発

```sh
go test ./...
go test -race ./...
go vet ./...
```

GitHub Actionsでもテスト、race検出、`go vet`を自動実行します。
