#!/bin/bash
# check-codex.sh - Codex availability check (for once hook)
# Executed once on first /harness-review run
#
# Usage: ./scripts/check-codex.sh

set -euo pipefail

# Project configuration file path
CONFIG_FILE=".claude-code-harness.config.yaml"

# Check if codex.enabled is already set
if [[ -f "$CONFIG_FILE" ]]; then
    if grep -q "codex:" "$CONFIG_FILE" 2>/dev/null; then
        # Do nothing if already configured
        exit 0
    fi
fi

# Check if Codex CLI is installed
if ! command -v codex &> /dev/null; then
    # Do nothing if Codex is not found
    exit 0
fi

# Get Codex version
CODEX_VERSION=$(codex --version 2>/dev/null | head -1 || echo "unknown")

# Get latest version from npm (3-second timeout)
LATEST_VERSION=$(npm show @openai/codex version 2>/dev/null || echo "unknown")

# Version comparison function
version_lt() {
    [ "$1" != "$2" ] && [ "$(printf '%s\n' "$1" "$2" | sort -V | head -n1)" = "$1" ]
}

# Notify user when Codex is found
cat << EOF

🤖 Codex detected

**Installed version**: ${CODEX_VERSION}
**Latest version**: ${LATEST_VERSION}
EOF

# Warn if version is outdated
if [[ "$LATEST_VERSION" != "unknown" && "$CODEX_VERSION" != "unknown" ]]; then
    # Extract numeric portion from version string
    CURRENT_NUM=$(echo "$CODEX_VERSION" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "0.0.0")
    LATEST_NUM=$(echo "$LATEST_VERSION" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "0.0.0")

    if version_lt "$CURRENT_NUM" "$LATEST_NUM"; then
        cat << EOF

⚠️ **Codex CLI is outdated**

To update:
\`\`\`bash
npm update -g @openai/codex
\`\`\`

Or ask Claude to update Codex for you.

EOF
    fi
fi

# timeout / gtimeout check (macOS compatibility)
TIMEOUT_CMD=""
if command -v timeout &> /dev/null; then
    TIMEOUT_CMD="timeout"
elif command -v gtimeout &> /dev/null; then
    TIMEOUT_CMD="gtimeout"
fi

if [[ -z "$TIMEOUT_CMD" ]]; then
    cat << 'EOF'

⚠️ **timeout command not found**

Codex CLI parallel review uses the `timeout` command for timeout control.
It is not included by default on macOS. Install with:

```bash
brew install coreutils
```

This enables `gtimeout`, which Harness auto-detects.
Codex works without it, but timeout control will not be available.

EOF
else
    echo ""
    echo "**Timeout command**: \`${TIMEOUT_CMD}\` ✅"
fi

cat << 'EOF'

To enable second-opinion review:

```yaml
# .claude-code-harness.config.yaml
review:
  codex:
    enabled: true
    model: gpt-5.2-codex  # Recommended model
```

Or run a Codex review individually with `/codex-review`

Details: skills/codex-review/SKILL.md

EOF

exit 0
