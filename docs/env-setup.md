# 🔐 対話型 `.env` 設定ガイド

Yomiage Keiryou v5.3.3のDocker構成は、初回起動前にDiscord Bot Tokenなどを対話形式で設定できます。Tokenは入力中に画面へ表示されません。

## Windows PowerShell

```powershell
cd windows/docker
.\setup-env.ps1
```

PowerShellの実行ポリシーでブロックされる場合:

```cmd
setup-env.cmd
```

その後に起動します。

```powershell
.\up.ps1
```

## Windowsで一度に起動

`.env`が存在しない、またはToken未設定の場合、`up.ps1`が自動的に設定スクリプトを呼び出します。

```powershell
.\up.ps1
```

## Linux / WSL

```bash
cd windows/docker
./setup-env.sh
```

初回設定をやり直す場合:

```bash
./setup-env.sh --force
```

```bash
docker compose -f compose.yml up -d --build
```

## 設定項目

| 項目 | 必須 | 既定値 | 説明 |
| --- | --- | --- | --- |
| `DISCORD_TOKEN` | 必須 | なし | Discord Developer PortalのBot Token。画面非表示入力 |
| `DISCORD_GUILD_ID` | 任意 | 空欄 | 開発中のコマンド即時反映用Guild ID |
| `VOICEVOX_SPEAKER` | 任意 | `3` | 初期話者ID |
| `VOICEVOX_IMAGE` | 任意 | `voicevox/voicevox_engine:latest` | VOICEVOXイメージ |
| `GOMEMLIMIT` | 任意 | `128MiB` | Goのメモリ上限 |
| `GOGC` | 任意 | `50` | Go GC比率 |
| `YOMIAGE_CACHE_TTL` | 任意 | `168h` | TTSキャッシュ保持期間 |
| `YOMIAGE_CACHE_MAX_BYTES` | 任意 | `268435456` | キャッシュ上限 |
| `YOMIAGE_MAX_AUDIO_BYTES` | 任意 | `33554432` | 1音声の上限 |

## セキュリティ

- `.env`はGitへコミットしないでください。
- Tokenをチャット、Issue、ログ、スクリーンショットへ貼らないでください。
- Tokenが漏えいした場合はDiscord Developer Portalで即時再生成してください。
- Linux/WSLでは`.env`を`chmod 600`で保存します。
- 設定の再生成は既存ファイルを自動上書きしません。明示的に`--force`または`-Force`を指定してください。

## 設定確認

Tokenそのものを表示せず、Composeの展開結果を確認します。

```bash
docker compose -f compose.yml config --quiet
```

Windowsでは:

```powershell
docker compose -f compose.yml config --quiet
```

`docker compose config`の通常出力には環境変数が含まれる場合があるため、共有ログへ貼り付けないでください。
