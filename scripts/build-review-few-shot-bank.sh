#!/bin/bash
# build-review-few-shot-bank.sh
# review-calibration.jsonl から few-shot 用バンクを生成する。

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
      execution,
      # 新フィールド: 旧レコード（フィールド欠如）は // 0 default で読み出す
      critical_count: (.critical_count // 0),
      major_count: (.major_count // 0),
      # score_delta は新フィールドが存在する場合のみ含める（null は除外しない）
      score_delta: (if has("score_delta") then .score_delta else null end)
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
