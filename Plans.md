# Powerball Harness — Plans.md

Last archive: 2026-04-15 (Phase 35–48 → `.claude/memory/archive/Plans-2026-04-15-phase35-48.md`)
Last release: v4.5.0 on 2026-04-15 (Phase 60+61)

---

## Phase 62: Advisor Strategy — read-only consultation agent + harness-loop skill

Created: 2026-04-16

**Goal**: Implement the Advisor Strategy introduced in upstream v4.1.1. Add a read-only `advisor` agent that executors (Worker, breezing Lead) can consult when they hit decision blockers — high-risk tasks, repeated failures from the same root cause, or plateau before user escalation. The advisor returns a structured `PLAN | CORRECTION | STOP` decision; it never writes code or invokes tools. Add `harness-loop` as a new skill for long-running autonomous execution loops that consult the advisor at its three trigger points. Adapt all paths to our `harness/` directory layout (v4.5.0).

**Reference**: upstream Chachamaru127/claude-code-harness PR #83 (v4.1.1). Do NOT cherry-pick — implement independently from scratch using the upstream as a feature specification.

### Stage 1: Config + advisor agent

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 62.1 | Add `advisor:` config block to `harness/.claude-code-harness.config.yaml` — fields: `enabled: true`, `mode: on-demand`, `max_consults_per_task: 3`, `retry_threshold: 2`, `consult_before_user_escalation: true`, `model_defaults.claude: opus` | `grep 'advisor:' harness/.claude-code-harness.config.yaml` returns a match; file is valid YAML | - | cc:Done [5f4ab76] |
| 62.2 | Create `harness/agents/advisor.md` — new read-only consultation agent. Frontmatter: `model: opus`, `allowed-tools: [Read, Grep, Glob]` only (no Write/Edit/Bash/Task). Body: role definition (read-only, no execution authority), response schema `advisor-response.v1` with three decision types (`PLAN` = replan approach, `CORRECTION` = apply local fix, `STOP` = escalate to reviewer), trigger inputs (risk flags, error signatures, plateau count), duplicate-suppression via `task_id + reason_code + error_sig` hash, state location `.claude/state/advisor/`. Description: `Use when consulting on blocked tasks, high-risk preflight, or repeated-failure patterns — returns PLAN/CORRECTION/STOP. Do NOT load for: implementation, review, planning.` | Agent file exists; `harness validate agents` passes; description ≤300 chars and starts with `Use when` | 62.1 | cc:Done [837cb1a] |

### Stage 2: Team composition + existing agent updates

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 62.3 | Update `harness/agents/team-composition.md` — expand from 3-agent to 4-agent model. Add Advisor as a "consult-only" lateral role between Worker and Reviewer in the team diagram. Update Role Definitions table with Advisor row (model: opus, tools: read-only, authority: guidance only). Update legacy-agent-mapping table to note advisor is new. Keep description ≤300 chars. | Diagram shows 4 roles; `grep 'Advisor' harness/agents/team-composition.md` ≥ 3 matches | 62.2 | cc:Done [d2062e0] |
| 62.4 | Update `harness/agents/worker.md` — add "Advisor Consultation" section describing the 3 trigger conditions (high-risk task marker `<!-- advisor:required -->`, same-cause failure ≥ `retry_threshold`, plateau before user escalation) and the consultation flow: read advisor config → invoke `powerball-harness:advisor` subagent → parse `PLAN/CORRECTION/STOP` → act accordingly. Add `--no-advisor` opt-out flag reference. | `grep 'advisor' harness/agents/worker.md` ≥ 5 matches; file passes validation | 62.2 | cc:Done [cdfebb1] |
| 62.5 | Update `harness/agents/reviewer.md` — add a short "Advisor vs Reviewer" boundary note: advisor gives mid-task guidance without final authority; reviewer gives final APPROVE/REQUEST_CHANGES verdict after implementation. Advisor cannot bypass the reviewer gate. | Note present in reviewer.md; no behavior changes to review flow | 62.2 | cc:Done [95bbdd0] |

### Stage 3: harness-loop skill (new)

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 62.6 | Create `harness/skills/harness-loop/SKILL.md` — new long-running autonomous loop skill. Frontmatter: `name: harness-loop`, `allowed-tools: [Read, Write, Edit, Bash, Grep, Glob, Task]`, `argument-hint: "[N-iterations|--until-done|--advisor|--no-advisor]"`. Body: loop execution model (iterate over Plans.md tasks in sequence), 3 advisor trigger points (pre-task risk check, post-failure retry gate, plateau detection), loop exit conditions (`cc:done` on all tasks, `--until-done` convergence, STOP from advisor), state file `.claude/state/loop-active.json`. Description ≤300 chars starting with `Use when`. | Skill dir + SKILL.md exist; `grep '^description:' harness/skills/harness-loop/SKILL.md` ≤300 chars and starts `"Use when`; `./local-scripts/audit-skill-descriptions.sh harness/skills/harness-loop` passes | 62.2 | cc:Done [8035338] |
| 62.7 | Add harness-loop to `harness/templates/codex-skills/harness-loop/SKILL.md` — codex-native variant adds `disable-model-invocation: true` frontmatter and strips interactive prompts (advisor flow becomes non-interactive: auto-accept CORRECTION, escalate STOP to codex-loop exit) | File exists at `harness/templates/codex-skills/harness-loop/SKILL.md`; contains `disable-model-invocation: true` | 62.6 | cc:Done [42713f1] |
| 62.8 | Add harness-loop to `harness/templates/opencode/skills/harness-loop/SKILL.md` — plain copy of 62.6 (opencode ignores unknown frontmatter fields) | File exists; content matches 62.6 | 62.6 | cc:Done [b9d8340] |

### Stage 4: Update existing skills for advisor integration

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 62.9 | Update `harness/skills/breezing/SKILL.md` — add `--advisor` / `--no-advisor` to Options table and argument-hint. Add "Advisor Integration" section: breezing Lead checks advisor config before spawning Workers for tasks tagged `<!-- advisor:required -->`; on Worker STOP signal, Lead invokes advisor before user escalation. Keep breezing as a thin alias to harness-work — only add the advisor hooks at the coordination layer. | `grep '\-\-advisor' harness/skills/breezing/SKILL.md` ≥ 2 matches; `validate-plugin.sh` passes | 62.4 | cc:Done [972690d] |
| 62.10 | Update `harness/skills/harness-work/SKILL.md` — add `--advisor` / `--no-advisor` flags to Options table and argument-hint. Add a short "Advisor Consultation" paragraph in the Execution section: when `--advisor` is active (or advisor.enabled in config), consult advisor at the 3 trigger points before escalating to user. | `grep '\-\-advisor' harness/skills/harness-work/SKILL.md` ≥ 2 matches | 62.4 | cc:Done [43d1a33] |

