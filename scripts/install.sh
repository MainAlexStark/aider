#!/usr/bin/env bash

set -e

REPO="https://github.com/YOUR_USERNAME/aider.git"
INSTALL_DIR="/usr/local/bin"
BINARY="aider"

echo "==> Installing $BINARY"

if ! command -v go >/dev/null 2>&1; then
    echo "Error: Go is not installed."
    exit 1
fi

TMP_DIR="$(mktemp -d)"

cleanup() {
    rm -rf "$TMP_DIR"
}

trap cleanup EXIT

echo "==> Cloning repository"

git clone "$REPO" "$TMP_DIR/aider"

cd "$TMP_DIR/aider"

echo "==> Building"

go build -o "$BINARY" ./cmd/agent

echo "==> Installing to $INSTALL_DIR"

sudo install \
    -m 755 \
    "$BINARY" \
    "$INSTALL_DIR/$BINARY"

echo
echo "Installation complete."
echo
echo "Try:"
echo "  aider explain \"connection refused\""
echo "  aider run echo hello"