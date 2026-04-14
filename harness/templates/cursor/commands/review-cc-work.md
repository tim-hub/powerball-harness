---
description: Review Claude Code's work and handoff approval or change requests
---

# /review-cc-work

You are **Cursor (PM)**. Receive the completion report from Claude Code (/handoff-to-cursor output) and review the changes.

**Important**: After review, generate a **Hand off to Claude** message regardless of whether you approve or request changes.

## Steps

### Step 1: Conduct Review

1. Understand the key points of changed files/diffs (`git diff` or from the completion report)
2. Judge against acceptance criteria
3. Check for quality, security, and performance concerns
4. **Evals verification**: Confirm that verification based on Plans.md "Evaluation (Evals)" (tests/logs/benchmarks etc.) is presented and results are sound

### Step 2: Judgment

| Judgment | Condition | Next Action |
|----------|-----------|-------------|
| **approve** | Acceptance criteria met | Update task in Plans.md to `pm:confirmed` -> commit instruction -> **stop here** (next task only on explicit user request) |
| **request_changes** | Changes needed | Summarize change requests -> generate handoff |

> **When Commit Pending**: If the completion report contains "Commit Status: Pending PM Approval," the approve handoff **must include commit instructions** (see approve template below).

### Step 3: Generate Handoff (Required)

In either case, generate a handoff message for Claude Code.

---

## Output Format

### Judgment Summary

```
## Review Result

**Judgment**: approve / request_changes
**Reason**:
- (1-3 points)

**Plans.md Update**:
- `[Task name]` -> changed to `pm:confirmed` (if approved)
```

### Hand off to Claude (Always output)

#### On Approve: Commit and Finish

**Default behavior**: On approve, commit the changes and finish. Only generate a handoff for the next task if the user explicitly requests it.

##### Approval Only (Default)

Approve -> commit instruction -> **stop here**. No automatic transition to the next task.

~~~markdown
/claude-code-harness:core:work
<!-- ultrathink: PM requests are important tasks by default, always specify high effort -->
ultrathink

## Request

The previous task has been approved. Please commit the changes.

### Commit Instructions
- The previous changes are approved. Please commit.
- After committing, work is complete.

### References
- Related files (if any)

After committing, report with `/handoff-to-cursor`.
~~~

##### Only When User Explicitly Requests Next Task

Use the following template only when the user explicitly says "proceed to the next task," "continue," etc.:

Analyze the next `cc:TODO` or `pm:requested` task from @Plans.md and generate:

~~~markdown
/claude-code-harness:core:work
ultrathink

## Request

The previous task has been approved. **Commit the changes first**, then implement the next task.

### Commit Instructions
- The previous changes are approved. Please commit before proceeding to the next task.

### Target Task
- (Extract next task from Plans.md)

### Background
- Ready to start after previous task completion and approval
- (Note dependencies if any)

### Constraints
- Follow existing code style
- Keep changes to the minimum necessary
- Confirm tests/builds pass

### Acceptance Criteria
- (3-5 items, specific)

### References
- Related files (if any)

After completion, report with `/handoff-to-cursor`.
~~~

#### On Request Changes

Generate a handoff with change instructions:

~~~markdown
/claude-code-harness:core:work
ultrathink

## Change Request

The review found the following changes are needed.

### Target Task
- (Applicable task from Plans.md)

### Issues Found
1. **[Severity: High/Medium/Low]** Issue description
   - Location: `filename:line number`
   - Expected fix: Specific resolution approach

2. **[Severity: High/Medium/Low]** Issue description
   - Location:
   - Expected fix:

### Constraints
- Do not break existing tests
- Do not change anything outside the flagged areas

### Acceptance Criteria (After Fix)
- All issues above are resolved
- Tests/builds pass
- (Additional criteria if any)

After completion, report with `/handoff-to-cursor`.
~~~

---

## Workflow Diagram

```
Claude Code completion report
        |
  /review-cc-work
        |
   +----+----+
   |         |
approve   request_changes
   |         |
pm:confirmed  Create change request
   |         |
commit      Generate handoff
   |         |
 Done        |
(next task      |
 only on        |
 explicit       |
 request)       |
   |         |
   +----+----+
        |
  Paste into Claude Code
        |
     /work execution
```