### Stage 5: Scripts + Go engine

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 62.11 | Create `harness/scripts/run-advisor-consultation.sh` — wrapper script that: (1) reads advisor config via `harness/.claude-code-harness.config.yaml`, (2) checks `enabled` flag and `max_consults_per_task` counter in `.claude/state/advisor/history.jsonl`, (3) writes `last-request.json` to `.claude/state/advisor/`, (4) outputs a structured prompt block for Claude to invoke the advisor subagent, (5) reads `last-response.json` and returns the `PLAN|CORRECTION|STOP` value. Uses `${BASH_SOURCE[0]}` for path resolution. | Script exists; `bash harness/scripts/run-advisor-consultation.sh --help` exits 0; `BASH_SOURCE` pattern used (not `$0`) | 62.1, 62.2 | cc:Done [ad1a2b0] |
| 62.12 | Add `go/internal/hookhandler/advisor_trigger.go` — new hook handler that detects advisor trigger conditions from hook events: (a) reads task markers for `<!-- advisor:required -->`, (b) tracks consecutive failure signatures in `.claude/state/advisor/failure-log.jsonl`, (c) increments plateau counter when task restarts without new commits. Exposes `ShouldConsultAdvisor(taskID, retryCount, errorSig string) bool`. Add corresponding `advisor_trigger_test.go`. | Both files exist; `go test ./go/internal/hookhandler/...` passes; new function exported correctly | 62.2 | cc:Done [9db1c5e] |

### Stage 6: Tests + docs + fixes

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 62.13 | Create `tests/test-advisor-protocol.sh` — verify: (a) advisor agent file has read-only `allowed-tools` (no Write/Edit/Bash/Task), (b) response schema contains PLAN/CORRECTION/STOP, (c) `run-advisor-consultation.sh` respects `max_consults_per_task` by counting `.claude/state/advisor/history.jsonl` entries, (d) STOP triggers user escalation path in mock invocation | `bash tests/test-advisor-protocol.sh` exits 0 | 62.2, 62.11 | cc:Done [5fed08d] |
| 62.14 | Create `tests/test-advisor-config.sh` — verify: (a) `advisor.enabled: false` in config disables consultation (script exits early with "advisor disabled"), (b) `retry_threshold` is read correctly, (c) `max_consults_per_task` ceiling is enforced, (d) config block is valid YAML and parses without error | `bash tests/test-advisor-config.sh` exits 0 | 62.1, 62.11 | cc:Done [c389189] |
| 62.15 | Create `docs/advisor-strategy.md` — strategy documentation covering: trigger conditions (with examples), decision type definitions (PLAN/CORRECTION/STOP), advisor vs reviewer authority boundary, duplicate-suppression mechanism, configuration reference (all `advisor:` fields), integration diagram showing 4-agent model, harness-loop interaction | File exists; linked from README.md | 62.2, 62.6 | cc:Done [64dbb30] |
| 62.16 | Update `README.md` — add "Advisor Strategy" section (after Agent Team section) describing what the advisor does, when it triggers, and linking to `docs/advisor-strategy.md` | `grep 'Advisor Strategy' README.md` returns a match | 62.15 | cc:Done [0a6b143] |
| 62.17 | Fix `harness/bin/harness` symlink resolution — update the shim script to use `readlink -f` (macOS: `realpath` fallback) so PATH-installed invocations resolve correctly regardless of symlink depth | `harness --version` works when `harness/bin/` is on PATH via symlink at `~/.local/bin/harness` | - | cc:Done [a91bb91] |

### Stage 7: Validation

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 62.18 | Run full validation: `./tests/validate-plugin.sh`, `./local-scripts/check-consistency.sh`, `./local-scripts/check-residue.sh`, `./local-scripts/audit-skill-descriptions.sh harness/skills/harness-loop`, `harness validate agents`, `go test ./go/internal/hookhandler/...` | All pass with 0 failures | 62.1–62.17 | cc:Done [b1a3fbe] |
| 62.19 | Add `[Unreleased]` CHANGELOG entry covering: (a) new advisor agent and 4-agent model, (b) harness-loop skill, (c) breezing/harness-work advisor flags, (d) advisor trigger hook in Go engine, (e) symlink fix. Use Before/After format per `.claude/rules/github-release.md` | CHANGELOG entry present under `[Unreleased]` with Before/After sections | 62.18 | cc:Done [20a7933] |

---

## Phase 61: Agent files optimization pass — `harness/agents/`

Created: 2026-04-15

