# harness-compact — Compaction Decision Framework

When to compact vs. when to wait.

---

## Decision Matrix (upstream)

| Situation | Recommendation | Reason |
|-----------|---------------|--------|
| Just finished research / exploration phase | **Compact now** | Research output is best summarised; raw tool call history is noise |
| Just finished planning (Plans.md written) | **Compact now** | Plan is on disk; context from the planning dialogue adds little |
| About to start a new phase or new task | **Compact now** | Fresh context for implementation; Plans.md re-read on resume |
| Mid-implementation (cc:WIP, files partially written) | **Wait** | Variable names, file paths, design decisions are live context |
| Mid-debugging (actively tracing a failure) | **Wait** | Compacting loses the stack of observations you've accumulated |
| Just after a failed approach | **Compact now** | Failed reasoning is noise; a clean slate avoids anchoring |
| About to switch to a completely unrelated task | **Compact now** | Prior task context pollutes the new one |
| After a long idle / context feels stale | **Compact now** | Stale noise outweighs the cost of re-reading key files |

---

## Harness-Specific Rules

### Rule H-1 — Do NOT compact mid-`[ralph]` loop

A `[ralph]` orchestrator loop keeps iteration history, worktree state, and the
verify-command contract in active context. Compacting mid-loop drops the
iteration count, the `FT-RALPH-*` suppression history, and the `MaxIter`
budget, causing the orchestrator to restart from scratch.

**Action**: Wait until `harness-ralph-loop` writes `cc:done [hash]` or a
`blocked` marker to Plans.md before compacting.

### Rule H-2 — Prefer compacting between phases when the previous phase is fully `cc:done`

When every task in the current phase shows `cc:done [hash]` or `pm:confirmed`,
compacting before starting the next phase gives maximum benefit:
- Planning dialogue for the completed phase is no longer needed
- Next-phase tasks will trigger fresh file reads anyway
- `handoff-artifact.json` already captures the WIP state snapshot

**Action**: Run `/compact` after all tasks in a phase are done, before running
`/harness-work` on the next phase.

### Rule H-3 — Prefer compacting after `/harness:harness-review` completes

A review session accumulates reading across many files. Once the verdict is
written to `.claude/state/review-result.json`, the in-context file-read
snapshots are redundant. Compacting at this point frees context for the fix
cycle without losing any artifact.

**Action**: After `/harness:harness-review` returns `APPROVE` or
`REQUEST_CHANGES`, compact before beginning the next major action.

---

## Persistence Model

Understanding what survives `/compact` helps you decide when it is safe.

### Survives compaction

| What | Why |
|------|-----|
| `CLAUDE.md` + all `.claude/rules/` | Re-injected by `InstructionsLoaded` hook |
| `Plans.md` | On disk; re-read by `PostCompact` re-injection |
| `.claude/memory/decisions.md`, `patterns.md` | On disk |
| `.claude/state/handoff-artifact.json` | Written by `PreCompact` hook before compaction |
| All git-committed code | On disk; Workers re-read as needed |
| `TodoWrite` task list | Persisted by CC separately |

### Lost at compaction

| What | Implication |
|------|-------------|
| Mid-conversation reasoning chains | Re-derive from Plans.md + files after resume |
| File snapshots read during the session | Claude will re-read on next reference |
| Verbal preferences stated in conversation | Write to `.claude/memory/` or CLAUDE.md first |
| The current debug / trace stack | Only compact after the investigation concludes |
| `[ralph]` iteration history (mid-loop) | See Rule H-1 above |

---

## Quick Checklist Before Compacting

- [ ] No active `cc:WIP` tasks in Plans.md that I haven't committed or checkpointed
- [ ] No `[ralph]` loop in progress
- [ ] Any verbal decisions or preferences written to `.claude/memory/`
- [ ] Current phase is at a natural boundary (phase complete, or just before starting new phase)
