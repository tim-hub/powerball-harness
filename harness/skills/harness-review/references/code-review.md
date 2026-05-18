# Code Review Flow

## Summary

Collect the diff, examine implementation, spec, Plans, regressions, and tests, then block only the issues that must be blocked.

## Step 1: collect diff

Check:

```bash
git status --short
git diff --stat "${BASE_REF:-HEAD}"
git diff "${BASE_REF:-HEAD}"
git ls-files --others --exclude-standard
```

Untracked files do not appear in `git diff`. Always include them in scope.

## Step 2: static scans

AI Residuals:

```bash
bash scripts/review-ai-residuals.sh --base "${BASE_REF:-HEAD}"
bash scripts/review-weak-supervision-report.sh
```

Candidates:

- `mockData`
- `dummy`
- `fake`
- `localhost`
- `TODO`
- `FIXME`
- `it.skip`
- `describe.skip`
- `test.skip`
- `expect(true).toBe(true)`

Finding a candidate alone does not make it major. Assess in diff context whether it directly leads to a shipping incident or misconfiguration.

## Step 3: eight review lenses

| Lens | What to check |
|---|---|
| Security | SQL injection, cross-site scripting, secret leak, permission bypass |
| Performance | N+1, needless heavy IO, blocking work |
| Quality | duplicate logic, unclear boundary, fragile parsing |
| Accessibility | labels, focus, contrast, keyboard path |
| AI Residuals | fake success, skipped tests, mock-only implementation |
| Spec Alignment | Contradictions with the spec source-of-truth |
| Plans Alignment | Alignment with `Plans.md` task / DoD / Depends |
| Regression Safety | Regressions in existing behavior, mirrors, CLI/skill UX |

## TDD compliance

For tasks that require TDD, look for evidence that a failing test was confirmed first.
For docs-only or refactor-only tasks where TDD is excessive, recording the skip reason is sufficient.

## Verdict

1. critical / major present → `REQUEST_CHANGES`
2. Spec source-of-truth / `Plans.md` / regression gate fails → `REQUEST_CHANGES`
3. Decision needed from the user → `decision_needed`
4. Minor / recommendation only → `APPROVE`
5. Insufficient evidence → `REQUEST_CHANGES` or `decision_needed`

## Post-fix re-review

After `REQUEST_CHANGES`, always re-review once fixes are applied.
If the same issue fails two consecutive post-fix re-reviews, force TeamAgent Debate.
