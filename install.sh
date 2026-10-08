#!/usr/bin/env bash
# ==============================================================================
# Vibe-Fi v2 (Go Edition) Universal Installer
# Fast, clean, modern music player for terminal
# Supports instant pre-built binary installation or source build
# ==============================================================================

set -e

REPO="Swadesh-c0de/vibe-fi-go"

# ANSI Color Codes
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m' # No Color

echo -e "${CYAN}======================================${NC}"
echo -e "${CYAN}${BOLD}     Vibe-Fi v2 (Go Edition)${NC}"
echo -e "${CYAN}======================================${NC}"
echo ""

# Parse Command-Line Options
FORCE_BUILD=false
GLOBAL_INSTALL=false
UNINSTALL=false
VERSION="latest"
NO_DEPS=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --build|-b)
            FORCE_BUILD=true
            shift
            ;;
        --global|--system|-g)
            GLOBAL_INSTALL=true
            shift
            ;;
        --uninstall|-u)
            UNINSTALL=true
            shift
            ;;
        --version|-v)
            VERSION="$2"
            shift 2
            ;;
        --no-deps)
            NO_DEPS=true
            shift
            ;;
        --help|-h)
            echo "Usage: ./install.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --global, -g    Install system-wide to /usr/local/bin"
            echo "  --build, -b     Force building from source"
            echo "  --version, -v   Specify release tag to install (e.g. v2.0.0)"
            echo "  --no-deps       Skip automatic runtime dependency checks"
            echo "  --uninstall, -u Remove vibe binary"
            echo "  --help, -h      Show this help message"
            exit 0
            ;;
        *)
            shift
            ;;
    esac
done

# Detect Operating System & CPU Architecture
RAW_OS="$(uname -s)"
RAW_ARCH="$(uname -m)"

case "$RAW_OS" in
    Linux)
        OS="linux"
        ;;
    Darwin)
        OS="darwin"
        ;;
    *)
        echo -e "${RED}Unsupported Operating System: ${RAW_OS}${NC}"
        exit 1
        ;;
esac

case "$RAW_ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported Architecture: ${RAW_ARCH}${NC}"
        exit 1
        ;;
esac

echo -e "  [✔] Platform: ${RAW_OS} (${ARCH})"

# Determine Installation Directory
if [ "$GLOBAL_INSTALL" = true ] || [ "$EUID" -eq 0 ]; then
    TARGET_DIR="/usr/local/bin"
else
    TARGET_DIR="${HOME}/.local/bin"
fi

if [ "$UNINSTALL" = true ]; then
    echo -e "${YELLOW}Uninstalling Vibe-Fi...${NC}"
    rm -f "${TARGET_DIR}/vibe"
    rm -f "${HOME}/.local/bin/vibe" 2>/dev/null || true
    echo -e "${GREEN}[✔] Removed vibe executable.${NC}"
    echo ""
    echo -e "To purge user config, playlists, and cache, run:"
    echo -e "  ${CYAN}rm -rf ~/.vibe-fi${NC}"
    exit 0
fi

mkdir -p "${TARGET_DIR}"

TMP_DIR="$(mktemp -d)"
cleanup() {
    rm -rf "${TMP_DIR}"
}
trap cleanup EXIT

INSTALL_SUCCESS=false

# Method 1: Pre-built Binary Download (Fast, zero dependencies needed)
if [ "$FORCE_BUILD" = false ]; then
    if [ "$VERSION" = "latest" ]; then
        DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/vibe-${OS}-${ARCH}.tar.gz"
    else
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/vibe-${OS}-${ARCH}.tar.gz"
    fi

    echo -e "${BOLD}:: Checking for pre-built binary (${OS}-${ARCH})...${NC}"

    DOWNLOAD_CMD=""
    if command -v curl >/dev/null 2>&1; then
        DOWNLOAD_CMD="curl -fsSL"
    elif command -v wget >/dev/null 2>&1; then
        DOWNLOAD_CMD="wget -qO-"
    fi

    if [ -n "$DOWNLOAD_CMD" ]; then
        if $DOWNLOAD_CMD "${DOWNLOAD_URL}" 2>/dev/null | tar -xz -C "${TMP_DIR}" 2>/dev/null; then
            if [ -f "${TMP_DIR}/vibe" ]; then
                echo -e "${GREEN}Downloaded pre-built binary successfully!${NC}"

                if [ "$GLOBAL_INSTALL" = true ] && [ "$EUID" -ne 0 ]; then
                    sudo install -m 755 "${TMP_DIR}/vibe" "${TARGET_DIR}/vibe"
                else
                    install -m 755 "${TMP_DIR}/vibe" "${TARGET_DIR}/vibe"
                fi

                INSTALL_SUCCESS=true
            fi
        fi
    fi

    if [ "$INSTALL_SUCCESS" = false ]; then
        echo -e "  ${YELLOW}[i] Pre-built release not found or network unavailable.${NC}"
    fi
fi

