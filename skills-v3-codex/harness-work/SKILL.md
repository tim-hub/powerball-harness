---
name: harness-work
description: "Use this skill whenever the user asks to implement, execute, build, 'do everything', or run tasks in the Codex environment. Do NOT load for: planning, code review, release, or setup. Unified execution skill for Harness v3 (Codex native) — implements Plans.md tasks from single task to full parallel team runs."
argument-hint: "[all] [task-number|range] [--parallel N] [--no-commit] [--breezing]"
effort: high
---

# Harness Work (v3) — Codex Native

> **This SKILL.md is the Codex CLI native version.**
> For the Claude Code version, see `skills-v3/harness-work/SKILL.md`.
> The subagent API uses Codex's `spawn_agent` / `send_input` / `wait_agent` / `close_agent`.

Unified execution skill for Harness v3.

## Quick Reference

| User Input | Mode | Behavior |
|------------|--------|------|
| `harness-work` | **solo** | Execute the next incomplete task |
| `harness-work all` | **sequential** | Execute all incomplete tasks sequentially |
| `harness-work 3` | solo | Execute only task 3 immediately |
| `harness-work --parallel 3` | parallel | Execute 3 tasks in parallel using companion `task` (Bash `&` + `wait`) |
| `harness-work --breezing` | breezing | Team execution via `spawn_agent` (only when explicitly specified) |

## Execution Mode Selection

> **Important**: In Codex, `spawn_agent` is used only when the user explicitly requests team execution or parallel work.
> Do not auto-escalate based solely on the number of tasks.

| Condition | Mode | Rationale |
|------|--------|------|
| No arguments / single task specified | **Solo** | Direct implementation is fastest |
| `all` / range specified (no flags) | **Sequential** | Safe sequential processing |
| `--parallel N` | **Parallel** | Bash parallel via companion `task` (only when explicitly specified) |
| `--breezing` | **Breezing** | `spawn_agent` team execution (only when explicitly specified) |

### Rules

1. **Explicit flags always override defaults**
2. **`--breezing` and `--parallel` activate only when explicitly specified**. No auto-escalation based on task count
3. `--parallel` and `--breezing` are mutually exclusive (cannot be specified together)

## Options

| Option | Description | Default |
|----------|------|----------|
| `all` | Target all incomplete tasks | - |
| `N` or `N-M` | Task number/range specification | - |
| `--parallel N` | Companion `task` Bash parallelism count | - |
| `--sequential` | Force sequential execution | - |
| `--no-commit` | Suppress final commit to main (Solo/Sequential only. Not supported in Breezing/Parallel) | false |
| `--breezing` | Team execution with Lead/Worker/Reviewer | false |
| `--no-tdd` | Skip TDD phase | false |

## Scope Dialog (when no arguments provided)

```
harness-work
How far should we go?
1) Next task: Execute the next incomplete task from Plans.md in Solo mode
2) All: Execute all remaining tasks sequentially
3) Specify number: Enter task number (e.g., 3, 5-7)
```

If arguments are provided, execution starts immediately (dialog is skipped).

## Execution Mode Details

### Solo Mode

1. Read Plans.md and identify the target task
   - **If Plans.md does not exist**: Automatically invoke `harness-plan create --ci` to generate Plans.md and continue
   - If the header is missing DoD / Depends columns: Stop
   - **If there are unrecorded tasks in the conversation**: Extract requirements from the preceding conversation context and auto-append to Plans.md as `cc:TODO`
1.5. **Task context verification** (30 seconds):
   - Infer and display the task purpose in one line based on the task's "content" and "DoD"
   - If confident in the inference: Proceed directly to implementation
   - If not confident: Ask the user one clarifying question
2. Update task to `cc:WIP`. Record `TASK_BASE_REF=$(git rev-parse HEAD)`
3. **TDD phase** (when `[skip:tdd]` is absent & test framework exists):
   a. Create test files first (Red)
   b. Verify failure
4. Implement the code (Green)
5. Auto-commit with `git commit` (can be skipped with `--no-commit`)
6. **Auto-review stage** (see "Review Loop") — Review the diff from TASK_BASE_REF..HEAD
7. Update task to `cc:Done [hash]`
8. **Rich completion report** (see "Completion Report Format")
9. **Auto re-planning on failure** (only on test/CI failure)

### Sequential Mode (default when `all` is specified)

Process Plans.md tasks one at a time in dependency order using Solo mode.
Update Plans.md after each task completes, then proceed to the next task.

### Parallel Mode (only when `--parallel N` is explicitly specified)

Execute independent tasks in parallel using Bash `&` + `wait`.
Uses the official plugin's companion `task` with worktree isolation per Worker.

> **Constraint**: Do not parallelize tasks that may modify the same file.
> Use `git worktree add` to isolate working directories per Worker; Lead reviews and cherry-picks afterward.

