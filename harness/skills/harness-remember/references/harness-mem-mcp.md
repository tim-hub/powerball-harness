# Unified Harness Memory — MCP Tools

The `harness_mem_*` MCP tools let Claude Code, Codex, and OpenCode share a runtime memory DB that lives **outside** the git-committed SSOT layer. Use them when a question needs cross-tool context or short-lived session state — things that don't belong in `decisions.md` or `patterns.md` but are still worth recalling across sessions.

> **Load this file only when**: the Harness MCP server is connected (tools prefixed `mcp__harness__harness_mem_*` are available) **and** the user's task involves cross-tool memory, shared DB search, session replay, or checkpoint/event recording.

## At a glance

- **Search**: `harness_mem_search`, `harness_mem_timeline`, `harness_mem_get_observations`
- **Injection**: `harness_mem_resume_pack`
- **Recording**: `harness_mem_record_checkpoint`, `harness_mem_record_event`, `harness_mem_finalize_session`

## Two layers, one vocabulary

| Layer | Where it lives | Authoring path | Best for |
|-------|----------------|----------------|----------|
| **SSOT (Layer 2)** | `.claude/memory/decisions.md`, `patterns.md` — committed to git | `memory ssot` / `memory record` | Durable decisions and reusable patterns the whole team needs |
| **Shared DB (this file)** | Harness MCP server (SQLite under the hood) | `harness_mem_record_*` MCP tools | Ephemeral cross-tool state: checkpoints, session timelines, free-form observations |

`/harness-remember sync` is the bridge: it promotes shared-DB observations into the SSOT layer when they matter long-term.

## Tool catalog

### Search and recall (read-only)

| Tool | Use when |
|------|----------|
| `harness_mem_search` | Free-text / keyword search across all recorded observations, regardless of which tool produced them |
| `harness_mem_timeline` | Chronological view — "what happened in this project between date X and Y" |
| `harness_mem_get_observations` | Fetch observations by id or filter (agent, session, tag) |
| `harness_mem_resume_pack` | Inject a compact resume bundle into the current session's context — use at session start when the user says "pick up where I left off" |

### Recording (write)

| Tool | Use when |
|------|----------|
| `harness_mem_record_checkpoint` | Save a durable "I reached this state" marker with enough context to resume later |
| `harness_mem_record_event` | Log a discrete event (task completed, decision drafted, error encountered). Lightweight compared to a checkpoint |
| `harness_mem_finalize_session` | Close out a session — triggers summarization and flushes buffered events |

## When to prefer MCP over file-based SSOT

| Situation | Prefer |
|-----------|--------|
| The fact is durable, team-visible, and would affect future decisions | **File SSOT** (`memory record`) |
| The fact spans tools (Claude / Codex / OpenCode) and is session-scoped | **MCP shared DB** |
| The user asks "what did we try last time?" across tools | `harness_mem_search` + `harness_mem_timeline` |
| Recording "we hit this issue" during a working session | `harness_mem_record_event` (cheap) |
| Marking a natural pause point with enough context to resume | `harness_mem_record_checkpoint` |
| Pulling Layer 1 observations into Layer 2 | `/harness-remember sync` (combines both layers) |

## Extension behavior: how `search` and `record` fan out

The `memory` skill treats local SSOT as authoritative and MCP as an extension. The table below is the contract the subcommands follow — it should match what's done in practice:

| Subcommand | Local stage (always runs first) | MCP extension (only if server is connected) |
|------------|----------------------------------|---------------------------------------------|
| `search`   | `rg`/`grep` over `.claude/memory/` and `.claude/agent-memory/` — returns file hits | `harness_mem_search` — returns shared-DB observations across Claude / Codex / OpenCode. Present results under a separate "cross-tool" header |
| `record`   | Validate SSOT-worthiness, append to `decisions.md` / `patterns.md`, update the index | `harness_mem_record_event(kind="ssot_entry", …)` — log a pointer to the new entry so other agents can discover it via `harness_mem_search` |

Two invariants:

1. **Local first.** The local step always runs before MCP. If local fails, stop — don't mask the failure behind an MCP success.
2. **Honest reporting.** Tell the user which layers ran. If MCP was unavailable, say so; don't silently degrade coverage without mentioning it.

### Example for `record`

After a successful local append of `D23` to `decisions.md`:

```
mcp__harness__harness_mem_record_event(
  kind     = "ssot_entry",
  agent    = "<current agent id>",
  payload  = {
    "id":      "D23",
    "type":    "decision",
    "file":    ".claude/memory/decisions.md",
    "title":   "Opus for high-effort tasks",
    "summary": "1-2 line summary",
    "commit":  "pending"
  }
)
```

Only a pointer plus metadata is mirrored — not the full body. The SSOT file remains the source of truth; the shared DB just makes the entry findable by other tools.

### Example for `search`

```
[Local stage] grep ".claude/memory/" for "ultrathink"
  → 3 hits in patterns.md

[MCP extension] harness_mem_search("ultrathink")
  → 2 additional hits from Codex sessions (checkpoints, events)

Present:
  ## Local SSOT hits (authoritative)
  .claude/memory/patterns.md:42:  ...
  .claude/memory/patterns.md:87:  ...
  .claude/memory/patterns.md:103: ...

  ## Cross-tool hits (shared DB)
  codex/session-2026-04-09 #checkpoint: ...
  codex/session-2026-04-12 #event: ...
```

## Session-start recall (`harness_mem_resume_pack`)

When the user says "continue from where we left off" at session start, call `harness_mem_resume_pack` early. The returned bundle gets injected into context; don't dump it verbatim — use it to orient your own reasoning.

## Promotion from Layer 1 observations into SSOT

If an MCP-only observation turns out to be important enough to commit:

1. Run `/harness-remember sync` — it scans Layer 1 observations and proposes promotions.
2. Review each proposed entry.
3. Approve the ones worth preserving — they go through the `record` validation gate and land in `decisions.md` or `patterns.md`, then `record` mirrors a pointer back to MCP.

Don't bypass `record` validation just because the observation came from MCP. The gate exists to keep SSOT clean regardless of source.

## Failure modes

- **MCP server not connected**: None of the tools above are callable. Fall back to file-based operations (grep over `.claude/memory/` for search, manual append for record). Warn the user clearly — don't silently degrade.
- **Stale observations**: MCP DB entries can become wrong over time. Before quoting an MCP result to the user, spot-check against current code (see "Before recommending from memory" in the global memory instructions).
- **Duplicate recording**: If the same observation might be recorded twice (once as a checkpoint, once as an event), prefer the checkpoint and drop the event. Checkpoints supersede events for the same moment.

## Related

- `skills/harness-remember/SKILL.md` — entry point (SSOT authoring layer)
- `references/record.md` — the validation gate that runs before any SSOT write
- `references/sync-ssot-from-memory.md` — Layer 1 → Layer 2 promotion flow
