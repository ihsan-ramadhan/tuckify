#!/bin/sh
set -e

REPO="ihsan-ramadhan/tuckify"
VERSION="v0.2.1"

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

GUI=false
for arg in "$@"; do
    case "$arg" in
        --gui|-g) GUI=true ;;
        *) ;; # ignore unknown arguments
    esac
done

INSTALL_DIR="${HOME}/.local/bin"
mkdir -p "${INSTALL_DIR}"

if [ "$GUI" = true ]; then
    if [ "$OS" = "darwin" ]; then
        URL="https://github.com/${REPO}/releases/download/${VERSION}/tuckify-gui-mac-universal.zip"
        TMP_ZIP=$(mktemp)
        echo "Downloading tuckify-gui ${VERSION} for macOS..."
        curl -fsSL "${URL}" -o "${TMP_ZIP}"
        echo "Installing to /Applications..."
        unzip -q -o "${TMP_ZIP}" -d "/Applications"
        rm -f "${TMP_ZIP}"
        echo "tuckify-gui successfully installed to /Applications/tuckify-gui.app"
    else
        if [ "$ARCH" != "amd64" ]; then
            echo "GUI only supports amd64 architecture on Linux currently."
            exit 1
        fi
        URL="https://github.com/${REPO}/releases/download/${VERSION}/tuckify-gui-linux-amd64"
        echo "Downloading tuckify-gui ${VERSION} for Linux (amd64)..."
        curl -fsSL "${URL}" -o "${INSTALL_DIR}/tuckify-gui"
        chmod +x "${INSTALL_DIR}/tuckify-gui"
        echo "tuckify-gui successfully installed to ${INSTALL_DIR}/tuckify-gui"
    fi
else
    BINARY="tuckify-${OS}-${ARCH}"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"
    echo "Downloading tuckify ${VERSION} for ${OS}/${ARCH}..."
    curl -fsSL "${URL}" -o "${INSTALL_DIR}/tuckify"
    chmod +x "${INSTALL_DIR}/tuckify"
    echo "tuckify successfully installed to ${INSTALL_DIR}/tuckify"
fi

case :$PATH: in
    *:"${INSTALL_DIR}":*) ;;
    *) echo "Please add ${INSTALL_DIR} to your PATH (e.g. in ~/.bashrc or ~/.zshrc)" ;;
esac
