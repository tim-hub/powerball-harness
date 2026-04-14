#!/bin/bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=/dev/null
source "${SCRIPT_DIR}/../lib/harness-mem-bridge.sh"
exec_harness_mem_script "scripts/hook-handlers/memory-session-start.sh" "$@"
