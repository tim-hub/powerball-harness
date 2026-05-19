#!/bin/bash
# test-terminal-notify.sh
# End-to-end tests for harness/scripts/lib/terminal-notify.sh and its integration
# into notification-handler.sh and webhook-notify.sh.
#
# Covers:
#   - Library presence and permissions
#   - All 5 HARNESS_TERMINAL_NOTIFY modes (off, bell, title, osc9, notify)
#   - Control-character sanitization
#   - Non-ASCII passthrough
#   - JSON encoding via encode_terminal_sequence_json
#   - notification-handler.sh terminalSequence integration
#   - webhook-notify.sh terminalSequence integration

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

TERMINAL_NOTIFY_LIB="${ROOT_DIR}/scripts/lib/terminal-notify.sh"
NOTIFICATION_HANDLER="${ROOT_DIR}/scripts/hook-handlers/notification-handler.sh"
WEBHOOK_NOTIFY="${ROOT_DIR}/scripts/hook-handlers/webhook-notify.sh"

failures=0
passed=0

err()  { echo "FAIL: $*" >&2; failures=$((failures + 1)); }
pass() { echo "pass: $*"; passed=$((passed + 1)); }

# ---------------------------------------------------------------------------
# Section 1: Library file presence and permissions
# ---------------------------------------------------------------------------

echo ""
echo "1. Library presence and permissions"
echo "-----------------------------------"

if [ ! -f "${TERMINAL_NOTIFY_LIB}" ]; then
  err "harness/scripts/lib/terminal-notify.sh not found"
else
  pass "terminal-notify.sh exists"
fi

if [ -f "${TERMINAL_NOTIFY_LIB}" ] && [ ! -x "${TERMINAL_NOTIFY_LIB}" ]; then
  err "terminal-notify.sh is not executable"
else
  [ -f "${TERMINAL_NOTIFY_LIB}" ] && pass "terminal-notify.sh is executable"
fi

# ---------------------------------------------------------------------------
# Section 2: Syntax check
# ---------------------------------------------------------------------------

echo ""
echo "2. Syntax check"
echo "---------------"

if bash -n "${TERMINAL_NOTIFY_LIB}" 2>/dev/null; then
  pass "terminal-notify.sh syntax OK"
else
  err "terminal-notify.sh has syntax errors"
fi

if bash -n "${NOTIFICATION_HANDLER}" 2>/dev/null; then
  pass "notification-handler.sh syntax OK"
else
  err "notification-handler.sh has syntax errors"
fi

if bash -n "${WEBHOOK_NOTIFY}" 2>/dev/null; then
  pass "webhook-notify.sh syntax OK"
else
  err "webhook-notify.sh has syntax errors"
fi

# ---------------------------------------------------------------------------
# Section 3: Required function exports
# ---------------------------------------------------------------------------

echo ""
echo "3. Required functions defined in library"
echo "----------------------------------------"

for fn in build_terminal_sequence encode_terminal_sequence_json; do
  if bash -c "source '${TERMINAL_NOTIFY_LIB}'; declare -f ${fn}" >/dev/null 2>&1; then
    pass "function ${fn} defined"
  else
    err "function ${fn} missing from terminal-notify.sh"
  fi
done

# ---------------------------------------------------------------------------
# Section 4: HARNESS_TERMINAL_NOTIFY mode behavior
# ---------------------------------------------------------------------------

echo ""
echo "4. Mode behavior"
echo "----------------"

_run_seq() {
  # Usage: _run_seq <mode> <title> <body>
  local mode="$1" title="$2" body="$3"
  HARNESS_TERMINAL_NOTIFY="${mode}" bash -c \
    "source '${TERMINAL_NOTIFY_LIB}'; build_terminal_sequence \"\$1\" \"\$2\"" -- "${title}" "${body}"
}

# Off modes: no output
for mode in "" "0"; do
  out="$(_run_seq "${mode}" "Hello" "World" || true)"
  if [ -z "${out}" ]; then
    pass "mode='${mode}' → no output (correct)"
  else
    err "mode='${mode}' should produce no output, got $(printf '%s' "${out}" | xxd | head -1)"
  fi
done

# Bell mode: single BEL byte
for mode in "1" "bell"; do
  out_hex="$(HARNESS_TERMINAL_NOTIFY="${mode}" bash -c \
    "source '${TERMINAL_NOTIFY_LIB}'; build_terminal_sequence '' ''" | xxd -p | tr -d '\n' || true)"
  if [ "${out_hex}" = "07" ]; then
    pass "mode='${mode}' → BEL (0x07) correct"
  else
    err "mode='${mode}' expected BEL (07) got '${out_hex}'"
  fi
