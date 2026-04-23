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
  "${ROOT_DIR}/skills/harness-work/SKILL.md"
  "${ROOT_DIR}/skills/harness-review/SKILL.md"
  "${ROOT_DIR}/skills/harness-plan/SKILL.md"
)
UPSTREAM_SKILL_NAMES=(
  "cc-update-review"
  "claude-codex-upstream-update"
)
AGENT_FILES=(
  "${ROOT_DIR}/agents/worker.md"
  "${ROOT_DIR}/agents/reviewer.md"
  "${ROOT_DIR}/agents/scaffolder.md"
)

jq -e '.env.CLAUDE_CODE_SUBPROCESS_ENV_SCRUB == "1"' "${SETTINGS_FILE}" >/dev/null || {
  echo "settings.json is missing CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1"
  exit 1
}

jq -e '.sandbox.failIfUnavailable == true' "${SETTINGS_FILE}" >/dev/null || {
  echo "settings.json is missing sandbox.failIfUnavailable=true"
  exit 1
}

jq -e '.sandbox.network.deniedDomains | type == "array" and (index("169.254.169.254") != null)' "${SETTINGS_FILE}" >/dev/null || {
  echo "settings.json is missing sandbox.network.deniedDomains metadata protection"
  exit 1
}

for hooks_file in "${HOOK_FILES[@]}"; do
  for event in TaskCreated CwdChanged FileChanged; do
    jq -e ".hooks.${event}[]?.hooks[]? | select(.command | contains(\"runtime-reactive\"))" "${hooks_file}" >/dev/null || {
      echo "${hooks_file} is missing ${event} -> runtime-reactive wiring"
      exit 1
    }
  done
done

for hooks_file in "${HOOK_FILES[@]}"; do
  jq -e '.hooks.PreToolUse[]? | select(.matcher == "AskUserQuestion") | .hooks[]? | select(.command | contains("ask-user-question-normalize"))' "${hooks_file}" >/dev/null || {
    echo "${hooks_file} is missing PreToolUse AskUserQuestion -> ask-user-question-normalize wiring"
    exit 1
  }

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

# v2.1.89: PermissionDenied hook wiring check
for hooks_file in "${HOOK_FILES[@]}"; do
  jq -e '.hooks.PermissionDenied[]?.hooks[]? | select(.command | contains("permission-denied"))' "${hooks_file}" >/dev/null || {
    echo "${hooks_file} is missing PermissionDenied -> permission-denied wiring"
    exit 1
  }
done

# v2.1.89: PermissionDenied handler script exists and is executable
PERM_DENIED_HANDLER="${ROOT_DIR}/scripts/hook-handlers/permission-denied-handler.sh"
[ -f "${PERM_DENIED_HANDLER}" ] || {
  echo "permission-denied-handler.sh does not exist"
  exit 1
}
[ -x "${PERM_DENIED_HANDLER}" ] || {
  echo "permission-denied-handler.sh is not executable"
  exit 1
}

# v2.1.89+: AskUserQuestion updatedInput answer bridge exists in Go fast path
ASK_NORMALIZER_GO="${ROOT_DIR}/go/internal/hookhandler/ask_user_question_normalizer.go"
[ -f "${ASK_NORMALIZER_GO}" ] || {
  echo "ask_user_question_normalizer.go does not exist"
  exit 1
}
grep -q 'HARNESS_ASK_USER_QUESTION_ANSWERS' "${ASK_NORMALIZER_GO}" || {
  echo "ask_user_question_normalizer.go is missing explicit answer source support"
  exit 1
}
grep -q 'updatedInput' "${ASK_NORMALIZER_GO}" || {
  echo "ask_user_question_normalizer.go is missing updatedInput output"
  exit 1
}

# v2.1.113: Bash hardening parity checks for find deletion and macOS dangerous removal paths
GUARDRAIL_HELPERS_GO="${ROOT_DIR}/go/internal/guardrail/helpers.go"
GUARDRAIL_RULES_TEST_GO="${ROOT_DIR}/go/internal/guardrail/rules_test.go"
grep -q 'hasDangerousFindDelete' "${GUARDRAIL_HELPERS_GO}" || {
  echo "guardrail helpers are missing find -delete / -exec rm detection"
  exit 1
}
grep -q 'hasDangerousMacOSRemovalPath' "${GUARDRAIL_HELPERS_GO}" || {
  echo "guardrail helpers are missing macOS dangerous removal path detection"
  exit 1
}
grep -q 'TestR05_FindDelete' "${GUARDRAIL_RULES_TEST_GO}" || {
  echo "guardrail tests are missing find -delete coverage"
  exit 1
}
grep -q 'TestR05_MacOSPrivatePath' "${GUARDRAIL_RULES_TEST_GO}" || {
  echo "guardrail tests are missing macOS dangerous path coverage"
  exit 1
}

