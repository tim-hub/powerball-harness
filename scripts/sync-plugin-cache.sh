#!/bin/bash
# sync-plugin-cache.sh — Harness plugin cache sync
# 1. Delegates CC file generation to `harness sync` (Go binary)
# 2. Syncs critical scripts to marketplace distribution cache
#
# Usage: Called from SessionStart hook or manually

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"

# --- Step 1: Run harness sync (Go binary) ---
if command -v harness >/dev/null 2>&1; then
  harness sync "$PROJECT_ROOT"
elif [ -x "${PROJECT_ROOT}/bin/harness" ]; then
  "${PROJECT_ROOT}/bin/harness" sync "$PROJECT_ROOT"
else
  echo "Error: harness binary not found. Run 'cd go && make install' first." >&2
  exit 1
fi

# --- Step 2: Sync critical files to marketplace cache ---
PLUGIN_NAME="claude-code-harness"
MARKETPLACE_NAME="claude-code-harness-marketplace"
SOURCE_VERSION="$(tr -d '[:space:]' < "${PROJECT_ROOT}/VERSION")"
CACHE_DIR="${HOME}/.claude/plugins/cache/${MARKETPLACE_NAME}/${PLUGIN_NAME}/${SOURCE_VERSION}"

sync_file() {
  local rel_path="$1"
  local src="${PROJECT_ROOT}/${rel_path}"
  local dst="${CACHE_DIR}/${rel_path}"
  if [ -f "$src" ]; then
    mkdir -p "$(dirname "$dst")"
    cp "$src" "$dst"
  fi
}

# Critical files to sync to distribution cache
critical_files=(
  "scripts/lib/harness-mem-bridge.sh"
  "scripts/hook-handlers/memory-bridge.sh"
  "scripts/hook-handlers/memory-session-start.sh"
  "scripts/hook-handlers/memory-user-prompt.sh"
  "scripts/hook-handlers/memory-post-tool-use.sh"
  "scripts/hook-handlers/memory-stop.sh"
  "scripts/hook-handlers/memory-codex-notify.sh"
  "scripts/hook-handlers/runtime-reactive.sh"
  "hooks/hooks.json"
  ".claude-plugin/hooks.json"
  ".claude-plugin/settings.json"
  ".claude-plugin/plugin.json"
  "VERSION"
)

for file in "${critical_files[@]}"; do
  sync_file "$file"
done
