---
name: harness-work
description: "Unified execution skill for Harness v3 (Codex native). Handles Plans.md tasks from single task to full parallel team runs. Triggers: implement, run, harness-work, do everything, breezing, team execution, parallel. Not for planning, review, release, or setup."
description-en: "Unified execution skill for Harness v3 (Codex native). Implements Plans.md tasks from single task to full parallel team runs."
argument-hint: "[all] [task-number|range] [--parallel N] [--no-commit] [--breezing]"
effort: high
---

# Harness Work — Codex Native

> **This SKILL.md is the Codex CLI native version.**
> For the Claude Code version, see `skills/harness-work/SKILL.md`.
> Subagent APIs use Codex's `spawn_agent` / `send_input` / `wait_agent` / `close_agent`.

Unified execution skill for Harness v3.

## Quick Reference

| User Input | Mode | Behavior |
|------------|--------|------|
| `harness-work` | **solo** | Execute the next incomplete task |
| `harness-work all` | **sequential** | Execute all incomplete tasks sequentially |
| `harness-work 3` | solo | Execute task 3 immediately |
| `harness-work --parallel 3` | parallel | Run 3 tasks in parallel via companion `task` (Bash `&` + `wait`) |
| `harness-work --breezing` | breezing | Team execution via `spawn_agent` (explicit only) |

## Execution Mode Selection

> **Important**: In Codex, `spawn_agent` is used only when the user explicitly requests team execution or parallel work.
> Do not auto-promote based solely on task count.

| Condition | Mode | Reason |
|------|--------|------|
| No args / single task specified | **Solo** | Direct implementation is fastest |
| `all` / range specified (no flags) | **Sequential** | Safe sequential processing |
| `--parallel N` | **Parallel** | Bash parallel via companion `task` (explicit only) |
| `--breezing` | **Breezing** | `spawn_agent` team execution (explicit only) |

### Rules

1. **Explicit flags always override defaults**
2. **`--breezing` and `--parallel` activate only when explicitly specified**. No auto-promotion based on task count
3. `--parallel` and `--breezing` are mutually exclusive (cannot be specified together)

## Options

| Option | Description | Default |
|----------|------|----------|
| `all` | Target all incomplete tasks | - |
| `N` or `N-M` | Task number / range specification | - |
| `--parallel N` | companion `task` Bash parallel count | - |
| `--sequential` | Force sequential execution | - |
| `--no-commit` | Suppress final commit to main (Solo/Sequential only. Not supported for Breezing/Parallel) | false |
| `--breezing` | Team execution with Lead/Worker/Reviewer | false |
| `--no-tdd` | Skip TDD phase | false |

## Scope Dialog (when no arguments provided)

```
harness-work
How far do you want to go?
1) Next task: the next incomplete task in Plans.md → execute in Solo mode
2) All: execute all remaining tasks sequentially
3) Specify number: enter a task number (e.g., 3, 5-7)
```

If arguments are provided, execute immediately (skip dialog).

## Execution Mode Details

### Solo Mode

1. Load Plans.md and identify the target task
   - **If Plans.md does not exist**: auto-call `harness-plan create --ci` → generate Plans.md and continue
   - If the header lacks DoD / Depends columns: stop
   - **If there are tasks not in the conversation**: extract requirements from the preceding conversation context and auto-append to Plans.md as `cc:TODO`
1.5. **Task background check** (30 seconds):
   - Infer and display the purpose of the task from its "content" and "DoD" in one line
   - If confident in the inference: proceed directly to implementation
   - If not confident: ask the user exactly one clarifying question
2. Update task to `cc:WIP`. Record `TASK_BASE_REF=$(git rev-parse HEAD)`
3. **TDD Phase** (when `[skip:tdd]` is absent & test framework exists):
   a. Create test file first (Red)
   b. Confirm failure
4. Implement code (Green)
5. Auto-commit with `git commit` (can be skipped with `--no-commit`)
6. **Auto-review stage** (see "Review Loop") — review diff from TASK_BASE_REF..HEAD
7. Update task to `cc:Done [hash]`
8. **Rich completion report** (see "Completion Report Format")
9. **Auto re-ticketing on failure** (only on test/CI failure)

### Sequential Mode (default when `all` is specified)

Process tasks in Plans.md one by one in dependency order using Solo mode.
Update Plans.md after each task completes and move to the next.

### Parallel Mode (only when `--parallel N` is explicitly specified)

Execute independent tasks in parallel using Bash `&` + `wait`.
Uses the official plugin's companion `task`, isolating each Worker in a worktree.

> **Constraint**: Do not parallelize tasks that may modify the same files.
> Use `git worktree add` to separate work directories per Worker, with Lead cherry-picking after review.

