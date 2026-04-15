---
name: harness-work
description: "Use when implementing, executing, or running Plans.md tasks — single task, parallel workers, or full team/breezing run. Accepts specific task numbers or ranges. Do NOT load for: planning, review, release, or setup."
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[all] [task-number|range] [--codex] [--parallel N] [--no-commit] [--resume id] [--breezing] [--auto-mode]"
effort: high
---

# Harness Work

Unified execution skill for Harness.
Consolidates the following legacy skills:

- `work` — Plans.md task implementation (auto scope detection)
- `impl` — Feature implementation (task-based)
- `breezing` — Full team auto-execution
- `parallel-workflows` — Parallel workflow optimization
- `ci` — CI failure recovery

## Quick Reference

| User Input | Mode | Behavior |
|------------|------|----------|
| `harness-work` | **auto** | Auto-selects based on task count (see below) |
| `harness-work all` | **auto** | Executes all incomplete tasks in auto mode |
| `harness-work 3` | solo | Immediately executes task 3 only |
| `harness-work --parallel 5` | parallel | Forces parallel execution with 5 workers |
| `harness-work --codex` | codex | Delegates to Codex CLI (explicit only) |
| `harness-work --breezing` | breezing | Forces team execution |

## Execution Mode Auto Selection (auto-selection when no flag is specified)

When no explicit mode flag (`--parallel`, `--breezing`, `--codex`) is provided,
the optimal mode is automatically selected based on the number of target tasks:

| Target Task Count | Auto-Selected Mode | Reason |
|-------------------|-------------------|--------|
| **1 task** | Solo | Minimal overhead. Direct implementation is fastest |
| **2-3 tasks** | Parallel (Task tool) | Threshold where Worker isolation benefits emerge |
| **4+ tasks** | Breezing | Lead coordination + Worker parallelism + Reviewer independence is effective |

### Rules

1. **Explicit flags always override auto mode**
   - `--parallel N` → Parallel mode (regardless of task count)
   - `--breezing` → Breezing mode (regardless of task count)
   - `--codex` → Codex mode (regardless of task count)
2. **`--codex` activates only when explicitly specified**. Not auto-selected because Codex CLI may not be installed in some environments
3. `--codex` can be combined with other modes: `--codex --breezing` → Codex + Breezing

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `all` | Target all incomplete tasks | - |
| `N` or `N-M` | Task number/range specification | - |
| `--parallel N` | Number of parallel workers | auto |
| `--sequential` | Force sequential execution | - |
| `--codex` | Delegate implementation to Codex CLI (explicit only, not auto-selected) | false |
| `--no-commit` | Suppress auto-commit | false |
| `--resume <id\|latest>` | Resume previous session | - |
| `--breezing` | Lead/Worker/Reviewer team execution | false |
| `--no-tdd` | Skip TDD phase | false |
| `--no-simplify` | Skip Auto-Refinement | false |
| `--auto-mode` | Explicitly enable Auto Mode rollout. Only considered when the parent session's permission mode is compatible | false |

> **Token Optimization (v2.1.69+)**: For lightweight tasks that don't involve git operations,
> enable `includeGitInstructions: false` in plugin settings to
> reduce prompt token usage.

## Scope Dialog (when no arguments provided)

```
harness-work
How far do you want to go?
1) Next task: The next incomplete task in Plans.md → Execute in Solo mode
2) All (recommended): Complete all remaining tasks → Auto-select mode based on task count
3) Specify numbers: Enter task numbers (e.g., 3, 5-7) → Auto-select mode based on count
```

If arguments are provided, execute immediately (skip dialog):
- `harness-work all` → All tasks, auto mode selection
- `harness-work 3-6` → 4 tasks, so Breezing is auto-selected

## Effort Level Control (v2.1.68+, simplified in v2.1.72)

Claude Code v2.1.68 sets **medium effort** (`◐`) as default for Opus 4.6.
v2.1.72 removed the `max` level, simplifying to 3 levels: `low(○)/medium(◐)/high(●)`.
`/effort auto` resets to default.
For complex tasks, use the `ultrathink` keyword to enable high effort (`●`).

### Multi-Factor Scoring

At task start, the following scores are summed, and **ultrathink is injected when the threshold reaches 3 or above**:

| Factor | Condition | Score |
|--------|-----------|-------|
| File count | 4+ files to be changed | +1 |
| Directory | Includes core/, guardrails/, security/ | +1 |
| Keywords | Contains architecture, security, design, migration | +1 |
| Failure history | Agent memory contains failure records for the same task | +2 |
| Explicit specification | PM template includes ultrathink notation | +3 (auto-adopted) |

