---
name: harness-review
description: "Use this skill whenever the user asks to review code, review a plan, check quality, analyze scope, run security checks, examine PRs or diffs, or runs /harness-review. Also use when the user wants a second opinion on changes, performance review, or pre-merge quality gate. Do NOT load for: code implementation (use harness-work), new features, bug fixes, project setup, or release. Unified review skill for Harness v3 — multi-angle code, plan, and scope review with optional dual-reviewer and security analysis."
allowed-tools: ["Read", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[code|plan|scope] [--dual] [--security]"
context: fork
effort: high
---

# Harness Review (v3)

Unified review skill for Harness v3.
Consolidates the following legacy skills:

- `harness-review` — Multi-angle code, plan, and scope review
- `codex-review` — Second opinion via Codex CLI
- `verify` — Build verification, error recovery, and review fix application
- `troubleshoot` — Error and failure diagnosis and repair

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|------|
| "review this" / "review" | `code` (auto) | Code review (recent changes) |
| "after `harness-plan`" | `plan` (auto) | Plan review |
| "check scope" | `scope` (auto) | Scope analysis |
| `harness-review code` | `code` | Force code review |
| `harness-review plan` | `plan` | Force plan review |
| `harness-review scope` | `scope` | Force scope analysis |
| `harness-review --dual` | `code` (auto) + Codex parallel | Claude + Codex dual review |
| `harness-review --security` | Security Review | OWASP Top 10 dedicated security review (read-only) |

## Options

| Option | Default | Description |
|-----------|-----------|------|
| `--dual` | none | Run Claude Reviewer and Codex Reviewer in parallel and merge verdicts. Auto-fallback when Codex is unavailable. Details: [`${CLAUDE_SKILL_DIR}/references/dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md) |
| `--security` | none | Run OWASP Top 10-based security-only review. Read-only (no Write/Edit/write Bash). Details: [`${CLAUDE_SKILL_DIR}/references/security-profile.md`](${CLAUDE_SKILL_DIR}/references/security-profile.md) |
| `--no-commit` | none | Disable auto-commit on APPROVE |

## Review Type Auto-Detection

| Recent Activity | Review Type | Perspective |
|--------------------|--------------|------|
| After `harness-work` | **Code Review** | Security, Performance, Quality, Accessibility, AI Residuals |
| After `harness-plan` | **Plan Review** | Clarity, Feasibility, Dependencies, Acceptance |
| After task addition | **Scope Review** | Scope-creep, Priority, Feasibility, Impact |

## Code Review Flow

### Step 1: Collect Change Diff

```bash
# Use BASE_REF if passed from harness-work, otherwise fall back to HEAD~1
CHANGED_FILES="$(git diff --name-only --diff-filter=ACMR "${BASE_REF:-HEAD~1}")"
git diff ${BASE_REF:-HEAD~1} --stat
git diff ${BASE_REF:-HEAD~1} -- ${CHANGED_FILES}
```

### Step 1.5: Static Scan for AI Residuals

Do not rely solely on LLM impressions; collect residual candidates in a reproducible manner. `scripts/review-ai-residuals.sh` returns stable JSON, which serves as review evidence.

```bash
# Diff-based
AI_RESIDUALS_JSON="$(bash scripts/review-ai-residuals.sh --base-ref "${BASE_REF:-HEAD~1}")"

# When specifying target files explicitly
bash scripts/review-ai-residuals.sh path/to/file.ts path/to/config.sh
```

### Step 2: Review from 5 Perspectives

| Perspective | Check Items |
|------|------------|
| **Security** | SQL injection, XSS, sensitive data exposure, input validation |
| **Performance** | N+1 queries, unnecessary re-renders, memory leaks |
| **Quality** | Naming, single responsibility, test coverage, error handling |
| **Accessibility** | ARIA attributes, keyboard navigation, color contrast |
| **AI Residuals** | `mockData`, `dummy`, `fake`, `localhost`, `TODO`, `FIXME`, `it.skip`, `describe.skip`, `test.skip`, hardcoded secrets/environment-specific URLs, obvious placeholder implementation comments |

### Step 2.2: AI Residuals Severity Classification Table

For `AI Residuals`, first check the JSON from `scripts/review-ai-residuals.sh`, then make a final judgment in the diff context to determine "is this truly a shipping risk?"

| Severity | Examples | Classification Rationale |
|--------|--------|-------------|
| **major** | `localhost` / `127.0.0.1` / `0.0.0.0` connections, `it.skip` / `describe.skip` / `test.skip`, hardcoded secret-like values, dev/staging hardcoded URLs | Likely to cause production incidents, misconfigurations, or missed tests. 1 item triggers `REQUEST_CHANGES` |
| **minor** | `mockData`, `dummy`, `fakeData`, `TODO`, `FIXME` | Likely residuals but not immediately incident-causing. Fix recommended but does not change verdict |
| **recommendation** | `temporary implementation`, `replace later`, `placeholder implementation` comments | Comments alone cannot be definitively classified as bugs, but should be tracked and clarified |

### Step 2.5: Verdict Determination by Threshold Criteria

Classify each finding by the following severity levels and determine the verdict **solely by these criteria**.

| Severity | Definition | Verdict Impact |
|--------|------|-----------------|
| **critical** | Security vulnerability, data loss risk, potential production outage | 1 item -> REQUEST_CHANGES |
| **major** | Breaking existing functionality, clear specification contradiction, failing tests | 1 item -> REQUEST_CHANGES |
| **minor** | Naming improvements, insufficient comments, style inconsistency | Does not affect verdict |
| **recommendation** | Best practice suggestions, future improvement proposals | Does not affect verdict |

> **Important**: When only minor / recommendation items exist, **always return APPROVE**.
> "Nice-to-have improvements" are not grounds for REQUEST_CHANGES.
> The same applies to `AI Residuals`. Only items that "directly cause shipping incidents or misconfigurations" qualify as `major`; mere residual candidates stay at `minor` or `recommendation`.

### Step 3: Review Result Output

```json
{
  "schema_version": "review-result.v1",
  "verdict": "APPROVE | REQUEST_CHANGES",
  "reviewer_profile": "static | runtime | browser",
  "calibration": {
    "label": "false_positive | false_negative | missed_bug | overstrict_rule",
    "source": "manual | post-review | retrospective",
    "notes": "observation notes",
    "prompt_hint": "key point for few-shot",
    "few_shot_ready": true
  },
  "critical_issues": [],
  "major_issues": [],
  "observations": [
    {
      "severity": "critical | major | minor | recommendation",
      "category": "Security | Performance | Quality | Accessibility | AI Residuals",
      "location": "filename:line_number",
      "issue": "issue description",
      "suggestion": "fix suggestion"
    }
  ],
  "recommendations": ["non-mandatory improvement suggestions"]
}
```

For browser reviews, `scripts/generate-browser-review-artifact.sh` determines the `browser_mode` and route / required artifacts, then `scripts/write-review-result.sh` normalizes and saves to `.claude/state/review-result.json`.
This file serves as the common input for the commit guard and subsequent flows.
Review results with `calibration` are appended to `.claude/state/review-calibration.jsonl` via `scripts/record-review-calibration.sh`, and the few-shot bank is updated via `scripts/build-review-few-shot-bank.sh`.

### Step 3.5: Codex Parallel Review with --dual Flag

When the `--dual` flag is specified, Codex review runs in parallel with Step 3's Claude review, and results are merged.

1. Check Codex availability (`scripts/codex-companion.sh setup --json`)
2. If available, launch `scripts/codex-companion.sh review --base "${BASE_REF:-HEAD~1}"`
3. Merge both verdicts using the Verdict Merge Rules
4. Add the `dual_review` field to the final review result

For detailed procedures, output schema, and fallback specifications, see [`${CLAUDE_SKILL_DIR}/references/dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md).

### Step 3.6: Security-Only Review with --security Flag

When the `--security` flag is specified, **skip** the normal 5-perspective review and execute the security-only flow.

**Read-only constraint**: During this flow, no Write / Edit / write Bash commands are executed.

1. Load the security profile:
   ```
   Read: ${CLAUDE_SKILL_DIR}/references/security-profile.md
   ```
2. Check all OWASP Top 10 categories against the change diff and related files
3. Check authentication/authorization flows, secret handling, and dependency package vulnerabilities
4. Set `reviewer_profile: "security"` and output results (conforming to Step 3 JSON schema)
5. Apply Security mode verdict criteria (see end of security-profile.md)

Choosing between normal Code Review and `--security`:

| | Normal Code Review | `--security` |
|---|---|---|
| Perspectives | Security, Performance, Quality, Accessibility, AI Residuals | Security only (all OWASP Top 10 items) |
| Depth | Overview check for security | Comprehensive coverage of auth, authorization, encryption, dependencies |
| Tool restrictions | None | Read / Grep / Glob / read-only Bash only |
| Use case | General pre-merge verification | Focused security audit, pre-release additional verification |

### Step 4: Commit Decision

- **APPROVE**: Execute auto-commit (unless `--no-commit`)
- **REQUEST_CHANGES**: Present critical/major findings and fix strategies. Auto-fix via `harness-work`'s fix loop, then re-review (max 3 times)

## Plan Review Flow

1. Read Plans.md
2. Review from the following **5 perspectives**:
   - **Clarity**: Are task descriptions clear?
   - **Feasibility**: Is it technically feasible?
   - **Dependencies**: Are inter-task dependencies correct? (Does the Depends column match actual dependencies?)
   - **Acceptance**: Are completion criteria (DoD column) defined and verifiable?
   - **Value**: Does this task solve a user problem?
     - Is "whose problem" clearly stated?
     - Have alternatives (the option of not building) been considered?
     - Are there Elephants (problems everyone notices but ignores)?
3. DoD / Depends column quality checks:
   - Tasks with empty DoD -> Warning ("Completion criteria undefined")
   - Unverifiable DoD ("looks good", "works properly", etc.) -> Warning + suggestion for specifics
   - Depends references non-existent task numbers -> Error
   - Circular dependencies -> Error
4. Present improvement suggestions

## Scope Review Flow

1. List added tasks/features
2. Analyze from the following perspectives:
   - **Scope-creep**: Deviation from original scope
   - **Priority**: Is the priority appropriate?
   - **Feasibility**: Achievable with current resources?
   - **Impact**: Impact on existing features
3. Present risks and recommended actions

## Anomaly Detection

| Situation | Action |
|------|----------|
| Security vulnerability | Immediately REQUEST_CHANGES |
| Suspected test tampering | Warning + fix request |
| Force push attempt | Reject + suggest alternative |

## Codex Environment

In Codex CLI environments (`CODEX_CLI=1`), some tools are unavailable, so the following fallbacks are used.

| Normal Environment | Codex Fallback |
|---------|-------------------|
| `TaskList` for task listing | `Read` Plans.md and check WIP/TODO tasks |
| `TaskUpdate` for status update | Directly `Edit` Plans.md markers (e.g., `cc:WIP` -> `cc:done`) |
| Write review results to Task | Output review results to stdout |

### Detection Method

```bash
if [ "${CODEX_CLI:-}" = "1" ]; then
  # Codex environment: Plans.md-based fallback
fi
```

### Review Output in Codex Environment

Since Task tool is not supported, review results are output in markdown format to standard output.
The Lead agent or user reads the results and decides the next action.

## Related Skills

- `harness-work` — Implement fixes after review
- `harness-plan` — Create or modify plans
- `harness-release` — Release after review passes
