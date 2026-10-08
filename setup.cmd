@echo off
setlocal
cd /d "%~dp0"
node scripts\setup.mjs %*
set "EXIT_CODE=%ERRORLEVEL%"
endlocal & exit /b %EXIT_CODE%
