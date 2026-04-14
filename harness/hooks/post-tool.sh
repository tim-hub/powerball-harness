#!/bin/bash
# post-tool.sh — Harness v4 PostToolUse thin shim
# Delegates to Go binary: stdin JSON → bin/harness hook post-tool → stdout JSON
"${CLAUDE_PLUGIN_ROOT}/bin/harness" hook post-tool
