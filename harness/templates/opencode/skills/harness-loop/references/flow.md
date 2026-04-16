# harness-loop: Wake-up Flow Details

Detailed reference for each wake-up entry procedure in `harness-loop`.
Supplements the summary in SKILL.md with implementation-level detail.

---

## Per Wake-up Entry Procedure (Detailed)

### Step 0: Concurrency Guard Lock (Idempotency Guard a)

```bash
LOCK_DIR=".claude/state/locks/loop-session.lock.d"
mkdir -p ".claude/state/locks"

# Atomic creation (fail immediately if exists — avoids TOCTOU race)
if ! mkdir "${LOCK_DIR}" 2>/dev/null; then
    existing=$(cat "${LOCK_DIR}/meta.json" 2>/dev/null || echo '{}')
    echo "ERROR: harness-loop is already running (lock dir exists: ${LOCK_DIR})" >&2
    echo "Lock contents: ${existing}" >&2
    echo "To force-clear, run: rm -rf ${LOCK_DIR}" >&2
    exit 10
fi

# Write lock metadata inside the lock directory
SESSION_ID="${CLAUDE_SESSION_ID:-unknown}"
ARGS_STR="$*"
cat > "${LOCK_DIR}/meta.json" <<EOF
{
  "pid": $$,
  "session_id": "${SESSION_ID}",
  "started_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "args": "${ARGS_STR}"
}
EOF

# Remove lock on exit (normal or abnormal)
cleanup_loop_lock() {
    rm -rf "${LOCK_DIR}" 2>/dev/null || true
}
trap cleanup_loop_lock EXIT INT TERM
```

- `LOCK_DIR` is `.claude/state/locks/loop-session.lock.d` (a directory)
- `mkdir` is atomic so TOCTOU races cannot occur (only one of two concurrent processes succeeds)
- Lock metadata is written to `${LOCK_DIR}/meta.json`: JSON with `pid`, `session_id`, `started_at`, `args`
- If a lock already exists: `already running` error (exit 10), stop immediately
- `EXIT` / `INT` / `TERM` all remove the lock (cleanup regardless of normal or abnormal exit)
- `rm -rf` is idempotent (safe to run twice)

### Step 0.5: State Consistency Check (Idempotency Guard b)

```bash
# Run lightweight consistency check in --quick mode at wake-up start
# Stop the loop immediately on failure (protects against corrupted Plans.md / uninitialized env)
if bash tests/validate-plugin.sh --quick; then
    : # OK — continue
else
    echo "harness-loop: state consistency check failed — stopping loop" >&2
    echo "Details: run bash tests/validate-plugin.sh --quick to investigate" >&2
    exit 1
fi
```

- `tests/validate-plugin.sh --quick` completes in seconds
- Checks: `.claude/state/` existence / Plans.md existence + v2 format / sprint-contract format
- Does NOT run full validate (39 verification items)
- If Plans.md is intentionally corrupted and this check fails, the loop stops immediately

### Step 1: Read Plans.md First

```bash
# Extract cc:WIP / cc:TODO tasks and identify the leading task's task_id
grep -E "cc:(WIP|TODO)" Plans.md | head -1
```

- If `cc:WIP` tasks remain: may have been interrupted in a previous cycle → get task_id and continue
- If `cc:TODO` tasks exist: get task_id as the next target
- If neither: **all tasks complete** → loop ends normally

> **Plans lock prerequisite**: If `plans-watcher.sh` is protecting Plans.md with flock,
> perform the Plans.md read within that flock scope.
> Without flock protection, direct read is fine.

### Step 2: Sprint-contract Existence Check & Generation

```bash
CONTRACT_PATH=".claude/state/contracts/${task_id}.sprint-contract.json"

if [ ! -f "${CONTRACT_PATH}" ]; then
    # Contract not yet generated → generate it
    node harness/scripts/generate-sprint-contract.js "${task_id}"

    # Step 2.5: Promote draft → approved (first generation only)
    # generate-sprint-contract.js initializes review.status == "draft"
    # ensure-sprint-contract-ready.sh (next step) requires "approved"
    # so we must promote before calling it
    bash harness/scripts/enrich-sprint-contract.sh "${CONTRACT_PATH}" \
      --check "auto-approve (harness-loop — confirm DoD from reviewer perspective)" \
      --approve
fi
```

