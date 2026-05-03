# Ralph Smoke Test: Happy Path SUCCESS

**Task**: 89.8
**Status**: Code-review-grade logic verification (no live orchestrator run — nested agents are disallowed in worker context)
**Reference implementation**: `harness/skills/harness-ralph-loop/references/loop-flow.md`
**Terminal state**: SUCCESS (all four terminal states now covered by 89.8–89.11)

---

## Test Scenario

**Input conditions**

- A `[ralph]` task is registered in Plans.md with:
  - Description: "Create the marker file `/tmp/ralph-marker.txt`"
  - DoD: "/tmp/ralph-marker.txt exists"
  - Verify: `test -f /tmp/ralph-marker.txt`
  - MaxIter: 10 (default)
- Worker iteration 0 is spawned into a persistent worktree.
- The worker runs `touch /tmp/ralph-marker.txt`, making **one file change** (or equivalent side effect detectable by verify).
- The verify command (`test -f /tmp/ralph-marker.txt`) exits **0** (file present).
- The worker emits `<promise>/tmp/ralph-marker.txt exists</promise>` in its final message.
- `ralph-worker-report.v1` is emitted with `verify.exit_code: 0` and `promise.asserted: true`.

**Simulated stub behaviour**

```
ralph-worker (iter 0):
  - Runs: touch /tmp/ralph-marker.txt
  - Commits the file change in the worktree
  - Emits in final message:
      <promise>/tmp/ralph-marker.txt exists</promise>
  - Emits ralph-worker-report.v1:
      {
        "schema": "ralph-worker-report.v1",
        "task_id": "89.8",
        "iteration": 0,
        "verify": { "exit_code": 0, "command": "test -f /tmp/ralph-marker.txt", "stderr_tail": "" },
        "promise": { "asserted": true, "dod": "/tmp/ralph-marker.txt exists" },
        "files_changed": [],
        "summary": "Created /tmp/ralph-marker.txt as required."
      }
```

---

## Expected Sequence

```
orchestrator starts
│
├─ Serialization guard: no other [ralph] task is cc:WIP → proceed
│
├─ Mark task cc:WIP in Plans.md
│
├─ iter = 0
│   ├─ Agent(subagent_type="ralph-worker", isolation="worktree")
│   │    └─ worker spawns
│   │    └─ runs: touch /tmp/ralph-marker.txt
│   │    └─ commits the change in the worktree
│   │    └─ runs verify: test -f /tmp/ralph-marker.txt → exits 0
│   │    └─ emits: <promise>/tmp/ralph-marker.txt exists</promise>
│   │    └─ emits ralph-worker-report.v1 with verify.exit_code: 0, promise.asserted: true
│   │
│   ├─ RALPH_WORKTREE = result.worktreePath  (set after iter 0 returns)
│   │
│   ├─ authoritative verify runs → exit code 0
│   ├─ prior_verify_result = {exit_code: 0, stderr_tail: ""}
│   │
│   ├─ parse ralph-worker-report.v1 from result.finalMessage
│   │    └─ ralph_report.verify.exit_code = 0
│   │    └─ ralph_report.promise.asserted = true
│   │
│   ├─ files_changed = git_diff_stat(RALPH_WORKTREE, since=PRIOR_COMMIT)
│   │    └─ returns [<worktree commit hash>]  ← non-empty (worker made changes)
│   ├─ PRIOR_COMMIT updated
│   │
│   └─ Decision matrix (priority order):
│       ├─ FT-RALPH-02 check: ralph_report.verify.exit_code (0) == verify_exit (0) → NOT triggered
│       ├─ FT-RALPH-01 check: files_changed.empty (false) → NOT triggered
│       ├─ SUCCESS check: promise_tag_found (true) AND promise.asserted (true) AND verify_exit == 0 (true)
│       │    ├─ TRIGGERED
│       │    ├─ commit_hash = merge_worktree_to_main(RALPH_WORKTREE)
│       │    ├─ update_plans_md(task_id, status="cc:done [<commit_hash>]")
│       │    ├─ print SUCCESS summary: iterations_used=1, files_changed_total, final_verify_output
│       │    ├─ cleanup_worktree(RALPH_WORKTREE)    ← worktree REMOVED on success
│       │    └─ exit 0
│       │
│       └─ [CONTINUE branch is NEVER reached]
│
└─ orchestrator exits 0 after iter 0
   iter 1 is NEVER started
```

