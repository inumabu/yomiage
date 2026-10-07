@echo off
setlocal
cd /d "%~dp0"
call "%~dp0setup-env.cmd"
if errorlevel 1 exit /b %errorlevel%
docker compose -f compose.yml up -d --build
if errorlevel 1 exit /b %errorlevel%
docker compose -f compose.yml ps