# v2.1.113: template parity for deniedDomains (consumer init must inherit metadata protection)
SECURITY_TEMPLATE="${ROOT_DIR}/templates/claude/settings.security.json.template"
[ -f "${SECURITY_TEMPLATE}" ] || {
  echo "settings.security.json.template does not exist"
  exit 1
}
jq -e '.sandbox.network.deniedDomains | type == "array" and (index("169.254.169.254") != null)' "${SECURITY_TEMPLATE}" >/dev/null || {
  echo "${SECURITY_TEMPLATE} is missing deniedDomains parity with .claude-plugin/settings.json"
  exit 1
}

# v2.1.113: end-to-end exercise of ask-user-question-normalize hook (binary contract smoke)
HARNESS_BIN="${ROOT_DIR}/bin/harness"
if [ -x "${HARNESS_BIN}" ]; then
  ASK_HOOK_INPUT='{"tool_name":"AskUserQuestion","tool_input":{"questions":[{"question":"Execution mode?","header":"Mode","options":[{"label":"solo"},{"label":"team"}],"multiSelect":false}],"answers":{"Execution mode?":"solo"}}}'
  ASK_HOOK_OUTPUT="$(printf '%s' "${ASK_HOOK_INPUT}" | "${HARNESS_BIN}" hook ask-user-question-normalize 2>/dev/null || true)"
  if [ -z "${ASK_HOOK_OUTPUT}" ]; then
    echo "ask-user-question-normalize hook produced no output for a valid single-select answer"
    exit 1
  fi
  printf '%s' "${ASK_HOOK_OUTPUT}" | jq -e '.hookSpecificOutput.permissionDecision == "allow" and .hookSpecificOutput.hookEventName == "PreToolUse"' >/dev/null || {
    echo "ask-user-question-normalize hook output missing expected permissionDecision/hookEventName"
    echo "output: ${ASK_HOOK_OUTPUT}"
    exit 1
  }
  printf '%s' "${ASK_HOOK_OUTPUT}" | jq -e '.hookSpecificOutput.updatedInput.answers["Execution mode?"] == "solo"' >/dev/null || {
    echo "ask-user-question-normalize hook did not echo answers in updatedInput"
    echo "output: ${ASK_HOOK_OUTPUT}"
    exit 1
  }
else
  echo "skip: bin/harness is not executable, skipping end-to-end ask-user-question-normalize exercise"
fi

# Phase 52: snapshot doc referenced from CHANGELOG / Feature Table / Plans must exist
UPSTREAM_SNAPSHOT_DOC="${ROOT_DIR}/docs/upstream-update-snapshot-2026-04-21.md"
[ -f "${UPSTREAM_SNAPSHOT_DOC}" ] || {
  echo "${UPSTREAM_SNAPSHOT_DOC} does not exist (referenced from CHANGELOG, Feature Table, and Plans)"
  exit 1
}
for referencing_file in \
  "${ROOT_DIR}/CHANGELOG.md" \
  "${ROOT_DIR}/docs/CLAUDE-feature-table.md" \
  "${ROOT_DIR}/Plans.md"; do
  grep -q 'upstream-update-snapshot-2026-04-21' "${referencing_file}" || {
    echo "${referencing_file} is missing the expected upstream-update-snapshot-2026-04-21 reference"
    exit 1
  }
done

# Phase 52: upstream update skill review contract and mirror drift checks
for skill_name in "${UPSTREAM_SKILL_NAMES[@]}"; do
  CANONICAL_SKILL="${ROOT_DIR}/skills/${skill_name}/SKILL.md"
  CODEX_SKILL="${ROOT_DIR}/codex/.codex/skills/${skill_name}/SKILL.md"
  LOCAL_AGENT_SKILL="${ROOT_DIR}/.agents/skills/${skill_name}/SKILL.md"

  [ -f "${CANONICAL_SKILL}" ] || {
    echo "${CANONICAL_SKILL} does not exist"
    exit 1
  }
  [ -f "${CODEX_SKILL}" ] || {
    echo "${CODEX_SKILL} does not exist"
    exit 1
  }
  cmp -s "${CANONICAL_SKILL}" "${CODEX_SKILL}" || {
    echo "${skill_name} skill mirror drift: skills/ and codex/.codex/ differ"
    exit 1
  }
  if [ -f "${LOCAL_AGENT_SKILL}" ]; then
    cmp -s "${CANONICAL_SKILL}" "${LOCAL_AGENT_SKILL}" || {
      echo "${skill_name} skill mirror drift: skills/ and .agents/ differ"
      exit 1
    }
  fi
done

