# Failure Taxonomy

Stable catalog of named failure modes across four Harness detection systems.
Inspired by the "Failure Taxonomy" element in *Natural-Language Agent Harnesses* (arXiv:2603.25723v1).

## Purpose

Give agents a shared vocabulary for detection, recovery, and escalation.
`FT-*` IDs are cited in:
- Go hook output and error messages (`go/internal/guardrail/`)
- Phase 72 trace events (`taxonomy_id` field in `.claude/state/traces/<task_id>.jsonl`)
- Advisor history records (`taxonomy_id` field in `.claude/state/advisor/history.jsonl`)
- Agent files (`harness/agents/advisor.md`, `harness/agents/worker.md`)

## ID Stability Rule

**Never reuse an `FT-*` ID once assigned, even after the mode is removed.**
If a mode is retired, mark it `(retired)` in the table and leave the row in place.
This preserves the meaning of IDs already written into trace events and history records.

---

## Catalog

### FT-TAMPER — Test and Quality Tampering

Detected by `go/internal/guardrail/tampering.go` during PostToolUse on test/CI files.

| ID | category | mode | detector | recovery | escalation | source |
|----|----------|------|----------|----------|------------|--------|
| FT-TAMPER-01 | TAMPER | Test skipping via `it.skip` / `describe.skip` / `test.skip` / `context.skip` | PostToolUse regex on test files | Revert the skip; implement missing behavior or fix the broken test | Reviewer REQUEST_CHANGES if skip is left in the diff | `go/internal/guardrail/tampering.go` T01 |
| FT-TAMPER-02 | TAMPER | Test disabling via `xit` / `xtest` / `xdescribe` | PostToolUse regex on test files | Rename back to `it` / `test` / `describe`; address the failing case | Reviewer REQUEST_CHANGES | `go/internal/guardrail/tampering.go` T02 |
| FT-TAMPER-03 | TAMPER | Test skipping via `@pytest.mark.skip` / `@pytest.mark.xfail` | PostToolUse regex on test files | Remove skip decorator; fix the underlying failure | Reviewer REQUEST_CHANGES | `go/internal/guardrail/tampering.go` T03 |
| FT-TAMPER-04 | TAMPER | Test skipping via `t.Skip()` / `t.Skipf()` / `t.SkipNow()` | PostToolUse regex on Go test files | Remove skip call; fix the underlying failure | Reviewer REQUEST_CHANGES | `go/internal/guardrail/tampering.go` T04 |
| FT-TAMPER-05 | TAMPER | `expect(...)` call commented out in test file | PostToolUse regex on test files | Restore the expect call; investigate why it was silenced | Reviewer REQUEST_CHANGES | `go/internal/guardrail/tampering.go` T05 |
| FT-TAMPER-06 | TAMPER | `assert*` call commented out in test file | PostToolUse regex on test files | Restore the assertion; fix the implementation if assertion was failing | Reviewer REQUEST_CHANGES | `go/internal/guardrail/tampering.go` T06 |
| FT-TAMPER-07 | TAMPER | Assertion replaced with a `// TODO` comment | PostToolUse regex on test files | Replace the TODO with a real assertion | Reviewer REQUEST_CHANGES | `go/internal/guardrail/tampering.go` T07 |
| FT-TAMPER-08 | TAMPER | Lint rule silenced via `eslint-disable` directive | PostToolUse regex on all files | Remove directive; fix the underlying lint issue; use targeted disable only if intentional | Reviewer REQUEST_CHANGES on blanket disables | `go/internal/guardrail/tampering.go` T08 |
| FT-TAMPER-09 | TAMPER | CI failure hidden via `continue-on-error: true` | PostToolUse regex on CI config files | Remove `continue-on-error`; fix the failing step | Reviewer REQUEST_CHANGES | `go/internal/guardrail/tampering.go` T09 |
| FT-TAMPER-10 | TAMPER | CI step forced unconditional via `if: always()` | PostToolUse regex on CI config files | Remove `always()` unless genuinely needed (e.g. cleanup step); document if intentional | Reviewer REQUEST_CHANGES | `go/internal/guardrail/tampering.go` T10 |
| FT-TAMPER-11 | TAMPER | Hardcoded answer dictionary (test expected values returned directly) | PostToolUse regex on test files | Replace with a real implementation; run tests with novel inputs to confirm generality | Reviewer REQUEST_CHANGES; Advisor `PLAN` if recurring across retries | `go/internal/guardrail/tampering.go` T11 |
| FT-TAMPER-12 | TAMPER | Implementation returns hardcoded test case values inline | PostToolUse regex on test files | Implement general logic; verify with inputs outside the test suite | Reviewer REQUEST_CHANGES | `go/internal/guardrail/tampering.go` T12 |

