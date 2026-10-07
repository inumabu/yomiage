# 🧯 Permission troubleshooting

## Linux / Debian
Run helper scripts with `sudo bash ...`. The runtime user must own `/var/lib/yomiage-keiryou` and `/var/cache/yomiage-keiryou`.

## WSL
Keep the repository inside the Linux filesystem when possible. `/mnt/c` can lack executable permissions for Linux files. Use `bash ./script.sh` if needed.

## Docker Desktop
The Compose file runs a one-shot root initializer against the named volume before starting the Bot. If an existing volume has the wrong owner, use `repair-volume.ps1`.


## 🧪 GitHub Actionsで `build/build.sh: Permission denied` が出る場合

GitHub上でシェルスクリプトの実行ビットが `0644` になっていてもCIが動くよう、v5.3.2以降のCIは `bash ./build/build.sh` を使用し、開始時に `.sh` の実行属性を正規化します。
古いワークフローを使っている場合は `./build/build.sh` を `bash ./build/build.sh` に変更してください。