CC_UPDATE_REVIEW="${ROOT_DIR}/skills/cc-update-review/SKILL.md"
grep -q 'allowed-tools: \["Read", "Grep", "Glob", "Bash"\]' "${CC_UPDATE_REVIEW}" || {
  echo "cc-update-review must allow read-only Bash for git diff inspection"
  exit 1
}
grep -q 'git diff -- docs/CLAUDE-feature-table.md' "${CC_UPDATE_REVIEW}" || {
  echo "cc-update-review is missing explicit Feature Table diff inspection guidance"
  exit 1
}
grep -q '## A/B/C/P 分類' "${CC_UPDATE_REVIEW}" || {
  echo "cc-update-review must name the actual A/B/C/P classification model"
  exit 1
}

UPSTREAM_UPDATE_SKILL="${ROOT_DIR}/skills/claude-codex-upstream-update/SKILL.md"
grep -q 'no-op adaptation' "${UPSTREAM_UPDATE_SKILL}" || {
  echo "claude-codex-upstream-update must allow documented no-op adaptation cycles"
  exit 1
}
grep -q 'Codex `0.122.0` 以降で確認する項目' "${UPSTREAM_UPDATE_SKILL}" || {
  echo "claude-codex-upstream-update is missing Codex 0.122.0+ watchlist"
  exit 1
}
grep -q 'Claude Code `2.1.116` 以降の UX / 運用改善' "${UPSTREAM_UPDATE_SKILL}" || {
  echo "claude-codex-upstream-update is missing Claude Code 2.1.116+ watchlist"
  exit 1
}

# Phase 53: snapshot doc and MCP hook safety decision
PHASE53_SNAPSHOT_DOC="${ROOT_DIR}/docs/upstream-update-snapshot-2026-04-23.md"
[ -f "${PHASE53_SNAPSHOT_DOC}" ] || {
  echo "${PHASE53_SNAPSHOT_DOC} does not exist"
  exit 1
}
for referencing_file in \
  "${ROOT_DIR}/CHANGELOG.md" \
  "${ROOT_DIR}/docs/CLAUDE-feature-table.md" \
  "${ROOT_DIR}/Plans.md"; do
  grep -q 'upstream-update-snapshot-2026-04-23' "${referencing_file}" || {
    echo "${referencing_file} is missing the expected upstream-update-snapshot-2026-04-23 reference"
    exit 1
  }
done
grep -q '53.1.2 MCP tool hook decision' "${PHASE53_SNAPSHOT_DOC}" || {
  echo "Phase 53 snapshot is missing the 53.1.2 MCP tool hook decision"
  exit 1
}
grep -q 'hooks/hooks.json` / `.claude-plugin/hooks.json` は今回は no-op' "${PHASE53_SNAPSHOT_DOC}" || {
  echo "Phase 53 snapshot must record that hook manifests are no-op for 53.1.2"
  exit 1
}
grep -q '読み取り専用の MCP health / resource list 診断' "${PHASE53_SNAPSHOT_DOC}" || {
  echo "Phase 53 snapshot must document the intended read-only MCP diagnostic use case"
  exit 1
}
grep -q '書き込み系 MCP tool は hook から呼ばない' "${PHASE53_SNAPSHOT_DOC}" || {
  echo "Phase 53 snapshot must forbid write-capable MCP tools from hooks"
  exit 1
}

# Phase 53.1.3: claude plugin tag must be visible in release preflight / dry-run guidance
HARNESS_RELEASE_SKILL="${ROOT_DIR}/skills/harness-release/SKILL.md"
grep -q 'claude plugin tag .claude-plugin --dry-run' "${HARNESS_RELEASE_SKILL}" || {
  echo "harness-release is missing claude plugin tag dry-run guidance"
  exit 1
}
grep -q 'claude plugin tag .claude-plugin --push --remote origin' "${HARNESS_RELEASE_SKILL}" || {
  echo "harness-release is missing claude plugin tag push guidance"
  exit 1
}
grep -q 'VERSION と .claude-plugin/plugin.json が不一致なら tag に進まない' "${HARNESS_RELEASE_SKILL}" || {
  echo "harness-release must stop before tagging when VERSION and plugin.json disagree"
  exit 1
}
grep -q '53.1.3 plugin tag release flow decision' "${PHASE53_SNAPSHOT_DOC}" || {
  echo "Phase 53 snapshot is missing the 53.1.3 plugin tag release flow decision"
  exit 1
}
grep -q 'claude plugin tag .claude-plugin --dry-run' "${PHASE53_SNAPSHOT_DOC}" || {
  echo "Phase 53 snapshot must record the claude plugin tag dry-run command"
  exit 1
}

