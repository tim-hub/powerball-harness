# Parallel Mode

Worker fan-out for 2-3 tasks with concurrent review phase — each Worker and its paired Reviewer run independently.

---

Execute tasks with N workers in parallel.
When explicitly specified with `--parallel N`, this mode is used regardless of task count.
If write conflicts to the same file occur, isolate with git worktree.

**Review Phase** (after all Workers complete):
Spawn one Reviewer agent per completed Worker, all running concurrently:
- Each Reviewer receives its Worker's diff independently
- Collect all verdicts; on any REQUEST_CHANGES, enter the fix loop for that task
- Constraint: Review phase wallclock ≤ max(single-review-time) + 10s coordination overhead
  (not the sum, which would be N × single-review-time)
- Verdict outcomes are identical to sequential review — each Reviewer sees the same diff

**Review execution order**:
1. Collect all completed Worker diffs
2. `for each worker_result in parallel:` spawn Reviewer agent (`run_in_background=true`)
3. `await_all reviewers` — collect verdicts
4. For any REQUEST_CHANGES: enter fix loop (sequential within that task)
5. For all APPROVE: proceed to commit
