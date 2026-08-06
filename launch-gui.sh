#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR"

while true; do
  echo "================================================================"
  echo " WyvDev Hub — Starting... (Auto-Restart Watchdog Active)"
  echo "================================================================"
  echo ""

  chmod +x ./wyvdev-linux-amd64 2>/dev/null
  ./wyvdev-linux-amd64

  echo ""
  echo "[WyvDev] Backend durdu. 3 saniye sonra yeniden baslatiliyor..."
  sleep 3
done
