#!/bin/sh
# download-binary.sh — Download the harness platform binary from GitHub releases.
# Called automatically by the Setup hook on first install (no binary required to run this).
#
# Usage:
#   bash skills/harness-setup/scripts/download-binary.sh
#   CLAUDE_PLUGIN_ROOT=/path/to/plugin bash skills/harness-setup/scripts/download-binary.sh

set -e

REPO="tim-hub/powerball-harness"

# Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
esac

BINARY_NAME="harness-${OS}-${ARCH}"

# Determine install directory
INSTALL_DIR="${CLAUDE_PLUGIN_ROOT:+${CLAUDE_PLUGIN_ROOT}/bin}"
if [ -z "$INSTALL_DIR" ]; then
  # Fall back to directory containing this script's parent (skills/../bin)
  SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
  INSTALL_DIR="${SCRIPT_DIR}/../../bin"
fi
INSTALL_DIR="$(cd "$INSTALL_DIR" 2>/dev/null && pwd || echo "$INSTALL_DIR")"

TARGET="${INSTALL_DIR}/${BINARY_NAME}"

# Already installed
if [ -x "$TARGET" ]; then
  echo "[harness] binary already installed: $TARGET"
  exit 0
fi

# Fetch latest release tag
echo "[harness] detecting latest release..."
LATEST_TAG=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | head -1 | cut -d'"' -f4)

if [ -z "$LATEST_TAG" ]; then
  echo "[harness] warning: could not fetch latest release tag, skipping binary download" >&2
  exit 0
fi

URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BINARY_NAME}"

echo "[harness] downloading ${BINARY_NAME} ${LATEST_TAG}..."
if curl -fsSL "$URL" -o "$TARGET"; then
  chmod +x "$TARGET"
  echo "[harness] installed: $TARGET"
  # Clear the "binary missing" warning flag so it won't show again
  rm -f "${HOME}/.claude/harness-binary-missing.warned"
else
  echo "[harness] warning: download failed (${URL}), skipping" >&2
  # Exit 0 so the hook doesn't block the user
  exit 0
fi
