# Powerball Harness — Plans.md

Last archive: 2026-04-17 (Phase 59–61 → `.claude/memory/archive/Plans-2026-04-17-phase59-61.md`)
Last release: v4.5.1 on 2026-04-16 (docs + hook removal + settings hardening)

---

## Phase 74: Code-space skill search — proof-of-concept on `harness-review`

Created: 2026-04-17

**Goal**: Borrow Meta-Harness's code-space search idea (arxiv 2603.28052) — instead of hand-editing `SKILL.md`, generate variants and score them against an evaluation suite. Scope: one skill (`harness-review`), 3-5 variants, measurable outcome. If the POC improves the skill, formalize the loop; if not, record why and discard.

**Depends on**: Phase 72 (traces provide failure signal for the proposer) + Phase 73 (advisor pattern reusable for proposer scaffolding).

**Open design decision** (user input during 74.1): score function — rubric-based (0-5 across N criteria like "catches real bugs", "no false positives", "clear verdict rationale") vs pass/fail against golden verdicts on a fixed PR corpus. Rubric gives gradient signal; golden outputs are objective but binary. Pick one.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 74.1 | Define score function + build eval runner `local-scripts/eval-skill.sh <skill-dir> <eval-suite-dir>`. Runs fixed inputs through the skill, captures outputs, emits JSON score report | `eval-skill.sh harness/skills/harness-review tests/skill-eval/harness-review` prints a reproducible score | 72.1 | cc:TODO |
| 74.2 | Build evaluation suite for `harness-review` at `tests/skill-eval/harness-review/` — 5 PR diffs (2 with real bugs, 2 clean, 1 scope-creep) + expected verdicts | `ls tests/skill-eval/harness-review/*.{diff,expected.json}` shows 10 files; running baseline skill against suite produces a score | 74.1 | cc:TODO |
| 74.3 | Write proposer script `local-scripts/propose-skill-variants.sh <skill-dir>`. Given SKILL.md + eval output + recent traces, generates 3 SKILL.md variants to `/tmp/skill-variants/harness-review-v{1,2,3}/SKILL.md` via Claude subagent | Running against a broken baseline (intentionally degraded `harness-review`) produces 3 syntactically valid SKILL.md files with meaningful diffs | 74.1 | cc:TODO |
| 74.4 | Run end-to-end search loop: baseline → generate 3 variants → score each → pick winner. Emit report at `.claude/state/code-search/harness-review-<YYYY-MM-DD>.md` with scores, diffs, chosen winner | Report exists; winner's score ≥ baseline's score; report includes rationale | 74.2, 74.3 | cc:TODO |
| 74.5 | Decision gate: if winner beats baseline by ≥10%, promote to main and add pattern to `patterns.md`. Otherwise document the null result. Update CHANGELOG [Unreleased] | Either a commit promoting the variant OR a patterns.md entry "code-space search POC attempted; no gain — reasons X, Y" | 74.4 | cc:TODO |

---

## Phase 73: Advisor full-history inspection — scoped raw-source loader

Created: 2026-04-17

**Goal**: Upgrade the Advisor (harness/agents/advisor.md) from `history.jsonl`-only reads to scoped raw-source inspection — recent session-log excerpts, relevant git diffs, and Phase 72 execution traces. The Meta-Harness paper's core claim is that compressed feedback (summaries) loses the causal signal needed to fix problems; raw sources restore it.

**Depends on**: Phase 72 (trace files are the richest new source).

**Non-goals**: Replacing `patterns.md` or `decisions.md`. Those remain summary SSOTs for humans. Advisor just gets more context *options* for edge cases where summaries don't suffice.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 73.1 | Extend advisor input schema with optional `context_sources` array (values: `session_log`, `git_diff`, `trace`, `patterns`). Caller chooses what to load per invocation | advisor.md schema updated; example in advisor.md shows all 4 source values | - | cc:TODO |
| 73.2 | Implement scoped loader `harness/scripts/advisor-load-context.sh <task_id> <sources...>`. Returns ≤10KB per source by default: session-log entries mentioning task_id, `git diff` since task-start commit, trace JSONL for this task, patterns.md sections matching task tags | Running with `--task 72.1 --sources trace,git_diff` on a real task returns valid excerpts totaling <20KB | 72.1 | cc:TODO |
| 73.3 | Update advisor.md prompt: add "When each source helps" table (e.g. trace → repeated_failure; git_diff → high_risk_preflight; session_log → plateau_before_escalation). Preserve PLAN/CORRECTION/STOP schema backward-compat | advisor.md passes `./tests/validate-plugin.sh`; schema section unchanged | 73.1 | cc:TODO |
| 73.4 | Keep duplicate suppression first — advisor only loads raw context when no `history.jsonl` cache hit exists for (task_id, reason_code, error_signature) tuple | Unit test: cached match → no context loaded; miss → context loader invoked once | 73.2 | cc:TODO |
| 73.5 | Add integration test: seed a trace showing "fix was a single-file rename in N files"; assert advisor returns `CORRECTION` with suggested_approach mentioning the rename pattern | Test at `tests/advisor/test-full-history-correction.sh` passes | 73.2, 73.3 | cc:TODO |

