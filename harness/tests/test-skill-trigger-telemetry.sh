#!/usr/bin/env bash
# test-skill-trigger-telemetry.sh
# Phase 95.5: verify skill_activated.invocation_trigger telemetry behavior
#
# Assertions:
#   (1) Records all 3 trigger types (human / model / skill-chain) separately
#   (2) Opt-out (HARNESS_SKILL_TELEMETRY_DISABLE=1) prevents writes
#   (3) Per-skill exclude list suppresses individual skills
#   (4) Ledger is append-only (existing records not modified)
#   (5) session_id is truncated to 12-char prefix (privacy)

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HANDLER="${ROOT_DIR}/scripts/skill-trigger-telemetry.sh"

[ -x "${HANDLER}" ] || chmod +x "${HANDLER}"
[ -x "${HANDLER}" ] || {
  echo "FAIL: ${HANDLER} is not executable"
  exit 1
}

# Isolate in a temp project root
TEST_PROJECT="$(mktemp -d)"
trap 'rm -rf "${TEST_PROJECT}"' EXIT
mkdir -p "${TEST_PROJECT}/.claude/state"
LEDGER="${TEST_PROJECT}/.claude/state/skill-trigger-stats.jsonl"

# (1) All 3 trigger types are recorded
for trigger in human model skill-chain; do
  INPUT="$(jq -nc --arg t "${trigger}" \
    '{skill_name: "harness-work", invocation_trigger: $t, session_id: "session-abcdefghijkl-rest", duration_ms: 100}')"
  printf '%s' "${INPUT}" | env CLAUDE_PROJECT_DIR="${TEST_PROJECT}" "${HANDLER}"
done

[ -f "${LEDGER}" ] || { echo "FAIL (1): ledger not created"; exit 1; }
LINE_COUNT="$(wc -l < "${LEDGER}" | tr -d ' ')"
if [ "${LINE_COUNT}" -ne 3 ]; then
  echo "FAIL (1): expected 3 records, got ${LINE_COUNT}"; cat "${LEDGER}"; exit 1
fi
for trigger in human model skill-chain; do
  grep -q "\"invocation_trigger\":\"${trigger}\"" "${LEDGER}" || {
    echo "FAIL (1): trigger ${trigger} not in ledger"; exit 1
  }
done

# (2) Opt-out: no write when HARNESS_SKILL_TELEMETRY_DISABLE=1
INPUT='{"skill_name":"harness-work","invocation_trigger":"human","session_id":"session-x"}'
printf '%s' "${INPUT}" | env CLAUDE_PROJECT_DIR="${TEST_PROJECT}" HARNESS_SKILL_TELEMETRY_DISABLE=1 "${HANDLER}"
NEW_COUNT="$(wc -l < "${LEDGER}" | tr -d ' ')"
[ "${NEW_COUNT}" -eq 3 ] || { echo "FAIL (2): opt-out should not write; got ${NEW_COUNT}"; exit 1; }

# (3) Per-skill exclude suppresses harness-loop
cat > "${TEST_PROJECT}/.claude/settings.local.json" <<'EOF'
{"harness":{"skill_telemetry_exclude":["harness-loop"]}}
EOF
INPUT='{"skill_name":"harness-loop","invocation_trigger":"human","session_id":"session-y"}'
printf '%s' "${INPUT}" | env CLAUDE_PROJECT_DIR="${TEST_PROJECT}" "${HANDLER}"
EXCL_COUNT="$(wc -l < "${LEDGER}" | tr -d ' ')"
[ "${EXCL_COUNT}" -eq 3 ] || { echo "FAIL (3): excluded skill should not be recorded; got ${EXCL_COUNT}"; exit 1; }

# Non-excluded skill is still recorded
INPUT='{"skill_name":"harness-review","invocation_trigger":"model","session_id":"session-z"}'
printf '%s' "${INPUT}" | env CLAUDE_PROJECT_DIR="${TEST_PROJECT}" "${HANDLER}"
NEW_COUNT="$(wc -l < "${LEDGER}" | tr -d ' ')"
[ "${NEW_COUNT}" -eq 4 ] || { echo "FAIL (3b): non-excluded skill should be recorded; got ${NEW_COUNT}"; exit 1; }

# (4) Append-only: first record unchanged after more writes
FIRST_LINE="$(head -n 1 "${LEDGER}")"
INPUT='{"skill_name":"harness-plan","invocation_trigger":"human","session_id":"session-q"}'
printf '%s' "${INPUT}" | env CLAUDE_PROJECT_DIR="${TEST_PROJECT}" "${HANDLER}"
NEW_FIRST="$(head -n 1 "${LEDGER}")"
[ "${FIRST_LINE}" = "${NEW_FIRST}" ] || {
  echo "FAIL (4): append-only violation -- first line changed"; exit 1
}

# (5) session_id truncated to 12 chars
TRUNCATED="$(jq -r '.session_id' < <(head -n 1 "${LEDGER}"))"
[ "${#TRUNCATED}" -eq 12 ] || {
  echo "FAIL (5): session_id should be 12 chars, got ${#TRUNCATED} (${TRUNCATED})"; exit 1
}

echo "PASS: test-skill-trigger-telemetry.sh -- all 5 assertions passed"