### Injection Method

When score >= 3, prepend `ultrathink` to the Worker spawn prompt.
The same logic applies in breezing mode (managed centrally by harness-work).

## Execution Mode Details

### Solo Mode (auto-selected for 1 task)

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
2. Update task to `cc:WIP`
3. **TDD Phase** (when `[skip:tdd]` is absent & test framework exists):
   a. Create test file first (Red)
   b. Confirm failure
4. Generate `sprint-contract.json` with `scripts/generate-sprint-contract.sh <task-id>`
5. Add Reviewer perspective with `scripts/enrich-sprint-contract.sh` and confirm approved status with `scripts/ensure-sprint-contract-ready.sh`
6. Implement code (Green) (Read/Write/Edit/Bash)
7. Auto-Refinement with `/simplify` (skip with `--no-simplify`)
8. **Auto Review Stage** (see "Review Loop"):
   - Execute review with Codex exec priority → fallback to internal Reviewer agent
   - If `sprint-contract.json`'s `reviewer_profile` is `runtime`, execute `scripts/run-contract-review-checks.sh`
   - On REQUEST_CHANGES: fix based on feedback → re-review (up to 3 times)
   - Proceed to next step on APPROVE. Self-check alone does not confirm completion
9. Normalize and save review artifact with `scripts/write-review-result.sh`
10. Auto-commit with `git commit` (skip with `--no-commit`)
11. Update task to `cc:Done` (with commit hash)
   - Get the latest commit hash (abbreviated 7 chars) with `git log --oneline -1`
   - Update Plans.md Status to `cc:Done [a1b2c3d]` format
   - If no commit (`--no-commit`), use `cc:Done` without hash
12. **Rich Completion Report** (see "Completion Report Format")
13. **Automatic Re-ticketing on Failure** (test/CI failure only):
    - Check test execution results
    - On failure: save fix task proposal to state, add to Plans.md via approval command (see "Automatic Re-ticketing of Failed Tasks")
    - On success: proceed to next task

### Parallel Mode (auto-selected for 2-3 tasks / forced with `--parallel N`)

Execute `[P]`-marked tasks with N workers in parallel.
When explicitly specified with `--parallel N`, this mode is used regardless of task count.
If write conflicts to the same file occur, isolate with git worktree.

### Codex Mode (`--codex` explicit only)

> Load [`${CLAUDE_SKILL_DIR}/references/codex-work.md`](${CLAUDE_SKILL_DIR}/references/codex-work.md)
> only when `command -v codex` succeeds **and** the user passes `--codex` or explicitly asks to use Codex.

### Breezing Mode (auto-selected for 4+ tasks / forced with `--breezing`)

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
1. Read Plans.md and identify target tasks
2. Analyze the dependency graph and determine execution order (Depends column)
3. Effort scoring for each task (ultrathink injection decision)
4. Generate `sprint-contract.json` with `scripts/generate-sprint-contract.sh`
5. Add Reviewer perspective with `scripts/enrich-sprint-contract.sh` and stop if unapproved with `scripts/ensure-sprint-contract-ready.sh`

**Phase B: Delegate (Worker spawn → review → cherry-pick)**:

Execute the following **sequentially** for each task (in dependency order):

> **API Note**: The following is written in Claude Code API syntax.
> In Codex environments, read `Agent(...)` as `spawn_agent(...)`, `SendMessage(...)` as `send_input(...)`.
> See the API mapping table in `team-composition.md` for details.

