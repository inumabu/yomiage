@echo off
setlocal
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0repair-volume.ps1"
if errorlevel 1 exit /b %errorlevel%
