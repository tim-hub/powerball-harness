# Powerball Harness — Plans.md

Last release: v4.11.2 on 2026-04-20 (mem health tri-state fix + active-watching test policy)

---

## Phase 82: Rename memory → harness-remember, create remember-this, fix release-this stale refs

Created: 2026-04-21

**Goal**: (A) Rename `harness/skills/memory/` to `harness/skills/harness-remember/` to avoid collision with Claude Code's built-in `/memory` command. (B) Extract the plugin-specific `sync-project-specs.md` from harness-remember into a new `.claude/skills/remember-this/` project-level skill. (C) Fix stale phase references in `.claude/skills/release-this/SKILL.md` Step 6 that still list old Phases 4-5 (marketplace.json, codex symlinks) removed in Phase 81.

**Motivation**:
- The `memory` skill name collides with Claude Code's built-in memory system — the auto-loader may route to the wrong one.
- `sync-project-specs.md` references `powerball-harness` operations — plugin-specific, same pattern as the harness-release split in Phase 81.
- release-this was created before the harness-release refactor landed, so its Step 6 description is stale.

**Agent names**: NOT changing. `claude-code-harness:worker`, `claude-code-harness:reviewer`, `powerball-harness:advisor` are part of the distributed plugin and work correctly as-is.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 82.1 | **Rename `harness/skills/memory/` → `harness/skills/harness-remember/`**: `mv` the directory. Update the `name:` field in SKILL.md frontmatter from `memory` to `harness-remember`. Update all external references: CLAUDE.md skills table, docs/CLAUDE-skill-catalog.md, harness/README.md, scripts referencing `skills/memory`, session-memory SKILL.md, workflows/default/init.yaml, and any other files found by `grep -r "skills/memory" . --include="*.{md,sh,json,yaml}"`. Preserve all subcommands (ssot, sync, sync-across, migrate, merge, search, record) | Directory at new path; old path gone; `name: harness-remember` in frontmatter; `grep -r "skills/memory" . --include="*.md" --include="*.sh" --include="*.json" --include="*.yaml"` returns 0 matches (excluding CHANGELOG/archive); skill auto-loads under new name | - | cc:Done [8068374] |
| 82.2 | **Create `.claude/skills/remember-this/SKILL.md`**: New project-level skill for this repo's memory operations. Move `sync-project-specs.md` from `harness/skills/harness-remember/references/` to `.claude/skills/remember-this/references/`. The SKILL.md should document the `sync-across` subcommand with references to this repo's specific workflows (powerball-harness operations, Plans.md marker conventions, AGENTS.md PM/Impl roles). Frontmatter follows skill-description.md rules | Skill exists at `.claude/skills/remember-this/SKILL.md`; `sync-project-specs.md` at `.claude/skills/remember-this/references/`; old copy removed from harness-remember; frontmatter ≤300 chars; description contains "Use when" | 82.1 | cc:Done [2afc534] |
| 82.3 | **Update harness-remember to drop sync-across**: Remove or update the `sync-across` row in harness-remember's Quick Reference table — it now points to `remember-this` instead. Add a note: "For project-specific spec sync, see `.claude/skills/remember-this/`". Keep all other subcommands (ssot, sync, migrate, merge, search, record) | harness-remember SKILL.md no longer references `sync-project-specs.md`; Quick Reference table updated; all other subcommands intact | 82.2 | cc:WIP |
| 82.4 | **Fix release-this SKILL.md Step 6 description**: Step 6 currently lists harness-release phases that no longer exist (Phase 4: marketplace.json sync, Phase 5: codex symlinks). Update to match actual current harness-release phases: Phase 0 (preflight), Phase 1-2 (version), Phase 3 (CHANGELOG), Phase 4 (commit & tag), Phase 5 (push), Phase 6 (GitHub Release). Also remove the "second pass" codex symlink note — codex symlinks are already checked in release-this Step 4 | Step 6 accurately reflects current harness-release phases 0-6; no mention of marketplace.json or codex symlinks in Step 6; `grep "marketplace.json\|Codex symlink" .claude/skills/release-this/SKILL.md` returns 0 matches in Step 6 section | - | cc:Done [a591ff1] |
| 82.5 | **Update documentation**: CLAUDE.md skills table (`memory` → `harness-remember`, add `remember-this`). docs/CLAUDE-skill-catalog.md entry update. CHANGELOG.md [Unreleased] Before/After entry. Update any doc that says "use `/memory`" → "use `/harness-remember`" | CLAUDE.md skills table lists `harness-remember` (not `memory`); CHANGELOG [Unreleased] has Before/After; docs/CLAUDE-skill-catalog.md updated; no docs reference `/memory` as the way to access harness SSOT | 82.1, 82.2, 82.3, 82.4 | cc:TODO |
| 82.6 | **Validation**: `make test`, `make check`, `./tests/validate-plugin.sh`. Verify skill auto-loads with new name. Verify old name no longer appears in skill listings (except CHANGELOG/archive) | `make test` passes; `make check` passes; `validate-plugin.sh` passes; `grep -r 'name: memory' harness/skills/` returns 0 matches | 82.5 | cc:TODO |

