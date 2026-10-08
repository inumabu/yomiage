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

通常はリポジトリのルートに戻り、構築自動化スクリプトを実行してください。個別スクリプトを順番に実行する必要はありません。

```powershell
cd ..\..
\.\setup.ps1
```

🔐 `setup.ps1`がDiscord Bot Tokenを非表示で受け取り、利用ルールの同意、`.env`作成、Dockerイメージのビルド、コンテナ起動まで実行します。

PowerShellが使えない場合:

```cmd
cd ..\..
setup.cmd
```

Linux/WSLでは次を使います。

```bash
cd ../..
./setup.sh
```

🔁 設定をやり直す場合は、既存`.env`を保護するため明示的に強制指定します。

```powershell
cd ..\..
\.\setup.ps1 -ForceEnv
```

```bash
./setup.sh --force-env
```

💡 `.env`は秘密情報を含むため、Git・Issue・スクリーンショットへ貼り付けないでください。

## 📋 操作

```powershell
docker compose -f windows/docker/compose.yml ps
docker compose -f windows/docker/compose.yml logs -f
docker compose -f windows/docker/compose.yml down
docker compose -f windows/docker/compose.yml up -d
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
setup.cmd
```

Docker named volume の所有者エラーが `settings.json` や `/var/lib/yomiage-keiryou` に対して出る場合:

```powershell
cd windows\docker
\.\repair-volume.ps1
cd ..\..
\.\setup.ps1 -NoBuild
```

この修復は named volume の所有者だけを直し、データを削除しません。
