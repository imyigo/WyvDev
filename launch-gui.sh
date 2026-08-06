#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR"

echo "================================================================"
echo " WyvDev Hub — Starting Live Server (Linux)"
echo "================================================================"
echo ""

chmod +x ./wyvdev-linux-amd64 2>/dev/null
./wyvdev-linux-amd64
