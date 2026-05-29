# OpenCode Environment Notes

Reference for running Worker or Reviewer inside OpenCode CLI environments.

## Invocation via Official Plugin `opencode-plugin-cc`

When calling OpenCode from Claude Code, execute via the official plugin:

```bash
# Task delegation (implementation, debugging, investigation)
bash scripts/opencode-companion.sh task "task content"

# Review
bash scripts/opencode-companion.sh review --base "${TASK_BASE_REF}"

# Setup check
/opencode:setup
```

> **Note**: Direct invocation of raw `opencode run` is prohibited.
> See `harness/rules/opencode-cli-only.md` (OpenCode Plugin Policy) for details.

## Tool Fallbacks When Running Inside OpenCode CLI

In OpenCode CLI environments, Claude Code's native Task tools are unavailable.
Use these fallbacks:

| Normal Environment | OpenCode Fallback |
|---|---|
| Get task list with `TaskList` | `harness plan-cli list --status cc:TODO,cc:WIP --pretty` |
| Update status with `TaskUpdate` | `harness plan-cli update <task-id> --status cc:WIP` |
| Write review result to Task | Output review result to stdout in markdown format |

### Detection

```bash
if [ "${OPENCODE_CLI:-}" = "1" ]; then
  # OpenCode environment: use harness plan-cli fallback
fi
```

> **Important**: `OPENCODE_CLI` is a harness convention — it is **NOT** automatically set
> by opencode-plugin-cc or the OpenCode CLI itself. The caller (harness skill or agent) sets it
> before delegating if needed. As an alternative, use a presence-check
> (`command -v opencode`) to detect the OpenCode environment.

### Review Output in OpenCode Environment

Since Task tools are not supported, review results are output to stdout in markdown format.
The Lead agent or user reads the results and decides the next action.

## Operation within OpenCode CLI (Incompatibilities)

The following Claude Code features are unavailable inside OpenCode sessions.

### memory frontmatter

```yaml
memory: project  # Claude Code only. Ignored in OpenCode
```

Alternatives in OpenCode environment:
- Document learnings in AGENTS.md (project root, loaded by OpenCode automatically)
- Use opencode.json's provider/model config to persist session preferences

### skills field

```yaml
skills:
  - harness-work  # References Claude Code's skills/ directory. Incompatible in OpenCode
```

Alternatives in OpenCode environment:
- Skills installed via `harness-setup opencode` are available in `.opencode/skills/`
- Reference them directly in session instructions

### Task Tool

OpenCode does not expose Claude Code's `TaskCreate` / `TaskUpdate` / `TaskList` tools.
State management is done via `harness plan-cli` subcommands (read: `list`/`get`; write: `update`/`add-task`).
