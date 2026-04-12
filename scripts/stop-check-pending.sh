#!/bin/bash
# stop-check-pending.sh
# Check and warn about unresolved pending-skills on Stop
#
# Usage: Auto-executed from Stop hook (type: command)
# Input: stdin JSON (Claude Code hooks)
# Output: Human-readable text warning (output directly to stdout)

set +e

STATE_DIR=".claude/state"
PENDING_DIR="${STATE_DIR}/pending-skills"

# Exit silently if pending directory does not exist
if [ ! -d "$PENDING_DIR" ]; then
  exit 0
fi

# Check pending files
PENDING_FILES=$(ls "$PENDING_DIR"/*.pending 2>/dev/null || true)

if [ -z "$PENDING_FILES" ]; then
  exit 0
fi

# When there are unresolved pending items
PENDING_COMMANDS=""
for f in $PENDING_FILES; do
  CMD_NAME=$(basename "$f" .pending)
  PENDING_COMMANDS="${PENDING_COMMANDS}${CMD_NAME}, "
done
PENDING_COMMANDS=$(echo "$PENDING_COMMANDS" | sed 's/, $//')

# Clear pending (already warned)
rm -f "$PENDING_DIR"/*.pending 2>/dev/null || true

# Output human-readable text warning to stdout
cat <<EOF

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  Quality gate not executed warning
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

The following commands were executed but their corresponding Skills were not invoked:
  → ${PENDING_COMMANDS}

This may cause the following issues:
  1. Missing usage statistics: Skill usage history was not recorded
  2. Quality guardrails not executed: Review/verification skills may not have been applied

Recommendation: Run /harness-review manually to perform quality checks.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

EOF

exit 0
