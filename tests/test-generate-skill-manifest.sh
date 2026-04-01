#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

OUTPUT_JSON="${TMP_DIR}/skill-manifest.json"
(cd "$PROJECT_ROOT" && "${PROJECT_ROOT}/scripts/generate-skill-manifest.sh" --output "${OUTPUT_JSON}" >/dev/null)

jq -e '
  .schema_version == "skill-manifest.v1" and
  .skill_count > 5 and
  ((.skills | map(.path) | sort) == (.skills | map(.path))) and
  any(.skills[]; .name == "harness-plan" and .path == "skills-v3/harness-plan/SKILL.md") and
  any(.skills[]; .name == "breezing" and (.path | test("skills-v3-codex/breezing/SKILL.md")))
' "${OUTPUT_JSON}" >/dev/null

jq -e '
  any(.skills[]; .name == "harness-plan" and (.allowed_tools | index("Read")) != null and (.allowed_tools | index("Task")) != null and .effort == "medium" and .surface == "skills-v3" and (.related_surfaces | index("skills")) != null and (.related_surfaces | index("codex/.codex/skills")) != null and (.do_not_use_for | index("implementation")) != null and (.do_not_use_for | index("release tasks")) != null)
' "${OUTPUT_JSON}" >/dev/null

echo "test-generate-skill-manifest: ok"
