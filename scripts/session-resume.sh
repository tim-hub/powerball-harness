#!/bin/bash
# session-resume.sh
# SessionStart Hook (resume): Harness セッション状態の自動復元
#
# Claude Code の /resume コマンド実行時に自動的に呼び出され、
# Harness のセッション状態（session.json, session.events.jsonl）を復元します。
#
# 入力: stdin から JSON（session_id, source などを含む）
# 出力: JSON 形式で hookSpecificOutput.additionalContext に情報を出力

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROGRESS_SNAPSHOT_LIB="${SCRIPT_DIR}/lib/progress-snapshot.sh"
if [ -f "${PROGRESS_SNAPSHOT_LIB}" ]; then
  # shellcheck source=/dev/null
  source "${PROGRESS_SNAPSHOT_LIB}"
fi

# ===== バナー表示 =====
VERSION=$(cat "$SCRIPT_DIR/../VERSION" 2>/dev/null || echo "unknown")
echo -e "\033[0;36m[claude-code-harness v${VERSION}]\033[0m Session resumed" >&2

# ===== stdin から JSON 入力を読み取り =====
INPUT=""
if [ -t 0 ]; then
  :
else
  INPUT=$(cat 2>/dev/null || true)
fi

# ===== Claude Code session_id を取得 =====
CC_SESSION_ID=""
if [ -n "$INPUT" ] && command -v jq >/dev/null 2>&1; then
  CC_SESSION_ID="$(echo "$INPUT" | jq -r '.session_id // empty' 2>/dev/null)"
fi

# ===== Harness 状態ディレクトリ (repo root 基準で統一) =====
REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || REPO_ROOT="$(pwd)"
STATE_DIR="${REPO_ROOT}/.claude/state"
SESSION_FILE="$STATE_DIR/session.json"
EVENT_LOG_FILE="$STATE_DIR/session.events.jsonl"
ARCHIVE_DIR="$STATE_DIR/sessions"
SESSION_MAP_FILE="$STATE_DIR/session-map.json"

mkdir -p "$STATE_DIR" "$ARCHIVE_DIR"

RESUME_CONTEXT_FILE="${STATE_DIR}/memory-resume-context.md"
RESUME_PENDING_FLAG="${STATE_DIR}/.memory-resume-pending"
RESUME_PROCESSING_FLAG="${STATE_DIR}/.memory-resume-processing"
RESUME_MAX_BYTES="${HARNESS_MEM_RESUME_MAX_BYTES:-32768}"
HANDOFF_ARTIFACT_FILE="${STATE_DIR}/handoff-artifact.json"
LEGACY_PRECOMPACT_SNAPSHOT_FILE="${STATE_DIR}/precompact-snapshot.json"

case "$RESUME_MAX_BYTES" in
  ''|*[!0-9]*) RESUME_MAX_BYTES=32768 ;;
esac
if [ "$RESUME_MAX_BYTES" -gt 65536 ]; then
  RESUME_MAX_BYTES=65536
fi
if [ "$RESUME_MAX_BYTES" -lt 4096 ]; then
  RESUME_MAX_BYTES=4096
fi

# ===== 出力メッセージを蓄積 =====
OUTPUT=""
add_line() {
  OUTPUT="${OUTPUT}$1\n"
}

count_matches() {
  local pattern="$1"
  local file="$2"
  local count
  count="$(grep -c "$pattern" "$file" 2>/dev/null || true)"
  printf '%s' "${count:-0}"
}

consume_memory_resume_context() {
  local file="$1"
  local max_bytes="$2"
  local total=0
  local line=""
  local line_bytes=0
  local out=""

  if [ ! -f "$file" ]; then
    return 0
  fi

  while IFS= read -r line || [ -n "$line" ]; do
    line_bytes="$(printf '%s\n' "$line" | wc -c | tr -d '[:space:]')"
    case "$line_bytes" in
      ''|*[!0-9]*) line_bytes=0 ;;
    esac
    if [ $((total + line_bytes)) -gt "$max_bytes" ]; then
      break
    fi
    out="${out}${line}