```bash
# Isolate each Worker with a worktree (note the order: -b <branch> <path>)
git worktree add -b worker-a-$$ /tmp/worker-a-$$
git worktree add -b worker-b-$$ /tmp/worker-b-$$

# Task A (cd into the worktree to switch cwd, then call companion)
PROMPT_A=$(mktemp /tmp/codex-prompt-XXXXXX.md)
cat > "$PROMPT_A" << EOF
Task A content...

Upon completion, output the following JSON to stdout:
{"commit": "<hash>", "files_changed": ["path1", "path2"], "summary": "..."}
EOF
(cd /tmp/worker-a-$$ && cat "$PROMPT_A" | bash "${PROJECT_ROOT}/scripts/codex-companion.sh" task --write) > /tmp/out-a-$$.json 2>>/tmp/harness-codex-$$.log &

# Task B (cd into the worktree to switch cwd, then call companion)
PROMPT_B=$(mktemp /tmp/codex-prompt-XXXXXX.md)
cat > "$PROMPT_B" << EOF
Task B content...

Upon completion, output the following JSON to stdout:
{"commit": "<hash>", "files_changed": ["path1", "path2"], "summary": "..."}
EOF
(cd /tmp/worker-b-$$ && cat "$PROMPT_B" | bash "${PROJECT_ROOT}/scripts/codex-companion.sh" task --write) > /tmp/out-b-$$.json 2>>/tmp/harness-codex-$$.log &

wait
rm -f "$PROMPT_A" "$PROMPT_B"

# Lead retrieves commit hashes from each Worker's output JSON, reviews individually, then cherry-picks
# ... review and cherry-pick processing ...

# Remove worktrees
git worktree remove /tmp/worker-a-$$
git worktree remove /tmp/worker-b-$$
```

### Breezing Mode (only when `--breezing` is explicitly specified)

Team execution with Lead / Worker / Reviewer role separation.
Uses Codex's native subagent API.

> **`--breezing` is only activated when explicitly specified**. Use only when the user instructs "team execution" or "breezing".

```
Lead (this agent)
├── Worker (spawn_agent) — Implementation
│   Each Worker operates in an isolated working directory via git worktree
└── Reviewer (companion review --base) — Review
```

**Phase A: Pre-delegate (Preparation)**:
1. Read Plans.md and identify target tasks
2. Analyze the dependency graph and determine execution order (Depends column)
3. Create a git worktree for each task
4. Generate `sprint-contract.json` with `scripts/generate-sprint-contract.sh`
5. Enrich with Reviewer perspective via `scripts/enrich-sprint-contract.sh`, then halt if unapproved via `scripts/ensure-sprint-contract-ready.sh`

**Phase B: Delegate (Worker spawn -> review -> cherry-pick)**:

Execute the following **sequentially** for each task (in dependency order):

