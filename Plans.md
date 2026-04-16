# Powerball Harness — Plans.md

Last archive: 2026-04-16 (Phase 55-56 → `.claude/memory/archive/Plans-2026-04-16-phase55-56.md`)
Last release: v4.5.1 on 2026-04-16 (docs + hook removal + settings hardening)

---

## Phase 68: Go guardrail — structured secret detection in post-tool pipeline

Created: 2026-04-16

Goal: Enhance the existing `securityPatterns` in `go/internal/guardrail/post_tool.go` to detect structured secret formats written via Write/Edit tools. Currently only catches `password = "..."` style assignments. Add regex patterns for JWT tokens, API key prefixes (`sk-`, `sk-ant-`, `AKIA*`, `ghp_*`, etc.), and other common secret formats. Advisory warnings (post-tool), not blocking.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 68.1 | Add structured secret regex patterns to `securityPatterns` in `post_tool.go`: JWT (`eyJ[A-Za-z0-9_-]{20,}\.eyJ`), Anthropic keys (`sk-ant-[a-zA-Z0-9-]{20,}`), OpenAI keys (`sk-[a-zA-Z0-9]{20,}`), AWS access keys (`AKIA[0-9A-Z]{16}`), GitHub tokens (`gh[pousr]_[A-Za-z0-9_]{36,}`), Stripe keys (`[sr]k_live_[a-zA-Z0-9]{20,}`), generic long base64 secrets (`['"][A-Za-z0-9+/=]{40,}['"]` near `key\|secret\|token` context) | `go test ./internal/guardrail/...` passes; each pattern fires a warning on matching content | - | cc:Done [a24dd6c] |
| 68.2 | Add test cases for each new pattern in `post_tool_test.go` — positive matches and negative cases (e.g. normal base64 images should not trigger) | All new patterns have ≥ 1 positive and ≥ 1 negative test; `go test -run TestPostTool` passes | 68.1 | cc:Done [a24dd6c] |

---

## Phase 67: harness-release skill cleanup — conditional codex, remove announce, consolidate scripts

Created: 2026-04-16

Goal: Clean up the harness-release skill: gate Codex symlink check on codex availability, remove the --announce subcommand (unused), and consolidate release-related scripts into the skill's own scripts/ folder for better cohesion.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 67.1 | Update SKILL.md: Phase 5 (Verify Codex Symlinks) — wrap in `command -v codex` guard so it only runs when Codex CLI is installed; Phase 10 (--announce) — remove entirely from subcommands table, argument-hint, Quick Reference, --dry-run section, and Phase 10 body | SKILL.md no longer references `--announce`; Phase 5 code block includes `if command -v codex` guard | - | cc:Done [b83340a] |
| 67.2 | Move `local-scripts/check-consistency.sh` → `harness/skills/harness-release/scripts/check-consistency.sh`. Update all references: Makefile, CONTRIBUTING.md, CLAUDE.md, harness-release SKILL.md, `harness/scripts/generate-sprint-contract.sh` | `make check` calls the new path; old path removed; `bash harness/skills/harness-release/scripts/check-consistency.sh` passes | - | cc:Done [b83340a] |
| 67.3 | Move `local-scripts/check-residue.sh` → `harness/skills/harness-release/scripts/check-residue.sh`. Update all references: Makefile, `tests/validate-plugin.sh`, `.claude/rules/migration-policy.md`, `.claude/rules/deleted-concepts.yaml`, harness-release SKILL.md | `make residue` calls the new path; old path removed; `bash harness/skills/harness-release/scripts/check-residue.sh` passes | - | cc:Done [b83340a] |
| 67.4 | Move `harness/scripts/validate-release-notes.sh` → `harness/skills/harness-release/scripts/validate-release-notes.sh`. Update reference in harness-release SKILL.md (currently uses `${CLAUDE_SKILL_DIR}/../../scripts/validate-release-notes.sh`) | SKILL.md uses `${CLAUDE_SKILL_DIR}/scripts/validate-release-notes.sh`; old location removed | - | cc:Done [b83340a] |
| 67.5 | Update `docs/repository-structure.md` to reflect the 3 script moves and verify `check-consistency.sh` still passes after all moves | `bash harness/skills/harness-release/scripts/check-consistency.sh` exits 0; `docs/repository-structure.md` matches actual layout | 67.2, 67.3, 67.4 | cc:Done [b83340a] |

---

## Phase 66: Simplify release tooling — CHANGELOG links, version sync scope

Created: 2026-04-16

