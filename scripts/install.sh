#!/usr/bin/env bash
set -e

APP_NAME="cryptoterm"
INSTALL_DIR="/usr/local/bin"
ALT_INSTALL_DIR="$HOME/.local/bin"

echo "==============================================="
echo "  Installing ${APP_NAME}..."
echo "==============================================="

# Determine target directory
if [ -n "$PREFIX" ] && [ -d "$PREFIX/bin" ]; then
    # Termux Android environment
    TARGET_DIR="$PREFIX/bin"
    USE_SUDO=0
elif [ -w "$INSTALL_DIR" ]; then
    TARGET_DIR="$INSTALL_DIR"
    USE_SUDO=0
elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
    TARGET_DIR="$INSTALL_DIR"
    USE_SUDO=1
else
    echo "Note: Installing to ${ALT_INSTALL_DIR}."
    mkdir -p "${ALT_INSTALL_DIR}"
    TARGET_DIR="${ALT_INSTALL_DIR}"
    USE_SUDO=0
fi

# Build binary
echo "Building ${APP_NAME} binary with Go..."
go build -ldflags="-s -w" -o "${APP_NAME}" ./cmd/cryptoterm

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
