# Session Memory

Manages cross-session learning and memory. Records and references past work, decisions, and learned patterns.

**Memory storage policy**: `decisions.md` and `patterns.md` are for Git sharing (project-wide SSOT). `session-log.md`, `context.json`, and `.claude/state/` should stay local (not committed) — they are prone to noise and bloat.

---

## Memory Structure

```
.claude/
├── memory/
│   ├── session-log.md      # Per-session log
│   ├── decisions.md        # Important decisions
│   ├── patterns.md         # Learned patterns
│   └── context.json        # Project context
└── state/
    └── agent-trace.jsonl   # Agent Trace (tool execution history)
```

**SSOT vs Local separation**:
- **SSOT (Git-shared)**: `decisions.md` / `patterns.md` — each entry needs a title + tags with an Index at the top
- **Local**: `session-log.md` / `context.json` / `.claude/state/` — noise-prone; keep out of Git unless needed

---

## Processing Flow

### At Session Start

1. Load `.claude/memory/context.json`
2. Review previous session log
3. Retrieve recent edit history from Agent Trace:
   ```bash
   tail -50 .claude/state/agent-trace.jsonl | jq -r '.files[].path' | sort -u
   ```
4. Identify incomplete tasks
5. Generate context summary

### During Session

1. Record important decisions in `decisions.md`
2. Add new patterns to `patterns.md`
3. Record file generation in `session-log.md`

### At Session End

1. Generate session summary
2. Update `context.json`
3. Record handoff items for next session

---

## Notes

- **Auto-save**: Use `hooks/Stop` to auto-append summaries to `session-log.md` at session end
- **Privacy**: Do not record confidential information
- **Git policy**: `decisions.md`/`patterns.md` for sharing; `session-log.md`/`context.json`/`.claude/state/` keep local
- **Capacity**: When logs grow large, "clean up session log" to split by month