done

# Title mode: ESC ]0;<title> BEL
out_hex="$(_run_seq "title" "My Title" "" | xxd -p | tr -d '\n' || true)"
# ESC=1b, ]=5d, 0=30, ;=3b, "My Title" in hex, BEL=07
if printf '%s' "${out_hex}" | grep -q '^1b5d30'; then
  pass "mode=title → starts with ESC ]0; (correct)"
else
  err "mode=title wrong prefix: ${out_hex}"
fi
if printf '%s' "${out_hex}" | grep -q '07$'; then
  pass "mode=title → ends with BEL (correct)"
else
  err "mode=title does not end with BEL: ${out_hex}"
fi

# OSC9 mode: ESC ]9;<title> BEL
out_hex="$(_run_seq "osc9" "Notify Me" "" | xxd -p | tr -d '\n' || true)"
if printf '%s' "${out_hex}" | grep -q '^1b5d39'; then
  pass "mode=osc9 → starts with ESC ]9; (correct)"
else
  err "mode=osc9 wrong prefix: ${out_hex}"
fi

# Notify mode: ESC ]777;notify;<title> BEL (or with body)
out_hex="$(_run_seq "notify" "Title" "Body" | xxd -p | tr -d '\n' || true)"
if printf '%s' "${out_hex}" | grep -q '^1b5d373737'; then
  pass "mode=notify → starts with ESC ]777 (correct)"
else
  err "mode=notify wrong prefix: ${out_hex}"
fi

# Unknown mode: no output
out="$(_run_seq "invalid_mode" "Test" "" || true)"
if [ -z "${out}" ]; then
  pass "unknown mode → no output (correct)"
else
  err "unknown mode should produce no output"
fi

# ---------------------------------------------------------------------------
# Section 5: Title required (non-bell modes skip empty title)
# ---------------------------------------------------------------------------

echo ""
echo "5. Empty-title behavior"
echo "-----------------------"

for mode in "title" "osc9" "notify"; do
  out="$(_run_seq "${mode}" "" "" || true)"
  if [ -z "${out}" ]; then
    pass "mode=${mode} with empty title → no output (correct)"
  else
    err "mode=${mode} with empty title should produce no output"
  fi
done

# Bell with empty title still fires
out_hex="$(HARNESS_TERMINAL_NOTIFY=bell bash -c \
  "source '${TERMINAL_NOTIFY_LIB}'; build_terminal_sequence '' ''" | xxd -p | tr -d '\n' || true)"
if [ "${out_hex}" = "07" ]; then
  pass "bell with empty title → BEL still fires (correct)"
else
  err "bell with empty title should still fire BEL"
fi

# ---------------------------------------------------------------------------
# Section 6: Control-character sanitization
# ---------------------------------------------------------------------------

echo ""
echo "6. Control-character sanitization"
echo "----------------------------------"

# ESC injected into title must be stripped
out_hex="$(_run_seq "osc9" $'\x1bINJECTED' "" | xxd -p | tr -d '\n' || true)"
# Should NOT contain a second ESC (the title ESC should be stripped)
# The output starts with the OSC9 ESC, followed by ]9; and the sanitized title
# If the title ESC was kept, we'd see 1b1b or 1b5d391bINJECTED early on
injected_count="$(printf '%s' "${out_hex}" | grep -o '1b' | wc -l | tr -d ' ')"
if [ "${injected_count}" -eq 1 ]; then
  pass "ESC in title stripped — only 1 ESC in output (the OSC prefix)"
else
  err "ESC in title not stripped: ${out_hex} (expected exactly 1 ESC byte)"
fi

# BEL injected into title must be stripped
out_hex="$(_run_seq "osc9" $'\x07BELTITLE' "" | xxd -p | tr -d '\n' || true)"
# Count BEL bytes — should be exactly 1 (the terminating BEL), not 2
bel_count="$(printf '%s' "${out_hex}" | grep -o '07' | wc -l | tr -d ' ')"
if [ "${bel_count}" -eq 1 ]; then
  pass "BEL in title stripped — only 1 BEL in output (the terminator)"
else
  err "BEL in title not stripped (found ${bel_count} BEL bytes): ${out_hex}"
fi

# ---------------------------------------------------------------------------
# Section 7: Non-ASCII passthrough
# ---------------------------------------------------------------------------

