#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR"

echo "================================================================"
echo " WyvDev Hub — Starting Live Server (macOS)"
echo "================================================================"
echo ""

if [[ "$(uname -m)" == "arm64" ]]; then
    chmod +x ./wyvdev-darwin-arm64 2>/dev/null
    ./wyvdev-darwin-arm64
else
    chmod +x ./wyvdev-darwin-amd64 2>/dev/null
    ./wyvdev-darwin-amd64
fi
