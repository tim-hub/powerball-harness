# Claude Code Harness — Plans.md

Last archive: 2026-03-08 (Phase 17–24 → `.claude/memory/archive/Plans-2026-03-08-phase17-24.md`)

---

## Phase 34: Feature Table Consistency Recovery + Unused Upstream Feature Implementation

Created: 2026-04-02
Purpose: Resolve all "documented but not leveraged" gaps in the Feature Table, improving Harness reliability and utilization across both Claude and Codex contexts

### Design Principles

- P0 is "fix the lies." Align Feature Table descriptions with actual implementation
- P1 is maximum impact with minimum effort. PostCompact WIP restoration and HTTP hook practical examples
- P2 is Codex parity and security. Effort propagation for Codex Worker and security review profile
- P3 is medium-term infrastructure strengthening. OTel format conversion and dual review

### Phase 34.0: Feature Table Exaggeration Fixes [P0]

Purpose: Resolve discrepancies between Feature Table descriptions and actual implementation, restoring Harness trustworthiness

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 34.0.1 | Fix 7 inaccurate entries in Feature Table (CLAUDE.md + docs/CLAUDE-feature-table.md): HTTP hooks→template only, OTel→custom JSONL, Analytics Dashboard→planned, LSP→CC native, Auto Mode→RP Phase 1, Slack→future support, Desktop Scheduled Tasks→CC native | All 7 entries match actual state with no wording that could be misread as "implemented" | - | cc:done |

### Phase 34.1: High-Impact Quick Wins [P1]

Purpose: Achieve Feature Table consistency with minimum effort and improve long-session quality

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 34.1.1 | Implement PostCompact hook to restore WIP task information saved by PreCompact as a systemMessage | post-compact.sh reads and re-injects WIP info, preserving task state after compaction | 34.0.1 | cc:done |
| 34.1.2 | Add HTTP hook for TaskCompleted to hooks.json, implementing opt-in external notification that fires only when `HARNESS_WEBHOOK_URL` is set | hooks.json contains at least one `type: "http"` entry, with non-blocking skip when URL is unset | 34.0.1 | cc:done |

### Phase 34.2: Codex Parity + Security [P2]

Purpose: Ensure Codex Worker implementation quality and security review independence

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 34.2.1 | Add `--security` flag to harness-review, implementing a reviewer profile focused on OWASP Top 10 + auth/authz + data exposure | `/harness-review --security` launches the security profile and includes security-specific checks in review-result | 34.1.1 | cc:done |
| 34.2.2 | Add mechanism to calculate and pass effort level from Plans.md task information during codex-companion.sh task invocation | Effort is propagated to Codex Worker, with high effort applied to complex tasks | - | cc:done |

### Phase 34.3: Monitoring Infrastructure + Codex Review Integration [P3]

Purpose: Improve operational monitoring and review quality through multi-model perspectives

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 34.3.1 | Align emit-agent-trace.js output with OTel Span JSON format, sending via OTLP HTTP only when `OTEL_EXPORTER_OTLP_ENDPOINT` is set | Spans are sent when OTel endpoint is configured; falls back to existing JSONL when unset | 34.0.1 | cc:done |
| 34.3.2 | Add `--dual` flag to harness-review, running Claude Reviewer and Codex Reviewer in parallel and merging verdicts | `/harness-review --dual` produces both verdicts with an integrated final judgment | 34.2.1 | cc:done |

---

## Phase 32: Long-running harness hardening from Anthropic article

Created: 2026-03-30
Purpose: Incorporate Anthropic's long-running apps design insights into Claude Harness, addressing self-evaluation bias, context anxiety, absence of Sprint Contracts, static review bias, and incomplete Codex-side continuity

### Design Principles

- Shift from "review" to "independent Evaluator judging by executable criteria"
- Reduce compaction dependency; make structured handoff artifacts + strategic resets the canonical path
- Start from Plans.md DoD and elevate to Sprint Contract before implementation
- Use common artifacts for contract / handoff / telemetry so meaning doesn't diverge between Claude and Codex
- Explicitly state assumptions for added components, including design for future removal decisions

### Priority Matrix

| Priority | Phase | Description | Tasks | Depends |
|----------|-------|-------------|-------|---------|
| **Required** | 32.0 | Apply independent Reviewer to all modes + reduce Self-review role | 3 | None |
| **Required** | 32.1 | Sprint Contract + Context Reset/Handoff + Codex continuity | 4 | 32.0 |
| **Recommended** | 32.2 | Runtime/Browser evaluator and calibration loop | 3 | 32.1 |
| **Recommended** | 32.3 | Assumption Registry + prompt language + per-agent telemetry | 3 | 32.1 |

Total: **13 tasks**

---

### Phase 32.0: Apply Independent Reviewer to All Modes [P0]

Purpose: Standardize the article's core insight of "avoiding self-evaluation bias" across Solo / Sequential / Codex paths, not just Breezing

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 32.0.1 | Revise Solo / Sequential / Codex execution flows in `harness-work`, downgrading Worker's self-review to "pre/post implementation preflight" and ensuring the final verdict always comes from an independent Reviewer or read-only review runner | All modes require an independent verdict before `cc:done`; Worker alone cannot confirm completion | - | cc:done |
| 32.0.2 | Unify Reviewer output contract to a common artifact equivalent to `review-result.json`, with machine-readable `verdict`, `checks`, `gaps`, `followups` | Review artifact format is unified across Claude / Codex / Breezing, enabling diff comparison and re-evaluation | 32.0.1 | cc:done |
| 32.0.3 | Update `README`, `team-composition`, `harness-work`, `harness-review`, and evidence docs to align with the new contract: "self-review is auxiliary" and "independent review is the completion condition" | Review responsibility is consistently described across all docs/skills/evidence with no legacy wording remaining | 32.0.1, 32.0.2 | cc:done |

### Phase 32.1: Sprint Contract and Context Reset/Handoff [P0]

Purpose: Make pre-implementation success criteria agreement and state inheritance in long-running execution first-class artifacts

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 32.1.1 | Design and implement a flow that generates `sprint-contract.json` from Plans.md DoD/Depends, with Reviewer adding `checks`, `non_goals`, `runtime_validation`, `risk_flags` before Worker starts | Contract artifact is generated before task start; Worker cannot begin implementation without agreement | 32.0.1 | cc:done |
| 32.1.2 | Extend `pre-compact` / `post-compact` / `session-init` / `session-resume` to save and reload `handoff artifact` with `previous_state`, `next_action`, `open_risks`, `failed_checks`, `decision_log` | Handoff artifact is saved in a stable format and reused as structured state (not summary) on resume | 32.0.2 | cc:done |
| 32.1.3 | Add "strategic context reset" policy for Claude, with reset candidates and handoff generation triggered by turn count / pre-compaction / Phase transitions | Reset conditions, generated artifacts, and restart procedures are defined and reproducible at least via dry-run/fixture | 32.1.2 | cc:done |
| 32.1.4 | Absorbing `31.1.2`, connect Codex's `plugin-first workflow` and `resume-aware effort continuity` to `sprint-contract` / `handoff artifact` / `session state` | Effort and incomplete contracts are preserved after resume/fork on the Codex path, providing sufficient grounds to mark `31.1.2` as complete | 32.1.1, 32.1.2 | cc:done |

### Phase 32.2: Runtime/Browser Evaluator and Calibration Loop [P1]

