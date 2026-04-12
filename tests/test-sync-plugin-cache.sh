#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_HOME="$(mktemp -d)"
trap 'rm -rf "${TMP_HOME}"' EXIT

SOURCE_VERSION="$(tr -d '[:space:]' < "${ROOT_DIR}/VERSION")"
CACHE_DIR="${TMP_HOME}/.claude/plugins/cache/claude-code-harness-marketplace/claude-code-harness/${SOURCE_VERSION}"
mkdir -p "${CACHE_DIR}"

# Prepare stale/missing cache entries and verify that passing CLAUDE_PLUGIN_ROOT
# as the plugin root resolves the sync source correctly.
printf 'stale\n' > "${CACHE_DIR}/VERSION"

HOME="${TMP_HOME}" CLAUDE_PLUGIN_ROOT="${ROOT_DIR}" bash "${ROOT_DIR}/scripts/sync-plugin-cache.sh" >/dev/null 2>&1

required_cached_files=(
  "${CACHE_DIR}/scripts/lib/harness-mem-bridge.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-bridge.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-session-start.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-user-prompt.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-post-tool-use.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-stop.sh"
  "${CACHE_DIR}/scripts/hook-handlers/runtime-reactive.sh"
  "${CACHE_DIR}/hooks/hooks.json"
  "${CACHE_DIR}/.claude-plugin/hooks.json"
  "${CACHE_DIR}/.claude-plugin/settings.json"
)

for file in "${required_cached_files[@]}"; do
  if [[ ! -f "${file}" ]]; then
    echo "sync-plugin-cache did not populate required file: ${file}"
    exit 1
  fi
done

echo "OK"
