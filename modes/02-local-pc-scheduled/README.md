# ⏰ 自宅PC 1台・時間帯起動

## 構成

```text
PC
├─ yomiage-keiryou
└─ VOICEVOX Engine
```

18:00〜02:00 のような利用時間だけ systemd timer で起動できます。

```bash
sudo install -m 0644 ../../pc-schedule/systemd/*.service ../../pc-schedule/systemd/*.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now yomiage-keiryou-stack-start.timer
sudo systemctl enable --now yomiage-keiryou-stack-stop.timer
```

⚠️ PC自体の電源断まで自動化する設定ではありません。OSが起動している間にサービスを開始・停止します。