Purpose: Supplement gaps that static code review alone cannot catch — UX, runtime behavior, and review accuracy drift

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 32.2.1 | Profile-ize the Reviewer with 3 types: `static`, `runtime`, `browser`. `runtime` runs test/lint/typecheck/API probe via Bash, executing contract-specified validations | Reviewer profile is selected per task type; runtime profile can execute contract-specified validation commands | 32.1.1 | cc:done |
| 32.2.2 | Add `browser` evaluator for web apps, integrating existing browser/Chrome pipeline into the reviewer flow. Validate layout integrity, major UI flows, and screenshot diffs on a contract basis | Browser profile is available; at least one fixture produces a UI flow verification artifact | 32.2.1 | cc:done |
| 32.2.3 | Accumulate review artifacts and create a calibration loop recording `false_positive`, `false_negative`, `missed_bug`, `overstrict_rule` with a few-shot update flow | A procedure and storage location exist for detecting drift from Reviewer judgment logs and updating criteria | 32.0.2, 32.2.1 | cc:done |
| 32.2.4 | Redesign browser reviewer route policy, switching default to `Playwright`-centric. Officially support 3 routes: `playwright | agent-browser | chrome-devtools` with priority order: `contract explicit > repo has Playwright infra > AgentBrowser available > Chrome fallback`. browser_mode: scripted | `sprint-contract`, browser artifact, docs, skill, fixture test express all 3 routes; route determination is repo/contract-based, not environment-dependent | 32.2.2 | cc:done |
| 32.2.5 | Add `browser_mode` (`scripted` / `exploratory`) to browser reviewer: `scripted` prioritizes Playwright, `exploratory` prioritizes AgentBrowser. Separate artifacts by role: trace/screenshot vs snapshot/ui-flow-log. browser_mode: exploratory | Contract can specify `browser_mode`; default route, required artifacts, and review procedure switch per mode | 32.2.4 | cc:done |

### Phase 32.3: Assumption Registry, Prompt Language, and Telemetry [P2]

Purpose: Enable tracking of added harness elements down to "why needed," "when removable," and "how expensive"

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 32.3.1 | Add an `Assumption Registry` with `assumption` and `retirement signal` per guardrail / skill / agent, cataloging "which model limitation each addresses" | At least major rules/agents/skills have assumptions and removal conditions recorded, usable for inventory during model updates | 32.1.1 | cc:done |
| 32.3.2 | Redesign prompt language for Worker / Reviewer / Lead, making quality posture explicit (not just procedures). Also create an A/B comparison procedure for wording diffs | Quality language is introduced in initialPrompt / review prompt; at least comparison observation logs can be captured | 32.0.1, 32.2.3 | cc:done |
| 32.3.3 | Add telemetry surface aggregating per-agent duration / token / cost / retry count / artifact count, enabling ROI comparison across Solo / Breezing / Codex modes | Per-Worker / Reviewer / Lead aggregation is visible; cost and success rate can be compared by mode | 32.0.2, 32.1.4 | cc:done |
## Phase 33: Claude 2.1.87-2.1.90 / Codex 0.118 upstream update integration

Created: 2026-04-02
Purpose: Incorporate CC 2.1.89 PermissionDenied hook / defer decision and 2.1.90 guardrail fixes into Harness, implementing auto mode denial tracking and Breezing safety valve. Auto-inherited items are explicitly classified in Feature Table

### Design Principles

- Claude side: implement "denial tracking via PermissionDenied handler" as immediate target
- defer decision: documentation only. Design concrete defer rules after accumulating operational patterns
- CC 2.1.90 PreToolUse exit 2 fix improves existing guardrail reliability (CC auto-inherited)
- Codex 0.118: retain prompt-plus-stdin as a comparison axis

### Phase 33.0: Immediate Implementation [P0]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 33.0.1 | Create `PermissionDenied` hook handler for telemetry recording of auto mode denials + Breezing Lead notification | `permission-denied-handler.sh` exists, is executable, and wired in both hooks.json files | - | cc:done |
| 33.0.2 | Add `PermissionDenied`, `defer`, `updatedInput+AskUserQuestion`, hook output >50K, and exit 2 fix to hooks-editing.md | hooks-editing.md event list and design guidelines are updated | 33.0.1 | cc:done |
| 33.0.3 | Add v2.1.84-2.1.90 to Feature Table (CLAUDE.md + docs/CLAUDE-feature-table.md) with A/C classification | Feature Table has new entries with zero B (documented-only) items | 33.0.1 | cc:done |
| 33.0.4 | Add PermissionDenied wiring validation to upstream integration test | Tests are green | 33.0.1, 33.0.2 | cc:done |

### Phase 33.1: Future Extensions [P1]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 33.1.1 | Design concrete rules for `defer` permission decision (production DB writes, external APIs, destructive git operations, etc.) | Defer conditions and resume flow exist as a design document | 33.0.2 | cc:done |
| 33.1.2 | ~~Codex 0.118 prompt-plus-stdin~~ | Companion script already passes stdin through via `exec node`. No Harness-side changes needed | - | cc:done(no-action-needed) |
| 33.1.3 | Add a pipeline for analyzing accumulated `PermissionDenied` data and suggesting auto mode permission setting optimizations | A script or skill exists for aggregating denial logs and generating improvement suggestions | 33.0.1 | cc:done |

---

## Phase 31: Claude 2.1.80-2.1.86 / Codex 0.117 upstream update integration

Created: 2026-03-28
Purpose: Avoid leaving Claude Code and Codex updates as "just documented" — incorporate Claude-side updates as real improvements to existing hooks, settings, agents, and rule generation in Harness, while organizing the next value axes for Codex

### Design Principles

- Check official changelogs / releases first; filter candidates based on primary sources
- Claude side: only implement items where "Harness can deliver 2x+ the value"
- Codex side: record comparison axes and future tasks; prioritize Claude implementation this round
- Save as a non-distributed internal skill, reusable for the same flow in future rounds

### Phase 31.0: Investigation and Immediate Implementation [P0]

Purpose: Select immediately effective items from latest updates and connect them to Harness experience improvements

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 31.0.1 | Investigate Claude changelog (`2.1.80`–`2.1.86`) and Codex releases (`0.117.0`), organizing candidates meaningful for Harness | Claude / Codex candidates are sorted into "implement now," "comparison axis," and "future" | - | cc:done |
| 31.0.2 | Incorporate Claude Code `hooks conditional if field` into `PermissionRequest`, firing the permission hook only for safe Bash commands | `.claude-plugin/hooks.json` and `hooks/hooks.json` have `if` on Bash `PermissionRequest`, passing `claude plugin validate` | 31.0.1 | cc:done |
| 31.0.3 | Align `PermissionRequest` edit matchers to `Edit|Write|MultiEdit`, ensuring hooks and core auto-approval surfaces match | `MultiEdit` is no longer missed in the hooks-side permission flow | 31.0.2 | cc:done |
| 31.0.4 | Add `sandbox.failIfUnavailable` and `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` to `.claude-plugin/settings.json`, preventing unsandboxed continuation on sandbox failure and credential leakage to subprocesses | Both items exist in settings, verifiable via validate / integration test | 31.0.1 | cc:done |
| 31.0.5 | Add `TaskCreated` / `CwdChanged` / `FileChanged` hooks and `runtime-reactive.sh`, enabling recording of background tasks, Plans updates, and worktree switches | Hook wiring and handler exist; runtime reactive test passes | 31.0.1 | cc:done |
| 31.0.6 | Migrate `paths:` in rules template and `scripts/localize-rules.sh` to YAML list format, making multi-glob patterns less fragile | Template / generated rule paths are in YAML list format; validate passes without breakage | 31.0.1 | cc:done |
| 31.0.7 | Add skill `effort` frontmatter and agent `initialPrompt` to `skills-v3` / `agents-v3` / mirror, stabilizing initial quality for heavy flows | Skill / agent frontmatter is updated; integration test and validate pass | 31.0.1 | cc:done |
| 31.0.8 | Update Feature Table / CHANGELOG / upstream integration test, recording not just "we tracked it" but "how Harness got stronger" | Docs / changelog / tests confirm v2.1.80–2.1.86 integration | 31.0.2, 31.0.3, 31.0.4, 31.0.5, 31.0.6, 31.0.7 | cc:done |

