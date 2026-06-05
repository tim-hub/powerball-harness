# Cleanup Execution Reference

Detailed step-by-step procedures for each `/maintenance` subcommand.

---

## --clear-state: Stale State File Removal

**Target directory**: `.claude/state/`

**Purpose**: Remove leftover files from completed loops, old contracts, and stale review results.

### Execution Steps

1. **List all files in `.claude/state/`**:
   ```bash
   find .claude/state -type f | sort
   ```

2. **Check each category**:

   #### loop-active.json
   - Check if a Breezing or harness-work loop is currently running:
     ```bash
     # If the file exists but no cc:WIP tasks are in Plans.md, it is stale
     grep 'cc:WIP' Plans.md 2>/dev/null || echo "No active WIP tasks"
     ```
   - If no `cc:WIP` tasks exist and no background agents are running: safe to remove.

   #### Sprint contracts (`.claude/state/contracts/*.sprint-contract.json`)
   - For each contract file, check the corresponding task in Plans.md:
     ```bash
     # Extract task ID from filename (e.g., 32.1.1.sprint-contract.json → task 32.1.1)
     grep 'cc:Done' Plans.md | grep '<task-id>'
     ```
   - If the task is marked `cc:Done`: safe to remove the contract.
   - If the task is `cc:WIP` or `cc:TODO`: keep the contract.

   #### review-result.json
   - Check the file modification time:
     ```bash
     find .claude/state -name 'review-result.json' -mtime +7
     ```
   - Files older than 7 days with no active review in progress: safe to remove.

   #### pending-fix-proposals.jsonl
   - **Never remove automatically** — these require explicit user approval via `approve fix` / `reject fix`.

3. **Dry-run output**:
   ```
   Safe to remove:
     .claude/state/loop-active.json (no active WIP tasks)
     .claude/state/contracts/32.1.1.sprint-contract.json (task cc:Done)
     .claude/state/contracts/32.1.2.sprint-contract.json (task cc:Done)
     .claude/state/review-result.json (8 days old)
   Will keep:
     .claude/state/contracts/33.1.sprint-contract.json (task cc:WIP)
     .claude/state/pending-fix-proposals.jsonl (requires manual approval)
   ```

4. **Confirm with user**, then remove identified files.

5. **Verify**:
   ```bash
   ls -la .claude/state/
   ```

---

## --purge-cache: Plugin Cache Purge

**Target directory**: `~/.claude/plugins/cache/`

**When to use**: After a major Harness version upgrade, or when the skill palette shows stale
entries that don't match the current plugin version.

### Execution Steps

1. **Check cache size**:
   ```bash
   du -sh ~/.claude/plugins/cache/ 2>/dev/null || echo "Cache directory not found"
   ```

2. **List cached plugin versions**:
   ```bash
   ls -la ~/.claude/plugins/cache/ 2>/dev/null
   ```

3. **Warn the user**: Purging the cache means all plugins must be re-fetched on next load.
   This adds ~5–10 seconds to the next session startup.

4. **Confirm with user** (required — this is the one destructive operation with no undo):
   ```
   This will delete ~/.claude/plugins/cache/ (SIZE).
   All plugins will be re-downloaded on next session start.
   Proceed? (y/N)
   ```

5. **Remove the cache**:
   ```bash
   rm -rf ~/.claude/plugins/cache/
   ```

6. **Confirm removal**:
   ```bash
   ls ~/.claude/plugins/cache/ 2>/dev/null || echo "Cache successfully removed"
   ```

### Safety Notes

- The local plugin source files in the repo are **not affected**
- The plugin marketplace URL and version pin are retained in `~/.claude/plugins/config.json`
- Claude Code will rebuild the cache automatically on next `/reload-plugins` or session start

---

## --clean-worktrees: Orphaned Worktree Cleanup

**Command**: `git worktree`

**When to use**: After completing a breezing run, when `git worktree list` shows entries for
branches that have been merged and deleted.

### Execution Steps

1. **List all worktrees**:
   ```bash
   git worktree list --porcelain
   ```
   Output format:
   ```
   worktree /path/to/main
   HEAD abc1234
   branch refs/heads/master

   worktree /path/to/worktree-task-32
   HEAD def5678
   branch refs/heads/worktree-task-32
   ```

2. **Identify orphaned worktrees** (branch no longer exists):
   ```bash
   git worktree list --porcelain | awk '/^branch/ {print $2}' | while read branch; do
     # Strip refs/heads/ prefix
     short="${branch#refs/heads/}"
     if ! git show-ref --verify --quiet "refs/heads/$short"; then
       echo "ORPHANED: $short"
     fi
   done
   ```

3. **Dry-run output**:
   ```
   Orphaned worktrees (branch deleted):
     /path/to/worktree-task-32  (branch: worktree-task-32 — not found)
     /path/to/worktree-task-33  (branch: worktree-task-33 — not found)
   Active worktrees (keep):
     /path/to/main  (branch: master)
   ```

4. **Confirm with user**, then remove each orphaned worktree:
   ```bash
   git worktree remove --force /path/to/worktree-task-32
   git worktree remove --force /path/to/worktree-task-33
   ```

5. **Prune stale administrative files**:
   ```bash
   git worktree prune
   ```

6. **Verify**:
   ```bash
   git worktree list
   ```

### Safety Notes

- `--force` is used because the worktree directory may already have been manually deleted
- The main worktree (current working directory) is **never** in the removal list
- Worktrees for branches that still exist are **never** removed

---

## --all: Run All Operations in Sequence

Executes the above operations in the following order:

1. `--clear-state`
2. `--clean-worktrees`
3. `--purge-cache` (last, because it requires the most explicit confirmation)

Each operation presents its dry-run summary and confirmation prompt independently.
If the user declines a step, that step is skipped and the next proceeds.

### Completion Summary

After all steps complete, print a summary:

```
Maintenance complete:
  .claude/state/: removed 3 stale files
  git worktrees: removed 2 orphaned entries
  plugin cache: skipped (user declined)
```
