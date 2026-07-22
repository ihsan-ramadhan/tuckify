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

BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"

verify_checksum() {
    file_path="$1"
    file_name=$(basename "$file_path")

    checksums_url="${BASE_URL}/checksums.txt"
    tmp_checksums=$(mktemp)
    if curl -fsSL "${checksums_url}" -o "${tmp_checksums}"; then
        expected=$(grep "  ${file_name}$" "${tmp_checksums}" | awk '{print $1}')
        rm -f "${tmp_checksums}"
        if [ -n "$expected" ]; then
            actual=$(sha256sum "$file_path" | awk '{print $1}')
            if [ "$expected" != "$actual" ]; then
                echo "Checksum verification failed for ${file_name}"
                echo "  expected: ${expected}"
                echo "  actual:   ${actual}"
                rm -f "$file_path"
                exit 1
            fi
            echo "Checksum verified: ${file_name}"
        else
            echo "Warning: ${file_name} not found in checksums.txt, skipping verification"
        fi
    else
        echo "Warning: could not download checksums.txt, skipping verification"
    fi
}

if [ "$GUI" = true ]; then
    if [ "$OS" = "darwin" ]; then
        URL="${BASE_URL}/tuckify-gui-mac-universal.zip"
        TMP_ZIP=$(mktemp)
        echo "Downloading tuckify-gui ${VERSION} for macOS..."
        curl -fsSL "${URL}" -o "${TMP_ZIP}"
        verify_checksum "${TMP_ZIP}"
        echo "Installing to /Applications..."
        unzip -q -o "${TMP_ZIP}" -d "/Applications"
        rm -f "${TMP_ZIP}"
        echo "tuckify-gui successfully installed to /Applications/tuckify-gui.app"
    else
        if [ "$ARCH" != "amd64" ]; then
            echo "GUI only supports amd64 architecture on Linux currently."
            exit 1
        fi
        URL="${BASE_URL}/tuckify-gui-linux-amd64"
        echo "Downloading tuckify-gui ${VERSION} for Linux (amd64)..."
        curl -fsSL "${URL}" -o "${INSTALL_DIR}/tuckify-gui"
        verify_checksum "${INSTALL_DIR}/tuckify-gui"
        chmod +x "${INSTALL_DIR}/tuckify-gui"
        echo "tuckify-gui successfully installed to ${INSTALL_DIR}/tuckify-gui"
    fi
else
    BINARY="tuckify-${OS}-${ARCH}"
    URL="${BASE_URL}/${BINARY}"
    echo "Downloading tuckify ${VERSION} for ${OS}/${ARCH}..."
    curl -fsSL "${URL}" -o "${INSTALL_DIR}/tuckify"
    verify_checksum "${INSTALL_DIR}/tuckify"
    chmod +x "${INSTALL_DIR}/tuckify"
    echo "tuckify successfully installed to ${INSTALL_DIR}/tuckify"
fi

case :$PATH: in
    *:"${INSTALL_DIR}":*) ;;
    *) echo "Please add ${INSTALL_DIR} to your PATH (e.g. in ~/.bashrc or ~/.zshrc)" ;;
esac
