#!/bin/bash
# permission.sh — Harness v3 PermissionRequest thin shim
# stdin JSON → core engine → stdout JSON
node "${CLAUDE_PLUGIN_ROOT}/core/dist/index.js" permission
