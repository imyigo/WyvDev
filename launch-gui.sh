#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$DIR"

# Mimariye göre doğru binary seç
ARCH=$(uname -m)
if [[ "$ARCH" == "aarch64" || "$ARCH" == "arm64" ]]; then
    BINARY="./wyvdev-linux-arm64"
else
    BINARY="./wyvdev-linux-amd64"
fi

if [[ ! -f "$BINARY" ]]; then
    echo "================================================================"
    echo " WyvDev Hub — HATA: $BINARY bulunamadı!"
    echo " Lütfen https://github.com/imyigo/WyvDev adresinden"
    echo " doğru binary'yi indirin veya 'go build -o $BINARY main.go' çalıştırın."
    echo "================================================================"
    read -p "Çıkmak için Enter tuşuna basın..."
    exit 1
fi

chmod +x "$BINARY" 2>/dev/null
FAIL_COUNT=0

while true; do
    echo "================================================================"
    echo " WyvDev Hub — Starting... (Watchdog Active | Arch: $ARCH)"
    echo "================================================================"
    echo ""

    "$BINARY"
    EXIT_CODE=$?

    if [[ $EXIT_CODE -eq 126 || $EXIT_CODE -eq 127 ]]; then
        FAIL_COUNT=$((FAIL_COUNT + 1))
        echo ""
        echo "[WyvDev] ❌ Binary çalıştırılamadı (kod: $EXIT_CODE, mimari: $ARCH)"
        if [[ $FAIL_COUNT -ge 3 ]]; then
            echo "[WyvDev] 3 başarısız denemeden sonra durduruluyor."
            echo "         Kaynak koddan derlemek için: go build -o $BINARY main.go"
            read -p "Çıkmak için Enter tuşuna basın..."
            exit 1
        fi
    else
        FAIL_COUNT=0
    fi

    echo ""
    echo "[WyvDev] Backend durdu. 3 saniye sonra yeniden baslatiliyor..."
    sleep 3
done
