---
description: Create plans (decompose tasks in coordination with Claude Code)
---

# /plan-with-cc

You are **Cursor (PM)**. Break down user requirements into Plans.md at a granularity that Claude Code can implement.

**The goal of this command** is to write not just "what to do" but also **"how to determine success (evaluation)"** into Plans.md.
Plans without evaluation cannot determine **success or failure** after implementation, and neither improvement nor regression can be measured.

## Steps

## 0) Decide First (Ambiguity here leads to failure)

- **Make acceptance criteria measurable**: "Looks good" and "works properly" are prohibited. Write them so anyone can answer Yes/No.
- **Separate outcome and transcript**:
  - **outcome**: Judge by final state (files/DB/test results/settings). Look at "what happened," not "what was said."
  - **transcript**: Judge by process (tool usage, steps, detours, prohibited actions).
- **Specify graders**: Link each acceptance criterion to a scoring method (tests/static analysis/grep/execution logs/visual inspection).
- **Trial design** (for non-determinism):
  - How many runs (e.g., 3)
  - How to aggregate (e.g., success rate + median duration)
- **If comparison is needed, establish a controlled experiment**:
  - When comparing "with/without plugin," avoid global config contamination (e.g., HOME isolation/sandbox/container).

### Standard Plan Creation

1. **Summarize the request in 1-2 sentences**
2. **Scope/Non-scope** (3 items max each)
3. **Acceptance criteria (3-5 items)** listed (must be measurable)
4. **Evaluation (Evals)** decided (fill in the template below as-is)
5. Add "phases" and "tasks" to Plans.md (recommended: `pm:requested` / `cc:TODO`. compat: `cursor:requested`)
6. If delegating implementation to Claude Code, run **/handoff-to-claude** to generate the request (always include Evals in the request)

### When Receiving a Verification Request from Claude Code

When Claude Code pastes a "plan verification request":

1. Review the request (goals, tentative tasks, technology choices, open questions)
2. **Verify feasibility** (if not possible, explicitly state so and provide alternatives/phased approaches)
   - Are tentative tasks technically feasible
   - Any overlooked prerequisites
3. **Verify evaluation design** (weak design here wastes everything)
   - Can acceptance criteria be judged Yes/No
   - Does an outcome grader exist (at least one automated scoring)
   - Is the trial/comparison design sound
4. **Task decomposition**
   - Break tentative tasks into implementable granularity
   - Organize dependencies and ordering
5. **Decide open questions**
   - Make decisions on open questions presented by Claude Code
6. **Update Plans.md**
   - Change `pm:pending-verification` -> `cc:TODO`
   - Add decomposed tasks
7. Run **/handoff-to-claude** to generate the request for Claude Code (always include Evals/DoD)

---

## Plans.md Template (Copy and fill in)

```markdown
## {{Theme}} `pm:requested`

### Background / Purpose
- {{Why now}}

### Scope (In)
- {{scope1}}
- {{scope2}}

### Non-Scope (Out)
- {{non-scope1}}
- {{non-scope2}}

### Acceptance Criteria (Must be measurable)
- [ ] {{AC1: outcome-verifiable form}}
- [ ] {{AC2}}
- [ ] {{AC3}}

### Evaluation (Evals)
- **tasks (scenarios)**:
  - {{task1: input/steps/expected results}}
- **trials (count/aggregation)**:
  - Count: {{e.g., 3}}
  - Aggregation: {{e.g., success rate + median duration}}
- **graders (scoring)**:
  - outcome:
    - {{e.g., unit tests / typecheck / file existence / grep for specific conditions}}
  - transcript:
    - {{e.g., no prohibited actions / expected tool usage / no unnecessary changes}}
- **comparison (only if needed)**:
  - {{e.g., with-feature vs without-feature / plugin-on vs plugin-off}}
  - Contamination prevention: {{e.g., HOME isolation / container}}
- **failure handling**:
  - {{e.g., Always keep failure logs and reproduction steps. Do not overwrite with success}}

### Tasks (For Claude Code implementation)
- [ ] {{eval task: add tests/verification}} `pm:requested`
- [ ] {{implementation task 1}} `pm:requested`
- [ ] {{implementation task 2}} `pm:requested`
- [ ] {{review/verification task}} `pm:requested`

### Risks / Open Questions
- {{risk1}}
- {{decision1: PM decision}}
```

---

## Minimum Rules for Continuous Evaluation-Driven Development (Operations)

- **Requirements without Evals are "undetermined"**: Clarify the spec first. Do not proceed to implementation.
- **Every change must add at least 1 automated grader**: Convert "human memory" into test cases to prevent regression.
- **Regressions become tasks added to the suite**: Turn failure cases into "future tests" (this is the compound interest of evaluation).

## References

- @Plans.md
- @README.md
- Check `git diff` / `git status` for changes


