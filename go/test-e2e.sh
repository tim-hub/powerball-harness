#!/bin/bash
# Phase 0 E2E test script
set -e
H="$(dirname "$0")/../bin/harness"

echo "=== Version ==="
"$H" version

echo ""
echo "=== R01: sudo deny ==="
result=$(echo '{"tool_name":"Bash","tool_input":{"command":"sudo apt install"}}' | "$H" hook pre-tool 2>/dev/null || true)
echo "$result"
echo "$result" | grep -q '"permissionDecision":"deny"' && echo "PASS" || echo "FAIL"

echo ""
echo "=== R06: force push deny ==="
result=$(echo '{"tool_name":"Bash","tool_input":{"command":"git push --force origin main"}}' | "$H" hook pre-tool 2>/dev/null || true)
echo "$result"
echo "$result" | grep -q '"permissionDecision":"deny"' && echo "PASS" || echo "FAIL"

echo ""
echo "=== Approve: safe command ==="
result=$(echo '{"tool_name":"Bash","tool_input":{"command":"ls -la"}}' | "$H" hook pre-tool 2>/dev/null)
if [ -z "$result" ]; then echo "PASS (empty = approve)"; else echo "FAIL: $result"; fi

echo ""
echo "=== R09: secret file warning ==="
result=$(echo '{"tool_name":"Read","tool_input":{"file_path":"/project/.env"}}' | "$H" hook pre-tool 2>/dev/null)
echo "$result"
echo "$result" | grep -q '"additionalContext"' && echo "PASS" || echo "FAIL"

echo ""
echo "=== PostToolUse: tampering ==="
result=$(echo '{"tool_name":"Write","tool_input":{"file_path":"test.spec.ts","content":"it.skip(\"broken\", () => {})"}}' | "$H" hook post-tool 2>/dev/null)
echo "$result"
echo "$result" | grep -q 'Test tampering' && echo "PASS" || echo "FAIL"

echo ""
echo "=== PermissionRequest: safe git ==="
result=$(echo '{"tool_name":"Bash","tool_input":{"command":"git status"}}' | "$H" hook permission 2>/dev/null)
echo "$result"
echo "$result" | grep -q '"behavior":"allow"' && echo "PASS" || echo "FAIL"

echo ""
echo "=== PermissionRequest: Write auto-allow ==="
result=$(echo '{"tool_name":"Write","tool_input":{"file_path":"src/main.ts"}}' | "$H" hook permission 2>/dev/null)
echo "$result"
echo "$result" | grep -q '"behavior":"allow"' && echo "PASS" || echo "FAIL"

echo ""
echo "=== Empty input: safe fallback ==="
result=$(echo '' | "$H" hook pre-tool 2>/dev/null)
echo "$result"
echo "$result" | grep -q '"decision":"approve"' && echo "PASS" || echo "FAIL"

echo ""
echo "=== All E2E tests done ==="