---

## Phase 81: Split harness-release into generic core + project-specific release-this

Created: 2026-04-21

**Goal**: Make `harness-release` a generic release engine usable by ANY project that installs the Harness plugin, not just this plugin itself. Extract all plugin-specific checks (check-consistency.sh, codex symlink verification, marketplace.json sync, completion marking) into a new `.claude/skills/release-this/` skill that runs project-specific validation first, then delegates to the generic harness-release for the actual release flow. Add `make build-all` as an early validation step in release-this.

**Motivation**: Currently harness-release is tightly coupled to the claude-code-harness plugin. Phases 4-5, 9, and `check-consistency.sh` are plugin-specific, which means any user installing Harness gets our plugin-specific checks injected into their release flow. Separating concerns lets harness-release be a clean, reusable release engine while release-this preserves every existing check this plugin needs.

**Design**:
- `harness/skills/harness-release/` → generic release engine (version bump, changelog, commit, tag, push, GitHub Release)
- `.claude/skills/release-this/` → project-specific orchestrator (build-all, check-consistency, codex symlinks, validate-plugin, then invoke harness-release)
- All existing functionality preserved — release-this does everything current harness-release does when releasing this plugin

**Non-goals**: Rewriting scripts from scratch (move, don't rewrite); changing the 9-phase release flow (it stays in harness-release, just without plugin-specific phases); changing version file locations.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 81.1 | **Audit & catalog**: Read every line of harness-release SKILL.md and all 5 scripts. Classify each section/script/phase as GENERIC or PLUGIN-SPECIFIC. Produce a migration manifest listing what moves where | Migration manifest created (can be in-memory or scratch file); every Phase (0-9), every script, every check within scripts categorized; reviewed for completeness | - | cc:Done [f7a08be] |
| 81.2 | **Refactor harness-release SKILL.md**: Remove plugin-specific Phases 4 (sync marketplace.json), 5 (verify codex symlinks), and 9 (completion marking commit). Remove references to check-consistency.sh. Add a "Project-specific pre-release hook" pattern: document that projects can create their own pre-release skill that runs before harness-release. Keep Phases 0-3, 6-8 as the generic flow. Update the Quick Reference table | harness-release SKILL.md contains only generic release logic; no references to marketplace.json, codex symlinks, or check-consistency.sh; "Project-specific pre-release" pattern documented; existing generic phases (preflight, version, changelog, commit, tag, push, GitHub Release) all preserved | 81.1 | cc:Done [1d55f1a] |
| 81.3 | **Refactor release-preflight.sh**: Review all checks in the script. Remove or make optional any checks that assume plugin-specific file structure. Keep: git cleanliness, CHANGELOG [Unreleased], .env parity, debug remnants scan, CI status. Ensure env var overrides (`HARNESS_RELEASE_PLUGIN_ROOT`, `HARNESS_RELEASE_HEALTHCHECK_CMD`, `HARNESS_RELEASE_CI_STATUS_CMD`) still work for generic use. Sprint contract schema validation should be conditional (skip if no contracts directory exists) | `release-preflight.sh` runs clean on a project that is NOT this plugin (no false failures from missing plugin-specific paths); all env var overrides documented; sprint contract check skipped gracefully when `.claude/state/contracts/` absent | 81.1 | cc:Done [32e4a30] |
| 81.4 | **Create `.claude/skills/release-this/SKILL.md`**: New project-specific skill that orchestrates the full release of this plugin. Flow: (1) `make build-all` — cross-platform binary compilation, (2) `check-consistency.sh` — all 13 plugin-specific checks, (3) `validate-plugin.sh` — full plugin validation, (4) codex symlink verification, (5) marketplace.json / harness.toml version sync check, (6) invoke generic `harness-release` for the actual release (version bump, changelog, commit, tag, push, GitHub Release), (7) completion marking commit (current Phase 9 logic). Skill must accept same arguments as harness-release (patch/minor/major/--dry-run/--complete) | Skill file exists at `.claude/skills/release-this/SKILL.md`; frontmatter follows skill-description.md rules; flow covers all 7 steps; accepts patch/minor/major/--dry-run/--complete; dry-run mode runs steps 1-5 without release | - | cc:Done [b151ad2] |
| 81.5 | **Move check-consistency.sh**: Move `harness/skills/harness-release/scripts/check-consistency.sh` to `.claude/skills/release-this/scripts/check-consistency.sh`. Update all references (Makefile `check` target, validate-plugin.sh section references, CLAUDE.md, any other files referencing the old path). Ensure `make check` still works | `check-consistency.sh` exists at new path; old path has no file; `make check` passes; `grep -r "harness-release/scripts/check-consistency" .` returns 0 matches (excluding CHANGELOG/archive); all 13 checks still pass | 81.4 | cc:Done [1df1180] |
| 81.6 | **Update sync-version.sh**: Currently syncs VERSION → harness.toml. The harness.toml sync is plugin-specific. Refactor so the script has a generic mode (just bumps VERSION + updates CHANGELOG compare links) and the harness.toml sync is triggered by release-this or via an env var flag. Keep the script in harness-release (it's mostly generic) | `sync-version.sh bump` works on a project with just a VERSION file (no harness.toml → no error); `sync-version.sh sync` with `HARNESS_RELEASE_EXTRA_VERSION_FILES=harness/harness.toml` syncs harness.toml; both paths tested | 81.2 | cc:Done [1b77f1d] |
| 81.7 | **End-to-end validation**: Run the full release-this flow in dry-run mode (`/release-this --dry-run`). Verify: (1) `make build-all` runs, (2) check-consistency.sh passes, (3) validate-plugin.sh passes, (4) codex symlinks verified, (5) version sync verified, (6) harness-release dry-run completes (preflight, version display, changelog preview). Then run `make test` to verify nothing is broken | Dry-run completes all 7 steps with no errors; `make test` passes; `make check` passes at new path; `make validate` passes | 81.2, 81.3, 81.4, 81.5, 81.6 | cc:Done — make build-all: 3 binaries built; make check: all 13 passes; validate-plugin.sh: 39/39 passed; version sync: 4.11.5 matches; make test: all checks passed |
| 81.8 | **Documentation update**: Update CLAUDE.md skills table to add `release-this` entry. Update `docs/CLAUDE-skill-catalog.md` if it exists. Update CHANGELOG.md [Unreleased] with Before/After entry describing the split. Update any docs that reference `/harness-release` as the way to release this plugin → point to `/release-this` instead | CLAUDE.md skills table lists `release-this`; CHANGELOG [Unreleased] has Before/After entry; no docs tell users to use `/harness-release` for this plugin's releases (they should use `/release-this`) | 81.7 | cc:Done [55ff1cd] |

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

## Future Considerations

(none currently)

---

## Archive

- Last archive: 2026-04-20 (Phase 74–76 → `.claude/memory/archive/Plans-2026-04-20-phase74-76.md`)
- Previous: 2026-04-18 (Phase 62–73 → `.claude/memory/archive/Plans-2026-04-18-phase62-73.md`)
- Other older phases have been moved to `.claude/memory/archive/` to keep this file lean.
