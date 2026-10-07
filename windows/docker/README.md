# 🪟 Windows + Docker Desktop

Windows 10/11 + Docker Desktop の Linux containers / WSL2 backend を使う構成です。

## 🧩 構成

```text
Windows
└─ Docker Desktop
   ├─ yomiage-keiryou
   └─ VOICEVOX Engine
```

VOICEVOX の `50021/tcp` はホストへ公開せず、Docker 内ネットワークだけで利用します。

## 🚀 起動

PowerShell でこのディレクトリに移動し、`up.ps1` を実行します。PowerShell の実行ポリシーでブロックされる場合は `up.cmd` を使えます。

```powershell
.\setup-env.ps1
.\up.ps1
```

🔐 `setup-env.ps1`ではDiscord Bot Tokenを非表示で入力できます。`.env`が存在しない状態で`.\up.ps1`を実行した場合も、自動的に対話設定が始まります。

PowerShellが使えない場合:

```cmd
setup-env.cmd
up.cmd
```

Linux/WSLでは次を使います。

```bash
./setup-env.sh
docker compose -f compose.yml up -d --build
```

🔁 設定をやり直す場合は、既存`.env`を保護するため明示的に強制指定します。

```powershell
.\setup-env.ps1 -Force
```

```bash
./setup-env.sh --force
```

💡 `.env`は秘密情報を含むため、Git・Issue・スクリーンショットへ貼り付けないでください。

## 📋 操作

```powershell
.\up.ps1
.\logs.ps1
.\down.ps1
.\backup.ps1
.\restore.ps1 .\backup\settings-YYYYMMDD-HHMMSS.zip
```

## 📨 Discordメッセージの表示範囲

- `/status`、`/queue`、`/help`などのスラッシュコマンド結果は、**実行した本人だけが見られる非公開Embed**です。
- ボタン操作の結果も、**ボタンを押した本人だけが見られる非公開Embed**です。
- 通常のメッセージコマンドや、読み上げ失敗などのBot通知は、チャンネル全体に見える**公開Embed**です。
- 成功は`✅`、注意・失敗は`⚠️`/`❌`の色と絵文字で区別します。

## 💾 データ

`yomiage-keiryou-data` というDocker named volumeに `settings.json` と TTS キャッシュが保存されます。TTS キャッシュは再生成可能なのでバックアップ対象外です。

## 🧠 メモリ

Docker Desktop のVM/WSL側メモリ設定とWindowsホストのRAMには余裕を持たせてください。VPSの「512MB」という数字を、そのままWindows Docker環境の必要メモリとはみなさないでください。

## 🛠️ GHCR

GitHub Actions のタグビルドでは `ghcr.io/<owner>/<repo>` へ yomiage-keiryou イメージをpushできます。公開レジストリを使いたくない場合は、ローカルbuildのままで利用できます。


## 🧯 Permission denied が出たら

`.ps1` が実行できない場合:

```cmd
up.cmd
```

Docker named volume の所有者エラーが `settings.json` や `/var/lib/yomiage-keiryou` に対して出る場合:

```powershell
.\repair-volume.ps1
.\up.ps1
```

この修復は named volume の所有者だけを直し、データを削除しません。
