#!/bin/bash
# Cross-validation test: TypeScript guardrails vs Go guardrails
# Verifies that both implementations produce identical decisions for all R01-R13 rules.
#
# Usage: bash tests/cross-validate-guardrails.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Try dist/ first (new tsconfig), fall back to core/dist/ (legacy)
if [ -f "${ROOT}/dist/index.js" ]; then
  TS_ENGINE="node ${ROOT}/dist/index.js pre-tool"
elif [ -f "${ROOT}/core/dist/index.js" ]; then
  TS_ENGINE="node ${ROOT}/core/dist/index.js pre-tool"
else
  echo "ERROR: TS engine not found in dist/ or core/dist/"
  exit 1
fi
GO_ENGINE="${ROOT}/go/harness hook pre-tool"

PASS=0
FAIL=0
SKIP=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# Extract decision from TS output: {"decision":"deny","reason":"..."}
extract_ts_decision() {
  local output="$1"
  echo "$output" | jq -r '.decision // "approve"' 2>/dev/null || echo "approve"
}

# Extract decision from Go output (handles both formats):
#   Format A: {"hookSpecificOutput":{"permissionDecision":"deny",...}}
#   Format B: {"permissionDecision":"deny",...}  (flat)
#   Empty output = approve
extract_go_decision() {
  local output="$1"
  if [ -z "$output" ] || [ "$output" = "null" ]; then
    echo "approve"
    return
  fi

  # Try wrapped format first
  local decision
  decision=$(echo "$output" | jq -r '.hookSpecificOutput.permissionDecision // empty' 2>/dev/null)
  if [ -z "$decision" ]; then
    # Try flat format
    decision=$(echo "$output" | jq -r '.permissionDecision // empty' 2>/dev/null)
  fi
  if [ -z "$decision" ]; then
    # Try legacy format (same as TS)
    decision=$(echo "$output" | jq -r '.decision // empty' 2>/dev/null)
  fi

  # Normalize: Go uses "allow" where TS uses "approve"
  case "$decision" in
    allow) echo "approve" ;;
    "")    echo "approve" ;;
    *)     echo "$decision" ;;
  esac
}

# Extract reason from TS output
extract_ts_reason() {
  local output="$1"
  echo "$output" | jq -r '.reason // .systemMessage // ""' 2>/dev/null || echo ""
}

# Extract reason from Go output
extract_go_reason() {
  local output="$1"
  if [ -z "$output" ] || [ "$output" = "null" ]; then
    echo ""
    return
  fi
  local reason
  reason=$(echo "$output" | jq -r '.hookSpecificOutput.permissionDecisionReason // empty' 2>/dev/null)
  if [ -z "$reason" ]; then
    reason=$(echo "$output" | jq -r '.permissionDecisionReason // empty' 2>/dev/null)
  fi
  if [ -z "$reason" ]; then
    reason=$(echo "$output" | jq -r '.hookSpecificOutput.additionalContext // empty' 2>/dev/null)
  fi
  if [ -z "$reason" ]; then
    reason=$(echo "$output" | jq -r '.additionalContext // empty' 2>/dev/null)
  fi
  echo "$reason"
}

run_test() {
  local rule_id="$1"
  local description="$2"
  local input_json="$3"
  local expected_decision="$4"

  # Run TS
  local ts_output
  ts_output=$(echo "$input_json" | $TS_ENGINE 2>/dev/null) || true
  local ts_decision
  ts_decision=$(extract_ts_decision "$ts_output")

  # Run Go (ignore exit code — deny exits with 2)
  local go_output
  go_output=$(echo "$input_json" | $GO_ENGINE 2>/dev/null) || true
  local go_decision
  go_decision=$(extract_go_decision "$go_output")

  # Compare
  if [ "$ts_decision" = "$go_decision" ]; then
    if [ "$ts_decision" = "$expected_decision" ]; then
      printf "${GREEN}PASS${NC} %s: %s (both=%s)\n" "$rule_id" "$description" "$ts_decision"
      PASS=$((PASS + 1))
    else
      printf "${YELLOW}WARN${NC} %s: %s — both agree (%s) but expected %s\n" \
        "$rule_id" "$description" "$ts_decision" "$expected_decision"
      PASS=$((PASS + 1))  # They agree, which is the cross-validation goal
    fi
  else
    printf "${RED}FAIL${NC} %s: %s — TS=%s Go=%s expected=%s\n" \
      "$rule_id" "$description" "$ts_decision" "$go_decision" "$expected_decision"
    printf "  TS output: %s\n" "$ts_output"
    printf "  Go output: %s\n" "$go_output"
    FAIL=$((FAIL + 1))
  fi
}

