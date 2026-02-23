#!/bin/bash
# posttooluse-tampering-detector.sh
# テスト改ざんパターンを検出して警告する（ブロックはしない）
#
# 用途: PostToolUse で Write|Edit 後に実行
# 動作:
#   - テストファイル（*.test.*, *.spec.*）への変更を監視
#   - 改ざんパターン（skip化、アサーション削除、eslint-disable）を検出
#   - 検出した場合は警告を additionalContext として出力
#   - ログに記録（.claude/state/tampering.log）
#
# 出力: JSON形式で hookSpecificOutput.additionalContext に警告を出力
#       → Claude Code が system-reminder として表示

set +e

# ===== 入力の取得 =====
INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null || true)"
fi

[ -z "$INPUT" ] && exit 0

# ===== JSON パース =====
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

# Write/Edit 以外はスキップ
[[ "$TOOL_NAME" != "Write" && "$TOOL_NAME" != "Edit" ]] && exit 0

# ファイルパスがなければスキップ
[ -z "$FILE_PATH" ] && exit 0

# ===== テストファイル判定 =====
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

# lint/CI 設定ファイル判定
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

# テストファイルでも設定ファイルでもなければスキップ
if ! is_test_file "$FILE_PATH" && ! is_config_file "$FILE_PATH"; then
  exit 0
fi

# ===== 改ざんパターン検出 =====
WARNINGS=""

# 検出対象のコンテンツ
CHECK_CONTENT="${NEW_STRING}${CONTENT}"