Goal: Optimize the 6 files under `harness/agents/` for frontmatter hygiene, token efficiency, wording clarity, and step logic. Use `/skill-creator` and `/skill-development` best-practice patterns as reference for description format and structure; apply agent-specific rules (frontmatter fields, model selection, tool allow-lists) from `plugin-dev:agent-development`. No behavioral changes — what each agent does stays the same; only how it is described and structured changes.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 61.1 | Audit all 6 agent files — record for each: line count, frontmatter completeness (description, model, allowed-tools, maxTurns, effort), description length and prefix, body section names, any duplicated content, and step-logic issues (unreachable steps, redundant retries, missing exit conditions) | Short audit table with one row per agent and a concrete issue list for each | - | cc:Done [6854c9f] |
| 61.2 | Normalize `description:` on all 6 agents: must start with `Use when <trigger>`, stay ≤300 chars, and not repeat capability prose that belongs in the body opening paragraph (mirrors `.claude/rules/skill-description.md`) | `awk -F'"' '/^description:/{print $2}' harness/agents/*.md` — all descriptions start with `Use when ` and are ≤300 chars | 61.1 | cc:Done [6854c9f] |
| 61.3 | Tighten step lists in `ci-cd-fixer.md` (475 lines) and `error-recovery.md` (348 lines) — merge duplicate guidance, remove redundant retry prose already implied by `maxTurns`, and cut steps that simply restate prior steps with different wording | Both files ≤ 350 lines; no content loss verified by re-reading the trimmed section against the original | 61.1 | cc:Done [6854c9f] |
| 61.4 | Reduce `team-composition.md` (570 lines) by moving reference tables (long agent capability matrices, example invocation blocks) into `harness/agents/references/team-composition-tables.md` and linking from the main file | Main file ≤ 400 lines; moved tables accessible via link; file passes `validate-plugin.sh` | 61.1 | cc:Done [6854c9f] |
| 61.5 | Audit and correct `allowed-tools` on every agent: remove tools the agent never invokes, add tools it does invoke but hasn't listed (e.g. `Grep`/`Glob` for search-heavy agents, `Bash` for script runners) | Each agent's `allowed-tools` list matches the set of tools actually called out in its body steps; no stale or missing entries | 61.1 | cc:Done [6854c9f] |
| 61.6 | Confirm or set `model:` on every agent: `haiku` for narrow high-frequency tasks, `sonnet` for main implementation work, `opus` only where the agent explicitly needs deep reasoning. Add a one-line rationale comment next to each choice | Every agent has an explicit `model:` field; rationale is in a comment or the audit note | 61.1 | cc:Done [6854c9f] |
| 61.7 | Remove `## Trigger Phrases`, `## When to Use`, and similar sections that duplicate the frontmatter description (Phase 59.5 applied the same cleanup to skills) | `grep -l '## Trigger Phrases\|## When to Use' harness/agents/*.md` returns 0 files | 61.1 | cc:Done [6854c9f] |
| 61.8 | Validate: `./tests/validate-plugin.sh` and `./local-scripts/check-consistency.sh` both pass; add CHANGELOG [Unreleased] entry under Changed describing the optimization pass in Before/After format | Both scripts pass; CHANGELOG entry present | 61.2–61.7 | cc:Done [6854c9f] |

---

## Phase 60: Commit prebuilt binaries, move build tooling, add go-change hook

Created: 2026-04-15

Goal: Ship `harness-darwin-arm64`, `harness-darwin-amd64`, `harness-linux-amd64` in-repo under `harness/bin/` (already permitted for plugins). Move `build-binary.sh` from the skill's `scripts/` folder to `local-scripts/` where it belongs alongside other dev helpers. Add a `build-all` target to the root Makefile that cross-compiles all three platforms into `harness/bin/`. Extend `.githooks/pre-commit` to auto-rebuild the current-platform binary when staged files include changes under `go/`. Remove the `binary` subcommand from `harness-setup` entirely — it is a dev concern, not a setup concern.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 60.1 | Cross-compile the two missing binaries (`harness-darwin-amd64`, `harness-linux-amd64`) alongside the existing `harness-darwin-arm64` using `go/scripts/build-all.sh`; confirm `VERSION` is embedded in each | All three files exist under `harness/bin/`, are executable, and each reports the correct version via `--version` | - | cc:done |
| 60.2 | Remove `harness/bin/harness-*` from `.gitignore` (currently line 285) so all three binaries become tracked | `grep 'harness/bin/harness-' .gitignore` returns no matches | - | cc:done |
| 60.3 | Stage all three binaries: `git add harness/bin/harness-darwin-arm64 harness/bin/harness-darwin-amd64 harness/bin/harness-linux-amd64` | `git status` shows all three binaries as tracked/staged; each <20 MB | 60.1, 60.2 | cc:done |
| 60.4 | Move `harness/skills/harness-setup/scripts/build-binary.sh` → `local-scripts/build-binary.sh`; update header comment to reflect its new dev-helper-only role | File exists at `local-scripts/build-binary.sh`; no longer present at `harness/skills/harness-setup/scripts/`; `grep -rn 'build-binary' harness/skills/` returns 0 hits | - | cc:done |
| 60.5 | Add `build` and `build-all` targets to the root `Makefile` that call `local-scripts/build-binary.sh` (current platform) and `go/scripts/build-all.sh` (all three platforms) respectively, both outputting to `harness/bin/`; retire the current `make build` that pointed at the old skill-scripts path | `make build` rebuilds `harness/bin/harness-<current-platform>`; `make build-all` rebuilds all three; both use `harness/bin/` as output dir | 60.4 | cc:done |
| 60.6 | Extend `.githooks/pre-commit` to detect staged changes under `go/`: if any `go/` file is staged, run `make build-all` to rebuild all platform binaries and re-stage them automatically | Committing after editing a `.go` file causes the pre-commit hook to rebuild and restage all `harness/bin/harness-*` binaries before the commit completes | - | cc:done |
| 60.7 | Remove the `binary` subcommand from `harness-setup/SKILL.md` completely: drop it from the Quick Reference table, `argument-hint`, and the `### binary — Platform Binary Build` section; remove the `binary` step from the `init` flow | No `binary` keyword in `harness-setup/SKILL.md`; `init` flow starts directly at step "Detect project type" | 60.4 | cc:done |
| 60.8 | Audit remaining references to `build-binary\|harness-setup binary` across `harness/`, `.claude-plugin/`, `hooks/` and fix any that still point to the old skill-scripts location | `grep -rn 'build-binary\|harness-setup binary' harness/ .claude-plugin/ hooks/` returns 0 hits | 60.4–60.7 | cc:done |
| 60.9 | Update `deleted-concepts.yaml`: add the old skill-scripts path `harness/skills/harness-setup/scripts/build-binary.sh` so residue scans catch any re-introduction; confirm `check-residue.sh` reports 0 on HEAD | `check-residue.sh` passes; entry present in `deleted-concepts.yaml` | 60.4 | cc:done |
| 60.10 | Validate: run `./tests/validate-plugin.sh` and `./local-scripts/check-consistency.sh`; add CHANGELOG [Unreleased] entry covering (a) binaries shipped prebuilt, (b) `binary` subcommand removed from setup, (c) `make build` / `make build-all` available for contributors, (d) pre-commit hook auto-rebuilds on go/ changes | Both scripts pass; CHANGELOG entry present with Before/After format | 60.1–60.9 | cc:done [866b895] |

---

## Phase 59: SKILL.md quality pass — all 26 skills

Created: 2026-04-15