```
for task in execution_order:
    # B-0. Working directory isolation
    worktree_path = "/tmp/worker-{task.number}-$$"
    branch_name = "worker-{task.number}-$$"
    git worktree add -b {branch_name} {worktree_path}
    TASK_BASE_REF = git rev-parse HEAD  # Base ref specific to this task

    # B-1. Generate sprint-contract
    contract_path = bash("scripts/generate-sprint-contract.sh {task.number}")
    contract_path = bash("scripts/enrich-sprint-contract.sh {contract_path} --check \"Verify DoD from reviewer perspective\" --approve")
    bash("scripts/ensure-sprint-contract-ready.sh {contract_path}")

    # B-2. Worker spawn (Codex native subagent)
    Plans.md: task.status = "cc:WIP"

    worker_id = spawn_agent({
        message: "Working directory: {worktree_path}. Please work here.\n\nTask: {task.content}\nDoD: {task.DoD}\ncontract_path: {contract_path}\n\nPlease implement. Run git commit when done.\n\nUpon completion, return the following JSON:\n{\"commit\": \"<hash>\", \"files_changed\": [\"path1\"], \"summary\": \"...\"}",
        fork_context: true
    })
    wait_agent({ ids: [worker_id] })
    # Retrieve commit hash, files_changed, and summary from Worker output

    # B-3. Lead executes review (companion review --base TASK_BASE_REF)
    # Review only this task's diff (starting from TASK_BASE_REF)
    # Uses the official plugin's structured review
    verdict = companion_review(TASK_BASE_REF)  # static review (see "Review Loop" for details)
    profile = jq(contract_path, ".review.reviewer_profile")
    browser_mode = jq(contract_path, ".review.browser_mode // \"scripted\"")
    review_input = "review-output.json"
    if profile == "runtime":
        # Execute runtime checks within the worktree (against Worker's artifacts, not Lead's cwd)
        review_input = bash("cd {worktree_path} && scripts/run-contract-review-checks.sh {contract_path}")
        runtime_verdict = jq(review_input, ".verdict")
        if runtime_verdict == "REQUEST_CHANGES":
            verdict = "REQUEST_CHANGES"
        elif runtime_verdict == "DOWNGRADE_TO_STATIC":
            # No runtime verification commands -> fall back to static review result
            review_input = "review-output.json"
    if profile == "browser":
        # Browser artifact is a PENDING_BROWSER scaffold. Actual browser execution is handled by the reviewer agent later.
        # Write the static review verdict to review-result (not PENDING_BROWSER).
        browser_artifact = bash("scripts/generate-browser-review-artifact.sh {contract_path}")
        # review_input remains as the static review
    # Ensure DOWNGRADE_TO_STATIC is not left in review_input
    if review_input != "review-output.json" and jq(review_input, ".verdict") == "DOWNGRADE_TO_STATIC":
        review_input = "review-output.json"
    bash("scripts/write-review-result.sh {review_input} {commit_hash}")

    # B-4. Fix loop (on REQUEST_CHANGES, up to 3 times)
    review_count = 0
    while verdict == "REQUEST_CHANGES" and review_count < 3:
        # Worker is still alive (not closed), so send instructions directly via send_input
        send_input({
            id: worker_id,
            message: "Review findings: {issues}\nPlease fix and run git commit --amend. Output the JSON again after fixing."
        })
        wait_agent({ ids: [worker_id] })
        # Re-review (diff from TASK_BASE_REF)
        diff_text = git("-C", worktree_path, "diff", TASK_BASE_REF, "HEAD")
        verdict = codex_exec_review(diff_text)
        review_count++

    close_agent({ id: worker_id })

    # B-5. Result processing
    if verdict == "APPROVE":
        # Cherry-pick the worktree commit to main
        commit_hash = git("-C", worktree_path, "rev-parse", "HEAD")
        git cherry-pick --no-commit {commit_hash}
        git commit -m "{task.content}"
        Plans.md: task.status = "cc:Done [{short_hash}]"
    else:
        -> Escalate to user (Plans.md remains cc:WIP)
        # Skip B-5 onward; also stop subsequent tasks

    # B-6. Worktree cleanup
    git worktree remove {worktree_path}
    git branch -D {branch_name}

    # B-7. Progress feed
    print("📊 Progress: Task {completed}/{total} done — {task.content}")
```

**Phase C: Post-delegate (Integration and Reporting)**:
1. Aggregate commit logs from all tasks
2. Output the **rich completion report**
3. Final verification of Plans.md (confirm all tasks are cc:Done)

## CI Failure Handling

1. Check logs and identify the error
2. Implement the fix
3. Stop the auto-fix loop if the same root cause fails 3 times
4. Escalate with a summary of failure logs, attempted fixes, and remaining issues

## Auto Re-ticketing of Failed Tasks

When tests/CI fail after task completion, auto-generate a fix task proposal and reflect it in Plans.md upon approval.

| Condition | Action |
|------|----------|
| Test failure after `cc:Done` | Present fix task proposal and wait for approval |
| CI failure (fewer than 3 times) | Implement the fix |
| CI failure (3rd time) | Present fix task proposal + escalate |

## Review Loop

Quality verification stage that runs automatically after implementation.
Applied uniformly across **all modes** (Solo / Sequential / Parallel / Breezing).

### Review Execution (via official plugin companion)

Uses the official plugin `codex-plugin-cc` companion review.
Obtains the verdict via structured output (conforming to `review-output.schema.json`).

> **Diff starting point**: Uses the task-specific `TASK_BASE_REF` (HEAD at task start).
> Reviews only that task's changes, not cumulative diffs.

```bash
# Record base ref at task start (execute before cc:WIP update)
TASK_BASE_REF=$(git rev-parse HEAD)

# ... after implementation ...

# Execute structured review via the official plugin
bash scripts/codex-companion.sh review --base "${TASK_BASE_REF}"
REVIEW_EXIT=$?
```

**Verdict mapping** (official plugin -> Harness format):

| Official Plugin | Harness | Verdict Impact |
|---|---|---|
| `approve` | `APPROVE` | - |
| `needs-attention` | `REQUEST_CHANGES` | - |
| `findings[].severity: critical` | `critical_issues[]` | Even 1 -> REQUEST_CHANGES |
| `findings[].severity: high` | `major_issues[]` | Even 1 -> REQUEST_CHANGES |
| `findings[].severity: medium/low` | `recommendations[]` | Does not affect verdict |

Determine the verdict from companion review output:
- `verdict` is `approve` -> `APPROVE`
- `verdict` is `needs-attention` -> `REQUEST_CHANGES`
- `findings` contain `critical` / `high` severity -> `REQUEST_CHANGES`

