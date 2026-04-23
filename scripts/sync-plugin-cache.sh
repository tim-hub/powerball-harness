#!/bin/bash
# sync-plugin-cache.sh — Harness plugin cache sync
# 1. Delegates CC file generation to `harness sync` (Go binary)
# 2. Syncs critical scripts to marketplace distribution cache
#
# Usage: Called from SessionStart hook or manually

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

is_harness_root() {
  local candidate="${1:-}"
  [ -n "$candidate" ] &&
    [ -x "$candidate/bin/harness" ] &&
    [ -f "$candidate/.claude-plugin/plugin.json" ] &&
    grep -q '"name"[[:space:]]*:[[:space:]]*"claude-code-harness"' "$candidate/.claude-plugin/plugin.json"
}

PROJECT_ROOT="${CLAUDE_PLUGIN_ROOT:-$DEFAULT_PROJECT_ROOT}"
if ! is_harness_root "$PROJECT_ROOT"; then
  if is_harness_root "$DEFAULT_PROJECT_ROOT"; then
    PROJECT_ROOT="$DEFAULT_PROJECT_ROOT"
  else
    echo "Error: could not resolve claude-code-harness plugin root." >&2
    exit 1
  fi
fi

# --- Step 1: Run harness sync (Go binary) ---
# Best-effort: if the binary is unavailable (e.g. in CI before build), skip
# and rely on the committed .claude-plugin/* files for Step 2 distribution.
sync_ok=0
if command -v harness >/dev/null 2>&1; then
  harness sync "$PROJECT_ROOT" && sync_ok=1
elif [ -x "${PROJECT_ROOT}/bin/harness" ] && "${PROJECT_ROOT}/bin/harness" sync "$PROJECT_ROOT" 2>/dev/null; then
  sync_ok=1
fi
if [ "$sync_ok" = 0 ]; then
  echo "Warning: harness binary not found or failed; using committed .claude-plugin/* files." >&2
fi

# --- Step 2: Sync critical files to marketplace cache ---
PLUGIN_NAME="claude-code-harness"
MARKETPLACE_NAME="claude-code-harness-marketplace"
SOURCE_VERSION="$(tr -d '[:space:]' < "${PROJECT_ROOT}/VERSION")"
CACHE_DIR="${HOME}/.claude/plugins/cache/${MARKETPLACE_NAME}/${PLUGIN_NAME}/${SOURCE_VERSION}"
MARKETPLACE_DIR="${HOME}/.claude/plugins/marketplaces/${MARKETPLACE_NAME}"

sync_file_to_dir() {
  local rel_path="$1"
  local target_dir="$2"
  local src="${PROJECT_ROOT}/${rel_path}"
  local dst="${target_dir}/${rel_path}"
  if [ -f "$src" ]; then
    mkdir -p "$(dirname "$dst")"
    cp "$src" "$dst"
  fi
}

sync_file() {
  local rel_path="$1"

  sync_file_to_dir "$rel_path" "$CACHE_DIR"

  # If a local marketplace checkout is installed, keep its hook definitions in
  # lockstep too. Claude may load hooks from this path before the versioned cache.
  if [ -d "$MARKETPLACE_DIR" ]; then
    sync_file_to_dir "$rel_path" "$MARKETPLACE_DIR"
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
