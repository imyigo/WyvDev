@echo off
title WyvDev Hub — Local Developer PaaS
cd /d "%~dp0"

:start
echo ================================================================
echo  WyvDev Hub — Starting... (Auto-Restart Watchdog Active)
echo ================================================================
echo.
wyvdev.exe
echo.
echo [WyvDev] Backend durdu. 3 saniye sonra yeniden baslatiliyor...
timeout /t 3 /nobreak >nul
goto start
