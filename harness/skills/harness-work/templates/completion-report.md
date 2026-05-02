# Completion Report Format

A visual summary auto-output on task completion (after `cc:Done` + commit).
Designed to convey change content and impact even to non-technical stakeholders.

## Solo / Parallel Template

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

## Generation Rules

1. **What was done**: Auto-extracted from `git diff --stat HEAD~1` and commit message. Minimize technical jargon, start with verbs
2. **What changed**: Infer Before/After from the task's "Content" and "DoD". Emphasize user experience changes
3. **Changed files**: Retrieved from `git diff --name-only HEAD~1`. Abbreviate with count when exceeding 5 files
4. **Remaining issues**: List `cc:TODO` / `cc:WIP` tasks from Plans.md. Indicate whether they are already in Plans.md
5. **Review**: Display review result (APPROVE / REQUEST_CHANGES → APPROVE)

## Reporting in Parallel Mode

- **1 task** (when forced with `--parallel`): Use Solo template above
- **Multiple tasks**: Use Breezing aggregate template (see below)

## Breezing Aggregate Template

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
