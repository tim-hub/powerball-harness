#!/usr/bin/env bash
# harness-setup codex — Set up Codex CLI with Harness skills
# Copies templates and skills from the plugin to the project's .codex/ directory.
#
# Usage: bash setup-codex.sh [--plugin-root PATH]
#
# Env:
#   CLAUDE_PLUGIN_ROOT  — set automatically by Claude Code when running inside a skill

set -euo pipefail

PLUGIN_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/../../../../" && pwd)}"
PROJECT_ROOT="$(pwd)"

# ─── 1. Check codex is installed ─────────────────────────────────────────────
if ! command -v codex &>/dev/null; then
  echo "❌  Codex CLI not found."
  echo ""
  echo "Install it with:"
  echo "  npm install -g @openai/codex"
  echo ""
  echo "Then re-run: harness-setup codex"
  exit 1
fi

CODEX_VERSION="$(codex --version 2>/dev/null || echo 'unknown')"
echo "✓  Codex CLI found: $CODEX_VERSION"

# ─── 2. Create .codex/ structure ─────────────────────────────────────────────
CODEX_DIR="$PROJECT_ROOT/.codex"
mkdir -p "$CODEX_DIR/rules" "$CODEX_DIR/skills"

# ─── 3. Copy template files: config, rules, .codexignore, AGENTS.md, README ──
TEMPLATE_DIR="$PLUGIN_ROOT/templates/codex"

if [[ ! -d "$TEMPLATE_DIR" ]]; then
  echo "❌  Template directory not found: $TEMPLATE_DIR"
  exit 1
fi

cp "$TEMPLATE_DIR/config.toml"         "$CODEX_DIR/config.toml"
cp "$TEMPLATE_DIR/rules/harness.rules" "$CODEX_DIR/rules/harness.rules"
cp "$TEMPLATE_DIR/.codexignore"        "$PROJECT_ROOT/.codexignore"
cp "$TEMPLATE_DIR/AGENTS.md"           "$PROJECT_ROOT/AGENTS.md"
echo "✓  Copied Codex config, rules, .codexignore, AGENTS.md"

# ─── 4. Copy skills/ → .codex/skills/ with disable-model-invocation patch ────
SKILLS_DIR="$PLUGIN_ROOT/skills"

if [[ ! -d "$SKILLS_DIR" ]]; then
  echo "❌  Plugin skills directory not found: $SKILLS_DIR"
  exit 1
fi

patch_skill_md() {
  local src="$1"
  local dst="$2"
  # If disable-model-invocation already present, copy as-is
  if grep -q "^disable-model-invocation:" "$src"; then
    cp "$src" "$dst"
  else
    # Insert 'disable-model-invocation: true' after the first '---' line
    awk 'NR==1 && /^---$/ { print; print "disable-model-invocation: true"; next } { print }' "$src" > "$dst"
  fi
}

skill_count=0
for skill_dir in "$SKILLS_DIR"/*/; do
  skill_name="$(basename "$skill_dir")"
  dst_skill_dir="$CODEX_DIR/skills/$skill_name"
  mkdir -p "$dst_skill_dir"

  # Copy all files in the skill directory
  cp -r "$skill_dir"* "$dst_skill_dir/" 2>/dev/null || true

  # Patch the main SKILL.md to add disable-model-invocation
  if [[ -f "$skill_dir/SKILL.md" ]]; then
    patch_skill_md "$skill_dir/SKILL.md" "$dst_skill_dir/SKILL.md"
  fi

  skill_count=$((skill_count + 1))
done

echo "✓  Copied $skill_count skills to .codex/skills/ (with disable-model-invocation patch)"

# ─── 5. Overlay codex-native skill overrides ─────────────────────────────────
CODEX_SKILLS_DIR="$PLUGIN_ROOT/templates/codex-skills"

if [[ -d "$CODEX_SKILLS_DIR" ]]; then
  override_count=0
  for override_dir in "$CODEX_SKILLS_DIR"/*/; do
    skill_name="$(basename "$override_dir")"
    dst_override_dir="$CODEX_DIR/skills/$skill_name"
    mkdir -p "$dst_override_dir"
    cp -r "$override_dir"* "$dst_override_dir/" 2>/dev/null || true
    override_count=$((override_count + 1))
  done
  echo "✓  Applied $override_count codex-native skill overrides from templates/codex-skills/"
fi

# ─── Done ─────────────────────────────────────────────────────────────────────
echo ""
echo "✅  Codex setup complete!"
echo ""
echo "Files created:"
echo "  .codex/config.toml       — Multi-agent configuration"
echo "  .codex/rules/harness.rules — Guardrail rules"
echo "  .codexignore             — Codex ignore patterns"
echo "  AGENTS.md                — Agent role reference"
echo "  .codex/skills/           — Harness skills ($skill_count skills)"
echo ""
echo "Next steps:"
echo "  1. Review .codex/config.toml and adjust agent profiles if needed"
echo "  2. Run: codex"
echo "  3. Or delegate from Claude Code: bash \"\${CLAUDE_PLUGIN_ROOT}/scripts/codex-companion.sh\" task --write \"your prompt\""