Goal: Systematic review and optimization of every skill under `harness/skills/`. Fix Quick Reference table format, anchor all script/reference paths to the correct tier (`${CLAUDE_SKILL_DIR}`), remove dead references, clean up deprecated slash commands, and remove redundant `## Trigger Phrases` sections. Executed as 7 rounds of parallel subagent pairs.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 59.1 | Fix Quick Reference table format — convert bash code blocks or wrong column headers to 3-column `User Input \| Subcommand \| Behavior` in `breezing`, `harness-sync`, `harness-release`, `writing-changelog`, `harness-work` | `grep -r '\| Mode \|' harness/skills/` returns 0 hits | - | cc:done |
| 59.2 | Add missing Quick Reference tables to `session` (3 tokens) and `notebook-lm` (2 tokens) where `argument-hint` had multiple pipe-separated subcommands but no table existed | Both skills have 3-column Quick Reference near top of SKILL.md | - | cc:done |
| 59.3 | Fix `${CLAUDE_SKILL_DIR}` path anchoring — replace bare `scripts/foo.sh` / `python3 scripts/foo.py` with skill-local or plugin-local anchored paths in `gogcli-ops` (4 paths), `session-control` (SKILL.md + reference file), `session-state` (reference file), `harness-review` (SKILL.md + `dual-review.md`), `harness-work` (SKILL.md + `codex-work.md`), `memory` | `grep -r 'bash scripts/' harness/skills/` returns 0 prose hits | - | cc:done |
| 59.4 | Remove dead file references: `docs/SESSION_ORCHESTRATION.md` (session-state), `docs/MEMORY_POLICY.md` (session-memory), `AGENTS.md` (deploy/health-checking.md), broken `${CLAUDE_SKILL_DIR}/../../docs/release-preflight.md` (harness-release) | No SKILL.md references non-existent docs/ files | - | cc:done |
| 59.5 | Remove `## Trigger Phrases` and `## Trigger Conditions` sections from SKILL.md and reference files (redundant with `description:` frontmatter) — affected: `session-init`, `session-memory`, `vibecoder-guide`, `agent-browser`, `ci`, and 5 reference files in `auth/` and `deploy/` | `grep -r '## Trigger Phrases' harness/skills/` returns 0 hits | - | cc:done |
| 59.6 | Update legacy slash commands to current skill invocation patterns in `workflow-guide` (6 commands), `session-init` (3 commands), `session` (2 commands), and `workflow-guide/examples/typical-workflow.md` | `grep -r '/harness-init\|/plan-with-agent\|/handoff-to-' harness/skills/` returns 0 hits | - | cc:done |
| 59.7 | Structural fixes: move Quick Reference from line 99 → top in `agent-browser`; remove non-standard `Trigger` column from `ci` Feature Details table; fix `harness-plan` sync row wording; expand truncated Step 1 in `session-control`; add vibecoder-guide/session-init distinction note | Each fix individually verified post-edit | - | cc:done |
| 59.8 | Delete `principles/references/vibecoder-guide.md` — content duplicated the standalone `vibecoder-guide` skill | File deleted; `principles` SKILL.md has no dangling reference | - | cc:done |

---

## Phase 56: Go validator + agent frontmatter fixes

Created: 2026-04-15

Goal: Fix the critical validator bug that rejects all agents, and clean up stale/missing fields in all agent frontmatter files. These are the highest-priority items from the project-wide review.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 56.1 | Add short model aliases (`sonnet`, `opus`, `haiku`) to `validModelNames` in `go/cmd/harness/validate.go` | `harness validate agents` passes on all 6 agent files; Go tests pass | - | cc:Done [4e9303f] |
| 56.2 | Fix `harness/agents/ci-cd-fixer.md` — remove `verify` from `skills:` (only keep `ci`); change `disallowedTools: [Task]` → `[Agent]`; add `permissionMode: bypassPermissions`, `effort: medium`, `maxTurns: 75`; align hook syntax to nested format matching worker.md | Agent passes `harness validate agents`; no references to non-existent skills | 56.1 | cc:Done [4e9303f] |
| 56.3 | Fix `harness/agents/error-recovery.md` — remove `skills: [verify, troubleshoot]` (both non-existent); change `disallowedTools: [Task]` → `[Agent]`; add `permissionMode: bypassPermissions`, `effort: medium`, `maxTurns: 75`; add deprecation notice header noting consolidation into `worker` per `team-composition.md` | Agent passes validation; deprecation status clear | 56.1 | cc:Done [4e9303f] |
| 56.4 | Fix `harness/agents/scaffolder.md` — update `"harness_version": "none | v2 | v3"` → `"none | v2 | v3 | v4"` in the output JSON schema | `grep 'v4' harness/agents/scaffolder.md` returns a match | - | cc:Done [4e9303f] |
| 56.5 | Rebuild Go binary after validate.go change | `harness/bin/harness-darwin-arm64` updated; `harness validate agents` succeeds end-to-end | 56.1 | cc:Done [4e9303f] |

---

## Phase 57: Documentation drift cleanup

Created: 2026-04-15

