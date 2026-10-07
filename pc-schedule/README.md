# ⏰ 自宅PCの時間帯起動

このディレクトリは `02-local-pc-scheduled` 用です。

## 🕒 例: 18:00〜02:00

まず通常の `voicevox-engine.service` と `yomiage-keiryou.service` がインストール済みであることを確認します。

その後、timer を入れます。

```bash
sudo install -m 0644 systemd/*.service systemd/*.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now yomiage-keiryou-stack-start.timer
sudo systemctl enable --now yomiage-keiryou-stack-stop.timer
```

時刻は以下を編集して変更します。

- `systemd/yomiage-keiryou-stack-start.timer`
- `systemd/yomiage-keiryou-stack-stop.timer`

## ⚠️ 注意

`02:00` に停止すると、その後に届いたDiscordメッセージは処理できません。利用開始時刻より少し前に起動する運用をおすすめします。
