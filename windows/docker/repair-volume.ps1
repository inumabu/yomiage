$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if (-not (docker info 2>$null)) {
    throw 'Docker Desktopが起動していません。'
}

# yomiageイメージをroot権限の一度限りの補助として再利用します。名前付きボリューム内の所有者だけを修正し、
# settings.jsonやキャッシュは削除しません。
docker compose -f compose.yml build yomiage-volume-init
docker compose -f compose.yml run --rm --user 0 --entrypoint /bin/sh yomiage-volume-init -lc 'uid=$(id -u yomiage-keiryou); gid=$(id -g yomiage-keiryou); install -d -m 0750 /var/lib/yomiage-keiryou /var/lib/yomiage-keiryou/cache/tts /var/lib/yomiage-keiryou/tmp; chown -R "$uid:$gid" /var/lib/yomiage-keiryou'
Write-Host 'Dockerボリュームの権限を修復しました。' -ForegroundColor Green