---

## Phase 72: Execution trace retention — per-task causal log

Created: 2026-04-17

**Goal**: Introduce structured per-task execution traces so future agents can reason causally about what was tried and what failed — not just what was decided. Complements `decisions.md` (why) and `patterns.md` (how) with a third layer: **attempts** (what actually happened).

**Motivation**: Borrowed from Meta-Harness (arxiv 2603.28052). The paper's proposer reads a median of 82 files per iteration — full execution history — and that raw context is what distinguishes it from template-based optimization. This project currently persists only summaries.

**Open design decision** (user input during 72.1): trace schema shape — flat JSONL where each line is one event (`tool_call`, `decision`, `error`, `fix_attempt`, `outcome`) vs nested per-attempt where each entry groups an attempt's tool_calls + outcome. Flat is simpler to append and tail; nested preserves attempt boundaries natively. Both are valid; pick one.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 72.1 | Define trace schema + storage layout. Write schema doc at `.claude/memory/schemas/trace.v1.md`. Storage: `.claude/state/traces/<task_id>.jsonl` — one file per Plans.md task. Capture: ts, event_type, tool (if tool_call), error_signature (if error), payload | schema doc exists; includes ≥1 concrete example per event type | - | cc:Done [a6bfd40] |
| 72.2 | Implement Go emitter `go/internal/trace/writer.go` with `AppendEvent(taskID, event)` — atomic append via flock, fsync on close. Add unit test for concurrent writers to different task files | `go test ./go/internal/trace/...` passes; concurrent test with 10 goroutines writing to 10 files produces 10 valid JSONL files with no corruption | 72.1 | cc:Done [26cb9d6] |
| 72.3 | Wire PostToolUse hook → trace emitter. Derive `task_id` from `.claude/state/plans-state.json` (active task). Skip if no active task | `/harness-work` on a test task produces a populated `.claude/state/traces/<task_id>.jsonl`; no-task sessions produce no trace file | 72.2 | cc:TODO |
| 72.4 | Archive policy script `harness/skills/maintenance/scripts/archive-traces.sh`. Moves traces for `cc:done` tasks >30 days old into `.claude/memory/archive/traces/YYYY-MM/`. Wire into existing maintenance skill | Script runs idempotently; second run is a no-op; files moved retain JSONL validity | 72.2 | cc:TODO |
| 72.5 | Documentation: add "Trace retention" section to `harness/README.md` memory diagram (L0/L1/L2) explaining where traces fit. Update `.claude/memory/patterns.md` with a `P10: Per-task execution traces` pattern. Add CHANGELOG [Unreleased] entry | README memory section mentions traces; patterns.md has P10; CHANGELOG entry present | 72.1, 72.3 | cc:TODO |

---

## Phase 71: Skill description reformat — capability summary + when_to_use field

Created: 2026-04-17

**Goal**: Reformat frontmatter in all 28 harness skills with two changes:
1. `description` → `"<Capability summary>. Use when <trigger>."` — remove `Do NOT load for` exclusions; lead with what the skill does
2. Add new `when_to_use` field — holds trigger phrases and example user requests (e.g. `"create a plan", "add a task", "mark done"`)

**Character budget**: `len(description) + len(when_to_use)` must stay ≤ 1500 chars (soft limit), hard limit 1536.

**Example**:
```yaml
description: "Plans and tracks tasks in Plans.md. Use when creating plans, adding tasks, or checking progress."
when_to_use: "create a plan, add a task, mark task done, check progress, where am I"
```

