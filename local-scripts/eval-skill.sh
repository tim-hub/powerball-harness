#!/usr/bin/env bash
# eval-skill.sh — Score a skill variant against a golden-verdict evaluation suite
#
# Usage:
#   bash local-scripts/eval-skill.sh <skill-dir> <eval-suite-dir> [--model <model>]
#
# Eval suite layout (one pair per test case):
#   <eval-suite-dir>/<id>.diff            PR diff text to review
#   <eval-suite-dir>/<id>.expected.json   {"verdict":"APPROVE"|"REQUEST_CHANGES","rationale":"..."}
#
# Score function: pass/fail against golden verdicts
#   score = correct_verdicts / total_cases
#
# Output (stdout): JSON score report
#   {
#     "skill":  "<skill-dir>",
#     "suite":  "<eval-suite-dir>",
#     "model":  "<model>",
#     "total":  N,
#     "passed": M,
#     "score":  0.XX,
#     "cases":  [
#       {"id":"case1","expected":"APPROVE","actual":"APPROVE","passed":true},
#       ...
#     ]
#   }
#
# Exit codes:
#   0 — completed (score may be 0)
#   1 — no .diff files found in eval-suite-dir
#   2 — invocation error

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# ---- arg parsing ----
SKILL_DIR=""
EVAL_SUITE_DIR=""
MODEL="haiku"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --model) MODEL="$2"; shift 2 ;;
    -*) echo "Unknown flag: $1" >&2; exit 2 ;;
    *)
      if [[ -z "$SKILL_DIR" ]]; then SKILL_DIR="$1"
      elif [[ -z "$EVAL_SUITE_DIR" ]]; then EVAL_SUITE_DIR="$1"
      else echo "Unexpected argument: $1" >&2; exit 2
      fi
      shift
      ;;
  esac
done

if [[ -z "$SKILL_DIR" || -z "$EVAL_SUITE_DIR" ]]; then
  echo "Usage: $0 <skill-dir> <eval-suite-dir> [--model <model>]" >&2
  exit 2
fi

if [[ ! -f "${SKILL_DIR}/SKILL.md" ]]; then
  echo "Error: ${SKILL_DIR}/SKILL.md not found" >&2
  exit 2
fi

if [[ ! -d "${EVAL_SUITE_DIR}" ]]; then
  echo "Error: eval suite directory not found: ${EVAL_SUITE_DIR}" >&2
  exit 2
fi

# ---- collect test cases ----
DIFF_FILES=()
while IFS= read -r f; do
  DIFF_FILES+=("$f")
done < <(find "${EVAL_SUITE_DIR}" -name "*.diff" | sort)

if [[ ${#DIFF_FILES[@]} -eq 0 ]]; then
  echo "Error: no .diff files in ${EVAL_SUITE_DIR}" >&2
  exit 1
fi

# ---- extract skill body (strip YAML frontmatter) ----
# Frontmatter is metadata for the skill loader; only the body drives review behavior.
SKILL_BODY=$(awk 'BEGIN{fence=0} /^---$/{fence++; next} fence>=2{print}' "${SKILL_DIR}/SKILL.md")

if [[ -z "$SKILL_BODY" ]]; then
  # No frontmatter found — use the whole file
  SKILL_BODY=$(cat "${SKILL_DIR}/SKILL.md")
fi

EVAL_SYSTEM_PROMPT="${SKILL_BODY}

---
EVALUATION INSTRUCTIONS (override any conflicting guidance above):
You are being evaluated. Review the diff provided by the user.
Your final line MUST be exactly one of: APPROVE or REQUEST_CHANGES
No other text on that final line. No punctuation. No explanation."

# ---- run each case ----
PASSED=0
TOTAL=0
CASES_JSON=""

for diff_file in "${DIFF_FILES[@]}"; do
  base="${diff_file%.diff}"
  case_id="$(basename "${base}")"
  expected_file="${base}.expected.json"

  if [[ ! -f "${expected_file}" ]]; then
    echo "Warning: no .expected.json for ${case_id}, skipping" >&2
    continue
  fi

  expected_verdict=$(jq -r '.verdict' "${expected_file}")
  diff_content=$(cat "${diff_file}")

  user_message="Review the following diff and apply the skill criteria above.

\`\`\`diff
${diff_content}
\`\`\`

Remember: your LAST line must be exactly APPROVE or REQUEST_CHANGES."

  # Run claude in non-interactive print mode.
  # --exclude-dynamic-system-prompt-sections removes per-machine variability (cwd, env)
  # so scores are more comparable across runs and machines.
  raw_output=$(printf '%s' "${user_message}" | \
    claude --print \
      --system-prompt "${EVAL_SYSTEM_PROMPT}" \
      --exclude-dynamic-system-prompt-sections \
      --model "${MODEL}" \
      --output-format text 2>/dev/null || true)

  # Extract verdict: take last occurrence of APPROVE or REQUEST_CHANGES
  actual_verdict=$(printf '%s' "${raw_output}" | \
    grep -oE 'APPROVE|REQUEST_CHANGES' | tail -1 || true)

  if [[ -z "${actual_verdict}" ]]; then
    actual_verdict="UNKNOWN"
  fi

  passed=false
  if [[ "${actual_verdict}" == "${expected_verdict}" ]]; then
    PASSED=$((PASSED + 1))
    passed=true
  fi

  TOTAL=$((TOTAL + 1))

  case_json=$(jq -n \
    --arg id "${case_id}" \
    --arg expected "${expected_verdict}" \
    --arg actual "${actual_verdict}" \
    --argjson passed "${passed}" \
    '{"id":$id,"expected":$expected,"actual":$actual,"passed":$passed}')

  if [[ -n "${CASES_JSON}" ]]; then
    CASES_JSON="${CASES_JSON},"
  fi
  CASES_JSON="${CASES_JSON}${case_json}"
done

if [[ $TOTAL -eq 0 ]]; then
  echo "Error: no valid test cases (each .diff needs a matching .expected.json)" >&2
  exit 1
fi

# ---- emit score report ----
SCORE=$(awk "BEGIN{printf \"%.4f\", ${PASSED}/${TOTAL}}")

jq -n \
  --arg skill "${SKILL_DIR}" \
  --arg suite "${EVAL_SUITE_DIR}" \
  --arg model "${MODEL}" \
  --argjson total "${TOTAL}" \
  --argjson passed "${PASSED}" \
  --argjson score "${SCORE}" \
  --argjson cases "[${CASES_JSON}]" \
  '{skill:$skill,suite:$suite,model:$model,total:$total,passed:$passed,score:$score,cases:$cases}'
