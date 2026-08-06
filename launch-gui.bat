@echo off
setlocal
echo ================================================================
echo  WyvDev Hub — Canli Sunucu Baslatiliyor (Windows)
echo  wyvdev.exe kendi HTTP sunucusunu acar ve tarayiciyi baslatir
echo ================================================================
echo.
start "WyvDev Hub" "%~dp0wyvdev.exe"
exit
