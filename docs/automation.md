# 🛠️ 構築自動化ガイド

Yomiage Keiryouには、Windows、Linux、WSLで同じ入口を使える構築自動化スクリプトを同梱しています。

## 必要条件

- Node.js 22以上
- Dockerモード: Docker DesktopまたはDocker Engine + Compose v2
- Buildモード: Bash、Git、Go 1.24以上、CMake、Make、`patch`、C/C++ツールチェーン
- Dockerモードの初回設定ではDiscord Bot Tokenと利用ルールへの`AGREE`入力が必要

## 最短手順

構築時は、Windows・Linux・WSLのいずれでも**リポジトリのルートにある自動化スクリプトを最初に実行**してください。`windows/docker`へ移動して`setup-env`、`up`、`logs`を順番に実行する必要はありません。

### エラー時の補助

まず環境診断を実行すると、不足しているコマンドとDocker Engineの状態を一覧表示できます。

```bash
./setup.sh --mode docker --doctor
./setup.sh --mode build --doctor
```

不足ツールをOSのパッケージ管理機能で導入してから再試行する場合は、`--repair`を付けます。Linux/WSLでは`apt-get`、Windowsでは`winget`を使用します。

```bash
./setup.sh --mode docker --repair
./setup.sh --mode build --repair
```

Windowsでは次のように実行できます。

```powershell
.\setup.ps1 -Mode docker -Doctor
.\setup.ps1 -Mode docker -Repair
```

`--repair`はNode.js、Git、Docker、Go、CMakeなどの不足コマンドを導入します。Docker Desktopそのものの初回起動、Windowsの再起動、利用規約への同意、Discord Token入力は自動完了できないため、画面の指示に従ってください。導入権限がない場合は、エラーに表示されたパッケージ名を管理者権限で導入してください。

### Windows PowerShell

```powershell
.\setup.ps1
```

PowerShellの実行ポリシーでブロックされる場合:

```cmd
setup.cmd
```

### Linux / WSL

```bash
chmod +x setup.sh
./setup.sh
```

既定のDockerモードでは、次を自動実行します。

1. Docker Engineの稼働確認
2. Docker Compose v2の確認
3. WindowsではPowerShell、Linux/WSLではBashを通じて対話型設定を起動
4. `.env`の作成と利用ルール同意の確認
5. Bot・VOICEVOX・Volume初期化コンテナのビルドと起動
6. `docker compose ps`による状態表示

## モード

### Docker構築

```bash
./setup.sh --mode docker
```

イメージを再ビルドしない場合:

```bash
./setup.sh --mode docker --no-build
```

環境設定を強制的にやり直す場合:

```bash
./setup.sh --mode docker --force-env
```

既存の`.env`を利用する場合:

```bash
./setup.sh --mode docker --skip-env
```

`--skip-env`を指定する場合は、事前に`windows/docker/.env`を作成してください。

`setup-env.ps1`、`setup-env.sh`、`up.ps1`などの個別スクリプトは、設定の再実行や復旧などの上級者向けです。通常の初回構築・再構築・確認にはルートの`setup.ps1`、`setup.cmd`、`setup.sh`を使用してください。

### Linux amd64ビルド

DAVE対応の正式ビルド対象はLinux amd64です。

```bash
./setup.sh --mode build
```

成果物:

```text
build/dist/yomiage-keiryou-amd64
```

このモードはLinux/WSL向けです。Windows PowerShellから実行する場合も、Bash、Go、CMake、C++ビルド環境が必要です。

### 検証

```bash
./setup.sh --mode verify
```

次を確認します。

- `git diff --check`
- 自動化スクリプトのNode.js構文
- PowerShellまたはBashスクリプトの構文

実行予定だけ確認する場合:

```bash
./setup.sh --mode docker --dry-run
```

PowerShell:

```powershell
.\setup.ps1 -Mode docker -DryRun
```

エラーが発生した場合も、入口スクリプトは実行コマンド・終了コード・次に実行する診断コマンドを表示します。通常の初回構築は従来どおり`--repair`なしで実行し、意図しないパッケージ導入を避けられます。

## 重要な設計

- Node.jsから実行ファイルを直接起動し、Unix専用の`&&`や`source`に依存しません。
- WindowsではPowerShellを`powershell.exe -File`として起動します。
- Linux/WSLでは`bash`を直接起動します。
- Tokenは自動化スクリプトの引数に渡しません。
- `.env`と`.yomiage-consent.jsonl`は既存の対話型設定スクリプトが管理します。
- 同意ログはローカルだけに保存され、外部送信されません。

## トラブルシューティング

### Docker Engineが起動していない

Docker Desktopを起動してから、次を確認します。

```bash
docker info
docker compose version
```

### `.env`がない

自動化スクリプトを通常モードで再実行してください。

```bash
./setup.sh --mode docker
```

### 同意版が古い

ライセンス版が更新された場合は再同意します。

```bash
./setup.sh --mode docker --force-env
```

### Dockerログを確認する

```bash
docker compose -f windows/docker/compose.yml logs -f
```

停止:

```bash
docker compose -f windows/docker/compose.yml down
```
