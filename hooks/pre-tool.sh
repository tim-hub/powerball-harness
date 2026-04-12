#!/bin/bash
# pre-tool.sh — Harness v3 PreToolUse thin shim (under 5 lines)
# stdin JSON → core engine → stdout JSON
node "${CLAUDE_PLUGIN_ROOT}/core/dist/index.js" pre-tool
