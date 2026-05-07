# Codex Orchestration in Breezing

## Codex Mode (`--codex`) — Implementation

Delegates all implementation to Codex CLI via the official plugin `codex-plugin-cc`:

```bash
# Task delegation (writable)
bash "${CLAUDE_SKILL_DIR}/../../scripts/codex-companion.sh" task --write "task content"

# Via stdin (for large prompts)
CODEX_PROMPT=$(mktemp /tmp/codex-prompt-XXXXXX.md)
# Write task content
cat "$CODEX_PROMPT" | bash "${CLAUDE_SKILL_DIR}/../../scripts/codex-companion.sh" task --write
rm -f "$CODEX_PROMPT"
```

## Codex Native Orchestration

Codex uses native subagents.
Key control surfaces are `spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`.

> **Claude Code vs Codex communication API** (SSOT: API mapping table in `team-composition.md`):
> - Claude Code: `SendMessage(to: agentId, message: "...")` to send fix instructions to Workers
> - Codex: `resume_agent(agent_id)` to resume Workers → `send_input(agent_id, "...")` to send instructions
>
> Pseudo-code in harness-work is written in Claude Code syntax. Translate to the above when running in a Codex environment.
