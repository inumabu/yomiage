# 🧯 権限トラブルシューティング

## Linux / Debian
補助スクリプトは`sudo bash ...`で実行してください。実行ユーザーには`/var/lib/yomiage-keiryou`と`/var/cache/yomiage-keiryou`の所有権が必要です。

## WSL
可能な限りリポジトリをLinuxファイルシステム内に置いてください。`/mnt/c`配下ではLinuxファイルの実行権限が不足する場合があります。必要なら`bash ./script.sh`で実行してください。

## Docker Desktop
ComposeファイルはBot起動前に、名前付きボリュームをroot権限で一度初期化します。既存ボリュームの所有者が誤っている場合は`repair-volume.ps1`を使用してください。


## 🧪 GitHub Actionsで `build/build.sh: Permission denied` が出る場合

通常のLinux amd64ビルドは、リポジトリのルートから `./setup.sh --mode build` を使用してください。GitHub上でシェルスクリプトの実行ビットが `0644` になっていてもCIが動くよう、v5.3.0以降のCIは `bash ./build/build.sh` を使用し、開始時に `.sh` の実行属性を正規化します。
低レベルのビルドスクリプトを直接呼び出す必要がある古いワークフローでは、`./build/build.sh` を `bash ./build/build.sh` に変更してください。
