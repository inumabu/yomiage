$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if (-not (Test-Path .env)) {
    Copy-Item .env.example .env
    Write-Host "Created .env. Edit DISCORD_TOKEN, then run this script again." -ForegroundColor Yellow
    exit 1
}

docker compose -f compose.yml up -d --build
docker compose -f compose.yml ps
