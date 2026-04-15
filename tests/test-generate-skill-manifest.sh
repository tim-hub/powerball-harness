#!/bin/bash

set -euo pipefail
export TMPDIR=/tmp  # Force /tmp for sandboxed execution (sandbox blocks /var/folders)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d "/tmp/harness-test.XXXXXX")"
trap 'rm -rf "$TMP_DIR"' EXIT

OUTPUT_JSON="${TMP_DIR}/skill-manifest.json"
# Run from harness/ so roots=['skills','templates/codex-skills'] resolve correctly (Phase 52+)
(cd "${PROJECT_ROOT}/harness" && "${PROJECT_ROOT}/harness/scripts/generate-skill-manifest.sh" --output "${OUTPUT_JSON}" >/dev/null)

jq -e '
  .schema_version == "skill-manifest.v1" and
  .skill_count > 5 and
  ((.skills | map(.path) | sort) == (.skills | map(.path))) and
  any(.skills[]; .name == "harness-plan" and .path == "skills/harness-plan/SKILL.md") and
  any(.skills[]; .name == "breezing" and (.path | test("templates/codex-skills/breezing/SKILL.md")))
' "${OUTPUT_JSON}" >/dev/null

jq -e '
  any(.skills[]; .name == "harness-plan" and (.allowed_tools | index("Read")) != null and (.allowed_tools | index("Task")) != null and .effort == "medium" and .surface == "skills" and (.do_not_use_for | any(.[]; contains("implementation"))) and (.do_not_use_for | any(.[]; contains("release"))))
' "${OUTPUT_JSON}" >/dev/null

echo "test-generate-skill-manifest: ok"
