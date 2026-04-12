#!/bin/bash
# webhook-notify.sh
# POST to external webhook only when HARNESS_WEBHOOK_URL is set
# Environment variables are not expanded in the HTTP hook url field,
# so this is implemented as a command hook + curl
#
# Usage: bash webhook-notify.sh <event-name>
# Input: stdin JSON from Claude Code hooks
# Env: HARNESS_WEBHOOK_URL (optional, skip if unset)

set -euo pipefail

EVENT_NAME="${1:-unknown}"

# Exit silently if HARNESS_WEBHOOK_URL is not set (opt-in)
if [ -z "${HARNESS_WEBHOOK_URL:-}" ]; then
  echo '{"decision":"approve","reason":"webhook URL not configured, skipping"}'
  exit 0
fi

# Read hook payload from stdin
PAYLOAD=""
if [ ! -t 0 ]; then
  PAYLOAD=$(cat)
fi

# Mask URL (secret protection: show scheme only)
# Hide user:pass@host, ?token=xxx, /services/T00/B00/xxx etc.
MASKED_URL="$(echo "${HARNESS_WEBHOOK_URL}" | sed -E 's|^(https?://).*|\1***/***|')"

# POST via curl (5-second timeout, continue with approve on failure but report result)
HTTP_CODE=""
CURL_EXIT=0
HTTP_CODE=$(curl --silent --output /dev/null --write-out "%{http_code}" --max-time 5 \
  --request POST \
  --header "Content-Type: application/json" \
  --header "X-Harness-Event: ${EVENT_NAME}" \
  --data "${PAYLOAD:-"{}"}" \
  "${HARNESS_WEBHOOK_URL}" 2>/dev/null) || CURL_EXIT=$?

# Build JSON safely with jq if available, otherwise use a fixed message
if [ "$CURL_EXIT" -ne 0 ]; then
  if command -v jq >/dev/null 2>&1; then
    jq -nc --arg reason "webhook delivery failed (curl exit $CURL_EXIT)" \
           --arg msg "[webhook-notify] POST to ${MASKED_URL} failed (curl exit $CURL_EXIT)" \
           '{"decision":"approve","reason":$reason,"systemMessage":$msg}'
  else
    echo "{\"decision\":\"approve\",\"reason\":\"webhook delivery failed\",\"systemMessage\":\"[webhook-notify] POST failed\"}"
  fi
elif [ "${HTTP_CODE:-000}" -ge 200 ] && [ "${HTTP_CODE:-000}" -lt 300 ] 2>/dev/null; then
  echo '{"decision":"approve","reason":"webhook notification sent"}'
else
  if command -v jq >/dev/null 2>&1; then
    jq -nc --arg reason "webhook returned HTTP ${HTTP_CODE}" \
           --arg msg "[webhook-notify] POST to ${MASKED_URL} returned HTTP ${HTTP_CODE}" \
           '{"decision":"approve","reason":$reason,"systemMessage":$msg}'
  else
    echo "{\"decision\":\"approve\",\"reason\":\"webhook returned HTTP ${HTTP_CODE}\",\"systemMessage\":\"[webhook-notify] POST returned HTTP ${HTTP_CODE}\"}"
  fi
fi
