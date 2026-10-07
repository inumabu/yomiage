$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
docker compose -f compose.yml logs -f --tail=100
