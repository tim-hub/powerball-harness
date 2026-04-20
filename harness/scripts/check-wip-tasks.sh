#!/usr/bin/env bash
# Stop hook: block session stop if Plans.md has cc:WIP tasks.
# Exit 0 = allow stop, Exit 2 + JSON = block stop.

PLANS="Plans.md"

# No Plans.md → allow stop
if [[ ! -f "$PLANS" ]]; then
  exit 0
fi

# Grep for cc:WIP lines, extract task identifiers
wip_tasks=$(grep -n 'cc:WIP' "$PLANS" | sed 's/:.*//') # line numbers

if [[ -z "$wip_tasks" ]]; then
  exit 0
fi

# Format task list for the reason message
wip_list=$(grep 'cc:WIP' "$PLANS" | head -5 | sed 's/^[[:space:]]*//' | tr '\n' '; ')

echo "{\"decision\":\"block\",\"reason\":\"WIP tasks remain: ${wip_list}Consider completing them or marking as blocked before stopping.\"}"
exit 2
