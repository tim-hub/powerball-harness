#!/bin/sh
# build-binary.sh — Dev helper: rebuild the harness binary from Go source for current platform.
#
# This is a CONTRIBUTOR tool. Users get prebuilt binaries from harness/bin/ in the repo —
# no Go toolchain is needed for fresh installs.
#
# Usage (from repo root):
#   bash local-scripts/build-binary.sh
#   CLAUDE_PLUGIN_ROOT=/path/to/plugin bash local-scripts/build-binary.sh

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

# Resolve directories relative to this script's location.
# Script is at: local-scripts/build-binary.sh
# Plugin root (harness/) is one level up from local-scripts/.
# Repo root is two levels up from local-scripts/.
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="${SCRIPT_DIR}/.."           # repo root
PLUGIN_DIR="${REPO_ROOT}/harness"      # harness/ (plugin root)

# Determine install directory (harness/bin/)
INSTALL_DIR="${CLAUDE_PLUGIN_ROOT:+${CLAUDE_PLUGIN_ROOT}/bin}"
if [ -z "$INSTALL_DIR" ]; then
  INSTALL_DIR="${PLUGIN_DIR}/bin"
fi
mkdir -p "$INSTALL_DIR"
INSTALL_DIR="$(cd "$INSTALL_DIR" && pwd)"

TARGET="${INSTALL_DIR}/${BINARY_NAME}"

# Already installed — skip rebuild
if [ -x "$TARGET" ]; then
  echo "[harness] binary already installed: $TARGET"
  exit 0
fi

# Locate Go source directory (always at repo root /go)
GO_DIR="$(cd "${REPO_ROOT}/go" 2>/dev/null && pwd)"

if [ ! -d "$GO_DIR" ]; then
  echo "[harness] error: Go source directory not found at ${REPO_ROOT}/go" >&2
  exit 1
fi

# Check Go is available
if ! command -v go >/dev/null 2>&1; then
  echo "[harness] error: 'go' is not installed or not in PATH" >&2
  echo "[harness] install Go from https://go.dev/dl/" >&2
  exit 1
fi

# Read version from harness/VERSION
VERSION="dev"
VERSION_FILE="${CLAUDE_PLUGIN_ROOT:+${CLAUDE_PLUGIN_ROOT}/VERSION}"
[ -z "$VERSION_FILE" ] && VERSION_FILE="${PLUGIN_DIR}/VERSION"
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
