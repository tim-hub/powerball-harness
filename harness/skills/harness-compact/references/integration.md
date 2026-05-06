# harness-compact — Integration with Existing Compaction Infrastructure

This document explains how `harness-compact` relates to the three existing harness
compaction components. The short answer: `harness-compact` is about **timing**; the
other components are about **state preservation**.

---

## Component Map

```
harness-compact (this skill)
  └── PostToolUse hook → suggest-compact.sh
      └── counts tool calls, emits /compact suggestion at threshold

PreCompact hook (pre-compact-save.js)
  └── Fires when /compact actually runs
      └── Writes .claude/state/handoff-artifact.json
          └── Captures WIP tasks, recent edits, session role, next action

PostCompact hook (post-compact.sh)
  └── Fires after /compact completes
      └── Reads handoff-artifact.json
          └── Re-injects Plans.md WIP context as systemMessage
```

These are orthogonal pipelines. `harness-compact` does not call or modify
`pre-compact-save.js` or `post-compact.sh`. They activate independently.

---

## Relationship 1: harness-compact ↔ PreCompact

**harness-compact** decides *when* to suggest running `/compact`.

**PreCompact** (`scripts/hook-handlers/pre-compact-save.js`) fires *automatically*
when `/compact` actually runs, regardless of whether `harness-compact` suggested it.

**Example flow**:
```
[Session — 52 tool calls reached]
suggest-compact.sh → stdout: {"systemMessage": "52 tool calls — consider /compact ..."}
  ↓ user sees the nudge, decides to run /compact
/compact runs
  ↓ PreCompact hook fires
pre-compact-save.js → writes .claude/state/handoff-artifact.json
  {version, wipTasks, recentEdits, nextAction, openRisks, ...}
```

If the user ignores the suggestion and never runs `/compact`, `pre-compact-save.js`
never fires. If `/compact` runs without `harness-compact` suggesting it (e.g., CC
auto-compaction), `pre-compact-save.js` still fires because it's wired to the
`PreCompact` event directly.

---

## Relationship 2: harness-compact ↔ PostCompact

**PostCompact** (`scripts/hook-handlers/post-compact.sh`) fires *after* `/compact`
completes and re-injects the context saved by `pre-compact-save.js`.

`harness-compact` has no interaction with `PostCompact`. Once compaction is done,
`PostCompact` handles re-orientation automatically.

**Example flow**:
```
/compact completes
  ↓ PostCompact hook fires
post-compact.sh → reads handoff-artifact.json
  → emits systemMessage: "## Structured Handoff\n- Previous state: ...\n- Next action: ..."
  ↓ model context is re-oriented

[user resumes work]
harness-compact PostToolUse hook → counter picks up from where it left off
  (counter file survived compaction; it lives in .claude/state/)
```

---

## Role-Gate Alignment

`suggest-compact.sh` reads `.claude/state/session.json` and suppresses the suggestion
when `role == worker` AND Plans.md has a `cc:WIP` task. This mirrors the blocking
logic in `pre-compact-save.js` (`shouldBlockCompaction`):

```
pre-compact-save.js shouldBlockCompaction():
  worker + WIP tasks → { block: true }   ← blocks actual compaction

suggest-compact.sh suppression:
  worker + cc:WIP   → exit 0 (no message) ← suppresses the nudge upstream
```

The suppression in `suggest-compact.sh` is a courtesy guard — it avoids nudging the
user toward an action that `pre-compact-save.js` would block anyway.

---

## handoff-artifact.json Relationship

`handoff-artifact.json` is written by `pre-compact-save.js` at compaction time. It
is read by `post-compact.sh` after compaction. `harness-compact` does not read or
write `handoff-artifact.json`.

However, the suggestion message emitted by `suggest-compact.sh` explicitly mentions
the handoff artifact mechanism:

> "Compacting now preserves the handoff artifact and re-injects Plans.md context
> after resume."

This guides users to understand the full lifecycle, not just the suggestion.

---

## Summary Table

| Component | Trigger | Writes | Reads | Concern |
|-----------|---------|--------|-------|---------|
| `suggest-compact.sh` | PostToolUse(Edit\|Write) | `compact-counter-<id>.json` | `session.json`, `Plans.md` | **Timing** |
| `pre-compact-save.js` | PreCompact event | `handoff-artifact.json`, `precompact-snapshot.json` | `Plans.md`, `session.json`, `work-active.json` | **State save** |
| `post-compact.sh` | PostCompact event | `compaction-events.jsonl` | `handoff-artifact.json`, `precompact-snapshot.json`, `Plans.md` | **State restore** |