echo ""
echo "7. Non-ASCII passthrough"
echo "------------------------"

out="$(_run_seq "osc9" "Café résumé" "" || true)"
if printf '%s' "${out}" | grep -Fq "Café"; then
  pass "non-ASCII characters pass through title unchanged"
else
  err "non-ASCII characters stripped from title (should be preserved)"
fi

# ---------------------------------------------------------------------------
# Section 8: encode_terminal_sequence_json
# ---------------------------------------------------------------------------

echo ""
echo "8. encode_terminal_sequence_json"
echo "---------------------------------"

# Verify encoding produces valid JSON string (with quotes)
if command -v jq >/dev/null 2>&1; then
  encoded="$(HARNESS_TERMINAL_NOTIFY=osc9 bash -c "
    source '${TERMINAL_NOTIFY_LIB}'
    seq=\$(build_terminal_sequence 'Test' '')
    encode_terminal_sequence_json \"\$seq\"
  " 2>/dev/null || true)"

  if [ -n "${encoded}" ]; then
    # Must start and end with double-quote (JSON string)
    if printf '%s' "${encoded}" | grep -q '^".*"$'; then
      pass "encode_terminal_sequence_json returns quoted JSON string"
    else
      err "encode_terminal_sequence_json output not a JSON string: ${encoded}"
    fi
    # Decoded value must contain the actual ESC byte (jq uses )
    decoded="$(printf '%s' "${encoded}" | jq -r . 2>/dev/null || true)"
    if printf '%s' "${decoded}" | xxd -p | grep -q '1b'; then
      pass "encoded JSON decodes back to OSC sequence (contains ESC byte)"
    else
      err "decoded terminalSequence missing ESC byte"
    fi
  else
    err "encode_terminal_sequence_json returned empty (jq available but encode failed)"
  fi
else
  pass "jq not available — skip encode_terminal_sequence_json round-trip test"
fi

# Empty input → no output (not empty string, but no bytes at all)
empty_out="$(bash -c "source '${TERMINAL_NOTIFY_LIB}'; encode_terminal_sequence_json ''" 2>/dev/null || true)"
if [ -z "${empty_out}" ]; then
  pass "encode_terminal_sequence_json('') → no output (correct)"
else
  err "encode_terminal_sequence_json('') should produce no output, got: ${empty_out}"
fi

# ---------------------------------------------------------------------------
# Section 9: notification-handler.sh integration
# ---------------------------------------------------------------------------

echo ""
echo "9. notification-handler.sh integration"
echo "---------------------------------------"

# Verify sources the lib
if grep -Fq 'terminal-notify.sh' "${NOTIFICATION_HANDLER}"; then
  pass "notification-handler.sh sources terminal-notify.sh"
else
  err "notification-handler.sh does not source terminal-notify.sh"
fi

# Known types must use build_terminal_sequence
if grep -Fq 'build_terminal_sequence' "${NOTIFICATION_HANDLER}"; then
  pass "notification-handler.sh calls build_terminal_sequence"
else
  err "notification-handler.sh missing build_terminal_sequence call"
fi

# Verify terminalSequence JSON output
if grep -Fq 'terminalSequence' "${NOTIFICATION_HANDLER}"; then
  pass "notification-handler.sh emits terminalSequence field"
else
  err "notification-handler.sh missing terminalSequence in output"
fi

# End-to-end: permission_prompt with osc9 → valid JSON with terminalSequence
if python3 -c "import sys" 2>/dev/null; then
  tmpdir="$(mktemp -d)"
  raw_out="$(PROJECT_ROOT="${tmpdir}" HARNESS_TERMINAL_NOTIFY=osc9 \
    bash "${NOTIFICATION_HANDLER}" \
    <<< '{"notification_type":"permission_prompt","session_id":"s1","agent_type":"worker"}' \
    2>/dev/null || true)"
  rm -rf "${tmpdir}"

  if [ -n "${raw_out}" ]; then
    ts_present="$(printf '%s' "${raw_out}" | python3 -c "