```bash
# Separate worktree per Worker (note the order: -b <branch> <path>)
git worktree add -b worker-a-$$ /tmp/worker-a-$$
git worktree add -b worker-b-$$ /tmp/worker-b-$$

# Task A (switch cwd to worktree via cd before calling companion)
PROMPT_A=$(mktemp /tmp/codex-prompt-XXXXXX.md)
cat > "$PROMPT_A" << EOF
Content of task A...

When done, output the following JSON to stdout:
{"commit": "<hash>", "files_changed": ["path1", "path2"], "summary": "..."}
EOF
(cd /tmp/worker-a-$$ && cat "$PROMPT_A" | bash "${PROJECT_ROOT}/scripts/codex-companion.sh" task --write) > /tmp/out-a-$$.json 2>>/tmp/harness-codex-$$.log &

# Task B (switch cwd to worktree via cd before calling companion)
PROMPT_B=$(mktemp /tmp/codex-prompt-XXXXXX.md)
cat > "$PROMPT_B" << EOF
Content of task B...

When done, output the following JSON to stdout:
{"commit": "<hash>", "files_changed": ["path1", "path2"], "summary": "..."}
EOF
(cd /tmp/worker-b-$$ && cat "$PROMPT_B" | bash "${PROJECT_ROOT}/scripts/codex-companion.sh" task --write) > /tmp/out-b-$$.json 2>>/tmp/harness-codex-$$.log &

wait
rm -f "$PROMPT_A" "$PROMPT_B"

# Lead gets commit hash from each Worker's output JSON, reviews individually → cherry-pick
# ... review / cherry-pick processing ...

# Remove worktrees
git worktree remove /tmp/worker-a-$$
git worktree remove /tmp/worker-b-$$
```

### Breezing Mode (only when `--breezing` is explicitly specified)

Team execution with role separation: Lead / Worker / Reviewer.
Uses Codex's native subagent API.

> **`--breezing` is explicit-only**. Use only when the user instructs "with team execution" or "with breezing".

```
Lead (this agent)
├── Worker (spawn_agent) — implementation
│   Each Worker operates in a git worktree-isolated work directory
└── Reviewer (companion review --base) — review
```

**Phase A: Pre-delegate (preparation)**:
1. Load Plans.md and identify target tasks
2. Analyze dependency graph and determine execution order (Depends column)
3. Create git worktree for each task
4. Generate `sprint-contract.json` with `scripts/generate-sprint-contract.sh`
5. Enrich with Reviewer perspective via `scripts/enrich-sprint-contract.sh`; stop if not approved via `scripts/ensure-sprint-contract-ready.sh`

**Phase B: Delegate (Worker spawn → review → cherry-pick)**:

Execute the following **sequentially** for each task (in dependency order):

