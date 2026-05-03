# Smoke Test: FT-RALPH-02 VERIFY-MISMATCH

**Test Type**: CODE-REVIEW-GRADE (static logic verification; no live subagent spawned)
**Task**: 89.10
**Date**: 2026-05-03

---

## Overview

This smoke test verifies that the `harness-ralph-loop` orchestrator correctly handles
**FT-RALPH-02 VERIFY-MISMATCH**: the case where a worker emits a `ralph-worker-report.v1`
with `verify.exit_code: 0` (claiming verify passed) while the orchestrator's own
authoritative `bash $VERIFY_CMD` exits 1.

This is an anti-tampering hard stop — the worker is hallucinating its verify result.
The orchestrator must halt immediately and record a diagnostic `blocked` reason in Plans.md.

---

## Test Scenario

### Setup

- Task in Plans.md with `Verify: bash /tmp/ralph-test-verify.sh`
- The actual verify script (`/tmp/ralph-test-verify.sh`) exits 1 (verify fails)
- The ralph-worker is stubbed to emit a `ralph-worker-report.v1` claiming `verify.exit_code: 0`

### Stubbed Worker Report (emitted by simulated worker)

```json
{
  "schema": "ralph-worker-report.v1",
  "task_id": "89.10-smoke",
  "iteration": 0,
  "verify": {
    "exit_code": 0,
    "command": "bash /tmp/ralph-test-verify.sh",
    "stderr_tail": ""
  },
  "promise": {
    "asserted": true,
    "dod": "verify exits 0"
  },
  "files_changed": ["src/output.txt"],
  "summary": "Created output file, verify passed (0)."
}
```

Worker also emits in its final message:
```
<promise>verify exits 0</promise>
```

### Actual Orchestrator Verify Result

```
$ bash /tmp/ralph-test-verify.sh
exit code: 1
stderr: "ERROR: expected output not found"
```

---

## Expected Execution Sequence

### iter 0

1. Orchestrator spawns ralph-worker with `isolation="worktree"`.
2. Worker runs, makes file changes, and emits `ralph-worker-report.v1` claiming `verify.exit_code: 0` plus `<promise>verify exits 0</promise>`.
3. **Orchestrator runs its own authoritative verify** (`bash $VERIFY_CMD` in `cwd=RALPH_WORKTREE`) → exits **1**.
4. `prior_verify_result` is set to `{exit_code: 1, stderr_tail: "ERROR: expected output not found"}`.
5. Orchestrator parses `ralph_report` from worker's final message → `ralph_report.verify.exit_code = 0`.
6. **Decision Matrix — Step 6, first check**:
   - `ralph_report.verify.exit_code (0) != verify_exit (1)` → **VERIFY-MISMATCH detected**.
7. Orchestrator writes to Plans.md:
   ```
   blocked (worker self-report disagrees with authoritative verify: worker=0, actual=1)
   ```
8. Orchestrator prints FT-RALPH-02 summary with both verify outputs side-by-side and the worker's `ralph-worker-report.v1` JSON.
9. Orchestrator exits non-zero.
10. **Loop does NOT continue to iter 1**.

---

## Expected Plans.md Final State

```
| 89.10 | Smoke test `FT-RALPH-02 VERIFY-MISMATCH`... | ... | 89.5 | blocked (worker self-report disagrees with authoritative verify: worker=0, actual=1) |
```

Key: the status column contains `blocked (worker self-report disagrees with authoritative verify: worker=0, actual=1)`.

---

## Pseudocode Walkthrough — Decision Matrix Path

From `loop-flow.md` lines 76–86 (Step 6 — Decision Matrix):

```
# ── 6. Decision matrix ───────────────────────────────────────────────────────
# Evaluate in this priority order:

# ── STOP: VERIFY-MISMATCH (FT-RALPH-02) ──────────────────────────────────────
if ralph_report.verify.exit_code != verify_exit:           # 0 != 1  → TRUE
    reason = "blocked (worker self-report disagrees with authoritative verify: worker={W} actual={A})"
              .format(W=ralph_report.verify.exit_code,      # W=0
                      A=verify_exit)                        # A=1
    update_plans_md(task_id, status=reason)                # writes blocked(...) to Plans.md
    print FT-RALPH-02 summary with diff history and last verify stderr
    # Do NOT clean up worktree — preserve for inspection
    exit 1                                                  # orchestrator halts; no further loop
```

The `SUCCESS` check at line 98, the `FT-RALPH-01 STUCK` check at line 89,
and the `CONTINUE` path at line 106 are **never reached** because the
`FT-RALPH-02` branch returns `exit 1` first.

---

## Verification: Line Numbers in loop-flow.md

| Logic element | Location in loop-flow.md |
|---------------|--------------------------|
| FT-RALPH-02 trigger condition (`ralph_report.verify.exit_code != verify_exit`) | Line 80 |
| `reason` string construction with `worker={W} actual={A}` | Line 81 |
| `update_plans_md(task_id, status=reason)` call | Line 83 |
| `print FT-RALPH-02 summary` | Line 84 |
| `exit 1` (hard stop — loop does NOT continue) | Line 86 |
| Decision Matrix table row for FT-RALPH-02 | Line 165 |
| Terminal State Handling — FT-RALPH-02 section (5-step procedure) | Lines 207–213 |

The FT-RALPH-02 branch is **evaluated first** in the decision matrix priority order
(line 79 comment: `# Evaluate in this priority order:`), before STUCK and SUCCESS checks.
This ensures a hallucinating worker cannot slip through by also satisfying the SUCCESS
conditions (promise tag present, verify.exit_code==0 in report, but actual verify==1).

---

## Wording Note: loop-flow.md vs. failure-taxonomy.md

Minor textual difference between the two authoritative sources:

| Source | Wording |
|--------|---------|
| `loop-flow.md` line 81 (pseudocode) | `worker={W} actual={A}` |
| `loop-flow.md` line 165 (Decision Matrix table) | `worker={W} actual={A}` |
| `loop-flow.md` lines 208 (Terminal State Handling) | `worker={W} actual={A}` |
| `failure-taxonomy.md` FT-RALPH-02 escalation column | `worker claimed {W}, actual {A}` |

The Plans.md update message uses `loop-flow.md`'s format (the orchestrator pseudocode is the
implementation reference). The taxonomy's "worker claimed {W}" phrasing is a prose description
in the escalation column and does not define the exact string written to Plans.md.

The DoD in the sprint contract (`89.10.sprint-contract.json`) tests for the substring
`blocked (worker self-report disagrees with authoritative verify)` which is present in
both formulations — so the DoD is satisfied regardless of the minor suffix difference.

---

## Conclusion

**FT-RALPH-02 VERIFY-MISMATCH is correctly handled in `loop-flow.md`.**

All required conditions from the task specification are implemented:

| Requirement | Status | Evidence |
|-------------|--------|---------|
| Trigger: orchestrator `bash $VERIFY_CMD` exit code differs from `ralph_report.verify.exit_code` | PASS | loop-flow.md line 80 |
| Action: hard stop immediately (anti-tampering) | PASS | loop-flow.md line 86 (`exit 1`); loop body never reaches `continue` |
| Plans.md update: `blocked (worker self-report disagrees with authoritative verify: worker={W}, actual={A})` | PASS | loop-flow.md lines 81–83 |
| Orchestrator does NOT continue the loop after mismatch detection | PASS | `exit 1` at line 86 terminates before the `CONTINUE` path at line 107 |
| Worktree preserved for inspection (not cleaned up) | PASS | loop-flow.md line 85 comment + Terminal State Handling line 212 |

No logic bugs or missing branches were found. No escalation required.