| Task | Description | DoD | Status |
|------|-------------|-----|--------|
| 71.1 | Update core workflow skills: `harness-plan`, `harness-work`, `harness-review`, `harness-release`, `harness-setup`, `harness-sync` | All 6 skills have new `description` + `when_to_use`; no `Do NOT load for` remains | cc:done |
| 71.2 | Update session skills: `session`, `session-init`, `session-control`, `session-state`, `session-memory` | All 5 updated | cc:done |
| 71.3 | Update infra/ops skills: `ci`, `deploy`, `maintenance`, `gogcli-ops`, `breezing`, `harness-loop` | All 6 updated | cc:done |
| 71.4 | Update feature/content skills: `auth`, `crud`, `ui`, `agent-browser`, `notebook-lm`, `writing-changelog` | All 6 updated | cc:done |
| 71.5 | Update guide/meta skills: `principles`, `vibecoder-guide`, `workflow-guide`, `cc-cursor-cc`, `memory` | All 5 updated | cc:done |
| 71.6 | Validate character budgets and run `./tests/validate-plugin.sh`; add CHANGELOG [Unreleased] entry | All 28 skills: `len(description)+len(when_to_use)` ≤ 1536; validation passes | cc:done |
| 71.7 | Fix `argument-hint` style in 4 inconsistent skills (`harness-work`, `harness-loop`, `harness-review`, `agent-browser`): consolidate separate `[a] [b]` brackets into single `[a\|b\|c]` style matching all other skills | All 4 updated; grep for `] [` in argument-hints returns 0 matches | cc:done |

---

## Phase 70: Code-review quality fixes (struct-field DI, hook ordering, go.work)

Created: 2026-04-17

**Goal**: Address three non-blocking code-review observations: move package-level test mocks to struct fields in `plans_watcher.go`, add an explicit hook-ordering comment in `post_batch.go`, and add `go.work` at the repo root to silence gopls false positives.

| Task | Description | DoD | Status |
|------|-------------|-----|--------|
| 70.1 | Refactor `plans_watcher.go`: replace package-level `flockCall`/`sleepCall`/`exitFailClosed` vars with `plansWatcherDeps` struct + `plansWatcher` type; keep `HandlePlansWatcher` as thin public wrapper | `go test ./go/internal/hookhandler/...` passes; `t.Parallel()`-safe | cc:done |
| 70.2 | Add hook-ordering comment to `postBatchHooks()` in `post_batch.go` explaining that slice position determines which output CC sees | Comment present; no test changes needed | cc:done |
| 70.3 | Add `go.work` at repo root pointing to `./go` to silence gopls false positives | `gopls` resolves imports from repo root without errors | cc:done |

---

## Phase 69: Upstream sync — PR #81 (Codex loop runtime + concurrency fixes) + PR #82 (mirror check fix)

Created: 2026-04-17

**Goal**: Sync missing upstream changes from Chachamaru127/claude-code-harness PR #81 and PR #82 that preceded PR #83 (our Phase 62 Advisor Strategy). PR #81 provides the foundational `harness-loop` runtime (ScheduleWakeup, pacing, sprint-contracts, flock locking, plateau detection) that PR #83's advisor integration was built on top of. PR #82 is a 1-line mirror-check fix.

**Reference**: Chachamaru127/claude-code-harness PR #81 (merged 2026-04-16T06:42Z), PR #82 (merged 2026-04-16T06:53Z).

**Path mapping**: upstream `skills/` → `harness/skills/`, `scripts/` → `harness/scripts/`, `opencode/skills/` → `harness/templates/opencode/skills/`, `go/` → `go/` (import path: `github.com/tim-hub/powerball-harness/...`).

