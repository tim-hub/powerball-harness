#!/bin/bash
# webhook-notify.sh
# POSTs to an external webhook only when HARNESS_WEBHOOK_URL is set.
# Environment variables are not expanded in the url field of HTTP hooks,
# so this is implemented as a command hook using curl.
#
# Usage: bash webhook-notify.sh <event-name>
# Input: stdin JSON from Claude Code hooks
# Env: HARNESS_WEBHOOK_URL (optional, skip if unset)

set -euo pipefail

EVENT_NAME="${1:-unknown}"

# Terminal sequence helper (CC 2.1.141+, opt-in via HARNESS_TERMINAL_NOTIFY)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
if [ -f "${PARENT_DIR}/lib/terminal-notify.sh" ]; then
  # shellcheck disable=SC1091
  source "${PARENT_DIR}/lib/terminal-notify.sh"
fi

# Render the terminalSequence JSON field (empty string when not opted in).
# Returns a comma-prefixed JSON field fragment, e.g. ,"terminalSequence":"..."
_render_terminal_sequence_field() {
  local title="${1:-}" body="${2:-}"
  if ! declare -f build_terminal_sequence >/dev/null 2>&1; then
    return 0
  fi
  local seq encoded
  seq="$(build_terminal_sequence "${title}" "${body}")"
  [ -n "${seq}" ] || return 0
  encoded="$(encode_terminal_sequence_json "${seq}")"
  printf ',"terminalSequence":%s' "${encoded}"
}

# Exit without doing anything if HARNESS_WEBHOOK_URL is not set (opt-in).
# Still emit terminalSequence when HARNESS_TERMINAL_NOTIFY is configured.
if [ -z "${HARNESS_WEBHOOK_URL:-}" ]; then
  _TS_FIELD="$(_render_terminal_sequence_field "Claude Code: ${EVENT_NAME}" "")"
  printf '{"decision":"approve","reason":"webhook URL not configured, skipping"%s}\n' "${_TS_FIELD}"
  exit 0
fi

# Read hook payload from stdin
PAYLOAD=""
if [ ! -t 0 ]; then
  PAYLOAD=$(cat)
fi

# Mask the URL (secret protection: show scheme only)
# Hides user:pass@host, ?token=xxx, /services/T00/B00/xxx, etc.
MASKED_URL="$(echo "${HARNESS_WEBHOOK_URL}" | sed -E 's|^(https?://).*|\1***/***|')"

# POST via curl (5-second timeout; approve and continue on failure, but report the result)
HTTP_CODE=""
CURL_EXIT=0
HTTP_CODE=$(curl --silent --output /dev/null --write-out "%{http_code}" --max-time 5 \
  --request POST \
  --header "Content-Type: application/json" \
  --header "X-Harness-Event: ${EVENT_NAME}" \
  --data "${PAYLOAD:-"{}"}" \
  "${HARNESS_WEBHOOK_URL}" 2>/dev/null) || CURL_EXIT=$?

# Compute terminal sequence fields once for both success and failure paths.
# MASKED_URL is already normalized to "https://***/***/", so values are JSON-safe.
_TS_FIELD_SUCCESS="$(_render_terminal_sequence_field "Claude Code: ${EVENT_NAME}" "webhook notified")"
_TS_FIELD_FAILURE="$(_render_terminal_sequence_field "Claude Code: ${EVENT_NAME}" "webhook failed")"

# Build response JSON (jq used for correct encoding when available; fallback uses safe literals)
if [ "$CURL_EXIT" -ne 0 ]; then
  if command -v jq >/dev/null 2>&1; then
    _base="$(jq -nc --arg reason "webhook delivery failed (curl exit $CURL_EXIT)" \
               --arg msg "[webhook-notify] POST to ${MASKED_URL} failed (curl exit $CURL_EXIT)" \
               '{"decision":"approve","reason":$reason,"systemMessage":$msg}')"
    printf '%s\n' "${_base%\}}${_TS_FIELD_FAILURE}}"
  else
    printf '{"decision":"approve","reason":"webhook delivery failed","systemMessage":"[webhook-notify] POST failed"%s}\n' \
      "${_TS_FIELD_FAILURE}"
  fi
elif [ "${HTTP_CODE:-000}" -ge 200 ] && [ "${HTTP_CODE:-000}" -lt 300 ] 2>/dev/null; then
  printf '{"decision":"approve","reason":"webhook notification sent"%s}\n' "${_TS_FIELD_SUCCESS}"
else
  if command -v jq >/dev/null 2>&1; then
    _base="$(jq -nc --arg reason "webhook returned HTTP ${HTTP_CODE}" \
               --arg msg "[webhook-notify] POST to ${MASKED_URL} returned HTTP ${HTTP_CODE}" \
               '{"decision":"approve","reason":$reason,"systemMessage":$msg}')"
    printf '%s\n' "${_base%\}}${_TS_FIELD_FAILURE}}"
  else
    printf '{"decision":"approve","reason":"webhook returned HTTP %s","systemMessage":"[webhook-notify] POST returned HTTP %s"%s}\n' \
      "${HTTP_CODE}" "${HTTP_CODE}" "${_TS_FIELD_FAILURE}"
  fi
fi
