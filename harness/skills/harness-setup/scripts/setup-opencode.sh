#!/usr/bin/env bash
# harness-setup opencode — Set up OpenCode with Harness skills
# Copies templates and skills from the plugin to the project's .opencode/ directory.
#
# Usage: bash setup-opencode.sh [--plugin-root PATH]
#
# Env:
#   CLAUDE_PLUGIN_ROOT  — set automatically by Claude Code when running inside a skill

set -euo pipefail

PLUGIN_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/../../../../" && pwd)}"
PROJECT_ROOT="$(pwd)"

# ─── 1. Check opencode is installed ──────────────────────────────────────────
if ! command -v opencode &>/dev/null; then
  echo "❌  OpenCode not found."
  echo ""
  echo "Install it from: https://opencode.ai/"
  echo ""
  echo "Then re-run: harness-setup opencode"
  exit 1
fi

OC_VERSION="$(opencode --version 2>/dev/null || echo 'unknown')"
echo "✓  OpenCode found: $OC_VERSION"

# ─── 2. Create .opencode/ structure ──────────────────────────────────────────
OC_DIR="$PROJECT_ROOT/.opencode"
mkdir -p "$OC_DIR/commands" "$OC_DIR/skills"

# ─── 3. Copy template config files ───────────────────────────────────────────
TEMPLATE_DIR="$PLUGIN_ROOT/templates/opencode"

if [[ ! -d "$TEMPLATE_DIR" ]]; then
  echo "❌  Template directory not found: $TEMPLATE_DIR"
  exit 1
fi

# Copy opencode.json
cp "$TEMPLATE_DIR/opencode.json" "$OC_DIR/opencode.json"
echo "✓  Copied opencode.json"

# Copy commands
cp -r "$TEMPLATE_DIR/commands/". "$OC_DIR/commands/"
cmd_count=$(find "$OC_DIR/commands" -name '*.md' | wc -l | tr -d ' ')
echo "✓  Copied $cmd_count commands to .opencode/commands/"

# Copy AGENTS.md to project root
cp "$TEMPLATE_DIR/AGENTS.md" "$PROJECT_ROOT/AGENTS.md"
echo "✓  Copied AGENTS.md"

# ─── 4. Copy skills/ → .opencode/skills/ as-is ───────────────────────────────
SKILLS_DIR="$PLUGIN_ROOT/skills"

if [[ ! -d "$SKILLS_DIR" ]]; then
  echo "❌  Plugin skills directory not found: $SKILLS_DIR"
  exit 1
fi

skill_count=0
for skill_dir in "$SKILLS_DIR"/*/; do
  skill_name="$(basename "$skill_dir")"
  dst_skill_dir="$OC_DIR/skills/$skill_name"
  mkdir -p "$dst_skill_dir"
  cp -r "$skill_dir"* "$dst_skill_dir/" 2>/dev/null || true
  skill_count=$((skill_count + 1))
done

echo "✓  Copied $skill_count skills to .opencode/skills/"

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo "✅  OpenCode setup complete!"
echo ""
echo "Files created:"
echo "  .opencode/opencode.json  — OpenCode project configuration"
echo "  .opencode/commands/      — Harness slash commands ($cmd_count commands)"
echo "  .opencode/skills/        — Harness skills ($skill_count skills)"
echo "  AGENTS.md                — Agent role reference"
echo ""
echo "Next steps:"
echo "  1. Review .opencode/opencode.json and adjust settings if needed"
echo "  2. Run: opencode"
