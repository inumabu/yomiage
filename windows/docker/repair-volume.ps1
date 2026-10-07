$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if (-not (docker info 2>$null)) {
    throw 'Docker Desktop is not running.'
}

# Re-use the yomiage image as a one-shot root helper. This only fixes ownership
# inside the named volume; it does not delete settings.json or cache files.
docker compose -f compose.yml build yomiage-volume-init
docker compose -f compose.yml run --rm --user 0 --entrypoint /bin/sh yomiage-volume-init -lc 'uid=$(id -u yomiage-keiryou); gid=$(id -g yomiage-keiryou); install -d -m 0750 /var/lib/yomiage-keiryou /var/lib/yomiage-keiryou/cache/tts /var/lib/yomiage-keiryou/tmp; chown -R "$uid:$gid" /var/lib/yomiage-keiryou'
Write-Host 'Docker volume permissions repaired.' -ForegroundColor Green
