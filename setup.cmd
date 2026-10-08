@echo off
setlocal
cd /d "%~dp0"
where node >nul 2>&1
if errorlevel 1 (
  where winget >nul 2>&1
  if errorlevel 1 (
    echo Node.jsが見つかりません。Node.js 22以上またはwingetを導入して再実行してください。
    exit /b 1
  )
  echo Node.jsが見つからないため、wingetで導入します。
  winget install --id OpenJS.NodeJS.LTS --exact --accept-source-agreements --accept-package-agreements
  set "PATH=%ProgramFiles%\nodejs;%PATH%"
)
where node >nul 2>&1
if errorlevel 1 (
  echo Node.jsを導入しましたがnodeコマンドが見つかりません。コマンドプロンプトを再起動して再実行してください。
  exit /b 1
)
node scripts\setup.mjs %*
set "EXIT_CODE=%ERRORLEVEL%"
endlocal & exit /b %EXIT_CODE%