Goal: Fix stale references, duplicate rows, and format mismatches across docs, memory, and rules. All items from the project-wide review classified as documentation issues.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 57.1 | Fix `docs/CLAUDE-skill-catalog.md` — remove references to non-existent skills `impl/`, `verify/`, `handoff/`, `maintenance/`, `troubleshoot/` from hierarchy diagram and category table | `grep -E 'impl/\|verify/\|handoff/\|maintenance/\|troubleshoot/' docs/CLAUDE-skill-catalog.md` returns 0 | - | cc:Done [625db0b] |
| 57.2 | Fix `CONTRIBUTING.md` CHANGELOG format section (lines 131-161) — replace Keep a Changelog style description with actual Before/After narrative format used by the project | CONTRIBUTING.md CHANGELOG section matches `github-release.md` rules | - | cc:Done [625db0b] |
| 57.3 | Fix `CONTRIBUTING.md` version management section (lines 98-103) — update "two places" to reference `harness/VERSION` + `harness/harness.toml` instead of `marketplace.json` | No mention of `marketplace.json` having a version field | - | cc:Done [625db0b] |
| 57.4 | Fix `CONTRIBUTING.md` Testing section (lines 212-216) — fix duplicate step number "3." | Sequential step numbering (1, 2, 3, 4) | - | cc:Done [625db0b] |
| 57.5 | Mark `.claude/memory/patterns.md` P1-P3 as superseded — add `_(superseded by D9/Go migration — see go/internal/guardrail/)_` markers; keep historical content but clearly flag it | P1, P2, P3 each have a superseded marker | - | cc:Done [c174a80] |
| 57.6 | Deduplicate `docs/CLAUDE-feature-table.md` — remove duplicate Slack Integration row (line ~256) and duplicate Auto Mode row (line ~187); review 3 "planned/future" items and mark with dates or remove | No duplicate rows; planned items either have target dates or are removed | - | cc:Done [c174a80] |
| 57.7 | Fix `go/DESIGN.md` — remove or annotate `internal/plans/` reference as "not yet implemented" | DESIGN.md accurately reflects actual package structure | - | cc:Done [711929a] |
| 57.8 | Fix `.claude/rules/hooks-editing.md` — remove stale dual-sync `.claude-plugin/hooks.json` requirement; update to reflect current architecture where `harness/hooks/hooks.json` is the SSOT | Rule matches actual file layout | - | cc:Done [711929a] |
| 57.9 | Register orphaned templates in `harness/templates/template-registry.json` or delete orphaned files — `sandbox-settings.json.template`, `rules/quality-gates.md.template`, `rules/security-guidelines.md.template`, `rules/tdd-guidelines.md.template` | Every `.template` file on disk has a registry entry, OR orphaned files are removed | - | cc:Done [711929a] |

---

## Phase 58: Script hygiene + settings hardening

Created: 2026-04-15

Goal: Fix shell script path conventions, strict mode, variable naming, and settings.json deny rule inconsistencies. All LOW-severity items from the project-wide review.

### Stage 1: Path convention fixes

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 58.1 | Fix `$0` → `${BASH_SOURCE[0]}` in 3 harness scripts: `harness/scripts/codex/sync-rules-to-agents.sh:15`, `harness/scripts/i18n/set-locale.sh:12`, `harness/scripts/i18n/check-translations.sh:9` | `grep -rn 'dirname "\$0"' harness/scripts/` returns 0 results | - | cc:Done [b4b5ef0] |
| 58.2 | Fix `$0` → `${BASH_SOURCE[0]}` in `local-scripts/check-consistency.sh:11` | Line 11 uses `${BASH_SOURCE[0]}` | - | cc:Done [b4b5ef0] |
| 58.3 | Fix misleading variable name in `harness/scripts/codex-setup-local.sh:55` — rename `repo_root` → `plugin_dir` with `# plugin-local:` comment | Variable name matches what it resolves to | - | cc:Done [b4b5ef0] |
| 58.4 | Fix `harness/scripts/generate-sprint-contract.sh` — add `__dirname` or `git rev-parse` based root resolution instead of `process.cwd()` | Script resolves project root from git, not CWD | - | cc:Done [b4b5ef0] |

### Stage 2: Test script fixes

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 58.5 | Fix `tests/validate-plugin.sh:206` — replace hardcoded `/tmp` with `${TMPDIR:-/tmp}` | No bare `/tmp` in mktemp calls | - | cc:Done [a5b2eb2] |
| 58.6 | Fix `tests/validate-plugin.sh:5` — add `set -e` to existing `set -u` and `set -o pipefail` | Line 5 reads `set -euo pipefail` | - | cc:Done [a5b2eb2] |

### Stage 3: Settings and skill description fixes

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 58.7 | Fix `harness/settings.json:21` — change `"Bash(* .env)"` to consistent `"Bash(cat .env:*)"` or equivalent `:*` syntax matching other rules | All deny rules use consistent syntax | - | cc:Done [a5b2eb2] |
| 58.8 | Fix `harness/settings.json:80-82` — move `export PATH=*`, `export LD_LIBRARY_PATH=*`, `export PYTHONPATH=*` from `deny` to `ask` | Three rules moved from deny to ask array | - | cc:Done [a5b2eb2] |
| 58.9 | Fix `harness/skills/vibecoder-guide/SKILL.md` description — rewrite to describe task shape not user attribute (remove "the user seems non-technical") | Description starts with `Use when ` and describes task shape per skill-description.md Rule 2 | - | cc:Done [a5b2eb2] |
| 58.10 | Clean up orphaned reference files — either link `harness/skills/harness-setup/references/codex.md` and `harness/skills/workflow-guide/references/commands.md` from their SKILL.md, or delete them | Every file in `references/` is linked from SKILL.md, OR orphaned files removed | - | cc:Done [a5b2eb2] |

### Stage 4: Validation

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 58.11 | Run full validation suite: `validate-plugin.sh` + `check-consistency.sh` + `check-residue.sh` + `harness validate all` | All pass with 0 failures | 56.1-56.5, 57.1-57.9, 58.1-58.10 | cc:Done [30a3866] |
| 58.12 | Record all changes under `[Unreleased]` in CHANGELOG.md in Before/After format | CHANGELOG entry added | 58.11 | cc:Done [35cb85c] |

---

## Phase 55: Path convention standardization — clear roots for all skills and scripts

Created: 2026-04-15

**Three-tier path convention** (per Opus consultation):
- **skill-local**: `${CLAUDE_SKILL_DIR}/...` — files inside the skill's own directory
- **plugin-local**: `${CLAUDE_SKILL_DIR}/../../...` — files elsewhere in the plugin (accepted `../../` since skills are always exactly at `skills/<name>/`, two levels below plugin root)
- **project-root**: `git rev-parse --show-toplevel` in scripts — never derive user project paths from script location

### Stage 1: Fix harness-release

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 55.1 | Fix `release-preflight.sh` CHANGELOG check — use `GIT_ROOT` (project root), not plugin root | Test passes: CHANGELOG found without env var override | - | cc:done |
| 55.2 | Update `release-preflight.sh` — derive `PROJECT_ROOT` from `git rev-parse --show-toplevel`, add tier comments to both scripts | `# project-root:` / `# plugin-local:` comments on key paths; tests pass | - | cc:done [50e78cd] |
| 55.3 | Update `SKILL.md` bash code blocks — replace bare `skills/harness-release/scripts/...` with `${CLAUDE_SKILL_DIR}/scripts/...` | `grep 'bash skills/' SKILL.md` returns 0 results | - | cc:done [6e3ec2c] |
| 55.4 | Update `SKILL.md` plugin-local links — standardize `${CLAUDE_SKILL_DIR}/../../` form; annotate `local-scripts/` and `validate-release-notes.sh` with `<!-- project-root -->` or `<!-- plugin-local -->` comments | All plugin-level links use consistent `../../` traversal; ownership clear in SKILL.md prose | 55.3 | cc:done [6e3ec2c] |

