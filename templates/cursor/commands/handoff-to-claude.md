---
description: Generate a work request prompt for Claude Code
---

# /handoff-to-claude

You are **Cursor (PM)**. Generate a request that can be copy-pasted directly to Claude Code.

## Input

- @Plans.md (to identify the target task)
- If possible, `git status -sb` and `git diff --name-only`

## Output (paste directly into Claude Code)

Output the following Markdown:

```markdown
/claude-code-harness:core:work
<!-- ultrathink: PM requests are important tasks by default, always specify high effort -->
ultrathink

## Request
Please implement the following.

- Target tasks:
  - (List applicable tasks from Plans.md)

## Constraints
- Follow existing code style
- Keep changes to the minimum necessary
- Provide test/build instructions if available

## Acceptance Criteria
- (3-5 items)

## Evals (Scoring/Verification)
Follow the "Evals" section in Plans.md and proceed in a way that **outcome/transcript can be scored**.

- tasks (scenarios):
  - (e.g., specific input/steps/expected results)
- trials (count/aggregation):
  - (e.g., 3 runs, success rate + median)
- graders (scoring):
  - outcome:
    - (e.g., unit tests / typecheck / file state)
  - transcript:
    - (e.g., no prohibited actions / no unnecessary changes)
- execution commands (if possible):
  - (e.g., `npm test`, `./tests/validate-plugin.sh`, etc.)

## References
- Related files (if any)

**After completion**: Run `/handoff-to-cursor` to submit the completion report
```


