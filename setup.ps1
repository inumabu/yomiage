[CmdletBinding()]
param(
    [ValidateSet('docker', 'build', 'verify')]
    [string]$Mode = 'docker',
    [switch]$SkipEnv,
    [switch]$ForceEnv,
    [switch]$NoBuild,
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$argsList = @('--mode', $Mode)
if ($SkipEnv) { $argsList += '--skip-env' }
if ($ForceEnv) { $argsList += '--force-env' }
if ($NoBuild) { $argsList += '--no-build' }
if ($DryRun) { $argsList += '--dry-run' }

& node (Join-Path $PSScriptRoot 'scripts/setup.mjs') @argsList
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
