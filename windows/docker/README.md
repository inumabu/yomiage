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
Copy-Item .env.example .env
notepad .env
.\up.ps1
```

`DISCORD_TOKEN` を設定してから起動してください。

## 📋 操作

```powershell
.\up.ps1
.\logs.ps1
.\down.ps1
.\backup.ps1
.\restore.ps1 .\backup\settings-YYYYMMDD-HHMMSS.zip
```

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
