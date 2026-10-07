# 🍓 Raspberry Pi + PC

## 役割

- Raspberry Pi: `yomiage-keiryou` を常時起動
- PC: VOICEVOX Engine を常時または時間帯起動

## ネットワーク例

```text
Pi   192.168.1.10
PC   192.168.1.20
```

Pi:

```dotenv
VOICEVOX_URL=http://192.168.1.20:50021
```

PC:

```dotenv
VOICEVOX_HOST=192.168.1.20
BOT_PRIVATE_IP=192.168.1.10
LAN_CIDR=192.168.1.0/24
```

VOICEVOX は PC の LAN IP に bind し、Firewall で Pi からのみ許可します。