---

### FT-ADVISE — Advisor Decision Triggers

Handled by `harness/agents/advisor.md` and the Worker retry loop in `harness/skills/harness-work/SKILL.md`.

| ID | category | mode | detector | recovery | escalation | source |
|----|----------|------|----------|----------|------------|--------|
| FT-ADVISE-01 | ADVISE | Repeated failure — same error signature after ≥2 retries | Worker retry loop increments `retry_count`; advisor triggered on `reason_code: repeated_failure`; duplicate-suppression skips re-consultation on identical `(task_id, reason_code, error_signature)` | Advisor returns `CORRECTION` (local fix) or `PLAN` (abandon current approach); Worker applies the suggestion | Advisor returns `STOP` when no known fix pattern; Worker escalates to Reviewer and surfaces rationale to user | `harness/agents/advisor.md` |
| FT-ADVISE-02 | ADVISE | High-risk preflight — destructive or irreversible operation detected before execution | `<!-- advisor:required -->` marker on task, or Worker detects destructive op (rm -rf, migration, force-push) in its plan; `reason_code: high_risk_preflight` | Advisor returns `PLAN` with a safer alternative approach, or `CORRECTION` with a targeted guard | Advisor returns `STOP` if `git_diff` shows a destructive op with no known-safe pattern in `patterns.md`; Worker surfaces to user before proceeding | `harness/agents/advisor.md` |
| FT-ADVISE-03 | ADVISE | Plateau before escalation — task stalled across multiple sessions without progress | Worker detects ≥3 CI failures from the same root cause across sessions, or has exhausted the review fix loop; `reason_code: plateau_before_escalation` | Advisor loads `session_log` + `trace` sources; returns `PLAN` with fresh approach incorporating cross-session context | Advisor returns `STOP` when all loaded sources are empty or no convergence path is evident; Worker generates re-ticket proposal and escalates | `harness/agents/advisor.md` |

---

### FT-RETRY — Worker Retry and Escalation Patterns

Handled by the harness-work SKILL.md review loop and CI failure handling.

| ID | category | mode | detector | recovery | escalation | source |
|----|----------|------|----------|----------|------------|--------|
| FT-RETRY-01 | RETRY | Review loop exhaustion — Reviewer returns REQUEST_CHANGES 3× without APPROVE | harness-work review loop counter (`review_count >= MAX_REVIEWS = 3`); triggered when critical/major findings remain after 3 fix cycles | Each cycle: analyze critical/major findings, implement targeted fixes, re-submit for review | After 3 failed cycles: display remaining critical/major issues to user; wait for human decision (continue / abort) | `harness/skills/harness-work/SKILL.md` Review Loop |
| FT-RETRY-02 | RETRY | CI failure retry limit — same root cause fails 3 consecutive times | harness-work CI failure handling; failure cause classified per FT-CI-* category; count incremented per same root cause | First 2 attempts: classify cause, implement targeted fix, re-run CI | 3rd failure: generate `.fix` task proposal saved to `.claude/state/pending-fix-proposals.jsonl`; present to user for approval; task status remains `cc:WIP` until user approves or rejects | `harness/skills/harness-work/SKILL.md` CI Failure Handling |
| FT-RETRY-03 | RETRY | Worker self-review gate failure — invalid `worker-report.v1` after 2 amendment cycles | Breezing Lead validates worker-report.v1 schema (B-2.5): all 5 SR rules present, `verified: true`, non-empty `evidence` | Lead sends amendment request (up to 2 cycles) with specific missing fields identified | After 2 failed amendments: Lead does not cherry-pick the Worker commit; escalates to user with summary of which SR rules failed | `harness/skills/harness-work/SKILL.md` Worker Self-Review Gate |