### Stage 2: Audit and fix all other skills

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 55.5 | Fix `harness-setup` — replace `${CLAUDE_PLUGIN_ROOT}/skills/harness-setup/scripts/...` with `${CLAUDE_SKILL_DIR}/scripts/...` | `grep 'CLAUDE_PLUGIN_ROOT' harness/skills/harness-setup/SKILL.md` returns 0 results | 55.3 | cc:done [0125611] |
| 55.6 | Audit `references/` links across all 28 SKILL.md files — ensure all use `${CLAUDE_SKILL_DIR}` | `grep -r 'references/' harness/skills/*/SKILL.md \| grep -v CLAUDE_SKILL_DIR` returns 0 results | 55.5 | cc:done [0125611] |
| 55.7 | Audit `scripts/` references across all SKILL.md files — annotate each as skill-local / plugin-local / project-root | `grep -r 'scripts/' harness/skills/*/SKILL.md \| grep -v CLAUDE_SKILL_DIR \| grep -v '^\#'` reviewed and classified | 55.5 | cc:done [0125611] |

### Stage 3: Document and enforce

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 55.8 | Create `.claude/rules/path-conventions.md` — document the three-tier convention with examples | Rule file exists; covers skill-local, plugin-local, project-root with code snippets | 55.6, 55.7 | cc:done [0125611] |
| 55.9 | Add path lint check to `validate-plugin.sh` — flag bare relative paths in bash code blocks in SKILL.md files | New check section passes on current HEAD | 55.8 | cc:done [0125611] |
| 55.10 | Run full validation suite (`validate-plugin.sh` + `check-consistency.sh` + `check-residue.sh`) | All pass with 0 failures | 55.9 | cc:done [0125611] |
| 55.11 | Record changes under `[Unreleased]` in CHANGELOG.md | Entry added | 55.10 | cc:done [7b0fc70] |

---

## Phase 52: Marketplace restructure — move plugin files to `harness/` subfolder

Created: 2026-04-15

Goal: Restructure the repo from single-plugin (`source: "./"`) to multi-plugin marketplace (`source: "./harness/"`). Move all harness-plugin-specific directories into a `harness/` subfolder. Keep repo-level files (docs, tests, Go source, CI, README, CHANGELOG, Plans.md) at root. Move `assets/` under `docs/`.

Design decisions (confirmed with Opus agent):
- `.claude-plugin/` at root keeps ONLY `marketplace.json`; plugin-specific `plugin.json`, `hooks.json`, `settings.json` move to `harness/.claude-plugin/`
- `.claude/rules/` stays at root (Claude Code reads rules from project root; SSOT for both dev and distribution)
- `local-scripts/` stays at root (dev/CI scripts for this repo)
- `.claude/memory/`, `.claude/settings.json`, `.claude/state/`, `.claude/sessions/`, `.claude/logs/` stay at root (project-level)
- `.claude/skills/`, `.claude/agents/`, `.claude/output-styles/` stay at root (project-level, not distributed with plugin)
- `hooks/` directory eliminated — canonical hooks.json is `harness/hooks/hooks.json`
- `VERSION`, `harness.toml` move to `harness/` (plugin-specific metadata)
- `benchmarks/` stays at root alongside `tests/`
- Config files (`claude-code-harness.config.*`) move to `harness/`
- CLAUDE.md stays at root; path references updated

### Batch 1: Create harness/ and move plugin directories

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 52.1 | Move `skills/` → `harness/skills/` | `harness/skills/` has all 25+ skill dirs; root `skills/` removed | -  | cc:Done [c19c6c7] |
| 52.2 | Move `agents/` → `harness/agents/` | `harness/agents/` has all agent .md files; root `agents/` removed | -  | cc:Done [c19c6c7] |
| 52.3 | Move `templates/` → `harness/templates/` | `harness/templates/` has codex/, opencode/, codex-skills/, modes/ | -  | cc:Done [c19c6c7] |
| 52.4 | Move `scripts/` → `harness/scripts/` | `harness/scripts/` has all script files; root `scripts/` removed | -  | cc:Done [c19c6c7] |
| 52.5 | Move `bin/` → `harness/bin/` | `harness/bin/` exists; root `bin/` removed | -  | cc:Done [c19c6c7] |
| 52.6 | Move `output-styles/` → `harness/output-styles/` | `harness/output-styles/` has all style files; root removed | -  | cc:Done [c19c6c7] |
| 52.7 | Move `workflows/` → `harness/workflows/` | `harness/workflows/` has all workflow files; root removed | -  | cc:Done [c19c6c7] |
| 52.8 | Move `VERSION`, `harness.toml`, `claude-code-harness.config.*` → `harness/` | Files in `harness/`; root copies removed | -  | cc:Done [c19c6c7] |

### Batch 2: Restructure .claude-plugin/ and marketplace config

Per official plugin docs: `settings.json` at plugin root; `hooks/hooks.json` at plugin root; only `plugin.json` inside `.claude-plugin/` — but we don't use plugin.json so it gets deleted.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 52.9 | Delete `.claude-plugin/plugin.json` (unused) | File no longer exists; root `.claude-plugin/` keeps only `marketplace.json` | -  | cc:Done [c19c6c7] |
| 52.10 | Move entire `hooks/` folder → `harness/hooks/` (hooks.json, BEST_PRACTICES.md, *.sh scripts) | `harness/hooks/` has all files; root `hooks/` removed | -  | cc:Done [c19c6c7] |
| 52.11 | Move `.claude-plugin/settings.json` → `harness/settings.json` (plugin root) | `harness/settings.json` exists; `.claude-plugin/settings.json` removed | -  | cc:Done [c19c6c7] |
| 52.12 | Update `marketplace.json`: `source: "./"` → `source: "./harness/"` and `outputStyles: "./harness/output-styles/"` | marketplace.json points to `./harness/` | 52.1–52.11  | cc:Done [c19c6c7] |

