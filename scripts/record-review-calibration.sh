#!/bin/bash
# record-review-calibration.sh
# Append to learning log when calibration is present in review-result.json.

set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 2
fi

INPUT_FILE="${1:-}"
OUTPUT_FILE="${2:-.claude/state/review-calibration.jsonl}"

if [ -z "$INPUT_FILE" ]; then
  echo "Usage: scripts/record-review-calibration.sh <review-result-file> [output-file]" >&2
  exit 1
fi

if [ ! -f "$INPUT_FILE" ]; then
  echo "Input file not found: $INPUT_FILE" >&2
  exit 3
fi

LABEL="$(jq -r '.calibration.label // empty' "$INPUT_FILE")"
if [ -z "$LABEL" ]; then
  exit 0
fi

case "$LABEL" in
  false_positive|false_negative|missed_bug|overstrict_rule) ;;
  *)
    echo "Invalid calibration label: $LABEL" >&2
    exit 4
    ;;
esac

mkdir -p "$(dirname "$OUTPUT_FILE")"

jq -c -n \
  --slurpfile src "$INPUT_FILE" \
  --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  '
  ($src[0] // {}) as $in
  | {
      schema_version: "review-calibration.v1",
      generated_at: $generated_at,
      task: ($in.task // {id: null, title: null}),
      reviewer_profile: ($in.reviewer_profile // "static"),
      review_type: ($in.type // "code"),
      verdict: ($in.verdict // "REQUEST_CHANGES"),
      calibration_label: ($in.calibration.label // null),
      calibration_source: ($in.calibration.source // "manual"),
      calibration_notes: ($in.calibration.notes // ""),
      prompt_hint: ($in.calibration.prompt_hint // ""),
      few_shot_ready: ($in.calibration.few_shot_ready // true),
      execution: ($in.execution // null),
      checks: ($in.checks // []),
      gaps: ($in.gaps // []),
      followups: ($in.followups // []),
      review_result_snapshot: {
        schema_version: ($in.schema_version // "review-result.v1"),
        verdict: ($in.verdict // null),
        reviewer_profile: ($in.reviewer_profile // null),
        commit_hash: ($in.commit_hash // null)
      }
    }' >> "$OUTPUT_FILE"

echo "$OUTPUT_FILE"
