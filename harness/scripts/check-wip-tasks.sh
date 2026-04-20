#!/usr/bin/env bash
# Stop hook: block session stop if Plans.md has cc:WIP tasks.
# Exit 0 = allow stop, Exit 2 + JSON = block stop.

PLANS="Plans.md"

# No Plans.md → allow stop
if [[ ! -f "$PLANS" ]]; then
  exit 0
fi

# Match cc:WIP only in the status column (last field before trailing |)
wip_tasks=$(awk -F'|' 'NF > 2 && $(NF-1) ~ /cc:WIP/' "$PLANS")

if [[ -z "$wip_tasks" ]]; then
  exit 0
fi

# Extract task IDs from first column (up to 5)
wip_list=$(awk -F'|' 'NF > 2 && $(NF-1) ~ /cc:WIP/ {gsub(/^[ \t]+|[ \t]+$/, "", $2); print $2}' "$PLANS" | head -5 | tr '\n' '; ')

echo "{\"decision\":\"block\",\"reason\":\"WIP tasks remain: ${wip_list}Consider completing them or marking as blocked before stopping.\"}"
exit 2
