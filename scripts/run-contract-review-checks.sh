#!/bin/bash
# run-contract-review-checks.sh
# sprint-contract.json に定義された runtime_validation を順番に実行し、
# 実行結果を review artifact として保存する。

set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 2
fi

CONTRACT_FILE="${1:-}"
OUTPUT_FILE="${2:-}"

if [ -z "$CONTRACT_FILE" ]; then
  echo "Usage: scripts/run-contract-review-checks.sh <contract-file> [output-file]" >&2
  exit 1
fi

if [ ! -f "$CONTRACT_FILE" ]; then
  echo "Contract file not found: $CONTRACT_FILE" >&2
  exit 3
fi

PROFILE="$(jq -r '.review.reviewer_profile // "static"' "$CONTRACT_FILE")"
TASK_ID="$(jq -r '.task.id // "unknown"' "$CONTRACT_FILE")"
TASK_TITLE="$(jq -r '.task.title // ""' "$CONTRACT_FILE")"
# 絶対パスで出力（worktree から呼ばれても Lead 側で解決できるように）
STATE_DIR="$(pwd)/.claude/state/review"
mkdir -p "$STATE_DIR"

if [ -z "$OUTPUT_FILE" ]; then
  OUTPUT_FILE="${STATE_DIR}/${TASK_ID}.runtime-review.json"
fi

if [ "$PROFILE" = "static" ]; then
  jq -n \
    --arg profile "$PROFILE" \
    --arg task_id "$TASK_ID" \
    --arg task_title "$TASK_TITLE" \
    --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    '{
      schema_version: "runtime-review.v1",
      generated_at: $generated_at,
      task: { id: $task_id, title: $task_title },
      reviewer_profile: $profile,
      verdict: "SKIPPED",
      note: "static profile does not execute runtime commands",
      checks: []
    }' > "$OUTPUT_FILE"
  echo "$OUTPUT_FILE"
  exit 0
fi

if [ "$PROFILE" = "browser" ]; then
  jq -n \
    --arg profile "$PROFILE" \
    --arg task_id "$TASK_ID" \
    --arg task_title "$TASK_TITLE" \
    --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    '{
      schema_version: "runtime-review.v1",
      generated_at: $generated_at,
      task: { id: $task_id, title: $task_title },
      reviewer_profile: $profile,
      verdict: "PENDING_BROWSER",
      note: "browser profile requires a browser-capable evaluator",
      checks: []
    }' > "$OUTPUT_FILE"
  echo "$OUTPUT_FILE"
  exit 0
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
CHECKS_FILE="$TMP_DIR/checks.jsonl"
FAILED=0

while IFS= read -r item; do
  label="$(printf '%s' "$item" | jq -r '.label // "unnamed-check"')"
  command="$(printf '%s' "$item" | jq -r '.command // empty')"
  if [ -z "$command" ]; then
    jq -nc --arg label "$label" '{"label":$label,"status":"skipped","reason":"empty command"}' >> "$CHECKS_FILE"
    continue
  fi

  set +e
  output="$(bash -lc "$command" 2>&1)"
  exit_code=$?
  set -e

  if [ $exit_code -ne 0 ]; then
    FAILED=1
    status="failed"
  else
    status="passed"
  fi

  jq -nc \
    --arg label "$label" \
    --arg command "$command" \
    --arg status "$status" \
    --arg output "$output" \
    --argjson exit_code "$exit_code" \
    '{label:$label,command:$command,status:$status,exit_code:$exit_code,output:$output}' >> "$CHECKS_FILE"
done < <(jq -c '.contract.runtime_validation[]? // empty' "$CONTRACT_FILE")

if [ -f "$CHECKS_FILE" ]; then
  checks_json="$(jq -s '.' "$CHECKS_FILE")"
else
  checks_json='[]'
fi

# runtime_validation が空の場合は static プロファイルへの降格を示す
# 空のまま APPROVE/REQUEST_CHANGES を返すのではなく、
# 呼び出し元がプロファイルを切り替えられるよう DOWNGRADE_TO_STATIC を返す
RAN_CHECKS="$(echo "$checks_json" | jq 'length')"
if [ "$RAN_CHECKS" -eq 0 ]; then
  verdict="DOWNGRADE_TO_STATIC"
  checks_json='[{"label":"no-runtime-checks","status":"skipped","reason":"runtime_validation is empty — fallback to static review"}]'
elif [ "$FAILED" -ne 0 ]; then
  verdict="REQUEST_CHANGES"
else
  verdict="APPROVE"
fi

jq -n \
  --arg generated_at "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg task_id "$TASK_ID" \
  --arg task_title "$TASK_TITLE" \
  --arg profile "$PROFILE" \
  --arg verdict "$verdict" \
  --argjson checks "$checks_json" \
  '{
    schema_version: "runtime-review.v1",
    generated_at: $generated_at,
    task: { id: $task_id, title: $task_title },
    reviewer_profile: $profile,
    verdict: $verdict,
    checks: $checks
  }' > "$OUTPUT_FILE"

echo "$OUTPUT_FILE"
