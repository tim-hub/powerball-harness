#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_HOME="$(mktemp -d)"
trap 'rm -rf "${TMP_HOME}"' EXIT

SOURCE_VERSION="$(tr -d '[:space:]' < "${ROOT_DIR}/VERSION")"
CACHE_DIR="${TMP_HOME}/.claude/plugins/cache/claude-code-harness-marketplace/claude-code-harness/${SOURCE_VERSION}"
MARKETPLACE_DIR="${TMP_HOME}/.claude/plugins/marketplaces/claude-code-harness-marketplace"
mkdir -p "${CACHE_DIR}" "${MARKETPLACE_DIR}/.claude-plugin"

# 古い/欠落したキャッシュと marketplace copy を用意して、CLAUDE_PLUGIN_ROOT を
# plugin root として渡したときに正しく同期元解決できることを確認する。
printf 'stale\n' > "${CACHE_DIR}/VERSION"
printf 'stale\n' > "${MARKETPLACE_DIR}/VERSION"
printf '{"hooks":{"SessionStart":[{"hooks":[{"command":"\"${CLAUDE_PLUGIN_ROOT}/bin/harness\" hook session-start"}]}]}}' > "${MARKETPLACE_DIR}/.claude-plugin/hooks.json"

HOME="${TMP_HOME}" CLAUDE_PLUGIN_ROOT="${ROOT_DIR}" bash "${ROOT_DIR}/scripts/sync-plugin-cache.sh" >/dev/null 2>&1

# 間違った CLAUDE_PLUGIN_ROOT が来ても、script path から実際の plugin root へ
# 戻れることを確認する。hook 実行環境の変数揺れに対する回帰テスト。
INVALID_ROOT="${TMP_HOME}/not-a-plugin-root"
mkdir -p "${INVALID_ROOT}"
HOME="${TMP_HOME}" CLAUDE_PLUGIN_ROOT="${INVALID_ROOT}" bash "${ROOT_DIR}/scripts/sync-plugin-cache.sh" >/dev/null 2>&1

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
  "${MARKETPLACE_DIR}/scripts/lib/harness-mem-bridge.sh"
  "${MARKETPLACE_DIR}/scripts/hook-handlers/memory-bridge.sh"
  "${MARKETPLACE_DIR}/scripts/hook-handlers/memory-session-start.sh"
  "${MARKETPLACE_DIR}/scripts/hook-handlers/memory-user-prompt.sh"
  "${MARKETPLACE_DIR}/scripts/hook-handlers/memory-post-tool-use.sh"
  "${MARKETPLACE_DIR}/scripts/hook-handlers/memory-stop.sh"
  "${MARKETPLACE_DIR}/scripts/hook-handlers/runtime-reactive.sh"
  "${MARKETPLACE_DIR}/hooks/hooks.json"
  "${MARKETPLACE_DIR}/.claude-plugin/hooks.json"
  "${MARKETPLACE_DIR}/.claude-plugin/settings.json"
)

for file in "${required_cached_files[@]}"; do
  if [[ ! -f "${file}" ]]; then
    echo "sync-plugin-cache did not populate required file: ${file}"
    exit 1
  fi
done

for file in "${CACHE_DIR}/.claude-plugin/hooks.json" "${MARKETPLACE_DIR}/.claude-plugin/hooks.json"; do
  if jq -e '.. | objects | select(.command? | strings | test("^\"\\\\$\\\\{CLAUDE_PLUGIN_ROOT\\\\}/bin/harness\"|^bash \"\\\\$\\\\{CLAUDE_PLUGIN_ROOT\\\\}/scripts/"))' "${file}" >/dev/null 2>&1; then
    echo "sync-plugin-cache left raw CLAUDE_PLUGIN_ROOT hook command in: ${file}"
    exit 1
  fi
done

echo "OK"
