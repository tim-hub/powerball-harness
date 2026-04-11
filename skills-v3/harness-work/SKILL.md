---
name: harness-work
description: "Use this skill whenever the user asks to implement, execute, build, code, 'do everything', 'run all tasks', or mentions harness-work, breezing, team run, parallel execution, or --codex. Also use when the user selects specific task numbers or ranges to execute. Do NOT load for: planning (use harness-plan), code review (use harness-review), release (use harness-release), or project setup (use harness-setup). Unified execution skill for Harness v3 — implements Plans.md tasks from single task to full parallel team runs."
allowed-tools: ["Read", "Write", "Edit", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[all] [task-number|range] [--codex] [--parallel N] [--no-commit] [--resume id] [--breezing] [--auto-mode]"
effort: high
---

# Harness Work (v3)

Unified execution skill for Harness v3.
Consolidates the following legacy skills:

- `work` — Plans.md task implementation (automatic scope detection)
- `impl` — Feature implementation (task-based)
- `breezing` — Full team auto-execution
- `parallel-workflows` — Parallel workflow optimization
- `ci` — CI failure recovery

## Quick Reference

| User Input | Mode | Behavior |
|------------|--------|------|
| `harness-work` | **auto** | Auto-determined by task count (see below) |
| `harness-work all` | **auto** | Execute all incomplete tasks in auto mode |
| `harness-work 3` | solo | Execute only task 3 immediately |
| `harness-work --parallel 5` | parallel | Parallel execution with 5 workers (forced) |
| `harness-work --codex` | codex | Delegate to Codex CLI (explicit only) |
| `harness-work --breezing` | breezing | Force team execution |

## Execution Mode Auto Selection (when no flags specified)

When no explicit mode flags (`--parallel`, `--breezing`, `--codex`) are provided,
the optimal mode is automatically selected based on the number of target tasks:

| Target Task Count | Auto-Selected Mode | Rationale |
|-------------|---------------|------|
| **1 task** | Solo | Minimal overhead. Direct implementation is fastest |
| **2-3 tasks** | Parallel (Task tool) | Threshold where Worker isolation benefits begin |
| **4+ tasks** | Breezing | Three-way separation of Lead coordination + Worker parallel + Reviewer independent is effective |

### Rules

1. **Explicit flags always override auto mode**
   - `--parallel N` -> Parallel mode (regardless of task count)
   - `--breezing` -> Breezing mode (regardless of task count)
   - `--codex` -> Codex mode (regardless of task count)
2. **`--codex` only activates when explicitly specified**. Not auto-selected because Codex CLI may not be installed in all environments
3. `--codex` can be combined with other modes: `--codex --breezing` -> Codex + Breezing

## Options

| Option | Description | Default |
|----------|------|----------|
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

> **Token Optimization (v2.1.69+)**: For lightweight tasks without git operations,
> enable `includeGitInstructions: false` in plugin settings to reduce
> prompt tokens.

## Scope Dialog (when no arguments provided)

```
harness-work
How far do you want to go?
1) Next task: Execute the next incomplete task from Plans.md -> Solo mode
2) All (recommended): Complete all remaining tasks -> Auto mode selection by task count
3) Specify numbers: Enter task numbers (e.g., 3, 5-7) -> Auto mode selection by count
```

If arguments are provided, execution starts immediately (dialog skipped):
- `harness-work all` -> All tasks, auto mode selection
- `harness-work 3-6` -> 4 tasks, so Breezing auto-selected

## Effort Level Control (v2.1.68+, v2.1.72 simplified)

Claude Code v2.1.68 defaults Opus 4.6 to **medium effort** (`◐`).
v2.1.72 deprecated the `max` level, simplifying to 3 levels: `low(○)/medium(◐)/high(●)`.
`/effort auto` resets to default.
For complex tasks, use the `ultrathink` keyword to enable high effort (`●`).

### Multi-Factor Scoring

When starting a task, the following scores are summed. **Threshold of 3 or above** triggers ultrathink injection:

| Factor | Condition | Score |
|------|------|--------|
| File count | 4+ files to be changed | +1 |
| Directory | Includes core/, guardrails/, security/ | +1 |
| Keyword | Contains architecture, security, design, migration | +1 |
| Failure history | Agent memory has failure records for the same task | +2 |
| Explicit specification | ultrathink noted in PM template | +3 (auto-adopted) |

### Injection Method

When score >= 3, prepend `ultrathink` to the Worker spawn prompt.
The same logic applies in breezing mode (managed uniformly by harness-work).

## Execution Mode Details

### Solo Mode (auto-selected for 1 task)

1. Read Plans.md and identify the target task
   - **If Plans.md does not exist**: Automatically invoke `harness-plan create --ci` -> Generate Plans.md and continue
   - If the header lacks DoD / Depends columns: `Plans.md is in the legacy format. Please regenerate with harness-plan create.` -> **Stop**
   - **If there are unrecorded tasks in the conversation**: Extract requirements from the most recent conversation context and auto-append to Plans.md with `cc:TODO`
     - Extraction logic: Detect action verbs from user statements ("add ...", "fix ...", "implement ...")
     - Append in v2 format (Task / Description / DoD / Depends / Status)
     - After appending, display "The following were added to Plans.md" (5-second timeout prompt, default: continue)
1.5. **Task Background Check** (30 seconds):
   - Infer and display the **purpose** (the problem this task solves) in one line from the task's "Description" and "DoD"
   - Infer and display the **impact scope** (affected files/modules) using `git grep` / `Glob`
   - If confident in the inference: Proceed directly to implementation (no flow delay)
   - If uncertain: Ask the user one question only ("Does this understanding look correct?")
2. Update task to `cc:WIP`
3. **TDD Phase** (when no `[skip:tdd]` & test framework exists):
   a. Create test files first (Red)
   b. Verify failure
4. Generate `sprint-contract.json` with `scripts/generate-sprint-contract.sh <task-id>`
5. Add Reviewer perspective via `scripts/enrich-sprint-contract.sh`, then confirm approved with `scripts/ensure-sprint-contract-ready.sh`
6. Implement code (Green) (Read/Write/Edit/Bash)
7. Auto-Refinement via `/simplify` (skippable with `--no-simplify`)
8. **Auto-review stage** (see "Review Loop"):
   - Execute review with Codex exec priority -> fallback to internal Reviewer agent
   - If `sprint-contract.json`'s `reviewer_profile` is `runtime`, execute `scripts/run-contract-review-checks.sh`
   - On REQUEST_CHANGES: Fix based on feedback -> re-review (max 3 times)
   - Proceed to next step on APPROVE. Self-check alone does not confirm completion
9. Normalize and save review artifact with `scripts/write-review-result.sh`
10. Auto-commit with `git commit` (skippable with `--no-commit`)
11. Update task to `cc:done` (with commit hash)
   - Get the latest commit hash (short form, 7 characters) with `git log --oneline -1`
   - Update Plans.md Status to `cc:done [a1b2c3d]` format
   - If no commit (`--no-commit`), use `cc:done` without hash
12. **Rich completion report** (see "Completion Report Format")
13. **Automatic re-ticketing on failure** (test/CI failure only):
    - Check test execution results
    - On failure: Save fix task proposal to state and add to Plans.md via approval command (see "Automatic Re-ticketing of Failed Tasks")
    - On success: Proceed to next task

### Parallel Mode (auto-selected for 2-3 tasks / forced with `--parallel N`)

Execute tasks marked with `[P]` using N workers in parallel.
When explicitly specified with `--parallel N`, this mode is used regardless of task count.
If write conflicts occur on the same file, isolate using git worktree.

### Codex Mode (`--codex` explicit only)

Delegate tasks to Codex CLI via the official plugin `codex-plugin-cc` companion.

```bash
# Task delegation (write-enabled)
bash scripts/codex-companion.sh task --write "task description"

# Via stdin (for large prompts)
CODEX_PROMPT=$(mktemp /tmp/codex-prompt-XXXXXX.md)
# Write task content
cat "$CODEX_PROMPT" | bash scripts/codex-companion.sh task --write
rm -f "$CODEX_PROMPT"

# Resume previous thread
bash scripts/codex-companion.sh task --resume-last --write "continue from where we left off"
```

The companion communicates with Codex via the App Server Protocol,
providing job management, thread resume, and structured output.
Results are validated and fixed manually if quality standards are not met.

### Breezing Mode (auto-selected for 4+ tasks / forced with `--breezing`)

Team execution with role separation of Lead / Worker / Reviewer.
In Codex, this assumes native subagent orchestration using `spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`,
not the legacy TeamCreate / TaskCreate-based approach.

**Permission Policy**:
- The current shipped default is `bypassPermissions`
- `--auto-mode` is treated as an opt-in rollout flag for compatible parent sessions
- Do not write the undocumented `autoMode` value to `permissions.defaultMode` or agent frontmatter `permissionMode`

> **CC v2.1.69+**: Nested teammates are prohibited by the platform,
> so do not add redundant nested-prevention wording to Worker/Reviewer prompts.

```
Lead (this agent)
├── Worker (task-worker agent) — Implementation
└── Reviewer (code-reviewer agent) — Review
```

**Phase A: Pre-delegate (Preparation)**:
1. Read Plans.md and identify target tasks
2. Analyze the dependency graph and determine execution order (Depends column)
3. Effort scoring for each task (ultrathink injection determination)
4. Generate `sprint-contract.json` with `scripts/generate-sprint-contract.sh`
5. Add Reviewer perspective via `scripts/enrich-sprint-contract.sh`, stop if unapproved via `scripts/ensure-sprint-contract-ready.sh`

**Phase B: Delegate (Worker spawn -> review -> cherry-pick)**:

Execute the following **sequentially** for each task (dependency order):

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
    # Agent tool return value includes agentId — used for SendMessage in the fix loop
    Plans.md: task.status = "cc:WIP"  # Update on start (unstarted tasks remain cc:TODO)

    worker_result = Agent(
        subagent_type="claude-code-harness:worker",
        prompt="Task: {task.description}\nDoD: {task.DoD}\ncontract_path: {contract_path}\nmode: breezing",
        isolation="worktree",
        run_in_background=false  # Execute in foreground -> wait for Worker completion
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
            pass  # No runtime validation command -> use static verdict as-is
    if profile == "browser":
        # browser artifact generates a PENDING_BROWSER scaffold.
        # Actual browser execution is handled by the reviewer agent subsequently.
        # review-result uses the static review verdict (not PENDING_BROWSER).
        browser_artifact = bash("scripts/generate-browser-review-artifact.sh {contract_path}")
        # browser artifact is saved for reference, but review-result verdict remains static
    # If review_input is DOWNGRADE_TO_STATIC, use the static review result
    if review_input != "review-output.json" and jq(review_input, ".verdict") == "DOWNGRADE_TO_STATIC":
        review_input = "review-output.json"  # Fall back to static review result
    bash("scripts/write-review-result.sh {review_input} {latest_commit}")

    # B-4. Fix loop (on REQUEST_CHANGES, max 3 times)
    # Worker completed in foreground but can be resumed via SendMessage
    # (CC: SendMessage(to: agentId) / Codex: resume_agent(agent_id) + send_input)
    review_count = 0
    latest_commit = worker_result.commit
    while verdict == "REQUEST_CHANGES" and review_count < 3:
        SendMessage(to=worker_id, message="Issues found: {issues}\nPlease fix and amend")
        # Worker fixes -> amends -> returns updated commit hash
        updated_result = wait_for_response(worker_id)
        latest_commit = updated_result.commit
        diff_text = git("-C", worker_result.worktreePath, "show", latest_commit)
        verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
        review_count++

    # B-5. APPROVE -> cherry-pick to main
    if verdict == "APPROVE":
        git cherry-pick --no-commit {latest_commit}  # worktree -> main
        git commit -m "{task.description}"
        Plans.md: task.status = "cc:done [{hash}]"
    else:
        -> Escalate to user

    # B-6. Progress feed
    print("📊 Progress: Task {completed}/{total} done — {task.description}")
```

### Sprint Contract

A `sprint-contract` is a small contract file that defines "what makes this task pass" in a format readable by both machines and humans.
The default save location is `.claude/state/contracts/<task-id>.sprint-contract.json`.

```bash
scripts/generate-sprint-contract.sh 32.1.1
```

The generated output includes:

- `checks`: Verification items decomposed from DoD
- `non_goals`: Out of scope for this iteration
- `runtime_validation`: Validation commands such as test, lint, typecheck
- `browser_validation`: UI flow verification items for the browser reviewer
- `browser_mode`: `scripted` or `exploratory`
- `route`: Whether the browser reviewer uses `playwright` / `agent-browser` / `chrome-devtools`
- `risk_flags`: `needs-spike`, `security-sensitive`, `ux-regression`, etc.
- `reviewer_profile`: `static`, `runtime`, `browser`

**Phase C: Post-delegate (Integration & Reporting)**:
1. Aggregate commit logs from all tasks
2. Output **rich completion report** (Breezing template from "Completion Report Format")
3. Final verification of Plans.md (confirm all tasks are cc:done)

## CI Failure Handling

When CI fails:

1. Check logs to identify the error
2. Implement the fix
3. If the same root cause fails 3 times, stop the auto-fix loop
4. Summarize failure logs, attempted fixes, and remaining issues for escalation

## Automatic Re-ticketing of Failed Tasks

When tests/CI fail after task completion, auto-generate fix task proposals and reflect them in Plans.md after approval:

### Trigger Conditions

| Condition | Action |
|------|----------|
| Test failure after `cc:done` | Save fix task proposal to state and wait for approval |
| CI failure (fewer than 3 times) | Implement fix and increment failure count |
| CI failure (3rd time) | Present fix task proposal + escalate |

### Auto-generation of Fix Tasks

1. Classify failure cause (syntax_error / import_error / type_error / assertion_error / timeout / runtime_error)
2. Save fix task proposal to `.claude/state/pending-fix-proposals.jsonl`:
   - Number: Original task number + `.fix` suffix (e.g., `26.1.fix`)
   - Description: `fix: [original task name] - [failure cause category]`
   - DoD: Tests/CI pass
   - Depends: Original task number
3. When user sends `approve fix <task_id>`, add to Plans.md with `cc:TODO`
4. `reject fix <task_id>` discards the proposal. When only 1 pending item exists, `yes` / `no` responses also work

## Review Loop

A quality verification stage that runs automatically after implementation (after step 5).
Applied **uniformly across all modes** (Solo / Parallel / Breezing).
In Parallel mode, each Worker executes the same loop as step 10 (external review acceptance).

### Review Execution Priority

```
1. Codex exec (preferred)
   ↓ codex command not found or timeout (120s)
2. Internal Reviewer agent (fallback)
```

### APPROVE / REQUEST_CHANGES Judgment Criteria

The following threshold criteria are passed to the reviewer, and the verdict is determined **solely by these criteria**.
Improvement suggestions outside these criteria are returned as `recommendations` but do not affect the verdict.

| Severity | Definition | Verdict Impact |
|--------|------|-----------------|
| **critical** | Security vulnerability, data loss risk, potential production outage | 1 item -> REQUEST_CHANGES |
| **major** | Breaking existing functionality, clear specification contradiction, failing tests | 1 item -> REQUEST_CHANGES |
| **minor** | Naming improvements, insufficient comments, style inconsistency | Does not affect verdict |
| **recommendation** | Best practice suggestions, future improvement proposals | Does not affect verdict |

> **Important**: When only minor / recommendation items exist, **always return APPROVE**.
> "Nice-to-have improvements" are not grounds for REQUEST_CHANGES.

### Codex Exec Review (via official plugin)

The HEAD at task start is retained as `BASE_REF`, and the diff from that ref is the review target.
Uses the official plugin `codex-plugin-cc` companion review.

```bash
# Record base ref at task start (execute before cc:WIP update in Step 2)
BASE_REF=$(git rev-parse HEAD)

# ... after implementation completes ...

# Execute structured review via official plugin
bash scripts/codex-companion.sh review --base "${BASE_REF}"
REVIEW_EXIT=$?
```

**Verdict mapping** (official plugin -> Harness format):

The official plugin returns structured output conforming to `review-output.schema.json`.
Conversion rules to Harness verdict format:

| Official Plugin | Harness | Verdict Impact |
|---|---|---|
| `approve` | `APPROVE` | - |
| `needs-attention` | `REQUEST_CHANGES` | - |
| `findings[].severity: critical` | `critical_issues[]` | 1 item -> REQUEST_CHANGES |
| `findings[].severity: high` | `major_issues[]` | 1 item -> REQUEST_CHANGES |
| `findings[].severity: medium/low` | `recommendations[]` | Does not affect verdict |

AI Residuals scanning continues via `scripts/review-ai-residuals.sh`,
and the final verdict is determined by combining it with the companion review results.

```bash
# AI Residuals scan (can run in parallel with companion review)
AI_RESIDUALS_JSON="$(bash scripts/review-ai-residuals.sh --base-ref "${BASE_REF}" 2>/dev/null || echo '{"tool":"review-ai-residuals","scan_mode":"diff","base_ref":null,"files_scanned":[],"summary":{"verdict":"APPROVE","major":0,"minor":0,"recommendation":0,"total":0},"observations":[]}')"
```

### Internal Reviewer Agent Fallback

When Codex exec is unavailable (`command -v codex` fails, or exit code != 0):

```
Agent tool: subagent_type="reviewer"
prompt: "Please review the following changes. Criteria: critical/major -> REQUEST_CHANGES, minor/recommendation only -> APPROVE. diff: {git diff ${BASE_REF}}"
```

The Reviewer agent executes reviews safely in read-only mode (Write/Edit/Bash disabled).

### Fix Loop (on REQUEST_CHANGES)

```
review_count = 0
MAX_REVIEWS = 3

while verdict == "REQUEST_CHANGES" and review_count < MAX_REVIEWS:
    1. Parse review feedback (critical / major only)
    2. Implement fixes for each issue
    3. Re-run review (same criteria, same priority)
    review_count++

if review_count >= MAX_REVIEWS and verdict != "APPROVE":
    -> Escalate to user
    -> "Fixed 3 times but the following critical/major issues remain" + display issue list
    -> Wait for user decision (continue / abort)
```

### Application in Breezing Mode

In Breezing mode, the **Lead** executes the review loop (see Phase B above):

1. Worker implements and commits in worktree -> Returns result to Lead
2. Lead reviews with Codex exec (preferred) / Reviewer agent (fallback)
3. REQUEST_CHANGES -> Lead sends fix instructions to Worker via SendMessage -> Worker amends
4. After fix, re-review (max 3 times)
5. APPROVE -> Lead cherry-picks to main -> Updates Plans.md to `cc:done [{hash}]`

## Completion Report Format

A visual summary automatically output when a task completes (`cc:done` + commit).
Designed so that non-technical stakeholders can understand the changes and their impact.

### Template

```
┌─────────────────────────────────────────────┐
│  ✓ Task {N} done: {task name}                   │
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
│  ■ Remaining items                           │
│    • Task {X} ({status}): {description}  <- Plans.md  │
│    • Task {Y} ({status}): {description}  <- Plans.md  │
│    ({M} incomplete tasks in Plans.md)        │
│                                              │
│  commit: {hash} | review: {APPROVE}          │
└─────────────────────────────────────────────┘
```

### Generation Rules

1. **What was done**: Auto-extracted from `git diff --stat HEAD~1` and commit message. Minimize technical jargon, start with verbs
2. **What changed**: Infer Before/After from the task's "Description" and "DoD". Emphasize user experience changes
3. **Changed files**: Obtained from `git diff --name-only HEAD~1`. If more than 5 files, truncate and show count
4. **Remaining items**: List `cc:TODO` / `cc:WIP` tasks from Plans.md. Indicate whether already tracked in Plans.md
5. **review**: Display review result (APPROVE / REQUEST_CHANGES -> APPROVE)

### Parallel Mode Reporting

- **1 task** (when `--parallel` forced): Use Solo template
- **Multiple tasks**: Use Breezing aggregate template (see below)

### Breezing Mode Reporting

Output collectively after all tasks complete. Each task is listed in abbreviated form (what was done + commit hash only),
followed by an overall summary (total changed files + remaining items):

```
┌─────────────────────────────────────────────┐
│  ✓ Breezing complete: {N}/{M} tasks          │
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
│  ■ Remaining items                           │
│    {K} incomplete tasks in Plans.md          │
│    • Task {X}: {description}                 │
│                                              │
└─────────────────────────────────────────────┘
```

## Related Skills

- `harness-plan` — Plan the tasks to execute
- `harness-sync` — Sync implementation with Plans.md
- `harness-review` — Review implementation
- `harness-release` — Version bump and release
