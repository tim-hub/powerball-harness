# Ralph Smoke Test: FT-RALPH-01 STUCK

**Task**: 89.9
**Status**: Code-review-grade logic verification (no live orchestrator run — nested agents are disallowed in worker context)
**Reference implementation**: `harness/skills/harness-ralph-loop/references/loop-flow.md`
**FT entry**: `harness/rules/failure-taxonomy.md` § FT-RALPH-01

---

## Test Scenario

**Input conditions**

- Worker iteration 0 is spawned into a persistent worktree.
- The worker makes **no file changes** (git diff against `PRIOR_COMMIT` returns empty).
- The verify command (`verify_cmd`) exits **non-zero** (exit code 1).
- `iter = 0`, `MaxIter = 10` (default).

**Simulated stub behaviour**

```
ralph-worker (iter 0):
  - Writes nothing to disk (no edits, no new files)
  - Returns ralph-worker-report.v1 with verify.exit_code: 1
  - Does NOT emit <promise>...</promise>
```

---

## Expected Sequence

```
orchestrator starts
│
├─ Mark task cc:WIP in Plans.md
│
├─ iter = 0
│   ├─ Agent(subagent_type="ralph-worker", isolation="worktree")
│   │    └─ worker spawns, writes nothing, verify exits 1
│   │
│   ├─ authoritative verify runs → exit code 1
│   ├─ parse ralph-worker-report.v1 from result.finalMessage
│   ├─ files_changed = git_diff_stat(RALPH_WORKTREE, since=PRIOR_COMMIT)
│   │    └─ returns []  ← EMPTY (no files changed)
│   ├─ PRIOR_COMMIT updated
│   │
│   └─ Decision matrix (priority order):
│       ├─ FT-RALPH-02 check: ralph_report.verify.exit_code (1) == verify_exit (1) → NOT triggered
│       ├─ FT-RALPH-01 check: files_changed.empty (true) AND verify_exit != 0 (true) → TRIGGERED
│       │    ├─ reason = "blocked (ralph stuck at iter 0: no file changes + verify exit 1)"
│       │    ├─ update_plans_md(task_id, status=reason)
│       │    ├─ print FT-RALPH-01 summary
│       │    └─ exit 1   ← HARD STOP
│       │
│       └─ [SUCCESS and CONTINUE branches are NEVER reached]
│
└─ orchestrator exits non-zero after iter 0
   iter 1 is NEVER started
```

**Expected Plans.md final state**

```
blocked (ralph stuck at iter 0: no file changes + verify exit 1)
```

---

## Pseudocode Walkthrough — Decision Matrix Path

The following traces through the loop-flow.md pseudocode step by step.

### Step 1–4: Setup and worker dispatch (iter 0)

```pseudocode
iter = 0
PRIOR_COMMIT = null    # no prior commit yet on iter 0

# Dispatch
result = Agent(subagent_type="ralph-worker", isolation="worktree")
RALPH_WORKTREE = result.worktreePath

# Authoritative verify (step 3)
verify_exit = 1    # command exited non-zero
prior_verify_result = {exit_code: 1, stderr_tail: "..."}

# Parse worker report (step 4)
ralph_report.verify.exit_code = 1    # worker reported same exit code (consistent)
ralph_report.promise.asserted = false
```

### Step 5: Stuck detection

```pseudocode
files_changed = git_diff_stat(RALPH_WORKTREE, since=PRIOR_COMMIT)
# PRIOR_COMMIT was null → diff against initial worktree HEAD
# Worker made NO changes → files_changed = []   (empty)

PRIOR_COMMIT = result.commit    # record current HEAD for next iter
```

### Step 6: Decision matrix — evaluated in priority order

**Branch 1: FT-RALPH-02 check (line 80)**

```pseudocode
if ralph_report.verify.exit_code != verify_exit:   # 1 != 1 → FALSE
    ...
# NOT triggered — worker self-report matches authoritative verify
```

**Branch 2: FT-RALPH-01 check (line 89)**

```pseudocode
if files_changed.empty AND verify_exit != 0:       # true AND true → TRUE
    reason = "blocked (ralph stuck at iter 0: no file changes + verify exit 1)"
    update_plans_md(task_id, status=reason)
    print FT-RALPH-01 summary with diff history and last verify stderr
    # Do NOT clean up worktree — preserve for inspection
    exit 1
```

