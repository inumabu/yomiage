$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if (-not (Test-Path .env)) {
    & (Join-Path $PSScriptRoot 'setup-env.ps1')
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

$tokenLine = Select-String -Path .env -Pattern '^DISCORD_TOKEN=([^\s].*)$' -ErrorAction SilentlyContinue
if (-not $tokenLine) {
    & (Join-Path $PSScriptRoot 'setup-env.ps1')
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

docker compose -f compose.yml up -d --build
docker compose -f compose.yml ps