# テストファイルの改ざん検出
if is_test_file "$FILE_PATH"; then
  # skip 化検出 (JS/TS) / Test skip detected
  if [[ "$CHECK_CONTENT" =~ (^|[^a-zA-Z_])(it|describe|test)\.skip[[:space:]]*\(|(^|[^a-zA-Z_])xit[[:space:]]*\(|(^|[^a-zA-Z_])xdescribe[[:space:]]*\( ]]; then
    WARNINGS="${WARNINGS}⚠️ Test skip detected / テストの skip 化を検出 (it.skip/describe.skip/xit)\n"
  fi

  # skip 化検出 (Python) / Python test skip detected
  # @pytest.mark.skip, @pytest.mark.skipIf, @unittest.skip, @unittest.skipIf, self.skipTest()
  if [[ "$CHECK_CONTENT" =~ @pytest\.mark\.skip|@unittest\.skip|self\.skipTest[[:space:]]*\( ]]; then
    WARNINGS="${WARNINGS}⚠️ Python test skip detected / Python テストの skip 化を検出 (@pytest.mark.skip / @unittest.skip / self.skipTest)\n"
  fi

  # .only 化検出 / Test .only detected
  if [[ "$CHECK_CONTENT" =~ (^|[^a-zA-Z_])(it|describe|test)\.only[[:space:]]*\(|(^|[^a-zA-Z_])fit[[:space:]]*\(|(^|[^a-zA-Z_])fdescribe[[:space:]]*\( ]]; then
    WARNINGS="${WARNINGS}⚠️ Test .only detected / テストの .only 化を検出（他のテストが実行されなくなります）\n"
  fi

  # eslint-disable 追加検出 / Lint/type suppression detected
  if [[ "$CHECK_CONTENT" =~ eslint-disable|@ts-ignore|@ts-expect-error|@ts-nocheck ]]; then
    WARNINGS="${WARNINGS}⚠️ Lint/type suppression detected / lint/型チェック無効化コメントを検出\n"
  fi

  # expect 削除検出（Edit の場合）/ Assertion removal detected
  if [ -n "$OLD_STRING" ] && [ -n "$NEW_STRING" ]; then
    OLD_EXPECTS=$(printf '%s' "$OLD_STRING" | grep -c 'expect\s*(' || true)
    NEW_EXPECTS=$(printf '%s' "$NEW_STRING" | grep -c 'expect\s*(' || true)
    if [ "$OLD_EXPECTS" -gt 0 ] && [ "$NEW_EXPECTS" -lt "$OLD_EXPECTS" ]; then
      WARNINGS="${WARNINGS}⚠️ Assertion removal detected / アサーション削除を検出 (expect: ${OLD_EXPECTS} → ${NEW_EXPECTS})\n"
    fi
  fi

  # assert 削除検出（Python）/ Assertion removal detected
  if [ -n "$OLD_STRING" ] && [ -n "$NEW_STRING" ]; then
    OLD_ASSERTS=$(printf '%s' "$OLD_STRING" | grep -cE '\bassert\b|self\.assert' || true)
    NEW_ASSERTS=$(printf '%s' "$NEW_STRING" | grep -cE '\bassert\b|self\.assert' || true)
    if [ "$OLD_ASSERTS" -gt 0 ] && [ "$NEW_ASSERTS" -lt "$OLD_ASSERTS" ]; then
      WARNINGS="${WARNINGS}⚠️ Assertion removal detected / アサーション削除を検出 (assert: ${OLD_ASSERTS} → ${NEW_ASSERTS})\n"
    fi
  fi

  # assertion weakening 検出（Edit の場合）/ Assertion weakening detected
  # toBe → toBeTruthy/toBeDefined/toBeUndefined/toBeNull/toBeFalsy のような緩いアサーションへの置き換えを検出
  if [ -n "$OLD_STRING" ] && [ -n "$NEW_STRING" ]; then
    # OLD に厳格なアサーションがあり、NEW で弱いアサーションに置き換えられたか確認
    OLD_STRICT=$(printf '%s' "$OLD_STRING" | grep -cE '\.toBe\(|\.toEqual\(|\.toStrictEqual\(|\.toHaveBeenCalledWith\(' || true)
    NEW_WEAK=$(printf '%s' "$NEW_STRING" | grep -cE '\.toBeTruthy\(|\.toBeDefined\(|\.toBeUndefined\(|\.toBeNull\(|\.toBeFalsy\(|\.toBeGreaterThanOrEqual\(0\)|\.toHaveBeenCalled\(\)' || true)
    NEW_STRICT=$(printf '%s' "$NEW_STRING" | grep -cE '\.toBe\(|\.toEqual\(|\.toStrictEqual\(|\.toHaveBeenCalledWith\(' || true)
    # 厳格なアサーションが減り、弱いアサーションが増えた場合に警告
    if [ "$OLD_STRICT" -gt 0 ] && [ "$NEW_STRICT" -lt "$OLD_STRICT" ] && [ "$NEW_WEAK" -gt 0 ]; then
      WARNINGS="${WARNINGS}⚠️ Assertion weakening detected / アサーション弱体化を検出 (strict: ${OLD_STRICT} → ${NEW_STRICT}, weak: +${NEW_WEAK}). e.g. toBe → toBeTruthy, toEqual → toBeDefined\n"
    fi
  fi

  # timeout 値の大幅引き上げ検出 / Large timeout increase detected
  # jest.setTimeout(N), jasmine.DEFAULT_TIMEOUT_INTERVAL = N, { timeout: N } 等の大きな値 (30000ms 以上) を検出
  TIMEOUT_THRESHOLD=30000
  TIMEOUT_HIT=$(printf '%s' "$CHECK_CONTENT" | grep -E 'jest\.setTimeout\(|jasmine\.DEFAULT_TIMEOUT_INTERVAL|[[:space:]]timeout[[:space:]]*:' | grep -oE '[0-9]+' | awk -v thr="$TIMEOUT_THRESHOLD" '$1 >= thr {found=1} END {print found+0}' || true)
  if [ "${TIMEOUT_HIT:-0}" -gt 0 ]; then
    WARNINGS="${WARNINGS}⚠️ Large timeout detected / タイムアウト値の大幅引き上げを検出 (≥${TIMEOUT_THRESHOLD}ms). e.g. jest.setTimeout(30000)\n"
  fi

  # catch-all assertion 検出 / Catch-all assertion detected
  # expect(true).toBe(true), expect(1).toBe(1) 等の常に成功する無意味なアサーションを検出
  if [[ "$CHECK_CONTENT" =~ expect\((true|false|1|0|null|undefined|[\"\']{2})\)\.(toBe|toEqual|toStrictEqual)\((true|false|1|0|null|undefined|[\"\']{2})\) ]]; then
    WARNINGS="${WARNINGS}⚠️ Catch-all assertion detected / 常に成功する無意味なアサーションを検出 (e.g. expect(true).toBe(true))\n"
  fi

  # toBeUndefined/toBeNull/toBeFalsy/toBeTruthy を定数値に適用するパターン
  if [[ "$CHECK_CONTENT" =~ expect\((true|false|null|undefined|0)\)\.(toBeUndefined|toBeNull|toBeFalsy|toBeTruthy)\(\) ]]; then
    WARNINGS="${WARNINGS}⚠️ Catch-all assertion detected / 定数に対する弱いアサーションを検出 (e.g. expect(false).toBeFalsy())\n"
  fi
fi

# 設定ファイルの緩和検出
if is_config_file "$FILE_PATH"; then
  # eslint ルール無効化 / Lint rule disabled
  if [[ "$CHECK_CONTENT" =~ \"off\"|:[[:space:]]*0|\"warn\".*→.*\"off\" ]]; then
    WARNINGS="${WARNINGS}⚠️ Lint rule disabled / lint ルールの無効化を検出\n"
  fi

  # CI continue-on-error / CI continue-on-error detected
  if [[ "$CHECK_CONTENT" =~ continue-on-error:[[:space:]]*true ]]; then
    WARNINGS="${WARNINGS}⚠️ CI continue-on-error detected / CI の continue-on-error 追加を検出\n"
  fi

  # strict モードの緩和 / TypeScript strict mode weakened
  if [[ "$CHECK_CONTENT" =~ \"strict\"[[:space:]]*:[[:space:]]*false|\"noImplicitAny\"[[:space:]]*:[[:space:]]*false ]]; then
    WARNINGS="${WARNINGS}⚠️ TypeScript strict mode weakened / TypeScript strict モードの緩和を検出\n"
  fi
fi

# ===== 警告がなければ終了 =====
[ -z "$WARNINGS" ] && exit 0

# ===== ログに記録 =====
STATE_DIR=".claude/state"
LOG_FILE="$STATE_DIR/tampering.log"

if [ -d "$STATE_DIR" ] || mkdir -p "$STATE_DIR" 2>/dev/null; then
  echo "[$(date -Iseconds 2>/dev/null || date '+%Y-%m-%dT%H:%M:%S')] FILE=$FILE_PATH TOOL=$TOOL_NAME" >> "$LOG_FILE" 2>/dev/null || true
  printf '%b' "$WARNINGS" | sed 's/^/  /' >> "$LOG_FILE" 2>/dev/null || true
fi

# ===== 警告を出力 =====
# Claude が次のターンで見られるように additionalContext として出力
WARNING_MSG="[Tampering Detector] Suspicious patterns detected in test/config file changes:
[Tampering Detector] テスト/設定ファイルの変更で以下のパターンを検出しました：

$(printf '%b' "$WARNINGS")
File / ファイル: $FILE_PATH

If this is an intentional change, no action is needed.
これが意図的な変更であれば問題ありませんが、テスト改ざんの可能性があります。

⚠️ Fix the implementation, not the tests. / テスト改ざん（skip化、アサーション削除）ではなく、実装の修正が正しい対応です。
⚠️ Fix the code, not the config. / 設定の緩和ではなく、コードの修正が正しい対応です。"

# JSON 出力
if command -v jq >/dev/null 2>&1; then
  jq -nc --arg ctx "$WARNING_MSG" \
    '{hookSpecificOutput:{hookEventName:"PostToolUse",additionalContext:$ctx}}'
else
  # jq がない場合は最小限のエスケープで出力
  ESCAPED_MSG=$(echo "$WARNING_MSG" | sed 's/\\/\\\\/g; s/"/\\"/g; s/$/\\n/' | tr -d '\n' | sed 's/\\n$//')
  echo "{\"hookSpecificOutput\":{\"hookEventName\":\"PostToolUse\",\"additionalContext\":\"${ESCAPED_MSG}\"}}"
fi

exit 0