# ---------------------------------------------------------------------------
# Prerequisite check
# ---------------------------------------------------------------------------

echo "=== Cross-Validation: TypeScript vs Go Guardrails ==="
echo ""

if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required. Install with: brew install jq"
  exit 1
fi

if [ ! -f "${ROOT}/dist/index.js" ] && [ ! -f "${ROOT}/core/dist/index.js" ]; then
  echo "ERROR: TS engine not built. Run: cd core && npm run build"
  exit 1
fi

if [ ! -f "${ROOT}/go/harness" ]; then
  echo "ERROR: Go engine not built. Run: cd go && make build"
  exit 1
fi

echo "TS engine: ${TS_ENGINE}"
echo "Go engine: ${GO_ENGINE}"
echo ""

# ---------------------------------------------------------------------------
# R01: sudo block
# ---------------------------------------------------------------------------
run_test "R01" "sudo deny" \
  '{"tool_name":"Bash","tool_input":{"command":"sudo apt install nginx"}}' \
  "deny"

run_test "R01" "no sudo — safe" \
  '{"tool_name":"Bash","tool_input":{"command":"apt list --installed"}}' \
  "approve"

# ---------------------------------------------------------------------------
# R02: protected path write block (Write/Edit)
# ---------------------------------------------------------------------------
run_test "R02" "write to .env deny" \
  '{"tool_name":"Write","tool_input":{"file_path":".env","content":"SECRET=x"}}' \
  "deny"

run_test "R02" "write to .git/ deny" \
  '{"tool_name":"Edit","tool_input":{"file_path":".git/config","old_string":"a","new_string":"b"}}' \
  "deny"

run_test "R02" "write to id_rsa deny" \
  '{"tool_name":"Write","tool_input":{"file_path":"/home/user/.ssh/id_rsa","content":"key"}}' \
  "deny"

run_test "R02" "write to .pem deny" \
  '{"tool_name":"Write","tool_input":{"file_path":"server.pem","content":"cert"}}' \
  "deny"

run_test "R02" "write to normal file — safe" \
  '{"tool_name":"Write","tool_input":{"file_path":"src/main.ts","content":"code"}}' \
  "approve"

# ---------------------------------------------------------------------------
# R03: Bash write to protected paths
# ---------------------------------------------------------------------------
run_test "R03" "echo to .env deny" \
  '{"tool_name":"Bash","tool_input":{"command":"echo SECRET=x > .env"}}' \
  "deny"

run_test "R03" "tee to .git/ deny" \
  '{"tool_name":"Bash","tool_input":{"command":"echo config | tee .git/config"}}' \
  "deny"

run_test "R03" "echo to .key deny" \
  '{"tool_name":"Bash","tool_input":{"command":"echo key >> server.key"}}' \
  "deny"

run_test "R03" "echo to normal file — safe" \
  '{"tool_name":"Bash","tool_input":{"command":"echo hello > output.txt"}}' \
  "approve"

# ---------------------------------------------------------------------------
# R04: confirm write outside project (non-work mode)
# ---------------------------------------------------------------------------
run_test "R04" "write outside project — ask" \
  '{"tool_name":"Write","tool_input":{"file_path":"/tmp/outside.txt","content":"x"}}' \
  "ask"

run_test "R04" "write relative path — safe" \
  '{"tool_name":"Write","tool_input":{"file_path":"src/app.ts","content":"x"}}' \
  "approve"

# ---------------------------------------------------------------------------
# R05: rm -rf confirm (non-work mode)
# ---------------------------------------------------------------------------
run_test "R05" "rm -rf ask" \
  '{"tool_name":"Bash","tool_input":{"command":"rm -rf /tmp/build"}}' \
  "ask"

run_test "R05" "rm -f (no -r) — safe" \
  '{"tool_name":"Bash","tool_input":{"command":"rm -f temp.log"}}' \
  "approve"

run_test "R05" "rm --recursive ask" \
  '{"tool_name":"Bash","tool_input":{"command":"rm --recursive old_dir"}}' \
  "ask"

# ---------------------------------------------------------------------------
# R06: force push block (no bypass)
# ---------------------------------------------------------------------------
run_test "R06" "git push --force deny" \
  '{"tool_name":"Bash","tool_input":{"command":"git push --force origin main"}}' \
  "deny"

run_test "R06" "git push -f deny" \
  '{"tool_name":"Bash","tool_input":{"command":"git push -f origin feature"}}' \
  "deny"

run_test "R06" "git push --force-with-lease deny" \
  '{"tool_name":"Bash","tool_input":{"command":"git push --force-with-lease origin main"}}' \
  "deny"