```
for task in execution_order:
    # B-0. Isolate work directory
    worktree_path = "/tmp/worker-{task.number}-$$"
    branch_name = "worker-{task.number}-$$"
    git worktree add -b {branch_name} {worktree_path}
    TASK_BASE_REF = git rev-parse HEAD  # base ref specific to this task

    # B-1. Generate sprint-contract
    contract_path = bash("scripts/generate-sprint-contract.sh {task.number}")
    contract_path = bash("scripts/enrich-sprint-contract.sh {contract_path} --check \"Verify DoD from reviewer perspective\" --approve")
    bash("scripts/ensure-sprint-contract-ready.sh {contract_path}")

    # B-2. Worker spawn (Codex native subagent)
    Plans.md: task.status = "cc:WIP"

    worker_id = spawn_agent({
        message: "Work in directory: {worktree_path}.\n\nTask: {task.content}\nDoD: {task.DoD}\ncontract_path: {contract_path}\n\nPlease implement. When done, git commit.\n\nWhen complete, return the following JSON:\n{\"commit\": \"<hash>\", \"files_changed\": [\"path1\"], \"summary\": \"...\"}",
        fork_context: true
    })
    wait_agent({ ids: [worker_id] })
    # Get commit hash, files_changed, summary from Worker output

    # B-3. Lead runs review (companion review --base TASK_BASE_REF)
    # Review only the diff specific to this task (from TASK_BASE_REF)
    # Use the official plugin's structured review
    verdict = companion_review(TASK_BASE_REF)  # static review (see "Review Loop" for details)
    profile = jq(contract_path, ".review.reviewer_profile")
    browser_mode = jq(contract_path, ".review.browser_mode // \"scripted\"")
    review_input = "review-output.json"
    if profile == "runtime":
        # Run runtime checks inside worktree (against Worker's artifacts, not Lead's cwd)
        review_input = bash("cd {worktree_path} && scripts/run-contract-review-checks.sh {contract_path}")
        runtime_verdict = jq(review_input, ".verdict")
        if runtime_verdict == "REQUEST_CHANGES":
            verdict = "REQUEST_CHANGES"
        elif runtime_verdict == "DOWNGRADE_TO_STATIC":
            # No runtime verification commands → fall back to static review result
            review_input = "review-output.json"
    if profile == "browser":
        # browser artifact is a PENDING_BROWSER scaffold. Actual browser execution is handled by reviewer agent later.
        # Write static review verdict to review-result (not PENDING_BROWSER).
        browser_artifact = bash("scripts/generate-browser-review-artifact.sh {contract_path}")
        # review_input stays as static review
    # Confirm DOWNGRADE_TO_STATIC does not remain in review_input
    if review_input != "review-output.json" and jq(review_input, ".verdict") == "DOWNGRADE_TO_STATIC":
        review_input = "review-output.json"
    bash("scripts/write-review-result.sh {review_input} {commit_hash}")

    # B-4. Fix loop (on REQUEST_CHANGES, max 3 times)
    review_count = 0
    while verdict == "REQUEST_CHANGES" and review_count < 3:
        # Worker is done but not closed, so send_input can be used to instruct directly
        send_input({
            id: worker_id,
            message: "Issues found: {issues}\nFix them and run git commit --amend. Output JSON again after fixing."
        })
        wait_agent({ ids: [worker_id] })
        # Re-review (diff from TASK_BASE_REF)
        diff_text = git("-C", worktree_path, "diff", TASK_BASE_REF, "HEAD")
        verdict = codex_exec_review(diff_text)
        review_count++

    close_agent({ id: worker_id })

    # B-5. Result handling
    if verdict == "APPROVE":
        # Cherry-pick worktree commit into main
        commit_hash = git("-C", worktree_path, "rev-parse", "HEAD")
        git cherry-pick --no-commit {commit_hash}
        git commit -m "{task.content}"
        Plans.md: task.status = "cc:Done [{short_hash}]"
    else:
        → Escalate to user (Plans.md stays as cc:WIP)
        # Skip B-5 onward, also stop subsequent tasks

    # B-6. Worktree cleanup
    git worktree remove {worktree_path}
    git branch -D {branch_name}

    # B-7. Progress feed
    print("📊 Progress: Task {completed}/{total} done — {task.content}")
```

**Phase C: Post-delegate (integration & report)**:
1. Aggregate commit log for all tasks
2. Output **rich completion report**
3. Final Plans.md check (verify all tasks are cc:Done)

## Handling CI Failures

1. Check logs and identify the error
2. Apply fixes
3. If the same cause fails 3 times, stop the auto-fix loop
4. Summarize the failure log, attempted fixes, and remaining issues for escalation

## Auto Re-Ticketing of Failed Tasks

When tests/CI fail after a task completes, auto-generate fix task proposals and reflect them in Plans.md after approval.

| Condition | Action |
|------|----------|
| Tests fail after `cc:Done` | Present fix task proposal and wait for approval |
| CI failure (fewer than 3 times) | Apply fixes |
| CI failure (3rd time) | Present fix task proposal + escalate |

## Review Loop

Quality verification stage that runs automatically after implementation.
Applied uniformly across **all modes** (Solo / Sequential / Parallel / Breezing).

### Running the Review (via official plugin companion)

Use the companion review from the official plugin `codex-plugin-cc`.
Obtain the verdict in structured output (conforming to `review-output.schema.json`).

> **Diff base**: Use the `TASK_BASE_REF` specific to each task (HEAD at the time the task started).
> Review only the changes for that task, not accumulated diffs.

```bash
# Record base ref at task start (run before updating cc:WIP)
TASK_BASE_REF=$(git rev-parse HEAD)

# ... after implementation ...

# Run structured review via official plugin
bash scripts/codex-companion.sh review --base "${TASK_BASE_REF}"
REVIEW_EXIT=$?
```

**Verdict mapping** (official plugin → Harness format):

| Official plugin | Harness | Verdict impact |
|---|---|---|
| `approve` | `APPROVE` | - |
| `needs-attention` | `REQUEST_CHANGES` | - |
| `findings[].severity: critical` | `critical_issues[]` | Even 1 → REQUEST_CHANGES |
| `findings[].severity: high` | `major_issues[]` | Even 1 → REQUEST_CHANGES |
| `findings[].severity: medium/low` | `recommendations[]` | No impact on verdict |

Determine verdict from companion review output:
- `verdict` is `approve` → `APPROVE`
- `verdict` is `needs-attention` → `REQUEST_CHANGES`
- `findings` contains `critical` / `high` severity → `REQUEST_CHANGES`

