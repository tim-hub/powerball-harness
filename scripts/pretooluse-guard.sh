#!/bin/bash
# pretooluse-guard.sh
# Claude Code Hooks: PreToolUse guardrail for dangerous operations.
# - Deny writes/edits to protected paths (e.g., .git/, .env, keys)
# - Ask for confirmation for writes outside the project directory
# - Deny sudo, ask for confirmation for rm -rf / git push
#
# Input: stdin JSON from Claude Code hooks
# Output: JSON to control PreToolUse permission decisions
#
# Cross-platform: Supports Windows (Git Bash/MSYS2/Cygwin/WSL), macOS, Linux

set +e

# Load cross-platform path utilities
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/path-utils.sh" ]; then
  # shellcheck source=./path-utils.sh
  source "$SCRIPT_DIR/path-utils.sh"
else
  # Fallback: define minimal path utilities if path-utils.sh not found
  is_absolute_path() {
    local p="$1"
    [[ "$p" == /* ]] && return 0
    [[ "$p" =~ ^[A-Za-z]:[\\/] ]] && return 0
    return 1
  }
  normalize_path() {
    local p="$1"
    p="${p//\\//}"
    echo "$p"
  }
  # Note: This expects already-normalized paths from caller for performance
  is_path_under() {
    local child="$1"
    local parent="$2"
    [[ "$parent" != */ ]] && parent="${parent}/"
    [[ "${child}/" == "${parent}"* ]] || [ "$child" = "${parent%/}" ]
  }
fi

detect_lang() {
  # Default to Japanese for this harness (can be overridden).
  # - CLAUDE_CODE_HARNESS_LANG=en for English
  # - CLAUDE_CODE_HARNESS_LANG=ja for Japanese
  if [ -n "${CLAUDE_CODE_HARNESS_LANG:-}" ]; then
    echo "${CLAUDE_CODE_HARNESS_LANG}"
    return 0
  fi
  echo "ja"
}

LANG_CODE="$(detect_lang)"

# ===== Work Mode Detection =====
# During /work (auto-iteration), skip certain confirmation prompts
# Security: limit bypass with expiration (24 hours)
# Note: CWD is retrieved from JSON later, only initialization here
# Backward compatibility: detect ultrawork-active.json as work-active.json

WORK_MODE="false"
WORK_BYPASS_RM_RF="false"
WORK_BYPASS_GIT_PUSH="false"
WORK_MAX_AGE_HOURS=24

# ===== Codex Mode Detection =====
# In --codex mode, Claude acts as PM and Edit/Write is prohibited
# (implementation is delegated to Codex Worker)
# Detected by codex_mode: true in work-active.json
CODEX_MODE="false"

# ===== Breezing Role Guard =====
# Agent Teams Teammate role-based access control
# Identify sessions by session_id / agent_id and restrict Write/Edit by role
BREEZING_ROLE=""
BREEZING_OWNS=""
SESSION_ID=""
AGENT_ID=""
AGENT_TYPE=""
BREEZING_ROLE_KEY=""

# ===== Breezing-Codex Mode Detection =====
# In breezing-codex mode (impl_mode: "codex"), block direct Write/Edit
# (implementation is delegated to Codex Implementer via codex exec CLI)
BREEZING_CODEX_MODE="false"

