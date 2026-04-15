# Team Composition — Codex CLI Environment

Reference for running Harness team patterns in Codex CLI environments.

## Bridge via Official Plugin `codex-plugin-cc`

When calling Codex from Claude Code, execute via the official plugin.
Direct invocation of raw `codex exec` is prohibited (see `.claude/rules/codex-cli-only.md`).

```bash
# Task delegation (implementation, debugging, investigation)
bash scripts/codex-companion.sh task --write "task content"

# Review
bash scripts/codex-companion.sh review --base "${TASK_BASE_REF}"

# Setup check (user interaction)
/codex:setup

# Task delegation (user interaction)
/codex:rescue "investigate the bug"
```

## API Mapping

| Claude Code | Codex CLI (native) | codex-plugin-cc | Notes |
|------------|-----------|-----------------|------|
| `Agent(subagent_type=...)` | `spawn_agent(...)` | `codex:codex-rescue` agent | CC->Codex via plugin |
| `SendMessage(to, message)` | `send_input(agent_id, message)` | — | Codex internal only |
| `bypassPermissions` | `--full-auto` | `task --write` | companion configures sandbox |
| Task tool | Direct Plans.md editing | — | |
| PreToolUse hooks | config.toml sandbox | — | |
| Codex review | `codex exec --sandbox read-only` | `review --base <ref>` | Structured output |

> **Important**: The Phase B flow of harness-work / breezing (Worker spawn -> review -> cherry-pick) is
> described using Claude Code's `Agent` + `SendMessage` syntax.
> In Codex environments, refer to the mapping table above for equivalent operations.

## Sequential Companion Pattern

```bash
# Worker equivalent (implementation tasks)
echo "task content" | bash scripts/codex-companion.sh task --write

# Reviewer equivalent (read-only review)
bash scripts/codex-companion.sh review --base "${BASE_REF}"
```

## Parallel Execution (Bash Level)

```bash
echo "Task A" | bash scripts/codex-companion.sh task --write > /tmp/out-a.txt &
echo "Task B" | bash scripts/codex-companion.sh task --write > /tmp/out-b.txt &
wait
```

Tasks without dependencies can be parallelized using Bash's `&` + `wait`.
Avoid parallel writes to the same file.

## Thread Management (Official Plugin)

```bash
bash scripts/codex-companion.sh task --resume-last --write "continue where you left off"
bash scripts/codex-companion.sh status
bash scripts/codex-companion.sh result <job-id>
bash scripts/codex-companion.sh cancel <job-id>
```

## codex exec Flag Reference (codex-cli 0.115.0+)

> Harness uses `scripts/codex-companion.sh` rather than raw `codex exec`.
> The following applies only within Codex native skills (`templates/codex-skills/`).

| Flag | Short Form | Description |
|---|---|---|
| `--sandbox` | `-s` | `read-only` / `workspace-write` / `danger-full-access` |
| `--full-auto` | - | Alias for `-a on-request` + `--sandbox workspace-write` |

Prompt input: `cat prompt.md | bash scripts/codex-companion.sh task --write` (stdin) or direct arg.

## Configurable Memory — Codex-Side Mapping

| Claude Code | Codex CLI | Notes |
|---|---|---|
| `memory: project` MEMORY.md | `AGENTS.md` hierarchy (global -> project -> subdir) | Persistent instructions |
| agent-memory directory | `agents.<name>.config_file` | Per-agent config files |
| spawn prompt | `AGENTS.override.md` | Temporary overrides |
| Session history | `history.persistence: save-all` | Saved to `history.jsonl` |

**Usage**: Consolidate project-specific learnings in `.codex/AGENTS.md` and periodically promote
`codex-learnings.md` content to AGENTS.md (SSOT maintenance).
