#!/bin/sh
set -e

# Ghost Installation Script
# This script downloads the latest release of Ghost and installs it to /usr/local/bin

REPO="swadhinbiswas/ghost"
BIN_DIR="/usr/local/bin"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux*) OS="Linux" ;;
    darwin*) OS="Darwin" ;;
    msys*|cygwin*|mingw*) OS="Windows" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    i386|i686) ARCH="i386" ;;
    *) echo "Unsupported Architecture: $ARCH"; exit 1 ;;
esac

echo "=> Detecting latest version of Ghost..."
if command -v curl >/dev/null 2>&1; then
    LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
elif command -v wget >/dev/null 2>&1; then
    LATEST_TAG=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
else
    echo "Error: curl or wget is required to download the binary."
    exit 1
fi

if [ -z "$LATEST_TAG" ]; then
    echo "Failed to fetch the latest version. Please check your internet connection or GitHub API limits."
    exit 1
fi

echo "=> Latest version is $LATEST_TAG"

# Determine the file extension based on OS
EXT="tar.gz"
if [ "$OS" = "Windows" ]; then
    EXT="zip"
fi

FILE_NAME="ghost_${OS}_${ARCH}.${EXT}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/${LATEST_TAG}/${FILE_NAME}"

TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo "=> Downloading $DOWNLOAD_URL ..."
if command -v curl >/dev/null 2>&1; then
    curl -L -o "$FILE_NAME" "$DOWNLOAD_URL"
else
    wget -O "$FILE_NAME" "$DOWNLOAD_URL"
fi

echo "=> Extracting archive..."
if [ "$EXT" = "tar.gz" ]; then
    tar -xzf "$FILE_NAME"
else
    unzip -q "$FILE_NAME"
fi

echo "=> Installing to $BIN_DIR/ghost (Requires sudo)"
if [ "$OS" = "Windows" ]; then
    # For Windows, just copy to the current directory or suggest path.
    mkdir -p "$HOME/bin"
    cp ghost.exe "$HOME/bin/"
    echo "Installed ghost.exe to $HOME/bin. Please ensure $HOME/bin is in your PATH."
else
    sudo mv ghost "$BIN_DIR/"
    sudo chmod +x "$BIN_DIR/ghost"
fi

cd - >/dev/null
rm -rf "$TMP_DIR"

echo "=> Ghost installation completed successfully!"
echo "Run 'ghost --help' to get started."