### Batch 3: Move assets and clean up empty dirs

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 52.13 | Move `assets/` → `docs/assets/` | `docs/assets/` has all SVGs; root `assets/` removed | -  | cc:Done [c19c6c7] |
| 52.14 | Delete empty `codex/` and `opencode/` dirs if still present | No empty ghost directories | -  | cc:Done [c19c6c7] |

### Batch 4: Update all path references

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 52.15 | Update `CLAUDE.md` path references | All paths in CLAUDE.md resolve correctly | 52.1–52.14  | cc:Done [c19c6c7] |
| 52.16 | Update `harness.toml` internal paths | `harness sync` produces correct output | 52.8  | cc:Done [c19c6c7] |
| 52.17 | Update `scripts/sync-version.sh` — read `VERSION` from `harness/VERSION`, write `harness/.claude-plugin/marketplace.json` | Version sync works from new locations | 52.4, 52.8  | cc:Done [c19c6c7] |
| 52.18 | Update `build-binary.sh` — change output from `bin/` → `harness/bin/`; no hook path changes needed (`CLAUDE_PLUGIN_ROOT` = harness/) | Binary lands in `harness/bin/harness-*` | 52.5  | cc:Done [c19c6c7] |
| 52.19 | Update `sync.go`: (1) read `harness.toml` from `harness/harness.toml`; (2) remove `syncHooksJSON` (hooks/ is now inside plugin dir, nothing to sync) | `harness sync` runs cleanly; `go test ./cmd/harness/` passes | 52.8, 52.10  | cc:Done [c19c6c7] |
| 52.20 | Update CI workflows (`.github/workflows/*.yml`) paths | CI passes | 52.1–52.14  | cc:Done [c19c6c7] |
| 52.21 | Update test files (`tests/validate-plugin.sh`, `tests/test-codex-package.sh`, etc.) paths | All tests pass with new paths | 52.1–52.14  | cc:Done [c19c6c7] |
| 52.22 | Update `.claude/scripts/check-consistency.sh` and `.claude/scripts/check-residue.sh` paths | Consistency check passes; residue check clean | 52.1–52.14  | cc:Done [c19c6c7] |
| 52.23 | Update README.md, CONTRIBUTING.md, docs/ path references | All doc links resolve | 52.1–52.14  | cc:Done [c19c6c7] |
| 52.24 | Update `docs/repository-structure.md` to reflect new layout | Matches actual directory tree | 52.1–52.14  | cc:Done [c19c6c7] |
| 52.25 | Update skill SKILL.md files that reference `${CLAUDE_PLUGIN_ROOT}/scripts/` or sibling paths | Skills resolve correct paths | 52.1, 52.4  | cc:Done [c19c6c7] |
| 52.26 | Update `deleted-concepts.yaml` with old root-level paths | `check-residue.sh` 0 detections on HEAD | 52.1–52.14  | cc:Done [c19c6c7] |
| 52.27 | Update `.gitignore` — replace `bin/harness-*` with `harness/bin/harness-*` | gitignore covers new binary path | 52.5  | cc:Done [c19c6c7] |

### Batch 5: Validation

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 52.28 | Full validation: `validate-plugin.sh` + `check-consistency.sh` + `check-residue.sh` + `test-codex-package.sh` | All pass (existing sandbox failures excepted) | 52.15–52.27  | cc:Done [c19c6c7] |
| 52.29 | Update CHANGELOG `[Unreleased]` with Phase 52 entry | Before/After documented | 52.28  | cc:Done [c19c6c7] |

---

## Phase 53: Add Makefile for local development

Created: 2026-04-15

Goal: Create a Makefile at the repo root to surface common dev/CI tasks as simple `make` targets. Includes validation, consistency checks, benchmark runs, and Go build.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 53.1 | Create `Makefile` with `validate`, `check`, `test`, `bench`, `build`, `lint` targets | `make validate` runs `tests/validate-plugin.sh`; `make check` runs `local-scripts/check-consistency.sh`; `make test` runs both; `make bench` runs `benchmarks/breezing-bench/run.sh`; `make build` runs Go binary build; `make lint` runs residue + skill-audit checks | - | cc:Done [a4aee1a] |
| 53.2 | Add `make` usage to CONTRIBUTING.md Testing section | CONTRIBUTING.md references `make test` as the recommended pre-submit check | 53.1 | cc:Done [990c129] |

---

## Phase 54: CI → Makefile — replace direct script calls with make targets

Created: 2026-04-15

Goal: Update `.github/workflows/validate-plugin.yml` to call `make` targets instead of raw script paths. Add missing make targets for CI-only steps (`version-bump`, `codex-test`). Fix stale paths in `compatibility-check.yml`. Do not touch hooks.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 54.1 | Add `version-bump` and `codex-test` targets to Makefile | `make version-bump` runs `local-scripts/check-version-bump.sh`; `make codex-test` runs `tests/test-codex-package.sh` | - | cc:Done [fa981b1] |
| 54.2 | Update `validate-plugin.yml` to use make targets | Steps use `make version-bump`, `make validate`, `make check`, `make codex-test` instead of `bash ./…` | 54.1 | cc:Done [fa981b1] |
| 54.3 | Fix stale paths in `compatibility-check.yml` (Phase 52 leftover) | Paths prefixed with `harness/`; workflow triggers updated | - | cc:Done [fa981b1] |

---

## Future Considerations

(none currently)

---

## Phase 51: Eliminate mirror directories — setup-time copy replaces build-time sync

Created: 2026-04-15

Goal: Delete `codex/`, `opencode/`, and `skills-codex/` directories. Move their config/templates to `templates/codex/` and `templates/opencode/`. Replace the mirror sync machinery with `harness-setup codex`, `harness-setup opencode`, and `harness-setup duo` subcommands that copy skills from the plugin's `skills/` to the user's project at setup time.

