#!/bin/bash

set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG="${PROJECT_ROOT}/harness/.claude-code-harness.config.yaml"
TMP_DIR="$(mktemp -d "/tmp/harness-test.XXXXXX")"

# Backup original config; restore on exit
BACKUP="${TMP_DIR}/config.yaml.bak"
cp "$CONFIG" "$BACKUP"
trap 'cp "$BACKUP" "$CONFIG"; rm -rf "$TMP_DIR"' EXIT

# ---------------------------------------------------------------------------
# Test 1 — Config block is present and valid
# ---------------------------------------------------------------------------
echo "Test 1: advisor config block present and all required fields exist"

grep 'advisor:' "$CONFIG" >/dev/null || {
  echo "FAIL: 'advisor:' block not found in config" >&2
  exit 1
}

for field in enabled mode max_consults_per_task retry_threshold consult_before_user_escalation model_defaults; do
  grep "${field}:" "$CONFIG" >/dev/null || {
    echo "FAIL: required field '${field}' not found in advisor config" >&2
    exit 1
  }
done

echo "  ok: advisor block present with all required fields"

# ---------------------------------------------------------------------------
# Test 2 — advisor.enabled: false disables consultation
# ---------------------------------------------------------------------------
echo "Test 2: advisor.enabled: false disables consultation"

# Write a minimal config with advisor disabled in place of the real one
cat > "$CONFIG" <<'YAML'
advisor:
  enabled: false
  mode: on-demand
  max_consults_per_task: 3
  retry_threshold: 2
  consult_before_user_escalation: true
  model_defaults:
    claude: opus
YAML

# run-advisor-consultation.sh reads CONFIG from PLUGIN_DIR at runtime;
# with enabled: false it should print "advisor disabled" to stderr and exit 0.
OUTPUT=$(
  bash "${PROJECT_ROOT}/harness/scripts/run-advisor-consultation.sh" \
    --task-id x \
    --reason-code repeated_failure \
    --error-sig "e" 2>&1 || true
)

# Restore real config immediately so subsequent tests use the correct values
cp "$BACKUP" "$CONFIG"

echo "$OUTPUT" | grep -qi "advisor disabled" || {
  echo "FAIL: expected 'advisor disabled' in output, got: ${OUTPUT}" >&2
  exit 1
}

echo "  ok: advisor disabled message emitted when enabled: false"

# ---------------------------------------------------------------------------
# Test 3 — retry_threshold is a positive integer
# ---------------------------------------------------------------------------
echo "Test 3: retry_threshold is a readable positive integer"

RETRY_THRESHOLD=$(grep 'retry_threshold:' "$CONFIG" | awk '{print $2}' | head -1)

[[ "$RETRY_THRESHOLD" =~ ^[0-9]+$ ]] || {
  echo "FAIL: retry_threshold '${RETRY_THRESHOLD}' is not a non-negative integer" >&2
  exit 1
}

[ "$RETRY_THRESHOLD" -gt 0 ] || {
  echo "FAIL: retry_threshold '${RETRY_THRESHOLD}' must be a positive integer (> 0)" >&2
  exit 1
}

echo "  ok: retry_threshold=${RETRY_THRESHOLD} is a positive integer"

# ---------------------------------------------------------------------------
# Test 4 — max_consults_per_task is a positive integer >= 1
# ---------------------------------------------------------------------------
echo "Test 4: max_consults_per_task is a readable positive integer >= 1"

MAX_CONSULTS=$(grep 'max_consults_per_task:' "$CONFIG" | awk '{print $2}' | head -1)

[[ "$MAX_CONSULTS" =~ ^[0-9]+$ ]] || {
  echo "FAIL: max_consults_per_task '${MAX_CONSULTS}' is not a non-negative integer" >&2
  exit 1
}

[ "$MAX_CONSULTS" -ge 1 ] || {
  echo "FAIL: max_consults_per_task '${MAX_CONSULTS}' must be >= 1" >&2
  exit 1
}

echo "  ok: max_consults_per_task=${MAX_CONSULTS} is a positive integer >= 1"

echo ""
echo "test-advisor-config: ok"
