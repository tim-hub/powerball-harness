# Breezing Mode

Full Lead/Worker/Reviewer team execution flow with Phase A (preparation), Phase B (delegate + review + cherry-pick), Phase C (reporting), and sprint-contract orchestration.

---

Team execution with Lead / Worker / Reviewer role separation.
In Codex, this assumes native subagent orchestration using `spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`,
and does not follow the old TeamCreate / TaskCreate-based approach.

**Permission Policy**:
- The current shipped default is `bypassPermissions`
- `--auto-mode` is treated as an opt-in rollout flag for compatible parent sessions
- Do not write the undocumented `autoMode` value to `permissions.defaultMode` or agent frontmatter `permissionMode`

> **CC v2.1.69+**: Nested teammates are prohibited by the platform,
> so do not add redundant nested prevention wording to Worker/Reviewer prompts.

```
Lead (this agent)
├── Worker (task-worker agent) — Implementation
└── Reviewer (code-reviewer agent) — Review
```

**Phase A: Pre-delegate (Preparation)**:
0. **Entry-point drift check** — run before dependency analysis:
   ```bash
   bash "${CLAUDE_PLUGIN_ROOT}/scripts/plans-drift-check.sh"
   ```
   - Exit 0: proceed to step 1
   - Exit 1: display drift report and prompt "Stale markers detected. Proceed anyway? (y/N)". Stop if user declines.