---

### FT-CI — CI Failure Categories

Detected by `harness/agents/ci-cd-fixer.md` during Phase 3 error classification.

| ID | category | mode | detector | recovery | escalation | source |
|----|----------|------|----------|----------|------------|--------|
| FT-CI-01 | CI | TypeScript compilation error (`TS\d{4}:` / `error TS`) | CI log pattern match: `TS\d{4}:` or `error TS` prefix in build output | Edit tool to fix type errors; `npx tsc --noEmit` to verify locally before re-running CI | After 3 failures from same TS error code: escalate per FT-RETRY-02 | `harness/agents/ci-cd-fixer.md` Phase 3 |
| FT-CI-02 | CI | Test assertion failure (`FAIL` / `AssertionError` in test output) | CI log pattern match: `FAIL` or `AssertionError` in test runner output | Fix failing assertion or the implementation under test; requires manual confirmation before auto-apply | After 3 failures: escalate; never auto-fix by weakening or removing the assertion | `harness/agents/ci-cd-fixer.md` Phase 3 |
| FT-CI-03 | CI | Dependency resolution error (`npm ERR!` / `Could not resolve`) | CI log pattern match: `npm ERR!` or `Could not resolve` | Clean install: `rm -rf node_modules package-lock.json && npm install`; requires `allow_rm_rf: true` in config | If clean install fails 3×: escalate; may indicate version conflict requiring manual `package.json` edit | `harness/agents/ci-cd-fixer.md` Phase 3 |
| FT-CI-04 | CI | Environment / secrets configuration error (env vars, permissions, external service) | CI log pattern match: `env`, `secret`, `permission` keywords; or external service HTTP 5xx | None — no auto-fix possible; display diagnosis and required manual actions | Immediate escalation to user; never attempt automated fix for credentials or external service outages | `harness/agents/ci-cd-fixer.md` Phase 3 |
| FT-CI-05 | CI | ESLint / lint rule violation (`eslint` / `Parsing error` in output) | CI log pattern match: `eslint` or `Parsing error` | `npx eslint --fix src/` for auto-fixable rules; Edit tool for manual fixes | After 3 auto-fix failures: surface specific rule violations to user for manual resolution | `harness/agents/ci-cd-fixer.md` Phase 3 |

---

## Cross-Reference Index

| Source system | File | IDs |
|---------------|------|-----|
| Go tampering patterns | `go/internal/guardrail/tampering.go` | FT-TAMPER-01 – FT-TAMPER-12 |
| Advisor error signatures | `harness/agents/advisor.md` | FT-ADVISE-01 – FT-ADVISE-03 |
| Worker retry patterns | `harness/skills/harness-work/SKILL.md` | FT-RETRY-01 – FT-RETRY-03 |
| CI fixer rules | `harness/agents/ci-cd-fixer.md` | FT-CI-01 – FT-CI-05 |

---

## Adding New Modes

1. Assign the next available `FT-<CATEGORY>-<NN>` ID (no gaps; do not skip numbers)
2. Fill all seven columns; `detector` and `recovery` must be non-empty
3. Add a cross-reference row in the index above if it's a new source system
4. **Never reuse or delete an ID** — retired modes get `(retired)` appended to the `mode` column
