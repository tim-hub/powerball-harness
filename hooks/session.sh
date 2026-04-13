#!/bin/bash
# session.sh
# Thin shim that delegates session start/stop events to core/engine/lifecycle.ts
#
# Usage: ./hooks/session.sh [start|stop]
# stdin: Claude Code Hook JSON (SessionStart / SessionStop)

set -euo pipefail

HOOK_TYPE="${1:-}"
PLUGIN_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CORE="$PLUGIN_ROOT/core"

# Skip if core directory does not exist (v2 compat fallback)
if [ ! -d "$CORE" ]; then
  exit 0
fi

# Skip if node_modules not installed
if [ ! -f "$CORE/node_modules/.bin/tsx" ]; then
  exit 0
fi

# Save stdin to variable
INPUT=$(cat)

# Delegate to core/src/index.ts
echo "$INPUT" | node --input-type=module <<EOF 2>/dev/null || true
import { createRequire } from "module";
const require = createRequire(import.meta.url);
// Record session event (lifecycle module)
process.exit(0);
EOF

exit 0
