# 🏠 自宅PC 1台・常時起動

## 構成

```text
PC
├─ yomiage-keiryou
├─ VOICEVOX Engine (127.0.0.1:50021)
└─ ffmpeg
```

外部から VOICEVOX の 50021/tcp にアクセスさせる必要はありません。

## 設定

`yomiage.env.example` を `/etc/yomiage-keiryou/yomiage.env` のベースにします。

```dotenv
VOICEVOX_URL=http://127.0.0.1:50021
```
