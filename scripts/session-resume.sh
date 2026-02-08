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

# ===== 出力メッセージを蓄積 =====
OUTPUT=""
add_line() {
  OUTPUT="${OUTPUT}$1\n"
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
  wip_count=$(grep -c "cc:WIP\|pm:依頼中\|cursor:依頼中" Plans.md 2>/dev/null || echo "0")
  todo_count=$(grep -c "cc:TODO" Plans.md 2>/dev/null || echo "0")
  PLANS_INFO="📄 Plans.md: 進行中 ${wip_count} / 未着手 ${todo_count}"
else
  PLANS_INFO="📄 Plans.md: 未検出"
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
add_line "${PLANS_INFO}"

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
