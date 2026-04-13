---
name: harness-review
description: "Use this skill whenever the user asks to review code, review a plan, check quality, analyze scope, run security checks, examine PRs or diffs, or runs /harness-review. Also use when the user wants a second opinion on changes, performance review, or pre-merge quality gate. Do NOT load for: code implementation (use harness-work), new features, bug fixes, project setup, or release. Multi-angle code, plan, and scope review with optional dual-reviewer and security analysis."
allowed-tools: ["Read", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[code|plan|scope] [--dual] [--security]"
context: fork
effort: high
model: opus
---

# Harness Review

## Quick Reference

| User Input | Subcommand | Behavior |
|------------|------------|----------|
| "Review this" / "review" | `code` (auto) | Code review (recent changes) |
| "After `harness-plan`" | `plan` (auto) | Plan review |
| "Check scope" | `scope` (auto) | Scope analysis |
| `harness-review code` | `code` | Force code review |
| `harness-review plan` | `plan` | Force plan review |
| `harness-review scope` | `scope` | Force scope analysis |
| `harness-review --dual` | `code` (auto) + Codex parallel | Claude + Codex dual review |
| `harness-review --security` | Security Review | OWASP Top 10 dedicated security review (read-only) |

## Options

| Option | Default | Description |
|--------|---------|-------------|
| `--dual` | none | Run Claude Reviewer and Codex Reviewer in parallel and merge verdicts. Auto-fallback when Codex is unavailable. Details: [`${CLAUDE_SKILL_DIR}/references/dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md) |
| `--security` | none | Execute OWASP Top 10-based security-only review. Read-only (no Write/Edit/Bash writes). Details: [`${CLAUDE_SKILL_DIR}/references/security-profile.md`](${CLAUDE_SKILL_DIR}/references/security-profile.md) |
| `--no-commit` | none | Disable auto-commit on APPROVE |

## Review Type Auto-Detection

| Recent Activity | Review Type | Perspectives |
|-----------------|-------------|--------------|
| After `harness-work` | **Code Review** | Security, Performance, Quality, Accessibility, AI Residuals |
| After `harness-plan` | **Plan Review** | Clarity, Feasibility, Dependencies, Acceptance |
| After task addition | **Scope Review** | Scope-creep, Priority, Feasibility, Impact |

## Code Review Flow

### Step 1: Collect Change Diff

```bash
# Use BASE_REF from harness-work if available, otherwise fall back to HEAD~1
CHANGED_FILES="$(git diff --name-only --diff-filter=ACMR "${BASE_REF:-HEAD~1}")"
git diff ${BASE_REF:-HEAD~1} --stat
git diff ${BASE_REF:-HEAD~1} -- ${CHANGED_FILES}
```

### Step 1.5: Static Scan for AI Residuals

Rather than relying solely on LLM impressions, pick up residual candidates in a reproducible way. `scripts/review-ai-residuals.sh` returns stable JSON, which is used as review evidence.

```bash
# Diff-based
AI_RESIDUALS_JSON="$(bash scripts/review-ai-residuals.sh --base-ref "${BASE_REF:-HEAD~1}")"

# To explicitly specify target files
bash scripts/review-ai-residuals.sh path/to/file.ts path/to/config.sh
```

### Step 2: Review from 5 Perspectives

| Perspective | Check Items |
|-------------|-------------|
| **Security** | SQL injection, XSS, credential exposure, input validation |
| **Performance** | N+1 queries, unnecessary re-renders, memory leaks |
| **Quality** | Naming, single responsibility, test coverage, error handling |
| **Accessibility** | ARIA attributes, keyboard navigation, color contrast |
| **AI Residuals** | `mockData`, `dummy`, `fake`, `localhost`, `TODO`, `FIXME`, `it.skip`, `describe.skip`, `test.skip`, hardcoded secrets/environment-dependent URLs, obvious placeholder implementation comments |

### Step 2.2: AI Residuals Severity Classification Table

For `AI Residuals`, first check the JSON from `scripts/review-ai-residuals.sh`, then make the final judgment of "is this truly a shipping risk?" from the diff context.

| Severity | Representative Examples | Classification Rationale |
|----------|------------------------|--------------------------|
| **major** | `localhost` / `127.0.0.1` / `0.0.0.0` connection targets, `it.skip` / `describe.skip` / `test.skip`, hardcoded secret-like values, dev/staging fixed URLs | Directly linked to production incidents, misconfiguration, or missed validation. 1 item → `REQUEST_CHANGES` |
| **minor** | `mockData`, `dummy`, `fakeData`, `TODO`, `FIXME` | Likely residuals, but not necessarily an immediate incident. Fix recommended but does not change verdict |
| **recommendation** | `temporary implementation`, `replace later`, `placeholder implementation` comments | Cannot be immediately classified as a bug from comments alone, but should be tracked and clarified |

### Step 2.5: Verdict Determination by Threshold Criteria

Classify each finding by the following severity levels and determine the verdict **solely by these criteria**.

| Severity | Definition | Verdict Impact |
|----------|------------|----------------|
| **critical** | Security vulnerabilities, data loss risk, potential production incidents | 1 item → REQUEST_CHANGES |
| **major** | Breaking existing functionality, clear contradiction with specifications, test failures | 1 item → REQUEST_CHANGES |
| **minor** | Naming improvements, insufficient comments, style inconsistencies | No impact on verdict |
| **recommendation** | Best practice suggestions, future improvement ideas | No impact on verdict |

> **Important**: When only minor / recommendation items exist, **always return APPROVE**.
> "Nice-to-have improvements" are not grounds for REQUEST_CHANGES.
> The same applies to `AI Residuals`. Only items that are "directly linked to shipping incidents or misconfiguration" are classified as `major`; mere residual candidates are kept at `minor` or `recommendation`.

### Step 3: Review Result Output

```json
{
  "schema_version": "review-result.v1",
  "verdict": "APPROVE | REQUEST_CHANGES",
  "reviewer_profile": "static | runtime | browser",
  "calibration": {
    "label": "false_positive | false_negative | missed_bug | overstrict_rule",
    "source": "manual | post-review | retrospective",
    "notes": "observation memo",
    "prompt_hint": "key points for few-shot",
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

For browser reviews, `scripts/generate-browser-review-artifact.sh` determines `browser_mode` and route / required artifacts, then `scripts/write-review-result.sh` normalizes and saves to `.claude/state/review-result.json`.
This file serves as the shared input for the commit guard and downstream flows.
Review results with `calibration` are appended to `.claude/state/review-calibration.jsonl` via `scripts/record-review-calibration.sh`, and the few-shot bank is updated via `scripts/build-review-few-shot-bank.sh`.

### Step 3.5: Codex Parallel Review with --dual Flag

When the `--dual` flag is specified, run a Codex review in parallel with the Claude review in Step 3, then merge the results.

1. Check Codex availability (`scripts/codex-companion.sh setup --json`)
2. If available, launch `scripts/codex-companion.sh review --base "${BASE_REF:-HEAD~1}"`
3. Integrate both verdicts using the Verdict Merge Rules
4. Add a `dual_review` field to the final review result

For detailed procedures, output schema, and fallback specifications, see [`${CLAUDE_SKILL_DIR}/references/dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md).

### Step 3.6: Security-Only Review with --security Flag

When the `--security` flag is specified, **skip** the standard 5-perspective review and execute the security-only flow.

**Read-only constraint**: No Write / Edit / write-mode Bash operations are executed during this flow.

1. Load the security profile:
   ```
   Read: ${CLAUDE_SKILL_DIR}/references/security-profile.md
   ```
2. Check all OWASP Top 10 categories against the change diff and related files
3. Check authentication/authorization flows, secret handling, and dependency vulnerabilities
4. Set `reviewer_profile: "security"` and output results (conforming to Step 3's JSON schema)
5. Apply the Security mode verdict criteria (see end of security-profile.md)

Choosing between standard Code Review and `--security`:

| | Standard Code Review | `--security` |
|---|---|---|
| Perspectives | Security, Performance, Quality, Accessibility, AI Residuals | Security only (all OWASP Top 10 items) |
| Depth | Security is overview-level | Comprehensive coverage of auth, authorization, encryption, dependencies |
| Tool restrictions | None | Read / Grep / Glob / read-only Bash only |
| Use case | Pre-merge comprehensive check | Security-focused audit, additional pre-release verification |

### Step 4: Commit Decision

- **APPROVE**: Execute auto-commit (unless `--no-commit`)
- **REQUEST_CHANGES**: Present critical/major findings and fix strategy. Auto-fix via `harness-work`'s fix loop followed by re-review (up to 3 times)

## Plan Review Flow

1. Read Plans.md
2. Review from the following **5 perspectives**:
   - **Clarity**: Are task descriptions clear?
   - **Feasibility**: Is it technically feasible?
   - **Dependencies**: Are inter-task dependencies correct? (Do Depends column entries match actual dependencies?)
   - **Acceptance**: Are completion criteria (DoD column) defined and verifiable?
   - **Value**: Does this task solve a user problem?
     - Is "whose problem" explicitly stated?
     - Were alternatives (including not building it) considered?
     - Are there Elephants (problems everyone notices but no one addresses)?
3. DoD / Depends column quality checks:
   - Tasks with empty DoD → Warning ("Completion criteria undefined")
   - DoD that is unverifiable ("looks good", "works properly", etc.) → Warning + concretization suggestion
   - Depends referencing non-existent task numbers → Error
   - Circular dependencies → Error
4. Present improvement suggestions

## Scope Review Flow

1. List added tasks/features
2. Analyze from the following perspectives:
   - **Scope-creep**: Deviation from original scope
   - **Priority**: Is the priority appropriate?
   - **Feasibility**: Achievable with current resources?
   - **Impact**: Impact on existing functionality
3. Present risks and recommended actions

## Anomaly Detection

| Situation | Action |
|-----------|--------|
| Security vulnerability | Immediately REQUEST_CHANGES |
| Suspected test tampering | Warning + fix request |
| Force push attempt | Reject + suggest alternative |

## Codex Environment

> Load [`${CLAUDE_SKILL_DIR}/references/codex-review.md`](${CLAUDE_SKILL_DIR}/references/codex-review.md)
> only when `command -v codex` succeeds **and** the user explicitly requests Codex or duo review.

## Related Skills

- `harness-work` — Implement fixes after review
- `harness-plan` — Create and modify plans
- `harness-release` — Release after review passes
