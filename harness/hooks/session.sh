#!/bin/bash
# session.sh — Harness v4 SessionStart thin shim
# Delegates to Go binary: stdin JSON → bin/harness hook session-start → stdout JSON
"${CLAUDE_PLUGIN_ROOT}/bin/harness" hook session-start
