# Harness for Codex CLI

Codex CLI compatible distribution of Claude Code Harness.

## Setup

### Option 0: Path-based loading (recommended, Codex 0.110.0+)

No file copy needed. Add skill paths directly to `config.toml`:

```bash
git clone https://github.com/Chachamaru127/claude-code-harness.git

# Add to ~/.codex/config.toml (or .codex/config.toml for project-local):
cat >> "${CODEX_HOME:-$HOME/.codex}/config.toml" <<TOML

# Harness skills (path-based, no copy needed)
[[skills.config]]
path = "$(pwd)/claude-code-harness/codex/.codex/skills/harness-work"
enabled = true

[[skills.config]]
path = "$(pwd)/claude-code-harness/codex/.codex/skills/harness-plan"
enabled = true

[[skills.config]]
path = "$(pwd)/claude-code-harness/codex/.codex/skills/harness-review"
enabled = true

[[skills.config]]
path = "$(pwd)/claude-code-harness/codex/.codex/skills/harness-release"
enabled = true

[[skills.config]]
path = "$(pwd)/claude-code-harness/codex/.codex/skills/harness-setup"
enabled = true

[[skills.config]]
path = "$(pwd)/claude-code-harness/codex/.codex/skills/breezing"
enabled = true
TOML
```

This loads skills directly from the cloned repo — `git pull` updates them in place.

### Option 1: Script (full install, user-based)

```bash
# Default: install to CODEX_HOME (user-based)
/path/to/claude-code-harness/scripts/setup-codex.sh --user
```

Project-local install is still available:

```bash
/path/to/claude-code-harness/scripts/setup-codex.sh --project
```

### Option 1.5: Claude Code (in-session)

If you use Claude Code Harness, run:

```bash
/setup codex
```

### Option 2: Manual (user-based)

```bash
git clone https://github.com/Chachamaru127/claude-code-harness.git

CODEX_HOME="${CODEX_HOME:-$HOME/.codex}"
BACKUP_ROOT="$CODEX_HOME/backups/manual-codex-setup"
mkdir -p "$CODEX_HOME/skills" "$CODEX_HOME/rules" "$BACKUP_ROOT"

# Prevent duplicate skill listings from legacy backup/archive directories.
for legacy in "$CODEX_HOME/skills"/_archived "$CODEX_HOME/skills"/*.backup.*; do
  [ -e "$legacy" ] || continue
  mv "$legacy" "$BACKUP_ROOT/"
done

for entry in claude-code-harness/codex/.codex/skills/*; do
  name="$(basename "$entry")"
  case "$name" in
    _archived|*.backup.*) continue ;;
  esac
  rm -rf "$CODEX_HOME/skills/$name"
  cp -R "$entry" "$CODEX_HOME/skills/"
done
cp -R claude-code-harness/codex/.codex/rules/* "$CODEX_HOME/rules/"
cp claude-code-harness/codex/.codex/config.toml "$CODEX_HOME/config.toml"
```

## Codex Multi-Agent Defaults

- `features.multi_agent = true`
- Harness role declarations are installed under `[agents.*]`
- Setup scripts always ensure `multi_agent` + role defaults in target `config.toml`
- Setup scripts keep backups in `$CODEX_HOME/backups/*` so Codex does not list old skills

## Runtime Behavior

- `$harness-work` and `$harness-review` are the primary Codex-facing workflow surfaces.
- `$work` / `$breezing` default to Codex native multi-agent orchestration.
- Native flow uses `spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`.
- `--claude` switches both implementation and review to Claude delegation.
- `--claude + --codex-review` is invalid and should fail before execution.

## State Path

Harness runtime state is written under:

```text
${CODEX_HOME:-~/.codex}/state/harness/
```

## Rules

`$CODEX_HOME/rules/harness.rules` provides command guardrails.

## Notes

- Codex reads skills from `$CODEX_HOME/skills/<skill-name>/SKILL.md`.
- Project-local `.codex/skills` overrides user-level skills.
