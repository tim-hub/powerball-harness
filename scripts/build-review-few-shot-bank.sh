#!/bin/bash
# build-review-few-shot-bank.sh
# Generate few-shot bank from review-calibration.jsonl.

set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 2
fi

INPUT_FILE="${1:-.claude/state/review-calibration.jsonl}"
OUTPUT_FILE="${2:-.claude/state/review-few-shot-bank.json}"

if [ ! -f "$INPUT_FILE" ]; then
  echo "Calibration log not found: $INPUT_FILE" >&2
  exit 3
fi

TMP_JSON="$(mktemp)"
trap 'rm -f "$TMP_JSON"' EXIT

jq -s -c '
  map(select(.calibration_label != null and .calibration_label != ""))
  | sort_by(.generated_at)
  | reverse
  | group_by([.calibration_label, .reviewer_profile, .review_type] | join("|"))
  | map(.[:3] | map({
      calibration_label,
      reviewer_profile,
      review_type,
      verdict,
      task,
      issue_summary: (
        (.gaps[0].issue // .review_result_snapshot.verdict // "")
      ),
      prompt_hint,
      calibration_notes,
      execution
    }))
  | flatten
' "$INPUT_FILE" > "$TMP_JSON"

jq -n \
  --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg source_file "$INPUT_FILE" \
  --slurpfile entries "$TMP_JSON" \
  '{
    schema_version: "review-few-shot-bank.v1",
    generated_at: $generated_at,
    source_file: $source_file,
    entries: $entries[0]
  }' > "$OUTPUT_FILE"

echo "$OUTPUT_FILE"
