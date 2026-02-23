#!/bin/bash
# ci-status-checker.sh
# PostToolUse (Bash matcher) で git push / gh pr 後の CI ステータスを非同期チェック
# CI 失敗検知時に additionalContext で ci-cd-fixer の spawn を推奨するメッセージを注入
#
# Input: stdin JSON from Claude Code hooks (PostToolUse/Bash)
# Output: JSON to approve the event (with optional additionalContext)

set +e  # エラーで停止しない

# === stdin から JSON ペイロードを読み取り ===
INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null)"
fi

# ペイロードが空の場合はスキップ
if [ -z "${INPUT}" ]; then
  echo '{"decision":"approve","reason":"ci-status-checker: no payload"}'
  exit 0
fi

# === Bash ツールの出力からコマンドと終了コードを取得 ===
TOOL_NAME=""
BASH_CMD=""
BASH_EXIT_CODE=""
BASH_OUTPUT=""

if command -v jq >/dev/null 2>&1; then
  _parsed="$(printf '%s' "${INPUT}" | jq -r '[
    (.tool_name // ""),
    (.tool_input.command // ""),
    ((.tool_response.exit_code // .tool_response.exitCode // -1) | tostring),
    ((.tool_response.output // .tool_response.stdout // "") | .[0:500])
  ] | @tsv' 2>/dev/null)"
  if [ -n "${_parsed}" ]; then
    IFS=$'\t' read -r TOOL_NAME BASH_CMD BASH_EXIT_CODE BASH_OUTPUT <<< "${_parsed}"
  fi
elif command -v python3 >/dev/null 2>&1; then
  _parsed="$(printf '%s' "${INPUT}" | python3 -c "
import json, sys
try:
    d = json.load(sys.stdin)
    tr = d.get('tool_response', {})
    ti = d.get('tool_input', {})
    print(d.get('tool_name', ''))
    print(ti.get('command', ''))
    print(str(tr.get('exit_code', tr.get('exitCode', -1))))
    out = tr.get('output', tr.get('stdout', ''))
    print(str(out)[:500])
except:
    print('')
    print('')
    print('-1')
    print('')
" 2>/dev/null)"
  TOOL_NAME="$(echo "${_parsed}" | sed -n '1p')"
  BASH_CMD="$(echo "${_parsed}" | sed -n '2p')"
  BASH_EXIT_CODE="$(echo "${_parsed}" | sed -n '3p')"
  BASH_OUTPUT="$(echo "${_parsed}" | sed -n '4p')"
fi

# === git push / gh pr コマンドかどうか判定 ===
is_push_or_pr_command() {
  local cmd="$1"
  # git push / gh pr create / gh pr merge / gh workflow run などを検知
  if echo "${cmd}" | grep -Eq '(^|[[:space:]])(git\s+push|gh\s+pr\s+(create|merge|edit)|gh\s+workflow\s+run)'; then
    return 0
  fi
  return 1
}

if ! is_push_or_pr_command "${BASH_CMD}"; then
  echo '{"decision":"approve","reason":"ci-status-checker: not a push/PR command"}'
  exit 0
fi

# === プロジェクトルートを検出 ===
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
if [ -f "${PARENT_DIR}/path-utils.sh" ]; then
  source "${PARENT_DIR}/path-utils.sh" 2>/dev/null || true
fi
PROJECT_ROOT="${PROJECT_ROOT:-$(detect_project_root 2>/dev/null || pwd)}"
STATE_DIR="${PROJECT_ROOT}/.claude/state"
mkdir -p "${STATE_DIR}" 2>/dev/null || true

# === 非同期で CI ステータスを確認（バックグラウンドジョブ）===
# CI チェックは最大 60 秒間ポーリング（gh コマンドが存在する場合のみ）
CI_STATUS_FILE="${STATE_DIR}/ci-status.json"
TS="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

check_ci_status_async() {
  if ! command -v gh >/dev/null 2>&1; then
    return
  fi

  local max_wait=60
  local poll_interval=10
  local elapsed=0
  local status="unknown"
  local conclusion="unknown"

  while [ "${elapsed}" -lt "${max_wait}" ]; do
    sleep "${poll_interval}"
    elapsed=$(( elapsed + poll_interval ))

    # 最新の PR チェックを取得
    local runs_json
    runs_json="$(gh run list --limit 1 --json status,conclusion,name,url 2>/dev/null)" || runs_json=""
    if [ -z "${runs_json}" ]; then
      continue
    fi

    if command -v jq >/dev/null 2>&1; then
      status="$(printf '%s' "${runs_json}" | jq -r '.[0].status // "unknown"' 2>/dev/null)" || status="unknown"
      conclusion="$(printf '%s' "${runs_json}" | jq -r '.[0].conclusion // "unknown"' 2>/dev/null)" || conclusion="unknown"
    fi

    # completed 以外はまだ実行中
    if [ "${status}" != "completed" ]; then
      continue
    fi

    # 結果を記録
    if command -v jq >/dev/null 2>&1; then
      jq -n \
        --arg ts "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
        --arg trigger_cmd "${BASH_CMD}" \
        --arg status "${status}" \
        --arg conclusion "${conclusion}" \
        '{timestamp:$ts, trigger_command:$trigger_cmd, status:$status, conclusion:$conclusion}' \
        > "${CI_STATUS_FILE}" 2>/dev/null || true
    fi

    # CI 失敗の場合はシグナルファイルを書き出す
    if [ "${conclusion}" = "failure" ] || [ "${conclusion}" = "timed_out" ] || [ "${conclusion}" = "cancelled" ]; then
      SIGNALS_FILE="${STATE_DIR}/breezing-signals.jsonl"
      if command -v jq >/dev/null 2>&1; then
        jq -nc \
          --arg signal "ci_failure_detected" \
          --arg timestamp "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" \
          --arg conclusion "${conclusion}" \
          --arg trigger_cmd "${BASH_CMD}" \
          '{signal:$signal, timestamp:$timestamp, conclusion:$conclusion, trigger_command:$trigger_cmd}' \
          >> "${SIGNALS_FILE}" 2>/dev/null || true
      fi
    fi

    return
  done
}

# バックグラウンドで CI チェックを実行（フックをブロックしない）
check_ci_status_async &
disown 2>/dev/null || true

# === 直近の CI 失敗シグナルをチェックして additionalContext を注入 ===
ADDITIONAL_CONTEXT=""
SIGNALS_FILE="${STATE_DIR}/breezing-signals.jsonl"

if [ -f "${SIGNALS_FILE}" ]; then
  # 直近1件の ci_failure_detected シグナルを取得（10分以内のもの）
  _recent_failure=""
  if command -v jq >/dev/null 2>&1; then
    _recent_failure="$(grep '"ci_failure_detected"' "${SIGNALS_FILE}" 2>/dev/null | tail -1)" || _recent_failure=""
  fi

  if [ -n "${_recent_failure}" ]; then
    _failure_conclusion=""
    if command -v jq >/dev/null 2>&1; then
      _failure_conclusion="$(printf '%s' "${_recent_failure}" | jq -r '.conclusion // ""' 2>/dev/null)" || _failure_conclusion=""
    fi

    ADDITIONAL_CONTEXT="[CI 失敗を検知しました]\nCI ステータス: ${_failure_conclusion}\nトリガーコマンド: ${BASH_CMD}\n\n推奨アクション: /breezing または ci-cd-fixer エージェントを spawn して CI 障害を自動修復してください。\n  例: ci-cd-fixer に「CI が失敗しました。ログを確認して修正してください。」と依頼"
  fi
fi

# === レスポンス ===
if [ -n "${ADDITIONAL_CONTEXT}" ]; then
  if command -v jq >/dev/null 2>&1; then
    jq -nc \
      --arg reason "ci-status-checker: push/PR detected, CI failure context injected" \
      --arg ctx "${ADDITIONAL_CONTEXT}" \
      '{"decision":"approve","reason":$reason,"additionalContext":$ctx}'
  else
    echo '{"decision":"approve","reason":"ci-status-checker: push/PR detected, CI failure context injected"}'
  fi
else
  echo '{"decision":"approve","reason":"ci-status-checker: push/PR detected, CI monitoring started"}'
fi
exit 0