"
    total=$((total + line_bytes))
  done < "$file"

  rm -f "$RESUME_PENDING_FLAG" "$RESUME_PROCESSING_FLAG" "$RESUME_CONTEXT_FILE" 2>/dev/null || true
  printf '%s' "$out"
}

get_handoff_artifact_path() {
  if [ -f "$HANDOFF_ARTIFACT_FILE" ]; then
    printf '%s' "$HANDOFF_ARTIFACT_FILE"
    return 0
  fi

  if [ -f "$LEGACY_PRECOMPACT_SNAPSHOT_FILE" ]; then
    printf '%s' "$LEGACY_PRECOMPACT_SNAPSHOT_FILE"
    return 0
  fi

  return 0
}

render_handoff_context() {
  local artifact_path="$1"
  if [ -z "$artifact_path" ] || [ ! -f "$artifact_path" ]; then
    return 0
  fi

  if command -v python3 >/dev/null 2>&1; then
    python3 - "$artifact_path" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
try:
    data = json.loads(path.read_text(encoding='utf-8'))
except Exception:
    sys.exit(0)

def normalize_text(value):
    if value is None:
        return ""
    if isinstance(value, str):
        return " ".join(value.split())
    if isinstance(value, dict):
        for key in ("summary", "task", "title", "detail", "message", "reason", "status"):
            candidate = value.get(key)
            if isinstance(candidate, str) and candidate.strip():
                return " ".join(candidate.split())
        return ""
    if isinstance(value, list):
        parts = []
        for item in value:
            text = normalize_text(item)
            if text:
                parts.append(text)
            if len(parts) >= 3:
                break
        return "; ".join(parts)
    return str(value)

def format_list(items, limit=4):
    parts = []
    for item in items or []:
        text = normalize_text(item)
        if text:
            parts.append(text)
        if len(parts) >= limit:
            break
    return "; ".join(parts)

previous = data.get("previous_state") or {}
next_action = data.get("next_action") or {}
open_risks = data.get("open_risks") or []
failed_checks = data.get("failed_checks") or []
decision_log = data.get("decision_log") or []
context_reset = data.get("context_reset") or {}
continuity = data.get("continuity") or {}
wip_tasks = data.get("wipTasks") or []
recent_edits = data.get("recentEdits") or []

lines = ["## Structured Handoff"]

previous_summary = normalize_text(previous.get("summary"))
if previous_summary:
    lines.append(f"- Previous state: {previous_summary}")

session_state = previous.get("session_state") or {}
if isinstance(session_state, dict):
    session_bits = []
    for key in ("state", "review_status", "active_skill", "resumed_at"):
        candidate = session_state.get(key)
        if isinstance(candidate, str) and candidate.strip():
            session_bits.append(f"{key}={candidate.strip()}")
    if session_bits:
        lines.append(f"- Session state: {', '.join(session_bits)}")

plan_counts = previous.get("plan_counts") or {}
if isinstance(plan_counts, dict):
    count_bits = []
    for key in ("total", "wip", "blocked", "recent_edits"):
        candidate = plan_counts.get(key)
        if isinstance(candidate, (int, float)) and int(candidate) > 0:
            count_bits.append(f"{key}={int(candidate)}")
    if count_bits:
        lines.append(f"- Plan counts: {', '.join(count_bits)}")

next_bits = []
next_summary = normalize_text(next_action.get("summary"))
if next_summary:
    next_bits.append(next_summary)
task_id = normalize_text(next_action.get("taskId"))
task = normalize_text(next_action.get("task"))
if task_id or task:
    next_bits.append(" ".join(part for part in [task_id, task] if part).strip())
depends = normalize_text(next_action.get("depends"))
dod = normalize_text(next_action.get("dod"))
if depends:
    next_bits.append(f"depends={depends}")
if dod:
    next_bits.append(f"DoD={dod}")
if next_bits:
    lines.append(f"- Next action: {' | '.join(next_bits)}")

if open_risks:
    lines.append(f"- Open risks: {format_list(open_risks)}")
if failed_checks:
    lines.append(f"- Failed checks: {format_list(failed_checks)}")
if decision_log:
    lines.append(f"- Decision log: {format_list(decision_log, 2)}")
context_reset_summary = normalize_text(context_reset.get("summary"))
if context_reset_summary:
    lines.append(f"- Context reset: {context_reset_summary}")
continuity_summary = normalize_text(continuity.get("summary"))
if continuity_summary:
    lines.append(f"- Continuity: {continuity_summary}")
if wip_tasks:
    lines.append(f"- WIP tasks: {format_list(wip_tasks, 5)}")
if recent_edits:
    lines.append(f"- Recent edits: {format_list(recent_edits, 5)}")

print("\n".join(lines))
PY
    return 0
  fi

  return 0
}