| Task | Description | Status |
|------|-------------|--------|
| 69.1–69.4 | harness-loop SKILL.md, references/flow.md, template variants, advisor layer | cc:Done [b51bef4] |
| 69.5 | Go plans_watcher concurrent safety (flock, fail-closed) | cc:Done [1b4c955] |
| 69.6 | Go sprint_contract.go + test | cc:Done [1b4c955] |
| 69.7 | Go CLI: codex_loop.go, sprint_contract.go, main.go, doctor.go | cc:Done [83b9da4] |
| 69.8–69.9 | harness-review: --ui-rubric flag, ui-rubric.md reference | cc:Done [0268737] |
| 69.10 | New scripts: codex-loop.sh, auto-checkpoint.sh, browser-review-runner.sh, detect-review-plateau.sh | cc:Done [83b9da4] |
| 69.11 | Script mods: plans-watcher flock, record-review-calibration, release-preflight, run-contract-review-checks, write-review-result, sync-version | cc:Done [83b9da4] |
| 69.12 | New script: sync-skill-mirrors.sh (PR #82 fix pre-applied) | cc:Done [83b9da4] |
| 69.13 | New maintenance skill (SKILL.md + references/cleanup.md) | cc:Done [0268737] |
| 69.14 | Port test files: test-auto-checkpoint.sh, test-codex-loop-cli.sh, test-detect-review-plateau.sh, test-harness-loop-flow.sh, test-harness-loop-guard.sh + integration tests | cc:Done [07f41c7] |
| 69.15 | Full validation pass + CHANGELOG [Unreleased] entry | cc:Done [2ad37ad] |
| 69.16 | Restore maintenance skill references in docs + hook messages | cc:Done [0268737] |

---

## Phase 68: Go guardrail — structured secret detection in post-tool pipeline

Created: 2026-04-16

| Task | Description | DoD | Status |
|------|-------------|-----|--------|
| 68.1 | Add structured secret regex patterns to `securityPatterns` in `post_tool.go`: JWT, Anthropic/OpenAI keys, AWS access keys, GitHub tokens, Stripe keys | `go test ./internal/guardrail/...` passes; each pattern fires a warning | cc:Done [a24dd6c] |
| 68.2 | Add test cases for each new pattern in `post_tool_test.go` | All new patterns have ≥1 positive and ≥1 negative test | cc:Done [a24dd6c] |

---

## Phase 67: harness-release skill cleanup

Created: 2026-04-16

| Task | Description | Status |
|------|-------------|--------|
| 67.1 | Gate Phase 5 Codex symlink check on `command -v codex`; remove --announce subcommand | cc:Done [b83340a] |
| 67.2–67.5 | Move check-consistency.sh, check-residue.sh, validate-release-notes.sh to harness-release/scripts/; update docs | cc:Done [b83340a] |

---

## Phase 66: Simplify release tooling

Created: 2026-04-16

| Task | Description | Status |
|------|-------------|--------|
| 66.1 | Replace CHANGELOG reference-link block with inline links | cc:Done [681bfc6] |
| 66.2 | Remove template version syncing from sync-version.sh | cc:Done [681bfc6] |
| 66.3 | Narrow version consistency check in check-consistency.sh and validate-plugin.sh | cc:Done [00b2afe] |

---

## Phase 65: harness-sync workflow reorder

Created: 2026-04-16

| Task | Description | Status |
|------|-------------|--------|
| 65.1 | Run Plans.md drift check before harness-work, not after review | cc:Done [479e3d3] |

---

## Phase 64: Agent orchestration optimizations

Created: 2026-04-16

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 64.1 | Parallelize reviewer invocations in `harness-work --parallel` mode | Review phase wallclock ≤ `max(single-review-time) + 10s overhead` | Phase 62 | cc:Done [1c83b89] |
| 64.2 | Parallelize sprint-contract generation in BREEZING Phase A | Phase A duration for 5-task run ≤ `max(per-task contract time) + 10s overhead` | Phase 62 | cc:Done [1c83b89] |
| 64.3 | Validate + record deltas in `benchmarks/phase64-results.json` + CHANGELOG | All pass; results file exists | 64.1–64.2 | cc:Done [1c83b89] |

---

## Phase 63: Hook chain optimizations

Created: 2026-04-16

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 63.1 | Consolidate `hook memory-bridge` invocations to single entry point with `--mode` flag | ≤1 memory-bridge subprocess spawn per hook event | Phase 62 | cc:Done [760d54f] |
| 63.2 | Make POST_BATCH (8 Write/Edit/Task hooks) concurrent fan-out in Go binary | PostToolUse batch wallclock ≤ 40% of Phase 62 baseline; `go test -race` passes | Phase 62 | cc:Done [760d54f] |
| 63.3 | Parallelize PreToolUse independent hooks on Write\|Edit | Total PreToolUse wallclock ≤ `max(individual hook time) + 20ms overhead` | Phase 62 | cc:Done [760d54f] |
| 63.4 | Validate + record deltas in `benchmarks/phase63-results.json` + CHANGELOG | Both scripts pass; results file exists | 63.1–63.3 | cc:Done [760d54f] |

---

## Phase 62: Go guardrail engine optimizations

Created: 2026-04-16

| Task | Description | DoD | Status |
|------|-------------|-----|--------|
| 62.1 | Capture baseline benchmarks in `benchmarks/phase62-baseline.json` | 5 metrics recorded | cc:Done [1ab7728] |
| 62.2 | Add resolved-path cache to `isProtectedPath` (max 256 entries) | Cache hit ≥80% on 50-edit benchmark | cc:Done [53892a4] |
| 62.3 | Split `detectTampering` by file type — skip non-test files | Post-tool latency on non-test file ≥40% faster | cc:Done [8fd5a43] |
| 62.4 | Fast-path in `normalizeCommand` for already-normalized commands | Zero allocs + early return for simple commands | cc:Done [53892a4] |
| 62.5 | Validate + `benchmarks/phase62-results.json` + CHANGELOG | All pass | cc:Done |

---

## Future Considerations

(none currently)

---
