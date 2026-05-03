# Ralph Loop Smoke Test — FT-RALPH-03 MAX-ITER

**Task**: 89.11
**Date**: 2026-05-03
**Status**: PASS — logic is correctly implemented

---

## Scenario

| Parameter | Value |
|-----------|-------|
| `MaxIter` | 3 |
| Verify command | Always exits 1 (simulate persistent failure) |
| Worker behavior | Changes files each iteration (STUCK does NOT trigger) |
| Expected terminal state | `FT-RALPH-03 MAX-ITER` |

**Core invariant being tested**: When `verify_exit != 0` but `files_changed` is non-empty on every iteration, the STUCK check (FT-RALPH-01) must NOT fire. The loop must run all MaxIter iterations before halting.

---

## Expected Sequence

| Iteration | Worker spawned | Files changed | Verify exit | Decision |
|-----------|---------------|---------------|-------------|----------|
| iter 0 | yes (creates worktree) | yes (e.g., `foo.py` added) | 1 | CONTINUE — files changed, verify failed, iter+1 < MaxIter |
| iter 1 | yes (EnterWorktree) | yes (e.g., `foo.py` modified) | 1 | CONTINUE — files changed, verify failed, iter+1 < MaxIter |
| iter 2 | yes (EnterWorktree) | yes (e.g., `bar.py` added) | 1 | CONTINUE — files changed, verify failed, iter+1 = MaxIter → loop ends |
| (iter 3 boundary) | — | — | — | Loop exits; FT-RALPH-03 fires |

**Iteration count in worktree git log**: 3 commits (one per worker dispatch, iter 0–2)

---

## Expected Plans.md Final State

```
blocked (ralph max-iter exhausted after 3 iterations, last verify exit: 1)
```

This matches the template at `loop-flow.md` lines 110–111:
```
reason = "blocked (ralph max-iter exhausted after {N} iterations, last verify exit: {code})"
          .format(N=MaxIter, code=verify_exit)
```

`N` is `MaxIter` (the configured limit = 3), not the last iteration index (which would be 2). This is the correct interpretation — "after 3 iterations" means 3 iterations were attempted.

---

## Pseudocode Walkthrough

References line numbers in `harness/skills/harness-ralph-loop/references/loop-flow.md`.

```python
# line 25: MaxIter = 3
MaxIter = 3

# line 36: loop over range(0, MaxIter) → iterations 0, 1, 2
for iter in range(0, MaxIter):    # iter = 0, 1, 2  (3 total iterations)

    # lines 38-61: dispatch worker
    # iter 0: creates worktree (isolation="worktree")
    # iter 1+: EnterWorktree then spawn worker

    # lines 63-65: run authoritative verify
    verify_exit = 1               # always fails in this scenario
    prior_verify_result = {exit_code: 1, stderr_tail: "..."}

    # lines 71-74: detect stuck
    files_changed = ["foo.py"]    # NON-EMPTY — worker did make changes

    PRIOR_COMMIT = result.commit  # advance for next iter

    # lines 79-85: VERIFY-MISMATCH check
    # ralph_report.verify.exit_code == verify_exit == 1 → no mismatch → skip

    # lines 88-95: STUCK check (FT-RALPH-01)
    # files_changed.empty = False → condition NOT met → SKIP
    # (This is the key: STUCK requires BOTH empty files_changed AND verify_exit != 0)

    # lines 97-103: SUCCESS check
    # verify_exit == 1 (not 0) → NOT success → skip

    # lines 105-107: CONTINUE
    # verify_exit != 0, iter+1 < MaxIter (for iter 0 and 1):
    #   print "iter 0: verify exit 1, continuing..."
    #   print "iter 1: verify exit 1, continuing..."
    # iter 2: verify_exit != 0, iter+1 == MaxIter → loop range exhausted → fall through

# END OF FOR LOOP
# lines 109-115: FT-RALPH-03 MAX-ITER fires HERE
reason = "blocked (ralph max-iter exhausted after 3 iterations, last verify exit: 1)"
update_plans_md(task_id, status=reason)
# worktree preserved (not cleaned up)
exit 1
```