### Phase 31.1: Saving Future Extensions [P1]

Purpose: Preserve high-value updates not implemented this round in a form ready for immediate pickup next time

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 31.1.1 | Save implementation plan for Claude Code `PreToolUse updatedInput` AskUserQuestion auto-completion / input normalization as an internal skill | Target surface and implementation plan are traceable for next pickup | 31.0.1 | cc:done |
| 31.1.2 | Record Codex `plugin-first workflow` and `resume-aware effort continuity` as comparison axes in Plans | Next-phase candidates for closing Claude / Codex gaps are documented | 31.0.1 | cc:done |
| 31.1.3 | Save this entire investigation/implementation flow as a non-distributed internal skill | `skills/claude-codex-upstream-update/SKILL.md` exists with local-only usage documented | 31.0.1 | cc:done |

## Phase 30: Claude Code / Codex Cross-Runtime Hardening Parity

Created: 2026-03-25
Purpose: Incorporate insights from `claude-code-hardened` into Harness — runtime enforcement via hooks for Claude Code, approximate enforcement via wrapper / quality gate / merge gate for Codex — reflecting both paths

### Design Principles

- Do not port shell hooks directly; absorb into existing Harness surfaces (`core/guardrails`, `hooks`, `scripts/codex*`, quality gate)
- What is shared is the "policy"; implementations differ between Claude Code and Codex
- Claude Code uses deny / warn / ask via PreToolUse / PostToolUse; Codex aims for equivalent accident prevention via pre-execution injection, post-execution inspection, and pre-merge verification
- Document differences that cannot be fully aligned due to platform constraints; make them visible to users via `validate` / `doctor` output

### Phase 30.0: Common Hardening Policy Definition [P0]

Purpose: Define "what to protect" first, so meaning doesn't diverge even when Claude Code / Codex have separate implementations

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 30.0.1 | Add hardening policy matrix to `docs/` or `skills-v3/harness-work` reference docs, defining intent and severity for target rules (`--no-verify` / `--no-gpg-sign` prohibition, `reset --hard` on protected branches prohibition, protected files warning, pre-push secrets check) | Target rules, applicable surfaces, and deny/warn/ask criteria are verifiable in a table | - | cc:done |
| 30.0.2 | Define application method mapping for Claude Code / Codex, documenting "common policy, implementation differences, known asymmetries" | For each rule, it's determined whether CC hook / Codex wrapper / quality gate / docs-only is used | 30.0.1 | cc:done |

### Phase 30.1: Claude Code Runtime Hardening Additions [P0]

Purpose: Increase what hooks can directly block on the Claude Code side, reducing Git accidents and critical file mis-edits before execution

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 30.1.1 | Add `--no-verify` / `--no-gpg-sign` prohibition rules and protected branch `git reset --hard` deny / direct push warn rules to `core/src/guardrails/rules.ts` | Dangerous Git commands are denied/warned as expected without conflicting with existing force-push rules | 30.0.2 | cc:done |
| 30.1.2 | Add a settings surface for warn / deny on protected files profile (e.g., `package.json`, `Dockerfile`, `.github/workflows`, `schema.prisma`) | A default or opt-in protected files list exists; warnings or blocks trigger on Write/Edit | 30.0.2 | cc:done |
| 30.1.3 | Update `core/src/guardrails/__tests__/rules.test.ts` and integration tests with regression tests for the above hardening | Tests exist for key deny/warn cases; all existing guardrail tests pass | 30.1.1, 30.1.2 | cc:done |

### Phase 30.2: Codex Path Parity Hardening Additions [P0]

Purpose: Achieve comparable safety under Codex's no-hooks constraint by combining wrapper / prompt injection / quality gate / merge gate

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 30.2.1 | Extend `scripts/codex/codex-exec-wrapper.sh` and related rules injection flow to inject hardening contract into base instructions before Codex execution | Hardening policy is always included in prompt context before Codex execution; skip mechanisms are explicitly restricted | 30.0.2 | cc:done |
| 30.2.2 | Add protected files / no-verify equivalent / secrets inspection to post-exec verification such as `scripts/codex-worker-quality-gate.sh`, catching dangerous changes before merge | Hardening policy violations are detected in Codex output with stable failure reason output | 30.2.1 | cc:done |
| 30.2.3 | Document known limitations of the Codex path (runtime asymmetry due to absence of hooks) in docs, making the gap with Claude Code visible to users | Docs explicitly state "what was equalized / what wasn't" | 30.2.2 | cc:done |

### Phase 30.3: Validate / Doctor / Docs Visibility [P1]

Purpose: Make hardening implementation presence and coverage human-readable and judgeable

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 30.3.1 | Add hardening surface inspection to `tests/validate-plugin.sh` or a new script, enabling verification of enablement status for both Claude Code / Codex paths | Running validate shows enabled/unconfigured status for key hardening policies | 30.1.3, 30.2.2 | cc:done |
| 30.3.2 | Add explanation and usage examples to README or docs: "Claude Code has direct enforcement, Codex has approximate enforcement" | Users can understand both paths' differences and recommended usage on a single page | 30.3.1 | cc:done |
| 30.3.3 | Record cross-runtime hardening addition in CHANGELOG `[Unreleased]` | User value and constraints are concisely documented in CHANGELOG | 30.3.2 | cc:done |

---

## Phase 29: Incorporating High-Value Elements from CCAGI Report into Harness

Created: 2026-03-24
Purpose: Decompose "operationally effective but overly vendor-dependent elements" identified in the CCAGI investigation, and re-implement them without compromising Harness's public / lightweight / general-purpose direction

### Design Principles

- What we incorporate is "typing failure-prone areas," not Issue-first or AWS/auth infrastructure lock-in
- Maintain current lightweight defaults; team-oriented features are opt-in
- Don't rely solely on LLM judgment; push re-executable validations down to scripts

### Recommended Execution Order (updated 2026-04-01)

- Start with `29.2.x`, standardizing pre-release reality checks first
- Then `29.1.x` to add opt-in bridge between Plans.md and GitHub Issues with dry-run assumption
- Follow with `29.3.x` for optional brief and machine-readable manifest, building scaffolding for comparison, audit, and auto-docs
- Finally `29.4.x` to run validate / consistency / CHANGELOG, closing Phase 29 as "not just planned but re-executable"
- `25.5.3` GitHub Release creation is only performed at the actual publication timing, kept separate from normal implementation backlog

### Phase 29.0: AI Residuals Review Gate [P0]

Purpose: Add commonly leaked AI implementation artifacts (mock, dummy, localhost, TODO remnants) as a default review perspective in Harness Review, reducing "works but can't ship" states

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 29.0.1 | Add 5th review perspective `AI Residuals` to `skills-v3/harness-review/SKILL.md`, defining verdict rules for detection targets (`mockData`, `dummy`, `localhost`, `TODO`, test disabling, hardcoded config) | SKILL.md has `AI Residuals` perspective with severity judgment table; minor/major boundary is specified | - | cc:done |
| 29.0.2 | Create `scripts/review-ai-residuals.sh` to statically scan diffs or target files, outputting detection results in a stable format | Script accepts target file input and produces stable output even with 0 detections | 29.0.1 | cc:done |
| 29.0.3 | Add procedure to call the above script from `harness-review code`, plus minimal regression test/fixture | Review flow and test/fixture are added; residual detection is reflected in review results | 29.0.2 | cc:done |

### Phase 29.1: Plans.md ⇄ GitHub Issue Bridge (opt-in) [P1]

