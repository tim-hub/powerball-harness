# Codex Environment Notes

Reference for running Worker inside Codex CLI environments.

## Invocation via Official Plugin `codex-plugin-cc`

When calling Codex from Claude Code, execute via the official plugin:

```bash
# Task delegation (implementation, debugging, investigation)
bash scripts/codex-companion.sh task --write "task content"

# Review
bash scripts/codex-companion.sh review --base "${TASK_BASE_REF}"

# Setup check
/codex:setup
```

> **Note**: Direct invocation of raw `codex exec` is prohibited.
> See `.claude/rules/codex-cli-only.md` (Codex Plugin Policy) for details.

## Operation within Codex CLI (Incompatibilities)

The following features are incompatible in Codex CLI environments (skills within `templates/codex-skills/`).

### memory frontmatter

```yaml
memory: project  # Claude Code only. Ignored in Codex
```

Alternatives in Codex environment:
- Document learnings in INSTRUCTIONS.md (project root)
- Use `config.toml`'s `[notify] after_agent` to write out memory at session end

### skills field

```yaml
skills:
  - harness-work  # References Claude Code's skills/ directory. Incompatible with Codex
  - harness-review
```

Alternatives in Codex environment:
- Call Codex skills using `$skill-name` syntax (e.g., `$harness-work`)
- Place skills in `~/.codex/skills/` or `.codex/skills/`

### Task Tool

Worker's `disallowedTools: [Agent]` is a Claude Code constraint (Task renamed to Agent in v2.1.63).
In Codex environment, the Task tool itself does not exist, so state management is done by directly Read/Edit-ing Plans.md.