sync_handoff_session_metadata() {
  local artifact_path="$1"
  local session_file="$2"
  [ -n "$artifact_path" ] || return 0
  [ -f "$artifact_path" ] || return 0
  [ -n "$session_file" ] || return 0
  [ -f "$session_file" ] || return 0
  command -v jq >/dev/null 2>&1 || return 0

  local context_reset_summary
  local context_reset_recommended
  local continuity_summary
  local continuity_effort
  local continuity_plugin_first
  local continuity_resume_aware
  local tmp_file

  context_reset_summary="$(jq -r '.context_reset.summary // empty' "$artifact_path" 2>/dev/null || true)"
  context_reset_recommended="$(jq -r '.context_reset.recommended // false' "$artifact_path" 2>/dev/null || echo false)"
  continuity_summary="$(jq -r '.continuity.summary // empty' "$artifact_path" 2>/dev/null || true)"
  continuity_effort="$(jq -r '.continuity.effort_hint // empty' "$artifact_path" 2>/dev/null || true)"
  continuity_plugin_first="$(jq -r '.continuity.plugin_first_workflow // false' "$artifact_path" 2>/dev/null || echo false)"
  continuity_resume_aware="$(jq -r '.continuity.resume_aware_effort_continuity // false' "$artifact_path" 2>/dev/null || echo false)"

  tmp_file="$(mktemp)"
  jq \
    --arg artifact_path "$artifact_path" \
    --arg context_reset_summary "$context_reset_summary" \
    --arg continuity_summary "$continuity_summary" \
    --arg continuity_effort "$continuity_effort" \
    --argjson context_reset_recommended "$context_reset_recommended" \
    --argjson continuity_plugin_first "$continuity_plugin_first" \
    --argjson continuity_resume_aware "$continuity_resume_aware" \
    '
    .harness = (.harness // {}) |
    .harness.last_handoff_artifact = $artifact_path |
    .harness.context_reset = {
      summary: $context_reset_summary,
      recommended: $context_reset_recommended
    } |
    .harness.continuity = {
      summary: $continuity_summary,
      effort_hint: $continuity_effort,
      plugin_first_workflow: $continuity_plugin_first,
      resume_aware_effort_continuity: $continuity_resume_aware
    }
    ' "$session_file" > "$tmp_file" && mv "$tmp_file" "$session_file"
}

# ===== セッション復元ロジック =====
RESTORED="false"
RESTORED_SESSION_ID=""
RESTORE_METHOD=""

# 方法1: セッションマッピングから検索
if [ -n "$CC_SESSION_ID" ] && [ -f "$SESSION_MAP_FILE" ] && command -v jq >/dev/null 2>&1; then
  HARNESS_SESSION_ID="$(jq -r --arg cc_id "$CC_SESSION_ID" '.[$cc_id] // empty' "$SESSION_MAP_FILE" 2>/dev/null)"

  if [ -n "$HARNESS_SESSION_ID" ]; then
    ARCHIVE_SESSION="$ARCHIVE_DIR/${HARNESS_SESSION_ID}.json"
    ARCHIVE_EVENTS="$ARCHIVE_DIR/${HARNESS_SESSION_ID}.events.jsonl"

    if [ -f "$ARCHIVE_SESSION" ]; then
      cp "$ARCHIVE_SESSION" "$SESSION_FILE"
      [ -f "$ARCHIVE_EVENTS" ] && cp "$ARCHIVE_EVENTS" "$EVENT_LOG_FILE"

      # 状態を initialized に更新し、resume イベントを記録
      NOW=$(date -u +%Y-%m-%dT%H:%M:%SZ)
      if command -v jq >/dev/null 2>&1; then
        tmp_file=$(mktemp)
        jq --arg state "initialized" \
           --arg resumed_at "$NOW" \
           '.state = $state | .resumed_at = $resumed_at' \
           "$SESSION_FILE" > "$tmp_file" && mv "$tmp_file" "$SESSION_FILE"
      fi

      # resume イベントを記録
      echo "{\"type\":\"session.resume\",\"ts\":\"$NOW\",\"state\":\"initialized\",\"data\":{\"cc_session_id\":\"$CC_SESSION_ID\",\"method\":\"mapping\"}}" >> "$EVENT_LOG_FILE"

      RESTORED="true"
      RESTORED_SESSION_ID="$HARNESS_SESSION_ID"
      RESTORE_METHOD="mapping"
    fi
  fi
fi

# 方法2: 最新の stopped セッションを自動復元（マッピングがない場合）
if [ "$RESTORED" = "false" ]; then
  LATEST_ARCHIVE=$(ls -t "$ARCHIVE_DIR"/*.json 2>/dev/null | head -n 1 || true)

  if [ -n "$LATEST_ARCHIVE" ] && [ -f "$LATEST_ARCHIVE" ]; then
    HARNESS_SESSION_ID=$(basename "$LATEST_ARCHIVE" .json)
    ARCHIVE_EVENTS="$ARCHIVE_DIR/${HARNESS_SESSION_ID}.events.jsonl"

    cp "$LATEST_ARCHIVE" "$SESSION_FILE"
    [ -f "$ARCHIVE_EVENTS" ] && cp "$ARCHIVE_EVENTS" "$EVENT_LOG_FILE"

    NOW=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    if command -v jq >/dev/null 2>&1; then
      tmp_file=$(mktemp)
      jq --arg state "initialized" \
         --arg resumed_at "$NOW" \
         '.state = $state | .resumed_at = $resumed_at' \
         "$SESSION_FILE" > "$tmp_file" && mv "$tmp_file" "$SESSION_FILE"
    fi

    echo "{\"type\":\"session.resume\",\"ts\":\"$NOW\",\"state\":\"initialized\",\"data\":{\"cc_session_id\":\"$CC_SESSION_ID\",\"method\":\"latest\"}}" >> "$EVENT_LOG_FILE"

    RESTORED="true"
    RESTORED_SESSION_ID="$HARNESS_SESSION_ID"
    RESTORE_METHOD="latest"
  fi
fi

# 方法3: 復元対象がない場合は新規初期化
if [ "$RESTORED" = "false" ]; then
  # session-init.sh と同等の初期化を行う
  NOW=$(date -u +%Y-%m-%dT%H:%M:%SZ)
  NEW_SESSION_ID="session-$(date +%s)"

  cat > "$SESSION_FILE" <<EOF
{
  "session_id": "$NEW_SESSION_ID",
  "parent_session_id": null,
  "state": "initialized",
  "started_at": "$NOW",
  "updated_at": "$NOW",
  "resumed_at": "$NOW",
  "event_seq": 0,
  "last_event_id": ""
}
EOF

  echo "{\"type\":\"session.start\",\"ts\":\"$NOW\",\"state\":\"initialized\",\"data\":{\"cc_session_id\":\"$CC_SESSION_ID\",\"note\":\"no_archive_found\"}}" > "$EVENT_LOG_FILE"

  RESTORED_SESSION_ID="$NEW_SESSION_ID"
  RESTORE_METHOD="new"
fi

# ===== Claude Code session_id とのマッピングを保存 =====
if [ -n "$CC_SESSION_ID" ] && [ -n "$RESTORED_SESSION_ID" ]; then
  if command -v jq >/dev/null 2>&1; then
    if [ -f "$SESSION_MAP_FILE" ]; then
      tmp_file=$(mktemp)
      jq --arg cc_id "$CC_SESSION_ID" --arg harness_id "$RESTORED_SESSION_ID" \
         '.[$cc_id] = $harness_id' "$SESSION_MAP_FILE" > "$tmp_file" && mv "$tmp_file" "$SESSION_MAP_FILE"
    else
      echo "{\"$CC_SESSION_ID\":\"$RESTORED_SESSION_ID\"}" > "$SESSION_MAP_FILE"
    fi
  fi
fi

# ===== セッション間通信用の登録 =====
# active.json に自分を登録（他セッションから認識可能にする）
if [ -f "$SCRIPT_DIR/session-register.sh" ]; then
  bash "$SCRIPT_DIR/session-register.sh" "$RESTORED_SESSION_ID" 2>/dev/null || true
fi

# ===== Skills Gate 初期化（session-init.sh と同様） =====
SESSION_SKILLS_USED_FILE="${STATE_DIR}/session-skills-used.json"
echo '{"used": [], "session_start": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}' > "$SESSION_SKILLS_USED_FILE"

# ===== SSOT 同期フラグをクリア（新セッション/復元セッション開始時） =====
# このフラグは /sync-ssot-from-memory 実行時に作成され、
# Plans.md クリーンアップ前の SSOT 同期確認に使用される
rm -f "${STATE_DIR}/.ssot-synced-this-session" 2>/dev/null || true

# ultrawork 警告フラグをクリア（セッション復元時）
# このフラグは userprompt-inject-policy.sh で一度だけ警告するために使用される
# 復元時にクリアすることで、復元後の最初のプロンプトで警告が再表示される
# 後方互換: 両方のフラグ名をクリア
rm -f "${STATE_DIR}/.work-review-warned" 2>/dev/null || true
rm -f "${STATE_DIR}/.ultrawork-review-warned" 2>/dev/null || true

# ===== Plans.md チェック =====
PLANS_INFO=""
if [ -f "Plans.md" ]; then
  wip_count="$(count_matches "cc:WIP\\|pm:依頼中\\|cursor:依頼中" "Plans.md")"
  todo_count="$(count_matches "cc:TODO" "Plans.md")"
  PLANS_INFO="📄 Plans.md: 進行中 ${wip_count} / 未着手 ${todo_count}"
else
  PLANS_INFO="📄 Plans.md: 未検出"
fi

SNAPSHOT_INFO=""
if declare -F progress_snapshot_summary >/dev/null 2>&1; then
  SNAPSHOT_INFO="$(progress_snapshot_summary "${STATE_DIR}" 2>/dev/null || true)"
fi

# ===== active_skill 検出（スキル再起動が必要かチェック） =====
ACTIVE_SKILL_INFO=""
if [ -f "$SESSION_FILE" ] && command -v jq >/dev/null 2>&1; then
  ACTIVE_SKILL=$(jq -r '.active_skill // empty' "$SESSION_FILE" 2>/dev/null)
  ACTIVE_SKILL_STARTED=$(jq -r '.active_skill_started_at // "不明"' "$SESSION_FILE" 2>/dev/null)

  if [ -n "$ACTIVE_SKILL" ]; then
    ACTIVE_SKILL_INFO="
## ⚠️ MANDATORY: ${ACTIVE_SKILL} Session Recovery

**前回のセッションで \`/${ACTIVE_SKILL}\` が実行中でした（開始: ${ACTIVE_SKILL_STARTED}）**

**必須アクション:**
1. \`/${ACTIVE_SKILL} 続きやって\` でスキルを再起動してください
2. スキルを再起動せずに直接実装を開始しないでください
3. スキル文脈なしでは review_status ガードが機能しません

スキルを再起動しない場合:
- レビュー強制が機能しません
- 前回の失敗からの学習が引き継がれません
- 完了チェックが不完全になります
"
  fi
fi

# ===== Work モード検出と harness-review 必須の再注入 =====
WORK_INFO=""
WORK_FILE="${STATE_DIR}/work-active.json"
# 後方互換: work-active.json がなければ ultrawork-active.json を試行
if [ ! -f "$WORK_FILE" ]; then
  WORK_FILE="${STATE_DIR}/ultrawork-active.json"
fi
if [ -f "$WORK_FILE" ] && command -v jq >/dev/null 2>&1; then
  REVIEW_STATUS=$(jq -r '.review_status // "pending"' "$WORK_FILE" 2>/dev/null)
  STARTED_AT=$(jq -r '.started_at // "不明"' "$WORK_FILE" 2>/dev/null)

  case "$REVIEW_STATUS" in
    "passed")
      WORK_INFO="⚡ **work モード継続中** (開始: ${STARTED_AT})\n   ✅ review_status: passed → 完了処理可能"
      ;;
    "failed")
      WORK_INFO="⚡ **work モード継続中** (開始: ${STARTED_AT})\n   ❌ review_status: failed → 修正後に /harness-review を再実行してください"
      ;;
    *)
      WORK_INFO="⚡ **work モード継続中** (開始: ${STARTED_AT})\n   ⚠️ review_status: pending → **完了前に /harness-review で APPROVE を得てください**"
      ;;
  esac
fi

# ===== 出力メッセージの構築 =====
add_line "# [claude-code-harness] セッション復元"
add_line ""

case "$RESTORE_METHOD" in
  "mapping")
    add_line "✅ セッション状態を復元しました（マッピングから検出）"
    add_line "   Harness Session: ${RESTORED_SESSION_ID}"
    ;;
  "latest")
    add_line "✅ 最新のセッション状態を復元しました"
    add_line "   Harness Session: ${RESTORED_SESSION_ID}"
    ;;
  "new")
    add_line "ℹ️ 復元対象のセッションがないため、新規初期化しました"
    add_line "   Harness Session: ${RESTORED_SESSION_ID}"
    ;;
esac

add_line ""

MEMORY_CONTEXT=""
if [ -f "$RESUME_PENDING_FLAG" ] || [ -f "$RESUME_CONTEXT_FILE" ]; then
  MEMORY_CONTEXT="$(consume_memory_resume_context "$RESUME_CONTEXT_FILE" "$RESUME_MAX_BYTES")"
fi

if [ -n "$MEMORY_CONTEXT" ]; then
  OUTPUT="${OUTPUT}${MEMORY_CONTEXT}"
  case "$MEMORY_CONTEXT" in
    *$'\n') ;;
    *) OUTPUT="${OUTPUT}\n" ;;
  esac
  add_line ""
fi

HANDOFF_CONTEXT=""
HANDOFF_ARTIFACT_PATH="$(get_handoff_artifact_path)"
if [ -n "$HANDOFF_ARTIFACT_PATH" ] && [ -f "$HANDOFF_ARTIFACT_PATH" ]; then
  # stale handoff を拒否: 24時間以上前の artifact は無視（session-init.sh と同じポリシー）
  HANDOFF_AGE_LIMIT=86400
  HANDOFF_MTIME="$(stat -f %m "$HANDOFF_ARTIFACT_PATH" 2>/dev/null || stat -c %Y "$HANDOFF_ARTIFACT_PATH" 2>/dev/null || echo 0)"
  NOW="$(date +%s)"
  HANDOFF_AGE=$(( NOW - HANDOFF_MTIME ))
  if [ "$HANDOFF_AGE" -lt "$HANDOFF_AGE_LIMIT" ]; then
    HANDOFF_CONTEXT="$(render_handoff_context "$HANDOFF_ARTIFACT_PATH")"
  fi
fi

if [ -n "$HANDOFF_CONTEXT" ]; then
  OUTPUT="${OUTPUT}${HANDOFF_CONTEXT}"
  case "$HANDOFF_CONTEXT" in
    *$'\n') ;;
    *) OUTPUT="${OUTPUT}\n" ;;
  esac
  add_line ""
fi

sync_handoff_session_metadata "$HANDOFF_ARTIFACT_PATH" "$SESSION_FILE"

add_line "${PLANS_INFO}"

if [ -n "${SNAPSHOT_INFO}" ]; then
  add_line "${SNAPSHOT_INFO}"
fi

# active_skill 再起動指示を追加（最優先で表示）
if [ -n "$ACTIVE_SKILL_INFO" ]; then
  add_line ""
  add_line "$ACTIVE_SKILL_INFO"
fi

# ultrawork モード情報を追加
if [ -n "$WORK_INFO" ]; then
  add_line ""
  add_line "$WORK_INFO"
fi

# ===== JSON 出力 =====
ESCAPED_OUTPUT=$(echo -e "$OUTPUT" | sed 's/\\/\\\\/g; s/"/\\"/g; s/$/\\n/' | tr -d '\n' | sed 's/\\n$//')

cat <<EOF
{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"${ESCAPED_OUTPUT}"}}
EOF

exit 0
