#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

SETTINGS_FILE="${ROOT_DIR}/.claude-plugin/settings.json"
HOOK_FILES=(
  "${ROOT_DIR}/hooks/hooks.json"
  "${ROOT_DIR}/.claude-plugin/hooks.json"
)
TEMPLATE_FILES=(
  "${ROOT_DIR}/templates/rules/coding-standards.md.template"
  "${ROOT_DIR}/templates/rules/testing.md.template"
  "${ROOT_DIR}/templates/rules/plans-management.md.template"
)
SKILL_FILES=(
  "${ROOT_DIR}/skills-v3/harness-work/SKILL.md"
  "${ROOT_DIR}/skills-v3/harness-review/SKILL.md"
  "${ROOT_DIR}/skills-v3/harness-plan/SKILL.md"
)
AGENT_FILES=(
  "${ROOT_DIR}/agents-v3/worker.md"
  "${ROOT_DIR}/agents-v3/reviewer.md"
  "${ROOT_DIR}/agents-v3/scaffolder.md"
)

jq -e '.env.CLAUDE_CODE_SUBPROCESS_ENV_SCRUB == "1"' "${SETTINGS_FILE}" >/dev/null || {
  echo "settings.json is missing CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1"
  exit 1
}

jq -e '.sandbox.failIfUnavailable == true' "${SETTINGS_FILE}" >/dev/null || {
  echo "settings.json is missing sandbox.failIfUnavailable=true"
  exit 1
}

for hooks_file in "${HOOK_FILES[@]}"; do
  for event in TaskCreated CwdChanged FileChanged; do
    jq -e ".hooks.${event}[]?.hooks[]? | select(.command | contains(\"hook-handlers/runtime-reactive\"))" "${hooks_file}" >/dev/null || {
      echo "${hooks_file} is missing ${event} -> runtime-reactive wiring"
      exit 1
    }
  done
done

for hooks_file in "${HOOK_FILES[@]}"; do
  jq -e '.hooks.PermissionRequest[]? | select(.matcher == "Edit|Write|MultiEdit")' "${hooks_file}" >/dev/null || {
    echo "${hooks_file} is missing PermissionRequest matcher for Edit|Write|MultiEdit"
    exit 1
  }

  jq -e '.hooks.PermissionRequest[]? | select(.matcher == "Bash" and (.if // "" | contains("Bash(git status*)")) and (.if // "" | contains("Bash(pytest*)")))' "${hooks_file}" >/dev/null || {
    echo "${hooks_file} is missing PermissionRequest Bash conditional if guard"
    exit 1
  }
done

for template_file in "${TEMPLATE_FILES[@]}"; do
  grep -q '^paths:$' "${template_file}" || {
    echo "${template_file} is missing YAML list paths header"
    exit 1
  }
  grep -q '^  - "' "${template_file}" || {
    echo "${template_file} does not use YAML list paths entries"
    exit 1
  }
done

for skill_file in "${SKILL_FILES[@]}"; do
  grep -q '^effort:' "${skill_file}" || {
    echo "${skill_file} is missing effort frontmatter"
    exit 1
  }
done

for agent_file in "${AGENT_FILES[@]}"; do
  grep -q '^initialPrompt:' "${agent_file}" || {
    echo "${agent_file} is missing initialPrompt frontmatter"
    exit 1
  }
done

echo "OK"
