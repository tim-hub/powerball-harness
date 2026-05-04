# Review Result Schema (review-result.v1)

Output JSON for all review types:

```json
{
  "schema_version": "review-result.v1",
  "verdict": "APPROVE | REQUEST_CHANGES",
  "verdict_reasoning": {
    "rationale": "Explanation of why this verdict was selected",
    "triggering_issues": ["List of critical/major findings that triggered REQUEST_CHANGES, if applicable"],
    "confidence": "high | medium"
  },
  "reviewer_profile": "static | runtime | browser",
  "critical_issues": [
    {
      "category": "Security | Performance | Quality | Accessibility | AI Residuals",
      "location": "filename:line_number",
      "issue": "issue description",
      "severity_justification": "Why this is critical",
      "suggestion": "fix suggestion"
    }
  ],
  "major_issues": [
    {
      "category": "Security | Performance | Quality | Accessibility | AI Residuals",
      "location": "filename:line_number",
      "issue": "issue description",
      "severity_justification": "Why this is major",
      "suggestion": "fix suggestion",
      "scope": "task-specific | universal"
    }
  ],
  "observations": [
    {
      "severity": "minor | recommendation",
      "category": "Security | Performance | Quality | Accessibility | AI Residuals",
      "location": "filename:line_number",
      "issue": "issue description",
      "suggestion": "fix suggestion",
      "scope": "task-specific | universal"
    }
  ],
  "recommendations": ["non-mandatory improvement suggestions"],
  "calibration": {
    "label": "false_positive | false_negative | missed_bug | overstrict_rule",
    "source": "manual | post-review | retrospective",
    "notes": "observation memo",
    "prompt_hint": "key points for few-shot",
    "few_shot_ready": true
  }
}
```

For browser reviews, `generate-browser-review-artifact.sh` determines `browser_mode` and route/artifacts.
`write-review-result.sh` normalizes and saves to `.claude/state/review-result.json`.
Calibration entries with `calibration` are appended to `.claude/state/review-calibration.jsonl` via `record-review-calibration.sh`; the few-shot bank is updated via `build-review-few-shot-bank.sh`.

For `--ui-rubric` flag, add this field:

```json
{
  "ui_rubric": {
    "design_quality": { "score": 8, "observations": ["..."] },
    "originality":    { "score": 6, "observations": ["..."] },
    "craft":          { "score": 9, "observations": ["..."] },
    "functionality":  { "score": 7, "observations": ["..."] }
  }
}
```

UI rubric scores are informational and do not affect the verdict unless `functionality` score ≤ 3.

For `--dual` flag, add a `dual_review` field — see [`${CLAUDE_SKILL_DIR}/references/dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md).
