#!/bin/bash
# posttool-output-normalize.sh
# Phase 95.4: PostToolUse.updatedToolOutput governance handler (opt-in)
#
# Claude Code 2.1.121+ allows PostToolUse hooks to return
# hookSpecificOutput.updatedToolOutput. This handler is OPT-IN and only
# applies to three permitted use cases: redaction, compaction, and
# machine-readable normalization. The original output and updated output are
# recorded in an append-only audit ledger.
#
# Enable: set HARNESS_OUTPUT_GOVERNANCE_ENABLE=1 in the environment or via
# .claude/settings.json env block. Disabled by default.
#
# Prohibited uses (Phase 58.2.2 governance):
#   - Tampering with review or test output
#   - Mixing human-readable explanations into JSON-contract tool output
#   - Hiding error evidence
#
# Input:  stdin JSON ({ tool_name, tool_input, tool_response, ... })
# Output: stdout JSON ({ hookSpecificOutput: { updatedToolOutput: "..." } })
#         or empty (= no-op)
# Audit:  .claude/state/output-audit.jsonl -- append-only before/after record

set -euo pipefail

PROJECT_ROOT="${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
STATE_DIR="${PROJECT_ROOT}/.claude/state"
AUDIT_LOG="${STATE_DIR}/output-audit.jsonl"

mkdir -p "${STATE_DIR}"

INPUT="$(cat)"

if [ -z "${INPUT}" ]; then
  exit 0
fi

# Default policy: no-op (silent passthrough).
# Only acts when HARNESS_OUTPUT_GOVERNANCE_ENABLE=1 AND a rule matcher fires.
if [ "${HARNESS_OUTPUT_GOVERNANCE_ENABLE:-0}" != "1" ]; then
  exit 0
fi

TOOL_NAME="$(printf '%s' "${INPUT}" | jq -r '.tool_name // ""' 2>/dev/null || echo "")"
TOOL_OUTPUT="$(printf '%s' "${INPUT}" | jq -r '.tool_response.output // .tool_response.stdout // ""' 2>/dev/null || echo "")"

# JSON-contract tools: skip to avoid mixing human-readable text into structured output.
JSON_CONTRACT_TOOLS_REGEX='^(Read|Grep|Glob|TodoWrite|Bash)$'
if printf '%s' "${TOOL_NAME}" | grep -Eq "${JSON_CONTRACT_TOOLS_REGEX}"; then
  exit 0
fi

NORMALIZED_OUTPUT="${TOOL_OUTPUT}"
APPLIED_RULE=""

# Rule 1: redact OpenAI / Anthropic API key patterns (defense-in-depth).
if printf '%s' "${TOOL_OUTPUT}" | grep -Eq 'sk-[A-Za-z0-9]{20,}|sk-ant-[A-Za-z0-9-]{20,}'; then
  NORMALIZED_OUTPUT="$(printf '%s' "${TOOL_OUTPUT}" | sed -E 's/sk-[A-Za-z0-9]{20,}/sk-REDACTED/g; s/sk-ant-[A-Za-z0-9-]{20,}/sk-ant-REDACTED/g')"
  APPLIED_RULE="redact-api-key"
fi

# If no rule matched, exit silently.
if [ -z "${APPLIED_RULE}" ]; then
  exit 0
fi

# Append audit record (before / after / rule / timestamp).
TIMESTAMP="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
AUDIT_RECORD="$(jq -nc \
  --arg timestamp "${TIMESTAMP}" \
  --arg tool "${TOOL_NAME}" \
  --arg rule "${APPLIED_RULE}" \
  --arg before "${TOOL_OUTPUT}" \
  --arg after "${NORMALIZED_OUTPUT}" \
  '{timestamp:$timestamp, tool:$tool, rule:$rule, before:$before, after:$after}')"

printf '%s\n' "${AUDIT_RECORD}" >> "${AUDIT_LOG}"

# Emit hookSpecificOutput.
jq -nc \
  --arg output "${NORMALIZED_OUTPUT}" \
  '{hookSpecificOutput: {updatedToolOutput: $output}}'