**Expected Plans.md final state**

```
cc:done [<commit_hash>]
```

**Expected side effects**

- `/tmp/ralph-marker.txt` exists on disk (created by the worker in iter 0)
- Commit `<commit_hash>` is present in main branch (merged from worktree)
- Worktree at `RALPH_WORKTREE` has been removed (`cleanup_worktree()` was called)

---

## Pseudocode Walkthrough — Decision Matrix Path

The following traces through the loop-flow.md pseudocode step by step.

### Step 1–4: Setup and worker dispatch (iter 0)

```pseudocode
iter = 0
PRIOR_COMMIT = null    # no prior commit yet on iter 0
MaxIter = 10           # default

# Step 1: Build prompt
prompt = build_ralph_prompt(task, iter=0, MaxIter=10, RALPH_WORKTREE=null, prior_verify_result=null)
# [prior iteration context] block is OMITTED because iter == 0

# Step 2: Dispatch worker
result = Agent(subagent_type="claude-code-harness:ralph-worker", isolation="worktree")
RALPH_WORKTREE = result.worktreePath   # captured immediately after iter 0

# Step 3: Authoritative verify
verify_exit = 0    # test -f /tmp/ralph-marker.txt → file present → exit 0
prior_verify_result = {exit_code: 0, stderr_tail: ""}

# Step 4: Parse worker report
ralph_report.verify.exit_code = 0
ralph_report.promise.asserted = true
```

### Step 5: Stuck detection (lines 71–74)

```pseudocode
files_changed = git_diff_stat(RALPH_WORKTREE, since=PRIOR_COMMIT)
# PRIOR_COMMIT was null → diff against initial worktree HEAD
# Worker committed touch /tmp/ralph-marker.txt → files_changed = ["/tmp/ralph-marker.txt"]  (non-empty)

PRIOR_COMMIT = result.commit    # update for next iteration (though next iteration won't occur)
```

### Step 6: Decision matrix — evaluated in priority order

**Branch 1: FT-RALPH-02 check (line 80)**

```pseudocode
if ralph_report.verify.exit_code != verify_exit:   # 0 != 0 → FALSE
    ...
# NOT triggered — worker self-report matches authoritative verify
```

**Branch 2: FT-RALPH-01 check (line 89)**

```pseudocode
if files_changed.empty AND verify_exit != 0:       # false AND false → FALSE
    ...
# NOT triggered — files changed AND verify passed
```

**Branch 3: SUCCESS check (line 98)**

```pseudocode
if promise_tag_found(result.finalMessage)          # TRUE: <promise>/tmp/ralph-marker.txt exists</promise> found
   AND ralph_report.promise.asserted               # TRUE: promise.asserted = true
   AND verify_exit == 0:                           # TRUE: authoritative verify returned 0
    # ALL THREE conditions satisfied → SUCCESS BRANCH TAKEN
    commit_hash = merge_worktree_to_main(RALPH_WORKTREE)
    update_plans_md(task_id, status="cc:done [<commit_hash>]")
    print SUCCESS summary: iterations_used=1, files_changed_total=1, final_verify_output=""
    cleanup_worktree(RALPH_WORKTREE)    # ← worktree REMOVED (unlike failure modes)
    exit 0
```

**This branch is taken. Execution terminates with exit 0.**

**Branch 4: CONTINUE (line 105–107)**

```pseudocode
# NEVER REACHED — exit 0 was called above
```

**FT-RALPH-03 MAX-ITER (line 109–115)**

