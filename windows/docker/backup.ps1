$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
New-Item -ItemType Directory -Force -Path backup | Out-Null

$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$dir = Join-Path $PSScriptRoot "backup\tmp-$stamp"
$dest = Join-Path $PSScriptRoot "backup\settings-$stamp.zip"
New-Item -ItemType Directory -Force -Path $dir | Out-Null
try {
    docker compose -f compose.yml cp yomiage:/var/lib/yomiage-keiryou/settings.json (Join-Path $dir 'settings.json')
    Compress-Archive -Path (Join-Path $dir 'settings.json') -DestinationPath $dest -Force
    Write-Host "Backup created: $dest"
} finally {
    Remove-Item -Recurse -Force $dir -ErrorAction SilentlyContinue
}
