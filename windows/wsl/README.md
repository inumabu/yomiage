# 🪟 Windows + WSL2 / Debian

Docker Desktopを使わず、Windows上のWSL2 DebianでLinux版を直接実行する構成です。

## 1️⃣ Debianを用意

PowerShell（管理者）:

```powershell
wsl --install -d Debian
```

既存のDebianを使う場合は不要です。

## 2️⃣ systemdを有効化

Debian内の `/etc/wsl.conf` に以下を追加します。

```ini
[boot]
systemd=true
```

PowerShellで:

```powershell
wsl --shutdown
```

その後Debianを起動し直します。

## 3️⃣ 必要パッケージ

```bash
sudo apt update
sudo apt install -y ffmpeg ca-certificates
```

BotバイナリはGitHub Actionsの `yomiage-keiryou-wsl-binaries` artifactから取得するか、リポジトリのルートで自動化スクリプトを実行して生成します。

```bash
./setup.sh --mode build
```

## 4️⃣ Bot + VOICEVOX

VOICEVOXもWSL内で動かす場合は、`voicevox/` の既存systemd手順を利用します。

```dotenv
VOICEVOX_URL=http://127.0.0.1:50021
```

BotだけWSL、VOICEVOXをWindows側のPCへ置く場合は、VOICEVOXの到達可能なLAN/ホストIPへ変更します。

> ⚠️ Windowsホストの`localhost`到達性はWindows/WSLネットワークモードによって挙動が異なる場合があります。安定性を優先するなら、両方をWSL内に置くか、Docker Desktop構成を推奨します。

## 🔨 Botのインストール補助

GitHub Actionsで生成したamd64 Linux artifactをWSL内へ置いた後、次のようにインストールできます。

```bash
sudo bash ./windows/wsl/install-bot.sh ./yomiage-keiryou-amd64
```

その後 `/etc/yomiage-keiryou/yomiage.env` を編集してください。VOICEVOXを同じWSL内で動かすなら `VOICEVOX_URL=http://127.0.0.1:50021`、別のLAN PCならそのIPを指定します。


## 🧯 Permission denied

WSL で `/mnt/c/...` 配下のファイルを `./script.sh` のように直接実行すると、Windows 側の実行属性が原因で `Permission denied` になることがあります。

```bash
cd ~/src
git clone https://github.com/inumabu/yomiage-keiryou.git
cd yomiage-keiryou
./setup.sh --mode build
```

Linux 側に配置できない事情がある場合も、通常は`./setup.sh --mode build`を使用してください。低レベルのスクリプトを直接実行する場合だけ、`bash ./...sh`の形式で起動してください。