```pseudocode
# NEVER REACHED — loop body exited at line 103 before the loop could exhaust MaxIter
```

---

## Critical Ordering: SUCCESS Before STUCK — Why It Matters

The task description flagged this concern explicitly:

> SUCCESS check must come BEFORE the failure mode checks (otherwise a passing worker could be falsely flagged as STUCK if no files changed in the last iter but verify still passes)

However, **loop-flow.md does NOT implement this ordering**. The actual priority in the pseudocode is:

```
1. FT-RALPH-02 VERIFY-MISMATCH  (line 80)
2. FT-RALPH-01 STUCK             (line 89)
3. SUCCESS                        (line 98)
4. CONTINUE                       (line 105)
```

The concern about false STUCK triggering would apply if a worker in a **later iteration** had already created the file in a prior iteration (so `files_changed` is empty for the current iter), AND verify still passes. Let's trace that edge case:

**Edge case: iter 1, marker already exists from iter 0**

```pseudocode
# Hypothetical iter 1 (file already existed from iter 0, no new changes needed)
files_changed = git_diff_stat(RALPH_WORKTREE, since=PRIOR_COMMIT)
# PRIOR_COMMIT = iter-0 commit
# Worker made NO new changes in iter 1
# → files_changed = []   (empty)

verify_exit = 0    # file still exists → exits 0

# FT-RALPH-01 check (line 89):
if files_changed.empty AND verify_exit != 0:   # true AND FALSE → FALSE
    ...
# NOT triggered — because verify_exit == 0, the AND condition is false
# SUCCESS check (line 98):
if promise_tag_found AND ralph_report.promise.asserted AND verify_exit == 0:
    # TRIGGERED — success on iter 1
```

**Conclusion**: The FT-RALPH-01 guard condition `verify_exit != 0` acts as a natural firewall. A passing verify (`verify_exit == 0`) makes the FT-RALPH-01 branch unreachable regardless of `files_changed.empty`. The SUCCESS branch at line 98 is therefore always reachable when verify passes, even though it comes after the STUCK check.

The ordering concern in the task description is addressed by the guard condition itself, not by moving the SUCCESS branch. The current ordering in loop-flow.md is logically correct.

---

## Verification: Loop-flow.md Lines Implementing the SUCCESS Branch

| Line(s) | Content |
|---------|---------|
| **97** | `# ── SUCCESS ───` comment marker |
| **98** | Guard: `if promise_tag_found(result.finalMessage) AND ralph_report.promise.asserted AND verify_exit == 0:` |
| **99** | `commit_hash = merge_worktree_to_main(RALPH_WORKTREE)` — squash-merge branch into main |
| **100** | `update_plans_md(task_id, status="cc:done [{hash}]".format(hash=commit_hash))` — writes cc:done marker |
| **101** | `print SUCCESS summary: iterations_used, files_changed_total, final_verify_output` — user-visible output |
| **102** | `cleanup_worktree(RALPH_WORKTREE)` — removes worktree (SUCCESS only; failure modes preserve it) |
| **103** | `exit 0` — clean exit |

The three conjuncts of the SUCCESS condition implement three independent checks:

1. **`promise_tag_found`** — scans only `result.finalMessage` (line 182–183 of loop-flow.md) via `PROMISE_REGEX`
2. **`ralph_report.promise.asserted`** — reads the structured `ralph-worker-report.v1` field (independent of tag scanning)
3. **`verify_exit == 0`** — the authoritative verify run by the orchestrator (line 64), not trusted from the worker

All three must be true simultaneously. A worker that emits a promise tag but whose verify fails, or that reports `promise.asserted: true` in its JSON but fails to emit the tag, will not trigger SUCCESS.

---

## Terminal State Handling: SUCCESS vs. Failure Modes

