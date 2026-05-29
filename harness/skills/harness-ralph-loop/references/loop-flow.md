# Ralph Loop Orchestrator — Full Flow

Reference for `harness-ralph-loop` orchestration logic.
Implements the persistent-worktree pattern confirmed FEASIBLE in `docs/spikes/ralph-worktree-persistence.md`.

---

## State Variables

```
RALPH_WORKTREE  = null          # absolute path to the persistent worktree (set on iter 0)
PRIOR_COMMIT    = null          # commit hash from the previous iteration (for diff detection)
ENTER_WT_LOADED = false         # EnterWorktree deferred-tool schema loaded flag
iter            = 0
prior_verify_result = null      # {exit_code, stderr_tail} from previous iteration's verify run
```

---

## Main Loop (Pseudocode)

```
# Read task from plans.json
task = json_result_of(harness plan-cli get ${task_id})  # reads from .claude/harness/plans.json
MaxIter = cli_override OR task.MaxIter OR 10
verify_cmd = cli_override OR task.verify_cmd

# Serialization guard: abort if another [ralph] task is already cc:WIP
if any_other_ralph_task_is_wip():  # harness plan-cli list --status cc:WIP
    print "Another [ralph] task is currently cc:WIP. Ralph tasks serialize — wait for it to finish."
    exit 1

# Mark task WIP
harness plan-cli update "${task_id}" --status cc:WIP

for iter in range(0, MaxIter):

    # ── 1. Build prompt ──────────────────────────────────────────────────────
    prompt = build_ralph_prompt(task, iter, MaxIter, RALPH_WORKTREE, prior_verify_result)

    # ── 2. Dispatch worker ───────────────────────────────────────────────────
    if iter == 0:
        # First iteration: create the persistent worktree
        result = Agent(
            subagent_type="harness:ralph-worker",
            prompt=prompt,
            isolation="worktree"        # ← creates the worktree; returns worktreePath
        )
        RALPH_WORKTREE = result.worktreePath

    else:
        # Subsequent iterations: enter the SAME worktree
        if NOT ENTER_WT_LOADED:
            ToolSearch(query="select:EnterWorktree")   # ← load deferred-tool schema (once)
            ENTER_WT_LOADED = true
        EnterWorktree(path=RALPH_WORKTREE)             # ← change CWD into the worktree
        result = Agent(
            subagent_type="harness:ralph-worker",
            prompt=prompt
            # isolation NOT passed — worker inherits current CWD (the persistent worktree)
        )

    # ── 3. Authoritative verify ──────────────────────────────────────────────
    verify_exit, verify_stderr = run_verify(verify_cmd, cwd=RALPH_WORKTREE)
    prior_verify_result = {exit_code: verify_exit, stderr_tail: last_40_lines(verify_stderr)}

    # ── 4. Parse worker report ───────────────────────────────────────────────
    ralph_report = parse_ralph_worker_report(result.finalMessage)
    # ralph_report fields: schema, task_id, iteration, verify.exit_code, promise.asserted, files_changed, summary

    # ── 5. Detect stuck (no file changes) ────────────────────────────────────
    files_changed = git_diff_stat(RALPH_WORKTREE, since=PRIOR_COMMIT)   # empty list = no changes

    PRIOR_COMMIT = result.commit   # update for next iteration

    # ── 6. Decision matrix ───────────────────────────────────────────────────
    # Evaluate in this priority order:

    # ── STOP: VERIFY-MISMATCH (FT-RALPH-02) ──────────────────────────────────
    if ralph_report.verify.exit_code != verify_exit:
        reason = "blocked (worker self-report disagrees with authoritative verify: worker={W} actual={A})"
                  .format(W=ralph_report.verify.exit_code, A=verify_exit)
        harness plan-cli update "${task_id}" --status blocked --reason "${reason}"
        print FT-RALPH-02 summary with diff history and last verify stderr
        # Do NOT clean up worktree — preserve for inspection
        exit 1

    # ── STOP: STUCK (FT-RALPH-01) ────────────────────────────────────────────
    if files_changed.empty AND verify_exit != 0:
        reason = "blocked (ralph stuck at iter {N}: no file changes + verify exit {code})"
                  .format(N=iter, code=verify_exit)
        harness plan-cli update "${task_id}" --status blocked --reason "${reason}"
        print FT-RALPH-01 summary with diff history and last verify stderr
        # Do NOT clean up worktree — preserve for inspection
        exit 1

    # ── SUCCESS ───────────────────────────────────────────────────────────────
    if promise_tag_found(result.finalMessage) AND ralph_report.promise.asserted AND verify_exit == 0:
        commit_hash = merge_worktree_to_main(RALPH_WORKTREE)
        harness plan-cli update "${task_id}" --status cc:done --hash "${commit_hash}"  # authoritative
        print SUCCESS summary: iterations_used, files_changed_total, final_verify_output
        cleanup_worktree(RALPH_WORKTREE)    # ← remove worktree after successful merge
        exit 0

    # ── CONTINUE ─────────────────────────────────────────────────────────────
    # verify_exit != 0 and iter+1 < MaxIter: loop back with failure context
    print "iter {iter}: verify exit {code}, continuing...".format(iter=iter, code=verify_exit)

# ── STOP: MAX-ITER (FT-RALPH-03) ─────────────────────────────────────────────
reason = "blocked (ralph max-iter exhausted after {N} iterations, last verify exit: {code})"
          .format(N=MaxIter, code=verify_exit)
harness plan-cli update "${task_id}" --status blocked --reason "${reason}"
print FT-RALPH-03 summary with iteration history
# Do NOT clean up worktree — preserve for manual inspection
exit 1
```