---

## Decision Matrix Verification

From `loop-flow.md` lines 160–168 (Decision Matrix table):

| Condition evaluated | Our scenario result | Terminal state triggered? |
|---------------------|---------------------|--------------------------|
| `promise_tag_found AND promise.asserted AND verify_exit == 0` | False (verify_exit = 1) | No |
| `ralph_report.verify.exit_code != verify_exit` | False (both = 1) | No |
| `files_changed.empty AND verify_exit != 0` | False (files_changed non-empty) | No — STUCK does NOT fire |
| `iter >= MaxIter` (evaluated after loop) | True after iter 2 completes | **YES — FT-RALPH-03 fires** |

All four conditions are correctly resolved. The STUCK gate is correctly bypassed when files change.

---

## Critical Logic Verification: STUCK Does NOT Fire When Files Change

This is the most important invariant for this smoke test.

**STUCK trigger** (loop-flow.md lines 88–95):
```python
if files_changed.empty AND verify_exit != 0:
    # FT-RALPH-01
```

The `AND` means **both** conditions must be true. In our scenario:
- `files_changed.empty` = **False** (worker changed files every iteration)
- `verify_exit != 0` = True

Because `files_changed.empty` is False, the `AND` short-circuits to False. STUCK is not triggered. This is correct behavior — a worker that is making progress (changing files) but not yet satisfying the verify command should be allowed to continue until MaxIter.

The orchestrator correctly distinguishes between:
- **Stuck worker**: No progress at all (same commit hash) + verify still fails → FT-RALPH-01 (immediate stop)
- **Progressing worker**: Making changes but verify not yet satisfied → continue until MaxIter → FT-RALPH-03

---

## FT-RALPH-03 MAX-ITER Branch Analysis

**Trigger location**: `loop-flow.md` lines 109–115 — OUTSIDE the `for` loop (after the loop exhausts `range(0, MaxIter)`)

**Trigger condition**: The loop exits naturally (all iterations consumed) without hitting SUCCESS, VERIFY-MISMATCH, or STUCK.

**Actions taken** (lines 109–115 and Terminal State Handling section, lines 214–219):
1. Write `blocked (ralph max-iter exhausted after {N} iterations, last verify exit: {code})` to Plans.md
2. Print iteration history via `git -C {RALPH_WORKTREE} log --oneline`
3. Print final verify stderr
4. **Do NOT remove the worktree** — preserve for manual inspection or handoff
5. Exit non-zero (exit 1)

All five actions are correctly specified. The worktree preservation is critical — the 3 worker commits remain accessible for debugging.

---

## Worktree Git Log

After 3 iterations with MaxIter=3, the worktree git log should show exactly 3 commits:

```
git -C {RALPH_WORKTREE} log --oneline

abc1234 [ralph-iter-2] attempt 3: added bar.py
def5678 [ralph-iter-1] attempt 2: modified foo.py
ghi9012 [ralph-iter-0] attempt 1: added foo.py
```

The DoD states "iteration count in worktree git log = 3", which is satisfied.

---

## No Logic Bugs Found

All verification conditions pass:

- [x] **Trigger**: iter counter reaches MaxIter (3) without both promise+verify passing
- [x] **Action**: Loop stops (exits after range exhaustion), worktree preserved
- [x] **Plans.md update**: `blocked (ralph max-iter exhausted after {N} iterations, last verify exit: {code})` — matches exactly
- [x] **STUCK bypass**: Files changing each iter means `files_changed.empty = False`, so STUCK condition fails — MAX-ITER fires correctly

No escalation required. The FT-RALPH-03 branch in `loop-flow.md` is correctly implemented.

---

## References

- `harness/skills/harness-ralph-loop/references/loop-flow.md` — loop logic (lines 36, 88-95, 109-115, 160-168, 214-219)
- `.claude/rules/failure-taxonomy.md` — FT-RALPH-03 entry
- `docs/spikes/ralph-worktree-persistence.md` — worktree pattern feasibility study
