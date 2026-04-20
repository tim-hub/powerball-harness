# Powerball Harness — Plans.md

Last release: v4.11.2 on 2026-04-20 (mem health tri-state fix + active-watching test policy)

---

## Phase 80: Port upstream tri-state mem health fix + active-watching test policy (v4.10.2)

Created: 2026-04-20

**Source**: Chachamaru127/claude-code-harness compare [888b195…2c6a66d](https://github.com/Chachamaru127/claude-code-harness/compare/888b19535149702fd05409b616bba9ac7111cb17...2c6a66d857d5d860d82a706ea1e6a1b0c210699f) (upstream v4.3.1 → v4.3.3). Opus agent review confirmed: Phase 77 ported the `mem health` subcommand but not the follow-up tri-state regression fix that landed upstream after our port.

**Goal**: Stop `bin/harness mem health` from returning `exit 1` for users who never installed harness-mem. Missing `~/.claude-mem/` is an opt-in-not-used state, not a failure. Port the governance rule (`active-watching-test-policy.md`) that codifies the tri-state naming convention so the next similar feature cannot regress the same way.

**Confirmed bug**: `go/cmd/harness/mem.go:100-102` returns `{Healthy: false, Reason: "not-initialized"}` + exit 1 whenever `~/.claude-mem/` is absent. Today only external callers of `bin/harness mem health` see the spurious failure (we have not yet wired mem health into the session monitor), but we are one wiring PR away from inheriting upstream's full regression.

**Non-goals**: Wiring mem health into session monitor (separate concern); porting upstream's Plans.md / CHANGELOG release sections (we run our own versioning pipeline); porting hook-wiring (Phase 77.5 already confirmed our Go handler emits `additionalContext` directly).

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 80.1 | Fix `go/cmd/harness/mem.go` tri-state: when `os.UserHomeDir()` fails OR `~/.claude-mem/` is absent, return `{Healthy: true, Reason: "not-configured"}` (exit 0). Rename reasons `not-initialized` → `not-configured` and `corrupted-settings` → `corrupted`. In the file-integrity check, accept `settings.json` OR `supervisor.json` (either readable + `json.Valid` satisfies the check); only flag `corrupted` when neither file is valid. Keep TCP probe behavior unchanged | `go test ./go/cmd/harness/... -run TestRunMemHealth -race` green; manual smoke on machine without `~/.claude-mem/` → stdout `{"healthy":true,"reason":"not-configured"}` + exit 0; manual smoke with dir + only `supervisor.json` present → healthy path reached; `go vet ./...` clean | - | cc:Done [70b6c28] |
| 80.2 | Rewrite `go/cmd/harness/mem_test.go` to assert the tri-state contract. Rename `TestRunMemHealth_NotInitialized` → `TestRunMemHealth_NotConfigured` (expect `Healthy=true`, `Reason="not-configured"`, daemon probe NOT called). Rename `TestRunMemHealth_CorruptedSettings` → `TestRunMemHealth_Corrupted` (expect `Reason="corrupted"`). Add `TestRunMemHealth_SupervisorJSONFallback` (only `supervisor.json` present → healthy path reached after probe). Keep existing `TestRunMemHealth_DaemonUnreachable` semantics. All tests must follow the `TestXxx_NotConfigured / _Unreachable / _Corrupted` naming convention | `go test ./go/cmd/harness/... -race -count=1` all green; `grep -n "not-initialized\|corrupted-settings" go/cmd/harness/` returns 0 matches; `go test -cover ./go/cmd/harness/...` coverage on `checkMemHealth` ≥ pre-change baseline | 80.1 | cc:Done [5c464b6] |
| 80.3 | Create `.claude/rules/active-watching-test-policy.md` (English) codifying the tri-state test requirement for any feature that probes an external daemon. Required content: (a) three mandatory test states — `_NotConfigured` (dependency not installed → healthy, no warning), `_Unreachable` (dependency installed but unreachable → unhealthy, warning), `_Corrupted` (dependency installed but malformed state → unhealthy, warning); (b) injection hook requirement (probes must be package-level `var` for test stubbing, mirroring `daemonProbe` in `mem.go`); (c) exit-code contract (`_NotConfigured` must NOT exit non-zero); (d) backlink from `CLAUDE.md` rules section. Translate any Japanese content from upstream | File exists at `.claude/rules/active-watching-test-policy.md`; includes all three required subsections; `CLAUDE.md` rules list links to it; `./tests/validate-plugin.sh` passes; `bash harness/skills/harness-release/scripts/check-residue.sh` returns 0 detections (or new detections explicitly allowlisted per migration-policy.md Rule 5) | 80.1 | cc:Done [6f2b7e8] |
| 80.4 | Residue + allowlist sweep. Run `bash harness/skills/harness-release/scripts/check-residue.sh` after 80.1-80.3 land. If the reason-string rename surfaces new stale-concept hits (e.g. `not-initialized` lingering in docs), fix at source. If `mem.go`/`mem_test.go` legitimately reference deleted concepts, add narrow allowlist entries to `.claude/rules/deleted-concepts.yaml` (file-path scope only — do NOT broaden to `go/` per migration-policy.md Rule 3) | `bash harness/skills/harness-release/scripts/check-residue.sh` returns `Detections: 0` on current HEAD; any new allowlist entries scoped to specific file paths; retroactive validation per migration-policy.md Rule 4 (checkout pre-80.1 commit, confirm scanner detects 1+ expected residues) | 80.1, 80.2, 80.3 | cc:Done — check-residue.sh: 0 detections; grep confirms 0 stale references to not-initialized/corrupted-settings; no allowlist entries needed |
| 80.5 | Release v4.11.2 (patch per `.claude/rules/versioning.md` — bug fix, no new user capability). CHANGELOG `[Unreleased]` Before/After entry under "Fixed": Before = "`bin/harness mem health` exits 1 for users without harness-mem installed, producing spurious `unhealthy: not-initialized` output." After = "Returns `healthy: true, reason: not-configured` + exit 0 when harness-mem is not installed; `unhealthy` is reserved for corrupted state or unreachable daemon." Bump `harness/VERSION` and `.claude-plugin/marketplace.json` via `./harness/skills/harness-release/scripts/sync-version.sh bump`. Tag + release per `harness-release` skill | `harness/VERSION` == `.claude-plugin/marketplace.json` version == `4.11.2`; CHANGELOG has dated `[4.11.2]` section in Before/After format; `bash harness/skills/harness-release/scripts/check-consistency.sh` passes; `./tests/validate-plugin.sh` passes; GitHub release created | 80.1, 80.2, 80.3, 80.4 | cc:Done |

**Port delta notes** (what we are NOT porting from the upstream diff):
- Upstream `hooks.json` wiring for `memory-session-start.sh` / `userprompt-inject-policy.sh` — Phase 77.5 already verified our Go `UserPromptInjectPolicyHandler` emits `additionalContext` via `consumeResumeContext()`; wiring the shell script would double-inject.
- Upstream `config.yaml` thresholds (`plans_drift`, `advisor_ttl_seconds`) — Phase 77.1 / 77.2 already present at `harness/.claude-code-harness.config.yaml:36,45`.
- Upstream Plans.md phase tracking and release CHANGELOG sections — project-specific, out of scope for our fork.

---

## Phase 79: Explicit Failure Taxonomy (borrowed from NL-Agent-Harnesses paper)

Created: 2026-04-20

**Goal**: Consolidate scattered failure-handling logic into a single SSOT catalog, inspired by the "Failure Taxonomy" element (one of 8 NLAH elements) in *Natural-Language Agent Harnesses* ([arXiv:2603.25723v1](https://arxiv.org/abs/2603.25723)). Failure handling is currently spread across Go tampering patterns T01–T12 (`go/internal/hookhandler/post_tool.go`), Advisor error signatures (`harness/agents/advisor.md`), the Worker 3-retry-then-reticket loop (`harness-work` Phase 13), and CI fixer rules. Stable IDs (`FT-*`) give agents a shared vocabulary for detection, recovery, and escalation — and compose naturally with Phase 72 traces (trace events can cite taxonomy IDs for richer post-hoc audit).

**Paper prescription**: "Named failure modes drive recovery strategies." Turning tacit practice into an inspectable, referenceable artifact improves auditability and enables future ablation study (paper RQ2).

**Depends on**: Phase 72 (trace events in `.claude/state/traces/<task_id>.jsonl` are the richest emission point for `FT-*` IDs; 79.2 wires them in).

**Non-goals**: Replacing advisor history duplicate suppression, tampering detection logic, or CI fixer rules — only naming and cross-referencing them.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 79.1 | Create `.claude/rules/failure-taxonomy.md` with columns `ID \| category \| mode \| detector \| recovery \| escalation \| source`. Inventory all modes across four source systems: Go tampering (T01–T12), Advisor error signatures, Worker retry patterns, CI fixer rules. Assign stable IDs `FT-<CATEGORY>-<NN>` (e.g. `FT-TAMPER-01`, `FT-RETRY-01`, `FT-ADVISE-01`). Document ID stability rule: never reuse an ID on removal | File exists; ≥15 modes catalogued covering all four sources; every row has non-empty detector + recovery; ID stability rule stated | - | cc:done [eedd811] |
| 79.2 | Annotate Go tampering patterns in `go/internal/hookhandler/post_tool.go` (+ tests) to emit `FT-TAMPER-*` IDs in hook output and error messages. Include `taxonomy_id` in the `error` event payload emitted to Phase 72 traces | `grep -rn "FT-TAMPER" go/` matches all T01–T12 patterns; `go test ./go/internal/hookhandler/...` passes; sample trace event contains `taxonomy_id` field | 79.1, 72.2 | cc:done [a80feb7] |
| 79.3 | Update `harness/agents/advisor.md` and `harness/agents/worker.md` to reference taxonomy IDs in error signatures. Extend advisor history schema: add optional `taxonomy_id` field to `.claude/state/advisor/history.jsonl` records (backward compat — old records without the field still match duplicate-suppression) | Both agent files cite `.claude/rules/failure-taxonomy.md`; advisor history schema documents `taxonomy_id`; duplicate-suppression logic handles mixed old/new records | 79.1 | cc:done [6b77ee1] |
| 79.4 | CHANGELOG `[Unreleased]` Before/After entry under "Added" per `.claude/rules/changelog.md`; link taxonomy from CLAUDE.md rules index | CHANGELOG entry in Before/After format; CLAUDE.md rules section links to `failure-taxonomy.md`; `./tests/validate-plugin.sh` passes | 79.1 | cc:done [6b77ee1] |

---

## Phase 78: Plans.md ordering convention — newest-first + archive footer

Created: 2026-04-20

**Goal**: Make the implicit "newest phase on top" convention explicit and enforced, and add a persistent archive navigation footer at the bottom of Plans.md. Currently: insertion order is undocumented in `harness-plan/SKILL.md`, no rule file states the convention, and a reader who scrolls to the bottom of Plans.md has no link to older phases in `.claude/memory/archive/`.

**Scope**:
1. Update `harness-plan/SKILL.md` to make insertion point and archive footer explicit
2. Fix current Plans.md phase ordering (74–77 are ascending; reorder to 77→76→75→74)
3. Add a persistent `## Archive` footer to Plans.md linking to `.claude/memory/archive/`
4. Extend `plans-format-check.sh` to validate non-ascending phase-number order

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 78.1 | Update `harness/skills/harness-plan/SKILL.md` `add` subcommand section: state insertion point explicitly ("Insert new phase block immediately after the `---` header separator, above any existing phase"). Update `archive` subcommand section: state it must maintain the `## Archive` footer after archiving | SKILL.md `add` section contains "insert above existing phases"; `archive` section mentions footer maintenance; `./tests/validate-plugin.sh` passes | - | cc:done [ca29894] |
| 78.2 | Fix current Plans.md: reorder phases 74–77 into non-ascending order (77 → 76 → 75 → 74) so all phases read newest-first. Add `## Archive` footer section at bottom of Plans.md with links to all existing archive files in `.claude/memory/archive/` | `grep -n "^## Phase" Plans.md` numbers are non-ascending top-to-bottom (gaps allowed); `## Archive` section exists at end of file with ≥1 link | - | cc:done [ca29894] |
| 78.3 | Extend `harness/scripts/plans-format-check.sh` with a phase-order check: extract all `## Phase N` numbers top-to-bottom, assert each number is strictly less than the previous one (non-ascending, gaps allowed), exit non-zero with "phase order violation: Phase M appears after Phase N" if not | `bash harness/scripts/plans-format-check.sh Plans.md` passes on correct file (e.g. 78,77,76); fails on ascending file (e.g. 74,75,76); gaps (78,76,74) pass | - | cc:done [ca29894] |
| 78.4 | CHANGELOG `[Unreleased]` Before/After entry under "Added"; update `archive` subcommand in `harness-plan/references/create.md` if it contains a Plans.md template that needs the footer | CHANGELOG entry present in Before/After format; no Plans.md template omits the `## Archive` footer | - | cc:done [ca29894] |

---

## Phase 77: Port upstream Session Monitor + memory-hooks wiring (PR #92, #93)

Created: 2026-04-19

**Source**: Chachamaru127/claude-code-harness PRs #91 (post-v4.3.0 cleanup, skip), #92 (v4.3.0 Phase 48 Session Monitor + Phase 49 XR-003 hooks), #93 (v4.3.1 release cut — **includes critical CodeRabbit security fixes** missing from #92).

**Goal**: Bring in the portable *runtime observability* and *memory-hook wiring* improvements from the upstream fork. Similar workflow to Phase 75 (v4.9.5 is diverged — our architecture is already Go-native, memory-bridge and inject-policy are Go handlers, and our Plans.md / hooks layout differs), so we cherry-pick by concern rather than rebase. Every port needs to be re-grounded in *our* file layout (`harness/hooks/hooks.json`, `harness/.claude-code-harness.config.yaml`) and *our* existing monitor.go, which is structurally different from upstream's.

**Investigation summary**:
- **PR #91** (1 commit): CLAUDE.md stale-ref cleanup + Plans.md archiving specific to upstream's v4.3.0 cut. Our CLAUDE.md is already Go-native-aware (v4.9.5) and we archive on our own cadence. **Skip entirely.**
- **PR #92** (6 commits, ~950 net lines): Phase 48 = three active `⚠️` drift monitors (harness-mem health, advisor/reviewer TTL, Plans.md thresholds) added into their `go/internal/session/monitor.go` + new `bin/harness mem health` subcommand. Phase 49 = wire `memory-session-start.sh` + `userprompt-inject-policy.sh` into hooks.json because upstream's Go `inject-policy` stub didn't return additionalContext. Mostly portable, but needs remapping into our codebase.
- **PR #93** (3 extra commits beyond #92): `2c60972` is just a rebase of the same Phase 49 commit. `14a18dc` is a release cut (skip — we manage our own VERSION). **`362e950` is load-bearing**: two CodeRabbit critical fixes — daemon TCP probe for `mem health` (file-only check gave false positives) and `os.Executable()` resolution replacing `projectRoot/bin/harness` (security: guardrail-bypass risk). Must bundle with any `mem health` port.

**What we already have and are NOT porting**:
- Go-native `hook memory-bridge --mode=user-prompt` (harness/hooks/hooks.json:283) — our memory-bridge is a real handler, not a stub, so the wiring rationale for `userprompt-inject-policy.sh` needs re-verification before we wire it (77.5 preflight).
- Release tooling, version bumps, CHANGELOG release sections — we run our own versioning pipeline (`sync-version.sh`).
- Dual-file hooks sync (upstream ships both `.claude-plugin/hooks.json` and `hooks/hooks.json`). We have a single `harness/hooks/hooks.json`, so the "single-hooks-file defective change" scenario #93 fixes does not apply.

**Portable scope mapped to our layout**:

| Upstream concern | Upstream file | Our target file | Portable? |
|------------------|---------------|-----------------|-----------|
| Plans.md drift thresholds | `go/internal/session/monitor.go` (+~80 lines) | `go/internal/session/monitor.go` (new function alongside existing SessionStart logic) | **Yes — highest ROI** |
| Advisor/Reviewer drift TTL | same | same; reads `.claude/state/session.events.jsonl` | **Yes — valuable, we already write events.jsonl** |
| Config schema additions | `.claude-code-harness.config.yaml` | `harness/.claude-code-harness.config.yaml` | Yes — add `orchestration.advisor_ttl_seconds` and `monitor.plans_drift` sections |
| `bin/harness mem health` | `go/cmd/harness/mem.go` + `main.go` | same (new) | **Optional** — harness-mem is not installed locally; still portable as opt-in |
| Daemon TCP probe + secure binary resolve (CodeRabbit) | `go/cmd/harness/mem.go` | same | Mandatory **iff** we do mem health |
| Shell-script hook wiring (XR-003) | `.claude-plugin/hooks.json` | `harness/hooks/hooks.json` | Conditional — pre-verify our Go `inject-policy` actually emits `additionalContext` before wiring the shell script (risk: duplicate injection) |
| `test-memory-hook-wiring.sh` presence + order assertions | `tests/test-memory-hook-wiring.sh` | ours already exists (60+ lines) — extend with the new assertions **iff** we do 77.5 | Conditional |

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 77.1 | Port Plans.md drift monitor (upstream PR #92 task 48.1.3). Add a `CheckPlansDrift` function to `go/internal/session/monitor.go` that reads Plans.md, counts `cc:WIP` tasks, and computes `stale_for` from the oldest WIP task's git-blame timestamp or file mtime fallback. Emit `⚠️ plans drift: WIP={n}, stale_for={hours}h` via the existing `writeSummary` sink when `WIP >= wip_threshold` OR `stale_for >= stale_hours`. Defaults: `wip_threshold=5`, `stale_hours=24`. Config via `harness/.claude-code-harness.config.yaml` `monitor.plans_drift.{wip_threshold,stale_hours}`. Include `now func() time.Time` injection for tests. Dead-code-free single-return style (learn from upstream 48.2.1 fix) | `go test ./internal/session/... -run TestMonitorHandler_PlansDrift -race` passes with ≥3 cases (under-threshold miss, WIP hit, stale hit); `go vet ./...` clean; `harness/.claude-code-harness.config.yaml` has `monitor.plans_drift` section with defaults; `validate-plugin.sh` still 41/0 | - | cc:Done [6c0b051] |
| 77.2 | Port Advisor/Reviewer drift monitor (upstream PR #92 task 48.1.2). Add `CheckAdvisorDrift` to `monitor.go`: scan the last 200 lines of `.claude/state/session.events.jsonl` (NOT the whole file — ring buffer style; upstream deferred the optimization, we adopt it from the start) for `advisor-request.v1` events without a matching `advisor-response.v1` within `orchestration.advisor_ttl_seconds` (default 600). Same pattern for `review-result.v1`. Emit `⚠️ advisor drift: request_id={id}, waiting {elapsed}s`. `filepath.Clean` the config path (also learned from 48.2.1). Respect missing jsonl gracefully | Tests cover hit / miss / config-override / file-missing (4 cases per event type = 8 total); `go vet ./...` clean; drift output appears exactly once per request (no duplicate on repeated runs — guard via `processed_request_ids` in-memory set scoped to the single `Handle` call) | 77.1 (shared scaffolding) | cc:Done [1173d25] |
| 77.3 | Add config schema entries and doc. Update `harness/.claude-code-harness.config.yaml` with `orchestration.advisor_ttl_seconds: 600` and `monitor.plans_drift: {wip_threshold: 5, stale_hours: 24}`. Add a short section to `docs/` or the relevant skill reference explaining how to tune the thresholds. Add CHANGELOG `[Unreleased]` entry in Before/After format (Before: drift invisible until user notices stale WIP / orphan advisor; After: session-monitor emits `⚠️` on each SessionStart). No code changes outside yaml + markdown | `harness/.claude-code-harness.config.yaml` has both new sections with comments; CHANGELOG entry exists; `check-consistency.sh` passes | 77.1, 77.2 | cc:Done [374e5c8] |
| 77.4 | (Optional, medium priority) Port `bin/harness mem health` subcommand (upstream PR #92 task 48.1.1 **bundled with PR #93 CodeRabbit fixes from commit 362e950**). Create `go/cmd/harness/mem.go` with `runMemHealthCheck()`: stage-1 file integrity (`~/.claude-mem/` + settings.json), stage-2 TCP probe to `HARNESS_MEM_HOST:HARNESS_MEM_PORT` (default `127.0.0.1:37888`, 500ms timeout). Package-level `daemonProbe` var for test injection. Binary resolution via `resolveHarnessBinary()`: `os.Executable()` → `CLAUDE_PLUGIN_ROOT/bin/harness` → `exec.LookPath("harness")` — **never `projectRoot/bin/harness`** (security: guardrail bypass). Wire into `go/cmd/harness/main.go` dispatcher. Add allowlist entries to `.claude/rules/deleted-concepts.yaml` for `mem.go`, `mem_test.go` — but keep `bin/` prefix narrow (upstream's initial entry was flagged too broad, deferred to 4.3.2). Unhealthy → exit 1 + `⚠️ harness-mem unhealthy: {reason}`; healthy → JSON `{healthy:true}` to stdout | `go test ./cmd/harness/... -run TestRunMemHealth -race` covers: not-initialized (early exit), corrupted settings, daemon-unreachable (TCP fail injected), healthy; `go vet ./...` clean; manual smoke `bin/harness mem health` on a machine without harness-mem returns `{healthy:false,reason:"not-initialized"}` exit 1; `check-residue.sh` still 0 detections | - | cc:Done [3f3bf93] |
| 77.5 | (Optional, conditional on preflight) Wire `memory-session-start.sh` and `userprompt-inject-policy.sh` into `harness/hooks/hooks.json` (upstream PR #92 Phase 49 / XR-003). **Preflight first**: run `bash harness/scripts/userprompt-inject-policy.sh < fixture.json` against a live session with memory-resume-context.md present and confirm our Go `hook inject-policy` does not already emit the same `additionalContext` (risk: double-injection). If our Go handler is a stub like upstream's was, add `bash "${CLAUDE_PLUGIN_ROOT}/harness/scripts/hook-handlers/memory-session-start.sh"` to the `startup\|resume` SessionStart hooks array (timeout 30, once=true) and add `bash "${CLAUDE_PLUGIN_ROOT}/harness/scripts/userprompt-inject-policy.sh"` to the UserPromptSubmit array between `memory-bridge` and `inject-policy` (timeout 15). If our Go handler already injects, document the finding and skip the wiring (task becomes a no-op confirmation) | Preflight output captured in a commit-attached note OR in `docs/`; either (a) hooks.json has the two new entries and `tests/test-memory-hook-wiring.sh` extended with presence + order assertions for the shell scripts, or (b) written confirmation that our Go stack already injects and the shell scripts remain bundled-but-dormant by design; `test-memory-hook-wiring.sh` passes in both branches | 77.4 (shares mem-daemon context; not a hard dep) | cc:Done — preflight: our Go `UserPromptInjectPolicyHandler` already emits `additionalContext` via `consumeResumeContext()`; wiring shell script would double-inject — scripts stay bundled-dormant by design |
| 77.6 | Bundle release: once 77.1–77.3 land (and whichever of 77.4/77.5 we opted into), bump `harness/VERSION` to the next minor (`4.10.0`) — Session Monitor active watching is a new capability (user-facing observability = minor per `.claude/rules/versioning.md`). Sync `.claude-plugin/marketplace.json` via `./harness/skills/harness-release/scripts/sync-version.sh bump`. CHANGELOG `[Unreleased]` → `[4.10.0] - YYYY-MM-DD` in Before/After format. Tag and release per `harness-release` skill | `harness/VERSION` == `.claude-plugin/marketplace.json` version; CHANGELOG has dated 4.10.0 section; `check-consistency.sh` passes; release tag created | 77.1, 77.2, 77.3 (77.4 / 77.5 conditional) | cc:Done |

**Explicit non-goals** (so future sessions don't re-litigate):
- Do not rebase `master` onto `pr-upstream-93` — the branches diverged at the Go-native split; a rebase would re-introduce upstream-specific CLAUDE.md and Plans.md churn we already moved past.
- Do not port upstream's release CHANGELOG text verbatim — our Before/After format in `.claude/rules/github-release.md` differs.
- Do not adopt upstream's dual-hooks-file scheme (the one #93 patches into existence); single `harness/hooks/hooks.json` is our SSOT.
- Do not port `deleted-concepts.yaml` entries 1:1 — upstream's initial `bin/` prefix was flagged too broad even by their own CodeRabbit reviewer; only add the specific file paths we actually create (mem.go / mem_test.go) if 77.4 lands.

---

## Phase 76: Fix WorktreeCreate hook — JSON-cwd guard parity in Go handler

Created: 2026-04-19

**Problem**: Users running `/harness-work` / `/harness-loop` sometimes end up with a literal folder named `{"decision":"approve","reason":"WorktreeCreate: initialized worktree state"}` in their project. Root cause: Claude Code occasionally feeds the previous `WorktreeCreate` hook's JSON output back as the `cwd` field on a subsequent invocation. The shell handler (`harness/scripts/hook-handlers/worktree-create.sh:58-63`) has a guard that rejects cwd values starting with `{`, but the Go port at `go/internal/hookhandler/worktree_create.go` lacks it — so `os.MkdirAll` happily creates a directory named after the JSON payload.

**Why it matters**: Silent filesystem pollution in user projects. No error, no warning, just a bizarrely-named folder that reappears on every worktree creation until someone notices.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 76.1 | Port the shell handler's JSON-cwd guard to `go/internal/hookhandler/worktree_create.go`. After the empty-cwd check, reject cwd values whose first byte is `{` and return the approve response with reason `"WorktreeCreate: skipped (invalid JSON cwd)"`. Matches shell handler exactly | `worktree_create.go` has the guard immediately after the `input.CWD == ""` check; running with a JSON-string cwd returns the skip reason and does not create a directory | - | cc:Done [local] |
| 76.2 | Add regression test `TestHandleWorktreeCreate_JSONCWDGuard` to `go/internal/hookhandler/worktree_create_test.go`. Feeds a payload whose `cwd` is the actual upstream hook output JSON; asserts the skip reason AND that no directory with that JSON name exists on disk | `go test ./go/internal/hookhandler/... -run TestHandleWorktreeCreate -race` passes all 7 tests including the new one | 76.1 | cc:Done [local] |
| 76.3 | Commit fix + test with Conventional Commits message (`fix(hooks): guard against JSON cwd in WorktreeCreate Go handler`). Record Before/After entry under CHANGELOG `[Unreleased]` | Commit landed on master; CHANGELOG has a Before/After entry describing the symptom (JSON-named folders) and the fix (Go handler now matches shell parity) | 76.2 | cc:Done [d87381f] |

---

## Phase 75: Port upstream Worker contract improvements (PR #88, #89)

Created: 2026-04-19

**Source**: Chachamaru127/claude-code-harness PRs #88 (v4.2.0, Hokage) / #89 (v4.3.0, Arcana) / #90 (release marker, no code).

**Goal**: Bring in the portable behavioral improvements from the upstream fork without rebasing — their v4.2/v4.3 diverged significantly from our line (we're at v4.9.3 with a different Go-native architecture). We cherry-pick what's a clean, low-coupling improvement and skip what our infra already covers differently.

**Investigation summary**:
- **PR #88** is a CC 2.1.99-110 + Opus 4.7 integration release. Most of it is CC-version-specific (guardrail rule refresh, agent frontmatter migration, plugin schema migration to `plugins-reference`) — we're already past it structurally. Two portable items: `scripts/enable-1h-cache.sh` (opt-in 1-hour prompt cache) and a role-based PreCompact decision (our `pre-compact-save.js` only warns; upstream's `pre_compact.go` blocks compaction for Worker sessions).
- **PR #89** is the high-value one: four independent Worker/Reviewer contract tightenings (#84–#87). All directly applicable.
- **PR #90** is a release-completion marker with no code changes — skip.

**What we already have and are NOT porting**:
- PreCompact hook: we have `pre-compact-save.js` warning on `cc:WIP` tasks. Upstream's role-based blocking is a nice-to-have but not urgent — deferred to 75.6 as optional.
- Guardrail engine: our Go-native `bin/harness` already covers bypass detection.
- Plugin schema: our v4.9 architecture is already past upstream's `plugins-reference` migration.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 75.1 | Add Worker NG rules to `harness-work` SKILL.md (upstream issue #85). NG-1: Plans.md `cc:*` markers are Lead-owned — Workers must not modify. NG-2: embedded git repositories inside the worktree are rejected. NG-3: nested teammate spawning is forbidden (reinforce the platform-level block with an explicit contract clause) | `harness-work/SKILL.md` has a "Worker NG Rules" section listing all three; each rule has a one-line rationale and a check point (manual or scripted). Optional: pre-delegate preflight script | - | cc:Done [6030060] |
| 75.2 | Add `REVIEW_AUTOSTART` handshake to `harness-review` SKILL.md Step 0 (upstream issue #84). Purpose: fix the fork-mode stall where `/harness-review` with no args waits forever for input. First 3 lines of Step 0 emit/detect `REVIEW_AUTOSTART` marker; add a prohibition list for 5 fork-context failure modes | `harness-review/SKILL.md` Step 0 top-matter includes the handshake marker and the prohibition list; a smoke test (`bash` or manual) confirms no stall when invoked in fork context with no args | - | cc:Done [390a473] |
| 75.3 | Add Worker self-review gate (`worker-report.v1` JSON schema) to `harness-work` SKILL.md (upstream issue #86). Before the Lead spawns a Reviewer, the Worker must emit a `worker-report.v1` JSON with 5 self-review rule verdicts + evidence fields. Lead rejects incomplete submissions; up to 2 amendment cycles before escalation | `worker-report.v1` schema documented in `harness-work/SKILL.md` or `references/`; Lead flow updated to validate and either accept or reject with amendment loop; schema has `rule_id`, `verified: bool`, `evidence: string` per entry | 75.1 | cc:Done [37b87cf] |
| 75.4 | Add universal violations session injection to `breezing` mode (upstream issue #87). Reviewer memory updates gain a `scope: universal \| task-specific` field; universal-scope items are collected and prepended to the next Worker's briefing within the same breezing session. Prevents repeat violations across tasks | Reviewer memory schema has `scope` field; breezing Phase B prepends universal items to Worker spawn prompt; documented in `harness-work/SKILL.md` breezing section | 75.3 | cc:Done [9d08753] |
| 75.5 | Port `scripts/enable-1h-cache.sh` (upstream PR #88). Opt-in shell script that exports Anthropic's 1-hour prompt cache setting for sessions expected to exceed 30 minutes. Uses `export` format so subprocesses inherit the setting | `harness/scripts/enable-1h-cache.sh` exists and is executable; `source`-able or `eval`-able from a parent shell; a short usage note in `harness-work/SKILL.md` or `docs/` explains when to use it | - | cc:Done [ae36e86] |
| 75.6 | (Optional, lower priority) Upgrade `pre-compact-save.js` from warn-only to role-based blocking (upstream PR #88 `pre_compact.go`). Worker sessions with `cc:WIP` tasks in Plans.md block compaction; Reviewer / Advisor roles pass through. Requires session-role detection in the hook | `pre-compact-save.js` detects session role from session-state and blocks compaction for Worker+WIP combo; test covers all 4 role × state permutations | - | cc:Done [84b0429] |

---

## Phase 74: Code-space skill search — proof-of-concept on `harness-review`

Created: 2026-04-17

**Goal**: Borrow Meta-Harness's code-space search idea (arxiv 2603.28052) — instead of hand-editing `SKILL.md`, generate variants and score them against an evaluation suite. Scope: one skill (`harness-review`), 3-5 variants, measurable outcome. If the POC improves the skill, formalize the loop; if not, record why and discard.

**Depends on**: Phase 72 (traces provide failure signal for the proposer) + Phase 73 (advisor pattern reusable for proposer scaffolding).

**Score function** (decided 2026-04-18): **golden verdict / pass-fail** — each test case has an expected APPROVE or REQUEST_CHANGES; score = correct / total. Binary, automatable, CI-friendly.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 74.1 | Define score function + build eval runner `local-scripts/eval-skill.sh <skill-dir> <eval-suite-dir>`. Runs fixed inputs through the skill, captures outputs, emits JSON score report | `eval-skill.sh harness/skills/harness-review tests/skill-eval/harness-review` prints a reproducible score | 72.1 | cc:Done [8ed646c] |
| 74.2 | Build evaluation suite for `harness-review` at `tests/skill-eval/harness-review/` — 5 PR diffs (2 with real bugs, 2 clean, 1 scope-creep) + expected verdicts | `ls tests/skill-eval/harness-review/*.{diff,expected.json}` shows 10 files; running baseline skill against suite produces a score | 74.1 | cc:Done [cbbe5af] |
| 74.3 | Write proposer script `local-scripts/propose-skill-variants.sh <skill-dir>`. Given SKILL.md + eval output + recent traces, generates 3 SKILL.md variants to `/tmp/skill-variants/harness-review-v{1,2,3}/SKILL.md` via Claude subagent | Running against a broken baseline (intentionally degraded `harness-review`) produces 3 syntactically valid SKILL.md files with meaningful diffs | 74.1 | cc:Done [9ef0a1e] |
| 74.4 | Run end-to-end search loop: baseline → generate 3 variants → score each → pick winner. Emit report at `.claude/state/code-search/harness-review-<YYYY-MM-DD>.md` with scores, diffs, chosen winner | Report exists; winner's score ≥ baseline's score; report includes rationale | 74.2, 74.3 | cc:Done [local] |
| 74.5 | Decision gate: if winner beats baseline by ≥10%, promote to main and add pattern to `patterns.md`. Otherwise document the null result. Update CHANGELOG [Unreleased] | Either a commit promoting the variant OR a patterns.md entry "code-space search POC attempted; no gain — reasons X, Y" | 74.4 | cc:Done [747f999] |

---

## Future Considerations

(none currently)

---

## Archive

- Last archive: 2026-04-18 (Phase 62–73 → `.claude/memory/archive/Plans-2026-04-18-phase62-73.md`)
- Other older phases have been moved to `.claude/memory/archive/` to keep this file lean.
