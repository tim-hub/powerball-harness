#!/usr/bin/env bash
# PreCompact hook: inject a systemMessage warning if Plans.md has cc:WIP tasks.
# Outputs JSON with systemMessage when WIP tasks exist; exits 0 either way (never blocks compaction).

PLANS="Plans.md"

# No Plans.md → nothing to warn about
if [[ ! -f "$PLANS" ]]; then
  exit 0
fi

# Collect WIP task lines (up to 5)
wip_lines=$(grep 'cc:WIP' "$PLANS" | head -5 | sed 's/^[[:space:]]*//' | tr '\n' '; ')

if [[ -z "$wip_lines" ]]; then
  exit 0
fi

echo "{\"systemMessage\": \"Warning: Compacting context with WIP tasks in progress: ${wip_lines}Key context about these tasks may be lost after compaction. Consider completing or checkpointing them first.\"}"
exit 0
