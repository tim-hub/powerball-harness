#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="${1:-}"
shift || true

# shellcheck source=/dev/null
source "${SCRIPT_DIR}/../lib/harness-mem-bridge.sh"

case "${TARGET}" in
  session-start)
    exec_harness_mem_script "scripts/hook-handlers/memory-session-start.sh" "$@"
    ;;
  user-prompt)
    exec_harness_mem_script "scripts/hook-handlers/memory-user-prompt.sh" "$@"
    ;;
  post-tool-use)
    exec_harness_mem_script "scripts/hook-handlers/memory-post-tool-use.sh" "$@"
    ;;
  stop)
    exec_harness_mem_script "scripts/hook-handlers/memory-stop.sh" "$@"
    ;;
  codex-notify)
    exec_harness_mem_script "scripts/hook-handlers/memory-codex-notify.sh" "$@"
    ;;
  *)
    echo "[claude-code-harness] unknown memory bridge target: ${TARGET}" >&2
    exit 0
    ;;
esac
