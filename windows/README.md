# 🪟 Windows系PC対応

`yomiage-keiryou` はWindows上でも次の2方式をサポートします。

| 方式 | おすすめ | 特徴 |
|---|---|---|
| 🐳 Docker Desktop | ⭐⭐⭐⭐⭐ | 一番簡単。Bot + VOICEVOXをまとめて管理 |
| 🐧 WSL2 / Debian | ⭐⭐⭐⭐ | Linuxとほぼ同じsystemd運用 |

## 🐳 Docker Desktop

リポジトリのルートにある構築自動化スクリプトを使用します。

```powershell
.\setup.ps1
```

🔐 Tokenは対話設定中に非表示入力できます。利用ルールへの`AGREE`入力、Docker構築、コンテナ起動まで自動で実行されます。PowerShellが使えない場合はルートの`setup.cmd`を使用してください。

## 🐧 WSL2 / Debian

`windows/wsl/README.md` を参照してください。

## 💡 どちらを選ぶ？

「設定してすぐ使いたい」ならDocker Desktop、「Linux VPS / Raspberry Piと同じ運用をしたい」ならWSL2 / Debianが向いています。
