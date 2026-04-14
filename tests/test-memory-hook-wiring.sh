#!/bin/bash
# SessionStart/UserPromptSubmit/PostToolUse/Stop should be wired to harness-mem and
# SessionStart should surface memory resume context immediately.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HARNESS_DIR="${ROOT_DIR}/harness"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

required_wrapper_files=(
  "${HARNESS_DIR}/scripts/lib/harness-mem-bridge.sh"
  "${HARNESS_DIR}/scripts/hook-handlers/memory-bridge.sh"
  "${HARNESS_DIR}/scripts/hook-handlers/memory-session-start.sh"
  "${HARNESS_DIR}/scripts/hook-handlers/memory-user-prompt.sh"
  "${HARNESS_DIR}/scripts/hook-handlers/memory-post-tool-use.sh"
  "${HARNESS_DIR}/scripts/hook-handlers/memory-stop.sh"
  "${HARNESS_DIR}/scripts/hook-handlers/memory-codex-notify.sh"
)

for wrapper_file in "${required_wrapper_files[@]}"; do
  [ -f "${wrapper_file}" ] || {
    echo "Required harness-mem wrapper is missing: ${wrapper_file}"
    exit 1
  }
done

for hooks_file in "${HARNESS_DIR}/hooks/hooks.json"; do
  # Matcher checks use strict pipe-token regex to avoid false positives on
  # typos like "startup-only" or "startup_special". The pattern matches
  # "startup" as a standalone token in pipe-separated matchers:
  #   - "startup"              → matches (whole string)
  #   - "startup|resume"       → matches (pipe-delimited token)
  #   - "resume|startup"       → matches (pipe-delimited token, end)
  #   - "startup-only"         → NO match (hyphen breaks boundary)
  jq -e '.hooks.SessionStart[] | select(.matcher | test("(^|\\|)startup($|\\|)")) | .hooks[] | select(.command | contains("memory-bridge"))' "${hooks_file}" >/dev/null || {
    echo "SessionStart startup is missing memory-bridge session-start in ${hooks_file}"
    exit 1
  }

  jq -e '.hooks.SessionStart[] | select(.matcher | test("(^|\\|)resume($|\\|)")) | .hooks[] | select(.command | contains("memory-bridge"))' "${hooks_file}" >/dev/null || {
    echo "SessionStart resume is missing memory-bridge session-start in ${hooks_file}"
    exit 1
  }

  jq -e '.hooks.UserPromptSubmit[] | .hooks[] | select(.command? | strings | contains("memory-bridge"))' "${hooks_file}" >/dev/null || {
    echo "UserPromptSubmit is missing memory-bridge user-prompt in ${hooks_file}"
    exit 1
  }

  jq -e '.hooks.PostToolUse[] | .hooks[] | select(.command? | strings | contains("memory-bridge"))' "${hooks_file}" >/dev/null || {
    echo "PostToolUse is missing memory-bridge post-tool-use in ${hooks_file}"
    exit 1
  }

  jq -e '.hooks.Stop[] | .hooks[] | select(.command? | strings | contains("memory-bridge"))' "${hooks_file}" >/dev/null || {
    echo "Stop is missing memory-bridge stop in ${hooks_file}"
    exit 1
  }
done

mkdir -p "${TMP_DIR}/.claude/state/snapshots"
mkdir -p "${TMP_DIR}/scripts/lib"
git -C "${TMP_DIR}" init -q

cp "${HARNESS_DIR}/VERSION" "${TMP_DIR}/VERSION"
cp "${HARNESS_DIR}/scripts/session-init.sh" "${TMP_DIR}/scripts/session-init.sh"
cp "${HARNESS_DIR}/scripts/session-resume.sh" "${TMP_DIR}/scripts/session-resume.sh"
cp "${HARNESS_DIR}/scripts/lib/progress-snapshot.sh" "${TMP_DIR}/scripts/lib/progress-snapshot.sh"

cat > "${TMP_DIR}/Plans.md" <<'EOF'
| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 1.0 | sample | done | - | cc:WIP |
EOF

seed_memory_context() {
  cat > "${TMP_DIR}/.claude/state/memory-resume-context.md" <<'EOF'
# Continuity Briefing

## Current Focus
- Continue from the previous session
EOF
  : > "${TMP_DIR}/.claude/state/.memory-resume-pending"
}

seed_memory_context
init_output="$(cd "${TMP_DIR}" && bash "${TMP_DIR}/scripts/session-init.sh")"
init_context="$(printf '%s' "${init_output}" | jq -r '.hookSpecificOutput.additionalContext')"

echo "${init_context}" | grep -q 'Continuity Briefing' || {
  echo "session-init additionalContext is missing memory continuity briefing"
  exit 1
}

[ ! -f "${TMP_DIR}/.claude/state/memory-resume-context.md" ] || {
  echo "session-init should consume memory-resume-context.md"
  exit 1
}

[ ! -f "${TMP_DIR}/.claude/state/.memory-resume-pending" ] || {
  echo "session-init should clear .memory-resume-pending"
  exit 1
}

seed_memory_context
resume_output="$(cd "${TMP_DIR}" && bash "${TMP_DIR}/scripts/session-resume.sh")"
resume_context="$(printf '%s' "${resume_output}" | jq -r '.hookSpecificOutput.additionalContext')"

echo "${resume_context}" | grep -q 'Continuity Briefing' || {
  echo "session-resume additionalContext is missing memory continuity briefing"
  exit 1
}

[ ! -f "${TMP_DIR}/.claude/state/memory-resume-context.md" ] || {
  echo "session-resume should consume memory-resume-context.md"
  exit 1
}

[ ! -f "${TMP_DIR}/.claude/state/.memory-resume-pending" ] || {
  echo "session-resume should clear .memory-resume-pending"
  exit 1
}

echo "OK"
