#!/bin/bash
# sync-plugin-cache.sh — Harness plugin cache sync
# Now delegates to `harness sync` (Go binary).
# Kept as a shell wrapper for backward compatibility with existing workflows.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Go バイナリ優先
if command -v harness >/dev/null 2>&1; then
  harness sync "$PROJECT_ROOT"
elif [ -x "${PROJECT_ROOT}/bin/harness" ]; then
  "${PROJECT_ROOT}/bin/harness" sync "$PROJECT_ROOT"
else
  echo "Error: harness binary not found. Run 'cd go && make install' first." >&2
  exit 1
fi