- Check `.claude/state/contracts/${task_id}.sprint-contract.json` for existence
- If absent: generate with `node harness/scripts/generate-sprint-contract.js ${task_id}`
- **On first generation only**: promote `draft` → `approved` via `enrich-sprint-contract.sh --approve`
  - `generate-sprint-contract.js` initializes `review.status == "draft"`
  - `ensure-sprint-contract-ready.sh` (Step 3) only accepts `approved`
  - Placing this inside the `if [ ! -f ... ]` block prevents applying it to contracts already approved in prior cycles
- After generation, reuse `${CONTRACT_PATH}` in subsequent steps

### Step 3: Contract Readiness Check

```bash
bash harness/scripts/ensure-sprint-contract-ready.sh "${CONTRACT_PATH}"
```

- Verifies sprint-contract `review.status == "approved"`
- Stops with an error if an unapproved contract remains

### Step 4: Resume Pack Reload

```
Step 4. harness-mem resume-pack reload:
  Call the mcp__harness__harness_mem_resume_pack tool.
  Required arguments:
    - project: current project name (e.g., derive with
              `basename $(git rev-parse --show-toplevel)`)
  Optional: session_id (when resuming from a prior session)

  Example (pseudocode):
    resume_pack = mcp__harness__harness_mem_resume_pack(
      project="claude-code-harness",
      session_id=<session_id from previous checkpoint>
    )
```

After wake-up with fresh context, the previous cycle's memory is lost.
Re-inject the following via `harness-mem resume-pack`:

- `decisions.md` — architecture decisions
- `patterns.md` — reusable patterns
- `session-state` — previous work state
- Most recent cycle's `checkpoint` — what was completed

> **Note**: Perform resume pack reload after Step 3 (contract readiness check).
> Skipping it risks re-implementing artifacts from prior cycles.

### Step 5: Execute 1 Task Cycle

Spawn `claude-code-harness:worker` via the Agent tool:

> **Important**: Specify `"claude-code-harness:worker"` for `subagent_type`, NOT `"harness-work"`.
> `harness-work` is a skill, not an agent. The actual agents are `worker` / `reviewer` / `scaffolder`.
> Specifying `"harness-work"` will cause Agent spawn failure, stopping the loop at the first Worker launch.

```python
worker_result = Agent(
    subagent_type="claude-code-harness:worker",  # worker agent (not a skill)
    prompt="""
    Task: ${task_id}
    DoD: <extracted from Plans.md>
    contract_path: ${CONTRACT_PATH}
    mode: breezing
    On completion: return commit hash, branch, and change summary.
    """,
    isolation="worktree",
    run_in_background=false  # foreground execution (wait for completion)
)
# worker_result: { commit, branch, worktreePath, files_changed, summary }
```

Worker runs in `mode: breezing`, so it:
- Only commits on the feature branch — does not touch main
- Stores changes in `worktreePath`
- Lead (harness-loop) handles review → cherry-pick in Steps 5.5/5.6

> **Implementation note**: `Bash("harness-work --breezing")` is also viable,
> but the Agent tool provides cleaner context isolation and is easier to debug.

### Step 5.5: Lead Review Execution

Lead reviews the commit returned by Worker:

```bash
# Get diff (targeting commit inside worktree)
diff_text=$(git -C "${worker_result.worktreePath}" show "${worker_result.commit}")

# ── (a) Codex companion review: run inside Worker's worktree directory ──────────────
# Running from Lead's main repo dir would yield an empty diff (risk of unconditional APPROVE).
# By cd-ing into worktreePath before calling review, we get the correct diff.
#
# If worktreePath is empty or identical to the main repo (no worktree isolation),
# fall back to running from Lead dir (same behavior as before).

MAIN_REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "")
WORKER_PATH="${worker_result.worktreePath:-}"

if [ -n "${WORKER_PATH}" ] && [ "${WORKER_PATH}" != "${MAIN_REPO_ROOT}" ]; then
    # Run review inside Worker's worktree → see actual diff of Worker feature branch
    ( cd "${WORKER_PATH}" && bash harness/scripts/codex-companion.sh review --base "${BASE_REF}" )
    REVIEW_EXIT=$?
    # review-output.json is created in Worker worktree dir — manage as absolute path
    REVIEW_OUTPUT_PATH="${WORKER_PATH}/review-output.json"
else
    # Fallback: run from Lead dir (when worktree isolation is not available)
    bash harness/scripts/codex-companion.sh review --base "${BASE_REF}"
    REVIEW_EXIT=$?
    REVIEW_OUTPUT_PATH="$(pwd)/review-output.json"
fi
# All subsequent steps use $REVIEW_OUTPUT_PATH (do not reference relative "review-output.json" directly)

# ── (b) reviewer_profile branching (check sprint-contract's review.reviewer_profile) ──
if command -v jq >/dev/null 2>&1; then
    REVIEWER_PROFILE=$(jq -r '.review.reviewer_profile // "static"' "${CONTRACT_PATH}" 2>/dev/null || echo "static")
else
    REVIEWER_PROFILE="static"
fi

case "${REVIEWER_PROFILE}" in
    runtime)
        # Run runtime validation command; may override verdict
        # run-contract-review-checks.sh runs inside Worker's worktree (test env is there)
        # Important: stdout of run-contract-review-checks.sh is the artifact FILE PATH (not JSON payload)
        if [ -n "${WORKER_PATH}" ] && [ "${WORKER_PATH}" != "${MAIN_REPO_ROOT}" ]; then
            RUNTIME_ARTIFACT_PATH=$(
                cd "${WORKER_PATH}" && bash harness/scripts/run-contract-review-checks.sh "${CONTRACT_PATH}" 2>/dev/null
            ) || RUNTIME_ARTIFACT_PATH=""
        else
            RUNTIME_ARTIFACT_PATH=$(
                bash harness/scripts/run-contract-review-checks.sh "${CONTRACT_PATH}" 2>/dev/null
            ) || RUNTIME_ARTIFACT_PATH=""
        fi

        # Empty (script failed) → treat as DOWNGRADE_TO_STATIC
        if [ -z "${RUNTIME_ARTIFACT_PATH}" ]; then
            RUNTIME_ARTIFACT_PATH=""
            RUNTIME_VERDICT="DOWNGRADE_TO_STATIC"
        else
            # Convert relative path to absolute using WORKER_PATH (or Lead dir) as base
            if [[ "${RUNTIME_ARTIFACT_PATH}" != /* ]]; then
                if [ -n "${WORKER_PATH}" ] && [ "${WORKER_PATH}" != "${MAIN_REPO_ROOT}" ]; then
                    RUNTIME_ARTIFACT_PATH="${WORKER_PATH}/${RUNTIME_ARTIFACT_PATH}"
                else
                    RUNTIME_ARTIFACT_PATH="$(pwd)/${RUNTIME_ARTIFACT_PATH}"
                fi
            fi

            # Read verdict from artifact file
            if command -v jq >/dev/null 2>&1; then
                RUNTIME_VERDICT=$(jq -r '.verdict // "DOWNGRADE_TO_STATIC"' "${RUNTIME_ARTIFACT_PATH}" 2>/dev/null || echo "DOWNGRADE_TO_STATIC")
            else
                RUNTIME_VERDICT=$(python3 -c "import json,sys; d=json.load(open(sys.argv[1])); print(d.get('verdict','DOWNGRADE_TO_STATIC'))" "${RUNTIME_ARTIFACT_PATH}" 2>/dev/null || echo "DOWNGRADE_TO_STATIC")
            fi
        fi

        if [ "${RUNTIME_VERDICT}" = "REQUEST_CHANGES" ]; then
            # Runtime validation failed → override verdict to REQUEST_CHANGES
            EFFECTIVE_VERDICT="REQUEST_CHANGES"
            REVIEW_RESULT_INPUT="${RUNTIME_ARTIFACT_PATH}"
        elif [ "${RUNTIME_VERDICT}" = "DOWNGRADE_TO_STATIC" ]; then
            # No runtime validation command → use static verdict as-is
            EFFECTIVE_VERDICT=""  # → read from REVIEW_OUTPUT_PATH
            REVIEW_RESULT_INPUT="${REVIEW_OUTPUT_PATH}"
        else
            EFFECTIVE_VERDICT="${RUNTIME_VERDICT}"
            REVIEW_RESULT_INPUT="${RUNTIME_ARTIFACT_PATH}"
        fi
        ;;
    browser)
        # Generate artifact for browser reviewer to use later
        # Browser artifact is a PENDING_BROWSER scaffold; actual browser execution is done by reviewer agent
        # review-result verdict remains static (not PENDING_BROWSER)
        bash harness/scripts/generate-browser-review-artifact.sh "${CONTRACT_PATH}" 2>/dev/null || true
        EFFECTIVE_VERDICT=""  # → read from REVIEW_OUTPUT_PATH (use static verdict)
        REVIEW_RESULT_INPUT="${REVIEW_OUTPUT_PATH}"
        ;;
    *)
        # static (default): use Codex companion review verdict as-is
        EFFECTIVE_VERDICT=""
        REVIEW_RESULT_INPUT="${REVIEW_OUTPUT_PATH}"
        ;;
esac

# If EFFECTIVE_VERDICT is not set, read from REVIEW_OUTPUT_PATH (absolute path)
if [ -z "${EFFECTIVE_VERDICT}" ]; then
    if command -v jq >/dev/null 2>&1; then
        EFFECTIVE_VERDICT=$(jq -r '.verdict // "REQUEST_CHANGES"' "${REVIEW_OUTPUT_PATH}" 2>/dev/null || echo "REQUEST_CHANGES")
    else
        EFFECTIVE_VERDICT=$(python3 -c "import json,sys; d=json.load(open(sys.argv[1])); print(d.get('verdict','REQUEST_CHANGES'))" "${REVIEW_OUTPUT_PATH}" 2>/dev/null || echo "REQUEST_CHANGES")
    fi
fi

# Normalize and save review result
# REVIEW_RESULT_INPUT is runtime artifact path when runtime REQUEST_CHANGES, otherwise REVIEW_OUTPUT_PATH
# This ensures runtime REQUEST_CHANGES propagates correctly to pretooluse-guard
bash harness/scripts/write-review-result.sh "${REVIEW_RESULT_INPUT}" "${worker_result.commit}"
```