**This branch is taken. Execution stops here.**

**Branches 3–4: SUCCESS and CONTINUE (lines 98 and 106)**

```pseudocode
# NEVER REACHED — exit 1 was called above
```

---

## Verification: Loop Does NOT Proceed to iter 1

The `exit 1` statement on line 95 of loop-flow.md terminates the for loop body before the loop
can increment `iter` to 1. The `for iter in range(0, MaxIter)` construct does not continue
after a non-local exit. No second Agent call is ever made.

---

## Lines in loop-flow.md That Implement This Branch

| Line(s) | Content |
|---------|---------|
| **71–72** | Step 5 header comment + `files_changed = git_diff_stat(...)` — computes the empty diff |
| **88** | FT-RALPH-01 comment marker `# ── STOP: STUCK (FT-RALPH-01) ────...` |
| **89** | Guard condition: `if files_changed.empty AND verify_exit != 0:` |
| **90–91** | Reason string construction: `"blocked (ralph stuck at iter {N}: no file changes + verify exit {code})"` |
| **92** | `update_plans_md(task_id, status=reason)` — writes the blocked marker |
| **93** | `print FT-RALPH-01 summary with diff history and last verify stderr` — surfaces diagnosis |
| **94** | Comment: `# Do NOT clean up worktree — preserve for inspection` |
| **95** | `exit 1` — hard stop; loop does not continue to iter 1 |

---

## Logic Correctness Assessment

### Trigger coverage: PASS

The condition `files_changed.empty AND verify_exit != 0` exactly matches the FT-RALPH-01
definition in failure-taxonomy.md:

> "Idle-iteration: no file changes between iterations + verify still failing"
> Detector: "`git diff --stat` against prior-iter commit returns empty + verify exits non-zero"

Both halves of the conjunction are independently verified:
- `files_changed.empty` → git diff stat comparison (line 72)
- `verify_exit != 0` → authoritative verify (line 64), not trusted from worker

### Action: PASS — Hard stop at iter 0; iter 1 never started

`exit 1` on line 95 terminates the orchestrator immediately. The for loop body cannot
advance `iter` to 1 after a non-local exit. This satisfies the DoD requirement:
"orchestrator did NOT proceed to iter 2".

### Plans.md update: PASS

`update_plans_md(task_id, status=reason)` on line 92 writes the exact format:
`blocked (ralph stuck at iter {N}: no file changes + verify exit {code})`
where `{N}=0` and `{code}=1` in this scenario.

### Priority ordering: PASS

FT-RALPH-01 STUCK is checked **after** FT-RALPH-02 VERIFY-MISMATCH (line 80) and **before**
the SUCCESS branch (line 98). This ordering is correct:
- VERIFY-MISMATCH must be checked first (it indicates worker hallucination of verification)
- STUCK only applies when the worker report is trustworthy (exit codes agree)
- SUCCESS is unreachable when verify_exit != 0

### Worktree preservation: PASS

The comment on line 94 and the absence of `cleanup_worktree()` confirm that the worktree
is preserved for manual inspection on FT-RALPH-01, matching the terminal-state handling
table in the Decision Matrix section (line 166).

---

## No Issues Found

The FT-RALPH-01 STUCK branch in `loop-flow.md` is logically correct.
No escalation required.

---

## Smoke Test Result Summary

| Check | Expected | Observed (logic) | Pass/Fail |
|-------|----------|-----------------|-----------|
| Trigger condition | `files_changed.empty AND verify_exit != 0` | Line 89 implements exact conjunction | PASS |
| Hard stop at iter 0 | `exit 1` before loop increment | Line 95, within the iter 0 body | PASS |
| iter 1 never started | No second Agent call | Exit before CONTINUE branch | PASS |
| Plans.md message | `blocked (ralph stuck at iter 0: ...)` | Lines 90–92 produce exact format | PASS |
| Worktree preserved | `cleanup_worktree()` NOT called | Absent from FT-RALPH-01 branch | PASS |
| Verify is authoritative | Worker exit code not trusted | Line 64 runs verify independently | PASS |