Design decisions (confirmed with opus agent):
- Codex: patch `disable-model-invocation: true` into SKILL.md frontmatter at copy time
- OpenCode: copy skills as-is (no frontmatter stripping — opencode ignores unknown fields)
- `skills-codex/`: move to `templates/codex-skills/` as codex-native skill overrides (breezing, harness-work); overlaid on top of skills/ copies during codex setup
- AGENTS.md: static template pointing to CLAUDE.md + agent role table (not generated)
- CI: replace mirror sync checks with template existence + setup idempotency tests

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 51.1 | Create `templates/codex/` with `config.toml`, `rules/harness.rules`, `.codexignore`, `AGENTS.md`, `README.md` from `codex/` | Files exist in `templates/codex/`; content matches originals | - | cc:done |
| 51.2 | Move opencode config to `templates/opencode/`: add `opencode.json`, `AGENTS.md`, `README.md`; commands already exist there | `templates/opencode/` has all config + commands | - | cc:done |
| 51.3 | Delete `codex/` directory | `codex/` no longer exists; git rm clean | 51.1 | cc:done |
| 51.4 | Delete `opencode/` directory | `opencode/` no longer exists | 51.2 | cc:done |
| 51.5 | Move `skills-codex/` → `templates/codex-skills/` (codex-native skill overrides: breezing, harness-work) | `templates/codex-skills/` exists with same content; `skills-codex/` removed | - | cc:done |
| 51.6 | Implement `harness-setup codex` subcommand in SKILL.md | Checks codex installed; copies `templates/codex/*` → `.codex/`; copies `skills/` → `.codex/skills/` with `disable-model-invocation: true` patch; then overlays `templates/codex-skills/` → `.codex/skills/` (overrides same-name skills with codex-native variants) | 51.1, 51.3, 51.5 | cc:done |
| 51.7 | Implement `harness-setup opencode` subcommand in SKILL.md | Checks opencode installed; copies `templates/opencode/*` → `.opencode/`; copies `skills/` → `.opencode/skills/` as-is | 51.2, 51.4 | cc:done |
| 51.8 | Implement `harness-setup duo` subcommand | Runs both codex + opencode setup | 51.6, 51.7 | cc:done |
| 51.9 | Remove mirror sync scripts: `sync-skill-mirrors.mjs`, `build-opencode.mjs`, `sync-skills.mjs`, `validate-opencode.mjs` | Scripts deleted; no remaining references | 51.3, 51.4 | cc:done |
| 51.10 | Remove mirror sync CI: update `compatibility-check.yml`, `check-consistency.sh` mirror section, `validate-plugin.sh` opencode refs | CI passes without mirror checks | 51.9 | cc:done |
| 51.11 | Add template existence check to `validate-plugin.sh` | `templates/codex/config.toml` and `templates/opencode/opencode.json` verified in CI | 51.10 | cc:done |
| 51.12 | Add `codex/`, `opencode/`, `skills-codex/` to `deleted-concepts.yaml` | `check-residue.sh` 0 detections on HEAD | 51.3–51.5 | cc:done |
| 51.13 | Update `tests/test-codex-package.sh` — remove refs to deleted paths | Test passes; no references to `codex/.codex/skills/` | 51.3 | cc:done |
| 51.14 | Update CHANGELOG [Unreleased] and run full validation | `validate-plugin.sh` + `check-consistency.sh` + `check-residue.sh` all pass | 51.1–51.13 | cc:done |

---

## Phase 50: Refocus skills/ on software development — move creative/content skills

Created: 2026-04-15

Goal: Move non-software-development skills (`allow1`, `generate-slide`, `generate-video`) from `skills/` to `.claude/skills/`, move `video-scene-generator.md` agent to `.claude/agents/`, and relocate `skills/routing-rules.md` to `.claude/rules/`. Remove their codex/opencode mirrors and update consistency checks.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 50.1 | Move `skills/allow1` → `.claude/skills/allow1`; remove codex/opencode mirrors | `.claude/skills/allow1/SKILL.md` exists; no mirror dirs remain; consistency check passes | - | cc:done [1a2fa24] |
| 50.2 | Move `skills/generate-slide` → `.claude/skills/generate-slide`; remove mirrors | `.claude/skills/generate-slide/SKILL.md` exists; no mirror dirs remain | - | cc:done [1a2fa24] |
| 50.3 | Move `skills/generate-video` → `.claude/skills/generate-video`; remove mirrors | `.claude/skills/generate-video/SKILL.md` exists; no mirror dirs remain | - | cc:done [1a2fa24] |
| 50.4 | Move `agents/video-scene-generator.md` → `.claude/agents/video-scene-generator.md` | File exists in new location; removed from `agents/` | - | cc:done [1a2fa24] |
| 50.5 | Move `skills/routing-rules.md` → `.claude/rules/skill-routing-rules.md` | File in `.claude/rules/`; update any references | - | cc:done [1a2fa24] |
| 50.6 | Update `build-opencode.mjs` skipSkills; confirm consistency check passes | 0 mirror check errors | 50.1–50.3 | cc:done [1a2fa24] |
| 50.7 | Add moved skills/agents to `deleted-concepts.yaml` residue scan | `check-residue.sh` 0 detections on HEAD | 50.1–50.5 | cc:done [1a2fa24] |
| 50.8 | Update CHANGELOG and validate | CHANGELOG has [Unreleased] entry; 0 residue violations | 50.1–50.7 | cc:done [1a2fa24] |

---

## Phase 49: harness-setup build-from-source + hooks.json SSOT consolidation

Created: 2026-04-14

Goal: Replace the network-dependent binary download with a local Go build, deduplicate the deny list in harness.toml, and make `.claude-plugin/hooks.json` the single source of truth by symlinking `hooks/hooks.json` to it.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 49.1 | Replace `download-binary.sh` with `build-binary.sh` that compiles from Go source for current platform | `build-binary.sh` exists; script builds and installs correct arch binary; SKILL.md and hooks.json updated | - | cc:done [569bf3b] |
| 49.2 | Deduplicate `harness.toml` deny list — remove 42 redundant entries subsumed by umbrella rules (`sudo:*`, `rm -rf:*`, `git reset --hard *`, `*bitcoin*`) | No duplicate entries; `python3` duplicate check returns "none" | - | cc:done [bdd816b] |
| 49.3 | Symlink `hooks/hooks.json` → `../.claude-plugin/hooks.json`; update `syncHooksJSON` in `sync.go` to detect symlink and skip copy | `ls -la hooks/hooks.json` shows symlink; `harness sync` prints "skipped (symlinked)"; all sync tests pass | 49.2 | cc:done [108441b] |

---

