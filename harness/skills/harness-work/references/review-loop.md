# Review Loop

Codex exec first / internal Reviewer agent fallback with severity-based verdict criteria and a 3-cycle fix cap — applied uniformly across solo, parallel, and breezing modes.

---

A quality verification stage that runs automatically after implementation completion (after step 5).
Applied uniformly across **all modes** (Solo / Parallel / Breezing).
In Parallel mode, each Worker executes the same loop as step 10 (external review acceptance).

## Review Execution Priority

```
1. Codex exec (priority, when available) — see ${CLAUDE_SKILL_DIR}/references/codex-work.md
   ↓ codex command does not exist or timeout (120s)
2. Internal Reviewer agent (fallback)
```

## APPROVE / REQUEST_CHANGES Verdict Criteria

The following threshold criteria are provided to reviewers, and the verdict is determined **solely by these criteria**.
Improvement suggestions outside these criteria are returned as `recommendations` but do not affect the verdict.

| Severity | Definition | Verdict Impact |
|----------|------------|----------------|
| **critical** | Security vulnerabilities, data loss risk, potential production incidents | 1 item → REQUEST_CHANGES |
| **major** | Breaking existing functionality, clear contradiction with specifications, test failures | 1 item → REQUEST_CHANGES |
| **minor** | Naming improvements, insufficient comments, style inconsistencies | No impact on verdict |
| **recommendation** | Best practice suggestions, future improvement ideas | No impact on verdict |

> **Important**: When only minor / recommendation items exist, **always return APPROVE**.
> "Nice-to-have improvements" are not grounds for REQUEST_CHANGES.

## Codex Exec Review (via official plugin)

> When Codex is available, load [`${CLAUDE_SKILL_DIR}/references/codex-work.md`](${CLAUDE_SKILL_DIR}/references/codex-work.md)
> for the full Codex exec review flow, verdict mapping, and AI Residuals scan details.

## Internal Reviewer Agent Fallback

When Codex exec is unavailable (`command -v codex` fails, or exit code != 0):

```
Agent tool: subagent_type="reviewer"
prompt: "Please review the following changes. Verdict criteria: critical/major → REQUEST_CHANGES, minor/recommendation only → APPROVE. diff: {git diff ${BASE_REF}}"
```

The Reviewer agent executes reviews safely in Read-only mode (Write/Edit/Bash disabled).

## Fix Loop (on REQUEST_CHANGES)

```
review_count = 0
MAX_REVIEWS = 3

while verdict == "REQUEST_CHANGES" and review_count < MAX_REVIEWS:
    1. Analyze review findings (critical / major only)
    2. Implement fixes for each finding
    3. Re-execute review (same criteria, same priority)
    review_count++

if review_count >= MAX_REVIEWS and verdict != "APPROVE":
    → Escalate to user
    → Display "Fixed 3 times but the following critical/major issues remain" + list of issues
    → Wait for user decision (continue / abort)
```

## Application in Breezing Mode

In Breezing mode, the **Lead** executes the review loop (see [`breezing-mode.md`](${CLAUDE_SKILL_DIR}/references/breezing-mode.md) Phase B):

1. Worker implements and commits in worktree → returns result to Lead
2. Lead reviews with Codex exec (priority) / Reviewer agent (fallback)
3. REQUEST_CHANGES → Lead sends fix instructions to Worker via SendMessage → Worker amends
4. After fix, re-review (up to 3 times)
5. APPROVE → Lead cherry-picks to main → Updates Plans.md to `cc:Done [{hash}]`