Purpose: Maintain Plans.md as SSOT while enabling bridge to Issues only when needed for team use. Does not force CCAGI's Issue-first approach; usable only when needed

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 29.1.1 | Add opt-in team mode specification to `harness-plan`, defining conversion rules for `Plans.md -> tracking issue / sub-issue` | SKILL.md or reference documents opt-in conditions, tracking issue format, and default behavior when not adopted | - | cc:done |
| 29.1.2 | Create `scripts/plans-issue-bridge.sh` to generate dry-run output (JSON or Markdown) of issue payloads from Plans.md | Script extracts task, DoD, Depends, Status from Plans.md and returns stable dry-run output | 29.1.1 | cc:done |
| 29.1.3 | Add usage differentiation and examples to `harness-plan` / docs: "unnecessary for solo development, recommended for team development" | README or docs has usage examples; default flow is not made heavier | 29.1.2 | cc:done |

### Phase 29.2: Vendor-Independent Pre-Release Verification [P1]

Purpose: Generalize only the CCAGI pre-deploy check philosophy, adding "reality checks before shipping" to Harness Release without AWS lock-in

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 29.2.1 | Add vendor-neutral pre-release check items to `skills-v3/harness-release/SKILL.md` (uncommitted diffs, env/healthcheck, debug/mock remnants, CHANGELOG, CI status) | Release flow has a pre-release verification section with abort conditions on check failure | - | cc:done |
| 29.2.2 | Create `scripts/release-preflight.sh` to execute the above checks with stable output | Script outputs pass/fail list for key checks, exiting non-zero on failure | 29.2.1 | cc:done |
| 29.2.3 | Add preflight execution pipeline to release skill / docs / tests, verifiable via dry-run | `/harness-release --dry-run` surfaces preflight guidance; related tests or fixtures are added | 29.2.2 | cc:done |

### Phase 29.3: Lightweight Brief and Machine-Readable Manifest [P2]

Purpose: Avoid importing CCAGI's document-first heaviness; surface design scaffolding only when needed. Also make Harness skills/commands easier to compare and audit

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 29.3.1 | Add optional brief generation rules to `harness-plan create`: `design brief` for UI tasks, `contract brief` for API tasks | Create flow has optional brief conditions; UI/API brief templates exist | - | cc:done |
| 29.3.2 | Add script or doc generation flow that produces `machine-readable manifest` from skill frontmatter / routing rules / mirror info | Manifest containing skill name, purpose, prohibited use, and related surfaces can be generated | - | cc:done |
| 29.3.3 | Document both artifacts in README or docs, stating usage for comparison, audit, and auto-docs | Docs describe brief/manifest purpose and generation method | 29.3.1, 29.3.2 | cc:done |

### Phase 29.4: Integration Verification and CHANGELOG [P2]

Purpose: Put Phase 29 additions onto the existing plugin/skill/docs verification loop so they don't remain just plans

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 29.4.1 | Run `validate-plugin.sh`, `check-consistency.sh`, and necessary skill mirror/checks to confirm no regressions after Phase 29 additions | All key validations pass; added scripts are re-executable in CI context | 29.0–29.3 | cc:done |
| 29.4.2 | Record "value incorporated from CCAGI investigation" in CHANGELOG `[Unreleased]` in Before/After format | CHANGELOG contains Phase 29 key points and user value | 29.4.1 | cc:done |

---

## Maintenance: Claude Code v2.1.77–v2.1.79 Integration

Created: 2026-03-20
Purpose: Integrate CC v2.1.77–v2.1.79 new features and fixes into Harness, updating Feature Table, hooks, and guardrail docs

### Phase M-CC79.0: Documentation and Hook Infrastructure Integration [P0]

Purpose: Establish the foundation with Feature Table updates and new StopFailure hook

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| CC79.0.1 | Add all new features from v2.1.77–v2.1.79 (21 items) to CLAUDE.md Feature Table, update version notation to 2.1.79+ | Feature Table has 21 new rows; notation shows 2.1.79+ | - | cc:done |
| CC79.0.2 | Add detailed sections to docs/CLAUDE-feature-table.md (behavior overview and Harness usage method for each feature) | Detailed sections exist and match existing format | CC79.0.1 | cc:done |
| CC79.0.3 | Add `StopFailure` event definition to hooks.json (×2) + create new `stop-failure.sh` handler | `StopFailure` exists in hooks.json; handler is executable | - | cc:done |
| CC79.0.4 | Update event type list, timeout table, and version notes in hooks-editing.md | `StopFailure`, `ConfigChange` exist in the list; v2.1.77/78 notes present | CC79.0.3 | cc:done |
| CC79.0.5 | Add `stop_failure` to `SignalType` in core/src/types.ts | Type definition exists | CC79.0.3 | cc:done |
| CC79.0.6 | Update session-control skill description from `/fork` to `/branch` | Description references `/branch` | - | cc:done |
| CC79.0.7 | Record integration changes in CHANGELOG.md [Unreleased] | Changes documented in Before/After format | CC79.0.1–CC79.0.6 | cc:done |

### Phase M-CC79.1: settings.json Deny Pattern Migration [P1]

Purpose: Migrate from hook-based MCP blocking to settings.json deny, leveraging v2.1.77 allow/deny priority

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| CC79.1.1 | Add `deny: ["mcp__codex__*"]` commented template to `.claude/settings.json` | settings.json has deny template | - | cc:done |
| CC79.1.2 | Add v2.1.78 settings.json deny pattern as recommendation to `codex-cli-only.md` | Rule file contains deny pattern explanation | CC79.1.1 | cc:done |

### Phase M-CC79.2: Plugin Persistent State Migration Preparation [P1]

Purpose: Leverage `${CLAUDE_PLUGIN_DATA}` variable to prevent state loss during plugin updates

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| CC79.2.1 | Gradually migrate hook handler state save paths to `${CLAUDE_PLUGIN_DATA}` support (with fallback) | Saves to `CLAUDE_PLUGIN_DATA` when set; falls back to legacy path when unset | - | cc:done |
| CC79.2.2 | Add `${CLAUDE_PLUGIN_DATA}` and `ANTHROPIC_CUSTOM_MODEL_OPTION` descriptions to harness-setup skill | SKILL.md has descriptions for both variables | - | cc:done |

### Phase M-CC79.3: CI Verification Enhancement + Agent Effort Declaration [P2]

Purpose: CI integration of `claude plugin validate` and Agent frontmatter effort field usage

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| CC79.3.1 | Add `claude plugin validate` to `tests/validate-plugin.sh` (requires v2.1.77+) | CI runs frontmatter + hooks.json syntax validation | - | cc:done |
| CC79.3.2 | Consider and add `effort` field to Worker/Reviewer agent definitions (Worker: medium, Reviewer: medium, high for security review) | Agent frontmatter has effort field | - | cc:done |

---

## Phase 28: CC Update Quality Revolution — "No Documentation-Only" + Value-Add Implementation

Created: 2026-03-20
Origin: Self-review of CC v2.1.77–v2.1.79 integration revealed "only 3 out of 21 items had Harness-specific value-add"
Purpose: (1) Create a guardrail skill to structurally prevent "documentation-only" entries in future CC update tracking (2) Implement real value-add for existing "documentation-only" items

### Background