# Work mode detection function (called after CWD is obtained)
# Prioritize work-active.json, fallback to ultrawork-active.json for backward compat
check_work_mode() {
  local cwd_path="$1"
  local active_file="${cwd_path}/.claude/state/work-active.json"

  # Backward compat: try ultrawork-active.json if work-active.json not found
  if [ ! -f "$active_file" ]; then
    active_file="${cwd_path}/.claude/state/ultrawork-active.json"
  fi

  [ ! -f "$active_file" ] && return

  if ! command -v jq >/dev/null 2>&1; then
    echo "[work] Warning: jq not installed, guard bypass disabled" >&2
    return
  fi

  local is_active
  is_active=$(jq -r '.active // false' "$active_file" 2>/dev/null || echo "false")
  [ "$is_active" != "true" ] && return

  # Expiration check (within 24 hours from started_at)
  local started_at
  started_at=$(jq -r '.started_at // empty' "$active_file" 2>/dev/null)
  [ -z "$started_at" ] && return

  # Parse ISO8601 (compatible with both macOS/Linux)
  # Remove Z suffix for parsing
  local started_clean="${started_at%%Z*}"
  started_clean="${started_clean%%+*}"  # Also remove timezone offset
  started_clean="${started_clean%%.*}"  # Also remove milliseconds

  local started_epoch=0
  local current_epoch
  current_epoch=$(date +%s)

  # macOS: date -j -f, Linux: date -d
  if [[ "$OSTYPE" == "darwin"* ]]; then
    started_epoch=$(date -j -f "%Y-%m-%dT%H:%M:%S" "$started_clean" +%s 2>/dev/null || echo 0)
  else
    started_epoch=$(date -d "${started_at}" +%s 2>/dev/null || echo 0)
  fi

  if [ "$started_epoch" -eq 0 ]; then
    echo "[work] Warning: failed to parse started_at, guard bypass disabled" >&2
    return
  fi

  # Future time check (prevent tampering)
  if [ "$started_epoch" -gt "$current_epoch" ]; then
    echo "[work] Warning: started_at is in the future, guard bypass disabled" >&2
    return
  fi

  local age_hours=$(( (current_epoch - started_epoch) / 3600 ))
  if [ "$age_hours" -ge "$WORK_MAX_AGE_HOURS" ]; then
    rm -f "$active_file" 2>/dev/null || true
    echo "[work] Warning: work-active.json expired (${age_hours}h >= ${WORK_MAX_AGE_HOURS}h), removed" >&2
    return
  fi

  WORK_MODE="true"
  # Performance: extract bypass_guards and codex_mode in one jq call to avoid re-reading
  local _work_extras
  _work_extras=$(jq -r '[
    (if .bypass_guards | type == "array" then (.bypass_guards | contains(["rm_rf"])) else false end),
    (if .bypass_guards | type == "array" then (.bypass_guards | contains(["git_push"])) else false end),
    (.codex_mode // false)
  ] | @tsv' "$active_file" 2>/dev/null)
  if [ -n "$_work_extras" ]; then
    IFS=$'\t' read -r WORK_BYPASS_RM_RF WORK_BYPASS_GIT_PUSH _work_codex_mode <<< "$_work_extras"
    # Cache codex_mode for check_codex_mode to avoid re-parsing
    WORK_CACHED_CODEX_MODE="${_work_codex_mode}"
  else
    WORK_BYPASS_RM_RF="false"
    WORK_BYPASS_GIT_PUSH="false"
  fi
}

# Codex mode detection function (called after CWD is obtained)
# Block Claude Edit/Write when codex_mode: true in work-active.json
# Prerequisite: only set CODEX_MODE when WORK_MODE is true and TTL is valid
# Performance: prefer cached values from check_work_mode
check_codex_mode() {
  # Skip if Work mode is not active (considering TTL expiration)
  [ "$WORK_MODE" != "true" ] && return

  # Use cached value from check_work_mode if available (avoids re-reading file)
  if [ -n "${WORK_CACHED_CODEX_MODE:-}" ]; then
    [ "$WORK_CACHED_CODEX_MODE" = "true" ] && CODEX_MODE="true"
    return
  fi

  # Fallback: read file directly (for python3-only environments where jq cache wasn't set)
  local cwd_path="$1"
  local active_file="${cwd_path}/.claude/state/work-active.json"

  # Backward compat: try ultrawork-active.json if work-active.json not found
  if [ ! -f "$active_file" ]; then
    active_file="${cwd_path}/.claude/state/ultrawork-active.json"
  fi

  [ ! -f "$active_file" ] && return

  local is_codex="false"

  if command -v python3 >/dev/null 2>&1; then
    is_codex=$(python3 -c '
import json, sys
try:
    with open(sys.argv[1]) as f:
        data = json.load(f)
    val = data.get("codex_mode", False)
    print("true" if val is True else "false")
except:
    print("false")
' "$active_file" 2>/dev/null || echo "false")
  fi

  [ "$is_codex" = "true" ] && CODEX_MODE="true"
}

# Breezing role detection function (called after CWD + SESSION_ID/AGENT_ID obtained)
# Look up role from .claude/state/breezing-session-roles.json
check_breezing_role() {
  local cwd_path="$1"
  local roles_file="${cwd_path}/.claude/state/breezing-session-roles.json"

  [ -z "$SESSION_ID" ] && [ -z "$AGENT_ID" ] && return
  [ ! -f "$roles_file" ] && return

  if ! command -v jq >/dev/null 2>&1; then
    return
  fi

  local lookup_key=""
  local role=""
  local owns=""

  for lookup_key in "$AGENT_ID" "$SESSION_ID"; do
    [ -z "$lookup_key" ] && continue
    role="$(jq -r --arg sid "$lookup_key" '.[$sid].role // empty' "$roles_file" 2>/dev/null)"
    [ -z "$role" ] && continue
    owns="$(jq -r --arg sid "$lookup_key" '.[$sid].owns // empty' "$roles_file" 2>/dev/null)"
    BREEZING_ROLE="$role"
    BREEZING_OWNS="$owns"
    BREEZING_ROLE_KEY="$lookup_key"
    return
  done
}

# Breezing-Codex mode detection function (called after CWD obtained)
# Block direct Write/Edit when impl_mode: "codex" in breezing-active.json
check_breezing_codex_mode() {
  local cwd_path="$1"
  local active_file="${cwd_path}/.claude/state/breezing-active.json"

  [ ! -f "$active_file" ] && return

  local is_codex="false"

  if command -v jq >/dev/null 2>&1; then
    local impl_mode
    impl_mode=$(jq -r '.impl_mode // empty' "$active_file" 2>/dev/null)
    [ "$impl_mode" = "codex" ] && is_codex="true"
  elif command -v python3 >/dev/null 2>&1; then
    is_codex=$(python3 -c '
import json, sys
try:
    with open(sys.argv[1]) as f:
        data = json.load(f)
    val = data.get("impl_mode", "")
    print("true" if val == "codex" else "false")
except:
    print("false")
' "$active_file" 2>/dev/null || echo "false")
  else
    echo "[Breezing-Codex] Warning: jq/python3 not found, breezing-codex mode detection disabled" >&2
    return
  fi

  [ "$is_codex" = "true" ] && BREEZING_CODEX_MODE="true"
}

# Detect and handle Breezing role registration Write
# Register session_id / agent_id → role on first Write (breezing-role-*.json)
try_register_breezing_role() {
  local file_path="$1"
  local cwd_path="$2"
  local roles_file="${cwd_path}/.claude/state/breezing-session-roles.json"

  # Only target writes to breezing-role-*.json
  BASENAME_ROLE="${file_path##*/}"
  case "$BASENAME_ROLE" in
    breezing-role-*.json) ;;
    *) return 1 ;;
  esac

  # Verify path is under .claude/state/
  case "$file_path" in
    .claude/state/breezing-role-*.json|*/.claude/state/breezing-role-*.json) ;;
    *) return 1 ;;
  esac

  [ -z "$SESSION_ID" ] && [ -z "$AGENT_ID" ] && return 1

  if ! command -v jq >/dev/null 2>&1; then
    return 1
  fi

  # Extract role info from tool_input.content
  local content role owns
  content=$(echo "$INPUT" | jq -r '.tool_input.content // empty' 2>/dev/null)
  [ -z "$content" ] && return 1

  role=$(echo "$content" | jq -r '.role // empty' 2>/dev/null)
  [ -z "$role" ] && return 1

  # Security: only allow known role values
  case "$role" in
    reviewer|implementer|lead) ;;
    *) return 1 ;;
  esac

  owns=$(echo "$content" | jq -c '.owns // []' 2>/dev/null || echo '[]')

  # Register session_id → role mapping
  mkdir -p "${cwd_path}/.claude/state" 2>/dev/null || true

  if [ ! -f "$roles_file" ]; then
    echo '{}' > "$roles_file"
  fi

  jq \
    --arg sid "$SESSION_ID" \
    --arg aid "$AGENT_ID" \
    --arg atype "$AGENT_TYPE" \
    --arg role "$role" \
    --argjson owns "$owns" \
    '
      (if $sid != "" then .[$sid] = {"role": $role, "owns": $owns, "agent_type": $atype} else . end)
      | (if $aid != "" then .[$aid] = {"role": $role, "owns": $owns, "agent_type": $atype} else . end)
    ' \
    "$roles_file" > "${roles_file}.tmp" && mv "${roles_file}.tmp" "$roles_file"

  return 0
}

