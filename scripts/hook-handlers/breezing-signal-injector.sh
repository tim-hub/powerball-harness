#!/bin/bash
# breezing-signal-injector.sh
# UserPromptSubmit フックで breezing-signals.jsonl から未消費シグナルを読み取り、
# systemMessage として注入する。
#
# Usage: 自動呼び出し（UserPromptSubmit hook）
# Input: stdin JSON from Claude Code hooks (UserPromptSubmit)
# Output: JSON with optional systemMessage

set +e  # エラーで停止しない

# === プロジェクトルートを検出 ===
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
if [ -f "${PARENT_DIR}/path-utils.sh" ]; then
  source "${PARENT_DIR}/path-utils.sh" 2>/dev/null || true
fi
PROJECT_ROOT="${PROJECT_ROOT:-$(cd "${PARENT_DIR}/.." && pwd)}"
STATE_DIR="${PROJECT_ROOT}/.claude/state"

# === breezing セッションが存在するかチェック ===
ACTIVE_FILE="${STATE_DIR}/breezing-active.json"
if [ ! -f "${ACTIVE_FILE}" ]; then
  # breezing セッション外はスキップ
  exit 0
fi

# === シグナルファイルが存在するかチェック ===
SIGNALS_FILE="${STATE_DIR}/breezing-signals.jsonl"
if [ ! -f "${SIGNALS_FILE}" ]; then
  exit 0
fi

# === 未消費シグナルを読み取り ===
# consumed_at が null または存在しない行を未消費とみなす
UNCONSUMED_SIGNALS=""
if command -v jq >/dev/null 2>&1; then
  # jq で consumed_at が null のシグナルを抽出
  UNCONSUMED_SIGNALS="$(grep -v '^$' "${SIGNALS_FILE}" 2>/dev/null | \
    while IFS= read -r line; do
      consumed="$(printf '%s' "${line}" | jq -r '.consumed_at // "null"' 2>/dev/null)"
      if [ "${consumed}" = "null" ]; then
        printf '%s\n' "${line}"
      fi
    done)" || UNCONSUMED_SIGNALS=""
elif command -v python3 >/dev/null 2>&1; then
  UNCONSUMED_SIGNALS="$(python3 -c "
import json, sys
lines = []
try:
    with open('${SIGNALS_FILE}', 'r') as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                d = json.loads(line)
                if d.get('consumed_at') is None:
                    lines.append(line)
            except:
                pass
print('\n'.join(lines))
" 2>/dev/null)" || UNCONSUMED_SIGNALS=""
fi

if [ -z "${UNCONSUMED_SIGNALS}" ]; then
  # 未消費シグナルなし
  exit 0
fi

# === シグナルをメッセージ形式に整形 ===
SYSTEM_MESSAGE=""
SIGNAL_COUNT=0