- Three-agent parallel review (Devil's Advocate / Product Value Architect / UX Analyst) conclusions aligned
- 14 out of 21 Feature Table items were "just documenting CC's benefits"
- Improvement perception by persona: Solo developer 4/10, Breezing user 7/10, VibeCoder 1/10
- Harness's true value lies in "governance across sessions and projects"
- The more CC perfects single-session automation, the more Harness should focus on being a "meta layer"

### Design Principles

1. **"Transcribing CC features" is not value-add** — If it goes in the Feature Table, an implementation of "how Harness leverages it" is required
2. **Experience must change automatically** — Design so users benefit without reading the Feature Table
3. **Only implement what CC cannot do** — Delegate single-session features to CC. Harness governs across multiple sessions and tasks

### Priority Matrix

| Priority | Phase | Description | Tasks | Depends |
|----------|-------|-------------|-------|---------|
| **Required** | 28.0 | "No documentation-only" guardrail skill | 3 | None |
| **Required** | 28.1 | StopFailure → auto-recovery (fundamental Breezing reliability improvement) | 3 | None |
| **Required** | 28.2 | Dynamic effort injection (connecting to existing scoring) | 2 | None |
| **Recommended** | 28.3 | StopFailure log visualization + allowRead sandbox auto-config | 3 | 28.1 |
| **Required** | 28.4 | Integration verification and CHANGELOG | 2 | 28.0–28.3 |

Total: **13 tasks**

---

### Phase 28.0: "No Documentation-Only" Guardrail Skill [P0]

Purpose: Structurally prevent "Feature Table documentation-only" in future CC update tracking. Non-distributed internal-only skill.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 28.0.1 | Create `skills/cc-update-review/` (`user-invocable: false`). Build a checklist skill that verifies "does each Feature Table item have a corresponding implementation change" targeting CC update integration PRs | Skill exists with `user-invocable: false` in frontmatter | - | cc:done |
| 28.0.2 | Define 3-category judgment criteria within the skill: (A) Has implementation = changes in hooks/scripts/agents/skills (B) Documentation-only = only Feature Table changes (C) CC auto-inherited = no Harness-side changes needed (perf improvements, bug fixes, etc.). Category B requires "implementation proposal" as mandatory output | Judgment criteria are in SKILL.md; B judgment triggers implementation proposal output | 28.0.1 | cc:done |
| 28.0.3 | Create `.claude/rules/cc-update-policy.md`. Rule: "Feature Table additions must be accompanied by corresponding implementation changes or explicit Category C (CC auto-inherited) classification" | Rule file exists and is linked from CLAUDE.md | 28.0.2 | cc:done |

### Phase 28.1: StopFailure → Auto-Recovery [P0]

Purpose: When a Worker dies from rate limiting in Breezing, Lead auto-detects, backs off, and restarts. "Team governance" value-add that CC alone cannot provide.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 28.1.1 | Add StopFailure detection logic to `breezing/SKILL.md` Lead Phase B. Periodically scan `.claude/state/stop-failures.jsonl` to identify Workers with 429 errors | Lead procedure for identifying failed Workers from StopFailure logs is documented in SKILL.md | - | cc:done |
| 28.1.2 | Define error-code-specific auto-actions in `breezing/SKILL.md`: 429 → exponential backoff (30s/60s/120s) then `SendMessage` restart instruction to Worker, 401 → Lead notifies user via systemMessage, 500 → record blocker in Plans.md | Error-code-specific action table exists in SKILL.md | 28.1.1 | cc:done |
| 28.1.3 | Add `systemMessage` output to `scripts/hook-handlers/stop-failure.sh`. On 429 detection, notify Lead: "Worker X stopped due to rate limit. Auto-restarting in 30 seconds" | stop-failure.sh outputs systemMessage JSON on 429 | 28.1.1 | cc:done |

### Phase 28.2: Dynamic Effort Injection [P0]

Purpose: Connect harness-work's existing scoring (threshold >= 3 triggers ultrathink) with Agent frontmatter's `effort` field.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 28.2.1 | Extend scoring section in `harness-work/SKILL.md`. When score >= 3, document effort specification via Agent tool `model` parameter in addition to `ultrathink` injection into spawn prompt (Note: Agent frontmatter `effort: medium` is default; spawn-time specification overrides) | Scoring → effort injection flow is documented in SKILL.md | - | cc:done |
| 28.2.2 | Add "dynamic effort override from Lead" description to Effort Control section in `agents-v3/worker.md`. Add instruction to record "was effort: high sufficient?" to agent memory post-completion | worker.md documents dynamic effort reception and post-task recording procedure | 28.2.1 | cc:done |

### Phase 28.3: Log Visualization + Sandbox Auto-Config [P1]

Purpose: From "just recording" to "visible and usable." Cross-project visibility that CC alone cannot provide.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 28.3.1 | Add `--show-failures` subcommand to `harness-sync/SKILL.md`. Aggregate `.claude/state/stop-failures.jsonl` and display summary by error code and time period | `/harness-sync --show-failures` displays recent error summary | 28.1 | cc:done |
| 28.3.2 | Add `allowRead` sandbox template to `.claude-plugin/settings.json`. Reviewer can read `.env.example`, `config/public-*`, `docs/` but not `.env` or private keys | settings.json has sandbox.allowRead, designed to improve Reviewer security review accuracy | - | cc:done |
| 28.3.3 | Add sandbox auto-config step to `harness-setup/SKILL.md` `init` subcommand. Auto-generate allowRead/denyRead based on project type | `harness-setup init` procedure for auto-generating sandbox config is documented | 28.3.2 | cc:done |

### Phase 28.4: Integration Verification and CHANGELOG [P2]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 28.4.1 | Full verification with `validate-plugin.sh` + `check-consistency.sh` | All validations pass | 28.0–28.3 | cc:done |
| 28.4.2 | Record Phase 28 changes in CHANGELOG.md [Unreleased] | Changes documented in Before/After format | 28.4.1 | cc:done |

### Phase 28.5: Runtime Certainty Reinforcement [P0]

Purpose: Promote to hooks/scripts only those items that should fire deterministically without LLM judgment, rather than relying on SKILL.md instructions (LLM judgment dependent)

**Criteria for scriptification**:
- Items that auto-fire via hooks with deterministic output without LLM judgment → scriptify
- Items requiring Lead's contextual judgment (backoff wait time, effort appropriateness) → keep in SKILL.md

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 28.5.1 | Create `scripts/show-failures.sh`. Standalone script that reads `stop-failures.jsonl` and outputs error-code aggregation, last 5 entries, and recommended actions to stdout | `bash scripts/show-failures.sh` displays aggregation summary. No error on empty JSONL | - | cc:done |
| 28.5.2 | Update `--show-failures` section in `harness-sync/SKILL.md`. Change from LLM manual aggregation to Bash execution of `scripts/show-failures.sh` | SKILL.md instructs `Bash("scripts/show-failures.sh")` | 28.5.1 | cc:done |
| 28.5.3 | Regression check with `validate-plugin.sh` + `check-consistency.sh` | All validations pass | 28.5.1–28.5.2 | cc:done |

**Items NOT scriptified (with rationale)**:
- Lead backoff + restart → Wait time varies by Worker status. Lead judgment is more appropriate than a fixed script
- Dynamic effort injection → Scoring depends on spawn prompt context (task content, impact scope). Hooks lack sufficient input information
- Sandbox auto-config → Template already applied to `settings.json`. Auto-generation at init time is the responsibility of `harness-setup` skill

---

## Fix: Plugin User Quality Improvements (Issue #64, #65)

Created: 2026-03-19
Purpose: Fix critical errors and UX issues encountered after plugin installation (Issue #64: MODULE_NOT_FOUND, Issue #65: HTTP hook errors)

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| F1 | Remove `/core/dist/` from `.gitignore` exclusion, including built JS in the repository | `core/dist/index.js` is git tracked; hooks work after `claude plugin install` | - | cc:done |
| F2 | Remove `localhost:9090` HTTP hook entries from `hooks.json` (×2), moving to `docs/examples/` as templates | No HTTP hook errors in default state. Templates are accessible in documentation | - | cc:done |
| F3 | Delete broken symbolic link `skills-v3/extensions/codex-review` | `find -type l -xtype l` returns 0 broken symlinks | - | cc:done |
| F4 | Unify `marketplace.json` license with `plugin.json` (MIT) | License fields match in both files | - | cc:done |
| F5 | Record plugin quality improvements in CHANGELOG.md `[Unreleased]` | CHANGELOG documents all changes in Before/After format | F1-F4 | cc:done |

---

## Maintenance: v3.10.3 release closeout

Created: 2026-03-14
Purpose: Bundle unpublished M10–M18 as a patch release, completing version / tag / GitHub Release / main push

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M19 | Update release metadata as `3.10.3`, complete verification / tag / push / GitHub Release | `VERSION` / `plugin.json` / `CHANGELOG` / tag / GitHub Release / `origin/main` are aligned at `3.10.3`; key validations pass | M10-M18 | cc:done |

---

## Maintenance: Claude Code 2.1.76 Integration

Created: 2026-03-14
Purpose: Incorporate CC 2.1.76 new features (MCP Elicitation, PostCompact hook, -n/--name, worktree.sparsePaths, etc.) into Harness, updating Feature Table, hooks, and skills

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M13 | Add Elicitation/ElicitationResult/PostCompact hook entries to hooks.json (×2) + create new handler scripts | 3 new hooks added to hooks.json; handlers executable via `${CLAUDE_PLUGIN_ROOT}` | - | cc:done |
| M14 | Add all CC 2.1.76 new feature rows (~10) to CLAUDE.md Feature Table, update version notation to 2.1.76+ | Feature Table has new feature rows; notation shows 2.1.76+ | - | cc:done |
| M15 | Add CC 2.1.76 detailed sections to docs/CLAUDE-feature-table.md | Behavior overview, Harness usage method, and constraints documented for each new feature | M14 | cc:done |
| M16 | Add `-n`/`--name`, `worktree.sparsePaths`, partial result retention, `/effort` command references to breezing/SKILL.md + harness-work/SKILL.md | 4 features reflected in skills | - | cc:done |
| M17 | Add Elicitation/ElicitationResult/PostCompact events to hooks-editing.md + reflect `--plugin-dir` breaking change in docs | hooks-editing.md has 3 events; docs have breaking change note | - | cc:done |
| M18 | Record CC 2.1.76 integration changes in CHANGELOG.md [Unreleased] | CHANGELOG documents all changes | M13-M17 | cc:done |

---

## Maintenance: Codex command surface + stale skill cleanup

Created: 2026-03-13
Purpose: Update Harness Codex-side commands to match Codex's native multi-agent / subagent pipeline, and resolve issue of legacy skills/commands lingering in `~/.codex/skills`

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M12 | Update Codex distribution docs / AGENTS / setup scripts / tests, aligning `harness-*` command surface and legacy skill cleanup with current Codex | `test-codex-package.sh` and related validations pass; recommended command surface and stale skill cleanup are documented for Codex | M11 | cc:done |

---

## Maintenance: PR61 selective merge rescue

Created: 2026-03-13
Purpose: Incorporate PR #61 without taking release metadata along — rescue only the substantive diff missing from current `main` and make it merge-ready

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M11 | Incorporate PR61 docs diff following current release-only policy, excluding unnecessary version bump / release entries, and pass regression checks | `check-version-bump.sh` / `check-consistency.sh` / `validate-plugin.sh` / `validate-plugin-v3.sh` / `test-codex-package.sh` pass; PR61 rescue approach can be explained | M10 | cc:done |

---

## Maintenance: release-only versioning workflow

Created: 2026-03-13
Purpose: Switch to a workflow where metadata is only updated at release time, preventing version / version badge / versioned CHANGELOG from racing ahead in feature PRs and causing conflicts / red CI

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M10 | Unify pre-commit / CI / docs / release skill to "don't touch VERSION in normal PRs; bump only at release time" policy, preventing PR61-like drift from recurring | `validate-plugin.sh` / `check-consistency.sh` / `test-codex-package.sh` / necessary additional regression tests pass; operational procedure and merge policy can be explained | - | cc:done |

---

## Maintenance: v3.10.2 release closeout

Created: 2026-03-12
Purpose: Finalize TaskCompleted finalize hardening and Claude Code 2.1.74 docs tracking through README / CHANGELOG / version metadata alignment for official release

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M9 | Sync `VERSION` / `plugin.json` / README (EN/JA) / CHANGELOG / compatibility docs to 3.10.2 with latest verification results, completing commit, push, tag, and GitHub Release | `check-consistency.sh` and related tests pass; `v3.10.2` tag / GitHub Release / main push confirmed | M8 | cc:done |

---

## Maintenance: TaskCompleted finalize hardening

Created: 2026-03-12
Purpose: Safely advance harness-mem finalize to TaskCompleted timing, reducing recording loss on pre-Stop crashes

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M8 | Add idempotent finalize call to `task-completed.sh` with dedicated regression tests verifying "finalize only on last task," "no duplicate finalize," and "skip when session_id is unresolved" | `tests/test-task-completed-finalize.sh` and existing related tests pass; TaskCompleted-based finalize behavior and safety conditions can be explained | - | cc:done |

---

## Maintenance: Auto Mode review follow-up

Created: 2026-03-12
Purpose: Correct discrepancies between Auto Mode default wording and implementation reality, align agent skill preload names and breezing mirror checks to pass review

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M6 | Revert Auto Mode to rollout/opt-in wording, unify agents-v3 skill names to actual `harness-*`, and make breezing mirror drift detectable in CI | `./scripts/sync-v3-skill-mirrors.sh --check` / `./scripts/ci/check-consistency.sh` / `./tests/validate-plugin.sh` / `./tests/test-codex-package.sh` pass; no critical findings in follow-up review | - | cc:done |

---

## Maintenance: PR59/60 Auto Mode default merge prep

Created: 2026-03-12
Purpose: Make PR #59 / #60 merge-ready with Auto Mode default policy — sync skill canonical source, docs, README version notation, and mirrors, resolving remaining validate blockers

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M5 | Unify Breezing Auto Mode default to teammate execution layer, sync README / feature docs / CHANGELOG / skill mirror to be merge-ready | `./scripts/sync-v3-skill-mirrors.sh --check` / `./scripts/ci/check-consistency.sh` / `./tests/validate-plugin.sh` pass; README (EN/JA), CHANGELOG, and skills-v3/mirror are aligned | - | cc:done [6983808] |

---

## Maintenance: PR58 pre-merge stabilization

Created: 2026-03-11
Purpose: Fix docs / CI / mirror alignment in PR #58 and restore to a state where merge eligibility can be re-evaluated

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M3 | Fix Auto Mode documentation errors, README/CHANGELOG version misalignment, validate-plugin baseline breakage, and opencode mirror drift | `validate-plugin.sh` / `check-consistency.sh` / `node scripts/build-opencode.js` / `core` tests pass; PR #58 remaining blockers are organized | - | cc:done [cb625b12] |

---

## Maintenance: v3.9.0 release redo

Created: 2026-03-11
Purpose: Redo v3.9.0 as an official release without cutting a new version, aligning README / CHANGELOG / tag / GitHub Release

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M4 | Clean up unpublished version notations in CHANGELOG, create v3.9.0 tag and GitHub Release to restore release consistency | README (EN/JA) / VERSION / plugin.json / CHANGELOG / tag / GitHub Release are aligned at v3.9.0 | - | cc:done [7618428c] |

---

## Maintenance: Claude-mem MCP Removal

Created: 2026-03-08
Purpose: Remove the path for connecting Claude-mem as MCP, along with its prerequisite docs/verification pipelines from the repo

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| M1 | Remove Claude-mem MCP wrapper, setup/verification scripts, and Cursor references, aligning remaining wording | `rg` shows zero target references in production files | - | cc:done |
| M2 | While keeping `harness-mem`, remove legacy memory name user-facing wording from live setup/hook/skill | `rg` shows zero target references in live config and primary skills. Internal compatibility paths excluded | M1 | cc:done |

---

## Phase 25: Solo Mode PM Framework Enhancement

Created: 2026-03-08
Origin: Comparative analysis with pm-skills (phuryn/pm-skills) — identified PM thinking framework gap in solo mode
Purpose: Embed "structured self-questioning mechanisms" into existing skills to compensate for PM absence in solo mode (standalone Claude Code operation)

### Background

- Harness was designed with a 2-Agent premise (Cursor PM + Claude Code Worker), leaving the PM-side thinking framework thin in solo mode
- pm-skills covers PM thought structuring (Discovery, Strategy, Execution) with 65 skills / 36 chain workflows
- Harness's strengths (mandatory Evals, Plans.md markers, guardrails) and pm-skills's strengths (framework application, staged checkpoints) are complementary
- No new skills/commands are created; everything is implemented as extensions to existing skills

### Completion Criteria

1. harness-plan create priority assessment uses a 2-axis matrix (Impact × Risk)
2. Plans.md table has DoD column, auto-generated during create
3. harness-review Plan Review has Value axis added
4. harness-plan sync has retrospective functionality integrated
5. breezing Phase 0 has structured 3-question check defined
6. harness-work Solo flow has task background verification step added
7. Plans.md table has Depends column, enabling breezing to leverage dependency graph

### Priority Matrix

| Priority | Phase | Description | Tasks | Depends |
|----------|-------|-------------|-------|---------|
| **Required** | 25.0 | Plans.md format extension (DoD + Depends columns) | 3 | None |
| **Required** | 25.1 | harness-plan create enhancement (2-axis matrix + DoD auto-generation) | 3 | 25.0 |
| **Required** | 25.2 | harness-review Plan Review extension (Value axis) | 2 | None |
| **Recommended** | 25.3 | harness-plan sync retro feature | 2 | None |
| **Recommended** | 25.4 | breezing Phase 0 structuring + harness-work Solo background verification | 3 | 25.0 |
| **Required** | 25.5 | Integration verification, versioning, and release | 3 | 25.0–25.4 |

Total: **16 tasks**

---

### Phase 25.0: Plans.md Format Extension [P0]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 25.0.1 | Extend Plans.md generation template (Step 6) in `harness-plan/references/create.md` to 5-column format: `| Task | Description | DoD | Depends | Status |` | Template is in 5-column format | - | cc:done |
| 25.0.2 | Update diff detection logic in `harness-plan/references/sync.md` for 5-column format (maintaining backward compatibility with 3-column Plans.md) | Works without error on legacy 3-column Plans.md | 25.0.1 | cc:done |
| 25.0.3 | Update Plans.md format specification section in `harness-plan/SKILL.md` to 5-column, adding DoD / Depends notation guide | Format specification in SKILL.md matches the new template | 25.0.1 | cc:done |

### Phase 25.1: harness-plan create Enhancement [P1]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 25.1.1 | Extend Step 5 in `harness-plan/references/create.md` to 2-axis matrix (Impact × Risk). Auto-assign `[needs-spike]` marker to high Impact × high Risk tasks and auto-generate spike tasks | Step 5 evaluates on 2 axes; high-risk tasks get spike | 25.0.1 | cc:done |
| 25.1.2 | Add logic to Step 6 in `harness-plan/references/create.md` to auto-infer and generate DoD column from task content | All tasks in generated Plans.md have DoD filled in | 25.0.1 | cc:done |
| 25.1.3 | Add logic to Step 6 in `harness-plan/references/create.md` to auto-infer and generate Depends column from intra-phase dependencies | Tasks without dependencies show `-`; tasks with dependencies show task numbers | 25.0.1 | cc:done |

### Phase 25.2: harness-review Plan Review Extension [P2] [P]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 25.2.1 | Add Value axis to Plan Review flow in `harness-review/SKILL.md` (5th axis: linkage to user problems, consideration of alternatives, Elephant detection) | Plan Review evaluates on 5 axes (Clarity / Feasibility / Dependencies / Acceptance / Value) | - | cc:done |
| 25.2.2 | Add DoD column and Depends column quality checks to Plan Review in `harness-review/SKILL.md` (empty field detection, unverifiable DoD warnings) | Tasks with empty DoD trigger warnings | - | cc:done |

### Phase 25.3: harness-plan sync Retro Feature [P3] [P]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 25.3.1 | Add `--retro` flag support to `harness-plan/references/sync.md`. Output retrospective of completed tasks (estimation accuracy, blocking cause patterns, scope drift) | `sync --retro` displays retrospective summary | - | cc:done |
| 25.3.2 | Add `--retro` to argument-hint and sync subcommand description in `harness-plan/SKILL.md` | SKILL.md has --retro description | 25.3.1 | cc:done |

### Phase 25.4: breezing Phase 0 Structuring + harness-work Solo Background Verification [P4]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 25.4.1 | Define structured 3-question check (scope confirmation, dependency confirmation, risk flags) in `breezing/SKILL.md` Phase 0: Planning Discussion | Phase 0 has 3 concrete check items | 25.0.1 | cc:done |
| 25.4.2 | Add Step 1.5 (30-second task background verification) between Step 1 and Step 2 in `harness-work/SKILL.md` Solo flow. Display inferred purpose and impact scope; confirm with 1 question only when uncertain | Solo flow has background verification step | - | cc:done |
| 25.4.3 | Add logic to `breezing/SKILL.md` Phase 0 to read Depends column and auto-determine task assignment order based on dependency graph | Tasks with empty Depends are assigned to Workers first | 25.0.1 | cc:done |

### Phase 25.5: Integration Verification, Versioning, and Release [P5]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 25.5.1 | Full verification with `./tests/validate-plugin.sh` + `./scripts/ci/check-consistency.sh` | All validations pass | 25.0–25.4 | cc:done |
| 25.5.2 | VERSION bump + plugin.json sync + CHANGELOG update | Versions are in sync | 25.5.1 | cc:done |
| 25.5.3 | Create GitHub Release | Release is published | 25.5.2 | cc:done |

---

## Phase 26: Masao Theory Application — Transition to State-Centric Architecture

Created: 2026-03-08
Origin: Analysis of Masao's "Macro Harness / Micro Harness / Project OS" 3-element theory
Purpose: Transition from conversation-centric to state-centric operation, improving autonomous execution reliability and session continuity

### Background

- Comparative analysis of Masao theory's 3 elements (Macro/Micro/Project OS) against Harness
- Micro Harness (breezing, guardrails, Agent Teams) is mature — no updates needed
- Gaps exist in Macro Harness (planning, monitoring, re-planning) and Project OS (state infrastructure)
- Multi-angle review by 3 agents (Red Team / Architect / PM-UX) confirmed:
  - KPI/Story layer demoted from P0 (for solo development, "automation" beats "management")
  - Plans.md format changes require unified design first (prevent conflicting changes)
  - Progress feed (visibility during breezing) added as new item

### Design Principles (derived from 3-agent discussion)

1. **Increase "automation," not "management"** — Making the management layer thicker creates the paradox of users managing the management layer
2. **Gradual transition from semi-auto to full-auto** — Use propose→approve flow until accuracy stabilizes
3. **Design Plans.md changes holistically before implementing** — Prevent conflicting changes to the same files
4. **Make optional fields the default** — Mandatory fields that go unused are harmful
5. **Leverage existing infrastructure** — Prefer extending existing hooks/skills over new mechanisms

### Priority Matrix

| Priority | Phase | Description | Tasks | Depends |
|----------|-------|-------------|-------|---------|
| **Required** | 26.0 | Failure → re-ticketing flow (semi-auto MVP) | 3 | None |
| **Required** | 26.1 | harness-sync --snapshot | 3 | None |
| **Recommended** | 26.2 | Lightweight artifact linking + progress feed | 4 | None |
| **Optional** | 26.3 | Plans.md v3 format unified design | 3 | 26.2 |
| **Required** | 26.4 | Integration verification, versioning, and release | 3 | 26.0–26.3 |

Total: **16 tasks**

---

### Phase 26.0: Failure → Re-Ticketing Flow (Semi-Auto MVP) [P0] [P]

Purpose: Transition from "just stopping" to "suggesting the next action" when the self-correction loop fails

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 26.0.1 | Add failure root cause analysis step to self-correction loop exit handling in `harness-work/SKILL.md`. On 3rd STOP, generate failure log summary + recommended action + fix task proposal | On 3rd STOP, root cause analysis and fix task proposal are output | - | cc:done |
| 26.0.2 | Add user approval flow for fix task proposals. On approval, auto-add to Plans.md as `cc:TODO`; on rejection, skip | Approval → Plans.md addition, rejection → skip works | 26.0.1 | cc:done |
| 26.0.3 | Record full-auto promotion conditions in `decisions.md` as D30 (consider full automation when proposal acceptance rate reaches 80%+) | D30 is recorded | 26.0.1 | cc:done |

### Phase 26.1: harness-sync --snapshot [P0] [P]

Purpose: Fundamental solution to the "where did I leave off?" problem on session restart

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 26.1.1 | Add `--snapshot` subcommand to `harness-sync/SKILL.md`. Consolidate Plans.md WIP/TODO count + latest 3 commits + unresolved blockers into a single output | `/harness-sync --snapshot` produces a state summary | - | cc:done |
| 26.1.2 | Add snapshot generation logic to `harness-sync/references/sync.md`. Read Plans.md + recent decisions.md entries + git log | Snapshot includes state beyond just Plans.md | 26.1.1 | cc:done |
| 26.1.3 | Add `--snapshot` to argument-hint and sync subcommand description in `harness-sync/SKILL.md` | SKILL.md has --snapshot description | 26.1.1 | cc:done |

### Phase 26.2: Lightweight Artifact Linking + Progress Feed [P1] [P]

Purpose: Improve task completion traceability + better UX during breezing

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 26.2.1 | Append recent commit hash to Status in `cc:done` marker update during task completion in `harness-work/SKILL.md` (e.g., `cc:done [a1b2c3d]`) | Commit hash is auto-appended on task completion | - | cc:done |
| 26.2.2 | Update diff detection logic in `harness-plan/references/sync.md` to support `cc:done [hash]` format (backward compatible: no error on hashless format) | Works without error on legacy format Plans.md | 26.2.1 | cc:done |
| 26.2.3 | Add 1-line progress summary output on Worker task completion to Lead flow in `breezing/SKILL.md` (format: "Task 3/7 complete: User auth API implementation") | Progress is displayed on each task completion during breezing | - | cc:done |
| 26.2.4 | Add progress summary output to `scripts/hook-handlers/task-completed.sh` (leveraging existing TaskCompleted hook infrastructure) | TaskCompleted hook outputs progress information | 26.2.3 | cc:done |

### Phase 26.3: Plans.md v3 Format Unified Design [P2]

Purpose: Holistically design future KPI/Story/Artifact column additions to prevent conflicting changes

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 26.3.1 | Design Plans.md v3 format specification. Document optional Purpose line (Phase header) + Artifact notation standardization + affected files list | Specification document is created with affected files list | - | cc:done |
| 26.3.2 | Add optional Purpose line to Plans.md generation template in `harness-plan/references/create.md`. Do not prompt for input by default | Purpose line can be generated (optional). Backward compatibility with existing Plans.md maintained | 26.3.1 | cc:done |
| 26.3.3 | Record Plans.md v3 format design decision in `decisions.md` as D31 | D31 is recorded | 26.3.1 | cc:done |

### Phase 26.4: Integration Verification, Versioning, and Release [P3]

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 26.4.1 | Full verification with `./tests/validate-plugin.sh` + `./scripts/ci/check-consistency.sh` | All validations pass | 26.0–26.3 | cc:done |
| 26.4.2 | VERSION bump + plugin.json sync + CHANGELOG update | Versions are in sync | 26.4.1 | cc:done |
| 26.4.3 | Create GitHub Release | Release is published | 26.4.2 | cc:done [56cdd77] |

---

## Phase 27: Masao Theory Application — Implementation Alignment Hardening

Created: 2026-03-10
Origin: Review of `56cdd777 feat: state-centric architecture with masao theory`
Purpose: Close gaps where Phase 26's "state-centric" features were description-first, ensuring they work end-to-end through implementation, restart pipeline, and traceability

### Background

- The direction alignment with Masao theory is correct, but some gaps exist between "what the description says is possible" and "what actually happens at runtime"
- In particular, "failure → re-ticketing" does not reach fix task addition on the TaskCompleted hook side — effectively stops at root cause analysis + escalation
- `--snapshot` has save design but auto-load/compare on session restart is not connected
- Among Project OS minimum requirements (purpose / acceptance criteria / upstream reference / artifact link), upstream reference is still thin

### Priority Matrix

| Priority | Phase | Description | Tasks | Depends |
|----------|-------|-------------|-------|---------|
| **Required** | 27.0 | Failure → re-ticketing runtime implementation | 3 | None |
| **Required** | 27.1 | Snapshot restart pipeline connection | 2 | None |
| **Recommended** | 27.2 | Project OS minimum traceability reinforcement | 3 | 27.1 |

Total: **8 tasks**

---

### Phase 27.0: Failure → Re-Ticketing Runtime Implementation [P0]

Purpose: Make Phase 26's "leave the next action as a ticket" work as actual runtime behavior, not just description

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 27.0.1 | Add `.fix` task proposal structured output to `scripts/hook-handlers/task-completed.sh`, returning failure category, source task number, DoD, and Depends in machine-readable format | On 3rd failure, fix task proposal is retrievable in JSON or stable format | - | cc:done |
| 27.0.2 | Implement approval flow to safely append fix task proposals to `Plans.md` (add only on approval, prevent duplicate additions) | Approval adds `.fix` task exactly once; rejection leaves Plans.md unchanged | 27.0.1 | cc:done |
| 27.0.3 | Align re-ticketing descriptions in `skills/harness-work/SKILL.md` / `CHANGELOG.md` / `.claude/memory/decisions.md` with implementation, adding regression verification | "Proposal only" / "addition after approval" / "full auto" boundaries are consistent across all files with a reproducible procedure | 27.0.2 | cc:done |

### Phase 27.1: Snapshot Restart Pipeline Connection [P0]

Purpose: Make saved state actually usable in the next session, closing the state-centric architecture loop

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 27.1.1 | Add latest snapshot loading to `session-init` / `session-resume.sh`, displaying diff summary from previous session on restart | Restarting with an existing snapshot shows a diff summary. Silently skips when none exists | - | cc:done |
| 27.1.2 | Fix verification procedure for initial save → 2nd comparison → restart load of `harness-sync --snapshot` as a reproducible script in `tests/` or documented procedure | Snapshot save → compare → restart verification can be reproduced following the procedure | 27.1.1 | cc:done |
| 27.1.3 | Separate `session-init` and usage tracking hook stdout noise, with regression verification that hook output is JSON body only | `session-init` / usage tracking telemetry output doesn't break hook stdout; direct execution verification exists | 27.1.2 | cc:done |
| 27.1.4 | Connect Claude SessionStart/UserPromptSubmit/PostToolUse/Stop to harness-mem runtime, displaying continuity briefing as the first thing in `session-init` / `session-resume` | hooks.json calls memory hook; Claude's SessionStart additionalContext shows `Continuity Briefing`; pending artifacts are not double-injected | 27.1.3 | cc:done |
| 27.1.5 | Add Claude memory lifecycle regression verification, fixing that start → prompt → stop flows through the same continuity chain | Lifecycle integration test confirms `record-event` / `resume-pack` / `finalize-session` use the same `correlation_id`, and SessionStart briefing is displayed | 27.1.4 | cc:done |

### Phase 27.2: Project OS Minimum Traceability Reinforcement [P1]

Purpose: Add minimal format for "why does this ticket exist" upstream tracing, within bounds that avoid management overhead
