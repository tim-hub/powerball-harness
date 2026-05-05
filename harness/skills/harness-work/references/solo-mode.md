# Solo Mode

Full 13-step single-task implementation flow with drift check, advisor preflight, TDD, sprint-contract generation, review, and completion report.

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

1. Read Plans.md and identify the target task
   - **If Plans.md does not exist**: Auto-invoke `harness-plan create --ci` → Generate Plans.md and continue
   - If header lacks DoD / Depends columns: `Plans.md is in the old format. Please regenerate with harness-plan create.` → **Stop**
   - **If the conversation contains unlisted tasks**: Extract requirements from the recent conversation context and auto-append to Plans.md as `cc:TODO`
     - Extraction logic: Detect action verbs from user statements ("add...", "fix...", "implement...")
     - Appended entries conform to v2 format (Task / Content / DoD / Depends / Status)
     - After appending, display "Added the following to Plans.md" with a 5-second timeout prompt (default: continue)
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
       Plans.md: task.status = "cc:WIP"
       ralph_result = Skill(name="harness-ralph-loop", args=task.id)
       # ralph_result terminal state determines Plans.md update:
       #   - SUCCESS         → cc:done [hash]   (already written by harness-ralph-loop orchestrator)
       #   - FT-RALPH-01     → blocked (ralph stuck — no progress across iterations)
       #   - FT-RALPH-02     → blocked (verify mismatch — promise/verify disagreement)
       #   - FT-RALPH-03     → blocked (max-iter exhausted without success)
       # All Plans.md updates are handled by harness-ralph-loop itself; skip steps 3–13.
       return ralph_result
   ```
   Ralph tasks serialize within a session (only one Ralph loop runs at a time). If the task does NOT
   have `[ralph]`, continue with the standard solo flow below.
2.5. Update task to `cc:WIP`
3. **TDD Phase** (when `[skip:tdd]` is absent & test framework exists):
   a. Create test file first (Red)
   b. Confirm failure
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
10. Auto-commit with `git commit` (skip with `--no-commit`)
11. Update task to `cc:Done` (with commit hash)
   - Get the latest commit hash (abbreviated 7 chars) with `git log --oneline -1`
   - Update Plans.md Status to `cc:Done [a1b2c3d]` format
   - If no commit (`--no-commit`), use `cc:Done` without hash
12. **Rich Completion Report** (see [`${CLAUDE_SKILL_DIR}/templates/completion-report.md`](${CLAUDE_SKILL_DIR}/templates/completion-report.md))
13. **Automatic Re-ticketing on Failure** (test/CI failure only):
    - Check test execution results
    - On failure: save fix task proposal to state, add to Plans.md via approval command (see [`re-ticketing.md`](${CLAUDE_SKILL_DIR}/references/re-ticketing.md))
    - On success: proceed to next task