while IFS= read -r signal_line; do
  [ -z "${signal_line}" ] && continue

  SIGNAL_COUNT=$((SIGNAL_COUNT + 1))
  signal_type=""
  signal_ts=""

  if command -v jq >/dev/null 2>&1; then
    signal_type="$(printf '%s' "${signal_line}" | jq -r '.signal // .type // "unknown"' 2>/dev/null)" || signal_type="unknown"
    signal_ts="$(printf '%s' "${signal_line}" | jq -r '.timestamp // ""' 2>/dev/null)" || signal_ts=""
  fi

  case "${signal_type}" in
    ci_failure_detected)
      conclusion=""
      trigger_cmd=""
      if command -v jq >/dev/null 2>&1; then
        conclusion="$(printf '%s' "${signal_line}" | jq -r '.conclusion // "unknown"' 2>/dev/null)"
        trigger_cmd="$(printf '%s' "${signal_line}" | jq -r '.trigger_command // ""' 2>/dev/null)"
      fi
      SYSTEM_MESSAGE="${SYSTEM_MESSAGE}[SIGNAL:ci_failure_detected] CI が失敗しました（${conclusion}）。トリガー: ${trigger_cmd}。ci-cd-fixer エージェントで自動修復することを検討してください。\n"
      ;;
    retake_requested)
      reason=""
      task_id=""
      if command -v jq >/dev/null 2>&1; then
        reason="$(printf '%s' "${signal_line}" | jq -r '.reason // ""' 2>/dev/null)"
        task_id="$(printf '%s' "${signal_line}" | jq -r '.task_id // ""' 2>/dev/null)"
      fi
      SYSTEM_MESSAGE="${SYSTEM_MESSAGE}[SIGNAL:retake_requested] タスク #${task_id} のやり直しが要求されました。理由: ${reason}\n"
      ;;
    reviewer_approved)
      task_id=""
      if command -v jq >/dev/null 2>&1; then
        task_id="$(printf '%s' "${signal_line}" | jq -r '.task_id // ""' 2>/dev/null)"
      fi
      SYSTEM_MESSAGE="${SYSTEM_MESSAGE}[SIGNAL:reviewer_approved] タスク #${task_id} がレビュアーに承認されました。\n"
      ;;
    escalation_required)
      reason=""
      task_id=""
      if command -v jq >/dev/null 2>&1; then
        reason="$(printf '%s' "${signal_line}" | jq -r '.reason // ""' 2>/dev/null)"
        task_id="$(printf '%s' "${signal_line}" | jq -r '.task_id // ""' 2>/dev/null)"
      fi
      SYSTEM_MESSAGE="${SYSTEM_MESSAGE}[SIGNAL:escalation_required] タスク #${task_id} でエスカレーションが必要です。理由: ${reason}\n"
      ;;
    *)
      # 未知のシグナルはそのまま通知
      SYSTEM_MESSAGE="${SYSTEM_MESSAGE}[SIGNAL:${signal_type}] ${signal_line}\n"
      ;;
  esac
done <<< "${UNCONSUMED_SIGNALS}"

if [ -z "${SYSTEM_MESSAGE}" ] || [ "${SIGNAL_COUNT}" -eq 0 ]; then
  exit 0
fi

# === consumed_at を設定してシグナルをマーク済みにする ===
# アトミック更新: 新しいファイルに consumed_at を付与して上書き
CONSUMED_TS="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
LOCK_DIR="${STATE_DIR}/.breezing-signals.lock"

_lock_acquired=0
for _i in $(seq 1 20); do
  if mkdir "${LOCK_DIR}" 2>/dev/null; then
    _lock_acquired=1
    break
  fi
  sleep 0.1
done

if [ "${_lock_acquired}" -eq 1 ]; then
  TMP_NEW_SIGNALS="$(mktemp /tmp/breezing-signals-new.XXXXXX)"

  if command -v jq >/dev/null 2>&1; then
    # consumed_at を付与して全シグナルを再書き込み
    while IFS= read -r line; do
      [ -z "${line}" ] && continue
      consumed="$(printf '%s' "${line}" | jq -r '.consumed_at // "null"' 2>/dev/null)"
      if [ "${consumed}" = "null" ]; then
        printf '%s' "${line}" | jq -c --arg ts "${CONSUMED_TS}" '. + {consumed_at: $ts}' 2>/dev/null >> "${TMP_NEW_SIGNALS}" || printf '%s\n' "${line}" >> "${TMP_NEW_SIGNALS}"
      else
        printf '%s\n' "${line}" >> "${TMP_NEW_SIGNALS}"
      fi
    done < "${SIGNALS_FILE}"

    mv "${TMP_NEW_SIGNALS}" "${SIGNALS_FILE}" 2>/dev/null || rm -f "${TMP_NEW_SIGNALS}"
  else
    rm -f "${TMP_NEW_SIGNALS}"
  fi

  rmdir "${LOCK_DIR}" 2>/dev/null || true
fi

# === systemMessage として出力 ===
HEADER="[breezing-signal-injector] ${SIGNAL_COUNT} 件の未消費シグナルがあります:\n"
FULL_MESSAGE="${HEADER}${SYSTEM_MESSAGE}"

if command -v jq >/dev/null 2>&1; then
  jq -nc --arg msg "${FULL_MESSAGE}" '{"systemMessage": $msg}'
else
  # jq がない場合は簡易エスケープ
  _escaped="${FULL_MESSAGE//\\/\\\\}"
  _escaped="${_escaped//\"/\\\"}"
  printf '{"systemMessage":"%s"}\n' "${_escaped}"
fi

exit 0
