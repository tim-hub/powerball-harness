# Harness for Codex CLI

Codex CLI compatible distribution of Claude Code Harness.

## Setup

### Option 0: Path-based loading (experimental; verify on your Codex build)

No file copy needed. Add skill paths directly to `config.toml`:

```bash
git clone https://github.com/tim-hub/powerball-harness.git

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
path = "$(pwd)/claude-code-harness/codex/.codex/skills/harness-sync"
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

If your Codex build picks up `[[skills.config]]`, `git pull` updates them in place.
Because support can drift by Codex build, verify this on a fresh Codex process before using it as the only onboarding path for end users.

### Option 1: Script (recommended, user-based)

```bash
# Default: install to CODEX_HOME (user-based)
/path/to/claude-code-harness/scripts/setup-codex.sh --user
```

This is the reliable default for end users today.
After updating Harness, rerun the same script to sync `~/.codex/skills` to the latest `harness-*` bundle.

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
git clone https://github.com/tim-hub/powerball-harness.git

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
- Setup scripts keep backups in `$CODEX_HOME/backups/*` and move removed Harness skills out of `skills/` so Codex does not keep listing stale commands

## Runtime Behavior

- `$harness-plan`, `$harness-sync`, `$harness-work`, `$breezing`, and `$harness-review` are the primary Codex-facing workflow surfaces.
- Codex should be driven from the `harness-*` skill names, not legacy aliases like `$work`, `$plan-with-agent`, or `$verify`.
- `$harness-work` and `$breezing` use Codex native multi-agent orchestration.
- Native flow uses `spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`.
- `breezing` keeps Lead/Worker/Reviewer separation while reusing Codex-native subagents instead of older teammate-only wording.

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
