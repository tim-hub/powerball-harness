#!/bin/bash
# pre-tool.sh — Harness PreToolUse hook shim
if command -v harness >/dev/null 2>&1; then
  harness hook pre-tool
elif [ -x "${CLAUDE_PLUGIN_ROOT}/bin/harness" ]; then
  "${CLAUDE_PLUGIN_ROOT}/bin/harness" hook pre-tool
else
  echo '{"decision":"approve","reason":"harness binary not found"}' >&2
  exit 1
fi
