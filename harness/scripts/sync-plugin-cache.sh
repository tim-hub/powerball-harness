#!/bin/bash
# sync-plugin-cache.sh — Harness plugin cache sync
# 1. Delegates CC file generation to `harness sync` (Go binary)
# 2. Syncs critical scripts to marketplace distribution cache
#
# Usage: Called from SessionStart hook or manually

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# PLUGIN_ROOT: the harness/ directory (where scripts, hooks, VERSION live)
PLUGIN_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"
# REPO_ROOT: the repository root (parent of harness/) — required by `harness sync`
REPO_ROOT="$(cd "$PLUGIN_ROOT/.." && pwd)"

# --- Step 1: Run harness sync (Go binary) ---
# `harness sync` expects the repo root and appends harness/ itself (outputDir="harness").
# Passing PLUGIN_ROOT would make it write to harness/harness/ — always use REPO_ROOT here.
# Best-effort: if the binary is unavailable (e.g. in CI before build), skip
# and rely on the committed .claude-plugin/* files for Step 2 distribution.
sync_ok=0
if command -v harness >/dev/null 2>&1; then
  harness sync "$REPO_ROOT" && sync_ok=1
elif [ -x "${PLUGIN_ROOT}/bin/harness" ] && "${PLUGIN_ROOT}/bin/harness" sync "$REPO_ROOT" 2>/dev/null; then
  sync_ok=1
fi
if [ "$sync_ok" = 0 ]; then
  echo "Warning: harness binary not found or failed; using committed .claude-plugin/* files." >&2
fi

# --- Step 2: Sync critical files to marketplace cache ---
# Paths are relative to PLUGIN_ROOT (harness/), not REPO_ROOT.
PLUGIN_NAME="harness"
MARKETPLACE_NAME="powerball-harness-marketplace"
SOURCE_VERSION="$(tr -d '[:space:]' < "${PLUGIN_ROOT}/VERSION")"
CACHE_DIR="${HOME}/.claude/plugins/cache/${MARKETPLACE_NAME}/${PLUGIN_NAME}/${SOURCE_VERSION}"

sync_file() {
  local rel_path="$1"
  local src="${PLUGIN_ROOT}/${rel_path}"
  local dst="${CACHE_DIR}/${rel_path}"
  if [ -f "$src" ]; then
    mkdir -p "$(dirname "$dst")"
    cp "$src" "$dst"
  fi
}

# Critical files to sync to distribution cache
# (paths are relative to CLAUDE_PLUGIN_ROOT = harness/)
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
  "settings.json"
  "VERSION"
)

for file in "${critical_files[@]}"; do
  sync_file "$file"
done
