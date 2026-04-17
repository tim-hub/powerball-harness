# Powerball Harness — Plans.md

Last archive: 2026-04-17 (Phase 59–61 → `.claude/memory/archive/Plans-2026-04-17-phase59-61.md`)
Last release: v4.5.1 on 2026-04-16 (docs + hook removal + settings hardening)

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
