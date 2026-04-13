#!/bin/bash
# posttooluse-tampering-detector.sh
# Detect test tampering patterns and warn (does not block)
#
# Purpose: Run after Write|Edit in PostToolUse
# Behavior:
#   - Monitor changes to test files (*.test.*, *.spec.*)
#   - Detect tampering patterns (skipping, assertion deletion, eslint-disable)
#   - Output a warning as additionalContext when detected
#   - Record to log (.claude/state/tampering.log)
#
# Output: Outputs warning in hookSpecificOutput.additionalContext as JSON
#       → Claude Code displays it as a system-reminder

set +e

# ===== Read input =====
INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null || true)"
fi

[ -z "$INPUT" ] && exit 0

# ===== Parse JSON =====
TOOL_NAME=""
FILE_PATH=""
OLD_STRING=""
NEW_STRING=""
CONTENT=""

if command -v jq >/dev/null 2>&1; then
  TOOL_NAME=$(printf '%s' "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null || true)
  FILE_PATH=$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // empty' 2>/dev/null || true)
  OLD_STRING=$(printf '%s' "$INPUT" | jq -r '.tool_input.old_string // empty' 2>/dev/null || true)
  NEW_STRING=$(printf '%s' "$INPUT" | jq -r '.tool_input.new_string // empty' 2>/dev/null || true)
  CONTENT=$(printf '%s' "$INPUT" | jq -r '.tool_input.content // empty' 2>/dev/null || true)
elif command -v python3 >/dev/null 2>&1; then
  eval "$(printf '%s' "$INPUT" | python3 -c '
import json, shlex, sys
try:
    data = json.load(sys.stdin)
except Exception:
    data = {}
tool_name = data.get("tool_name") or ""
tool_input = data.get("tool_input") or {}
file_path = tool_input.get("file_path") or ""
old_string = tool_input.get("old_string") or ""
new_string = tool_input.get("new_string") or ""
content = tool_input.get("content") or ""
print(f"TOOL_NAME={shlex.quote(tool_name)}")
print(f"FILE_PATH={shlex.quote(file_path)}")
print(f"OLD_STRING={shlex.quote(old_string)}")
print(f"NEW_STRING={shlex.quote(new_string)}")
print(f"CONTENT={shlex.quote(content)}")
' 2>/dev/null)"
fi

# Skip if not Write/Edit
[[ "$TOOL_NAME" != "Write" && "$TOOL_NAME" != "Edit" ]] && exit 0

# Skip if file path is missing
[ -z "$FILE_PATH" ] && exit 0

# ===== Test file detection =====
is_test_file() {
  local path="$1"
  case "$path" in
    *.test.ts|*.test.tsx|*.test.js|*.test.jsx) return 0 ;;
    *.spec.ts|*.spec.tsx|*.spec.js|*.spec.jsx) return 0 ;;
    *.test.py|test_*.py|*_test.py) return 0 ;;
    *.test.go|*_test.go) return 0 ;;
    */__tests__/*|*/tests/*) return 0 ;;
  esac
  return 1
}

# lint/CI config file detection
is_config_file() {
  local path="$1"
  case "$path" in
    .eslintrc*|eslint.config.*) return 0 ;;
    .prettierrc*|prettier.config.*) return 0 ;;
    tsconfig.json|tsconfig.*.json) return 0 ;;
    biome.json|.stylelintrc*) return 0 ;;
    jest.config.*|vitest.config.*) return 0 ;;
    .github/workflows/*.yml|.github/workflows/*.yaml) return 0 ;;
    .gitlab-ci.yml|Jenkinsfile) return 0 ;;
  esac
  return 1
}

# Skip if neither a test file nor a config file
if ! is_test_file "$FILE_PATH" && ! is_config_file "$FILE_PATH"; then
  exit 0
fi

# ===== Tampering pattern detection =====
WARNINGS=""

# Content to check
CHECK_CONTENT="${NEW_STRING}${CONTENT}"

# Test file tampering detection
if is_test_file "$FILE_PATH"; then
  # Test skip detected (JS/TS)
  if [[ "$CHECK_CONTENT" =~ (^|[^a-zA-Z_])(it|describe|test)\.skip[[:space:]]*\(|(^|[^a-zA-Z_])xit[[:space:]]*\(|(^|[^a-zA-Z_])xdescribe[[:space:]]*\( ]]; then
    WARNINGS="${WARNINGS}⚠️ Test skip detected (it.skip/describe.skip/xit)\n"
  fi

  # Python test skip detected
  # @pytest.mark.skip, @pytest.mark.skipIf, @unittest.skip, @unittest.skipIf, self.skipTest()
  if [[ "$CHECK_CONTENT" =~ @pytest\.mark\.skip|@unittest\.skip|self\.skipTest[[:space:]]*\( ]]; then
    WARNINGS="${WARNINGS}⚠️ Python test skip detected (@pytest.mark.skip / @unittest.skip / self.skipTest)\n"
  fi

  # Test .only detected (.only conversion)
  if [[ "$CHECK_CONTENT" =~ (^|[^a-zA-Z_])(it|describe|test)\.only[[:space:]]*\(|(^|[^a-zA-Z_])fit[[:space:]]*\(|(^|[^a-zA-Z_])fdescribe[[:space:]]*\( ]]; then
    WARNINGS="${WARNINGS}⚠️ Test .only detected (other tests will no longer run)\n"
  fi

  # Lint/type suppression detected
  if [[ "$CHECK_CONTENT" =~ eslint-disable|@ts-ignore|@ts-expect-error|@ts-nocheck ]]; then
    WARNINGS="${WARNINGS}⚠️ Lint/type suppression detected (eslint-disable/@ts-ignore/@ts-nocheck)\n"
  fi

  # Assertion removal detected (for Edit, JS/TS)
  if [ -n "$OLD_STRING" ] && [ -n "$NEW_STRING" ]; then
    OLD_EXPECTS=$(printf '%s' "$OLD_STRING" | grep -c 'expect\s*(' || true)
    NEW_EXPECTS=$(printf '%s' "$NEW_STRING" | grep -c 'expect\s*(' || true)
    if [ "$OLD_EXPECTS" -gt 0 ] && [ "$NEW_EXPECTS" -lt "$OLD_EXPECTS" ]; then
      WARNINGS="${WARNINGS}⚠️ Assertion removal detected (expect: ${OLD_EXPECTS} → ${NEW_EXPECTS})\n"
    fi
  fi

  # Assertion removal detected (Python)
  if [ -n "$OLD_STRING" ] && [ -n "$NEW_STRING" ]; then
    OLD_ASSERTS=$(printf '%s' "$OLD_STRING" | grep -cE '\bassert\b|self\.assert' || true)
    NEW_ASSERTS=$(printf '%s' "$NEW_STRING" | grep -cE '\bassert\b|self\.assert' || true)
    if [ "$OLD_ASSERTS" -gt 0 ] && [ "$NEW_ASSERTS" -lt "$OLD_ASSERTS" ]; then
      WARNINGS="${WARNINGS}⚠️ Assertion removal detected (assert: ${OLD_ASSERTS} → ${NEW_ASSERTS})\n"
    fi
  fi

  # Assertion weakening detected (for Edit)
  # Detect replacement with weaker assertions like toBe → toBeTruthy/toBeDefined/toBeUndefined/toBeNull/toBeFalsy
  if [ -n "$OLD_STRING" ] && [ -n "$NEW_STRING" ]; then
    # Check if OLD has strict assertions that were replaced by weaker ones in NEW
    OLD_STRICT=$(printf '%s' "$OLD_STRING" | grep -cE '\.toBe\(|\.toEqual\(|\.toStrictEqual\(|\.toHaveBeenCalledWith\(' || true)
    NEW_WEAK=$(printf '%s' "$NEW_STRING" | grep -cE '\.toBeTruthy\(|\.toBeDefined\(|\.toBeUndefined\(|\.toBeNull\(|\.toBeFalsy\(|\.toBeGreaterThanOrEqual\(0\)|\.toHaveBeenCalled\(\)' || true)
    NEW_STRICT=$(printf '%s' "$NEW_STRING" | grep -cE '\.toBe\(|\.toEqual\(|\.toStrictEqual\(|\.toHaveBeenCalledWith\(' || true)
    # Warn when strict assertions decrease and weak assertions increase
    if [ "$OLD_STRICT" -gt 0 ] && [ "$NEW_STRICT" -lt "$OLD_STRICT" ] && [ "$NEW_WEAK" -gt 0 ]; then
      WARNINGS="${WARNINGS}⚠️ Assertion weakening detected (strict: ${OLD_STRICT} → ${NEW_STRICT}, weak: +${NEW_WEAK}). e.g. toBe → toBeTruthy, toEqual → toBeDefined\n"
    fi
  fi

  # Large timeout increase detected
  # Detect large values (30000ms or more) for jest.setTimeout(N), jasmine.DEFAULT_TIMEOUT_INTERVAL = N, { timeout: N }, etc.
  TIMEOUT_THRESHOLD=30000
  TIMEOUT_HIT=$(printf '%s' "$CHECK_CONTENT" | grep -E 'jest\.setTimeout\(|jasmine\.DEFAULT_TIMEOUT_INTERVAL|[[:space:]]timeout[[:space:]]*:' | grep -oE '[0-9]+' | awk -v thr="$TIMEOUT_THRESHOLD" '$1 >= thr {found=1} END {print found+0}' || true)
  if [ "${TIMEOUT_HIT:-0}" -gt 0 ]; then
    WARNINGS="${WARNINGS}⚠️ Large timeout detected (≥${TIMEOUT_THRESHOLD}ms). e.g. jest.setTimeout(30000)\n"
  fi

  # Catch-all assertion detected
  # Detect meaningless assertions that always succeed like expect(true).toBe(true), expect(1).toBe(1)
  if [[ "$CHECK_CONTENT" =~ expect\((true|false|1|0|null|undefined|[\"\']{2})\)\.(toBe|toEqual|toStrictEqual)\((true|false|1|0|null|undefined|[\"\']{2})\) ]]; then
    WARNINGS="${WARNINGS}⚠️ Catch-all assertion detected: meaningless assertion that always succeeds (e.g. expect(true).toBe(true))\n"
  fi

  # Pattern of applying toBeUndefined/toBeNull/toBeFalsy/toBeTruthy to constant values
  if [[ "$CHECK_CONTENT" =~ expect\((true|false|null|undefined|0)\)\.(toBeUndefined|toBeNull|toBeFalsy|toBeTruthy)\(\) ]]; then
    WARNINGS="${WARNINGS}⚠️ Catch-all assertion detected: weak assertion on constant value (e.g. expect(false).toBeFalsy())\n"
  fi
fi

# Config file weakening detection
if is_config_file "$FILE_PATH"; then
  # Lint rule disabled
  if [[ "$CHECK_CONTENT" =~ \"off\"|:[[:space:]]*0|\"warn\".*→.*\"off\" ]]; then
    WARNINGS="${WARNINGS}⚠️ Lint rule disabled: lint rule disable detected\n"
  fi

  # CI continue-on-error detected
  if [[ "$CHECK_CONTENT" =~ continue-on-error:[[:space:]]*true ]]; then
    WARNINGS="${WARNINGS}⚠️ CI continue-on-error detected: continue-on-error added to CI\n"
  fi

  # TypeScript strict mode weakened
  if [[ "$CHECK_CONTENT" =~ \"strict\"[[:space:]]*:[[:space:]]*false|\"noImplicitAny\"[[:space:]]*:[[:space:]]*false ]]; then
    WARNINGS="${WARNINGS}⚠️ TypeScript strict mode weakened: TypeScript strict mode relaxation detected\n"
  fi
fi

# ===== Exit if no warnings =====
[ -z "$WARNINGS" ] && exit 0

# ===== Record to log =====
STATE_DIR=".claude/state"
LOG_FILE="$STATE_DIR/tampering.log"

if [ -d "$STATE_DIR" ] || mkdir -p "$STATE_DIR" 2>/dev/null; then
  echo "[$(date -Iseconds 2>/dev/null || date '+%Y-%m-%dT%H:%M:%S')] FILE=$FILE_PATH TOOL=$TOOL_NAME" >> "$LOG_FILE" 2>/dev/null || true
  printf '%b' "$WARNINGS" | sed 's/^/  /' >> "$LOG_FILE" 2>/dev/null || true
fi

# ===== Output warning =====
# Output as additionalContext so Claude can see it on the next turn
WARNING_MSG="[Tampering Detector] Suspicious patterns detected in test/config file changes:

$(printf '%b' "$WARNINGS")
File: $FILE_PATH

If this is an intentional change, no action is needed. Otherwise, there may be test tampering.

⚠️ Fix the implementation, not the tests.
⚠️ Fix the code, not the config."

# JSON output
if command -v jq >/dev/null 2>&1; then
  jq -nc --arg ctx "$WARNING_MSG" \
    '{hookSpecificOutput:{hookEventName:"PostToolUse",additionalContext:$ctx}}'
else
  # Output with minimal escaping if jq is not available
  ESCAPED_MSG=$(echo "$WARNING_MSG" | sed 's/\\/\\\\/g; s/"/\\"/g; s/$/\\n/' | tr -d '\n' | sed 's/\\n$//')
  echo "{\"hookSpecificOutput\":{\"hookEventName\":\"PostToolUse\",\"additionalContext\":\"${ESCAPED_MSG}\"}}"
fi

exit 0
