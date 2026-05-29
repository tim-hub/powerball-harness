#!/bin/bash
# calculate-effort.sh
# Reads task information from plans.json and calculates the effort level, printing it to stdout.
#
# Usage:
#   bash scripts/calculate-effort.sh "task description or task ID"
#   echo "task description" | bash scripts/calculate-effort.sh
#
# Output: low / medium / high (stdout)
#
# Scoring criteria:
#   4+ candidate file changes → +2
#   2+ dependent tasks → +1
#   Keywords (refactor, migration, security, cross-cutting) → +1
#   2+ conditions in DoD → +1
#
# Score: 0-2 → low, 3-4 → medium, 5+ → high

set -euo pipefail

# Get task description from arguments or stdin
TASK_INPUT=""
if [ $# -gt 0 ]; then
  TASK_INPUT="$*"
elif [ ! -t 0 ]; then
  # Input from stdin (pipe)
  TASK_INPUT="$(cat)"
fi

if [ -z "$TASK_INPUT" ]; then
  # No input → fallback
  echo "medium"
  exit 0
fi

# Initialize score
SCORE=0

# Resolve plans.json path (fallback order: git root → PROJECT_ROOT → cwd)
# SSOT: <root>/.claude/harness/plans.json
_GIT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
PLANS_JSON="${PROJECT_ROOT:-${_GIT_ROOT:-$(pwd)}}/.claude/harness/plans.json"

# Extract task information from plans.json
# Per-task fields: id, name, dod, depends (array), qualityMarkers (array)
TASK_CONTENT=""
TASK_DOD=""
TASK_DEPENDS=""

if command -v jq >/dev/null 2>&1 && [ -f "$PLANS_JSON" ]; then
  # Search by task ID pattern (#123, 34.2.2, #34.2.2 formats)
  if echo "$TASK_INPUT" | grep -qE '^#?[0-9]+(\.[0-9]+)*$'; then
    TASK_ID_PATTERN=$(echo "$TASK_INPUT" | tr -d '#')
    # Look up the task by exact id. depends is rendered as a comma-joined list.
    _task_tsv=$(jq -r --arg id "$TASK_ID_PATTERN" '
      [.phases[].tasks[]? | select(.id == $id)] | .[0] // empty |
      [ (.name // ""), (.dod // ""), ((.depends // []) | join(",")) ] | @tsv
    ' "$PLANS_JSON" 2>/dev/null || true)
    if [ -n "$_task_tsv" ]; then
      IFS=$'\t' read -r TASK_CONTENT TASK_DOD TASK_DEPENDS <<< "$_task_tsv"
    fi
  fi

  # If not found by task ID, search by keyword in the task name
  if [ -z "$TASK_CONTENT" ]; then
    _needle="$(echo "$TASK_INPUT" | cut -c1-50)"
    _task_tsv=$(jq -r --arg needle "$_needle" '
      [.phases[].tasks[]? | select((.name // "") | contains($needle))] | .[0] // empty |
      [ (.name // ""), (.dod // ""), ((.depends // []) | join(",")) ] | @tsv
    ' "$PLANS_JSON" 2>/dev/null || true)
    if [ -n "$_task_tsv" ]; then
      IFS=$'\t' read -r TASK_CONTENT TASK_DOD TASK_DEPENDS <<< "$_task_tsv"
    fi
  fi
fi

# If not retrievable from plans.json, use the task input itself as the analysis target
if [ -z "$TASK_CONTENT" ]; then
  TASK_CONTENT="$TASK_INPUT"
fi

# Analysis target text (combine task content + DoD + input text)
ANALYSIS_TEXT="${TASK_CONTENT} ${TASK_DOD} ${TASK_INPUT}"

# ---- Scoring ----

# 1. 4+ candidate file changes → +2
# Count file references in the task description (.ts .js .sh .json .md .go .py .rb .tsx .jsx)
FILE_REFS=$(echo "$ANALYSIS_TEXT" | { grep -oE '[a-zA-Z0-9_/-]+\.(ts|tsx|js|jsx|sh|json|md|go|py|rb|css|scss|yaml|yml)' || true; } | wc -l | tr -d '[:space:]')
if [ "${FILE_REFS:-0}" -ge 4 ]; then
  SCORE=$((SCORE + 2))
fi

# 2. 2+ dependent tasks → +1
if [ -n "$TASK_DEPENDS" ]; then
  # Count dependent tasks in the Depends column (dotted ID: 34.1.1, simple ID: #123, comma-separated, etc.)
  # Count dotted IDs first, then count remaining simple numeric IDs
  DEP_COUNT=$(echo "$TASK_DEPENDS" | { grep -oE '#?[0-9]+(\.[0-9]+)+' || true; } | wc -l | tr -d '[:space:]')
  SIMPLE_COUNT=$(echo "$TASK_DEPENDS" | sed -E 's/#?[0-9]+(\.[0-9]+)+//g' | { grep -oE '#?[0-9]+' || true; } | wc -l | tr -d '[:space:]')
  DEP_COUNT=$((DEP_COUNT + SIMPLE_COUNT))
  if [ "${DEP_COUNT:-0}" -ge 2 ]; then
    SCORE=$((SCORE + 1))
  fi
fi

# 3. Keyword check → +1 (added for 1+ matches, no double counting)
KEYWORDS="refactor migration security cross-cutting"
KEYWORD_MATCH=0
for kw in $KEYWORDS; do
  if echo "$ANALYSIS_TEXT" | grep -qi "$kw" 2>/dev/null; then
    KEYWORD_MATCH=1
    break
  fi
done
SCORE=$((SCORE + KEYWORD_MATCH))

# 4. 2+ conditions in DoD → +1
if [ -n "$TASK_DOD" ]; then
  # Count the number of conditions delimited by semicolons, Japanese commas, or commas (number of delimiters + 1 = condition count)
  DOD_DELIMITERS=$(echo "$TASK_DOD" | { grep -oE '[;,]' || true; } | wc -l | tr -d '[:space:]')
  DOD_TOTAL=$(( DOD_DELIMITERS + 1 ))
  if [ "${DOD_TOTAL:-1}" -ge 2 ]; then
    SCORE=$((SCORE + 1))
  fi
fi

# ---- Effort determination ----
if [ "$SCORE" -ge 5 ]; then
  echo "high"
elif [ "$SCORE" -ge 3 ]; then
  echo "medium"
else
  echo "low"
fi