```
for task in execution_order:
    # B-1. Generate sprint-contract
    contract_path = bash("scripts/generate-sprint-contract.sh {task.number}")
    contract_path = bash("scripts/enrich-sprint-contract.sh {contract_path} --check \"Verify DoD from reviewer perspective\" --approve")
    bash("scripts/ensure-sprint-contract-ready.sh {contract_path}")

    # B-2. Worker spawn (foreground, worktree isolation)
    # Agent tool return value contains agentId — used for SendMessage in fix loop
    Plans.md: task.status = "cc:WIP"  # Update on start (unstarted tasks remain cc:TODO)

    worker_result = Agent(
        subagent_type="claude-code-harness:worker",
        prompt="Task: {task.content}\nDoD: {task.DoD}\ncontract_path: {contract_path}\nmode: breezing",
        isolation="worktree",
        run_in_background=false  # Foreground execution → wait for Worker completion
    )
    worker_id = worker_result.agentId  # Retain for SendMessage
    # worker_result contains {commit, worktreePath, files_changed, summary}

    # B-3. Lead executes review (Codex exec priority)
    diff_text = git("-C", worker_result.worktreePath, "show", worker_result.commit)
    verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
    profile = jq(contract_path, ".review.reviewer_profile")
    review_input = "review-output.json"
    if profile == "runtime":
        review_input = bash("cd {worker_result.worktreePath} && scripts/run-contract-review-checks.sh {contract_path}")
        runtime_verdict = jq(review_input, ".verdict")
        if runtime_verdict == "REQUEST_CHANGES":
            verdict = "REQUEST_CHANGES"
        elif runtime_verdict == "DOWNGRADE_TO_STATIC":
            pass  # No runtime validation command → use static verdict as-is
    if profile == "browser":
        # browser artifact generates a PENDING_BROWSER scaffold.
        # Actual browser execution is handled by the reviewer agent in a subsequent step.
        # Write the static review verdict to review-result (not PENDING_BROWSER).
        browser_artifact = bash("scripts/generate-browser-review-artifact.sh {contract_path}")
        # browser artifact is saved for reference, but review-result verdict remains static
    # If review_input is DOWNGRADE_TO_STATIC, use the static review result
    if review_input != "review-output.json" and jq(review_input, ".verdict") == "DOWNGRADE_TO_STATIC":
        review_input = "review-output.json"  # Fall back to static review result
    bash("scripts/write-review-result.sh {review_input} {latest_commit}")

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
        Plans.md: task.status = "cc:Done [{hash}]"
    else:
        → Escalate to user

    # B-6. Progress feed
    print("📊 Progress: Task {completed}/{total} done — {task.content}")
```

### Sprint Contract

A `sprint-contract` is a small contract file that defines "what passes this task" in a format readable by both machines and humans.
The default storage location is `.claude/state/contracts/<task-id>.sprint-contract.json`.

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
2. Output a **Rich Completion Report** (Breezing template from "Completion Report Format")
3. Final check of Plans.md (verify all tasks are cc:Done)

## CI Failure Handling

When CI fails:

1. Check logs and identify the error
2. Implement fixes
3. Stop the auto-fix loop after 3 failures from the same cause
4. Summarize failure logs, attempted fixes, and remaining issues for escalation

## Automatic Re-ticketing of Failed Tasks

When tests/CI fail after task completion, auto-generate fix task proposals and reflect them in Plans.md after approval:

### Trigger Conditions

| Condition | Action |
|-----------|--------|
| Test failure after `cc:Done` | Save fix task proposal to state and wait for approval |
| CI failure (fewer than 3 times) | Implement fix and increment failure count |
| CI failure (3rd time) | Present fix task proposal + escalate |

### Auto-Generation of Fix Tasks

1. Classify failure cause (syntax_error / import_error / type_error / assertion_error / timeout / runtime_error)
2. Save fix task proposal to `.claude/state/pending-fix-proposals.jsonl`:
   - Number: Original task number + `.fix` suffix (e.g., `26.1.fix`)
   - Content: `fix: [original task name] - [failure cause category]`
   - DoD: Tests/CI pass
   - Depends: Original task number
3. When user sends `approve fix <task_id>`, add to Plans.md as `cc:TODO`
4. `reject fix <task_id>` discards the proposal. When there is only one pending item, `yes` / `no` responses are also accepted

## Review Loop

A quality verification stage that runs automatically after implementation completion (after step 5).
Applied uniformly across **all modes** (Solo / Parallel / Breezing).
In Parallel mode, each Worker executes the same loop as step 10 (external review acceptance).

### Review Execution Priority

```
1. Codex exec (priority, when available) — see ${CLAUDE_SKILL_DIR}/references/codex-work.md
   ↓ codex command does not exist or timeout (120s)
2. Internal Reviewer agent (fallback)
```

### APPROVE / REQUEST_CHANGES Verdict Criteria

The following threshold criteria are provided to reviewers, and the verdict is determined **solely by these criteria**.
Improvement suggestions outside these criteria are returned as `recommendations` but do not affect the verdict.

| Severity | Definition | Verdict Impact |
|----------|------------|----------------|
| **critical** | Security vulnerabilities, data loss risk, potential production incidents | 1 item → REQUEST_CHANGES |
| **major** | Breaking existing functionality, clear contradiction with specifications, test failures | 1 item → REQUEST_CHANGES |
| **minor** | Naming improvements, insufficient comments, style inconsistencies | No impact on verdict |
| **recommendation** | Best practice suggestions, future improvement ideas | No impact on verdict |

> **Important**: When only minor / recommendation items exist, **always return APPROVE**.
> "Nice-to-have improvements" are not grounds for REQUEST_CHANGES.

### Codex Exec Review (via official plugin)

