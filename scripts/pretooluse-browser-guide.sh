#!/bin/bash
# pretooluse-browser-guide.sh
# Hook to suggest agent-browser when using MCP browser tools
#
# Target tools:
#   - mcp__chrome-devtools__*
#   - mcp__playwright__* / mcp__plugin_playwright__*
#
# Behavior:
#   - If agent-browser is installed, recommend its use
#   - Non-blocking (informational only)
#
# Input: stdin JSON from Claude Code hooks (pre-filtered by matcher)
# Output: JSON with hookSpecificOutput format

set -euo pipefail

# Read JSON from stdin
INPUT=""
if [ ! -t 0 ]; then
  INPUT="$(cat 2>/dev/null)"
fi

# Do nothing if no input
[ -z "$INPUT" ] && exit 0

# Check if agent-browser is installed
if command -v agent-browser &> /dev/null; then
  # Output recommendation message (hookSpecificOutput format)
  # Already filtered to MCP browser tools by matcher, no additional tool name check needed
  if command -v jq >/dev/null 2>&1; then
    CONTEXT="💡 **Consider using agent-browser first**

agent-browser is a browser automation tool optimized for AI agents.

\`\`\`bash
# Basic usage
agent-browser open <url>
agent-browser snapshot -i -c  # AI-optimized snapshot
agent-browser click @e1        # Click by element reference
\`\`\`

The current MCP tools are also available, but agent-browser is simpler and faster.

Details: \`docs/OPTIONAL_PLUGINS.md\`"

    jq -nc --arg ctx "$CONTEXT" '{
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        additionalContext: $ctx
      }
    }'
  else
    # Try Python when jq is unavailable
    if command -v python3 >/dev/null 2>&1; then
      python3 - <<'PY'
import json
context = """💡 **Consider using agent-browser first**

agent-browser is a browser automation tool optimized for AI agents.

```bash
# Basic usage
agent-browser open <url>
agent-browser snapshot -i -c  # AI-optimized snapshot
agent-browser click @e1        # Click by element reference
```

The current MCP tools are also available, but agent-browser is simpler and faster.

Details: `docs/OPTIONAL_PLUGINS.md`"""
print(json.dumps({
    "hookSpecificOutput": {
        "hookEventName": "PreToolUse",
        "additionalContext": context
    }
}))
PY
    fi
  fi
fi

# Normal exit when agent-browser not installed or output complete
exit 0
