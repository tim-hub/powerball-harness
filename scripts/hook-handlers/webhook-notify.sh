#!/bin/bash
# webhook-notify.sh
# HARNESS_WEBHOOK_URL が設定されている場合のみ外部 webhook に POST する
# HTTP hook の url フィールドでは環境変数が展開されないため、
# command hook + curl で実装している
#
# Usage: bash webhook-notify.sh <event-name>
# Input: stdin JSON from Claude Code hooks
# Env: HARNESS_WEBHOOK_URL (optional, skip if unset)

set -euo pipefail

EVENT_NAME="${1:-unknown}"

# HARNESS_WEBHOOK_URL が未設定なら何もせず終了（opt-in）
if [ -z "${HARNESS_WEBHOOK_URL:-}" ]; then
  echo '{"decision":"approve","reason":"webhook URL not configured, skipping"}'
  exit 0
fi

# stdin から hook payload を読み取る
PAYLOAD=""
if [ ! -t 0 ]; then
  PAYLOAD=$(cat)
fi

# URL をマスク（シークレット保護: スキームのみ表示）
# user:pass@host, ?token=xxx, /services/T00/B00/xxx 等を全て隠す
MASKED_URL="$(echo "${HARNESS_WEBHOOK_URL}" | sed -E 's|^(https?://).*|\1***/***|')"

# curl で POST（タイムアウト 5 秒、失敗しても approve で続行だが結果を報告）
HTTP_CODE=""
CURL_EXIT=0
HTTP_CODE=$(curl --silent --output /dev/null --write-out "%{http_code}" --max-time 5 \
  --request POST \
  --header "Content-Type: application/json" \
  --header "X-Harness-Event: ${EVENT_NAME}" \
  --data "${PAYLOAD:-"{}"}" \
  "${HARNESS_WEBHOOK_URL}" 2>/dev/null) || CURL_EXIT=$?

# jq があれば安全に JSON を構築、なければ固定メッセージ
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
