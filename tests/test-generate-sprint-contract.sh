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
      {"id": "32.1.1", "name": "Create contract", "description": "", "dod": "Load runtime validation into contract", "depends": ["32.0.1"], "status": "cc:TODO", "urgency": "medium", "importance": "medium", "qualityMarkers": [], "comments": []},
      {"id": "32.2.2", "name": "Add browser evaluator", "description": "", "dod": "Verify UI flow in browser", "depends": ["32.2.1"], "status": "cc:TODO", "urgency": "medium", "importance": "medium", "qualityMarkers": [], "comments": []},
      {"id": "32.2.5", "name": "Handle browser_mode: exploratory", "description": "", "dod": "Prioritize AgentBrowser in exploratory mode", "depends": ["32.2.2"], "status": "cc:TODO", "urgency": "medium", "importance": "medium", "qualityMarkers": [], "comments": []}
    ]
  }],
  "futureConsiderations": []
}
EOF
PLANS_JSON="${TMP_DIR}/.claude/harness/plans.json"

cat > "${TMP_DIR}/package.json" <<'EOF'
{
  "scripts": {
    "test": "vitest run",
    "test:e2e": "playwright test"
  },
  "devDependencies": {
    "@playwright/test": "^1.52.0"
  }
}
EOF

OUTPUT_PATH="${TMP_DIR}/out/32.1.1.sprint-contract.json"
(cd "${TMP_DIR}" && "${PROJECT_ROOT}/harness/scripts/generate-sprint-contract.sh" "32.1.1" "${PLANS_JSON}" "${OUTPUT_PATH}" >/dev/null)

jq -e '
  .schema_version == "sprint-contract.v1" and
  .task.id == "32.1.1" and
  .task.depends_on == ["32.0.1"] and
  .review.reviewer_profile == "runtime" and
  (.contract.runtime_validation | type) == "array" and
  (.contract.runtime_validation[0].command | test("npm test"))
' "${OUTPUT_PATH}" >/dev/null

BROWSER_OUTPUT="${TMP_DIR}/out/32.2.2.sprint-contract.json"
(cd "${TMP_DIR}" && "${PROJECT_ROOT}/harness/scripts/generate-sprint-contract.sh" "32.2.2" "${PLANS_JSON}" "${BROWSER_OUTPUT}" >/dev/null)

jq -e '
  .task.id == "32.2.2" and
  .review.reviewer_profile == "browser" and
  .review.browser_mode == "scripted" and
  .review.route == "playwright" and
  (.contract.browser_validation[0].required_artifacts | index("trace")) != null
' "${BROWSER_OUTPUT}" >/dev/null

EXPLORATORY_OUTPUT="${TMP_DIR}/out/32.2.5.sprint-contract.json"
(cd "${TMP_DIR}" && HARNESS_BROWSER_REVIEW_DISABLE_AGENT_BROWSER=1 \
  "${PROJECT_ROOT}/harness/scripts/generate-sprint-contract.sh" "32.2.5" "${PLANS_JSON}" "${EXPLORATORY_OUTPUT}" >/dev/null)

jq -e '
  .task.id == "32.2.5" and
  .review.reviewer_profile == "browser" and
  .review.browser_mode == "exploratory" and
  (.contract.browser_validation[0].required_artifacts | index("snapshot")) != null
' "${EXPLORATORY_OUTPUT}" >/dev/null

echo "test-generate-sprint-contract: ok"
