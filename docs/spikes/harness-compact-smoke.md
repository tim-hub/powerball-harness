# harness-compact — Smoke Test Transcript

Phase 92, task 92.7 — 2026-05-06

Two scenarios executed against
`harness/skills/harness-compact/scripts/suggest-compact.sh`.

---

## Scenario A — Threshold Trigger

**Setup**: Fresh counter file (no prior session state).

**Method**: Run `suggest-compact.sh` 55 times via loop using
`CLAUDE_SESSION_ID=smoke`. Capture stdout+stderr on runs 50 and 51.

**Run 50 — stdout**:
```json
{"systemMessage": "50 tool calls this session — consider `/compact` if transitioning phases or completing a milestone. Compacting now preserves the handoff artifact and re-injects Plans.md context after resume. Decision guide: harness/skills/harness-compact/references/decision-framework.md"}
```

**Run 50 — stderr**:
```
[HarnessCompact] 50 tool calls — strategic compact checkpoint
```

**Run 51 — stdout**: *(empty — no suggestion between thresholds)*

**Result**: PASS — `systemMessage` emitted on run 50 (contains `consider /compact`);
run 51 silent as expected. Interval reminder verified separately (run 75 in prior
Wave 2 DoD check also emitted a `systemMessage`).

---

## Scenario B — Worker+WIP Suppression

**Setup**:
- `.claude/state/session.json` set to `{"role":"worker","state":"active"}`
- Plans.md contains `cc:WIP` rows (live Plans.md during Phase 92 implementation)
- Counter seeded to 60 (well past threshold 50)

**Method**: Run `suggest-compact.sh` once more as `CLAUDE_SESSION_ID=suppress`.

**stdout**: *(empty)*

**stderr**:
```
[HarnessCompact] suppressed (worker session with cc:WIP tasks)
```

**Result**: PASS — no `systemMessage` emitted; suppression line present on stderr.
The role-gate mirrors `shouldBlockCompaction()` in `pre-compact-save.js` which
blocks actual compaction for the same condition.

---

## Summary

| Scenario | Expected | Actual | Pass? |
|----------|----------|--------|-------|
| Run 50: threshold trigger | stdout contains `consider /compact` | ✓ | ✅ |
| Run 51: silent between reminders | stdout empty | ✓ | ✅ |
| Worker+WIP: suppression | stdout empty, stderr has `suppressed` | ✓ | ✅ |
