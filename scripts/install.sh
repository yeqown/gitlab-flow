#!/bin/bash

set -euo pipefail

DEFAULT_INSTALL_DIR="/usr/local/bin"
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOCAL_FLOW3="$SCRIPT_DIR/flow3"
FLOW3_URL="${FLOW3_URL:-https://raw.githubusercontent.com/yeqown/gitlab-flow/main/scripts/flow3}"
FALLBACK_DIR="$HOME/.local/bin"
TMP_FILE=""

cleanup() {
    if [ -n "$TMP_FILE" ] && [ -f "$TMP_FILE" ]; then
        rm -f "$TMP_FILE"
    fi
}

md5_of_file() {
    local file="$1"

    if command -v md5 >/dev/null; then
        md5 -q "$file"
    elif command -v md5sum >/dev/null; then
        md5sum "$file" | awk '{print $1}'
    else
        echo "❌ neither md5 nor md5sum is available"
        exit 1
    fi
}

trap cleanup EXIT

echo "Installing flow3..."

if ! command -v curl >/dev/null; then
    echo "❌ curl not found. Please install curl first."
    exit 1
fi

TMP_FILE="$(mktemp)"
echo "Downloading flow3 from $FLOW3_URL"
curl -fsSL "$FLOW3_URL" -o "$TMP_FILE"
SOURCE_FLOW3="$TMP_FILE"

TARGET_DIR="$INSTALL_DIR"
USE_SUDO="false"

if [ -d "$TARGET_DIR" ] && [ -w "$TARGET_DIR" ]; then
    :
elif [ ! -d "$TARGET_DIR" ] && [ -w "$(dirname "$TARGET_DIR")" ]; then
    mkdir -p "$TARGET_DIR"
elif command -v sudo >/dev/null; then
    USE_SUDO="true"
else
    TARGET_DIR="$FALLBACK_DIR"
    mkdir -p "$TARGET_DIR"
    echo "No write permission for $INSTALL_DIR and sudo is unavailable. Falling back to $TARGET_DIR"
fi

TARGET_PATH="$TARGET_DIR/flow3"
INSTALLED_PATH=""

if command -v flow3 >/dev/null; then
    INSTALLED_PATH="$(command -v flow3)"
fi

if [ -n "$INSTALLED_PATH" ] && [ -f "$INSTALLED_PATH" ]; then
    INSTALLED_MD5="$(md5_of_file "$INSTALLED_PATH")"
    REMOTE_MD5="$(md5_of_file "$SOURCE_FLOW3")"

    if [ "$INSTALLED_MD5" = "$REMOTE_MD5" ]; then
        echo "✅ flow3 is already up to date ($INSTALLED_PATH)"
        exit 0
    fi

    if [ -z "${INSTALL_DIR:-}" ] || [ "$INSTALL_DIR" = "$DEFAULT_INSTALL_DIR" ]; then
        TARGET_PATH="$INSTALLED_PATH"
        TARGET_DIR="$(dirname "$TARGET_PATH")"
        if [ ! -w "$TARGET_DIR" ] && command -v sudo >/dev/null; then
            USE_SUDO="true"
        fi
    fi

    echo "Found installed flow3: $INSTALLED_PATH"
    echo "Local md5 : $INSTALLED_MD5"
    echo "Remote md5: $REMOTE_MD5"
    read -r -p "MD5 differs, do you want to upgrade? [y/N] " confirm_upgrade
    case "$confirm_upgrade" in
        y|Y|yes|YES)
            ;;
        *)
            echo "Upgrade cancelled."
            exit 0
            ;;
    esac
else
    if [ -f "$LOCAL_FLOW3" ]; then
        echo "No installed flow3 found in PATH. Proceeding with fresh install."
    fi
fi

if [ "$USE_SUDO" = "true" ]; then
    echo "Need sudo permission to install to $TARGET_DIR"
    sudo mkdir -p "$TARGET_DIR"
    sudo install -m 755 "$SOURCE_FLOW3" "$TARGET_PATH"
else
    install -m 755 "$SOURCE_FLOW3" "$TARGET_PATH"
fi

echo "✅ flow3 installed to $TARGET_PATH"

if [ "$TARGET_DIR" = "$FALLBACK_DIR" ] && [[ ":$PATH:" != *":$FALLBACK_DIR:"* ]]; then
    echo "⚠️  Add $FALLBACK_DIR to PATH:"
    echo "   export PATH=\"$FALLBACK_DIR:\$PATH\""
fi

echo "Checking dependencies..."

if ! command -v glab >/dev/null; then
    echo "❌ glab not found. Please install glab first:"
    echo "   brew install glab"
    echo "   or visit: https://gitlab.com/gitlab-org/cli"
    exit 1
fi

if ! command -v git >/dev/null; then
    echo "❌ git not found. Please install git first."
    exit 1
fi

echo "✅ All dependencies satisfied"
echo ""
echo "Usage: flow3 help"
echo "Initialize: flow3 init"
