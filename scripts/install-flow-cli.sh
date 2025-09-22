#!/bin/bash

set -e

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing flow3..."

# 检查权限
if [ ! -w "$INSTALL_DIR" ]; then
    echo "Need sudo permission to install to $INSTALL_DIR"
    sudo cp "$SCRIPT_DIR/flow3" "$INSTALL_DIR/flow3"
    sudo chmod +x "$INSTALL_DIR/flow3"
else
    cp "$SCRIPT_DIR/flow3" "$INSTALL_DIR/flow3"
    chmod +x "$INSTALL_DIR/flow3"
fi

echo "✅ flow3 installed to $INSTALL_DIR/flow3"

# 检查依赖
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