# Method 2: Build from Source (Fallback or if --build is specified)
if [ "$INSTALL_SUCCESS" = false ]; then
    if [ ! -f "cmd/vibe/main.go" ]; then
        echo -e "${RED}Error: Cannot build from source because source files are not present.${NC}"
        echo "Please clone the repository first:"
        echo "  git clone https://github.com/${REPO}.git"
        echo "  cd vibe-fi-go && ./install.sh"
        exit 1
    fi

    echo -e "${BOLD}:: Building from source...${NC}"

    if ! command -v go >/dev/null 2>&1; then
        echo -e "${RED}Error: Go compiler is not installed.${NC}"
        echo "Please install Go 1.20+ from https://go.dev/dl/"
        exit 1
    fi

    # On macOS, export Homebrew pkg-config paths
    if [ "$OS" = "darwin" ]; then
        if [ -d "/opt/homebrew/lib/pkgconfig" ]; then
            export PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
        fi
        if [ -d "/usr/local/lib/pkgconfig" ]; then
            export PKG_CONFIG_PATH="/usr/local/lib/pkgconfig${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
        fi
    fi

    mkdir -p build
    if go build -ldflags="-s -w" -o build/vibe ./cmd/vibe; then
        cp build/vibe ./vibe
        echo -e "${GREEN}Build successful:${NC} ./vibe"

        if [ "$GLOBAL_INSTALL" = true ] && [ "$EUID" -ne 0 ]; then
            sudo install -m 755 build/vibe "${TARGET_DIR}/vibe"
        else
            install -m 755 build/vibe "${TARGET_DIR}/vibe"
        fi
        INSTALL_SUCCESS=true
    else
        echo -e "${RED}Build from source failed.${NC}"
        exit 1
    fi
fi

# Post-Install Verification
echo ""
echo -e "${GREEN}Installed:${NC} ${TARGET_DIR}/vibe"

# Verify Runtime Dependencies (libmpv)
if [ "$NO_DEPS" = false ] && [ -f "${TARGET_DIR}/vibe" ]; then
    if "${TARGET_DIR}/vibe" --version 2>&1 | grep -qi "libmpv"; then
        echo ""
        echo -e "${YELLOW}:: Notice: Runtime dependency 'libmpv' is missing.${NC}"

        SUDO=""
        if [ "$EUID" -ne 0 ] && command -v sudo >/dev/null 2>&1; then
            SUDO="sudo"
        fi

        AUTO_INSTALLED=false
        if [ "$EUID" -eq 0 ] || [ -n "$SUDO" ]; then
            if command -v apt-get >/dev/null 2>&1; then
                echo -e "${CYAN}:: Installing libmpv via apt...${NC}"
                if $SUDO apt-get update -qq && ($SUDO apt-get install -y --no-install-recommends libmpv2 2>/dev/null || $SUDO apt-get install -y --no-install-recommends mpv 2>/dev/null); then
                    AUTO_INSTALLED=true
                fi
            elif command -v pacman >/dev/null 2>&1; then
                echo -e "${CYAN}:: Installing mpv via pacman...${NC}"
                if $SUDO pacman -Sy --noconfirm mpv; then
                    AUTO_INSTALLED=true
                fi
            elif command -v dnf >/dev/null 2>&1; then
                echo -e "${CYAN}:: Installing mpv-libs via dnf...${NC}"
                if $SUDO dnf install -y mpv-libs 2>/dev/null || $SUDO dnf install -y mpv 2>/dev/null; then
                    AUTO_INSTALLED=true
                fi
            elif command -v zypper >/dev/null 2>&1; then
                echo -e "${CYAN}:: Installing libmpv2 via zypper...${NC}"
                if $SUDO zypper install -y libmpv2 2>/dev/null || $SUDO zypper install -y mpv 2>/dev/null; then
                    AUTO_INSTALLED=true
                fi
            elif command -v apk >/dev/null 2>&1; then
                echo -e "${CYAN}:: Installing mpv-libs via apk...${NC}"
                if $SUDO apk add mpv-libs 2>/dev/null; then
                    AUTO_INSTALLED=true
                fi
            elif [ "$OS" = "darwin" ] && command -v brew >/dev/null 2>&1; then
                echo -e "${CYAN}:: Installing mpv via Homebrew...${NC}"
                if brew install mpv; then
                    AUTO_INSTALLED=true
                fi
            fi
        fi

        if [ "$AUTO_INSTALLED" = true ]; then
            echo -e "${GREEN}[✔] Successfully installed libmpv.${NC}"
        else
            echo -e "${YELLOW}[!] Action Required: Install libmpv using your package manager:${NC}"
            if command -v apt-get >/dev/null 2>&1; then
                echo -e "    ${BOLD}apt update && apt install -y libmpv2${NC} (or mpv)"
            elif command -v pacman >/dev/null 2>&1; then
                echo -e "    ${BOLD}sudo pacman -S mpv${NC}"
            elif command -v dnf >/dev/null 2>&1; then
                echo -e "    ${BOLD}sudo dnf install mpv-libs${NC}"
            elif command -v zypper >/dev/null 2>&1; then
                echo -e "    ${BOLD}sudo zypper install libmpv2${NC}"
            elif command -v apk >/dev/null 2>&1; then
                echo -e "    ${BOLD}apk add mpv-libs${NC}"
            elif [ "$OS" = "darwin" ]; then
                echo -e "    ${BOLD}brew install mpv${NC}"
            else
                echo -e "    ${BOLD}Please install libmpv / mpv for your system.${NC}"
            fi
            echo ""
        fi
    fi
fi

if [ "$TARGET_DIR" = "${HOME}/.local/bin" ]; then
    if [[ ":$PATH:" != *":${TARGET_DIR}:"* ]]; then
        echo ""
        echo -e "${YELLOW}Note:${NC} ${TARGET_DIR} is not in your \$PATH."
        echo "Add the following line to your ~/.bashrc or ~/.zshrc:"
        echo -e "  ${BOLD}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
    fi

    echo ""
    echo -e "To install system-wide to /usr/local/bin, run:"
    echo -e "  ${BOLD}./install.sh --global${NC}"
fi

echo ""
echo -e "${GREEN}${BOLD}Done!${NC} Run '${CYAN}vibe${NC}' to start listening."