msg() {
  # msg <key> [arg]
  local key="$1"
  local arg="${2:-}"

  if [ "$LANG_CODE" = "en" ]; then
    case "$key" in
      deny_path_traversal) echo "Blocked: path traversal in file_path ($arg)" ;;
      ask_write_outside_project) echo "Confirm: writing outside project directory ($arg)" ;;
      deny_protected_path) echo "Blocked: protected path ($arg)" ;;
      deny_sudo) echo "Blocked: sudo is not allowed via Claude Code hooks" ;;
      ask_git_push) echo "Confirm: git push requested ($arg)" ;;
      ask_rm_rf) echo "Confirm: rm -rf requested ($arg)" ;;
      deny_git_commit_no_review) echo "Blocked: Run /harness-review before committing. After review approval, run git commit again." ;;
      deny_codex_mode) echo "[Codex Mode] Claude is the PM. Direct Edit/Write is prohibited. Delegate implementation to Codex Worker via codex exec (CLI)." ;;
      deny_breezing_codex_mode) echo "[Breezing-Codex] Direct Edit/Write is prohibited in codex impl mode. Implementation must go through codex exec (CLI)." ;;
      deny_codex_mcp) echo "Blocked: Codex MCP is deprecated. Use 'codex exec' (Bash) instead. See .claude/rules/codex-cli-only.md" ;;
      *) echo "$key $arg" ;;
    esac
    return 0
  fi

  # ja (default)
  case "$key" in
    deny_path_traversal) echo "Blocked: suspected path traversal (file_path: $arg)" ;;
    ask_write_outside_project) echo "Confirm: writing outside project (file_path: $arg)" ;;
    deny_protected_path) echo "Blocked: operation on protected path (path: $arg)" ;;
    deny_sudo) echo "Blocked: sudo is not allowed via hooks" ;;
    ask_git_push) echo "Confirm: about to execute git push (command: $arg)" ;;
    ask_rm_rf) echo "Confirm: about to execute rm -rf (command: $arg)" ;;
    deny_git_commit_no_review) echo "Blocked: please run /harness-review before committing. After review, you can run git commit again." ;;
    deny_codex_mode) echo "[Codex Mode] In --codex mode, Claude acts as PM. Direct Edit/Write is prohibited. Delegate implementation to Codex Worker via codex exec (CLI)." ;;
    deny_breezing_codex_mode) echo "[Breezing-Codex] Direct Edit/Write is prohibited in codex implementation mode. Implementation must go through codex exec (CLI)." ;;
    deny_codex_mcp) echo "Blocked: Codex MCP is deprecated. Use 'codex exec' (Bash) instead. See .claude/rules/codex-cli-only.md" ;;
    *) echo "$key $arg" ;;
  esac
}

INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null)"
fi

[ -z "$INPUT" ] && exit 0

TOOL_NAME=""
FILE_PATH=""
COMMAND=""
CWD=""

if command -v jq >/dev/null 2>&1; then
  # Performance: extract all fields in one jq call instead of 5 separate invocations
  _jq_parsed="$(echo "$INPUT" | jq -r '[
    (.tool_name // ""),
    (.tool_input.file_path // ""),
    (.tool_input.command // ""),
    (.cwd // ""),
    (.session_id // ""),
    (.agent_id // ""),
    (.agent_type // "")
  ] | @tsv' 2>/dev/null)"
  if [ -n "$_jq_parsed" ]; then
    IFS=$'\t' read -r TOOL_NAME FILE_PATH COMMAND CWD SESSION_ID AGENT_ID AGENT_TYPE <<< "$_jq_parsed"
  fi
  unset _jq_parsed
elif command -v python3 >/dev/null 2>&1; then
  # Performance+Security: extract all fields in one python3 call (no eval)
  _py_parsed="$(echo "$INPUT" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
except Exception:
    data = {}
def get_nested(d, path):
    for k in path.split('.'):
        if isinstance(d, dict):
            d = d.get(k) or ''
        else:
            return ''
    return d if isinstance(d, str) else ''
fields = ['tool_name', 'tool_input.file_path', 'tool_input.command', 'cwd', 'session_id', 'agent_id', 'agent_type']
print('\t'.join(get_nested(data, f) for f in fields))
" 2>/dev/null)"
  if [ -n "$_py_parsed" ]; then
    IFS=$'\t' read -r TOOL_NAME FILE_PATH COMMAND CWD SESSION_ID AGENT_ID AGENT_TYPE <<< "$_py_parsed"
  fi
  unset _py_parsed
fi

[ -z "$TOOL_NAME" ] && exit 0

# ===== Execute Work mode detection (after CWD obtained) =====
if [ -n "$CWD" ]; then
  check_work_mode "$CWD"
  check_codex_mode "$CWD"
  check_breezing_role "$CWD"
  check_breezing_codex_mode "$CWD"
fi

# ===== Cost Control: track tool call counts per session =====
CONFIG_FILE=".claude-code-harness.config.yaml"
STATE_DIR=".claude/state"
COST_STATE_FILE="$STATE_DIR/cost-state.json"

check_cost_control() {
  local tool="$1"

  # Check cost_control.enabled
  if [ ! -f "$CONFIG_FILE" ]; then
    return 0
  fi

  local cost_enabled
  cost_enabled=$(grep -E "^  enabled:" "$CONFIG_FILE" 2>/dev/null | head -n 1 | awk '{print $2}' || echo "false")
  if [ "$cost_enabled" != "true" ]; then
    return 0
  fi

  # Initialize cost-state.json if not present
  # Security: refuse if state dir or file is a symlink (prevents symlink-based overwrites)
  if [ -L "$STATE_DIR" ] || [ -L "$COST_STATE_FILE" ]; then
    return 0
  fi
  if [ ! -f "$COST_STATE_FILE" ]; then
    mkdir -p "$STATE_DIR" 2>/dev/null || true
    echo '{"total_tool_calls":0,"edit_calls":0,"bash_calls":0}' > "$COST_STATE_FILE"
  fi

  if command -v jq >/dev/null 2>&1; then
    # Get current counts
    local total_calls edit_calls bash_calls
    total_calls=$(jq -r '.total_tool_calls // 0' "$COST_STATE_FILE" 2>/dev/null || echo 0)
    edit_calls=$(jq -r '.edit_calls // 0' "$COST_STATE_FILE" 2>/dev/null || echo 0)
    bash_calls=$(jq -r '.bash_calls // 0' "$COST_STATE_FILE" 2>/dev/null || echo 0)

    # Get limits from config
    local total_limit edit_limit bash_limit warn_percent
    total_limit=$(grep -A5 "session_limits:" "$CONFIG_FILE" 2>/dev/null | grep "total_tool_calls:" | awk '{print $2}' || echo 500)
    edit_limit=$(grep -A5 "session_limits:" "$CONFIG_FILE" 2>/dev/null | grep "edit_calls:" | awk '{print $2}' || echo 100)
    bash_limit=$(grep -A5 "session_limits:" "$CONFIG_FILE" 2>/dev/null | grep "bash_calls:" | awk '{print $2}' || echo 200)
    warn_percent=$(grep "warn_threshold_percent:" "$CONFIG_FILE" 2>/dev/null | awk '{print $2}' || echo 80)

    # Increment counts
    total_calls=$((total_calls + 1))
    case "$tool" in
      Write|Edit) edit_calls=$((edit_calls + 1)) ;;
      Bash) bash_calls=$((bash_calls + 1)) ;;
    esac

    # Update cost-state.json
    jq --argjson t "$total_calls" --argjson e "$edit_calls" --argjson b "$bash_calls" \
      '.total_tool_calls = $t | .edit_calls = $e | .bash_calls = $b' \
      "$COST_STATE_FILE" > "${COST_STATE_FILE}.tmp" && mv "${COST_STATE_FILE}.tmp" "$COST_STATE_FILE"

    # Limit check
    if [ "$total_calls" -ge "$total_limit" ]; then
      echo "[Cost Control] Session tool call limit ($total_limit) reached. Please start a new session."
      return 1
    fi

    case "$tool" in
      Write|Edit)
        if [ "$edit_calls" -ge "$edit_limit" ]; then
          echo "[Cost Control] Edit/Write call limit ($edit_limit) reached."
          return 1
        fi
        ;;
      Bash)
        if [ "$bash_calls" -ge "$bash_limit" ]; then
          echo "[Cost Control] Bash call limit ($bash_limit) reached."
          return 1
        fi
        ;;
    esac

    # Warning threshold check (warn via additionalContext)
    local warn_total=$((total_limit * warn_percent / 100))
    local warn_edit=$((edit_limit * warn_percent / 100))
    local warn_bash=$((bash_limit * warn_percent / 100))

    local warnings=""
    if [ "$total_calls" -ge "$warn_total" ] && [ "$total_calls" -lt "$total_limit" ]; then
      warnings="${warnings}[Cost Warning] Total tool calls: ${total_calls}/${total_limit} (exceeded ${warn_percent}%)\n"
    fi
    case "$tool" in
      Write|Edit)
        if [ "$edit_calls" -ge "$warn_edit" ] && [ "$edit_calls" -lt "$edit_limit" ]; then
          warnings="${warnings}[Cost Warning] Edit/Write: ${edit_calls}/${edit_limit}\n"
        fi
        ;;
      Bash)
        if [ "$bash_calls" -ge "$warn_bash" ] && [ "$bash_calls" -lt "$bash_limit" ]; then
          warnings="${warnings}[Cost Warning] Bash: ${bash_calls}/${bash_limit}\n"
        fi
        ;;
    esac

    if [ -n "$warnings" ]; then
      echo -e "$warnings"
      return 2  # Warning (not a block)
    fi
  fi

  return 0
}

