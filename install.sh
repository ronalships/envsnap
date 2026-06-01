#!/bin/sh
set -e

# envsnap installer
# Usage: curl -fsSL https://raw.githubusercontent.com/ronalships/envsnap/main/install.sh | sh

REPO="ronalships/envsnap"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"

    case "$OS" in
        darwin) OS="darwin" ;;
        linux) OS="linux" ;;
        *) echo "Unsupported OS: $OS"; exit 1 ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    echo "${OS}_${ARCH}"
}

# Get latest release tag from GitHub
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
        grep '"tag_name":' |
        sed -E 's/.*"([^"]+)".*/\1/'
}

main() {
    echo "Installing envsnap..."

    PLATFORM="$(detect_platform)"
    VERSION="$(get_latest_version)"

    if [ -z "$VERSION" ]; then
        echo ""
        echo "No releases available yet."
        echo ""
        echo "envsnap is pre-release. Install via Go instead:"
        echo ""
        echo "  go install github.com/${REPO}/cmd/envsnap@latest"
        echo ""
        exit 1
    fi

    FILENAME="envsnap_${VERSION#v}_${PLATFORM}.tar.gz"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    echo "Downloading ${FILENAME}..."

    TMPDIR="$(mktemp -d)"
    trap "rm -rf ${TMPDIR}" EXIT

    curl -fsSL "$URL" -o "${TMPDIR}/envsnap.tar.gz"
    tar -xzf "${TMPDIR}/envsnap.tar.gz" -C "${TMPDIR}"

    echo "Installing to ${INSTALL_DIR}/envsnap..."

    if [ -w "$INSTALL_DIR" ]; then
        mv "${TMPDIR}/envsnap" "${INSTALL_DIR}/envsnap"
    else
        sudo mv "${TMPDIR}/envsnap" "${INSTALL_DIR}/envsnap"
    fi

    chmod +x "${INSTALL_DIR}/envsnap"

    echo ""
    echo "envsnap ${VERSION} installed successfully!"
    echo ""
    echo "Run 'envsnap --help' to get started."
}

main
