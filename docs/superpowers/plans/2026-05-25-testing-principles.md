# Testing Principles Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-extended-cc:subagent-driven-development (recommended) or superpowers-extended-cc:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade three harness files so downstream projects get enforced TDD task-splitting and mandatory phase-bottom verification instead of advisory suggestions.

**Architecture:** Three independent markdown edits — one template file, one planner reference, one worker reference. No new files. No script changes. Changes propagate to downstream projects via `harness-setup`.

**Spec:** `docs/superpowers/specs/2026-05-25-testing-principles-design.md`

---

### Task 1: Upgrade `tdd-guidelines.md.template`

**Goal:** Flip "suggestion → enforcement", update the decision logic tree, and add the Phase-bottom Verification section with three agent-executable lanes.

**Files:**
- Modify: `harness/templates/rules/tdd-guidelines.md.template`

**Acceptance Criteria:**
- [ ] File no longer contains the phrase "Suggestion, not enforcement"
- [ ] Decision Logic block shows the `.a`/`.b` split path
- [ ] `## Phase-bottom Verification` section present with lane table
- [ ] Lane table lists curl / `/chrome` / CLI lanes plus the docs/config skip row
- [ ] `./tests/validate-plugin.sh` exits 0

**Verify:** `grep -n "Suggestion, not enforcement" harness/templates/rules/tdd-guidelines.md.template` → no output; `grep -n "Phase-bottom Verification" harness/templates/rules/tdd-guidelines.md.template` → prints a line number

**Steps:**

- [ ] **Step 1: Replace the `## Notes` section**

Find this exact block at the bottom of the file:

```
## Notes

- **Suggestion, not enforcement**: The user decides
- **Respect existing tests**: Don't create duplicates
- **When no test framework is detected**: Skip TDD suggestion
```

Replace with:

```
## Notes

- **Enforcement**: Test-first is the default. Tasks tagged `[feature]` or `[bugfix]`
  are automatically split by the planner (see below). Add `[skip:tdd]` to a task to
  opt out.
- **Respect existing tests**: Don't create duplicates.
- **When no test framework is detected**: Auto-apply `[skip:tdd]`.
```

- [ ] **Step 2: Replace the `## Decision Logic (For Skills)` block**

Find this exact block:

````
## Decision Logic (For Skills)

```
Task received
    |
Check tags: [feature] / [bugfix] / [config] / [docs]
    |
Check file paths: src/core/, src/services/, src/api/
    |
    |-- [feature] + business logic -> Recommend TDD
    |-- [bugfix] -> Recommend reproduction test-first
    |-- [config] / [docs] -> TDD not needed
    |-- UI style only -> TDD not needed
```
````

Replace with:

````
## Decision Logic (For Skills)

```
Task received
    |
Check tags and [skip:tdd] absence
    |
    |-- [feature] or [bugfix] (no [skip:tdd])
    |       → Planner SPLITS into:
    |           N.a  [tdd:test-first]  Write failing tests
    |           N.b                    Implement to pass (blocked-by N.a)
    |
    |-- [skip:tdd] present → single task, no split
    |
    |-- [config] / [docs] / [style] / [refactor] → auto-apply [skip:tdd]
```
````

- [ ] **Step 3: Append the Phase-bottom Verification section at end of file**

Add after the `## Notes` section (i.e., at the very end of the file):

```markdown

## Phase-bottom Verification

At the end of every phase, the planner appends one verification task.
The Worker cannot close a phase until this task is `cc:done`.

### Lane Selection

All lanes are **agent-executable** — the Worker runs them autonomously.

| Project type | Lane | Agent execution | Verification criterion |
|---|---|---|---|
| Backend / API | **curl** | `bash`: `curl -f <endpoint>` | All critical routes return HTTP 2xx |
| Frontend / UI | **`/chrome`** | Claude Chrome extension — activate `/chrome` before the task | Golden path completes, 0 console errors |
| CLI tool | **CLI** | `bash`: run command end-to-end | Exit 0 + expected stdout matches |
| All tasks are `[skip:tdd]` | **— skip —** | N/A | No phase-bottom task appended |

Lane is inferred from the phase's task content. If mixed, use the dominant type.

### Verification Task Format

```markdown
| N.e2e | [verify:e2e] Phase N E2E — <lane> | <lane-specific criterion> | <last-impl-task> | cc:TODO |
```

**Skip rule**: If every task in the phase carries `[skip:tdd]` (docs/config-only phase),
no `N.e2e` task is appended.
```