# Phase 53.1.4: Auto Mode policy must extend built-in defaults instead of replacing them
grep -Fq '53.1.4 Auto Mode "$defaults" permission and sandbox policy' "${PHASE53_SNAPSHOT_DOC}" || {
  echo "Phase 53 snapshot is missing the 53.1.4 Auto Mode defaults policy"
  exit 1
}
grep -Fq 'Auto Mode built-in defaults stay in place through "$defaults"' "${PHASE53_SNAPSHOT_DOC}" || {
  echo 'Phase 53 snapshot must say Auto Mode built-in defaults are extended with $defaults'
  exit 1
}
grep -Fq 'R05 guardrail and sandbox.network.deniedDomains are not duplicated by Auto Mode' "${PHASE53_SNAPSHOT_DOC}" || {
  echo "Phase 53 snapshot must explain why R05 / deniedDomains remain separate guardrails"
  exit 1
}
jq -e '._harness_auto_mode_note | contains("Auto Mode guidance: keep \"$defaults\" and append only project-specific entries")' "${SECURITY_TEMPLATE}" >/dev/null || {
  echo 'settings.security.json.template must document additive Auto Mode $defaults guidance'
  exit 1
}
if jq -e 'has("autoMode")' "${SETTINGS_FILE}" >/dev/null; then
  jq -e '
    .autoMode
    | [
        to_entries[]
        | select(.key == "allow" or .key == "soft_deny" or .key == "environment")
        | (.value | type == "array" and index("$defaults") != null)
      ]
    | all
  ' "${SETTINGS_FILE}" >/dev/null || {
    echo 'settings.json autoMode allow/soft_deny/environment entries must include $defaults when present'
    exit 1
  }
fi
jq -e '
  (.permissions.deny | index("Bash(sudo:*)") != null and index("Bash(rm -rf:*)") != null and index("Bash(rm -fr:*)") != null)
  and
  (.permissions.ask | index("Bash(git reset --hard:*)") != null and index("Bash(git push --force:*)") != null)
  and
  (.sandbox.network.deniedDomains | index("169.254.169.254") != null and index("metadata.google.internal") != null and index("metadata.azure.com") != null)
' "${SETTINGS_FILE}" >/dev/null || {
  echo "settings.json must preserve existing deny, ask, and deniedDomains guardrails"
  exit 1
}

# Phase 53.1.5: plugin / managed settings policy docs must stay explicit
PLUGIN_POLICY_DOC="${ROOT_DIR}/docs/plugin-managed-settings-policy.md"
[ -f "${PLUGIN_POLICY_DOC}" ] || {
  echo "${PLUGIN_POLICY_DOC} does not exist"
  exit 1
}
grep -q 'DISABLE_UPDATES は手動 `claude update` まで止める' "${PLUGIN_POLICY_DOC}" || {
  echo "plugin policy doc must explain DISABLE_UPDATES vs DISABLE_AUTOUPDATER"
  exit 1
}
grep -q '通常ユーザー向け default には入れない' "${PLUGIN_POLICY_DOC}" || {
  echo "plugin policy doc must not over-apply managed marketplace restrictions to normal defaults"
  exit 1
}
grep -q 'Harness 独自の dependency resolver は追加しない' "${PLUGIN_POLICY_DOC}" || {
  echo "plugin policy doc must leave dependency resolution to Claude Code"
  exit 1
}
grep -q 'plugin `themes/` directory は今回は P' "${PLUGIN_POLICY_DOC}" || {
  echo "plugin policy doc must record the themes decision"
  exit 1
}
grep -q 'plugin-managed-settings-policy.md' "${ROOT_DIR}/skills/harness-setup/SKILL.md" || {
  echo "harness-setup must link to plugin managed settings policy"
  exit 1
}
grep -q '53.1.5 plugin / managed settings policy' "${PHASE53_SNAPSHOT_DOC}" || {
  echo "Phase 53 snapshot is missing the 53.1.5 plugin managed settings policy"
  exit 1
}
grep -q 'Plugin themes / managed settings / dependency auto-resolve.*A: docs 化済み' "${ROOT_DIR}/docs/CLAUDE-feature-table.md" || {
  echo "Feature Table must mark 53.1.5 plugin policy docs as done"
  exit 1
}

for hooks_file in "${HOOK_FILES[@]}"; do
  MCP_TOOL_COUNT="$(jq '[.. | objects | select(.type? == "mcp_tool")] | length' "${hooks_file}")"
  if [ "${MCP_TOOL_COUNT}" -eq 0 ]; then
    continue
  fi

  jq -e '
    [.. | objects | select(.type? == "mcp_tool")] |
    all(
      ((.tool // .tool_name // .name // "") | test("(health|list|read|get|status|diagnostic|resource)"; "i"))
      and
      ((.tool // .tool_name // .name // "") | test("(write|create|update|delete|remove|record|mutate|set|insert|upsert|patch)"; "i") | not)
    )
  ' "${hooks_file}" >/dev/null || {
    echo "${hooks_file} has an mcp_tool hook that is not clearly read-only"
    exit 1
  }
done

echo "OK"
