# Sync Project Specs Reference

**Run this when you're worried "Did Plans.md actually get updated?" after completing work.**

## When to Use

| Situation | Command to Use |
|-----------|----------------|
| "How far along? What's next?" | `harness-sync` (use this first) |
| "Worked on it but forgot if I updated Plans.md" | **This command** |
| "Started from old template, format might be outdated" | **This command** |

> Tip: Usually `harness-sync` is sufficient. Use this for "just in case" or "format migration".

---

## Purpose

Aligns project specs' docs (e.g., `Plans.md`, `AGENTS.md`, `.claude/rules/*`) with latest powerball-harness operations (**PM <-> Impl**, `pm:*` markers, handoff commands).

## VibeCoder Phrases

- "**Worked on it but unsure if Plans.md is updated**" -> this command
- "**Want to align old format files to latest**" -> Unifies markers and descriptions
- "**Keep manual changes, fix only needed parts**" -> Preserves existing text, applies only diffs

---

## Sync Targets (Only Existing Files)

- `Plans.md`
- `AGENTS.md`
- `CLAUDE.md` (only if has operation description)
- `.claude/rules/workflow.md`
- `.claude/rules/plans-management.md`

---

## Sync Content (Minimal Diff Policy)

### 1. Marker Normalization

- **Standard**: `pm:requested`, `pm:confirmed`
- **Compatible**: `cursor:requested`, `cursor:confirmed` (treated as synonyms)

### 2. State Transition Documentation

```
pm:requested -> cc:WIP -> cc:done -> pm:confirmed
```

### 3. Handoff Routes Addition

- PM->Impl: use `harness-work` to implement tasks
- Impl->PM: use `harness-review` for review handoff
- Cursor workflow: use `cc-cursor-cc` skill for Claude↔Cursor handoffs

### 4. Notification File Description

- `.claude/state/pm-notification.md` (compatible: `.claude/state/cursor-notification.md`)

---

## Execution Steps

### Step 1: Collect Current State (Required)

- Check target file existence and extract relevant sections
- Tally `Plans.md` marker occurrences (pm/cursor/cc)

### Step 2: Declare Change Policy (Required)

Tell user:
- Preserve existing text in principle (no destructive rewrites)
- Additions/replacements limited to "minimum necessary for operation"
- Changes shown as diffs, adjust if needed

### Step 3: Sync (Apply Diffs)

- **Plans.md**: Add `pm:*` to marker legend, note `cursor:*` as compatible
- **AGENTS.md**: Update roles to PM/Impl
- **rules/*.md**: Change `cursor:*` to `pm:*` standard + compatibility note
- **CLAUDE.md**: Add PM<->Impl routes if operation section exists

### Step 4: Finish (Required)

- Run `harness-sync` to verify markers
- If the Harness MCP server is connected, call `harness_mem_record_checkpoint` to mark the sync point — captures enough state (which specs were aligned, current marker counts, any manual tweaks) for other agents or future sessions to resume from a known-good baseline. If MCP is unavailable, skip silently and mention in the report.

---

## Parallel Execution

File reads can be parallelized:

| Process | Parallel |
|---------|----------|
| Plans.md read | Yes, Independent |
| AGENTS.md read | Yes, Independent |
| CLAUDE.md read | Yes, Independent |
| rules/*.md read | Yes, Independent |

Updates run serially for consistency.
