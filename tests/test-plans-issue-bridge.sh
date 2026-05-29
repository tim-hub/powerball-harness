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
  "phases": [
    {
      "id": 291, "title": "Plans.md to GitHub Issue bridge (opt-in)", "created": "2026-01-01",
      "goal": "Bridge", "status": "active", "urgency": "medium", "importance": "medium", "comments": [],
      "tasks": [
        {"id": "29.1.1", "name": "Define team mode", "description": "", "dod": "opt-in conditions are clearly documented", "depends": [], "status": "cc:TODO", "urgency": "medium", "importance": "medium", "qualityMarkers": [], "comments": []},
        {"id": "29.1.2", "name": "Create issue payload dry-run", "description": "", "dod": "Tasks can be extracted from plans.json", "depends": ["29.1.1"], "status": "cc:TODO", "urgency": "medium", "importance": "medium", "qualityMarkers": [], "comments": []}
      ]
    },
    {
      "id": 293, "title": "Lightweight brief and machine-readable manifest", "created": "2026-01-01",
      "goal": "Brief", "status": "active", "urgency": "medium", "importance": "medium", "comments": [],
      "tasks": [
        {"id": "29.3.1", "name": "Add design brief", "description": "", "dod": "UI brief template exists", "depends": [], "status": "cc:TODO", "urgency": "medium", "importance": "medium", "qualityMarkers": [], "comments": []}
      ]
    }
  ],
  "futureConsiderations": []
}
EOF
PLANS_JSON="${TMP_DIR}/.claude/harness/plans.json"

JSON_OUTPUT="${TMP_DIR}/bridge.json"
(cd "$TMP_DIR" && "${PROJECT_ROOT}/harness/scripts/plans-issue-bridge.sh" --plans "${PLANS_JSON}" --team-mode --format json --output "${JSON_OUTPUT}" >/dev/null)

jq -e '
  .schema_version == "plans-issue-bridge.v1" and
  .team_mode.enabled == true and
  .summary.phase_count == 2 and
  .summary.task_count == 3 and
  (.tracking_issue.labels | index("team-mode")) != null and
  (.sub_issues | length) == 3 and
  .sub_issues[1].depends_on == ["29.1.1"] and
  .sub_issues[2].phase.id == "293" and
  .sub_issues[2].status == "cc:TODO"
' "${JSON_OUTPUT}" >/dev/null

MARKDOWN_OUTPUT="${TMP_DIR}/bridge.md"
(cd "$TMP_DIR" && "${PROJECT_ROOT}/harness/scripts/plans-issue-bridge.sh" --plans "${PLANS_JSON}" --team-mode --format markdown --output "${MARKDOWN_OUTPUT}" >/dev/null)

grep -q "issue bridge dry-run" "${MARKDOWN_OUTPUT}"
grep -q "Team mode: enabled" "${MARKDOWN_OUTPUT}"
grep -q "29.1.2 Create issue payload dry-run" "${MARKDOWN_OUTPUT}"

echo "test-plans-issue-bridge: ok"
