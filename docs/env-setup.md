# 🔐 対話型 `.env` 設定ガイド

Yomiage Keiryou v5.3.4のDocker構成は、初回起動前にDiscord Bot Tokenなどを対話形式で設定できます。Tokenは入力中に画面へ表示されません。

初回設定では、[`LICENSE`](../LICENSE)の利用ルールを表示し、`AGREE`の入力を求めます。同意しない場合は設定を完了できません。

## 推奨: 構築自動化スクリプト

通常は個別の`setup-env`や`up`を実行せず、リポジトリのルートから構築自動化スクリプトを実行してください。

Windows PowerShell:

```powershell
\.\setup.ps1
```

PowerShellの実行ポリシーでブロックされる場合:

```cmd
setup.cmd
```

Linux / WSL:

```bash
./setup.sh
```

自動化スクリプトは、Docker確認、対話型設定、同意確認、Dockerイメージのビルド、コンテナ起動、状態表示をまとめて実行します。個別スクリプトはトラブルシューティングや高度な運用が必要な場合だけ使用してください。

## 個別設定が必要な場合

自動化スクリプトを使わず、環境設定だけをやり直す場合は、設定ディレクトリから個別スクリプトを実行できます。

```powershell
cd windows/docker
\.\setup-env.ps1
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

設定完了後はリポジトリのルートへ戻り、次のコマンドで構築できます。

```bash
cd ../..
./setup.sh --skip-env
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
| `YOMIAGE_LICENSE_ACCEPTED` | 自動 | `true` | 利用ルールへの同意済みフラグ |
| `YOMIAGE_LICENSE_VERSION` | 自動 | `2026-10-08` | 同意したルールの版 |

## セキュリティ

- `.env`はGitへコミットしないでください。
- Tokenをチャット、Issue、ログ、スクリーンショットへ貼らないでください。
- Tokenが漏えいした場合はDiscord Developer Portalで即時再生成してください。
- Linux/WSLでは`.env`を`chmod 600`で保存します。
- 設定の再生成は既存ファイルを自動上書きしません。明示的に`--force`または`-Force`を指定してください。

## 🧾 同意ログの最小収集

同意確認の記録は、設定ディレクトリ内の`.yomiage-consent.jsonl`へ**ローカル保存**します。中央サーバーへの送信や外部アップロードは行いません。

記録するのは次の項目だけです。

- 同意イベント名
- `accepted: true`
- ライセンス版
- UTCの同意日時
- 実行経路（`setup-env.ps1`または`setup-env.sh`）

次の情報は記録・収集しません。

- DiscordユーザーID、Guild ID
- Discord Bot Token
- IPアドレス、ホスト名、OSユーザー名
- メッセージ本文、音声本文、利用コマンド

Linux/WSLでは同意ログも`0600`で保存します。ルールが更新されると版が変わり、次回設定時に再同意が必要になります。ログを削除したい場合は、Botを停止してから`.yomiage-consent.jsonl`を削除してください。削除してもBotの動作には影響しません。

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
