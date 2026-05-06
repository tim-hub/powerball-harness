---
name: harness-compact
description: "Suggests strategic /compact checkpoints to preserve context across phase transitions. Use when deciding when to compact, checking session tool-call stats, or resetting the compact counter."
when_to_use: "suggest compact, check compact stats, reset compact counter, /compact timing, strategic compaction, context window"
allowed-tools: ["Read", "Bash", "Grep"]
---

# harness-compact

Strategic compaction suggester — nudges `/compact` at meaningful task boundaries so auto-compaction does not interrupt mid-implementation.

Inspired by [`affaan-m/everything-claude-code`](https://github.com/affaan-m/everything-claude-code/tree/main/skills/strategic-compact).

## Quick Reference

| Subcommand | Behavior |
|------------|----------|
| `suggest` (default) | Analyse current session state and print a compact recommendation now |
| `stats` | Show tool-call counter, threshold, next reminder at, session age, WIP count |
| `reset` | Reset the tool-call counter for this session |

## Automatic Hook

A `PostToolUse(Edit|Write)` hook counts change events per session.  
Counter file: `.claude/state/compact-counter-<CLAUDE_SESSION_ID>.json`  
Cleaned automatically by the `SessionEnd` → `session-cleanup` hook.

When the count crosses a threshold a `systemMessage` is emitted so the model sees the suggestion:

| Count | Action |
|-------|--------|
| Reaches `HARNESS_COMPACT_THRESHOLD` (default 50) | First suggestion |
| Every `HARNESS_COMPACT_INTERVAL` (default 25) thereafter | Periodic reminder |

Override thresholds via env vars: `HARNESS_COMPACT_THRESHOLD`, `HARNESS_COMPACT_INTERVAL`.

## Suppression Rules

The hook stays silent when either condition holds:

- Session role is `worker` **AND** Plans.md contains any `cc:WIP` task — mirrors the existing PreCompact role-gate; Workers must not compact mid-implementation.
- An active `[ralph]` orchestration loop is in progress.

## Decision Guide

See [`references/decision-framework.md`](${CLAUDE_SKILL_DIR}/references/decision-framework.md) — when to compact vs. when to wait, including harness-specific rules for `[ralph]` loops and phase boundaries.

## Relationship to PreCompact / PostCompact

`harness-compact` is about **timing** only — it suggests *when* to `/compact`.

The existing `PreCompact` hook (`scripts/hook-handlers/pre-compact-save.js`) and `PostCompact` hook (`scripts/hook-handlers/post-compact.sh`) handle **state preservation** — they save and restore context *when* compaction occurs. These concerns are orthogonal.

See [`references/integration.md`](${CLAUDE_SKILL_DIR}/references/integration.md) for the full integration diagram.

## Running Manually

```bash
# Analyse and print recommendation now
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-compact/scripts/suggest-compact.sh"

# Show session stats
bash "${CLAUDE_PLUGIN_ROOT}/skills/harness-compact/scripts/show-stats.sh"

# Reset counter
rm -f ".claude/state/compact-counter-${CLAUDE_SESSION_ID}.json"
```

## Related Skills

- `harness-plan` — Re-read Plans.md after compacting to orient on the next task
- `session` — Inspect current session state
- `harness-work` — Resume implementation after compacting
