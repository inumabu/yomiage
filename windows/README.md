# 🪟 Windows系PC対応

`yomiage-keiryou` はWindows上でも次の2方式をサポートします。

| 方式 | おすすめ | 特徴 |
|---|---|---|
| 🐳 Docker Desktop | ⭐⭐⭐⭐⭐ | 一番簡単。Bot + VOICEVOXをまとめて管理 |
| 🐧 WSL2 / Debian | ⭐⭐⭐⭐ | Linuxとほぼ同じsystemd運用 |

## 🐳 Docker Desktop

`windows/docker/` を使用します。

```powershell
cd windows\docker
.\setup-env.ps1
.\up.ps1
```

🔐 Tokenは対話設定中に非表示入力できます。PowerShellが使えない場合は`setup-env.cmd`と`up.cmd`を使用してください。

## 🐧 WSL2 / Debian

`windows/wsl/README.md` を参照してください。

## 💡 どちらを選ぶ？

「設定してすぐ使いたい」ならDocker Desktop、「Linux VPS / Raspberry Piと同じ運用をしたい」ならWSL2 / Debianが向いています。