**Verdict determination**:

| verdict | Action |
|---------|--------|
| `APPROVE` | Proceed to Step 5.6 (cherry-pick) |
| `REQUEST_CHANGES` | Enter fix loop (max 3 iterations) |

**Fix loop (on REQUEST_CHANGES)**:

```python
review_count = 0
latest_commit = worker_result.commit
worker_id = worker_result.agentId
# Read max_iterations from sprint-contract if present; fallback to 3 (backward compat)
MAX_REVIEWS = read_contract(contract_path, ".review.max_iterations") or 3

while verdict == "REQUEST_CHANGES" and review_count < MAX_REVIEWS:
    # Send fix instructions to Worker (resume via SendMessage)
    SendMessage(to=worker_id, message=f"Issues found: {issues}\nPlease fix and amend")
    updated_result = wait_for_response(worker_id)
    latest_commit = updated_result.commit
    diff_text = git("-C", worker_result.worktreePath, "show", latest_commit)
    verdict = codex_exec_review(diff_text) or reviewer_agent_review(diff_text)
    review_count += 1

if review_count >= MAX_REVIEWS and verdict != "APPROVE":
    # Escalation
    raise PivotRequired(f"Still REQUEST_CHANGES after {MAX_REVIEWS} fix attempts: {issues}")
```

### Step 5.6: APPROVE → Cherry-pick to Main