| Action | SUCCESS | FT-RALPH-01 | FT-RALPH-02 | FT-RALPH-03 |
|--------|---------|-------------|-------------|-------------|
| Plans.md marker | `cc:done [hash]` | `blocked (ralph stuck ...)` | `blocked (worker self-report disagrees ...)` | `blocked (ralph max-iter exhausted ...)` |
| Worktree | **Removed** (`cleanup_worktree`) | Preserved | Preserved | Preserved |
| Exit code | 0 | 1 | 1 | 1 |
| Line in loop-flow.md | 100–103 | 92–95 | 83–86 | 110–115 |

---

## Logic Correctness Assessment

### Trigger coverage: PASS

The triple conjunction `promise_tag_found AND ralph_report.promise.asserted AND verify_exit == 0`
correctly implements the SUCCESS terminal state:

- `promise_tag_found` — prevents false positives from a worker that forgets the promise tag
- `ralph_report.promise.asserted` — prevents a worker from hallucinating success via tag alone
- `verify_exit == 0` — authoritative verify (not worker-trusted); this is the load-bearing hard guarantee

### Action: PASS — merge, mark, clean, exit 0

`merge_worktree_to_main` on line 99 commits the worktree state to main.
`update_plans_md` on line 100 writes `cc:done [<hash>]`.
`cleanup_worktree` on line 102 removes the worktree (distinguishing SUCCESS from the preserved-on-failure modes).
`exit 0` on line 103 signals orchestrator completion.

### Plans.md update: PASS

`update_plans_md(task_id, status="cc:done [{hash}]")` on line 100 writes the exact format.
The `{hash}` placeholder is the commit hash returned by `merge_worktree_to_main`.

### Critical ordering (SUCCESS after STUCK): PASS WITH ANALYSIS

SUCCESS check at line 98 comes **after** FT-RALPH-01 at line 89. The task description flagged this as a potential false-STUCK risk. Analysis (above) shows this is safe because FT-RALPH-01's guard `verify_exit != 0` is mutually exclusive with SUCCESS's `verify_exit == 0`. No scenario where verify passes can trigger FT-RALPH-01.

### Worktree cleanup: PASS

`cleanup_worktree(RALPH_WORKTREE)` is called only in the SUCCESS branch (line 102).
All three failure modes (`exit 1` branches at lines 86, 95, and 115) explicitly omit cleanup
with `# Do NOT clean up worktree — preserve for inspection` comments. This asymmetry is correct:
successful merges no longer need the worktree; failures need it for manual diagnosis.

### Promise tag vs. structured report: PASS

Two independent signals must agree:
- `promise_tag_found` scans text output (line 181–183)
- `ralph_report.promise.asserted` reads structured JSON (line 68–69)

Neither alone is sufficient. This dual-check prevents both accidental tag matches in tool-call echoes
and hallucinated structured-report fields.

---

## No Issues Found

The SUCCESS branch in `loop-flow.md` is logically correct.
No escalation required.

---

## Smoke Test Result Summary

| Check | Expected | Observed (logic) | Pass/Fail |
|-------|----------|-----------------|-----------|
| Trigger condition | `promise_tag_found AND promise.asserted AND verify_exit == 0` | Line 98 implements exact triple conjunction | PASS |
| clean exit at iter 0 | `exit 0` before loop continues | Line 103, within the iter 0 body | PASS |
| iter 1 never started | No second Agent call | Exit before CONTINUE branch | PASS |
| Plans.md message | `cc:done [<hash>]` | Lines 100 produces exact format | PASS |
| Worktree removed | `cleanup_worktree()` called | Line 102, SUCCESS branch only | PASS |
| Verify is authoritative | Worker exit code not trusted | Line 64 runs verify independently | PASS |
| Ordering safe despite SUCCESS-after-STUCK | `files_changed.empty AND verify_exit != 0` cannot trigger when `verify_exit == 0` | Mutual exclusivity of guard conditions | PASS |
| `/tmp/ralph-marker.txt` exists | File created in worktree, merged to main | Worker `touch` + `merge_worktree_to_main` | PASS |