1. Identify target tasks with `harness plan-cli list` — the SSOT is `.claude/harness/plans.json`
2. Analyze the dependency graph and determine execution order (each task's `depends` field)
3. Effort scoring for each task (ultrathink injection decision)
4. Generate all sprint contracts concurrently (contracts are independent):
   ```
   contracts = parallel_map(tasks, λ task:
       generate-sprint-contract.sh {task.number}
       → enrich-sprint-contract.sh
       → ensure-sprint-contract-ready.sh
   )
   ```
   - Target: Phase A duration ≤ max(per-task contract time) + 10s overhead
     (not sum(per-task contract time) which scales linearly with task count)
   - Contract content is byte-identical to sequential generation (stateless scripts)
   - If any contract fails ensure-sprint-contract-ready check → stop Phase A and escalate

**Phase B: Delegate (Worker spawn → review → cherry-pick)**:

Execute the following **sequentially** for each task (in dependency order):

> **API Note**: The following is written in Claude Code API syntax.
> In Codex environments, read `Agent(...)` as `spawn_agent(...)`, `SendMessage(...)` as `send_input(...)`.
> See the API mapping table in `team-composition.md` for details.

```
universal_violations = []  # Accumulated across tasks; injected into next Worker briefing

for task in execution_order:
    # B-1. Generate sprint-contract
    contract_path = bash("${CLAUDE_SKILL_DIR}/../../scripts/generate-sprint-contract.sh {task.number}")  # pseudocode — plugin-local
    contract_path = bash("${CLAUDE_SKILL_DIR}/../../scripts/enrich-sprint-contract.sh {contract_path} --check \"Verify DoD from reviewer perspective\" --approve")  # pseudocode — plugin-local
    bash("${CLAUDE_SKILL_DIR}/../../scripts/ensure-sprint-contract-ready.sh {contract_path}")  # pseudocode — plugin-local

    # B-2 (NEW): [ralph] marker check — pre-dispatch, before standard worker spawn
    # If the task description contains "[ralph]", delegate to harness-ralph-loop instead.
    # Ralph tasks serialize within a session (only one Ralph loop runs at a time).
    if "[ralph]" in task.description:
        harness plan-cli update <task-id> --status cc:WIP  # authoritative (plans.json)
        ralph_result = Skill(name="harness-ralph-loop", args=task.id)
        # ralph_result terminal state determines the plans.json update:
        #   - SUCCESS         → cc:done [hash]   (already written by harness-ralph-loop orchestrator)
        #   - FT-RALPH-01     → blocked (ralph stuck — no progress across iterations)
        #   - FT-RALPH-02     → blocked (verify mismatch — promise/verify disagreement)
        #   - FT-RALPH-03     → blocked (max-iter exhausted without success)
        # All plans.json updates are handled by harness-ralph-loop itself; Lead skips B-2.5 through B-5.
        continue  # Skip the standard worker/reviewer loop for this task

    # B-2. Worker spawn (foreground, worktree isolation)
    # Agent tool return value contains agentId — used for SendMessage in fix loop
    harness plan-cli update <task-id> --status cc:WIP  # Update on start (unstarted tasks remain cc:TODO) — authoritative (plans.json)

    violation_preamble = ""
    if universal_violations:
        violation_preamble = "Universal violations from prior tasks in this session — do NOT repeat these:\n" + "\n".join(f"- {v}" for v in universal_violations) + "\n\n"

    worker_result = Agent(
        subagent_type="harness:worker",
        prompt=f"{violation_preamble}Task: {task.content}\nDoD: {task.DoD}\ncontract_path: {contract_path}\nmode: breezing",
        isolation="worktree",
        run_in_background=false  # Foreground execution → wait for Worker completion
    )
    worker_id = worker_result.agentId  # Retain for SendMessage
    # worker_result contains {commit, worktreePath, files_changed, summary}

    # B-2.5. Worker self-review gate (worker-report.v1)
    # Worker must emit worker-report.v1 before Lead spawns Reviewer.
    # Lead validates; up to 2 amendment cycles if the report is incomplete.
    # See worker-self-review.md for schema details.
    self_review = worker_result.worker_report  # JSON block emitted at end of Worker output
    amendment_count = 0
    while not is_valid_worker_report(self_review) and amendment_count < 2:
        SendMessage(to=worker_id, message="worker-report.v1 incomplete — all 5 SR rules must have verified:true and non-empty evidence. Please re-emit.")
        updated = wait_for_response(worker_id)
        self_review = updated.worker_report
        amendment_count++
    if not is_valid_worker_report(self_review):
        → Escalate: "Worker failed to produce valid self-review after 2 amendments" — stop task, do not cherry-pick

    # B-3. Lead executes review (Codex exec priority)
    diff_text = git("-C", worker_result.worktreePath, "show", worker_result.commit)
    verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
    profile = jq(contract_path, ".review.reviewer_profile")
    review_input = "review-output.json"
    if profile == "runtime":
        review_input = bash("cd {worker_result.worktreePath} && ${CLAUDE_SKILL_DIR}/../../scripts/run-contract-review-checks.sh {contract_path}")  # pseudocode — plugin-local
        runtime_verdict = jq(review_input, ".verdict")
        if runtime_verdict == "REQUEST_CHANGES":
            verdict = "REQUEST_CHANGES"
        elif runtime_verdict == "DOWNGRADE_TO_STATIC":
            pass  # No runtime validation command → use static verdict as-is
    if profile == "browser":
        # browser artifact generates a PENDING_BROWSER scaffold.
        # Actual browser execution is handled by the reviewer agent in a subsequent step.
        # Write the static review verdict to review-result (not PENDING_BROWSER).
        browser_artifact = bash("${CLAUDE_SKILL_DIR}/../../scripts/generate-browser-review-artifact.sh {contract_path}")  # pseudocode — plugin-local
        # browser artifact is saved for reference, but review-result verdict remains static
    # If review_input is DOWNGRADE_TO_STATIC, use the static review result
    if review_input != "review-output.json" and jq(review_input, ".verdict") == "DOWNGRADE_TO_STATIC":
        review_input = "review-output.json"  # Fall back to static review result
    bash("${CLAUDE_SKILL_DIR}/../../scripts/write-review-result.sh {review_input} {latest_commit}")  # pseudocode — plugin-local

    # B-3.5. Collect universal violations (scope: universal findings → injected into next Worker)
    # Reviewer memory updates may include a "scope" field: "universal" or "task-specific".
    # Universal-scope findings are patterns that should never recur in ANY task in this session.
    review_result = jq(".claude/state/review-result.json", ".")
    for finding in review_result.get("observations", []) + review_result.get("major_issues", []):
        if finding.get("scope") == "universal":
            universal_violations.append(f"[{finding['category']}] {finding['issue']}")

    # B-4. Fix loop (on REQUEST_CHANGES, up to 3 times)
    # Worker has completed in foreground, but can be resumed via SendMessage
    # (CC: SendMessage(to: agentId) / Codex: resume_agent(agent_id) + send_input)
    review_count = 0
    latest_commit = worker_result.commit
    while verdict == "REQUEST_CHANGES" and review_count < 3:
        SendMessage(to=worker_id, message="Issues found: {issues}\nPlease fix and amend")
        # Worker fixes → amends → returns updated commit hash
        updated_result = wait_for_response(worker_id)
        latest_commit = updated_result.commit
        diff_text = git("-C", worker_result.worktreePath, "show", latest_commit)
        verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
        review_count++

    # B-5. APPROVE → cherry-pick to main
    if verdict == "APPROVE":
        git cherry-pick --no-commit {latest_commit}  # worktree → main
        git commit -m "{task.content}"
        harness plan-cli update <task-id> --status cc:done --hash {hash}  # authoritative (plans.json)
    else:
        → Escalate to user

    # B-6. Progress feed
    print("📊 Progress: Task {completed}/{total} done — {task.content}")
```

## Sprint Contract

A `sprint-contract` is a small contract file that defines "what passes this task" in a format readable by both machines and humans.
The default storage location is `.claude/state/contracts/<task-id>.sprint-contract.json`.

> **Concurrency note**: In Breezing Phase A, contracts for all tasks are generated concurrently before any Worker is spawned.

```bash
"${CLAUDE_SKILL_DIR}/../../scripts/generate-sprint-contract.sh" 32.1.1
```

The generated artifact includes:

- `checks`: Verification items decomposed from the DoD
- `non_goals`: What is out of scope for this task
- `runtime_validation`: Validation commands such as test, lint, typecheck
- `browser_validation`: UI flow verification items for the browser reviewer
- `browser_mode`: `scripted` or `exploratory`
- `route`: Whether the browser reviewer uses `playwright` / `agent-browser` / `chrome-devtools`
- `risk_flags`: `needs-spike`, `security-sensitive`, `ux-regression`, etc.
- `reviewer_profile`: `static`, `runtime`, `browser`

**Phase C: Post-delegate (Integration & Reporting)**:
1. Aggregate commit logs for all tasks
2. Output a **Rich Completion Report** (see [`${CLAUDE_SKILL_DIR}/templates/completion-report.md`](${CLAUDE_SKILL_DIR}/templates/completion-report.md))
3. Final check with `harness plan-cli list --status all` (verify all tasks are cc:done)