> When Codex is available, load [`${CLAUDE_SKILL_DIR}/references/codex-work.md`](${CLAUDE_SKILL_DIR}/references/codex-work.md)
> for the full Codex exec review flow, verdict mapping, and AI Residuals scan details.

### Internal Reviewer Agent Fallback

When Codex exec is unavailable (`command -v codex` fails, or exit code != 0):

```
Agent tool: subagent_type="reviewer"
prompt: "Please review the following changes. Verdict criteria: critical/major → REQUEST_CHANGES, minor/recommendation only → APPROVE. diff: {git diff ${BASE_REF}}"
```

The Reviewer agent executes reviews safely in Read-only mode (Write/Edit/Bash disabled).

### Fix Loop (on REQUEST_CHANGES)

```
review_count = 0
MAX_REVIEWS = 3

while verdict == "REQUEST_CHANGES" and review_count < MAX_REVIEWS:
    1. Analyze review findings (critical / major only)
    2. Implement fixes for each finding
    3. Re-execute review (same criteria, same priority)
    review_count++

if review_count >= MAX_REVIEWS and verdict != "APPROVE":
    → Escalate to user
    → Display "Fixed 3 times but the following critical/major issues remain" + list of issues
    → Wait for user decision (continue / abort)
```

### Application in Breezing Mode

In Breezing mode, the **Lead** executes the review loop (see Phase B above):

1. Worker implements and commits in worktree → returns result to Lead
2. Lead reviews with Codex exec (priority) / Reviewer agent (fallback)
3. REQUEST_CHANGES → Lead sends fix instructions to Worker via SendMessage → Worker amends
4. After fix, re-review (up to 3 times)
5. APPROVE → Lead cherry-picks to main → Updates Plans.md to `cc:Done [{hash}]`

## Completion Report Format

A visual summary auto-output on task completion (after `cc:Done` + commit).
Designed to convey change content and impact even to non-technical stakeholders.

### Template

```
┌─────────────────────────────────────────────┐
│  ✓ Task {N} Done: {task name}                │
├─────────────────────────────────────────────┤
│                                              │
│  ■ What was done                             │
│    • {change 1}                              │
│    • {change 2}                              │
│                                              │
│  ■ What changed                              │
│    Before: {old behavior}                    │
│    After:  {new behavior}                    │
│                                              │
│  ■ Changed files ({N} files)                 │
│    {file path 1}                             │
│    {file path 2}                             │
│                                              │
│  ■ Remaining issues                          │
│    • Task {X} ({status}): {content}  ← Plans.md  │
│    • Task {Y} ({status}): {content}  ← Plans.md  │
│    ({M} incomplete tasks in Plans.md)        │
│                                              │
│  commit: {hash} | review: {APPROVE}           │
└─────────────────────────────────────────────┘
```

### Generation Rules

1. **What was done**: Auto-extracted from `git diff --stat HEAD~1` and commit message. Minimize technical jargon, start with verbs
2. **What changed**: Infer Before/After from the task's "Content" and "DoD". Emphasize user experience changes
3. **Changed files**: Retrieved from `git diff --name-only HEAD~1`. Abbreviate with count when exceeding 5 files
4. **Remaining issues**: List `cc:TODO` / `cc:WIP` tasks from Plans.md. Indicate whether they are already in Plans.md
5. **Review**: Display review result (APPROVE / REQUEST_CHANGES → APPROVE)

### Reporting in Parallel Mode

- **1 task** (when forced with `--parallel`): Use Solo template
- **Multiple tasks**: Use Breezing aggregate template (see below)

### Reporting in Breezing Mode

Output collectively after all tasks are complete. Each task is listed in abbreviated form (what was done + commit hash only),
followed by an overall summary (total changed files + remaining issues):

```
┌─────────────────────────────────────────────┐
│  ✓ Breezing Complete: {N}/{M} tasks          │
├─────────────────────────────────────────────┤
│                                              │
│  1. ✓ {task name 1}            [{hash1}]     │
│  2. ✓ {task name 2}            [{hash2}]     │
│  3. ✓ {task name 3}            [{hash3}]     │
│                                              │
│  ■ Overall changes                           │
│    {N} files changed, {A} insertions(+),     │
│    {D} deletions(-)                          │
│                                              │
│  ■ Remaining issues                          │
│    {K} incomplete tasks in Plans.md          │
│    • Task {X}: {content}                     │
│                                              │
└─────────────────────────────────────────────┘
```

## Related Skills

- `harness-plan` — Plan the tasks to execute
- `harness-sync` — Sync implementation with Plans.md
- `harness-review` — Review implementations
- `harness-release` — Version bump and release