import sys, json
d = json.loads(sys.stdin.buffer.read().decode('utf-8'))
print('yes' if 'terminalSequence' in d else 'no')
" 2>/dev/null || echo 'parse_fail')"
    if [ "${ts_present}" = "yes" ]; then
      pass "notification-handler.sh: permission_prompt + osc9 → terminalSequence in JSON"
    else
      err "notification-handler.sh: terminalSequence missing in JSON (got: ${ts_present})"
    fi
  else
    err "notification-handler.sh produced no output for permission_prompt"
  fi

  # Unknown type → no terminalSequence (output is empty — no JSON at all)
  tmpdir="$(mktemp -d)"
  raw_out2="$(PROJECT_ROOT="${tmpdir}" HARNESS_TERMINAL_NOTIFY=osc9 \
    bash "${NOTIFICATION_HANDLER}" \
    <<< '{"notification_type":"unknown_xyz","session_id":"s1"}' \
    2>/dev/null || true)"
  rm -rf "${tmpdir}"

  if [ -z "${raw_out2}" ]; then
    pass "notification-handler.sh: unknown type → no JSON output (correct)"
  else
    err "notification-handler.sh: unknown type should not emit JSON: ${raw_out2}"
  fi
fi

# ---------------------------------------------------------------------------
# Section 10: webhook-notify.sh integration
# ---------------------------------------------------------------------------

echo ""
echo "10. webhook-notify.sh integration"
echo "----------------------------------"

if grep -Fq 'terminal-notify.sh' "${WEBHOOK_NOTIFY}"; then
  pass "webhook-notify.sh sources terminal-notify.sh"
else
  err "webhook-notify.sh does not source terminal-notify.sh"
fi

if grep -Fq '_render_terminal_sequence_field' "${WEBHOOK_NOTIFY}"; then
  pass "webhook-notify.sh defines _render_terminal_sequence_field helper"
else
  err "webhook-notify.sh missing _render_terminal_sequence_field"
fi

# No-URL path + osc9 → terminalSequence present
if python3 -c "import sys" 2>/dev/null; then
  raw_out="$(HARNESS_TERMINAL_NOTIFY=osc9 HARNESS_WEBHOOK_URL="" \
    bash "${WEBHOOK_NOTIFY}" "task-completed" \
    <<< '{}' 2>/dev/null || true)"

  if [ -n "${raw_out}" ]; then
    ts_present="$(printf '%s' "${raw_out}" | python3 -c "
import sys, json
d = json.loads(sys.stdin.buffer.read().decode('utf-8'))
print('yes' if 'terminalSequence' in d else 'no')
" 2>/dev/null || echo 'parse_fail')"
    if [ "${ts_present}" = "yes" ]; then
      pass "webhook-notify.sh: no-URL + osc9 → terminalSequence in JSON"
    else
      err "webhook-notify.sh: terminalSequence missing when URL unset + osc9 (got: ${ts_present})"
    fi
  else
    err "webhook-notify.sh produced no output"
  fi

  # No terminal notify env → no terminalSequence
  raw_out2="$(HARNESS_TERMINAL_NOTIFY="" HARNESS_WEBHOOK_URL="" \
    bash "${WEBHOOK_NOTIFY}" "task-completed" \
    <<< '{}' 2>/dev/null || true)"

  if [ -n "${raw_out2}" ]; then
    ts_present2="$(printf '%s' "${raw_out2}" | python3 -c "
import sys, json
d = json.loads(sys.stdin.buffer.read().decode('utf-8'))
print('yes' if 'terminalSequence' in d else 'no')
" 2>/dev/null || echo 'parse_fail')"
    if [ "${ts_present2}" = "no" ]; then
      pass "webhook-notify.sh: no env → no terminalSequence (correct)"
    else
      err "webhook-notify.sh: terminalSequence should be absent when env unset"
    fi
  fi
fi

# ---------------------------------------------------------------------------
# Section 11: No Japanese in shell files
# ---------------------------------------------------------------------------

echo ""
echo "11. No Japanese in shell files"
echo "------------------------------"

if python3 -c "import sys" 2>/dev/null; then
  for f in "${TERMINAL_NOTIFY_LIB}" "${NOTIFICATION_HANDLER}" "${WEBHOOK_NOTIFY}"; do
    if python3 -c "
import sys, re
data = open(sys.argv[1]).read()
sys.exit(0 if re.search(r'[぀-ヿ一-鿿]', data) else 1)
" "$f" 2>/dev/null; then
      err "Japanese characters found in ${f##*/}"
    else
      pass "no Japanese in ${f##*/}"
    fi
  done
else
  pass "python3 not available — skip Japanese check"
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

echo ""
echo "=========================================="
echo "Results: ${passed} passed, ${failures} failed"
echo "=========================================="

if [ "${failures}" -gt 0 ]; then
  exit 1
fi

echo "test-terminal-notify: ok"
