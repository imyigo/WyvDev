@echo off
setlocal
echo ================================================================
echo  AI Toolkit - Global kurulum baslatiliyor (PowerShell)
echo  Bu makinede kurulu IDE'ler otomatik tespit edilir, workspace'e
echo  hicbir sey yazilmaz - sadece gercek IDE global klasorlerine.
echo ================================================================
echo.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0install-skills.ps1"
echo.
pause
