# Solo Mode

Full 12-step single-task implementation flow with drift check, advisor preflight, TDD, sprint-contract generation, review, and completion report.

---

#### Step 0: Entry-point drift check

Before selecting execution mode, run a lightweight sync pass to catch stale Plans.md markers:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/scripts/plans-drift-check.sh"
```

- **Exit 0** (no stale markers): proceed immediately to Step 1
- **Exit 1** (stale markers detected): display the drift report, then prompt the user:
  "Stale markers detected. Proceed anyway? (y/N) [default: N]"
  - If user confirms (`y`): continue to Step 1
  - If user declines or no response: stop with message "Run /harness-plan sync first to resolve marker drift"

This check is intentionally lightweight — it only inspects commit messages, not file content. For a thorough sync, run `/harness-plan sync` explicitly.

1. Identify the target task with `harness plan-cli list` — the SSOT is `.claude/harness/plans.json`
   - **Dependency check**: Before claiming a task, inspect its `depends` (via `harness plan-cli get <task-id>`). If any listed dependency is not yet `cc:done`/`pm:confirmed`, skip this task and select the next eligible one. For `.b` tasks whose `.a` is still open: redirect the Worker to `.a` first.
   - **If plans.json does not exist**: if a legacy `Plans.md` is present, run `harness plan-cli migrate` to convert it; otherwise auto-invoke `harness-plan create --ci` to generate it, then continue
   - **If the conversation contains unlisted tasks**: Extract requirements from the recent conversation context and add them with `harness plan-cli add-task <phase-id> --name "…" --dod "…"` (status defaults to `cc:TODO`)
     - Extraction logic: Detect action verbs from user statements ("add...", "fix...", "implement...")
     - After adding, display "Added the following tasks" with a 5-second timeout prompt (default: continue)
1.5. **Task Background Check** (30 seconds):
   - Infer and display the **purpose** (the problem this task solves) in one line from the task's "Content" and "DoD"
   - Use `git grep` / `Glob` to infer and display the **impact scope** (files/modules affected by changes)
   - If confident in the inference: proceed directly to implementation (no flow delay)
   - If not confident: ask the user one question only ("Is this understanding correct?")
1.6. **Advisor Preflight** (when `advisor.enabled` or `--advisor`):
   - If task has `<!-- advisor:required -->` marker: consult `harness:advisor` with `reason_code: high_risk_preflight`
   - On `PLAN`: proceed with suggested approach
   - On `CORRECTION`: apply correction before starting
   - On `STOP`: escalate to user immediately
2. **`[ralph]` marker pre-dispatch check**: If the task description contains `[ralph]`, delegate to
   `harness-ralph-loop` instead of the standard solo flow:
   ```
   # [ralph] pre-dispatch (solo mode)
   if "[ralph]" in task.description:
       harness plan-cli update <task-id> --status cc:WIP
       ralph_result = Skill(name="harness-ralph-loop", args=task.id)
       # ralph_result terminal state determines the plans.json update:
       #   - SUCCESS         → cc:done [hash]   (already written by harness-ralph-loop orchestrator)
       #   - FT-RALPH-01     → blocked (ralph stuck — no progress across iterations)
       #   - FT-RALPH-02     → blocked (verify mismatch — promise/verify disagreement)
       #   - FT-RALPH-03     → blocked (max-iter exhausted without success)
       # All plans.json updates are handled by harness-ralph-loop itself; skip steps 3–13.
       return ralph_result
   ```
   Ralph tasks serialize within a session (only one Ralph loop runs at a time). If the task does NOT
   have `[ralph]`, continue with the standard solo flow below.
2.5. Update task to `cc:WIP`
   - `harness plan-cli update <task-id> --status cc:WIP` (writes to plans.json, the authoritative SSOT)
3. **TDD Phase** — behaviour depends on task tag:

   | Task type | Step 3 action |
   |---|---|
   | `[tdd:test-first]` (an `.a` task) | **This task IS the TDD phase.** Write the failing test file; confirm it runs red. Commit as `test: failing tests for {{feature}}`. |
   | `.b` task (follows a `.a`) | Confirm tests from `.a` still run red, then proceed to Step 6. No new test file needed. |
   | No split (legacy task, no `[skip:tdd]`) | Existing behaviour — create test file first, confirm failure. |
   | `[skip:tdd]` present | Skip Step 3 entirely. |
4. Generate `sprint-contract.json` with `"${CLAUDE_SKILL_DIR}/../../scripts/generate-sprint-contract.sh" <task-id>`
5. Add Reviewer perspective with `"${CLAUDE_SKILL_DIR}/../../scripts/enrich-sprint-contract.sh"` and confirm approved status with `"${CLAUDE_SKILL_DIR}/../../scripts/ensure-sprint-contract-ready.sh"`
6. Implement code (Green) (Read/Write/Edit/Bash)
7. Auto-Refinement with `/simplify` (skip with `--no-simplify`)
8. **Auto Review Stage** (see [`review-loop.md`](${CLAUDE_SKILL_DIR}/references/review-loop.md)):
   - Execute review with Codex exec priority → fallback to internal Reviewer agent
   - If `sprint-contract.json`'s `reviewer_profile` is `runtime`, execute `"${CLAUDE_SKILL_DIR}/../../scripts/run-contract-review-checks.sh"`
   - On REQUEST_CHANGES: fix based on feedback → re-review (up to 3 times)
   - Proceed to next step on APPROVE. Self-check alone does not confirm completion
9. Normalize and save review artifact with `"${CLAUDE_SKILL_DIR}/../../scripts/write-review-result.sh"`
10. Update task to `cc:done`
   - `harness plan-cli update <task-id> --status cc:done` (writes to plans.json, the authoritative SSOT). No commit is made by default.
   - If `--commit` was passed: run `git commit`, get the abbreviated 7-char hash with `git log --oneline -1`, then `harness plan-cli update <task-id> --status cc:done --hash a1b2c3d`.
   - **Phase-close check**: Scan the current phase with `harness plan-cli list --phase <id>`.
     - All tasks except `[verify:e2e]` are `cc:done` AND `N.e2e` is `cc:TODO`: → "Phase N implementation complete. Next: run the E2E verification task (N.e2e)." Auto-select `N.e2e` as the next task (or surface it in manual mode).
     - `[verify:e2e]` task is `cc:done`: phase is fully closed.
     - No `[verify:e2e]` task (docs/config-only phase): phase closes normally.
11. **Rich Completion Report** (see [`${CLAUDE_SKILL_DIR}/templates/completion-report.md`](${CLAUDE_SKILL_DIR}/templates/completion-report.md))
12. **Automatic Re-ticketing on Failure** (test/CI failure only):
    - Check test execution results
    - On failure: save fix task proposal to state, add it via `harness plan-cli add-task` on approval (see [`re-ticketing.md`](${CLAUDE_SKILL_DIR}/references/re-ticketing.md))
    - On success: proceed to next task