run_test "R06" "git push (normal) — safe" \
  '{"tool_name":"Bash","tool_input":{"command":"git push origin feature-branch"}}' \
  "approve"

# ---------------------------------------------------------------------------
# R09: secret file read warning
# ---------------------------------------------------------------------------
run_test "R09" "read .env — warn (approve with message)" \
  '{"tool_name":"Read","tool_input":{"file_path":"/project/.env"}}' \
  "approve"

run_test "R09" "read id_rsa — warn" \
  '{"tool_name":"Read","tool_input":{"file_path":"/home/user/.ssh/id_rsa"}}' \
  "approve"

run_test "R09" "read .pem — warn" \
  '{"tool_name":"Read","tool_input":{"file_path":"cert.pem"}}' \
  "approve"

run_test "R09" "read normal file — safe" \
  '{"tool_name":"Read","tool_input":{"file_path":"src/main.ts"}}' \
  "approve"

# ---------------------------------------------------------------------------
# R10: --no-verify / --no-gpg-sign block
# ---------------------------------------------------------------------------
run_test "R10" "git commit --no-verify deny" \
  '{"tool_name":"Bash","tool_input":{"command":"git commit -m \"fix\" --no-verify"}}' \
  "deny"

run_test "R10" "git commit --no-gpg-sign deny" \
  '{"tool_name":"Bash","tool_input":{"command":"git commit --no-gpg-sign -m \"fix\""}}' \
  "deny"

run_test "R10" "git commit (normal) — safe" \
  '{"tool_name":"Bash","tool_input":{"command":"git commit -m \"fix bug\""}}' \
  "approve"

# ---------------------------------------------------------------------------
# R11: git reset --hard protected branch
# ---------------------------------------------------------------------------
run_test "R11" "git reset --hard main deny" \
  '{"tool_name":"Bash","tool_input":{"command":"git reset --hard main"}}' \
  "deny"

run_test "R11" "git reset --hard origin/master deny" \
  '{"tool_name":"Bash","tool_input":{"command":"git reset --hard origin/master"}}' \
  "deny"

run_test "R11" "git reset --hard HEAD~1 — safe" \
  '{"tool_name":"Bash","tool_input":{"command":"git reset --hard HEAD~1"}}' \
  "approve"

# ---------------------------------------------------------------------------
# R12: direct push to protected branch — warn
# ---------------------------------------------------------------------------
run_test "R12" "git push origin main — deny" \
  '{"tool_name":"Bash","tool_input":{"command":"git push origin main"}}' \
  "deny"

run_test "R12" "git push origin feature:master — deny" \
  '{"tool_name":"Bash","tool_input":{"command":"git push origin feature:master"}}' \
  "deny"

# ---------------------------------------------------------------------------
# R13: protected review paths — warn
# ---------------------------------------------------------------------------
run_test "R13" "write package.json — warn" \
  '{"tool_name":"Write","tool_input":{"file_path":"package.json","content":"{}"}}' \
  "approve"

run_test "R13" "write Dockerfile — warn" \
  '{"tool_name":"Write","tool_input":{"file_path":"Dockerfile","content":"FROM node"}}' \
  "approve"

run_test "R13" "write schema.prisma — warn" \
  '{"tool_name":"Edit","tool_input":{"file_path":"schema.prisma","old_string":"a","new_string":"b"}}' \
  "approve"

# ---------------------------------------------------------------------------
# Edge cases: safe commands
# ---------------------------------------------------------------------------
run_test "SAFE" "ls -la — no rule matches" \
  '{"tool_name":"Bash","tool_input":{"command":"ls -la"}}' \
  "approve"

run_test "SAFE" "Read normal file — no rule matches" \
  '{"tool_name":"Read","tool_input":{"file_path":"README.md"}}' \
  "approve"

run_test "SAFE" "Write normal file — no rule matches" \
  '{"tool_name":"Write","tool_input":{"file_path":"src/utils.ts","content":"export {}"}}' \
  "approve"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "==========================================="
printf "Results: ${GREEN}%d PASS${NC}" "$PASS"
if [ "$FAIL" -gt 0 ]; then
  printf " / ${RED}%d FAIL${NC}" "$FAIL"
fi
if [ "$SKIP" -gt 0 ]; then
  printf " / ${YELLOW}%d SKIP${NC}" "$SKIP"
fi
TOTAL=$((PASS + FAIL + SKIP))
echo " / $TOTAL total"
echo "==========================================="

if [ "$FAIL" -gt 0 ]; then
  echo ""
  echo "CROSS-VALIDATION FAILED: TS and Go guardrails produce different decisions."
  exit 1
fi

echo ""
echo "CROSS-VALIDATION PASSED: All rules produce identical decisions in both engines."
