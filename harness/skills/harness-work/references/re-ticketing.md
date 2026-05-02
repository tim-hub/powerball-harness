# Automatic Re-ticketing of Failed Tasks

Fix task auto-generation with approval flow when tests/CI fail after task completion — proposes a numbered `.fix` task to Plans.md pending user approval.

---

When tests/CI fail after task completion, auto-generate fix task proposals and reflect them in Plans.md after approval:

## Trigger Conditions

| Condition | Action |
|-----------|--------|
| Test failure after `cc:Done` | Save fix task proposal to state and wait for approval |
| CI failure (fewer than 3 times) | Implement fix and increment failure count |
| CI failure (3rd time) | Present fix task proposal + escalate |

## Auto-Generation of Fix Tasks

1. Classify failure cause (syntax_error / import_error / type_error / assertion_error / timeout / runtime_error)
2. Save fix task proposal to `.claude/state/pending-fix-proposals.jsonl`:
   - Number: Original task number + `.fix` suffix (e.g., `26.1.fix`)
   - Content: `fix: [original task name] - [failure cause category]`
   - DoD: Tests/CI pass
   - Depends: Original task number
3. When user sends `approve fix <task_id>`, add to Plans.md as `cc:TODO`
4. `reject fix <task_id>` discards the proposal. When there is only one pending item, `yes` / `no` responses are also accepted
