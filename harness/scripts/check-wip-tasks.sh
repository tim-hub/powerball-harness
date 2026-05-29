#!/usr/bin/env bash
# Stop hook: block session stop if plans.json has cc:WIP tasks.
# Exit 0 = allow stop, Exit 2 + JSON = block stop.
# SSOT: .claude/harness/plans.json (via config-utils.sh helpers).

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./config-utils.sh
source "${SCRIPT_DIR}/config-utils.sh"

# No plans.json → nothing to block on
plans_json_exists || exit 0

# No WIP tasks → allow stop
plans_has_wip || exit 0

# Names of WIP tasks (up to 5), joined for the message
wip_list=$(plans_wip_names 5 | tr '\n' '; ')

echo "{\"decision\":\"block\",\"reason\":\"plans.json has WIP task(s) remaining (the SSOT is .claude/harness/plans.json): ${wip_list}Consider completing them or marking as blocked before stopping.\"}"
exit 2
