#!/usr/bin/env bash
# PreCompact hook: inject a systemMessage warning if plans.json has cc:WIP tasks.
# Outputs JSON with systemMessage when WIP tasks exist; exits 0 either way (never blocks compaction).
# SSOT: .claude/harness/plans.json (via config-utils.sh helpers).

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./config-utils.sh
source "${SCRIPT_DIR}/config-utils.sh"

# No plans.json or no WIP tasks → nothing to warn about
plans_json_exists || exit 0
plans_has_wip || exit 0

wip_lines=$(plans_wip_names 5 | tr '\n' '; ')

echo "{\"systemMessage\": \"Warning: Compacting context with WIP tasks in progress: ${wip_lines}Key context about these tasks may be lost after compaction. Consider completing or checkpointing them first.\"}"
exit 0
