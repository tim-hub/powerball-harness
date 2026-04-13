#!/bin/bash
# pre-tool.sh — Harness v4 PreToolUse thin shim
# Delegates to Go binary: stdin JSON → bin/harness hook pre-tool → stdout JSON
"${CLAUDE_PLUGIN_ROOT}/bin/harness" hook pre-tool
