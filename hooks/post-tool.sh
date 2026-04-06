#!/bin/bash
# post-tool.sh — Harness PostToolUse hook shim
if command -v harness >/dev/null 2>&1; then
  harness hook post-tool
elif [ -x "${CLAUDE_PLUGIN_ROOT}/bin/harness" ]; then
  "${CLAUDE_PLUGIN_ROOT}/bin/harness" hook post-tool
else
  echo '{"decision":"approve","reason":"harness binary not found"}' >&2
  exit 1
fi
