#!/bin/bash
# test-run-advisor-consultation.sh
# advisor consultation wrapper の回帰テスト

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
WRAPPER="${PROJECT_ROOT}/scripts/run-advisor-consultation.sh"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

pass() {
  echo "[PASS] $1"
}

fail() {
  echo "[FAIL] $1" >&2
  exit 1
}

cat > "${TMP_DIR}/request.json" <<'EOF'
{
  "schema_version": "advisor-request.v1",
  "task_id": "43.2.2",
  "reason_code": "retry-threshold",
  "trigger_hash": "43.2.2:retry-threshold:abc",
  "question": "同じ失敗が2回続いた。次に何を変えるべきか",
  "attempt": 2,
  "last_error": "schema parse failed",
  "context_summary": ["wrapper 実装中", "history 追記が必要"]
}
EOF

cat > "${TMP_DIR}/fake-companion.sh" <<'EOF'
#!/bin/bash
set -euo pipefail
MODE="${FAKE_ADVISOR_MODE:-PLAN}"
if [ "${1:-}" != "task" ]; then
  echo "unexpected subcommand: ${1:-}" >&2
  exit 2
fi
case "${MODE}" in
  PLAN)
    cat <<'JSON'
{"schema_version":"advisor-response.v1","decision":"PLAN","summary":"順番を入れ替える","executor_instructions":["status を先に固定する"],"confidence":0.81,"stop_reason":null}
JSON
    ;;
  CORRECTION)
    cat <<'JSON'
{"schema_version":"advisor-response.v1","decision":"CORRECTION","summary":"局所修正で足りる","executor_instructions":["JSON validation を先に通す"],"confidence":0.72,"stop_reason":null}
JSON
    ;;
  STOP)
    cat <<'JSON'
{"schema_version":"advisor-response.v1","decision":"STOP","summary":"ここで止める","executor_instructions":["ユーザー判断を待つ"],"confidence":0.93,"stop_reason":"dangerous-migration"}
JSON
    ;;
  INVALID)
    echo '{"schema_version":"advisor-response.v1","decision":"PLAN"'
    ;;
  TIMEOUT)
    sleep 3
    ;;
  TIMEOUT_WITH_OUTPUT)
    echo "partial stdout before timeout"
    echo "partial stderr before timeout" >&2
    sleep 3
    ;;
  *)
    echo "unknown mode: ${MODE}" >&2
    exit 2
    ;;
esac
EOF
chmod +x "${TMP_DIR}/fake-companion.sh"

run_case() {
  local mode="$1"
  local expected_decision="$2"
  local response_file="${TMP_DIR}/${mode}.response.json"
  local output_file="${TMP_DIR}/${mode}.stdout"
  CODEX_ADVISOR_COMPANION="${TMP_DIR}/fake-companion.sh" \
    FAKE_ADVISOR_MODE="${mode}" \
    bash "${WRAPPER}" \
      --request-file "${TMP_DIR}/request.json" \
      --response-file "${response_file}" \
      --model fake-model > "${output_file}"

  jq -e --arg decision "${expected_decision}" '.decision == $decision' "${response_file}" >/dev/null \
    || fail "${mode}: decision mismatch"
  grep -q "${expected_decision}" "${output_file}" || fail "${mode}: stdout missing response"
  pass "${mode}: decision ${expected_decision}"
}

run_case PLAN PLAN
run_case CORRECTION CORRECTION
run_case STOP STOP

set +e
CODEX_ADVISOR_COMPANION="${TMP_DIR}/fake-companion.sh" \
  FAKE_ADVISOR_MODE="INVALID" \
  bash "${WRAPPER}" \
    --request-file "${TMP_DIR}/request.json" \
    --response-file "${TMP_DIR}/invalid.response.json" >"${TMP_DIR}/invalid.stdout" 2>"${TMP_DIR}/invalid.stderr"
INVALID_EXIT=$?
set -e
[ "${INVALID_EXIT}" -ne 0 ] || fail "INVALID: wrapper should fail"
[ ! -f "${TMP_DIR}/invalid.response.json" ] || fail "INVALID: broken response file should not be written"
pass "INVALID: broken JSON rejected"

set +e
CODEX_ADVISOR_COMPANION="${TMP_DIR}/fake-companion.sh" \
  FAKE_ADVISOR_MODE="TIMEOUT" \
  bash "${WRAPPER}" \
    --request-file "${TMP_DIR}/request.json" \
    --response-file "${TMP_DIR}/timeout.response.json" \
    --timeout-sec 1 >"${TMP_DIR}/timeout.stdout" 2>"${TMP_DIR}/timeout.stderr"
TIMEOUT_EXIT=$?
set -e
[ "${TIMEOUT_EXIT}" -eq 124 ] || fail "TIMEOUT: expected exit 124 got ${TIMEOUT_EXIT}"
[ ! -f "${TMP_DIR}/timeout.response.json" ] || fail "TIMEOUT: timeout response file should not be written"
grep -q "timed out" "${TMP_DIR}/timeout.stderr" || fail "TIMEOUT: stderr should mention timeout"
pass "TIMEOUT: standardized timeout exit"

# Regression: when the subprocess emits output before the timeout fires,
# TimeoutExpired.stdout/stderr arrive as bytes even with text=True, and the
# old handler crashed with "TypeError: can't concat str to bytes".
set +e
CODEX_ADVISOR_COMPANION="${TMP_DIR}/fake-companion.sh" \
  FAKE_ADVISOR_MODE="TIMEOUT_WITH_OUTPUT" \
  bash "${WRAPPER}" \
    --request-file "${TMP_DIR}/request.json" \
    --response-file "${TMP_DIR}/timeout-output.response.json" \
    --timeout-sec 1 >"${TMP_DIR}/timeout-output.stdout" 2>"${TMP_DIR}/timeout-output.stderr"
TIMEOUT_OUTPUT_EXIT=$?
set -e
[ "${TIMEOUT_OUTPUT_EXIT}" -eq 124 ] || fail "TIMEOUT_WITH_OUTPUT: expected exit 124 got ${TIMEOUT_OUTPUT_EXIT}"
[ ! -f "${TMP_DIR}/timeout-output.response.json" ] || fail "TIMEOUT_WITH_OUTPUT: response file should not be written"
grep -q "timed out" "${TMP_DIR}/timeout-output.stderr" || fail "TIMEOUT_WITH_OUTPUT: stderr should mention timeout"
grep -qv "TypeError" "${TMP_DIR}/timeout-output.stderr" || fail "TIMEOUT_WITH_OUTPUT: stderr should not contain TypeError"
pass "TIMEOUT_WITH_OUTPUT: bytes output before timeout handled cleanly"

echo "test-run-advisor-consultation: ok"
