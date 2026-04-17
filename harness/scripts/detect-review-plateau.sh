#!/bin/bash
# detect-review-plateau.sh
# Detects whether the review fix loop is stuck (plateau) and prompts the Lead to change strategy (pivot).
#
# Usage: ./harness/scripts/detect-review-plateau.sh <task_id> [--calibration-file <path>]
#
# Exit codes:
#   0 = PIVOT_NOT_REQUIRED
#   1 = INSUFFICIENT_DATA
#   2 = PIVOT_REQUIRED
#
# Output (stdout):
#   STATUS: PIVOT_REQUIRED | PIVOT_NOT_REQUIRED | INSUFFICIENT_DATA
#   ENTRIES: <N>
#   JACCARD_AVG: <0.XX>  (only when N>=3)
#   REASON: <description>

set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 1
fi

# --- Argument parsing ---
TASK_ID=""
CALIBRATION_FILE=".claude/state/review-calibration.jsonl"

_positional=()
while [ "$#" -gt 0 ]; do
  case "$1" in
    --help|-h)
      cat <<'EOF'
Usage: harness/scripts/detect-review-plateau.sh <task_id> [--calibration-file <path>]

Detects whether the review fix loop is stuck by analyzing the Jaccard similarity
of the file sets touched in successive review iterations.

Exit codes:
  0 = PIVOT_NOT_REQUIRED  (similarity <= threshold; review is making progress)
  1 = INSUFFICIENT_DATA   (fewer than 3 calibration entries for this task)
  2 = PIVOT_REQUIRED      (similarity > threshold; stuck in same files)
EOF
      exit 0
      ;;
    --calibration-file)
      shift
      CALIBRATION_FILE="${1:-}"
      ;;
    --*)
      # Ignore unknown options
      ;;
    *)
      _positional+=("$1")
      ;;
  esac
  shift
done

if [ "${#_positional[@]}" -ge 1 ]; then
  TASK_ID="${_positional[0]}"
fi

if [ -z "$TASK_ID" ]; then
  echo "Usage: harness/scripts/detect-review-plateau.sh <task_id> [--calibration-file <path>]" >&2
  exit 1
fi

if [ ! -f "$CALIBRATION_FILE" ]; then
  echo "STATUS: INSUFFICIENT_DATA"
  echo "ENTRIES: 0"
  echo "REASON: calibration file not found: $CALIBRATION_FILE"
  exit 1
fi

# --- Extract entries for this task_id (last 3) ---
ENTRIES_JSON="$(jq -c --arg tid "$TASK_ID" \
  'select(.task.id == $tid)' \
  "$CALIBRATION_FILE" 2>/dev/null | tail -3)"

ENTRY_COUNT="$(printf '%s\n' "$ENTRIES_JSON" | jq -s 'length' 2>/dev/null || printf '0')"

# Get total count (check is based on total entries)
ALL_ENTRIES_JSON="$(jq -c --arg tid "$TASK_ID" \
  'select(.task.id == $tid)' \
  "$CALIBRATION_FILE" 2>/dev/null)"

TOTAL_COUNT="$(printf '%s\n' "$ALL_ENTRIES_JSON" | jq -s 'length' 2>/dev/null || printf '0')"

# --- N < 3: INSUFFICIENT_DATA ---
if [ "$TOTAL_COUNT" -lt 3 ]; then
  echo "STATUS: INSUFFICIENT_DATA"
  echo "ENTRIES: $TOTAL_COUNT"
  echo "REASON: not enough calibration entries (need >= 3, found $TOTAL_COUNT)"
  exit 1
fi

# --- N >= 3: analyze the last 3 entries ---

# Extract file sets from an entry
# Priority:
#   1. review_result_snapshot.files_changed[]
#   2. gaps[].location split by ':' (first segment = filename)
#   3. empty set if neither available
extract_files() {
  local entry="$1"
  local files=""

  # 1. review_result_snapshot.files_changed
  files="$(echo "$entry" | jq -r '
    (.review_result_snapshot.files_changed // []) | .[]
  ' 2>/dev/null)"

  if [ -z "$files" ]; then
    # 2. gaps[].location split by ':'
    files="$(echo "$entry" | jq -r '
      (.gaps // [])
      | map(select(.location != null and .location != ""))
      | map(.location | split(":")[0])
      | .[]
    ' 2>/dev/null)"
  fi

  echo "$files"
}

# Store the last 3 entries
ENTRY1="$(echo "$ENTRIES_JSON" | sed -n '1p')"
ENTRY2="$(echo "$ENTRIES_JSON" | sed -n '2p')"
ENTRY3="$(echo "$ENTRIES_JSON" | sed -n '3p')"

# Extract sorted unique file sets
FILES1="$(extract_files "$ENTRY1" | sort -u)"
FILES2="$(extract_files "$ENTRY2" | sort -u)"
FILES3="$(extract_files "$ENTRY3" | sort -u)"

# Jaccard similarity: |A ∩ B| / |A ∪ B|
jaccard() {
  local set_a="$1"
  local set_b="$2"

  # Both empty → similarity 1.0 (same empty set)
  if [ -z "$set_a" ] && [ -z "$set_b" ]; then
    echo "1.0"
    return
  fi

  # One empty → similarity 0.0
  if [ -z "$set_a" ] || [ -z "$set_b" ]; then
    echo "0.0"
    return
  fi

  # Intersection count
  local intersection
  intersection="$(comm -12 \
    <(echo "$set_a" | sort -u) \
    <(echo "$set_b" | sort -u) \
    | wc -l | tr -d ' ')"

  # Union = |A| + |B| - |A ∩ B|
  local count_a count_b union
  count_a="$(echo "$set_a" | sort -u | wc -l | tr -d ' ')"
  count_b="$(echo "$set_b" | sort -u | wc -l | tr -d ' ')"
  union=$(( count_a + count_b - intersection ))

  if [ "$union" -eq 0 ]; then
    echo "1.0"
    return
  fi

  # Floating point division via awk
  awk "BEGIN { printf \"%.4f\", $intersection / $union }"
}

# Compute pairwise Jaccard for 3 pairs
J12="$(jaccard "$FILES1" "$FILES2")"
J13="$(jaccard "$FILES1" "$FILES3")"
J23="$(jaccard "$FILES2" "$FILES3")"

# Average Jaccard
JACCARD_AVG="$(awk "BEGIN { printf \"%.4f\", ($J12 + $J13 + $J23) / 3 }")"

# Plateau threshold: average Jaccard > 0.7 means stuck in same files
THRESHOLD="0.7"
IS_PLATEAU="$(awk "BEGIN { print ($JACCARD_AVG > $THRESHOLD) ? \"yes\" : \"no\" }")"

if [ "$IS_PLATEAU" = "yes" ]; then
  echo "STATUS: PIVOT_REQUIRED"
  echo "ENTRIES: $TOTAL_COUNT"
  echo "JACCARD_AVG: $JACCARD_AVG"
  echo "REASON: review iterations >= 3 and file-set similarity (Jaccard avg $JACCARD_AVG) > $THRESHOLD — stuck in same files"
  exit 2
else
  echo "STATUS: PIVOT_NOT_REQUIRED"
  echo "ENTRIES: $TOTAL_COUNT"
  echo "JACCARD_AVG: $JACCARD_AVG"
  echo "REASON: file-set similarity (Jaccard avg $JACCARD_AVG) <= $THRESHOLD — review is making progress"
  exit 0
fi
