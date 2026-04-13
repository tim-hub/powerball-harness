---
name: harness-mem
description: "Alias for the `memory` skill. Use when user invokes /harness-mem, types `harness-mem`, or mentions the Harness memory DB, MCP memory tools, recording a decision, searching memory, or cross-tool memory. This skill delegates to `skills/memory/SKILL.md` — all logic lives there, not here."
allowed-tools: ["Read", "Write", "Edit", "Bash", "Grep", "mcp__harness__harness_mem_*"]
argument-hint: "[ssot|sync|sync-across|migrate|merge|search|record]"
context: fork
---

# Harness Memory (alias for `memory`)

`harness-mem` is a user-facing alias for the canonical `memory` skill. It exists so that invocations like `/harness-mem`, `harness-mem record …`, or phrases mentioning "harness-mem" resolve cleanly without the model having to guess which skill to open.

## Action

Read the canonical skill and execute it with the argument you received:

1. Read `skills/memory/SKILL.md`
2. Dispatch the subcommand (`ssot`, `sync`, `sync-across`, `migrate`, `merge`, `search`, `record`) exactly as that skill's Quick Reference table specifies
3. Do **not** duplicate any logic here — if a behavior needs to change, change it in `skills/memory/SKILL.md` so both entry points stay in sync automatically

Effective dispatch:

```
/harness-mem <subcommand> [args]   →   skills/memory/SKILL.md  subcommand=<subcommand>
```

## Related

- `skills/memory/SKILL.md` — canonical implementation (all subcommand logic lives here)
- `skills/memory/references/` — subcommand reference files (`record.md`, `ssot-initialization.md`, etc.)
