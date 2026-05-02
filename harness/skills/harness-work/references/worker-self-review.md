# Worker Self-Review Gate (worker-report.v1)

Schema and Lead-side validation rules for the worker-report.v1 JSON block the Worker must emit before the Lead spawns a Reviewer; up to 2 amendment cycles before escalation.

---

Before the Lead spawns a Reviewer, the Worker must emit a `worker-report.v1` JSON block as the final output of its implementation turn. The Lead validates the report and rejects incomplete submissions with up to 2 amendment cycles before escalating.

## Schema

See the canonical template at [`${CLAUDE_SKILL_DIR}/templates/worker-report.v1.json`](${CLAUDE_SKILL_DIR}/templates/worker-report.v1.json).

```json
{
  "schema": "worker-report.v1",
  "task_id": "<task number>",
  "self_review": [
    {
      "rule_id": "SR-1",
      "rule": "All DoD items from the sprint contract are addressed",
      "verified": true,
      "evidence": "<what was done / test output / file reference>"
    },
    {
      "rule_id": "SR-2",
      "rule": "No NG rule violations: Plans.md untouched (NG-1), no embedded git repos (NG-2), no nested spawn (NG-3)",
      "verified": true,
      "evidence": "<git diff HEAD Plans.md output / find .git output>"
    },
    {
      "rule_id": "SR-3",
      "rule": "Tests pass or no test framework present and no regressions detected",
      "verified": true,
      "evidence": "<test run output or rationale>"
    },
    {
      "rule_id": "SR-4",
      "rule": "No AI residuals: no localhost/127.0.0.1, no it.skip/describe.skip/test.skip, no hardcoded secrets",
      "verified": true,
      "evidence": "<grep output or 'residuals scan: clean'>"
    },
    {
      "rule_id": "SR-5",
      "rule": "Commit is self-contained: no unrelated file changes, no debug artifacts, meaningful commit message",
      "verified": true,
      "evidence": "<git diff --stat output>"
    }
  ]
}
```

## Validation Rules (Lead-side)

A `worker-report.v1` is **valid** iff:
- All 5 `rule_id` entries (SR-1 through SR-5) are present
- Every entry has `"verified": true`
- Every `evidence` field is non-empty (non-empty string, not `""` or `null`)

If any rule has `"verified": false` or an empty `evidence` field, the report is **invalid** — Lead sends an amendment request (see B-2.5 in [`breezing-mode.md`](${CLAUDE_SKILL_DIR}/references/breezing-mode.md)). After 2 failed amendments, the task is escalated to the user; the Worker commit is not cherry-picked.