- [ ] **Step 4: Commit**

```bash
git add harness/templates/rules/tdd-guidelines.md.template
git commit -m "feat: upgrade tdd-guidelines template — enforce split, add phase-bottom verification"
```

---

### Task 2: Update `harness-plan` create reference

**Goal:** Add the task-split rule to Step 5.5, add `[tdd:test-first]` to Step 6 marker logic, update the generation template to show `.a`/`.b`/`.e2e` rows, and add two new Step 6.5 self-review checks.

**Files:**
- Modify: `harness/skills/harness-plan/references/create.md`

**Acceptance Criteria:**
- [ ] Step 5.5 contains "### Task Split — Test-First Pair" subsection with the `.a`/`.b` format table
- [ ] Step 6 generation template example shows `1.1.a`, `1.1.b`, and `1.e2e` rows
- [ ] Step 6.5 self-review contains check 4 (TDD pair completeness) and check 5 (phase-bottom verification)
- [ ] `./tests/validate-plugin.sh` exits 0

**Verify:** `grep -n "Task Split" harness/skills/harness-plan/references/create.md` → prints a line number; `grep -n "1.1.a" harness/skills/harness-plan/references/create.md` → prints a line number

**Steps:**

- [ ] **Step 1: Append task-split rule to Step 5.5**

Find this exact line at the end of Step 5.5:

```
Tasks not matching the above have TDD automatically applied (test-first).
```

Replace with:

```
Tasks not matching the above have TDD automatically applied (test-first).

### Task Split — Test-First Pair

For every task that does NOT receive `[skip:tdd]`, split into a blocking pair:

| Suffix | Tag | Description | DoD | Depends |
|--------|-----|-------------|-----|---------|
| `N.a` | `[tdd:test-first]` | Write failing tests: {{task name}} | Failing test file exists and runs red | prior dep or `-` |
| `N.b` | _(original tags)_ | {{original description}} | {{original DoD}} | `N.a` |

Rules:
- `N.b` always lists `N.a` in `Depends`.
- Original DoD moves to `N.b`; `N.a` DoD is always "Failing test file exists and runs red".
- If the original task had a prior dependency (e.g., `1.1`), `N.a` inherits it; `N.b` depends on `N.a`.
- `[bugfix]` tasks: `N.a` DoD is "Reproduction test exists and fails on current code".
```

- [ ] **Step 2: Add split line to the Quality Marker Assignment Logic block**

Find this exact line inside the marker logic code block:

```
    +-- other -> no marker (TDD enabled by default)
```

Replace with:

```
    +-- other -> no marker (TDD enabled by default)
    +-- [feature] / [bugfix] (no [skip:tdd]) -> split into N.a [tdd:test-first] + N.b (see Step 5.5)
```

- [ ] **Step 3: Replace the generation template task table example**

Find this exact block in the generation template:

```markdown
| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 1.1  | [Task description] [feature:security] | [Verifiable completion criteria] | - | cc:TODO |
| 1.2  | [Task description] | [Verifiable completion criteria] | 1.1 | cc:TODO |
```

Replace with:

```markdown
| Task  | Description | DoD | Depends | Status |
|-------|-------------|-----|---------|--------|
| 1.1.a | [tdd:test-first] Write failing tests: {{task name}} | Failing test runs red | - | cc:TODO |
| 1.1.b | [Task description] [feature:security] | [Verifiable completion criteria] | 1.1.a | cc:TODO |
| 1.2.a | [tdd:test-first] Write failing tests: {{task name}} | Failing test runs red | - | cc:TODO |
| 1.2.b | [Task description] | [Verifiable completion criteria] | 1.2.a | cc:TODO |
| 1.e2e | [verify:e2e] Phase 1 E2E — curl: critical endpoints return 2xx | `curl -f` exits 0 for all routes | 1.2.b | cc:TODO |
```

- [ ] **Step 4: Add two checks to Step 6.5 Self-Review**

Find the end of the Step 6.5 block:

```
3. **Dependency consistency** — Do all task numbers in the `Depends` column reference tasks that exist in the phase?

Fix issues inline. No need to re-review — just fix and move on.
```

Replace with:

```
3. **Dependency consistency** — Do all task numbers in the `Depends` column reference tasks that exist in the phase?
4. **TDD pair completeness** — Every task without `[skip:tdd]` must have a corresponding `.a` row. Fix any missing splits.
5. **Phase-bottom verification** — Every phase with at least one non-`[skip:tdd]` task must end with an `N.e2e` row. Fix any missing.

Fix issues inline. No need to re-review — just fix and move on.
```

- [ ] **Step 5: Commit**

```bash
git add harness/skills/harness-plan/references/create.md
git commit -m "feat: harness-plan enforces test-first split and phase-bottom E2E task"
```

---

### Task 3: Update `harness-work` solo-mode reference

**Goal:** Add a dependency guard at Step 1, replace the TDD Phase block at Step 3 with a behaviour table distinguishing `.a`/`.b`/legacy tasks, and add a phase-close gate at Step 11.

**Files:**
- Modify: `harness/skills/harness-work/references/solo-mode.md`

**Acceptance Criteria:**
- [ ] Step 1 contains "Dependency check" paragraph blocking `.b` tasks while `.a` is open
- [ ] Step 3 contains a behaviour table with rows for `[tdd:test-first]`, `.b`, legacy, and `[skip:tdd]`
- [ ] Step 11 contains "Phase-close check" with three outcome bullets
- [ ] `./tests/validate-plugin.sh` exits 0

**Verify:** `grep -n "Dependency check" harness/skills/harness-work/references/solo-mode.md` → prints a line number; `grep -n "Phase-close check" harness/skills/harness-work/references/solo-mode.md` → prints a line number

**Steps:**

- [ ] **Step 1: Add dependency guard after Step 1 task identification**

Find this exact block at Step 1:

```
1. Read Plans.md and identify the target task
   - **If Plans.md does not exist**: Auto-invoke `harness-plan create --ci` → Generate Plans.md and continue
   - If header lacks DoD / Depends columns: `Plans.md is in the old format. Please regenerate with harness-plan create.` → **Stop**
   - **If the conversation contains unlisted tasks**: Extract requirements from the recent conversation context and auto-append to Plans.md as `cc:TODO`
     - Extraction logic: Detect action verbs from user statements ("add...", "fix...", "implement...")
     - Appended entries conform to v2 format (Task / Content / DoD / Depends / Status)
     - After appending, display "Added the following to Plans.md" with a 5-second timeout prompt (default: continue)
```

Replace with:

```
1. Read Plans.md and identify the target task
   - **If Plans.md does not exist**: Auto-invoke `harness-plan create --ci` → Generate Plans.md and continue
   - If header lacks DoD / Depends columns: `Plans.md is in the old format. Please regenerate with harness-plan create.` → **Stop**
   - **If the conversation contains unlisted tasks**: Extract requirements from the recent conversation context and auto-append to Plans.md as `cc:TODO`
     - Extraction logic: Detect action verbs from user statements ("add...", "fix...", "implement...")
     - Appended entries conform to v2 format (Task / Content / DoD / Depends / Status)
     - After appending, display "Added the following to Plans.md" with a 5-second timeout prompt (default: continue)
   - **Dependency check**: Before claiming a task, verify its `Depends` column. If any listed dependency is not yet `cc:done`, skip this task and select the next eligible one. For `.b` tasks whose `.a` is still open: redirect to `.a` first.
```

- [ ] **Step 2: Replace the Step 3 TDD Phase block**

Find this exact block:

```
3. **TDD Phase** (when `[skip:tdd]` is absent & test framework exists):
   a. Create test file first (Red)
   b. Confirm failure
```

Replace with:

```
3. **TDD Phase** — behaviour depends on task tag:

   | Task type | Step 3 action |
   |---|---|
   | `[tdd:test-first]` (an `.a` task) | **This task IS the TDD phase.** Write the failing test file; confirm it runs red. Commit as `test: failing tests for {{feature}}`. Stop here — do not proceed to Step 6 implementation. |
   | `.b` task (Depends on a `.a`) | Confirm tests from `.a` still run red, then proceed directly to Step 6. No new test file needed. |
   | No split (legacy task, no `[skip:tdd]`) | Existing behaviour: create test file first, confirm failure. |
   | `[skip:tdd]` present | Skip Step 3 entirely. |
```

