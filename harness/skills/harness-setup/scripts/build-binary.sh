#!/bin/sh
# build-binary.sh — Build the harness binary from source for the current platform.
# Replaces download-binary.sh to avoid network-dependent failures.
#
# Usage:
#   bash skills/harness-setup/scripts/build-binary.sh
#   CLAUDE_PLUGIN_ROOT=/path/to/plugin bash skills/harness-setup/scripts/build-binary.sh

set -e

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
  SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
  INSTALL_DIR="${SCRIPT_DIR}/../../../bin"
fi
mkdir -p "$INSTALL_DIR"
INSTALL_DIR="$(cd "$INSTALL_DIR" && pwd)"

TARGET="${INSTALL_DIR}/${BINARY_NAME}"

# Already installed — skip rebuild
if [ -x "$TARGET" ]; then
  echo "[harness] binary already installed: $TARGET"
  exit 0
fi

# Locate Go source directory
GO_DIR="${CLAUDE_PLUGIN_ROOT:+${CLAUDE_PLUGIN_ROOT}/go}"
if [ -z "$GO_DIR" ] || [ ! -d "$GO_DIR" ]; then
  SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
  GO_DIR="${SCRIPT_DIR}/../../../go"
fi

if [ ! -d "$GO_DIR" ]; then
  echo "[harness] error: Go source directory not found at $GO_DIR" >&2
  exit 1
fi

GO_DIR="$(cd "$GO_DIR" && pwd)"

# Check Go is available
if ! command -v go >/dev/null 2>&1; then
  echo "[harness] error: 'go' is not installed or not in PATH" >&2
  echo "[harness] install Go from https://go.dev/dl/" >&2
  exit 1
fi

# Read version
VERSION="dev"
VERSION_FILE="${GO_DIR}/../VERSION"
if [ -f "$VERSION_FILE" ]; then
  VERSION="$(cat "$VERSION_FILE")"
fi

LDFLAGS="-s -w -X main.version=${VERSION}"

echo "[harness] building ${BINARY_NAME} v${VERSION} from source..."
(cd "$GO_DIR" && CGO_ENABLED=0 GOOS="$OS" GOARCH="$ARCH" go build -ldflags="$LDFLAGS" -o "$TARGET" ./cmd/harness/)
chmod +x "$TARGET"
echo "[harness] installed: $TARGET"

# Clear the "binary missing" warning flag
rm -f "${HOME}/.claude/harness-binary-missing.warned"
