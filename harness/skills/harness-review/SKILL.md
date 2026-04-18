---
name: harness-review
description: "Multi-angle code and plan review with security, scope, and UI profiles. Use when reviewing code, plans, PRs, or running pre-merge quality gates."
when_to_use: "review code, review plan, review PR, security audit, pre-merge check, scope analysis, quality gate"
allowed-tools: ["Read", "Grep", "Glob", "Bash", "Task"]
argument-hint: "[code|plan|scope|--dual|--security|--ui-rubric]"
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
| `harness-review --ui-rubric` | UI Rubric Review | 4-axis design quality scoring (Design Quality, Originality, Craft, Functionality) |

## Options

| Option | Default | Description |
|--------|---------|-------------|
| `--dual` | none | Run Claude Reviewer and Codex Reviewer in parallel and merge verdicts. Auto-fallback when Codex is unavailable. Details: [`${CLAUDE_SKILL_DIR}/references/dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md) |
| `--security` | none | Execute OWASP Top 10-based security-only review. Read-only (no Write/Edit/Bash writes). Details: [`${CLAUDE_SKILL_DIR}/references/security-profile.md`](${CLAUDE_SKILL_DIR}/references/security-profile.md) |
| `--ui-rubric` | none | Score UI changes on 4 axes (Design Quality, Originality, Craft, Functionality) using a 0–10 rubric. Details: [`${CLAUDE_SKILL_DIR}/references/ui-rubric.md`](${CLAUDE_SKILL_DIR}/references/ui-rubric.md) |
| `--no-commit` | none | Disable auto-commit on APPROVE |

## Review Type Auto-Detection

| Recent Activity | Review Type | Perspectives |
|-----------------|-------------|--------------|
| After `harness-work` | **Code Review** | Security, Performance, Quality, Accessibility, AI Residuals |
| After `harness-plan` | **Plan Review** | Clarity, Feasibility, Dependencies, Acceptance |
| After task addition | **Scope Review** | Scope-creep, Priority, Feasibility, Impact |

## Verdict Framework (Establish Before Review)

### Severity Classification

Before conducting the review, establish the severity framework that will determine the verdict:

| Severity | Definition | Verdict Impact |
|----------|------------|----------------|
| **critical** | Security vulnerabilities, data loss risk, potential production incidents | 1 item → REQUEST_CHANGES |
| **major** | Breaking existing functionality, clear contradiction with specifications, test failures | 1 item → REQUEST_CHANGES |
| **minor** | Naming improvements, insufficient comments, style inconsistencies | No impact on verdict |
| **recommendation** | Best practice suggestions, future improvement ideas | No impact on verdict |

### AI Residuals Severity Classification

For `AI Residuals` findings, classify using this mapping:

| Severity | Representative Examples | Classification Rationale |
|----------|------------------------|--------------------------|
| **major** | `localhost` / `127.0.0.1` / `0.0.0.0` connection targets, `it.skip` / `describe.skip` / `test.skip`, hardcoded secret-like values, dev/staging fixed URLs | Directly linked to production incidents, misconfiguration, or missed validation. 1 item → `REQUEST_CHANGES` |
| **minor** | `mockData`, `dummy`, `fakeData`, `TODO`, `FIXME` | Likely residuals, but not necessarily an immediate incident. Fix recommended but does not change verdict |
| **recommendation** | `temporary implementation`, `replace later`, `placeholder implementation` comments | Cannot be immediately classified as a bug from comments alone, but should be tracked and clarified |

### Verdict Decision Rule

The verdict is determined **solely** by the presence of critical or major findings:

- **If critical or major findings exist**: Verdict = REQUEST_CHANGES (with explicit reasoning citing which finding(s) triggered the change request)
- **If only minor and recommendation findings exist**: Verdict = APPROVE

## Code Review Flow

### Step 0: Reviewer Mode Auto-Detection (Browser vs Static)

> **Fork context auto-start (`REVIEW_AUTOSTART`)**: When invoked in a forked session (context: fork), emit `REVIEW_AUTOSTART` as the very first output token before any other processing. This signals to the parent session that the fork has initialised correctly and auto-review has begun. A parent waiting on this marker can proceed; a parent that never sees it after 10 seconds should re-invoke.
>
> **Fork-context prohibition list** — the following 5 failure modes are forbidden and must never occur in fork context:
> 1. Waiting for user confirmation before starting (auto-start is required — no prompts)
> 2. Returning an empty output if the diff is empty (output `{"verdict":"APPROVE","rationale":"no changes detected"}` instead)
> 3. Spawning a browser reviewer without first checking `reviewer_profile` in the sprint contract
> 4. Writing to `Plans.md` or any `cc:*` marker (review is read-only with respect to Plans.md)
> 5. Exiting without emitting a JSON verdict conforming to the Step 4 schema

Before collecting the diff, determine whether to use the browser reviewer or static reviewer:

```
Does the change include UI files (.tsx, .jsx, .vue, .css, .html)?
├─ No  → Static reviewer (proceed to Step 1)
└─ Yes → Does the sprint-contract specify reviewer_profile: "browser"?
    ├─ Yes → Browser reviewer (launch browser-review-runner.sh)
    └─ No  → Is this a visual/design change (layout, color, spacing)?
        ├─ Yes → Browser reviewer recommended (confirm with user if --ui-rubric not set)
        └─ No  → Static reviewer (proceed to Step 1)
```

**Browser reviewer path** (when `reviewer_profile: "browser"` or UI change detected):
```bash
bash "${CLAUDE_SKILL_DIR}/../../scripts/browser-review-runner.sh" --contract "${CONTRACT_PATH}"
```

The browser reviewer captures screenshots, verifies interaction flows, and outputs findings conforming to the Step 4 JSON schema. Build the few-shot bank after browser review:
```bash
bash "${CLAUDE_SKILL_DIR}/../../scripts/build-review-few-shot-bank.sh"
```

### Step 1: Collect Change Diff

```bash
# Use BASE_REF from harness-work if available, otherwise fall back to HEAD~1
CHANGED_FILES="$(git diff --name-only --diff-filter=ACMR "${BASE_REF:-HEAD~1}")"
git diff ${BASE_REF:-HEAD~1} --stat
git diff ${BASE_REF:-HEAD~1} -- ${CHANGED_FILES}
```

### Step 2: Static Scan for AI Residuals

Rather than relying solely on LLM impressions, pick up residual candidates in a reproducible way. `"${CLAUDE_SKILL_DIR}/../../scripts/review-ai-residuals.sh"` returns stable JSON, which is used as review evidence.

```bash
# Diff-based
AI_RESIDUALS_JSON="$(bash "${CLAUDE_SKILL_DIR}/../../scripts/review-ai-residuals.sh" --base-ref "${BASE_REF:-HEAD~1}")"

# To explicitly specify target files
bash "${CLAUDE_SKILL_DIR}/../../scripts/review-ai-residuals.sh" path/to/file.ts path/to/config.sh
```

### Step 3: Review from 5 Perspectives

| Perspective | Check Items |
|-------------|-------------|
| **Security** | SQL injection, XSS, credential exposure, input validation |
| **Performance** | N+1 queries, unnecessary re-renders, memory leaks |
| **Quality** | Naming, single responsibility, test coverage, error handling |
| **Accessibility** | ARIA attributes, keyboard navigation, color contrast |
| **AI Residuals** | `mockData`, `dummy`, `fake`, `localhost`, `TODO`, `FIXME`, `it.skip`, `describe.skip`, `test.skip`, hardcoded secrets/environment-dependent URLs, obvious placeholder implementation comments |

**During review**: Apply the severity framework established above. Classify each finding against the critical/major/minor/recommendation matrix. **Document your severity classification and rationale for each finding in the output.**

### Step 4: Review Result Output with Explicit Verdict Reasoning

```json
{
  "schema_version": "review-result.v1",
  "verdict": "APPROVE | REQUEST_CHANGES",
  "verdict_reasoning": {
    "rationale": "Explanation of why this verdict was selected (reference the severity framework above)",
    "triggering_issues": ["List of critical/major findings that triggered REQUEST_CHANGES, if applicable"],
    "confidence": "high | medium"
  },
  "reviewer_profile": "static | runtime | browser",
  "critical_issues": [
    {
      "category": "Security | Performance | Quality | Accessibility | AI Residuals",
      "location": "filename:line_number",
      "issue": "issue description",
      "severity_justification": "Why this is critical (reference severity framework)",
      "suggestion": "fix suggestion"
    }
  ],
  "major_issues": [
    {
      "category": "Security | Performance | Quality | Accessibility | AI Residuals",
      "location": "filename:line_number",
      "issue": "issue description",
      "severity_justification": "Why this is major (reference severity framework)",
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

For browser reviews, `"${CLAUDE_SKILL_DIR}/../../scripts/generate-browser-review-artifact.sh"` determines `browser_mode` and route / required artifacts, then `"${CLAUDE_SKILL_DIR}/../../scripts/write-review-result.sh"` normalizes and saves to `.claude/state/review-result.json`.
This file serves as the shared input for the commit guard and downstream flows.
Review results with `calibration` are appended to `.claude/state/review-calibration.jsonl` via `"${CLAUDE_SKILL_DIR}/../../scripts/record-review-calibration.sh"`, and the few-shot bank is updated via `"${CLAUDE_SKILL_DIR}/../../scripts/build-review-few-shot-bank.sh"`.

### Step 4.1: Codex Parallel Review with --dual Flag

When the `--dual` flag is specified, run a Codex review in parallel with the Claude review in Step 4, then merge the results.

1. Check Codex availability (`"${CLAUDE_SKILL_DIR}/../../scripts/codex-companion.sh" setup --json`)
2. If available, launch `"${CLAUDE_SKILL_DIR}/../../scripts/codex-companion.sh" review --base "${BASE_REF:-HEAD~1}"`
3. Integrate both verdicts using the Verdict Merge Rules
4. Add a `dual_review` field to the final review result

For detailed procedures, output schema, and fallback specifications, see [`${CLAUDE_SKILL_DIR}/references/dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md).

### Step 4.2: Security-Only Review with --security Flag

When the `--security` flag is specified, **skip** the standard 5-perspective review and execute the security-only flow.

**Read-only constraint**: No Write / Edit / write-mode Bash operations are executed during this flow.

1. Load the security profile:
   ```
   Read: ${CLAUDE_SKILL_DIR}/references/security-profile.md
   ```
2. Check all OWASP Top 10 categories against the change diff and related files
3. Check authentication/authorization flows, secret handling, and dependency vulnerabilities
4. Set `reviewer_profile: "security"` and output results (conforming to Step 4's JSON schema)
5. Apply the Security mode verdict criteria (see end of security-profile.md)

Choosing between standard Code Review and `--security`:

| | Standard Code Review | `--security` |
|---|---|---|
| Perspectives | Security, Performance, Quality, Accessibility, AI Residuals | Security only (all OWASP Top 10 items) |
| Depth | Security is overview-level | Comprehensive coverage of auth, authorization, encryption, dependencies |
| Tool restrictions | None | Read / Grep / Glob / read-only Bash only |
| Use case | Pre-merge comprehensive check | Security-focused audit, additional pre-release verification |

### Step 4.3: UI Rubric Scoring with --ui-rubric Flag

When the `--ui-rubric` flag is specified, run the 4-axis design quality scoring flow **in addition to** the standard review.

1. Load the UI rubric definition:
   ```
   Read: ${CLAUDE_SKILL_DIR}/references/ui-rubric.md
   ```
2. Score each axis on a 0–10 scale based on the diff and any browser screenshots available
3. Add a `ui_rubric` field to the final review result:

```json
{
  "ui_rubric": {
    "design_quality": { "score": 8, "observations": ["Strong visual hierarchy", "Consistent spacing"] },
    "originality":    { "score": 6, "observations": ["Functional but template-like layout"] },
    "craft":          { "score": 9, "observations": ["Pixel-perfect implementation", "Smooth transitions"] },
    "functionality":  { "score": 7, "observations": ["All happy paths covered", "Missing empty state"] }
  }
}
```

4. UI rubric scores are informational — they do **not** affect the APPROVE/REQUEST_CHANGES verdict unless a `functionality` score of 3 or below is detected (indicates broken features).

### Step 5: Commit Decision

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