---

## Prompt Template

The following structured prompt is passed to each worker iteration.
The `[prior iteration context]` block is **omitted entirely** when `iter == 0`.

```
[ralph context]
You are iteration {iter} of up to {MaxIter}. Worktree: {RALPH_WORKTREE}

[task]
{task.description}
DoD: {task.DoD}

[promise format]
When DoD is met AND verify exits 0, emit on its own line:
<promise>{task.DoD}</promise>
Then emit ralph-worker-report.v1 JSON.

[verify]
Run this command BEFORE emitting the promise:
  {verify_cmd}
Only emit the promise if exit code is 0.

[prior iteration context]  ← omit this entire block when iter == 0
Iteration {iter-1} verify exit code: {code}
Last 40 lines of stderr (ANSI-stripped, max 2KB):
```
{stderr_tail}
```
Files changed last iteration: {files_changed_last_iter}
(Read prior scratch/output files in the worktree for full context.)

[constraints]
Do not spawn nested agents. Do not modify files outside {RALPH_WORKTREE}.
```

---

## Decision Matrix

All four terminal states are handled. Evaluated in priority order after each iteration:

| Condition | Terminal State | Action |
|-----------|----------------|--------|
| `promise_tag_found AND ralph_report.promise.asserted AND verify_exit == 0` | SUCCESS | Commit + merge worktree to main, mark `cc:done [hash]`, clean up worktree |
| `ralph_report.verify.exit_code != verify_exit` | FT-RALPH-02 VERIFY-MISMATCH | Hard stop; mark `blocked (worker self-report disagrees: worker={W} actual={A})`; preserve worktree |
| `files_changed.empty AND verify_exit != 0` | FT-RALPH-01 STUCK | Hard stop; mark `blocked (ralph stuck at iter {N}: no file changes + verify exit {code})`; preserve worktree |
| `iter >= MaxIter` | FT-RALPH-03 MAX-ITER | Hard stop; mark `blocked (ralph max-iter exhausted after {N} iterations, last verify exit: {code})`; preserve worktree |
| Otherwise | CONTINUE | Build next prompt with failure context; increment iter |

---

## Promise Tag Scanning

Promise detection scans **only the final assistant message** of the worker result — not tool call echoes, not intermediate output.

```
# Regex: match <promise>...</promise> anchored to content boundaries
PROMISE_REGEX = r'<promise>(.*?)</promise>'

# Scan only result.finalMessage (the last assistant text block)
match = regex_search(PROMISE_REGEX, result.finalMessage)
promise_text = match.group(1) if match else None
promise_tag_found = (promise_text is not None)
```

The promise text must match `task.DoD` (case-insensitive, whitespace-normalized) to be considered valid.
A promise tag appearing in a tool call output echo or prior-context block is ignored.

---

## Terminal State Handling

### SUCCESS
1. Run `git -C {RALPH_WORKTREE} log --oneline` to collect the iteration commit history
2. Cherry-pick or merge the worktree branch into main: `git -C {PROJECT_ROOT} merge --squash {RALPH_WORKTREE_BRANCH}` then commit
3. Run `harness plan-cli update "${task_id}" --status cc:done --hash "${commit_hash}"` (authoritative)
4. Remove the worktree: `git worktree remove --force {RALPH_WORKTREE}`
5. Print a success summary: iterations used, files changed, verify command output

### FT-RALPH-01 STUCK
1. Run `harness plan-cli update "${task_id}" --status blocked --reason "ralph stuck at iter {N}: no file changes + verify exit {code}"`
2. Print the last 40 lines of verify stderr
3. Print `git -C {RALPH_WORKTREE} log --oneline` showing prior iterations
4. **Do not remove the worktree** — preserve at `RALPH_WORKTREE` for manual inspection
5. Exit non-zero

### FT-RALPH-02 VERIFY-MISMATCH
1. Run `harness plan-cli update "${task_id}" --status blocked --reason "worker self-report disagrees with authoritative verify: worker={W} actual={A}"`
2. Print both verify outputs side by side for diagnosis
3. Print the worker's `ralph-worker-report.v1` JSON for inspection
4. **Do not remove the worktree** — preserve for inspection
5. Exit non-zero

### FT-RALPH-03 MAX-ITER
1. Run `harness plan-cli update "${task_id}" --status blocked --reason "ralph max-iter exhausted after {N} iterations, last verify exit: {code}"`
2. Print iteration history: `git -C {RALPH_WORKTREE} log --oneline`
3. Print final verify stderr
4. **Do not remove the worktree** — preserve for manual inspection or handoff
5. Exit non-zero

---

## Worktree Pattern — Key Constraints

These constraints come directly from `docs/spikes/ralph-worktree-persistence.md` (FEASIBLE classification):

1. `ToolSearch(query="select:EnterWorktree")` must be called before the **first** `EnterWorktree` call — it is a deferred tool and its schema must be fetched first. Call it once and set `ENTER_WT_LOADED = true`.
2. Do NOT pass `isolation="worktree"` on iterations 1..N — that would create a new fresh worktree and break persistence.
3. `RALPH_WORKTREE` is the single source of truth for worktree state. Store it immediately after iter 0 completes.
4. The orchestrator (not the worker) is responsible for `EnterWorktree`. The worker receives no special isolation param on iterations 1+.
5. The Verify command is always run **by the orchestrator** (not trusted from the worker) to prevent FT-RALPH-02 hallucination.
