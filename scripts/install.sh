#!/usr/bin/env bash
set -e

APP_NAME="binance-terminal"
INSTALL_DIR="/usr/local/bin"
ALT_INSTALL_DIR="$HOME/.local/bin"

echo "==============================================="
echo "  Installing ${APP_NAME}..."
echo "==============================================="

# Determine target directory
TARGET_DIR="$INSTALL_DIR"
if [ ! -w "$INSTALL_DIR" ]; then
    if sudo -n true 2>/dev/null; then
        USE_SUDO=1
    else
        echo "Note: No write permission for ${INSTALL_DIR}. Installing to ${ALT_INSTALL_DIR} instead."
        mkdir -p "${ALT_INSTALL_DIR}"
        TARGET_DIR="${ALT_INSTALL_DIR}"
        USE_SUDO=0
    fi
fi

# Build binary
echo "Building ${APP_NAME} binary with Go..."
go build -ldflags="-s -w" -o "${APP_NAME}" cmd/binance-terminal/main.go

# Install
echo "Installing to ${TARGET_DIR}/${APP_NAME}..."
if [ "${USE_SUDO}" = "1" ]; then
    sudo cp "${APP_NAME}" "${TARGET_DIR}/${APP_NAME}"
    sudo chmod +x "${TARGET_DIR}/${APP_NAME}"
else
    cp "${APP_NAME}" "${TARGET_DIR}/${APP_NAME}"
    chmod +x "${TARGET_DIR}/${APP_NAME}"
fi

echo "==============================================="
echo "✅ Installation complete!"
echo ""
echo "Run '${APP_NAME}' in your terminal to get started!"
if [ "${TARGET_DIR}" = "${ALT_INSTALL_DIR}" ]; then
    echo "Make sure ${ALT_INSTALL_DIR} is in your PATH."
fi
echo "==============================================="
