#!/bin/bash
# write-review-result.sh
# さまざまな review 出力を review-result.v1 に正規化し、後方互換の review-approved.json も更新する。

set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 2
fi

INPUT_FILE="${1:-}"
COMMIT_HASH="${2:-}"
OUTPUT_FILE="${3:-.claude/state/review-result.json}"
LEGACY_FILE=".claude/state/review-approved.json"

if [ -z "$INPUT_FILE" ]; then
  echo "Usage: scripts/write-review-result.sh <input-json-file> [commit-hash] [output-file]" >&2
  exit 1
fi

if [ ! -f "$INPUT_FILE" ]; then
  echo "Input file not found: $INPUT_FILE" >&2
  exit 3
fi

mkdir -p "$(dirname "$OUTPUT_FILE")"
mkdir -p "$(dirname "$LEGACY_FILE")"

jq -n \
  --slurpfile src "$INPUT_FILE" \
  --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg commit_hash "$COMMIT_HASH" '
  def as_array(v):
    if v == null then []
    elif (v | type) == "array" then v
    else [v]
    end;
  def normalize_gap(item; severity):
    if (item | type) == "string" then
      {severity: severity, issue: item}
    else
      item + {severity: (item.severity // severity)}
    end;
  # companion verdict 正規化: approve→APPROVE, needs-attention→REQUEST_CHANGES
  def normalize_verdict(v):
    if v == "approve" then "APPROVE"
    elif v == "needs-attention" then "REQUEST_CHANGES"
    else v
    end;
  # companion findings[] → gaps[] マッピング（引数で入力を受け取る）
  # companion findings → gaps（critical/high のみブロッキング）
  def findings_to_gaps(input):
    as_array(input.findings) | map({
      severity: .severity,
      issue: (.title // .body // ""),
      file: (.file // null),
      line_start: (.line_start // null),
      line_end: (.line_end // null),
      recommendation: (.recommendation // "")
    }) | map(select(.severity | IN("critical","high")));
  # companion findings → followups（medium/low は非ブロッキング）
  def findings_to_followups(input):
    as_array(input.findings) | map({
      severity: .severity,
      issue: (.title // .body // ""),
      file: (.file // null),
      recommendation: (.recommendation // "")
    }) | map(select(.severity | IN("medium","low")));
  ($src[0] // {}) as $in
  | {
      schema_version: "review-result.v1",
      generated_at: $generated_at,
      verdict: normalize_verdict($in.verdict // $in.judgment // "REQUEST_CHANGES"),
      reviewer_profile: ($in.reviewer_profile // "static"),
      task: ($in.task // null),
      type: ($in.type // $in.review_type // null),
      commit_hash: (if $commit_hash == "" then ($in.commit_hash // null) else $commit_hash end),
      execution: (
        if (($in.route // null) != null) or (($in.mode // null) != null) or (($in.browser_mode // null) != null) or (($in.tool_matcher // null) != null) or (($in.required_artifacts // null) != null) or (($in.execution_instructions // null) != null) then
          {
            route: ($in.route // null),
            mode: ($in.mode // $in.browser_mode // null),
            tool_matcher: ($in.tool_matcher // null),
            browser_mode: ($in.browser_mode // null),
            required_artifacts: (
              if ($in.required_artifacts // null) == null then []
              else $in.required_artifacts
              end
            ),
            instructions: (
              if ($in.execution_instructions // null) == null then []
              elif ($in.execution_instructions | type) == "array" then $in.execution_instructions
              else [$in.execution_instructions]
              end
            )
          }
        else null
        end
      ),
      calibration: (
        if (($in.calibration // null) != null) or (($in.calibration_label // null) != null) then
          {
            label: ($in.calibration.label // $in.calibration_label // null),
            source: ($in.calibration.source // "manual"),
            notes: ($in.calibration.notes // ""),
            prompt_hint: ($in.calibration.prompt_hint // ""),
            few_shot_ready: ($in.calibration.few_shot_ready // true)
          }
        else null
        end
      ),
      checks: (
        if ($in.checks // null) != null then $in.checks
        else []
        end
      ),
      gaps: (
        (as_array($in.gaps) | map(normalize_gap(.; "major")))
        + (as_array($in.critical_issues) | map(normalize_gap(.; "critical")))
        + (as_array($in.major_issues) | map(normalize_gap(.; "major")))
        + (as_array($in.observations) | map(select((.severity // "minor") | IN("critical","major"))))
        + findings_to_gaps($in)
      ),
      followups: (
        as_array($in.followups)
        + as_array($in.recommendations)
        + (as_array($in.observations) | map(select((.severity // "minor") | IN("minor","recommendation"))))
        + findings_to_followups($in)
      )
    }' > "$OUTPUT_FILE"

# blocking gaps がある場合は verdict を REQUEST_CHANGES に強制
BLOCKING_GAPS="$(jq '[.gaps[] | select(.severity == "critical" or .severity == "high" or .severity == "major")] | length' "$OUTPUT_FILE" 2>/dev/null || echo 0)"
if [ "$BLOCKING_GAPS" -gt 0 ]; then
  jq '.verdict = "REQUEST_CHANGES"' "$OUTPUT_FILE" > "${OUTPUT_FILE}.tmp" && mv "${OUTPUT_FILE}.tmp" "$OUTPUT_FILE"
fi

VERDICT="$(jq -r '.verdict' "$OUTPUT_FILE")"
CALIBRATION_PRESENT="$(jq -r '.calibration.label // empty' "$OUTPUT_FILE")"
if [ "$VERDICT" = "APPROVE" ]; then
  jq -n \
    --arg approved_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --arg judgment "APPROVE" \
    --arg commit_hash "$COMMIT_HASH" \
    '{approved_at:$approved_at, judgment:$judgment, commit_hash:(if $commit_hash == "" then null else $commit_hash end)}' > "$LEGACY_FILE"
else
  rm -f "$LEGACY_FILE" 2>/dev/null || true
fi

if [ -n "$CALIBRATION_PRESENT" ]; then
  CALIBRATION_SCRIPT="$(dirname "$0")/record-review-calibration.sh"
  FEW_SHOT_SCRIPT="$(dirname "$0")/build-review-few-shot-bank.sh"
  CALIBRATION_LOG=".claude/state/review-calibration.jsonl"
  FEW_SHOT_BANK=".claude/state/review-few-shot-bank.json"
  if [ -x "$CALIBRATION_SCRIPT" ]; then
    "$CALIBRATION_SCRIPT" "$OUTPUT_FILE" >/dev/null
  fi
  if [ -x "$FEW_SHOT_SCRIPT" ] && [ -f "$CALIBRATION_LOG" ]; then
    "$FEW_SHOT_SCRIPT" "$CALIBRATION_LOG" "$FEW_SHOT_BANK" >/dev/null
  fi
fi

echo "$OUTPUT_FILE"