Goal: Remove unnecessary coupling between the plugin version and template `_harness_version` fields. Templates are scaffolded once into user projects and should allow backward compatibility — they don't need to track every plugin patch. Also simplify CHANGELOG by replacing the bottom-of-file reference-link block with inline links at each release header.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 66.1 | Replace CHANGELOG reference-link block with inline links. Remove the ~80 `[X.Y.Z]: https://...` lines at the bottom of `CHANGELOG.md`; convert each `## [X.Y.Z] - DATE` heading to `## [X.Y.Z](https://github.com/tim-hub/powerball-harness/compare/vPREV...vX.Y.Z) - DATE` with an inline URL. Update `[Unreleased]` the same way | `grep -c '^\[.*\]: https://github.com' CHANGELOG.md` returns 0; every `## [X.Y.Z]` heading is a clickable inline link; CI "compare link missing" check passes or is removed | - | cc:Done [681bfc6] |
| 66.2 | Remove template version syncing from `sync-version.sh`. The `sync` and `bump` subcommands currently update `_harness_version` in all `harness/templates/*.template` files and `template-registry.json`. Remove that loop — only sync `harness/VERSION`, `harness/harness.toml`, and `.claude-plugin/marketplace.json` | `bash harness/skills/harness-release/scripts/sync-version.sh bump` only modifies VERSION, harness.toml, and marketplace.json; `grep -rl '_harness_version.*4\.5' harness/templates/` still shows the old version (unchanged) | - | cc:Done [681bfc6] |
| 66.3 | Narrow version consistency check in `local-scripts/check-consistency.sh` and `tests/validate-plugin.sh`. Remove checks that compare `_harness_version` in templates or `template-registry.json` against `VERSION`. Only validate: VERSION == harness.toml version == marketplace.json plugin version | Version drift between templates and VERSION no longer fails `check-consistency.sh` or `validate-plugin.sh`; drift between VERSION / harness.toml / marketplace.json still fails | 66.2 | cc:Done [00b2afe] |

---

## Phase 65: harness-sync workflow reorder

Created: 2026-04-16

