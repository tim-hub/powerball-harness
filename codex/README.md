# Harness for Codex CLI

Codex CLI compatible distribution of Claude Code Harness.

## Setup

### Option 0: Path-based loading (experimental)

No file copy needed. Add skill paths directly to `config.toml`:

```bash
git clone https://github.com/tim-hub/powerball-harness.git

# Add to ~/.codex/config.toml (or .codex/config.toml for project-local):
cat >> "${CODEX_HOME:-$HOME/.codex}/config.toml" <<TOML
[[skills.config]]
path = "$(pwd)/powerball-harness/codex/.codex/skills/harness-work"
enabled = true
TOML
```

### Option 1: Script (recommended)

```bash
/path/to/powerball-harness/scripts/setup-codex.sh --user
```

### Option 2: Manual

```bash
git clone https://github.com/tim-hub/powerball-harness.git

CODEX_HOME="${CODEX_HOME:-$HOME/.codex}"
mkdir -p "$CODEX_HOME/skills" "$CODEX_HOME/rules"

for entry in powerball-harness/codex/.codex/skills/*; do
  name="$(basename "$entry")"
  rm -rf "$CODEX_HOME/skills/$name"
  cp -R "$entry" "$CODEX_HOME/skills/"
done
cp -R powerball-harness/codex/.codex/rules/* "$CODEX_HOME/rules/"
cp powerball-harness/codex/.codex/config.toml "$CODEX_HOME/config.toml"
```

## Skill Architecture

Skills are **symlinks** to `../../../skills/` — single source of truth, no duplication.

```
codex/.codex/skills/
├── harness-plan    → ../../../skills/harness-plan
├── harness-work    → ../../../skills/harness-work
├── harness-review  → ../../../skills/harness-review
├── harness-release → ../../../skills/harness-release
├── harness-setup   → ../../../skills/harness-setup
├── harness-sync    → ../../../skills/harness-sync
├── breezing        → ../../../skills/breezing
└── memory          → ../../../skills/memory
```

## Runtime

- `$harness-plan`, `$harness-work`, `$breezing`, `$harness-review` are the primary surfaces
- `$harness-work` and `$breezing` use Codex native multi-agent orchestration

## Rules

`$CODEX_HOME/rules/harness.rules` provides command guardrails.
