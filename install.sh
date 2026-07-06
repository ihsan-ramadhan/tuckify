#!/bin/sh
set -e

REPO="ihsan-ramadhan/tuckify"
VERSION="v0.2.0"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

BINARY="tuckify-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"
INSTALL_DIR="${HOME}/.local/bin"

mkdir -p "${INSTALL_DIR}"
echo "Downloading tuckify ${VERSION} for ${OS}/${ARCH}..."
curl -fsSL "${URL}" -o "${INSTALL_DIR}/tuckify"

chmod +x "${INSTALL_DIR}/tuckify"
echo "tuckify successfully installed to ${INSTALL_DIR}/tuckify"

case :$PATH: in
    *:"${INSTALL_DIR}":*) ;;
    *) echo "Please add ${INSTALL_DIR} to your PATH (e.g. in ~/.bashrc or ~/.zshrc)" ;;
esac
