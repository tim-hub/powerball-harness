#!/bin/bash
# permission.sh — Harness v3 PermissionRequest 薄いシム
# stdin JSON → core エンジン → stdout JSON
node "${CLAUDE_PLUGIN_ROOT}/core/dist/index.js" permission
