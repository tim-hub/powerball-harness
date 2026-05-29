#!/bin/bash

set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d "/tmp/harness-test.XXXXXX")"
trap 'rm -rf "$TMP_DIR"' EXIT

mkdir -p "${TMP_DIR}/.claude/harness"
cat > "${TMP_DIR}/.claude/harness/plans.json" <<'EOF'
{
  "project": "test", "meta": {"lastRelease": "", "lastReleaseDate": ""},
  "phases": [{
    "id": 32, "title": "Test Phase", "created": "2026-01-01", "goal": "Test",
    "status": "active", "urgency": "medium", "importance": "medium", "comments": [],
    "tasks": [
      {"id": "32.1.1", "name": "Create contract", "description": "", "dod": "Load runtime validation into contract", "depends": ["32.0.1"], "status": "cc:TODO", "urgency": "medium", "importance": "medium", "qualityMarkers": [], "comments": []}
    ]
  }],
  "futureConsiderations": []
}
EOF
PLANS_JSON="${TMP_DIR}/.claude/harness/plans.json"

CONTRACT_PATH="${TMP_DIR}/contract.json"
"${PROJECT_ROOT}/harness/scripts/generate-sprint-contract.sh" "32.1.1" "${PLANS_JSON}" "$CONTRACT_PATH" >/dev/null

if "${PROJECT_ROOT}/harness/scripts/ensure-sprint-contract-ready.sh" "$CONTRACT_PATH" >/dev/null 2>&1; then
  echo "contract should fail before approval"
  exit 1
fi

"${PROJECT_ROOT}/harness/scripts/enrich-sprint-contract.sh" "$CONTRACT_PATH" \
  --check "Re-verify DoD from reviewer perspective" \
  --non-goal "UI polish is out of scope this time" \
  --risk "needs-spike" \
  --note "reviewer checked runtime path" \
  --approve >/dev/null

"${PROJECT_ROOT}/harness/scripts/ensure-sprint-contract-ready.sh" "$CONTRACT_PATH" >/dev/null

jq -e '
  .review.status == "approved" and
  (.contract.non_goals | length) == 1 and
  (.contract.risk_flags | index("needs-spike")) != null
' "$CONTRACT_PATH" >/dev/null

echo "test-sprint-contract-approval: ok"
