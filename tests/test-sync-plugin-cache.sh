#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HARNESS_DIR="${ROOT_DIR}/harness"
TMP_HOME="$(mktemp -d)"
trap 'rm -rf "${TMP_HOME}"' EXIT

SOURCE_VERSION="$(tr -d '[:space:]' < "${HARNESS_DIR}/VERSION")"
CACHE_DIR="${TMP_HOME}/.claude/plugins/cache/claude-code-harness-marketplace/claude-code-harness/${SOURCE_VERSION}"
mkdir -p "${CACHE_DIR}"

# Prepare a stale/missing cache and verify that passing CLAUDE_PLUGIN_ROOT
# as the plugin root correctly resolves the sync source.
printf 'stale\n' > "${CACHE_DIR}/VERSION"

HOME="${TMP_HOME}" CLAUDE_PLUGIN_ROOT="${HARNESS_DIR}" bash "${HARNESS_DIR}/scripts/sync-plugin-cache.sh" >/dev/null 2>&1

required_cached_files=(
  "${CACHE_DIR}/scripts/lib/harness-mem-bridge.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-bridge.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-session-start.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-user-prompt.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-post-tool-use.sh"
  "${CACHE_DIR}/scripts/hook-handlers/memory-stop.sh"
  "${CACHE_DIR}/scripts/hook-handlers/runtime-reactive.sh"
  "${CACHE_DIR}/hooks/hooks.json"
  "${CACHE_DIR}/settings.json"
)

for file in "${required_cached_files[@]}"; do
  if [[ ! -f "${file}" ]]; then
    echo "sync-plugin-cache did not populate required file: ${file}"
    exit 1
  fi
done

echo "OK"
