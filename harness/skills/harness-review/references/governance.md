# Review Governance

## Summary

Only return `APPROVE` when you can say with evidence that there are no critical issues.

## Pass criteria

`APPROVE` conditions:

- Zero critical or major findings
- No contradiction with the spec source-of-truth (`spec_path`) or `spec_skip_reason`
- No contradiction with `Plans.md` task / DoD / Depends
- No evidence of regression in existing behavior, tests, UX, CLI, configuration, docs, or distribution mirrors
- Evidence exists: verification commands, diffs, file:line citations, test results
- No unresolved TeamAgent Debate disagreements

## Severity

| Severity | Meaning | Verdict |
|---|---|---|
| critical | Secret exposure, data destruction, permission destruction, directly causes a release incident | REQUEST_CHANGES |
| major | DoD not met, spec source-of-truth violation, clear regression, dangerous without test execution | REQUEST_CHANGES |
| minor | Quality improves but not a shipping blocker | APPROVE allowed |
| recommendation | Optional improvement | APPROVE allowed |

If findings are minor / recommendation only, blocking is not required.
If blocking, explain concretely why the finding is major, not just minor.

## AskUserQuestion / decision_needed

For decisions that would break things if guessed wrong, use `decision_needed` instead of `REQUEST_CHANGES`.

`decision_needed` examples:

- The spec source-of-truth needs to be changed
- A `Plans.md` DoD / Depends needs to be changed
- The user must choose between security and UX priority
- Whether to preserve or remove backward compatibility is a business decision

Use `AskUserQuestion` when available.
In Codex environments where it is unavailable, emit `decision_needed.v1` to stdout and do not proceed on a guess.

## Side effects

Review default read-only boundary:

- Do not auto-commit even on `APPROVE`
- Do not push just to review
- commit / push / release is the responsibility of `harness-work` / `harness-release` / explicit user request

## Output evidence

Required:

- Target scope
- Review command executed
- Tests executed
- Accepted findings
- Rejected findings
- Clean result or remaining issues
- Pass/fail for spec source-of-truth / Plans.md / regression gate

An `APPROVE` with empty evidence is invalid.
