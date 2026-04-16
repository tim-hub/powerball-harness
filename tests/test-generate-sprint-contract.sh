#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

cat > "${TMP_DIR}/Plans.md" <<'EOF'
| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 32.1.1 | contract を作る | runtime validation を contract に載せる | 32.0.1 | cc:TODO |
| 32.2.2 | browser evaluator を追加する | browser で UI フローを確認する | 32.2.1 | cc:TODO |
| 32.2.5 | browser_mode: exploratory を扱う | exploratory mode で AgentBrowser を優先する | 32.2.2 | cc:TODO |
| 43.1.1 | [needs-spike] security migration contract | state migration の guard を確認する <!-- advisor:required --> | - | cc:TODO |
EOF

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
(cd "${TMP_DIR}" && node "${PROJECT_ROOT}/scripts/generate-sprint-contract.js" "32.1.1" "${TMP_DIR}/Plans.md" "${OUTPUT_PATH}" >/dev/null)

jq -e '
  .schema_version == "sprint-contract.v1" and
  .task.id == "32.1.1" and
  .task.depends_on == ["32.0.1"] and
  .review.reviewer_profile == "runtime" and
  .advisor.enabled == true and
  .advisor.mode == "on-demand" and
  .advisor.max_consults == 3 and
  .advisor.retry_threshold == 2 and
  .advisor.pre_escalation_consult == true and
  .advisor.model_policy.claude_default == "opus" and
  .advisor.model_policy.codex_default == "gpt-5.4" and
  (.advisor.triggers | length) == 0 and
  (.contract.runtime_validation | type) == "array" and
  .contract.runtime_validation[0].command == "CI=true npm test"
' "${OUTPUT_PATH}" >/dev/null

BROWSER_OUTPUT="${TMP_DIR}/out/32.2.2.sprint-contract.json"
(cd "${TMP_DIR}" && node "${PROJECT_ROOT}/scripts/generate-sprint-contract.js" "32.2.2" "${TMP_DIR}/Plans.md" "${BROWSER_OUTPUT}" >/dev/null)

jq -e '
  .task.id == "32.2.2" and
  .review.reviewer_profile == "browser" and
  .review.browser_mode == "scripted" and
  .review.route == "playwright" and
  (.contract.browser_validation[0].required_artifacts | index("trace")) != null
' "${BROWSER_OUTPUT}" >/dev/null

EXPLORATORY_OUTPUT="${TMP_DIR}/out/32.2.5.sprint-contract.json"
(cd "${TMP_DIR}" && HARNESS_BROWSER_REVIEW_DISABLE_AGENT_BROWSER=1 \
  node "${PROJECT_ROOT}/scripts/generate-sprint-contract.js" "32.2.5" "${TMP_DIR}/Plans.md" "${EXPLORATORY_OUTPUT}" >/dev/null)

jq -e '
  .task.id == "32.2.5" and
  .review.reviewer_profile == "browser" and
  .review.browser_mode == "exploratory" and
  .review.route == null and
  (.contract.browser_validation[0].required_artifacts | index("snapshot")) != null
' "${EXPLORATORY_OUTPUT}" >/dev/null

ADVISOR_OUTPUT="${TMP_DIR}/out/43.1.1.sprint-contract.json"
(cd "${TMP_DIR}" && node "${PROJECT_ROOT}/scripts/generate-sprint-contract.js" "43.1.1" "${TMP_DIR}/Plans.md" "${ADVISOR_OUTPUT}" >/dev/null)

jq -e '
  .task.id == "43.1.1" and
  .advisor.enabled == true and
  .advisor.mode == "on-demand" and
  .advisor.max_consults == 3 and
  .advisor.retry_threshold == 2 and
  .advisor.pre_escalation_consult == true and
  .advisor.model_policy.claude_default == "opus" and
  .advisor.model_policy.codex_default == "gpt-5.4" and
  .advisor.triggers == ["needs-spike", "security-sensitive", "state-migration", "<!-- advisor:required -->"] and
  (.contract.risk_flags | index("security-sensitive")) != null and
  (.contract.risk_flags | index("state-migration")) != null
' "${ADVISOR_OUTPUT}" >/dev/null

cat > "${TMP_DIR}/ui-plans.md" <<'EOF'
| Task | 内容 | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 41.3.1 | design-heavy task | design と styling と aesthetic を見ながら UI layout を磨く | 41.2.1 | cc:TODO |
EOF

UI_RUBRIC_OUTPUT="${TMP_DIR}/out/41.3.1.sprint-contract.json"
(cd "${TMP_DIR}" && node "${PROJECT_ROOT}/scripts/generate-sprint-contract.js" "41.3.1" "${TMP_DIR}/ui-plans.md" "${UI_RUBRIC_OUTPUT}" >/dev/null)

jq -e '
  .task.id == "41.3.1" and
  .review.reviewer_profile == "ui-rubric" and
  .review.max_iterations == 10 and
  .review.rubric_target.design == 6 and
  .review.rubric_target.originality == 6 and
  .review.rubric_target.craft == 6 and
  .review.rubric_target.functionality == 6
' "${UI_RUBRIC_OUTPUT}" >/dev/null

echo "test-generate-sprint-contract: ok"