- [ ] **Step 3: Add phase-close gate after Step 11**

Find this exact block:

```
11. Update task to `cc:Done` (with commit hash)
   - Get the latest commit hash (abbreviated 7 chars) with `git log --oneline -1`
   - Update Plans.md Status to `cc:Done [a1b2c3d]` format (authoritative)
   - If no commit (`--no-commit`), use `cc:Done` without hash
   - **Mirror to native task list**: run `TaskList`, locate any native task whose title starts with the Plans.md task ID prefix (e.g., `97.1`), and call `TaskUpdate(status="completed")` on it. If no matching native task exists, skip silently — the native task list is a mirror, never authoritative.
```

Replace with:

```
11. Update task to `cc:Done` (with commit hash)
   - Get the latest commit hash (abbreviated 7 chars) with `git log --oneline -1`
   - Update Plans.md Status to `cc:Done [a1b2c3d]` format (authoritative)
   - If no commit (`--no-commit`), use `cc:Done` without hash
   - **Mirror to native task list**: run `TaskList`, locate any native task whose title starts with the Plans.md task ID prefix (e.g., `97.1`), and call `TaskUpdate(status="completed")` on it. If no matching native task exists, skip silently — the native task list is a mirror, never authoritative.
   - **Phase-close check**: Scan the current phase in Plans.md.
     - All tasks except `[verify:e2e]` are `cc:done` AND `N.e2e` is `cc:TODO` → surface "Phase N implementation complete — run E2E verification task (N.e2e) next" and select it as the next task.
     - `[verify:e2e]` task is `cc:done` → phase is fully closed.
     - No `[verify:e2e]` task in phase (docs/config-only) → phase closes normally.
```

- [ ] **Step 4: Commit**

```bash
git add harness/skills/harness-work/references/solo-mode.md
git commit -m "feat: harness-work enforces .a/.b task order and phase-close E2E gate"
```

---

### Task 4: Validate all changes

**Goal:** Confirm the three edits pass the plugin validation suite and the consistency check.

**Files:**
- Read: `tests/validate-plugin.sh` output

**Acceptance Criteria:**
- [ ] `./tests/validate-plugin.sh` exits 0
- [ ] `./.claude/skills/release-this/scripts/check-consistency.sh` exits 0
- [ ] `grep -n "Suggestion, not enforcement" harness/templates/rules/tdd-guidelines.md.template` → empty
- [ ] `grep -n "Phase-bottom Verification" harness/templates/rules/tdd-guidelines.md.template` → match found
- [ ] `grep -n "Task Split" harness/skills/harness-plan/references/create.md` → match found
- [ ] `grep -n "Dependency check" harness/skills/harness-work/references/solo-mode.md` → match found
- [ ] `grep -n "Phase-close check" harness/skills/harness-work/references/solo-mode.md` → match found

**Verify:** `./tests/validate-plugin.sh && echo PASS` → prints PASS

**Steps:**

- [ ] **Step 1: Run plugin validation**

```bash
./tests/validate-plugin.sh
```

Expected: exits 0. If it fails, read the error output and fix the relevant file.

- [ ] **Step 2: Run consistency check**

```bash
./.claude/skills/release-this/scripts/check-consistency.sh
```

Expected: exits 0.

- [ ] **Step 3: Verify grep checks**

```bash
grep -n "Suggestion, not enforcement" harness/templates/rules/tdd-guidelines.md.template
# Expected: no output

grep -n "Phase-bottom Verification" harness/templates/rules/tdd-guidelines.md.template
grep -n "Task Split" harness/skills/harness-plan/references/create.md
grep -n "Dependency check" harness/skills/harness-work/references/solo-mode.md
grep -n "Phase-close check" harness/skills/harness-work/references/solo-mode.md
# Expected: each prints a line number
```

- [ ] **Step 4: Commit validation result**

```bash
git add -p  # review any incidental changes
git commit -m "chore: validate testing principles implementation"
```
