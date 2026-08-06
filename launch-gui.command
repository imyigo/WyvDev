#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR"

while true; do
  echo "================================================================"
  echo " WyvDev Hub — Starting... (Auto-Restart Watchdog Active)"
  echo "================================================================"
  echo ""

  if [[ "$(uname -m)" == "arm64" ]]; then
    chmod +x ./wyvdev-darwin-arm64 2>/dev/null
    ./wyvdev-darwin-arm64
  else
    chmod +x ./wyvdev-darwin-amd64 2>/dev/null
    ./wyvdev-darwin-amd64
  fi

  echo ""
  echo "[WyvDev] Backend durdu. 3 saniye sonra yeniden baslatiliyor..."
  sleep 3
done