# Cost control check runs after emit_deny is defined (executed later)

emit_decision() {
  local decision="$1"
  local reason="$2"
  local additional_context="${3:-}"

  if command -v jq >/dev/null 2>&1; then
    if [ -n "$additional_context" ]; then
      jq -nc --arg decision "$decision" --arg reason "$reason" --arg ctx "$additional_context" \
        '{hookSpecificOutput:{hookEventName:"PreToolUse", permissionDecision:$decision, permissionDecisionReason:$reason, additionalContext:$ctx}}'
    else
      jq -nc --arg decision "$decision" --arg reason "$reason" \
        '{hookSpecificOutput:{hookEventName:"PreToolUse", permissionDecision:$decision, permissionDecisionReason:$reason}}'
    fi
    return 0
  fi

  if command -v python3 >/dev/null 2>&1; then
    DECISION="$decision" REASON="$reason" ADDITIONAL_CONTEXT="$additional_context" python3 - <<'PY'
import json, os
output = {
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": os.environ.get("DECISION", ""),
    "permissionDecisionReason": os.environ.get("REASON", ""),
  }
}
ctx = os.environ.get("ADDITIONAL_CONTEXT", "")
if ctx:
    output["hookSpecificOutput"]["additionalContext"] = ctx
print(json.dumps(output))
PY
    return 0
  fi

  # Fallback: omit reason and additionalContext to avoid JSON escaping issues.
  printf '%s' "{\"hookSpecificOutput\":{\"hookEventName\":\"PreToolUse\",\"permissionDecision\":\"${decision}\"}}"
}

emit_deny() {
  # Record hook blocking event (non-blocking, fire-and-forget)
  local SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  if [ -x "$SCRIPT_DIR/record-usage.js" ] && command -v node >/dev/null 2>&1; then
    node "$SCRIPT_DIR/record-usage.js" hook pretooluse-guard --blocked >/dev/null 2>&1 &
  fi
  emit_decision "deny" "$1"
}
emit_ask() { emit_decision "ask" "$1"; }

# ===== Codex MCP block (deprecated) =====
# MCP server has been removed. Failsafe in case text fixes were missed.
if [[ "$TOOL_NAME" == mcp__codex__* ]]; then
  emit_deny "$(msg deny_codex_mcp)"
  exit 0
fi

# ===== Execute cost control check =====
COST_CHECK_MSG=""
COST_CHECK_MSG=$(check_cost_control "$TOOL_NAME")
COST_CHECK_RESULT=$?

if [ "$COST_CHECK_RESULT" -eq 1 ]; then
  # Limit reached → deny
  emit_deny "$COST_CHECK_MSG"
  exit 0
fi
# Warning (result=2) will be included in additionalContext in subsequent processing

# ===== Generate additionalContext guidelines (Claude Code v2.1.9+) =====
# Return guidelines based on file path during Write/Edit operations

TEST_QUALITY_GUIDELINE="[Test Quality Guidelines]
- Prohibited: changing to it.skip() / test.skip()
- Prohibited: removing or weakening assertions
- Prohibited: adding eslint-disable comments"

IMPL_QUALITY_GUIDELINE="[Implementation Quality Guidelines]
- Prohibited: hardcoding test expected values
- Prohibited: stubs, mocks, or empty implementations
- Must implement meaningful logic"