### APPROVE / REQUEST_CHANGES Criteria

| Severity | Definition | Impact on verdict |
|--------|------|-----------------|
| **critical** | Security vulnerability, data loss risk, potential production incident | Even 1 → REQUEST_CHANGES |
| **major** | Breaks existing functionality, clear contradiction with spec, test failure | Even 1 → REQUEST_CHANGES |
| **minor** | Naming improvement, missing comments, style inconsistency | No impact on verdict |
| **recommendation** | Best practice suggestion, future improvement | No impact on verdict |

> **Important**: If only minor / recommendation issues are found, **always return APPROVE**.

### Fix Loop (on REQUEST_CHANGES)

```
review_count = 0
MAX_REVIEWS = 3

while verdict == "REQUEST_CHANGES" and review_count < MAX_REVIEWS:
    1. Analyze review findings (critical / major only)
    2. Implement fixes for each finding
    3. git commit --amend
    4. Run companion review again (from TASK_BASE_REF)
    review_count++

if review_count >= MAX_REVIEWS and verdict != "APPROVE":
    → Escalate to user
    → Display "Fixed 3 times but the following critical/major issues remain" + issue list
    → Wait for user decision (continue / abort)
```

### Application in Breezing Mode

1. Worker implements and commits in worktree → wait for completion with `wait_agent`
2. Lead reviews with companion review (from TASK_BASE_REF)
3. REQUEST_CHANGES → instruct Worker to fix via `send_input` → Worker amends
4. Re-review after fix (max 3 times)
5. Terminate Worker with `close_agent`
6. APPROVE → Lead cherry-picks to main → update Plans.md to `cc:Done [{hash}]`

### Worker Output Contract

The Worker prompt must explicitly specify that the following JSON is returned on completion:

```json
{
  "commit": "a1b2c3d",
  "files_changed": ["src/foo.ts", "tests/foo.test.ts"],
  "summary": "Add bar feature to foo module"
}
```

Lead parses this JSON to get the commit hash and file list.
If Worker does not return JSON, get the most recent commit with `git log --oneline -1`.

## Completion Report Format

Visual summary automatically output when a task completes.

### Solo Template

```
┌─────────────────────────────────────────────┐
│  ✓ Task {N} Done: {Task Name}               │
├─────────────────────────────────────────────┤
│  ■ What was done                            │
│    • {Change 1}                             │
│    • {Change 2}                             │
│  ■ What changed                             │
│    Before: {old behavior}                   │
│    After:  {new behavior}                   │
│  ■ Changed files ({N} files)                │
│    {file path 1}                            │
│  ■ Remaining items                          │
│    {M} incomplete tasks remain in Plans.md  │
│  commit: {hash} | review: {APPROVE}         │
└─────────────────────────────────────────────┘
```

### Breezing Template

```
┌─────────────────────────────────────────────┐
│  ✓ Breezing Done: {N}/{M} tasks             │
├─────────────────────────────────────────────┤
│  1. ✓ {Task Name 1}            [{hash1}]    │
│  2. ✓ {Task Name 2}            [{hash2}]    │
│  ■ Overall changes                          │
│    {N} files changed, {A} insertions(+),    │
│    {D} deletions(-)                         │
│  ■ Remaining items                          │
│    {K} incomplete tasks remain in Plans.md  │
└─────────────────────────────────────────────┘
```

## Differences from Claude Code Version

| Item | Claude Code version | Codex native version (this file) |
|------|---------------|-------------------------------|
| Worker spawn | `Agent(subagent_type="worker")` | `spawn_agent({message, fork_context})` |
| Wait for completion | `Agent` return value | `wait_agent({ids: [id]})` |
| Fix instruction | `SendMessage(to: agentId)` | `send_input({id, message})` |
| Worker termination | Automatic (Agent tool return value) | Explicit via `close_agent({id})` |
| Worktree isolation | `isolation="worktree"` auto-managed | Manual isolation via `git worktree add` |
| Permissions | `bypassPermissions` | companion `task --write` / `spawn_agent`: inherits session permissions |
| Review | Codex exec → Reviewer agent fallback | companion `review --base` (structured output) |
| Verdict retrieval | Parse Agent response | companion review `verdict` field (approve/needs-attention) |
| Mode auto-promotion | Auto-determined by task count | Explicit flags only (no auto-promotion) |
| Effort control | `ultrathink` + `/effort` | `model_reasoning_effort` in config.toml |
| Auto-Refinement | `/simplify` | None |

## Related Skills

- `harness-plan` — Plan the tasks to execute
- `harness-sync` — Sync implementation with Plans.md
- `harness-review` — Review implementation
- `harness-release` — Version bump and release