```bash
# Return to trunk branch (Worker worked on feature branch)
TRUNK=$(git symbolic-ref refs/remotes/origin/HEAD 2>/dev/null | sed 's|refs/remotes/origin/||' || echo "main")
git checkout "${TRUNK}"

# Verify commit is not already in trunk (re-entry guard)
if ! git merge-base --is-ancestor "${latest_commit}" HEAD; then
    git cherry-pick --no-commit "${latest_commit}"
    git commit -m "${task_title}"
fi

# ── (c) Cleanup order: worktree remove → branch -D ────────────────────────────────
# `git branch -D` fails with "branch is checked out at <path>" if the feature branch
# is still checked out in a worktree. Remove the worktree first to make branch -D safe.
#
# Order:
#   1. cherry-pick → incorporated into main (git commit above)
#   2. worktree remove (remove the worktree where feature branch was checked out)
#   3. branch -D (safe now that worktree is removed)

MAIN_REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || echo "")
WORKER_PATH="${worker_result.worktreePath:-}"

# Step 2: worktree remove
if [ -n "${WORKER_PATH}" ] && [ "${WORKER_PATH}" != "${MAIN_REPO_ROOT}" ]; then
    git worktree remove "${WORKER_PATH}" --force 2>/dev/null || true
fi

# Step 3: branch -D (safe after worktree remove)
if [ -n "${worker_result.branch}" ] && \
   [ "${worker_result.branch}" != "main" ] && \
   [ "${worker_result.branch}" != "master" ] && \
   [ "${worker_result.branch}" != "${TRUNK}" ]; then
    git branch -D "${worker_result.branch}" 2>/dev/null || true
fi
```

Update Plans.md:

```bash
# cc:WIP → cc:Done [{hash}]
HASH=$(git rev-parse --short HEAD)
# Update the relevant task line in Plans.md
```

### Step 6: Plateau Detection

```bash
bash harness/scripts/detect-review-plateau.sh ${current_task_id}
PLATEAU_EXIT=$?
# Note: current_task_id is the task_id identified in Step 1
```

| exit code | Meaning | Action |
|-----------|---------|--------|
| `0` | `PIVOT_NOT_REQUIRED` | Continue |
| `1` | `INSUFFICIENT_DATA` | Continue (insufficient data) |
| `2` | `PIVOT_REQUIRED` | **Stop loop** + escalation |

**Advisor consultation on PIVOT_REQUIRED** (when `--advisor` is enabled):

Before escalating to the user, call the advisor:

```bash
bash harness/scripts/run-advisor-consultation.sh \
  --reason-code plateau_before_escalation \
  --task-id "${current_task_id}"
```

If advisor returns `PLAN`, retry with the suggested approach.
If advisor returns `STOP`, proceed to user escalation.

**User escalation message on PIVOT_REQUIRED**:

```
harness-loop: stopped due to plateau detection (cycle {N}/{max})

Detected issue:
  {plateau details: output from detect-review-plateau.sh}

Suggested actions:
  1. Manually review and revise task content
  2. Re-run with --pacing plateau to extend the interval
  3. Skip the problem task and restart /harness-loop

Please review the current Plans.md state.
```

**Pre-escalation advisor check** (when `--advisor` is enabled):

Before presenting any STOP/failure to the user, call:

```bash
bash harness/scripts/run-advisor-consultation.sh \
  --reason-code pre_user_escalation \
  --task-id "${current_task_id}"
```

### Step 7: Cycle Count Check

```
cycles_completed += 1
if cycles_completed >= max_cycles:
    stop loop
    print(f"harness-loop: stopped after {max_cycles} cycles")
    return
```

- Default `max_cycles = 8`
- With `--max-cycles N`: stop after N cycles

**Persisting cycle count**:
- Embed the count in the `prompt` argument of `ScheduleWakeup`:
  ```
  /harness-loop all --max-cycles 8 --cycles-done {N} --pacing worker
  ```
- On wake-up, read `--cycles-done N` to restore the count

### Step 8: Record Checkpoint

