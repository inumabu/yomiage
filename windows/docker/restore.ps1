param(
    [Parameter(Mandatory=$true)]
    [string]$BackupZip
)

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if (-not (Test-Path $BackupZip)) { throw "バックアップが見つかりません: $BackupZip" }
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$tmp = Join-Path $PSScriptRoot "backup\restore-$stamp"
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
try {
    Expand-Archive -Path $BackupZip -DestinationPath $tmp -Force
    $settings = Join-Path $tmp 'settings.json'
    if (-not (Test-Path $settings)) { throw 'バックアップアーカイブ内にsettings.jsonがありません' }
    docker compose -f compose.yml up -d
    docker compose -f compose.yml cp $settings yomiage:/tmp/yomiage-settings-restore.json
    docker compose -f compose.yml exec -T -u 0 yomiage sh -c "install -o yomiage-keiryou -g yomiage-keiryou -m 0640 /tmp/yomiage-settings-restore.json /var/lib/yomiage-keiryou/settings.json && rm -f /tmp/yomiage-settings-restore.json"
    docker compose -f compose.yml restart yomiage
    Write-Host "復元しました: $BackupZip"
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
