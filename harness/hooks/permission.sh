#!/bin/bash
# permission.sh — Harness v4 PermissionRequest thin shim
# Delegates to Go binary: stdin JSON → bin/harness hook permission → stdout JSON
"${CLAUDE_PLUGIN_ROOT}/bin/harness" hook permission