# Return guidelines based on file path
# Argument: $1 = file path (relative or absolute)
# Return: guideline string (empty if no match)
get_guideline_for_path() {
  local path="$1"

  # Test file pattern
  case "$path" in
    tests/*|test/*|__tests__/*|*.spec.ts|*.spec.tsx|*.spec.js|*.spec.jsx|*.test.ts|*.test.tsx|*.test.js|*.test.jsx)
      echo "$TEST_QUALITY_GUIDELINE"
      return 0
      ;;
  esac

  # Implementation file pattern
  case "$path" in
    src/*.ts|src/*.tsx|src/*.js|src/*.jsx|lib/*.ts|lib/*.tsx|lib/*.js|lib/*.jsx)
      echo "$IMPL_QUALITY_GUIDELINE"
      return 0
      ;;
  esac

  # No match
  echo ""
}

# Return explicit "allow" with additionalContext
# Omitting permissionDecision causes ambiguous behavior and prompts even in bypass mode
# Explicitly allow with permissionDecision: "allow" to avoid prompts
emit_approve_with_context() {
  local context="$1"
  if [ -n "$context" ]; then
    if command -v jq >/dev/null 2>&1; then
      jq -nc --arg ctx "$context" \
        '{hookSpecificOutput:{hookEventName:"PreToolUse", permissionDecision:"allow", additionalContext:$ctx}}'
    elif command -v python3 >/dev/null 2>&1; then
      ADDITIONAL_CONTEXT="$context" python3 -c '
import json, os
print(json.dumps({"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","additionalContext":os.environ["ADDITIONAL_CONTEXT"]}}))
'
    fi
  fi
  # Output nothing for empty context (default behavior)
}

is_path_traversal() {
  local p="$1"
  [[ "$p" == ".." ]] && return 0
  [[ "$p" == "../"* ]] && return 0
  [[ "$p" == *"/../"* ]] && return 0
  [[ "$p" == *"/.." ]] && return 0
  return 1
}

# Resolve symlinks and return the canonical (real) path.
# Falls back to the input path if realpath is unavailable or the path doesn't exist yet.
resolve_real_path() {
  local p="$1"
  local base_dir="${2:-}"

  # If relative path and base_dir given, prepend it
  if [ -n "$base_dir" ] && ! is_absolute_path "$p"; then
    p="${base_dir}/${p}"
  fi

  # Try realpath (GNU/macOS) first, then readlink -f (Linux), then Python fallback
  if command -v realpath >/dev/null 2>&1; then
    realpath "$p" 2>/dev/null && return 0
  fi
  if command -v readlink >/dev/null 2>&1; then
    readlink -f "$p" 2>/dev/null && return 0
  fi
  if command -v python3 >/dev/null 2>&1; then
    python3 -c "import os,sys; print(os.path.realpath(sys.argv[1]))" "$p" 2>/dev/null && return 0
  fi

  # Fallback: return normalized input
  echo "$p"
}

is_protected_path() {
  local p="$1"
  case "$p" in
    .git/*|*/.git/*) return 0 ;;
    .env|.env.*|*/.env|*/.env.*) return 0 ;;
    secrets/*|*/secrets/*) return 0 ;;
    *.pem|*.key|*id_rsa*|*id_ed25519*|*/.ssh/*) return 0 ;;
  esac
  return 1
}


if [ "$TOOL_NAME" = "Write" ] || [ "$TOOL_NAME" = "Edit" ]; then
  [ -z "$FILE_PATH" ] && exit 0

  if is_path_traversal "$FILE_PATH"; then
    emit_deny "$(msg deny_path_traversal "$FILE_PATH")"
    exit 0
  fi

  # ===== Symlink bypass protection =====
  # Resolve the real path to prevent symlink-based bypasses of protected path checks.
  # Example: attacker creates symlink "safe.txt -> ../../.env" to bypass is_protected_path.
  RESOLVED_FILE_PATH="$(resolve_real_path "$FILE_PATH" "$CWD")"

  # If the resolved path differs from the original, re-check for path traversal
  if [ "$RESOLVED_FILE_PATH" != "$FILE_PATH" ]; then
    # Check if symlink target points to a protected path
    RESOLVED_REL_PATH="$RESOLVED_FILE_PATH"
    if [ -n "$CWD" ]; then
      RESOLVED_NORM_CWD="$(normalize_path "$CWD")"
      RESOLVED_CWD_SLASH="${RESOLVED_NORM_CWD%/}/"
      if [[ "$RESOLVED_FILE_PATH" == "$RESOLVED_CWD_SLASH"* ]]; then
        RESOLVED_REL_PATH="${RESOLVED_FILE_PATH#$RESOLVED_CWD_SLASH}"
      fi
    fi
    if is_protected_path "$RESOLVED_REL_PATH"; then
      emit_deny "$(msg deny_protected_path "$FILE_PATH -> $RESOLVED_REL_PATH")"
      exit 0
    fi
    # Check if symlink escapes project directory
    if [ -n "$CWD" ] && is_absolute_path "$RESOLVED_FILE_PATH"; then
      if ! is_path_under "$RESOLVED_FILE_PATH" "$CWD"; then
        emit_deny "$(msg deny_path_traversal "$FILE_PATH -> $RESOLVED_FILE_PATH")"
        exit 0
      fi
    fi
  fi

  # ===== Codex Mode: PM cannot Edit/Write (Plans.md is allowed) =====
  if [ "$CODEX_MODE" = "true" ]; then
    # Allow Plans.md status marker updates (legitimate PM operation)
    # Strict pattern: only allow when filename is exactly "Plans.md"
    # Reject symbolic links (security measure)
    if [ -L "$FILE_PATH" ]; then
      emit_deny "[Codex Mode] Symbolic links are not allowed for Plans.md"
      exit 0
    fi
    # Not using local since this is outside a function
    BASENAME_FILE="${FILE_PATH##*/}"
    if [ "$BASENAME_FILE" = "Plans.md" ]; then
      : # Allow (exactly Plans.md only)
    else
      emit_deny "$(msg deny_codex_mode)"
      exit 0
    fi
  fi

  # ===== Breezing-Codex Mode: block direct Edit/Write =====
  if [ "$BREEZING_CODEX_MODE" = "true" ]; then
    if [ -L "$FILE_PATH" ]; then
      emit_deny "[Breezing-Codex] Symbolic links are not allowed"
      exit 0
    fi
    # Allowlist: breezing-related state, review state, *.md (documents)
    # Security: do not allow control files like ultrawork-active.json
    case "$FILE_PATH" in
      .claude/state/breezing*|*/.claude/state/breezing*) ;; # breezing state is allowed
      .claude/state/review*|*/.claude/state/review*) ;; # review state is allowed
      *.md) ;; # Document files are allowed
      *)
        emit_deny "$(msg deny_breezing_codex_mode)"
        exit 0
        ;;
    esac
  fi

  # ===== Breezing Role Guard: Teammate role-based access control =====
  if { [ -n "$SESSION_ID" ] || [ -n "$AGENT_ID" ]; } && [ -n "$CWD" ]; then
    # Detect role registration Write (Write to breezing-role-*.json is registration)
    if try_register_breezing_role "$FILE_PATH" "$CWD" 2>/dev/null; then
      exit 0  # Allow registration Write
    fi

    # Reviewer: block Write/Edit (.claude/state/ is allowed)
    if [ "$BREEZING_ROLE" = "reviewer" ]; then
      case "$FILE_PATH" in
        .claude/state/*|*/.claude/state/*) ;; # state files are allowed
        *)
          emit_deny "[Breezing] Reviewer is Read-only. Code modifications are the Implementer responsibility."
          exit 0
          ;;
      esac
    fi

    # Implementer: block Write/Edit to files outside owns
    if [ "$BREEZING_ROLE" = "implementer" ] && [ -n "$BREEZING_OWNS" ] && [ "$BREEZING_OWNS" != "null" ]; then
      # .claude/state/ is always allowed
      case "$FILE_PATH" in
        .claude/state/*|*/.claude/state/*) ;; # state files are allowed
        *.md) ;; # Document files are allowed
        *)
          # Match against owns paths
          BREEZING_FILE_ALLOWED="false"

          # Calculate relative path from CWD (REL_PATH is undefined at this point)
          BREEZING_REL_PATH="$FILE_PATH"
          if [ -n "$CWD" ]; then
            BREEZING_REL_PATH="${FILE_PATH#${CWD}/}"
          fi

          # Get owns array with jq and match
          if [ -f "${CWD}/.claude/state/breezing-session-roles.json" ]; then
            ROLE_KEY="${BREEZING_ROLE_KEY:-$SESSION_ID}"
            while IFS= read -r OWNED_PATTERN; do
              [ -z "$OWNED_PATTERN" ] && continue
              # Match by absolute path
              case "$FILE_PATH" in
                $OWNED_PATTERN*) BREEZING_FILE_ALLOWED="true"; break ;;
              esac
              # Match by relative path too
              case "$BREEZING_REL_PATH" in
                $OWNED_PATTERN*) BREEZING_FILE_ALLOWED="true"; break ;;
              esac
            done < <(jq -r --arg sid "$ROLE_KEY" '.[$sid].owns[]? // empty' \
              "${CWD}/.claude/state/breezing-session-roles.json" 2>/dev/null)
          fi

          if [ "$BREEZING_FILE_ALLOWED" = "false" ]; then
            emit_deny "[Breezing] This file is outside the owns scope: $FILE_PATH"
            exit 0
          fi
          ;;
      esac
    fi
  fi

  # Normalize paths for cross-platform comparison
  NORM_FILE_PATH="$(normalize_path "$FILE_PATH")"
  NORM_CWD="$(normalize_path "$CWD")"

  # If absolute and outside project cwd, ask for confirmation.
  # Supports both Unix (/path) and Windows (C:/path, C:\path) absolute paths
  if [ -n "$NORM_CWD" ] && is_absolute_path "$NORM_FILE_PATH"; then
    if ! is_path_under "$NORM_FILE_PATH" "$NORM_CWD"; then
      emit_ask "$(msg ask_write_outside_project "$FILE_PATH")"
      exit 0
    fi
  fi

  # Normalize to relative when possible for pattern matching.
  REL_PATH="$NORM_FILE_PATH"
  if [ -n "$NORM_CWD" ] && is_path_under "$NORM_FILE_PATH" "$NORM_CWD"; then
    # Remove the CWD prefix to get relative path
    # Not using local since this is outside a function
    CWD_WITH_SLASH="${NORM_CWD%/}/"
    if [[ "$NORM_FILE_PATH" == "$CWD_WITH_SLASH"* ]]; then
      REL_PATH="${NORM_FILE_PATH#$CWD_WITH_SLASH}"
    fi
  fi

  if is_protected_path "$REL_PATH"; then
    emit_deny "$(msg deny_protected_path "$REL_PATH")"
    exit 0
  fi

  # ===== LSP/Skills gate (Phase0+) =====
  STATE_DIR=".claude/state"
  SESSION_FILE="$STATE_DIR/session.json"
  TOOLING_POLICY_FILE="$STATE_DIR/tooling-policy.json"
  SKILLS_POLICY_FILE="$STATE_DIR/skills-policy.json"
  SKILLS_CONFIG_FILE="$STATE_DIR/skills-config.json"
  SESSION_SKILLS_USED_FILE="$STATE_DIR/session-skills-used.json"

  # Default exclusion patterns (applied even without policy file)
  is_default_excluded() {
    local path="$1"
    # .md, .txt, .json files are always excluded (documents/config files)
    case "$path" in
      *.md|*.txt|*.json) return 0 ;;
    esac
    # .claude/ subdirectory is always excluded
    case "$path" in
      .claude/*) return 0 ;;
    esac
    # docs/, templates/, benchmarks/ are always excluded
    case "$path" in
      docs/*|templates/*|benchmarks/*) return 0 ;;
    esac
    return 1
  }

  # Excluded path check function
  is_excluded_path() {
    local path="$1"
    local policy_file="$2"

    # Check default exclusions first
    is_default_excluded "$path" && return 0

    # If no policy file, finish with defaults only
    [ ! -f "$policy_file" ] && return 1

    if command -v jq >/dev/null 2>&1; then
      # Check skills_gate.exclude_paths
      local exclude_paths
      exclude_paths=$(jq -r '.skills_gate.exclude_paths[]? // empty' "$policy_file" 2>/dev/null)

      while IFS= read -r pattern; do
        [ -z "$pattern" ] && continue
        case "$path" in
          $pattern*) return 0 ;;
        esac
        case "$pattern" in
          \*.*)
            local ext="${pattern#\*}"
            [[ "$path" == *"$ext" ]] && return 0
            ;;
        esac
      done <<< "$exclude_paths"

      # Check exclude_extensions
      local exclude_exts
      exclude_exts=$(jq -r '.skills_gate.exclude_extensions[]? // empty' "$policy_file" 2>/dev/null)
      local file_ext=".${path##*.}"

      while IFS= read -r ext; do
        [ -z "$ext" ] && continue
        [ "$file_ext" = "$ext" ] && return 0
      done <<< "$exclude_exts"
    fi

    return 1
  }

  # ===== Skills Gate: check skill usage per session =====
  # Only apply gate when skills-config.json exists and enabled=true
  if [ -f "$SKILLS_CONFIG_FILE" ]; then
    if command -v jq >/dev/null 2>&1; then
      SKILLS_GATE_ACTIVE=$(jq -r '.enabled // false' "$SKILLS_CONFIG_FILE" 2>/dev/null || echo "false")
      
      if [ "$SKILLS_GATE_ACTIVE" = "true" ]; then
        # Excluded path check
        if is_excluded_path "$REL_PATH" "$SKILLS_POLICY_FILE"; then
          : # Excluded path → skip
        else
          # Check session-skills-used.json
          SKILL_USED_THIS_SESSION="false"
          if [ -f "$SESSION_SKILLS_USED_FILE" ]; then
            USED_COUNT=$(jq -r '.used | length' "$SESSION_SKILLS_USED_FILE" 2>/dev/null || echo "0")
            if [ "$USED_COUNT" -gt 0 ]; then
              SKILL_USED_THIS_SESSION="true"
            fi
          fi
          
          if [ "$SKILL_USED_THIS_SESSION" = "false" ]; then
            # Skill not used → block
            AVAILABLE_SKILLS=$(jq -r '.skills // [] | join(", ")' "$SKILLS_CONFIG_FILE" 2>/dev/null || echo "impl, harness-review")
            DENY_MSG="[Skills Gate] Please use a skill before editing code.

Skills Gate is enabled for this project.
Please invoke an appropriate skill via the Skill tool before making code changes.

Available skills: ${AVAILABLE_SKILLS}

Example: Invoke 'impl' or 'harness-review' via the Skill tool

After using a skill, try Write/Edit again."
            emit_deny "$DENY_MSG"
            exit 0
          fi
        fi
      fi
    fi
  fi

  # ===== LSP Gate: recommend LSP usage for semantic changes =====
  if [ -f "$SESSION_FILE" ] && [ -f "$TOOLING_POLICY_FILE" ]; then
    if command -v jq >/dev/null 2>&1; then
      CURRENT_PROMPT_SEQ=$(jq -r '.prompt_seq // 0' "$SESSION_FILE" 2>/dev/null || echo 0)
      INTENT=$(jq -r '.intent // "literal"' "$SESSION_FILE" 2>/dev/null || echo "literal")
      LSP_AVAILABLE=$(jq -r '.lsp.available // false' "$TOOLING_POLICY_FILE" 2>/dev/null || echo false)
      LSP_LAST_USED_SEQ=$(jq -r '.lsp.last_used_prompt_seq // 0' "$TOOLING_POLICY_FILE" 2>/dev/null || echo 0)

      FILE_EXT="${FILE_PATH##*.}"
      LSP_AVAILABLE_FOR_EXT=$(jq -r ".lsp.available_by_ext[\"$FILE_EXT\"] // false" "$TOOLING_POLICY_FILE" 2>/dev/null || echo false)

      if [ "$INTENT" = "semantic" ] && [ "$LSP_AVAILABLE" = "true" ] && [ "$LSP_AVAILABLE_FOR_EXT" = "true" ]; then
        if [ "$LSP_LAST_USED_SEQ" != "$CURRENT_PROMPT_SEQ" ]; then
          DENY_MSG="[LSP Policy] Please analyze the impact scope using LSP tools before making code changes.

Recommended LSP tools:
- Go-to-definition to check symbol definitions
- Find-references to check usage locations
- Diagnostics to detect type errors

After understanding the impact scope with LSP tools, try Write/Edit again."
          emit_deny "$DENY_MSG"
          exit 0
        fi
      fi
    fi
  fi

  # ===== additionalContext output (Claude Code v2.1.9+) =====
  # Return file path-based guidelines when all guards pass
  GUIDELINE="$(get_guideline_for_path "$REL_PATH")"
  if [ -n "$GUIDELINE" ]; then
    emit_approve_with_context "$GUIDELINE"
  fi

  exit 0
fi


if [ "$TOOL_NAME" = "Bash" ]; then
  [ -z "$COMMAND" ] && exit 0

  if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])sudo([[:space:]]|$)'; then
    emit_deny "$(msg deny_sudo)"
    exit 0
  fi

  # ===== Breezing Role Guard: Bash command restrictions =====
  if [ -n "$BREEZING_ROLE" ]; then
    # Reviewer: block write-type Bash commands
    if [ "$BREEZING_ROLE" = "reviewer" ]; then
      # Read-only commands (cat, grep, ls, git status/diff/log, echo) are allowed
      # Write-type commands (redirect, sed -i, tee, mv, cp, rm, git commit/push) are blocked
      # 2>&1 (stderr→stdout) is read-safe and excluded
      BREEZING_SANITIZED_CMD=$(echo "$COMMAND" | sed 's/2>&1//g; s/>&2//g')
      if echo "$BREEZING_SANITIZED_CMD" | grep -Eq '(>|>>|2>|&>|(^|[[:space:]])tee([[:space:]]|$)|sed[[:space:]]+-i)'; then
        emit_deny "[Breezing] Reviewer cannot execute write-type Bash commands."
        exit 0
      fi
      if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])(mv|cp|rm|mkdir|touch)[[:space:]]'; then
        emit_deny "[Breezing] Reviewer cannot execute file operation commands."
        exit 0
      fi
      if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])git[[:space:]]+(commit|push|add|checkout|reset|rebase|merge|cherry-pick)([[:space:]]|$)'; then
        emit_deny "[Breezing] Reviewer cannot execute git mutation commands."
        exit 0
      fi
    fi

    # Implementer: block git commit (only Lead can commit)
    if [ "$BREEZING_ROLE" = "implementer" ]; then
      if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])git[[:space:]]+commit([[:space:]]|$)'; then
        emit_deny "[Breezing] Implementer cannot execute git commit. Commits are batch-executed by Lead at the completion stage."
        exit 0
      fi
      if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])git[[:space:]]+push([[:space:]]|$)'; then
        emit_deny "[Breezing] Implementer cannot execute git push."
        exit 0
      fi
    fi
  fi

  # ===== Breezing-Codex Mode: restrict write-type Bash commands =====
  if [ "$BREEZING_CODEX_MODE" = "true" ]; then
    # Block redirect and in-place editing (2>&1 is read-safe and excluded)
    BREEZING_CODEX_SANITIZED_CMD=$(echo "$COMMAND" | sed 's/2>&1//g; s/>&2//g')
    if echo "$BREEZING_CODEX_SANITIZED_CMD" | grep -Eq '(>|>>|2>|&>|(^|[[:space:]])tee([[:space:]]|$)|sed[[:space:]]+-i|awk[[:space:]]+-i[[:space:]]+inplace)'; then
      emit_deny "$(msg deny_breezing_codex_mode)"
      exit 0
    fi
    # Block file operation commands
    if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])(mv|cp|rm|mkdir|touch)[[:space:]]'; then
      emit_deny "[Breezing-Codex] File operation commands are prohibited in codex impl mode."
      exit 0
    fi
    # Block git mutation commands
    if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])git[[:space:]]+(commit|push|add|checkout|reset|rebase|merge|cherry-pick|apply|am|switch|restore|stash|pull|clean|rm|mv|submodule)([[:space:]]|$)'; then
      emit_deny "[Breezing-Codex] Git mutation commands are prohibited in codex impl mode."
      exit 0
    fi
  fi

  # ===== Codex Mode: PM is restricted from write commands in Bash =====
  if [ "$CODEX_MODE" = "true" ]; then
    # Detect write operation patterns
    # - Redirect: >, >>, 2>, &>
    # - tee command
    # - sed -i (in-place edit)
    # - awk -i inplace
    # Note: Read-only commands (cat, grep, ls, git status, etc.) are allowed
    # Note: rm is excluded here as it is handled by the rm -rf allowlist below
    # 2>&1 (stderr→stdout) is read-safe and excluded
    CODEX_SANITIZED_CMD=$(echo "$COMMAND" | sed 's/2>&1//g; s/>&2//g')
    if echo "$CODEX_SANITIZED_CMD" | grep -Eq '(>|>>|2>|&>|(^|[[:space:]])tee([[:space:]]|$)|sed[[:space:]]+-i|awk[[:space:]]+-i[[:space:]]+inplace)'; then
      emit_deny "$(msg deny_codex_mode)"
      exit 0
    fi
    # mv, cp require confirmation (ask)
    # rm is handled by the rm -rf allowlist (avoids ordering issues)
    if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])(mv|cp)[[:space:]]'; then
      emit_ask "[Codex Mode] Execute file operations (mv/cp) in PM mode? Delegating to Codex Worker is recommended."
      exit 0
    fi
  fi

  # ===== Commit Guard: Block commits before review completion =====
  if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])git[[:space:]]+commit([[:space:]]|$)'; then
    REVIEW_STATE_FILE=".claude/state/review-approved.json"
    REVIEW_RESULT_FILE=".claude/state/review-result.json"
    COMMIT_GUARD_ENABLED="true"

    # Check if disabled in configuration file
    CONFIG_FILE=".claude-code-harness.config.yaml"
    if [ -f "$CONFIG_FILE" ] && command -v grep >/dev/null 2>&1; then
      if grep -q "commit_guard:[[:space:]]*false" "$CONFIG_FILE" 2>/dev/null; then
        COMMIT_GUARD_ENABLED="false"
      fi
    fi

    if [ "$COMMIT_GUARD_ENABLED" = "true" ]; then
      # Check review approval status
      REVIEW_APPROVED="false"
      if command -v jq >/dev/null 2>&1; then
        if [ -f "$REVIEW_RESULT_FILE" ]; then
          RESULT_VERDICT=$(jq -r '.verdict // empty' "$REVIEW_RESULT_FILE" 2>/dev/null)
          if [ "$RESULT_VERDICT" = "APPROVE" ]; then
            REVIEW_APPROVED="true"
          fi
        fi

        if [ "$REVIEW_APPROVED" = "false" ] && [ -f "$REVIEW_STATE_FILE" ]; then
          APPROVED_AT=$(jq -r '.approved_at // empty' "$REVIEW_STATE_FILE" 2>/dev/null)
          JUDGMENT=$(jq -r '.judgment // empty' "$REVIEW_STATE_FILE" 2>/dev/null)
          if [ -n "$APPROVED_AT" ] && [ "$JUDGMENT" = "APPROVE" ]; then
            REVIEW_APPROVED="true"
          fi
        fi
      fi

      if [ "$REVIEW_APPROVED" = "false" ]; then
        emit_deny "$(msg deny_git_commit_no_review)"
        exit 0
      fi

      # Clear approval status after commit (require re-review before next commit)
      # Note: This should be done in PostToolUse, but here we only warn
    fi
  fi

  if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])git[[:space:]]+push([[:space:]]|$)'; then
    # Can be bypassed during Work mode
    if [ "$WORK_MODE" = "true" ] && [ "$WORK_BYPASS_GIT_PUSH" = "true" ]; then
      : # Skip (auto-approved)
    else
      emit_ask "$(msg ask_git_push "$COMMAND")"
      exit 0
    fi
  fi

  # Detect dangerous recursive rm patterns
  # Note: Only rm -rf / rm -r -f are bypass targets. Other flag combinations require confirmation
  if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])rm[[:space:]]+-[a-z]*r[a-z]*[[:space:]]' || \
     echo "$COMMAND" | grep -Eiq '(^|[[:space:]])rm[[:space:]]+--recursive'; then

    # ===== Work allowlist approach (Codex approved) =====
    # Default: require confirmation
    RM_AUTO_APPROVE="false"

    # Only check if Work mode is active and rm_rf bypass is allowed
    if [ "$WORK_MODE" = "true" ] && [ "$WORK_BYPASS_RM_RF" = "true" ]; then

      # 0. Only allowed flag forms (rm -rf or rm -r -f)
      # rm -rfv, rm -fr, rm --recursive and other forms require confirmation
      if ! echo "$COMMAND" | grep -Eq '(^|[[:space:]])rm[[:space:]]+(-rf|-r[[:space:]]+-f)[[:space:]]+'; then
        : # Require confirmation (disallowed flag form)
      # 1. Require confirmation if dangerous shell syntax is present (* ? $ ( ) { } ; | & < > \ `)
      elif echo "$COMMAND" | grep -Eq '[\*\?\$\(\)\{\};|&<>\\`]'; then
        : # Require confirmation
      # 2. Require confirmation if sudo/xargs/find is present
      elif echo "$COMMAND" | grep -Eiq '(sudo|xargs|find)[[:space:]]'; then
        : # Require confirmation
      else
        # Extract rm targets (strip flag portion)
        RM_TARGET=$(echo "$COMMAND" | sed -E 's/^.*rm[[:space:]]+(-rf|-r[[:space:]]+-f)[[:space:]]+//' | sed 's/[[:space:]].*//')

        # 3. Single target check (ensure no multiple targets separated by spaces)
        RM_TARGET_COUNT=$(echo "$COMMAND" | sed -E 's/^.*rm[[:space:]]+(-rf|-fr|-r[[:space:]]+-f|-f[[:space:]]+-r)[[:space:]]+//' | wc -w | tr -d ' ')
        if [ "$RM_TARGET_COUNT" -eq 1 ]; then

          # 4. Relative paths only (must not start with / or ~)
          # 5. No parent references (must not contain ..)
          # 6. No trailing slash
          # 7. No path separators (only basename allowed)
          # 8. Must not contain . or //
          case "$RM_TARGET" in
            /*|~*|*..*)
              : # Require confirmation
              ;;
            */)
              : # Require confirmation (trailing slash)
              ;;
            *//*|*/.*)
              : # Require confirmation (contains // or /.)
              ;;
            */*)
              : # Require confirmation (contains path separator)
              ;;
            .)
              : # Require confirmation (current directory)
              ;;
            *)
              # 9. Protected path check
              case "$RM_TARGET" in
                .git*|.env*|*secrets*|*keys*|*.pem|*.key|*id_rsa*|*id_ed25519*|.ssh*|.npmrc*|.aws*|.gitmodules*)
                  : # Require confirmation (protected path)
                  ;;
                *)
                  # 10. Allowlist check
                  if [ -n "$CWD" ]; then
                    WORK_FILE="$CWD/.claude/state/work-active.json"
                    # Backward compat: try ultrawork-active.json if work-active.json not found
                    if [ ! -f "$WORK_FILE" ]; then
                      WORK_FILE="$CWD/.claude/state/ultrawork-active.json"
                    fi
                    if [ -f "$WORK_FILE" ] && command -v jq >/dev/null 2>&1; then
                      # Get allowlist from allowed_rm_paths
                      ALLOWED_PATHS=$(jq -r '.allowed_rm_paths[]? // empty' "$WORK_FILE" 2>/dev/null)
                      if [ -n "$ALLOWED_PATHS" ]; then
                        while IFS= read -r ALLOWED; do
                          if [ "$RM_TARGET" = "$ALLOWED" ]; then
                            RM_AUTO_APPROVE="true"
                            break
                          fi
                        done <<< "$ALLOWED_PATHS"
                      fi
                    fi
                  fi
                  ;;
              esac
              ;;
          esac
        fi
      fi
    fi

    # If not auto-approved, require confirmation
    if [ "$RM_AUTO_APPROVE" != "true" ]; then
      # Add PM-specific message in Codex mode
      if [ "$CODEX_MODE" = "true" ]; then
        emit_ask "[Codex Mode] Execute rm -rf in PM mode? Delegating to Codex Worker is recommended. ($COMMAND)"
      else
        emit_ask "$(msg ask_rm_rf "$COMMAND")"
      fi
      exit 0
    fi
    # else: auto-approved (pass through without output)
  fi

  # ===== Codex Mode: Confirm simple rm (without -r) too =====
  if [ "$CODEX_MODE" = "true" ]; then
    if echo "$COMMAND" | grep -Eiq '(^|[[:space:]])rm[[:space:]]'; then
      emit_ask "[Codex Mode] Execute rm in PM mode? Delegating to Codex Worker is recommended."
      exit 0
    fi
  fi

  exit 0
fi

exit 0
