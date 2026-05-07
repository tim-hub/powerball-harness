# Planning Discussion & Dependency Execution

## Phase 0: Planning Discussion (structured 3-question check)

Before executing all tasks, verify plan health with the following 3 questions.
All are skipped when `--no-discuss` is specified.

**Q1. Scope confirmation**:
> "Executing {{N}} tasks. Is the scope appropriate?"

If too many, suggest narrowing by priority (Required > Recommended > Optional).

**Q2. Dependency confirmation** (only when Plans.md has a Depends column):
> "Task {{X}} depends on {{Y}}. Is the execution order correct?"

Read the Depends column and display the dependency chain. Error if circular dependencies exist.

**Q3. Risk flag** (only when `[needs-spike]` tasks exist):
> "Task {{Z}} is [needs-spike]. Should we spike it first?"

If there are incomplete `[needs-spike]` tasks, confirm whether to run the spike first.

If all 3 questions pass, proceed to Phase A (designed to complete in 30 seconds total).

## Task Assignment Based on Dependency Graph

When Plans.md has a Depends column (v2 format), tasks are executed following the dependency graph:

1. Execute **tasks with Depends set to `-`** first. If multiple independent tasks exist, they can be spawned in parallel
2. After each Worker completes, Lead reviews → cherry-picks (see harness-work Phase B)
3. Once a dependency source task is cherry-picked to main, execute tasks that depended on it next
4. Repeat until all tasks are complete

> **Note**: The "Worker complete → review → cherry-pick" cycle for each task is sequential.
> Only the Worker spawn portion of independent tasks (Depends is `-`) can be parallelized.