```json
{
  "session_id": "<current session ID>",
  "title": "harness-loop cycle {N}/{max}: {task_completed}",
  "content": "Cycle {N} complete. commit: {commit}. changes: {files_changed}. next: {next_task}"
}
```

Record in memory via the `harness_mem_record_checkpoint` tool.
Automatically included in the next wake-up's resume pack.

### Step 9: Schedule Next Wake-up

```
ScheduleWakeup(
    delaySeconds=<value for pacing>,
    prompt="/harness-loop <same args> --cycles-done {N}",
    reason="Cycle {N}/{max} complete: {task_completed}"
)
```

**delaySeconds per pacing**:

| pacing | delaySeconds | Selection rationale |
|--------|-------------|---------------------|
| `worker` | 270 | Re-entry immediately after Worker completion (within 5 min cache warm) |
| `ci` | 270 | Assumes shortest CI job completion time |
| `plateau` | 1200 | 20 min cooling period (plateau avoidance) |
| `night` | 3600 | Overnight batch (maximum clamp value) |

> **Clamp constraint**: `ScheduleWakeup` clamps `delaySeconds` to `[60, 3600]` at runtime.
> Values below 60 are raised to 60; values above 3600 are lowered to 3600.
> All design values are within range, but note this if changing values in the future.

---

## Cycle Stop Conditions Matrix

| Condition | Cycle count | exit | Stop reason | User notification |
|-----------|------------|------|-------------|-------------------|
| `cycles >= max_cycles` | N (limit) | 0 | Normal limit | "Stopped after {N} cycles" |
| `PIVOT_REQUIRED` | Any | 2 | Plateau detected | Escalation details |
| No incomplete tasks | Any | 0 | All tasks done | Completion report |
| User cancel | Any | - | Manual interrupt | - |

---

## Pacing Selection Guide

### Which pacing to use

```
What is the nature of the task?
│
├── Want to re-enter immediately after Worker completion
│     → worker (270s)
│
├── Need to wait for CI / test completion
│     → ci (270s)
│     (If CI takes more than 270s, adjust --pacing manually)
│
├── Detected plateau and want to add spacing
│     → plateau (1200s)
│
└── Want to leave overnight and check in the morning
      → night (3600s)
```

### When to change pacing

- **On initial launch**: `worker` (default) is usually fine
- **When lots of CI waiting**: switch to `--pacing ci`
- **After plateau detection**: consider auto-switching to `--pacing plateau` (see Step 5)
- **Overnight runs**: launch with `--pacing night` and go to sleep

---

## ScheduleWakeup Constraint Details

### Runtime constraint on delaySeconds

```
ScheduleWakeup(delaySeconds=X)
  → X < 60  → clamp to 60
  → X > 3600 → clamp to 3600
  → 60 <= X <= 3600 → used as-is
```

### Relationship with cache TTL

ScheduleWakeup's cache TTL is **5 min (300s)**.

- `worker` / `ci` at 270s is within 5 min → wake-up with warm cache
- `plateau` at 1200s and `night` at 3600s exceed cache expiry → wake-up after cache miss
  → Step 4 (resume pack reload) is especially important in these cases

### Passing arguments to the next wake-up

Method for carrying cycle count to the next wake-up:

```bash
# Embed current cycle count in prompt
NEXT_PROMPT="/harness-loop ${SCOPE} --max-cycles ${MAX_CYCLES} --cycles-done ${CYCLES_DONE} --pacing ${PACING}"

ScheduleWakeup(
    delaySeconds=${DELAY},
    prompt="${NEXT_PROMPT}",
    reason="cycle ${CYCLES_DONE}/${MAX_CYCLES} complete"
)
```

---

## Reference: spike 41.0.0 Verification Results

This design is based on the empirical results from spike 41.0.0:

- `ScheduleWakeup`: confirmed as an internal tool. delay [60, 3600] clamp, cache 5min TTL
- `/loop`: confirmed as CC dynamic mode. sentinel `<<autonomous-loop-dynamic>>`
- `harness_mem_record_checkpoint`: confirmed (schema: session_id / title / content required)

Update this file if any of these assumptions change.
