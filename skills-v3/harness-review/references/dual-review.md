# Dual Review (--dual)

Run Claude Reviewer and Codex Reviewer in parallel to improve review quality through different model perspectives.

## Prerequisites

- Codex CLI is installed (verify with `scripts/codex-companion.sh setup --json`)
- If Codex is unavailable, falls back to Claude-only review

## Execution Flow

1. Check Codex availability

   ```bash
   CODEX_AVAILABLE="$(bash scripts/codex-companion.sh setup --json 2>/dev/null | jq -r '.ready // false')"
   ```

2. Launch Claude Reviewer via Task tool (normal review flow)

3. If Codex is available, launch `scripts/codex-companion.sh review` in parallel

   ```bash
   # Specify --base when BASE_REF is provided. --json for structured output
   bash scripts/codex-companion.sh review --base "${BASE_REF:-HEAD~1}" --json
   ```

4. Wait for both results

5. Verdict merge rules (evaluated in order):
   - Both APPROVE -> `APPROVE`
   - Either is REQUEST_CHANGES -> `REQUEST_CHANGES` (adopt the stricter verdict)
   - `critical_issues` are merged from both lists (no deduplication)
   - `major_issues` are merged from both lists (no deduplication)
   - `recommendations` are merged with deduplication

## Output Format

The standard `review-result.v1` schema with an added `dual_review` field:

```json
{
  "schema_version": "review-result.v1",
  "verdict": "APPROVE | REQUEST_CHANGES",
  "dual_review": {
    "claude_verdict": "APPROVE | REQUEST_CHANGES",
    "codex_verdict": "APPROVE | REQUEST_CHANGES | unavailable | timeout",
    "merged_verdict": "APPROVE | REQUEST_CHANGES",
    "divergence_notes": "Reason when verdicts diverge. Example: Claude found a major in Performance, Codex found no issues"
  },
  "critical_issues": [],
  "major_issues": [],
  "observations": [],
  "recommendations": []
}
```

### Special `codex_verdict` Values

| Value | Meaning |
|----|------|
| `"unavailable"` | Codex CLI is not installed or unavailable |
| `"timeout"` | Codex review timed out (no response within 120 seconds) |

## Fallback

- **Codex unavailable**: Execute Claude-only review and record `codex_verdict: "unavailable"`
- **Codex timeout**: Adopt Claude's verdict as-is and record `codex_verdict: "timeout"`
- **Invalid Codex review output**: Treated as parse failure and record `codex_verdict: "unavailable"`

In all fallback cases, the Claude-only review result becomes the final verdict.

## Divergence Notes Format

When verdicts match (`claude_verdict == codex_verdict`), set `divergence_notes` to an empty string.

When verdicts diverge, record in the following format:

```
Claude: REQUEST_CHANGES (Security - SQL injection risk)
Codex: APPROVE (judged the same location as no issue)
Adopted: REQUEST_CHANGES (prefer the stricter verdict)
```