### APPROVE / REQUEST_CHANGES Criteria

| Severity | Definition | Verdict Impact |
|--------|------|-----------------|
| **critical** | Security vulnerabilities, data loss risk, potential production outages | Even 1 -> REQUEST_CHANGES |
| **major** | Breaking existing functionality, clear spec contradictions, test failures | Even 1 -> REQUEST_CHANGES |
| **minor** | Naming improvements, missing comments, style inconsistencies | Does not affect verdict |
| **recommendation** | Best practice suggestions, future improvement proposals | Does not affect verdict |

> **Important**: If only minor / recommendation findings exist, **always return APPROVE**.

### Fix Loop (on REQUEST_CHANGES)

```
review_count = 0
MAX_REVIEWS = 3

while verdict == "REQUEST_CHANGES" and review_count < MAX_REVIEWS:
    1. Analyze review findings (only critical / major)
    2. Implement fixes for each finding
    3. git commit --amend
    4. Re-run companion review (from TASK_BASE_REF)
    review_count++

if review_count >= MAX_REVIEWS and verdict != "APPROVE":
    -> Escalate to user
    -> "Fixed 3 times but the following critical/major findings remain" + display findings list
    -> Wait for user decision (continue / abort)
```

### Application in Breezing Mode

1. Worker implements and commits in the worktree -> Wait for completion via `wait_agent`
2. Lead reviews via companion review (from TASK_BASE_REF)
3. REQUEST_CHANGES -> Send fix instructions to Worker via `send_input` -> Worker amends
4. Re-review after fix (up to 3 times)
5. Terminate Worker via `close_agent`
6. APPROVE -> Lead cherry-picks to main -> Update Plans.md to `cc:Done [{hash}]`

### Worker Output Contract

The Worker prompt explicitly requires returning the following JSON upon completion:

```json
{
  "commit": "a1b2c3d",
  "files_changed": ["src/foo.ts", "tests/foo.test.ts"],
  "summary": "Added bar feature to foo module"
}
```

Lead parses this JSON to obtain the commit hash and file list.
If the Worker does not return JSON, retrieve the latest commit via `git log --oneline -1`.

## Completion Report Format

Visual summary automatically output upon task completion.

### Solo Template

```
┌─────────────────────────────────────────────┐
│  ✓ Task {N} Done: {task name}                   │
├─────────────────────────────────────────────┤
│  ■ What was done                              │
│    • {change 1}                               │
│    • {change 2}                               │
│  ■ What changed                               │
│    Before: {old behavior}                     │
│    After:  {new behavior}                     │
│  ■ Changed files ({N} files)                  │
│    {file path 1}                              │
│  ■ Remaining issues                           │
│    {M} incomplete tasks remain in Plans.md    │
│  commit: {hash} | review: {APPROVE}           │
└─────────────────────────────────────────────┘
```

### Breezing Template

```
┌─────────────────────────────────────────────┐
│  ✓ Breezing Done: {N}/{M} tasks              │
├─────────────────────────────────────────────┤
│  1. ✓ {task name 1}            [{hash1}]      │
│  2. ✓ {task name 2}            [{hash2}]      │
│  ■ Overall changes                            │
│    {N} files changed, {A} insertions(+),     │
│    {D} deletions(-)                          │
│  ■ Remaining issues                           │
│    {K} incomplete tasks remain in Plans.md    │
└─────────────────────────────────────────────┘
```

## Differences from the Claude Code Version

| Item | Claude Code Version | Codex Native Version (this file) |
|------|---------------|-------------------------------|
| Worker spawn | `Agent(subagent_type="worker")` | `spawn_agent({message, fork_context})` |
| Completion wait | `Agent` return value | `wait_agent({ids: [id]})` |
| Fix instructions | `SendMessage(to: agentId)` | `send_input({id, message})` |
| Worker termination | Automatic (Agent tool return value) | Explicit termination via `close_agent({id})` |
| Worktree isolation | `isolation="worktree"` auto-managed | Manual isolation via `git worktree add` |
| Permissions | `bypassPermissions` | companion `task --write` / `spawn_agent`: session permission inheritance |
| Review | Codex exec -> Reviewer agent fallback | companion `review --base` (structured output) |
| Verdict retrieval | Parse Agent response | companion review verdict field (approve/needs-attention) |
| Mode auto-escalation | Auto-determined by task count | Explicit flags only (no auto-escalation) |
| Effort control | `ultrathink` + `/effort` | `model_reasoning_effort` in config.toml |
| Auto-Refinement | `/simplify` | None |

## Related Skills

- `harness-plan` — Plan tasks for execution
- `harness-sync` — Sync implementation with Plans.md
- `harness-review` — Code review for implementation
- `harness-release` — Version bump and release
