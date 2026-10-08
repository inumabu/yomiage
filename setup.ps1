[CmdletBinding()]
param(
    [ValidateSet('docker', 'build', 'verify')]
    [string]$Mode = 'docker',
    [switch]$SkipEnv,
    [switch]$ForceEnv,
    [switch]$NoBuild,
    [switch]$DryRun,
    [switch]$Doctor,
    [switch]$Repair
)

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    if (Get-Command winget -ErrorAction SilentlyContinue) {
        Write-Host 'Node.jsが見つからないため、wingetで導入します。' -ForegroundColor Yellow
        winget install --id OpenJS.NodeJS.LTS --exact --accept-source-agreements --accept-package-agreements
        $env:Path = "$env:ProgramFiles\nodejs;$env:Path"
    } else {
        throw 'Node.jsが見つかりません。Node.js 22以上またはWindows App Installer（winget）を導入して再実行してください。'
    }
}
if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    throw 'Node.jsを導入しましたがnodeコマンドが見つかりません。PowerShellを再起動して再実行してください。'
}

$argsList = @('--mode', $Mode)
if ($SkipEnv) { $argsList += '--skip-env' }
if ($ForceEnv) { $argsList += '--force-env' }
if ($NoBuild) { $argsList += '--no-build' }
if ($DryRun) { $argsList += '--dry-run' }
if ($Doctor) { $argsList += '--doctor' }
if ($Repair) { $argsList += '--repair' }

& node (Join-Path $PSScriptRoot 'scripts/setup.mjs') @argsList
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