Goal: Catch stale Plans.md markers before implementation starts, not after review ends. This is a workflow-flow change (not a latency optimization), so it lives in its own phase rather than with the perf work.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 65.1 | Reorder `harness-sync` to run before `harness-work` (not after review). Today sync/drift-check happens post-review, which means stale Plans.md markers are only caught after implementation. Add an entry-point drift check in `harness-work` that runs a lightweight sync pass before mode selection; full sync still runs at session boundaries | `harness-work` invocation on a Plans.md with stale markers (e.g. cc:TODO for a task that's already committed) prints a drift summary and exits non-zero when ≥ 1 stale marker is detected, prompting the user to confirm before implementation starts; zero behavioral change when Plans.md is already in sync | - | cc:Done [479e3d3] |

---

## Phase 64: Agent orchestration optimizations

Created: 2026-04-16

Goal: Cut wallclock time in breezing and parallel work modes by parallelizing reviewer invocations and sprint-contract generation. Depends on Phase 62 baseline numbers for before/after comparison.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 64.1 | Parallelize reviewer invocations in `harness-work --parallel` mode. Today 2–3 workers run concurrently but feed into a single reviewer that processes verdicts one at a time. Reviewer is read-only (no Write/Edit/Bash in allowed-tools), so multiple reviewer instances can run concurrently without state conflict. Spawn one reviewer per completed worker | Review phase wallclock for 2 concurrent tasks ≤ `max(single-review-time) + 10s overhead`, measured against Phase 62 baseline; verdict outcomes identical on a fixed test scenario before vs after | Phase 62 | cc:TODO |
| 64.2 | Parallelize sprint-contract generation in BREEZING Phase A (`skills/harness-work` Lead orchestration). Today Lead generates sprint contracts for each task sequentially before spawning workers. Contracts are independent of each other; generate them concurrently and spawn workers against already-ready contracts | BREEZING Phase A duration for a 5-task run ≤ `max(per-task contract time) + 10s overhead` vs Phase 62 baseline `sum(per-task contract time)`; worker spawn order unchanged; contract content byte-identical to sequential version on a fixed fixture | Phase 62 | cc:TODO |
| 64.3 | Validate: `make validate` and `make check` pass; re-run review-phase and BREEZING Phase A metrics and record deltas in `benchmarks/phase64-results.json`; add CHANGELOG `[Unreleased]` entry in Before/After format for the orchestration layer | Both scripts pass; results file exists with `{baseline_ms, optimized_ms, improvement_pct}` for both metrics; CHANGELOG entry present | 64.1–64.2 | cc:TODO |

---

## Phase 63: Hook chain optimizations

Created: 2026-04-16

Goal: Reduce subprocess count and serial latency in the shell hook chain (`harness/hooks/`, `harness/scripts/`, `go/internal/hook/`). Depends on Phase 62 baseline numbers for before/after comparison.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 63.1 | Consolidate `hook memory-bridge` invocations. Today it fires separately in PreToolUse, PostToolUse, SessionStart, and Stop — 3–4 subprocess spawns per session. Replace with a single entry point that accepts a `--mode={pre,post,start,stop,user-prompt}` flag and dispatch internally | `grep -c '"command".*memory-bridge' harness/hooks/hooks.json` shows one handler per event but all point to the same underlying binary/script; total memory-bridge subprocess spawns per hook event is ≤ 1 vs Phase 62 baseline | Phase 62 | cc:TODO |
| 63.2 | Make `POST_BATCH` (the 8 Write/Edit/Task hooks: emit-trace, auto-cleanup, track-changes, auto-test, quality-pack, plans-watcher, tdd-check, auto-broadcast) a concurrent fan-out in the Go binary (`bin/harness hook post-tool`). Today 7 of the 8 run serially in shell. Update `internal/hook/post_tool.go` to launch them as goroutines, await all, and merge output | PostToolUse batch wallclock on a representative Write/Edit drops to ≤ 40% of Phase 62 baseline; no hook output is lost or reordered; `go test -race ./internal/hook/...` passes | Phase 62 | cc:TODO |
| 63.3 | Parallelize PreToolUse independent hooks on `Write\|Edit`: `inbox-check`, secrets-scanning agent, and `browser-guide` (where applicable) have no dependencies on each other. Run them concurrently from a single dispatcher entry | Total PreToolUse wallclock on a Write ≤ `max(individual hook time) + 20ms overhead`, not the sum; verified against Phase 62 baseline | Phase 62 | cc:TODO |
| 63.4 | Validate: `make validate` and `make check` pass; re-run PostToolUse batch and PreToolUse metrics and record deltas in `benchmarks/phase63-results.json`; add CHANGELOG `[Unreleased]` entry in Before/After format for the hooks layer | Both scripts pass; results file exists with `{baseline_ms, optimized_ms, improvement_pct}` for both metrics; CHANGELOG entry present | 63.1–63.3 | cc:TODO |

---

## Phase 62: Go guardrail engine optimizations

Created: 2026-04-16

Goal: Cut latency in `go/internal/guardrail/` — the fast path that runs on every pre-tool and post-tool invocation. All three changes are purely additive (cache, gate, fast-path) with zero change to allow/deny/ask decisions or rule logic. Baseline captured here is shared by Phases 63 and 64.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 62.1 | Capture baseline benchmarks for all five metrics (pre-tool median, PostToolUse batch wallclock, SessionStart→first prompt, BREEZING Phase A for 5 tasks, post-tool tampering+security scan on a non-test file). Store raw numbers in `benchmarks/phase62-baseline.json` | File exists with all five metrics recorded as `{baseline_ms, sample_count, platform}`; results reproducible across 3 consecutive runs within ±10% | - | cc:Done [1ab7728] |
| 62.2 | Add a per-session resolved-path cache to `internal/guardrail/helpers.go::isProtectedPath`. `filepath.EvalSymlinks` currently stats the filesystem on every Write/Edit protected-path check, even for paths already resolved earlier in the same session. Cache `{inputPath → resolvedPath}` in a bounded map (max 256 entries, evict oldest on overflow) | Repeated Write/Edit on the same file in a session calls `EvalSymlinks` exactly once; cache hit ratio ≥ 80% on a 50-edit workflow benchmark; map never exceeds 256 entries; no change in deny/allow decisions on the existing rule test suite | 62.1 | cc:Done [53892a4] |
| 62.3 | Split `detectTampering` (T01–T12) and `detectSecurityRisks` by file type in `internal/guardrail/post_tool.go`. Tampering patterns are only meaningful on test and CI-config files; today all 12 tampering regexes run on every Write/Edit regardless of file path. Add an early path-based gate that skips the tampering scan for non-test files | Post-tool latency on a non-test file drops ≥ 40% vs 62.1 baseline; tampering still detected on test files (`*_test.*`, `test_*`, `*.spec.*`, `.github/workflows/*.yml`, etc.) with no regression in existing post_tool test suite | 62.1 | cc:Done [8fd5a43] |
| 62.4 | Add a fast-path to `internal/guardrail/helpers.go::normalizeCommand`. Today it runs unconditionally on every Bash command before regex matching, allocating a new string even when the command is already normalized. Add `if !strings.ContainsAny(cmd, "\t\r\n") && !strings.Contains(cmd, "  ") { return cmd }` before the allocation | Benchmark shows zero allocations and early return for simple single-spaced commands; normalize still correctly collapses whitespace when present; all existing rule tests pass unchanged | 62.1 | cc:Done [53892a4] |
| 62.5 | Validate: `go test -race ./...` passes; re-run pre-tool and post-tool metrics from 62.1 and record deltas in `benchmarks/phase62-results.json`; add CHANGELOG `[Unreleased]` entry in Before/After format for the guardrail layer | All tests pass; results file exists with `{baseline_ms, optimized_ms, improvement_pct}` for pre-tool and post-tool metrics; CHANGELOG entry present | 62.2–62.4 | cc:Done |

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
| 62.18 | Run full validation: `./tests/validate-plugin.sh`, `./local-scripts/check-consistency.sh`, `./local-scripts/check-residue.sh`, `./local-scripts/audit-skill-descriptions.sh harness/skills/harness-loop`, `harness validate agents`, `go test ./go/internal/hookhandler/...` | All pass with 0 failures | 62.1–62.17 | cc:WIP |
| 62.19 | Add `[Unreleased]` CHANGELOG entry covering: (a) new advisor agent and 4-agent model, (b) harness-loop skill, (c) breezing/harness-work advisor flags, (d) advisor trigger hook in Go engine, (e) symlink fix. Use Before/After format per `.claude/rules/github-release.md` | CHANGELOG entry present under `[Unreleased]` with Before/After sections | 62.18 | cc:TODO |

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

## Future Considerations

(none currently)

---
