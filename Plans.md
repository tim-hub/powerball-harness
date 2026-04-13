# Claude Code Harness — Plans.md

Last archive: 2026-04-12 (Phase 25–34 → `.claude/memory/archive/Plans-2026-04-12-phase25-34.md`)

---

## Future Considerations

(none currently)

---

## Phase 40: Migration Residue Scanner — inclusion → exclusion verification

Created: 2026-04-11
Purpose: Add an **exclusion-based verification layer** (systematic detection of references to deleted concepts) to Harness's verification stack. Currently only inclusion-based checks ("does X exist?") are in place, relying on accidental discovery for post-major-migration (v3→v4 etc.) residue. Make it so that the 13 v3 residue bugs discovered in this session would be automatically caught in the future.

### Background (Why this phase exists)

Between the v4.0.0 "Hokage" release (2026-04-09) and v4.0.1, 13 v3 residue bugs were **accidentally discovered**:

| # | Residue Bug | How Found | Impact |
|---|---|---|---|
| 1 | `validate-plugin.sh` grep'ing deleted `core/src/guardrails/rules.ts` | 4 failures on first validate run | Verification script producing false negatives |
| 2 | `check-consistency.sh` expecting `"TypeScript guardrail engine"` in README | 2 failures on first consistency run | Same |
| 3 | `tests/test-memory-hook-wiring.sh` expecting v3 shell path exact match | Downstream validate failures | Same |
| 4 | `tests/test-claude-upstream-integration.sh` `permission-denied-handler` | Rediscovered after Worker C partial fix | Same |
| 5 | `"Harness v3"` string in 18 SKILL.md frontmatters | User noticed in slash palette | User confusion |
| 6 | `v3` narrative in `agents/*.md` | Found as side effect of grep re-scan | Same |
| 7 | `(v3)` suffix in SKILL.md H1 titles | Same | Same |
| 8 | `harness.toml` → `plugin.json` sync dropping `skills: ["./"]` | Noticed via auto-revert phenomenon | Functional regression |
| 9 | `/HAR:review` SKILL.md body was English-centric | Fork subagent responding in English | UX problem |
| 10 | `core/ engine` reference in `README.md` file tree | User's final feedback | Misleading |
| 11 | `skills/`/`agents/` duplication bug in `README.md` file tree | Same | Same |
| 12 | Same problem in `README_ja.md` | Same | Same |
| 13 | `Node.js 18+` requirement in `README.md` troubleshooting | Same | Misleading |

All were **accidentally found**: via test failures, user reports, or review feedback. A systematic scanner would have **caught all of these before the v4.0.0 release** — they are a class of bug that systematic detection can handle. This reveals a **fundamental gap** in Harness's verification strategy: it has only inclusion-based checking ("does X exist?") and lacks the exclusion-based perspective ("does deleted X still remain?").

### Priority Matrix

| Priority | Phase | Content | Task Count | Depends |
|--------|-------|------|---------|------|
| **Required** | 40.0 | Foundation (deleted-concepts.yaml + check-residue.sh) | 2 | None |
| **Required** | 40.1 | Integration (doctor --residue + validate-plugin + release preflight) | 3 | 40.0 |
| **Required** | 40.2 | Documentation (migration-policy.md) | 1 | 40.0, 40.1 |

Total: **6 tasks**

### Completion Criteria (Definition of Done — Phase 40 overall)

| # | Criterion | Verification | Required/Recommended |
|---|------|---------|----------|
| 1 | Scanner retroactively detects all 13 v3 residue bugs from this session | Roll back to 1 commit before v4.0.1 and run `bash scripts/check-residue.sh` → 13 detected / post-release commit → 0 | Required |
| 2 | Scanner false positive rate zero (historical records like CHANGELOG.md correctly excluded via allowlist) | Run at v4.0.1 HEAD → 0 hits | Required |
| 3 | `bin/harness doctor --residue` calls scanner and displays results | Run command, see expected output (count + files + line numbers) | Required |
| 4 | Residue check integrated into `validate-plugin.sh`, failures added to total fail count | Intentionally inject v3 residue → validate-plugin fails | Required |
| 5 | `harness-release` skill preflight includes `harness doctor --residue`, aborts release on detection | Documented in SKILL.md + dry-run confirmed | Required |
| 6 | `.claude/rules/migration-policy.md` exists with documented rules for updating deleted-concepts.yaml | File exists + content confirmed | Required |
| 7 | `.claude/rules/deleted-concepts.yaml` has all 13 detected bugs from this session as entries | Parse yaml, minimum 8-10 entries (13 bugs can be consolidated into patterns) | Required |
| 8 | All Go tests pass, validate-plugin 43+/0 (increased by residue check), check-consistency all pass | Run existing test suite | Required |

---

### Phase 40.0: Foundation [P0]

Purpose: Define `.claude/rules/deleted-concepts.yaml` as SSOT and implement `check-residue.sh` to read it and scan the repo

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 40.0.1 | Create `.claude/rules/deleted-concepts.yaml`. Schema has 2 sections: `deleted_paths[]` and `deleted_concepts[]`. Each entry contains `path`/`term`, `term_localized` (optional), `replacement`/`replacement_localized` (optional), `deleted_in` (version), `deleted_by` (commit hash optional), `reason`, `allowlist[]`. Consolidate the 13 v3 residue bugs from this session into patterns as entries: (a) `core/src/guardrails`, (b) `core/dist`, (c) `core/package.json`, (d) `scripts/run-hook.sh`, (e) term `"TypeScript guardrail engine"` / `"TypeScript guardrail engine"`, (f) term `"Harness v3"` / `"(v3)"`, (g) troubleshooting text `"Node.js 18+ is installed"` / `"Node.js 18+ required"` / `"Ensure Node.js"`, (h) v3 shell invocation patterns `"hook-handlers/memory-bridge"` / `"hook-handlers/runtime-reactive"` / `"hook-handlers/permission-denied-handler"`. Allowlist includes `CHANGELOG.md`, `.claude/memory/archive/**`, `benchmarks/**`, `README.md` "Before / After" table regions | YAML is valid (parseable by `yq`). 8-10 entries + `allowlist` array per entry. `reason` field explains why each entry was deleted (v4.0.0 Hokage migration, CC 2.1.94 compliance, etc.) | - | cc:done [20654143, 191cdde4] |
| 40.0.2 | Implement `scripts/check-residue.sh`. Read `.claude/rules/deleted-concepts.yaml` with `yq`, scan `deleted_paths[]` and `deleted_concepts[]` sequentially using `grep -rln -F`. Apply allowlist (gitignore-style matching or prefix match). Exit code 1 on hit, output detailed report to stdout (count, files, line numbers, matching string, which entry it violates). Works under `set -euo pipefail`. Also implement error fallback (if yq not installed → fallback to python3 + yaml module) | (a) Run scanner at v4.0.0 release commit (`8d8ce3c8`) and point before v4.0.1 → detects all 13 v3 residue bugs (retroactive validation), (b) Run at v4.0.1+ HEAD → 0 residue (zero false positives), (c) Intentionally add reference to `core/src/guardrails/rules.ts` in README → detected immediately. `bash scripts/check-residue.sh` standalone works for all behaviors | 40.0.1 | cc:done [20654143, 191cdde4] |

---

### Phase 40.1: Integration [P0]

Purpose: Embed scanner into 3 verification points (developer ad-hoc / per-PR / pre-release)

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 40.1.1 | Add `--residue` flag to `go/cmd/harness/doctor.go`. Internally call `scripts/check-residue.sh` as subprocess and format results for display. Alternatively, implement Go-native yaml parsing (add `gopkg.in/yaml.v3` etc. to go.mod). Works independently from existing `--migration` and other flags. Skip residue check when running `bin/harness doctor` alone (opt-in, for speed) | Running `bin/harness doctor --residue`: (a) displays scanning message, (b) shows file + line number + matching entry when found, (c) displays "✓ No migration residue detected" when 0, (d) exit code 0 (clean) or 1 (residue). `go test ./cmd/harness/ -run TestDoctor_Residue` PASS with intentional residue injection test | 40.0.2 | cc:done [470a05bd] |
| 40.1.2 | Integrate residue scan into `tests/validate-plugin.sh`. Add as new test category "migration residue check", integrate into existing pass/fail/warn counts (0 residue → +1 pass, 1+ residue → +1 fail) | When running `./tests/validate-plugin.sh`, residue check runs as final section. In clean state, pass count +1 (42 → 43). With intentional residue, fail count +1. Existing tests all maintained | 40.0.2 | cc:done [1e886ad9] |
| 40.1.3 | Add `bin/harness doctor --residue` to preflight section (equivalent to Step 1) in `skills/harness-release/SKILL.md`. On residue detection, abort release + display fix instructions to user. Sync 3 mirrors (`skills/`, `codex/.codex/skills/`, `opencode/skills/`) | preflight table in `harness-release` has residue check row added. Error message on residue detection clearly states "Phase 40 scanner detected N references to deleted concepts. Please fix and re-run." Sync with `check-consistency.sh` PASS (mirrors fully match) | 40.0.2 | cc:done [60199c01] |

---

### Phase 40.2: Documentation [P0]

Purpose: Codify rules for correctly operating the scanner in future major migrations

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 40.2.1 | Create `.claude/rules/migration-policy.md`. Content: (1) Obligation to update `.claude/rules/deleted-concepts.yaml` during major version migrations, (2) Update timing must be simultaneous with the deletion PR (no delays), (3) Allowlist operation standards (historical records like CHANGELOG always in allowlist, Before/After tables in allowlist based on context, docs/archive always in allowlist), (4) How to perform retroactive validation, (5) Record the 13 v3 residue bugs found in this session (v4.0.0 → v4.0.1) as an appendix (with the story of why this feature was born), (6) Add a reference to migration-policy.md in `CLAUDE.md` | File exists, markdown valid. A non-expert can understand "why deleted-concepts.yaml is needed" in 5 minutes. 1-line reference added to CLAUDE.md | 40.0.1, 40.1.3 | cc:done [719c08bd] |

---


## Phase 39: Review Experience Improvement + v4.0.1 Pre-Release Polish

Created: 2026-04-11
Purpose: After Phase 38 completion, tidy up improvement opportunities discovered (/HAR:review output non-expert-friendliness, sync.go plugin.json auto-revert root fix, test assertion strictening, v3 cleanup residue, bare review scope cap) and update CHANGELOG before v4.0.1 release to reach a shippable state

Background: An independent review after Phase 38 found 3 follow-ups (HAR:* validation / jq assertion looseness / excessive scope). While fixing these, deeper issues cascaded: namespace inconsistencies, review output UX problems, a root bug where `harness sync` drops the skills field when regenerating plugin.json. All are outside Phase 38 scope but are quality improvements that should be resolved before v4.0.1 release, so they are consolidated here.

### Priority Matrix

| Priority | Phase | Content | Task Count | Depends |
|--------|-------|------|---------|------|
| **Done** | 39.0 | /HAR:review output improvement (bare flow + English + non-expert friendly) | 3 | - |
| **Done** | 39.1 | Infrastructure fixes (sync.go Skills + test assertion + scope cap) | 3 | - |
| **Done** | 39.2 | Name consistency (HAR:* → harness-* revert + SSOT recovery) | 1 | - |
| **Done** | 39.3 | v3 cleanup residue removal + test script v4 migration | 5 | - |
| **Required** | 39.4 | CHANGELOG [Unreleased] update | 1 | 39.0-39.3 |
| **Recommended** | 39.5 | Reviewer-flagged follow-ups (shell pipeline + memory_bridge) | 2 | - |

Total: **15 tasks** (12 complete + 3 pending, 1 of which is a release blocker)

### Completion Criteria (Definition of Done — Phase 39 overall)

| # | Criterion | Verification | Required/Recommended |
|---|------|---------|----------|
| 1 | /HAR:review works with bare invocation, English result summary at top + non-expert 4-section structure | Run `/harness-review` and visually confirm | Required |
| 2 | `plugin.json.skills = ["./"]` preserved after running `harness sync` | `jq '.skills' .claude-plugin/plugin.json` | Required |
| 3 | `go test ./...` all 12 packages PASS | Run command | Required |
| 4 | `./tests/validate-plugin.sh` 42 pass / 0 fail | Run command | Required |
| 5 | `./scripts/ci/check-consistency.sh` all pass | Run command | Required |
| 6 | CHANGELOG [Unreleased] contains Phase 38 + Phase 39 improvements in Before/After format | Visual inspection | Required |
| 7 | Phase 39.5 to be addressed in v4.0.2 or later (not a blocker before release) | Remains cc:TODO in Plans.md | Recommended |

---

### Phase 39.0: /HAR:review Output Improvement [P0 Done]

Purpose: Overhaul the review experience from "English JSON output for engineers" to "Japanese summary readable by non-experts + technical details folded away"

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.0.1 | Add bare invocation default flow (Step 0) to `skills/harness-review/SKILL.md`. Auto BASE_REF determination logic: `git describe --tags` → main → HEAD~10 + auto-transition to Step 1. Sync 3 mirrors | Visually confirm bare invocation reaches Step 1 automatically. validate-plugin / check-consistency maintained | - | cc:done [8d2b89cc] |
| 39.0.2 | Add "Output language/format (strictly required)" block to Step 0, add result summary top-output rule to Step 3. Mandate English output + JSON as supplementary at the end | Explicitly cite CLAUDE.md "context: fork skills also in English" rule. Confirm English output on `/harness-review` run | 39.0.1 | cc:done [af915fb4] |
| 39.0.3 | Redesign non-expert template (information granularity MID / cognitive load MIN). 6-section structure: verdict → ✨ What was good → ⚠️ Concerns (4 parts: English title → problem → response → severity → technical details) → 🎬 Next actions → 📊 Auto verification → 📦 Detailed data (JSON demoted). Sync 3 mirrors | `/harness-review` output follows new template. Severity uses "🔴 Critical / 🟠 Important / 🟡 Minor / 🟢 Recommended" format. English severity words and technical jargon isolated from body text | 39.0.2 | cc:done [7481f98f] |

---

### Phase 39.1: Infrastructure Fixes [P0 Done]

Purpose: Fix 3 root bugs identified in review (plugin.json auto-revert, jq assertion looseness, bare review excessive scope)

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.1.1 | Add `Skills []string` field to `pluginJSON` struct in `go/cmd/harness/sync.go`, hardcode `[]string{"./"}` in `generatePluginJSON()`. Root-fixes the auto-revert loop where `harness sync` kept removing `"skills": ["./"]` from plugin.json | Add `skills == ["./"]` assertion to `TestSync_GeneratesPluginJSON` and PASS. After running `harness sync .`, `jq '.skills' plugin.json` returns `["./"]` | - | cc:done [009faf74] |
| 39.1.2 | Tighten SessionStart matcher check in `tests/test-memory-hook-wiring.sh` from `contains("startup")` to pipe-token regex `test("(^|\\|)startup($|\\|)")`. Prevents typos like `startup-only` from silently passing as false positives | Verified with 6 edge cases (`startup`, `startup|resume`, `resume|startup` match; `startup-only`, `startup_special`, `resume|startup-only` reject). `bash tests/test-memory-hook-wiring.sh` OK | - | cc:done [f7146d3e] |
| 39.1.3 | Add upper-bound fallback to Step 0.1 in `skills/harness-review/SKILL.md`: if commit count from last tag to HEAD exceeds 10, auto-clamp to HEAD~10. Prevent excessive scope on bare invocation | Auto-clamp fires with 10+ commits. On review run, display both original candidate and narrowed result in summary. Sync 3 mirrors | - | cc:done [9103377f] |

---

### Phase 39.2: Name Consistency [P0 Done]

Purpose: ff4ee422 changed frontmatter `name` to `HAR:*`, but this mismatches directory name, violates skill-editing.md SSOT rule, and creates a 3-way split where internal text still says `harness-*`. Restore consistency.

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.2.1 | Revert frontmatter `name:` from `HAR:*` back to `harness-*` in 18 files (6 skills × 3 locations: main/codex/opencode). Keep "HAR:" brand in description prefix (for visual identification) | All 18 files have `name:` as `harness-*` (grep verified). `"HAR:` in description maintained at 54 locations (18 files × 3 description fields). validate-plugin / check-consistency maintained | - | cc:done [af915fb4] |

---

### Phase 39.3: v3 Cleanup Residue Removal + Test Script v4 Migration [P0 Done]

Purpose: Remove v3-era references missed in the v4.0.0 release cleanup (references to deleted TypeScript rules.ts, "TypeScript engine" in README, tests for v3 hook call patterns, v3-named artifacts)

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.3.1 | Change RULES_FILE path in `tests/validate-plugin.sh` from deleted `core/src/guardrails/rules.ts` to `go/internal/guardrail/rules.go`, sync R12 expected pattern from `warn-direct-push-protected-branch` to `deny-direct-push-protected-branch` | 4 R10-R13 failures disappear from validate-plugin.sh, pass count increases from 35 → 40 | - | cc:done [cbea4620] |
| 39.3.2 | Sync expected strings in `scripts/ci/check-consistency.sh` from `"TypeScript guardrail engine"` / `"TypeScript guardrail engine"` to `"Go-native guardrail engine"` / `"Go-native guardrail engine"` | 2 TypeScript reference failures disappear from check-consistency.sh, all pass | - | cc:done [cbea4620] |
| 39.3.3 | Migrate jq queries in `tests/test-memory-hook-wiring.sh` and `tests/test-claude-upstream-integration.sh` from v3 shell path (`hook-handlers/memory-bridge`) to v4 Go binary format (`bin/harness hook memory-bridge` equivalent). Also add `command` null handling for agent-type hooks | Both test scripts run OK directly, 2 "missing wiring" failures disappear from validate-plugin.sh | - | cc:done [c91b21c1] |
| 39.3.4 | Sync PermissionDenied wiring check in `tests/test-claude-upstream-integration.sh` from `contains("permission-denied-handler")` to `contains("permission-denied")` (matching v4's `bin/harness hook permission-denied` format) | test-claude-upstream-integration.sh runs OK directly, last remaining failure in validate-plugin.sh resolved, reaching 42/0 all pass | 39.3.3 | cc:done [04026f3a] |
| 39.3.5 | Delete v4 cleanup residue: 2 JSON-named ghost directories (side effects of Agent tool isolation errors), `core/` residue (node_modules + package-lock.json), `infographic-check.png` (debug screenshot), `.orphaned_at` (old session marker) | None of these 5 items exist at repo root. No impact on git status (all untracked) | - | cc:done [Lead direct execution] |

---

### Phase 39.4: CHANGELOG Update [P0 Required — v4.0.1 Release Blocker]

Purpose: Record all improvements from Phase 38 and Phase 39 in the [Unreleased] section of CHANGELOG.md in Before/After format so users correctly understand the changes when v4.0.1 is released

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.4.1 | Append Phase 39 improvements to `[Unreleased]` section of `CHANGELOG.md`. Keep existing Phase 38 entries (CC 2.1.89-2.1.100 tracking) and add new sections such as "#### 8. Review Experience Improvements (Phase 39)" below them. Each change described in Before/After format (per `.claude/rules/github-release.md`). Covers: (a) /HAR:review bare invocation + non-expert template, (b) sync.go Skills field root fix, (c) SessionStart matcher strictening, (d) bare review scope cap, (e) name revert, (f) v3 cleanup residue removal, (g) test scripts v4 migration | CHANGELOG.md Unreleased has Phase 39 sub-entries (items 8-12) added. VERSION / plugin.json version / harness.toml version not changed (not a release operation). `./scripts/ci/check-consistency.sh` PASS, `./tests/validate-plugin.sh` 42 pass / 0 fail | 39.0.1-39.3.5 | cc:done [c96ca7d1] |

---

### Phase 39.5: Reviewer-Flagged Follow-ups [P1 Recommended — v4.0.2 and later]

Purpose: Track the 2 recommendations found in previous and current /HAR:review sessions. Neither blocks current operation; neither should block the v4.0.1 release.

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 39.5.1 | Harden `grep -c '^plan:' \| awk '$1 > 2 {exit 0} {exit 1}'` pipeline in Step 0.2 of `skills/harness-review/SKILL.md` for strict-mode compatibility. Use single awk integration like `awk '/^plan:/ {n++} END {exit (n<=2)}'`, or add `\|\| true` as a safety net. Sync 3 mirrors | Bare invocation does not falsely stop on 0 plan: entries under `set -euo pipefail`. Existing behavior (plan review mode when plan: > 2 entries) maintained | - | cc:TODO |
| 39.5.2 | Fix issue in `go/internal/hookhandler/memory_bridge.go` where `validTargets` map expects kebab-case (`session-start`, `user-prompt` etc.) but CC sends PascalCase (`SessionStart`, `UserPromptSubmit` etc.) in `hook_event_name`, causing it to always fall into "unknown target" fail-open branch and rendering memory bridge effectively non-functional. Normalize HookEventName or align `validTargets` keys to PascalCase | Memory bridge actually dispatches each of session-start / user-prompt / post-tool-use / stop events. Add PascalCase input test cases to hook handler tests | - | cc:TODO |

---


## Phase 38: CC 2.1.89-2.1.100 Tracking + Go v4 Release Polish

Created: 2026-04-10
Purpose: Incorporate unabsorbed hook/permission/plugin changes from Claude Code 2.1.89-2.1.100 into the Harness v4 Go guardrails, achieving a perfect zero-security-regression state before the Go v4.0.0 release

Background: Two vulnerabilities patched in CC 2.1.98 (backslash-escaped flag bypass, env-var prefix bypass) remain open in Harness's second-layer guardrails. Additionally, `DecisionDefer` from CC 2.1.89 has a type definition in `go/pkg/hookproto/types.go` but is not caught in the `PreToolToOutput()` switch in `go/internal/guardrail/pre_tool.go` — a known gap. .husky protection / symlink resolution / wildcard normalization / plugin.json skills explicit declaration / Monitor tool adoption should all be incorporated before the v4 release.

### Priority Matrix

| Priority | Phase | Content | Task Count | Depends |
|--------|-------|------|---------|------|
| **Required** | 38.0 | Security emergency fix (permission.go hardening + DecisionDefer wiring) | 2 | None |
| **Required** | 38.1 | Security hardening (.husky + symlink + wildcard normalization) | 2 | None |
| **Required** | 38.2 | Plugin/skill alignment + Monitor tool adoption | 2 | None |
| **Required** | 38.3 | Integration verification, binary rebuild, CHANGELOG | 3 | All of 38.0-38.2 |

Total: **9 tasks**

### Completion Criteria (Definition of Done — Phase 38 overall)

| # | Criterion | Verification | Required/Recommended |
|---|------|---------|----------|
| 1 | 16+ new guardrail security tests added, all PASS | `go test -v ./internal/guardrail/...` | Required |
| 2 | All Go tests PASS | `go test ./...` | Required |
| 3 | Plugin validation PASS | `./tests/validate-plugin.sh` | Required |
| 4 | Consistency check PASS | `./scripts/ci/check-consistency.sh` | Required |
| 5 | Feature Table has CC 2.1.98 Monitor entry + value-add column "A: Has implementation" | Visual inspection of `docs/CLAUDE-feature-table.md` | Required |
| 6 | CHANGELOG.md Unreleased has 7 Before/After entries for Phase 38 | Visual inspection of `CHANGELOG.md` | Required |
| 7 | Binaries rebuilt for 3 platforms | `ls -la bin/harness-*` | Required |
| 8 | CC 2.1.89-2.1.100 security hardening fully reflected in Go v4 | All criteria 1-7 met | Required |

---

### Phase 38.0: Security Emergency Fix (permission.go + DecisionDefer) [P0]

Purpose: Patch in Harness's second-layer guard the 2 vulnerabilities fixed in CC 2.1.98 (backslash-escape / env-var prefix). Also resolve the known gap where `DecisionDefer` added in 2.1.89 is not caught in the switch case.

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 38.0.1 | `go/internal/guardrail/permission.go` hardening: (1) Add `hasBackslashEscape(cmd string) bool` function, detect patterns like `\-`, `\\ ` (space-escape), `\--` using regex `\\[\-\s]`, (2) Add `stripSafeEnvPrefix(cmd string) (string, bool)` function and `knownSafeEnvVars` map (`LANG`, `LANGUAGE`, `TZ`, `NO_COLOR`, `FORCE_COLOR`). `LC_*` allowed via prefix match, (3) At the top of `isSafeCommand()`, call both; return `false` immediately on backslash escape detection or unknown env-var detection [feature:security] | Minimum 8 tests added to `go/internal/guardrail/permission_test.go`, all PASS: (a) `git\ status` reject, (b) `git\ push\ --force` reject, (c) `rm\ -rf\ /` reject, (d) `LANG=C git status` pass, (e) `TZ=UTC git log` pass, (f) `EVIL=x git status` reject, (g) `LANG=C NO_COLOR=1 git status` pass, (h) `LANG=C EVIL=x git status` reject. `go test ./internal/guardrail/...` all PASS | - | cc:done [aa9f4bb] |
| 38.0.2 | Add `case hookproto.DecisionDefer:` to the switch statement in `PreToolToOutput()` function in `go/internal/guardrail/pre_tool.go`, setting `inner.PermissionDecision = "defer"` + `inner.PermissionDecisionReason = result.Reason`. Resolves the known issue where `DecisionDefer` constant already defined at `go/pkg/hookproto/types.go:39` but not caught in switch [feature:security] | Test added to `go/internal/guardrail/pre_tool_test.go` for DecisionDefer return case, output JSON contains `"permissionDecision": "defer"` and `"permissionDecisionReason": "<reason>"`. `go test ./internal/guardrail/...` all PASS | - | cc:done [aa9f4bb] |

---

### Phase 38.1: Security Hardening (helpers.go + rules.go) [P0]

Purpose: Port 3 items to Harness Go guardrails: CC 2.1.89 symlink target resolution, CC 2.1.90 `.husky` protected path addition, CC 2.1.98 wildcard whitespace normalization

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 38.1.1 | `go/internal/guardrail/helpers.go` extension: (1) Add `.husky/` pattern to `protectedPathPatterns` slice (protect git hooks directory, following CC 2.1.90), (2) Inside `isProtectedPath(path string) bool`, call `filepath.EvalSymlinks()` and also deny when the resolved real path after symlink resolution matches protected patterns (following CC 2.1.89). Return `true` (deny) as fail-safe when `EvalSymlinks` errors [feature:security] | Minimum 5 tests added to `go/internal/guardrail/helpers_test.go`, all PASS: (a) `.husky/pre-commit` write deny, (b) `.husky/hooks/commit-msg` deny, (c) Create symlink `link-env → .env` in temp dir and verify access is denied, (d) 2-level symlink chain (link1 → link2 → .env) also denied, (e) `EvalSymlinks` error on symlink loop also fail-safes to deny. `go test ./internal/guardrail/...` all PASS | - | cc:done [aa9f4bb] |
| 38.1.2 | In wildcard pattern evaluation in `go/internal/guardrail/rules.go`, normalize consecutive whitespace (spaces/tabs) in user command to single space using `regexp.MustCompile(\`\s+\`).ReplaceAllString(cmd, " ")` before pattern matching. Reproduce in second layer the CC 2.1.98 fix where `Bash(git push -f:*)` matches commands with multiple spaces/tabs [feature:security] | Minimum 3 tests added to `go/internal/guardrail/rules_test.go`, all PASS: (a) `git  push  --force` (consecutive spaces) denied by force-push rule, (b) `git\tpush\t-f` (tab-separated) denied, (c) `git push   --force-with-lease` denied. Existing tests all maintained. `go test ./internal/guardrail/...` all PASS | - | cc:done [aa9f4bb] |

---

### Phase 38.2: Plugin/Skill Alignment + Monitor Tool Adoption [P0]

Purpose: Explicitly support the CC 2.1.94 plugin skill invocation name spec, declare and document the Monitor tool added in CC 2.1.98 for use in long-running skills like Breezing

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 38.2.1 | Add `"skills": ["./"]` field to `.claude-plugin/plugin.json`. Explicitly support the CC 2.1.94+ spec where plugin skill invocation names are based on the frontmatter `name` field. Maintain compatibility with existing auto-discover (all existing 32 skills remain invocable) | `.claude-plugin/plugin.json` has `"skills": ["./"]`. `./tests/validate-plugin.sh` PASS. `./scripts/ci/check-consistency.sh` PASS. Confirm `.skills` is `["./"]` with jq | - | cc:done [ebdf47b] |
| 38.2.2 | CC 2.1.98 Monitor tool support: (1) Add `"Monitor"` to `allowed-tools` array in frontmatter of `skills/breezing/SKILL.md`, `skills/harness-work/SKILL.md`, `skills/ci/SKILL.md`, `skills/deploy/SKILL.md`, `skills/harness-review/SKILL.md`, (2) Add "### Monitor Tool Usage Guide (CC 2.1.98+)" section to `skills/breezing/SKILL.md` (document usage guidelines: Worker observation done via Agent layer completion notification so Monitor not needed there; for shell long-running command monitoring prefer Monitor; list concrete examples like `gh run watch`, `go test -v`, `codex-companion.sh watch <job-id>`), (3) Add "Monitor tool" row to `docs/CLAUDE-feature-table.md` with "A: Has implementation (allowed-tools + usage guide + Feature Table)" in value-add column | 5 SKILL.md files contain `"Monitor"` in frontmatter (`grep -l '"Monitor"' skills/*/SKILL.md` returns 5 hits). breezing SKILL.md has `### Monitor Tool Usage Guide (CC 2.1.98+)` section. `docs/CLAUDE-feature-table.md` has Monitor row with "A: Has implementation" in value-add column. Does not trigger "documentation only" detection in `.claude/rules/cc-update-policy.md` (has implementation) | - | cc:done [ebdf47b] |

---

### Phase 38.3: Integration Verification, Binary Rebuild, CHANGELOG [P0]

Purpose: Integrate all changes from Phase 38.0-38.2, rebuild binaries on 3 platforms, record in CHANGELOG in Before/After format, and reach a state ready for Go v4.0.0 release

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 38.3.1 | Rebuild `bin/harness` for 3 platforms: darwin-arm64, darwin-amd64, linux-amd64. Use existing cross-platform build script (`scripts/build-cross-platform.sh` etc. if available), otherwise run `GOOS=darwin GOARCH=arm64 go build -o bin/harness-darwin-arm64 ./go/cmd/harness` for all 3 patterns | `bin/harness-darwin-arm64`, `bin/harness-darwin-amd64`, `bin/harness-linux-amd64` all updated (latest timestamp in `ls -la`). Each binary returns v4.0.0 from `--version`. Binary size increases to ~10-11MB due to Phase 37.5 additions (Phase 38 impact negligible) | 38.0.1, 38.0.2, 38.1.1, 38.1.2 | cc:done [fbed2f9] |
| 38.3.2 | Append Phase 38 items to `[Unreleased]` section of `CHANGELOG.md` in Before/After format. Cover all 7 items from CC 2.1.89-2.1.100 (backslash / env-var / defer / .husky+symlink / wildcard / plugin.json / Monitor). Use "CC Update → Harness Usage" format from `.claude/rules/github-release.md` | CHANGELOG.md Unreleased has 7 entries added in `#### N. Claude Code 2.1.98 Integration` + `##### N-X.` format. Each entry has 2-part format: "CC Update" and "Harness Usage". VERSION / plugin.json version / harness.toml version not changed (not a release operation) | 38.0.1, 38.0.2, 38.1.1, 38.1.2, 38.2.1, 38.2.2 | cc:done [fbed2f9] |
| 38.3.3 | Integration tests: (1) `go test ./...` all PASS, (2) `./tests/validate-plugin.sh` PASS, (3) `./scripts/ci/check-consistency.sh` PASS — all 3 confirmed. If any fail, identify root cause and fix before marking task complete | All 3 commands PASS. Confirm all 24 newly added security tests are included and no regression in existing tests. validate-plugin.sh improves from baseline 7 failures to 6 failures (1 improvement from plugin.json skills field addition); check-consistency.sh maintains baseline 2 (all v4 cleanup residue, unrelated to Phase 38) | 38.3.1, 38.3.2 | cc:done [fbed2f9] |

---


## Phase 37: Full Go Port of All Hook Handlers — "Complete Hokage"

Created: 2026-04-09
Purpose: Port the remaining 37 bash/Node.js handlers in hooks.json to Go subcommands, fully eliminating run-hook.sh + run-script.js + Node.js runtime dependency
Prerequisites: Phase 35.0-35.2 complete, v4.0 "Hokage" committed (14 hooks already ported to Go)

### Design Principles

- Implement each handler as a `runHook()` subcommand in `go/cmd/harness/main.go`
- Port existing bash behavior 1:1 (no feature additions)
- Centralize implementation in `internal/hookhandler/` package
- Create `_test.go` for each handler (minimum input/output tests)
- Rewrite the corresponding hooks.json entry to `bin/harness hook <name>`

### Node.js Dependency Status

| Dependency | Count | Target |
|------|------|------|
| Node.js required | 2 | pre-compact-save.js (783 lines), emit-agent-trace.js (808 lines) |
| Pure bash | 35 | All others |

**Only 2 files need to be ported to achieve zero Node.js dependency.**

---

### Phase 37.1: Trivial Handlers (10) [cc:TODO]

Difficulty: Low / ~50-100 lines each / File I/O + JSON output only

| Task | Handler | Source File | Lines | Description | Status |
|------|---------|-----------|------|------|--------|
| 37.1.1 | pretooluse-inbox-check | scripts/pretooluse-inbox-check.sh | 82 | Check for unread messages from other sessions (5-minute throttle) | cc:TODO |
| 37.1.2 | pretooluse-browser-guide | scripts/pretooluse-browser-guide.sh | 84 | Detect agent-browser CLI + recommend MCP browser tools | cc:TODO |
| 37.1.3 | memory-bridge | scripts/hook-handlers/memory-bridge.sh + 4 sub-handlers | 55 | harness-mem MCP bridge dispatcher (session-start/user-prompt/post-tool/stop) | cc:TODO |
| 37.1.4 | worktree-create | scripts/hook-handlers/worktree-create.sh | 93 | Create .claude/state/ + record worktree-info.json | cc:TODO |
| 37.1.5 | worktree-remove | scripts/hook-handlers/worktree-remove.sh | 73 | Delete tmp files + delete worktree-info.json | cc:TODO |
| 37.1.6 | posttooluse-commit-cleanup | scripts/posttooluse-commit-cleanup.sh | 50 | Detect git commit → delete review-approved.json | cc:TODO |
| 37.1.7 | posttooluse-clear-pending | scripts/posttooluse-clear-pending.sh | 28 | Delete pending-skills/*.pending (skill completion signal) | cc:TODO |
| 37.1.8 | session-auto-broadcast | scripts/session-auto-broadcast.sh | 103 | Notify teammates on changes to src/api/, types/, schema | cc:TODO |
| 37.1.9 | config-change | scripts/hook-handlers/config-change.sh | 92 | ConfigChange → record to breezing-timeline.jsonl | cc:TODO |
| 37.1.10 | instructions-loaded | scripts/hook-handlers/instructions-loaded.sh | 86 | InstructionsLoaded → jsonl log + hooks.json existence verification | cc:TODO |

---

### Phase 37.2: Medium Handlers (12) [cc:TODO]

Difficulty: Medium / ~100-350 lines each / JSONL management, state tracking, conditional branching

| Task | Handler | Source File | Lines | Description | Status |
|------|---------|-----------|------|------|--------|
| 37.2.1 | setup-hook | scripts/setup-hook.sh | 188 | Plugin cache sync + .claude/state initialization + template validation | cc:TODO |
| 37.2.2 | runtime-reactive | scripts/hook-handlers/runtime-reactive.sh | 168 | FileChanged/CwdChanged/TaskCreated → context injection | cc:TODO |
| 37.2.3 | teammate-idle | scripts/hook-handlers/teammate-idle.sh | 186 | Record teammate idle + continue:false stop signal | cc:TODO |
| 37.2.4 | userprompt-track-command | scripts/userprompt-track-command.sh | 107 | Detect /slash commands + record usage + pending-skills markers | cc:TODO |
| 37.2.5 | breezing-signal-injector | scripts/hook-handlers/breezing-signal-injector.sh | 183 | breezing-signals.jsonl → systemMessage injection + mark consumed | cc:TODO |
| 37.2.6 | ci-status-checker | scripts/hook-handlers/ci-status-checker.sh | 192 | Detect git push/gh pr → async CI status check | cc:TODO |
| 37.2.7 | usage-tracker | scripts/usage-tracker.sh | 108 | Track Skill/Task tool usage | cc:TODO |
| 37.2.8 | todo-sync | scripts/todo-sync.sh | 118 | TodoWrite → Plans.md marker sync (pending→cc:TODO etc.) | cc:TODO |
| 37.2.9 | auto-cleanup-hook | scripts/auto-cleanup-hook.sh | 118 | File size warning after Write/Edit (>10KB) | cc:TODO |
| 37.2.10 | track-changes | scripts/track-changes.sh | 185 | Record file changes + 2-hour dedup + path normalization | cc:TODO |
| 37.2.11 | plans-watcher | scripts/plans-watcher.sh | 201 | Detect Plans.md changes + inject WIP/TODO/done marker summary | cc:TODO |
| 37.2.12 | tdd-order-check | scripts/tdd-order-check.sh | 115 | Warning for implementation files edited before tests (enforce TDD order) | cc:TODO |

---

### Phase 37.3: Medium Handlers — Supplement Existing Go (7) [cc:TODO]

Go binary already has routing, but hooks.json still calls bash

| Task | Handler | Source File | Lines | Description | Status |
|------|---------|-----------|------|------|--------|
| 37.3.1 | elicitation-handler | scripts/hook-handlers/elicitation-handler.sh | 139 | MCP Elicitation → log + auto-skip during Breezing | cc:TODO |
| 37.3.2 | elicitation-result | scripts/hook-handlers/elicitation-result.sh | 123 | ElicitationResult → jsonl log | cc:TODO |
| 37.3.3 | stop-session-evaluator | scripts/hook-handlers/stop-session-evaluator.sh | 106 | Stop → session state evaluation + session.json update | cc:TODO |
| 37.3.4 | stop-failure | scripts/hook-handlers/stop-failure.sh | 178 | StopFailure → API error log (rate limit, auth) | cc:TODO |
| 37.3.5 | notification-handler | scripts/hook-handlers/notification-handler.sh | 166 | Notification → record to notification-events.jsonl | cc:TODO |
| 37.3.6 | permission-denied-handler | scripts/hook-handlers/permission-denied-handler.sh | 197 | PermissionDenied → denial log + notify Breezing Lead | cc:TODO |
| 37.3.7 | posttooluse-quality-pack | scripts/posttooluse-quality-pack.sh | 190 | Quality checks after Write/Edit (Prettier, tsc, console.log detection) | cc:TODO |

---

### Phase 37.4: Hard Handlers (8) [cc:TODO]

Difficulty: High / ~300-900 lines each / state machines, process control, Node.js ports

| Task | Handler | Source File | Lines | Description | Status |
|------|---------|-----------|------|------|--------|
| 37.4.1 | userprompt-inject-policy | scripts/userprompt-inject-policy.sh | 351 | Memory resume context injection + semaphore lock + RESUME_MAX_BYTES limit | cc:TODO |
| 37.4.2 | fix-proposal-injector | scripts/hook-handlers/fix-proposal-injector.sh | 338 | pending-fix-proposals.jsonl → display proposals + approve/reject → Plans.md sync | cc:TODO |
| 37.4.3 | posttooluse-log-toolname | scripts/posttooluse-log-toolname.sh | 333 | Tool usage log + LSP tracking + session event log (500-line rotation) + flock | cc:TODO |
| 37.4.4 | auto-test-runner | scripts/auto-test-runner.sh | 326 | Detect source file changes → auto-run tests (async) + auto-detect Vitest/Jest/pytest | cc:TODO |
| 37.4.5 | task-completed | scripts/hook-handlers/task-completed.sh | 911 | Record task completion + generate fix proposals + Breezing timeline + Plans.md sync (largest) | cc:TODO |
| 37.4.6 | **pre-compact-save.js** ⚡ | scripts/hook-handlers/pre-compact-save.js | 783 | **Node.js** — Generate handoff-artifact.json + precompact-snapshot.json + collect Git info | cc:TODO |
| 37.4.7 | **emit-agent-trace.js** ⚡ | scripts/emit-agent-trace.js | 808 | **Node.js** — Record agent-trace.jsonl + OpenTelemetry span + 10MB/3-generation rotation | cc:TODO |
| 37.4.8 | post-compact (extended) | scripts/hook-handlers/post-compact.sh | 380 | PostCompact extension — WIP context + handoff artifact re-injection (supplement current Go version) | cc:TODO |

⚡ = Node.js dependency. Porting these 2 files achieves zero Node.js dependency.

---

### Phase 37.5: Final hooks.json Rewrite + Legacy Deletion [cc:TODO]

| Task | Description | DoD | Status |
|------|------|-----|--------|
| 37.5.1 | Rewrite all remaining 37 entries in hooks.json to `bin/harness hook <name>` | `grep -c 'run-hook.sh' hooks/hooks.json` returns 0 | cc:TODO |
| 37.5.2 | Delete `scripts/run-hook.sh` + `scripts/run-script.js` | Files do not exist | cc:TODO |
| 37.5.3 | Delete `package.json` (complete removal of npm dependency) | File does not exist | cc:TODO |
| 37.5.4 | Delete `core/` (TypeScript engine) | Directory does not exist | cc:TODO |
| 37.5.5 | E2E tests: all hook events work via Go binary | `go/scripts/test-e2e.sh` all pass | cc:TODO |
| 37.5.6 | `harness doctor` confirms zero Node.js dependency | `grep -rE "node\|run-script" hooks/` returns 0 hits | cc:TODO |

---

### Phase 37 Completion Criteria

| # | Criterion | Verification |
|---|------|---------|
| 1 | 0 references to `run-hook.sh` in hooks.json | `grep 'run-hook' hooks/hooks.json` |
| 2 | `scripts/run-hook.sh`, `run-script.js`, `package.json`, `core/` are deleted | Verified with `ls` |
| 3 | 0 references to `node` command inside harness (except codex-companion) | `grep -r 'node ' scripts/ hooks/` |
| 4 | Go tests (`_test.go`) exist for all 37 handlers | `go test ./internal/hookhandler/...` |
| 5 | `harness doctor` passes all checks | `bin/harness doctor` |
| 6 | `go/scripts/test-e2e.sh` covers all hook events | All E2E pass |

Total: **37 handler ports + 6 cleanups = 43 tasks**

---


## Phase 37: Upstream Merge — upstream/main → local master (EN translation)

Created: 2026-04-13
Purpose: Merge upstream/main (v4.0.2) into local master (v3.17.5), translating all Japanese content to English. Upstream is the source of truth; local master diverged with EN translation + simplification.

### Merge Principles

1. **Upstream is source of truth** — all upstream changes take precedence
2. **English only** — all Japanese content from upstream must be translated to English
3. **Deletions follow upstream** — anything deleted upstream must be deleted locally
4. **Creations follow upstream** — anything new upstream must be created locally (in English)
5. **Conflicts resolve to upstream** — when both sides changed, use upstream content (translated)
6. **README uses local master** — local README.md is already English and well-maintained
7. **Plans.md merges both** — combine upstream phases (35–40) with local phases (35–36)
8. **Skills follow upstream** — new skills added, updated skills synced, deleted skills removed (all in EN)
9. **Go engine adopted** — the "Hokage" Go hook engine replaces shell + TypeScript

### Priority Matrix

| Priority | Phase | Content | Tasks | Depends |
|----------|-------|---------|-------|---------|
| **Required** | 37.0 | Go engine + core infrastructure | 5 | - |
| **Required** | 37.1 | Hooks + hook handlers migration | 4 | 37.0 |
| **Required** | 37.2 | Scripts sync (deletions + modifications) | 4 | 37.0 |
| **Required** | 37.3 | Skills sync (deletions + modifications + creations) | 5 | 37.0 |
| **Required** | 37.4 | Agents sync | 3 | 37.0 |
| **Required** | 37.5 | Rules, config, and docs sync | 5 | 37.3, 37.4 |
| **Required** | 37.6 | Tests migration to v4 patterns | 4 | 37.1, 37.2 |
| **Required** | 37.7 | Templates, workflows, benchmarks sync | 3 | 37.5 |
| **Required** | 37.8 | Plans.md merge + CHANGELOG + final validation | 4 | 37.0–37.7 |

Total: **37 tasks**

### Completion Criteria (Definition of Done — Phase 37)

| # | Criterion | Verification | Required |
|---|-----------|-------------|----------|
| 1 | `go/` directory exists with full Hokage engine | `ls go/cmd/harness/main.go` exists | Required |
| 2 | All upstream deletions applied | `git diff --name-status upstream/main master` shows no unexpected D entries | Required |
| 3 | All Japanese content translated to English | `grep -rP '[\x{3040}-\x{309F}\x{30A0}-\x{30FF}\x{4E00}-\x{9FFF}]' --include='*.md' --include='*.sh' --include='*.go'` returns only CHANGELOG historical entries | Required |
| 4 | `./tests/validate-plugin.sh` passes | 43+ pass / 0 fail | Required |
| 5 | `./scripts/ci/check-consistency.sh` passes | All checks pass | Required |
| 6 | README.md is local master version (English) | No upstream README overwrite | Required |
| 7 | Plans.md contains both upstream phases and local phases | Phases 35-40 all present | Required |

---

### Phase 37.0: Go Engine + Core Infrastructure [P0]

Purpose: Bring the entire `go/` directory and updated `core/` from upstream, translate comments/docs to English

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 37.0.1 | Checkout `go/` directory from upstream/main (entire Go engine: cmd/, internal/, pkg/, go.mod, go.sum, DESIGN.md, SPEC.md) and translate all Japanese comments/docs to English | `go/` exists, `go build ./...` succeeds, no Japanese in `.go` or `.md` files under `go/` | - | cc:done |
| 37.0.2 | Sync `core/` from upstream (bun.lock, dist/ updates, src/ updates for v4 compat) | `core/` matches upstream structure, `ls core/bun.lock` exists | 37.0.1 | cc:done |
| 37.0.3 | Sync root config files: `.claude-code-harness.config.yaml`, `.gitignore` from upstream | Files match upstream content | - | cc:done |
| 37.0.4 | Delete files removed by upstream: `VERSION`, `LICENSE.ja.md`, `README_ja.md`, `.claude-plugin/plugin.json`, `.cursor/rules/skill-subagent-usage.mdc`, `.github/workflows/opencode-compat.yml` | None of these files exist | - | cc:done |
| 37.0.5 | Sync `.claude-plugin/marketplace.json`, `.claude-plugin/hooks.json` from upstream (translate JP descriptions to EN) | Files match upstream structure, all descriptions in English | 37.0.1 | cc:done |

---

### Phase 37.1: Hooks + Hook Handlers Migration [P0]

Purpose: Replace shell+TypeScript hook engine with Go binary hook invocations from upstream

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 37.1.1 | Sync `hooks/hooks.json` from upstream — update all hook commands from shell/node invocations to Go binary `bin/harness hook <name>` pattern | `hooks/hooks.json` matches upstream structure | 37.0.1 | cc:TODO |
| 37.1.2 | Sync hook shim scripts (`hooks/pre-tool.sh`, `hooks/post-tool.sh`, `hooks/permission.sh`, etc.) to call Go binary instead of node | Hook shims delegate to `bin/harness` | 37.1.1 | cc:TODO |
| 37.1.3 | Sync all `hooks/*.sh` files from upstream, translate any Japanese comments to English | All hook scripts match upstream behavior, English only | 37.1.1 | cc:TODO |
| 37.1.4 | Delete hook files removed by upstream (if any remain from local Phase 36) | No stale hook files | 37.1.1 | cc:TODO |

---

### Phase 37.2: Scripts Sync [P0]

Purpose: Apply upstream deletions and modifications to scripts/

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 37.2.1 | Delete scripts already removed in Phase 36 that upstream also removed (verify alignment): `pretooluse-guard.sh`, `posttooluse-*.sh`, `stop-*.sh`, `permission-request.sh`, `skill-child-reminder.sh`, `sync-v3-skill-mirrors.sh`, `build-opencode.js`, `validate-opencode.js`, `setup-opencode.sh`, `opencode-setup-local.sh` | All 13 deleted scripts confirmed absent | - | cc:TODO |
| 37.2.2 | Sync modified scripts from upstream (translate JP→EN): `sync-plugin-cache.sh`, `codex-companion.sh`, `check-consistency.sh`, `generate-skill-manifest.sh`, `validate-release-notes.sh`, `write-review-result.sh`, `build-cross-platform.sh`, `i18n/set-locale.sh` | Modified scripts match upstream behavior, English only | 37.0.1 | cc:TODO |
| 37.2.3 | Add new scripts from upstream (translate JP→EN): `check-residue.sh`, `harness-mem-bridge.sh` and any others created in Phases 35-40 | New scripts exist and are English | 37.0.1 | cc:TODO |
| 37.2.4 | Sync CI workflows from upstream: `.github/workflows/validate-plugin.yml`, `.github/workflows/release.yml`, `.github/workflows/benchmark.yml` (translate JP→EN) | CI files match upstream, English only | 37.2.2 | cc:TODO |

---

### Phase 37.3: Skills Sync [P0]

Purpose: Sync all skills with upstream — deletions, modifications, and new additions (all in English)

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 37.3.1 | Delete skills removed by upstream: `skills/allow1/`, `skills/claude-codex-upstream-update/` | Directories do not exist | - | cc:TODO |
| 37.3.2 | Sync all 85+ modified skills from upstream (translate JP→EN): frontmatter descriptions, SKILL.md content, references/*.md files. Key skills: `harness-review`, `harness-work`, `harness-plan`, `harness-release`, `harness-setup`, `breezing`, `ci`, `deploy`, `auth`, `crud`, `generate-video`, `generate-slide`, `agent-browser`, `vibecoder-guide`, `workflow-guide`, etc. | All skills match upstream structure and features, English only | 37.0.1 | cc:TODO |
| 37.3.3 | Add new skill files from upstream: `skills/generate-video/schemas/IMPLEMENTATION_SUMMARY.md` | File exists, English content | 37.3.2 | cc:TODO |
| 37.3.4 | Verify skill frontmatter: all `name:` fields match directory names, `description:` fields are English, version references say v4 not v3 | `grep -r 'Harness v3' skills/` returns no results, all `name:` match dirs | 37.3.2 | cc:TODO |
| 37.3.5 | Remove `codex/.codex/skills/` symlinks and `opencode/skills/` mirrors if upstream removed them; OR sync if upstream retained them | Mirror state matches upstream | 37.3.2 | cc:TODO |

---

### Phase 37.4: Agents Sync [P0]

Purpose: Sync agent files with upstream consolidation

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 37.4.1 | Verify agents already deleted in Phase 36 align with upstream: `code-reviewer.md`, `codex-implementer.md`, `plan-analyst.md`, `plan-critic.md`, `project-analyzer.md`, `project-scaffolder.md`, `project-state-updater.md`, `task-worker.md` | All 8 legacy agents absent | - | cc:TODO |
| 37.4.2 | Sync existing agents from upstream (translate JP→EN): `ci-cd-fixer.md`, `error-recovery.md`, `video-scene-generator.md` | Content matches upstream, English only | 37.0.1 | cc:TODO |
| 37.4.3 | Add new agents from upstream (translate JP→EN): `reviewer.md`, `scaffolder.md`, `worker.md`, `team-composition.md` — verify these supersede any local versions | 4 new agents exist, English only, v4 patterns | 37.4.1 | cc:TODO |

---

### Phase 37.5: Rules, Config, and Docs Sync [P0]

Purpose: Sync .claude/rules/, docs/, and other config files

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 37.5.1 | Sync modified rules from upstream (translate JP→EN): `cc-update-policy.md`, `codex-cli-only.md`, `github-release.md`, `hooks-editing.md`, `implementation-quality.md`, `skill-editing.md`, `test-quality.md`, `v3-architecture.md`, `versioning.md` | Rules match upstream features, English only | 37.0.1 | cc:TODO |
| 37.5.2 | Delete rules removed by upstream: `command-editing.md` (already done in Phase 36), `deleted-concepts.yaml`, `migration-policy.md`, `self-audit.md`, `version-drift.md` | Files do not exist | - | cc:TODO |
| 37.5.3 | Add new rules from upstream (translate JP→EN): any new rules in `.claude/rules/` not present locally | New rules exist, English only | 37.5.1 | cc:TODO |
| 37.5.4 | Sync `CLAUDE.md` from upstream (translate JP→EN) — merge carefully preserving local EN improvements | CLAUDE.md has v4 references, English only | 37.3, 37.4 | cc:TODO |
| 37.5.5 | Sync docs/ directory: `CLAUDE-feature-table.md`, `CLAUDE-skill-catalog.md`, `distribution-scope.md`, `CLAUDE_CODE_COMPATIBILITY.md`, architecture docs, plans/ docs (translate JP→EN). Keep `README.md` as local version | docs/ matches upstream features, English only. README.md untouched | 37.5.1 | cc:TODO |

---

### Phase 37.6: Tests Migration to v4 [P0]

Purpose: Update all test scripts to v4 patterns (Go binary, no TypeScript refs)

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 37.6.1 | Sync `tests/validate-plugin.sh` and `tests/validate-plugin-v3.sh` from upstream — update to reference Go guardrails instead of TypeScript | Tests reference `go/internal/guardrail/rules.go` | 37.1, 37.2 | cc:TODO |
| 37.6.2 | Sync all modified test scripts (30+ files) from upstream (translate JP→EN): hook wiring tests, integration tests, guardrail tests | All test scripts match upstream v4 patterns, English only | 37.6.1 | cc:TODO |
| 37.6.3 | Sync test fixtures from upstream: `tests/fixtures/` | Fixtures match upstream | 37.6.2 | cc:TODO |
| 37.6.4 | Run `./tests/validate-plugin.sh` and verify 43+ pass / 0 fail | Test suite green | 37.6.2 | cc:TODO |

---

### Phase 37.7: Templates, Workflows, Benchmarks Sync [P1]

Purpose: Sync remaining directories

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 37.7.1 | Sync `templates/` from upstream (translate JP→EN): all `.template` files, `template-registry.json` | Templates match upstream, English only | 37.5 | cc:TODO |
| 37.7.2 | Sync `workflows/` from upstream (translate JP→EN): `default/init.yaml`, `plan.yaml`, `review.yaml`, `work.yaml` | Workflows match upstream, English only | 37.5 | cc:TODO |
| 37.7.3 | Sync `benchmarks/` from upstream (translate JP→EN): `breezing-bench/` reports, analyzer, eval prompts | Benchmarks match upstream, English only | 37.5 | cc:TODO |

---

### Phase 37.8: Plans.md Merge + CHANGELOG + Final Validation [P0]

Purpose: Merge Plans.md from both branches, update CHANGELOG, run final validation

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 37.8.1 | Merge Plans.md: keep local Phase 35 (rebrand) and Phase 36 (simplification) as completed history. Add upstream Phases 35-40 (Go rewrite through Migration Residue Scanner) translated to English. Renumber to avoid conflicts (upstream Phase 35 → Phase 38 in merged plan, etc.) or keep as-is with clear labels | Plans.md contains all phases from both branches, English only | 37.0–37.7 | cc:TODO |
| 37.8.2 | Sync `CHANGELOG.md` from upstream, translate all entries to English. Preserve any local-only entries | CHANGELOG has v4.0.0–v4.0.2 entries in English | 37.8.1 | cc:TODO |
| 37.8.3 | Sync `CONTRIBUTING.md`, `.claude/output-styles/harness-ops.md`, and remaining misc files from upstream (translate JP→EN) | All misc files synced, English only | 37.8.1 | cc:TODO |
| 37.8.4 | Final validation: run `./tests/validate-plugin.sh`, `./scripts/ci/check-consistency.sh`, verify no Japanese remains outside CHANGELOG historical entries | All checks pass, English-only codebase | 37.8.1–37.8.3 | cc:TODO |

---


## Phase 36: skills-v3/ Integration — Consolidate SSOT into skills/

Created: 2026-04-07
Purpose: Retire the `skills-v3/` directory and make `skills/` the sole SSOT. Simplify mirror sync to only 2 directions: `skills/ → codex/.codex/skills/` and `opencode/skills/`.

### Background

Phase A (completed in previous session) finished renaming `agents-v3/ → agents/` and `skills-v3-codex/ → skills-codex/`.
Phase B is the `skills-v3/` consolidation, which was separated into its own PR because it requires redesigning the mirror script group.

### Design Principles

- The 6 core skills + breezing + routing-rules.md from `skills-v3/` already have mirror copies in `skills/`
- After consolidation, `skills/` is the source of truth. `skills-v3/` is deleted
- The 10 symlinks in `skills-v3/extensions/` are unnecessary since the actual files exist in `skills/`
- Change source paths in mirror sync scripts from `skills-v3/` → `skills/`
- Remove "v3" from all paths and documentation (excluding past CHANGELOG entries)

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 36.1 | Final sync of `skills-v3/` content to `skills/`, confirm no diff | `diff -r skills-v3/{skill} skills/{skill}` shows diff 0 | - | cc:done |
| 36.2 | Rewrite `sync-v3-skill-mirrors.sh`: change source to `skills/`, rename script to `sync-skill-mirrors.sh` | `./scripts/sync-skill-mirrors.sh` and `--check` work correctly | 36.1 | cc:done |
| 36.3 | Update section [10/12] in `check-consistency.sh`: change `skills-v3/` references to `skills/` | `./scripts/ci/check-consistency.sh` passes all paths | 36.2 | cc:done |
| 36.4 | Update `validate-plugin-v3.sh`: `skills-v3/` → `skills/` | Script works correctly | 36.2 | cc:done |
| 36.5 | Remove `skills-v3` from roots in `generate-skill-manifest.sh` | Manifest generation scans only `skills/` | 36.2 | cc:done |
| 36.6 | Change source paths in `fix-symlinks.sh` to `skills/` | Windows-compatible repair logic treats `skills/` as source of truth | 36.2 | cc:done |
| 36.7 | Remove `skills-v3` references from `set-locale.sh` | Locale switching processes only `skills/` | 36.2 | cc:done |
| 36.8 | Bulk documentation update: update `skills-v3` references in README.md, README_ja.md, v3-architecture.md, CLAUDE.md, etc. | `grep -r 'skills-v3' --include='*.md'` returns 0 hits outside CHANGELOG | 36.1 | cc:done |
| 36.9 | Update `skills-v3` references in SKILL.md files (including all mirrors) | No `skills-v3` references in any SKILL.md | 36.8 | cc:done |
| 36.10 | Delete `skills-v3/` directory + old `sync-v3-skill-mirrors.sh` | `ls skills-v3/` shows it does not exist | 36.1-36.9 | cc:done |
| 36.11 | Update `skills-v3` references in test files | `grep -r 'skills-v3' tests/` returns 0 hits | 36.10 | cc:done |
| 36.12 | Integration verification: `check-consistency.sh` + `validate-plugin.sh` + `go build` + `go test` all pass | All CI-equivalent verifications pass (excluding known pre-existing issues) | 36.10, 36.11 | cc:done |

---


## Phase 36: Project Simplification — Dead Code Cleanup + Codex Restoration

Created: 2026-04-12
Purpose: Remove dead OpenCode scripts, pre-consolidation agent duplicates, and unwired scripts. Restore Codex integration with symlinked skills.

### Design Principles

- **Git is the archive**: Removed code is recoverable from git history
- **Keep Codex alive**: Codex scripts and integration stay; restore `codex/` directory with symlinks
- **Remove OpenCode**: OpenCode platform is fully retired
- **Zero functional regression**: Only remove files not wired into hooks or active workflows

### Phase 36.0: Codex Restoration (symlink-based)

Purpose: Restore `codex/` directory with symlinked skills so Codex CLI integration works again

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 36.0.1 | Create `codex/.codex/skills/` with symlinks to `../../skills/` for core harness skills | All symlinks resolve correctly | - | cc:done |
| 36.0.2 | Restore `codex/.codex/config.toml` (multi-agent config) from git history | File exists and is valid TOML | 36.0.1 | cc:done |
| 36.0.3 | Restore `codex/.codex/rules/harness.rules` from git history | File exists | 36.0.1 | cc:done |
| 36.0.4 | Create English `codex/AGENTS.md` and `codex/README.md` with correct `tim-hub/powerball-harness` URLs | Files exist, no stale URLs | 36.0.2 | cc:done |
| 36.0.5 | Restore `codex/.codexignore` | File exists | - | cc:done |

### Phase 36.1: OpenCode Script Removal

Purpose: Remove OpenCode platform scripts (fully retired, no restoration planned)

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 36.1.1 | Remove `scripts/build-opencode.js`, `scripts/validate-opencode.js`, `scripts/setup-opencode.sh`, `scripts/opencode-setup-local.sh` | Files do not exist | - | cc:done |
| 36.1.2 | Remove `.github/workflows/opencode-compat.yml` | File does not exist | 36.1.1 | cc:done |

### Phase 36.2: Pre-Consolidation Agent Removal (~2,251 lines)

Purpose: Remove old agents superseded by v3 consolidation (worker.md, reviewer.md, scaffolder.md already exist)

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 36.2.1 | Remove `agents/task-worker.md` (consolidated into `worker.md`) | File does not exist | - | cc:done |
| 36.2.2 | Remove `agents/code-reviewer.md` (consolidated into `reviewer.md`) | File does not exist | - | cc:done |
| 36.2.3 | Remove `agents/plan-analyst.md`, `agents/plan-critic.md` (consolidated into `reviewer.md`) | Files do not exist | - | cc:done |
| 36.2.4 | ~~Remove `agents/error-recovery.md`~~ — Restored: still actively referenced by `worker.md`, `team-composition.md`, and `workflows/` | File exists (kept) | - | cc:done |
| 36.2.5 | Remove `agents/project-analyzer.md`, `agents/project-state-updater.md`, `agents/project-scaffolder.md` (consolidated into `scaffolder.md`) | Files do not exist | - | cc:done |
| 36.2.6 | Remove `agents/codex-implementer.md` (uses Codex MCP which is deprecated; v3 worker handles Codex via companion script) | File does not exist | - | cc:done |

### Phase 36.3: Unwired Script Removal (~1,279+ lines)

Purpose: Remove scripts not referenced in hooks.json or any active hook handler

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 36.3.1 | Remove `scripts/pretooluse-guard.sh` (1,279 lines, replaced by `hooks/pre-tool.sh` + core TypeScript) | File does not exist | - | cc:done |
| 36.3.2 | Remove old stop scripts: `scripts/stop-check-pending.sh`, `scripts/stop-cleanup-check.sh`, `scripts/stop-plans-reminder.sh` (replaced by `hook-handlers/stop-session-evaluator`) | Files do not exist | - | cc:done |
| 36.3.3 | Remove `scripts/posttooluse-security-review.sh`, `scripts/posttooluse-tampering-detector.sh` (consolidated into core TypeScript + haiku agent hook) | Files do not exist | - | cc:done |
| 36.3.4 | Remove `scripts/permission-request.sh`, `scripts/skill-child-reminder.sh`, `scripts/sync-v3-skill-mirrors.sh` (unwired, referencing removed dirs) | Files do not exist | - | cc:done |

### Phase 36.4: Stale Reference Cleanup (rules + architecture)

Purpose: Remove deprecated rules, update stale references

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 36.4.1 | Remove `.claude/rules/command-editing.md` (self-labeled DEPRECATED since v2.17.0) | File does not exist | - | cc:done |
| 36.4.2 | Update `.claude/rules/v3-architecture.md` to reflect actual structure | v3-architecture.md matches reality | - | cc:done |

### Phase 36.5: Stale Reference Cleanup (skills, scripts, tests, docs)

Purpose: Fix references to removed files across the project. CHANGELOG.md is excluded (historical record).

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 36.5.1 | Clean `sync-v3-skill-mirrors.sh` references from `skills/harness-release/SKILL.md` (6 refs) and `skills/harness-setup/SKILL.md` (2 refs) | `grep -r sync-v3-skill-mirrors skills/` returns no results | 36.3.4 | cc:done |
| 36.5.2 | Clean `opencode/` references from `skills/harness-release/SKILL.md`, `skills/harness-setup/SKILL.md`, `scripts/i18n/set-locale.sh`, `scripts/generate-skill-manifest.sh`, `scripts/ci/check-consistency.sh` | `grep -rn 'opencode/' --include='*.sh' --include='*.js' --include='*.md' scripts/ skills/` returns only CHANGELOG.md hits | 36.1 | cc:done |
| 36.5.3 | Update `tests/test-codex-package.sh` to reflect symlink-based codex/ (remove opencode refs, update skill path checks) | Test passes against new codex/ structure | 36.0, 36.1 | cc:done |
| 36.5.4 | Clean `pretooluse-guard.sh` references from `scripts/sync-plugin-cache.sh`, `tests/test-commit-guard.sh` | `grep -rn pretooluse-guard scripts/ tests/` returns only CHANGELOG.md and core/ source comments | 36.3.1 | cc:done |
| 36.5.5 | Clean `stop-cleanup-check.sh`, `stop-plans-reminder.sh` references from `scripts/sync-plugin-cache.sh` | `grep -rn 'stop-cleanup-check\|stop-plans-reminder' scripts/` returns no results | 36.3.2 | cc:done |
| 36.5.6 | Update `docs/distribution-scope.md` to reflect current structure (no opencode, codex is symlinks) | Doc matches reality | 36.0, 36.1 | cc:done |
| 36.5.7 | Clean `opencode/` patterns from `.gitignore` | No opencode patterns in .gitignore | 36.1 | cc:done |
| 36.5.8 | Update `docs/CLAUDE_CODE_COMPATIBILITY.md` to remove opencode references | No opencode references | 36.1 | cc:done |
| 36.5.9 | Update `docs/plans/briefs-manifest.md` to remove opencode surface reference | No opencode references | 36.1 | cc:done |
| 36.5.10 | Remove dead link in `.claude/rules/skill-editing.md` to `command-editing.md` | No reference to command-editing.md | 36.4.1 | cc:done |

---


## Phase 35: Harness v4 — Go Zero-Based Rebuild

Created: 2026-04-05
Purpose: Consolidate 127 shell scripts + TypeScript core into a single Go binary, reduce hook response time to under 5ms, and eliminate the dual-file configuration management problem

Design details: [go/DESIGN.md](go/DESIGN.md)

### Design Principles

- CC plugin protocol compliance is the top priority. Maintain official formats for `plugin.json`, `hooks.json`, `settings.json`, `agents/*.md`, `skills/*/SKILL.md`
- `harness.toml` is the only file users need to edit. CC required files are auto-generated by `harness sync`
- Run E2E verification at each Phase; unmigrated scripts continue to work via existing shims
- `bin/harness` is resolved via CC plugin PATH

### Priority Matrix

| Priority | Phase | Content | Task Count | Depends |
|--------|-------|------|---------|------|
| **Required** | 35.0 | Protocol + guardrails (minimal MVP) | 4 | None |
| **Required** | 35.1 | SQLite state layer | 3 | 35.0 |
| **Required** | 35.2 | Unified configuration (harness.toml → CC file generation) | 4 | 35.1 |
| **Recommended** | 35.3 | Handler integration (incremental absorption of 127 scripts) | 5 | 35.2 |
| **Recommended** | 35.4 | Agent lifecycle state machine | 3 | 35.3 |
| **Recommended** | 35.5 | Skill validation + SKILL.md validator | 2 | 35.2 |
| **Optional** | 35.6 | Breezing concurrency (goroutine/worktree) | 3 | 35.4 |
| **Optional** | 35.7 | npm distribution + cross-compilation | 3 | 35.6 |

Total: **27 tasks**

### Completion Criteria (Definition of Done — Phase 35 overall)

Phase 35 is complete when **all** of the following criteria are met.

| # | Criterion | Verification | Required/Recommended |
|---|------|---------|----------|
| 1 | **Zero Node.js runtime dependency**: No `node`/`core/dist` references from Go hook. Allowlist: codex-companion.sh, unmigrated scripts/*.js | `grep -rE "node\|core/dist" hooks/` + allowlist cross-check | Required |
| 2 | **Hook response p99 < 10ms**: Heaviest PreToolUse path (with SQLite lookup) | `hyperfine` 100 runs (empty DB / bloated DB / contended DB) | Required |
| 3 | **Guardrail parity**: All test cases for R01-R13 PASS in Go tests | `go test ./internal/guard/...` | Required |
| 4 | **Official protocol compliance**: Documented fields in Protocol Truth Table (SPEC.md §2) work correctly. experimental fields unimplemented, unknown fields ignored | E2E tests verifying each field | Required |
| 5 | **Unified configuration**: `harness.toml` → `harness sync` → CC operates correctly | `harness sync && validate-plugin.sh` PASS | Required |
| 6 | **Dual hooks.json eliminated**: `harness sync` auto-syncs hooks.json + .claude-plugin/hooks.json | `check-consistency.sh` PASS | Required |
| 7 | **Existing skills/agents work**: All 30+ skills + 3 agents work correctly in Go environment | E2E via hook event matrix (all hook events covered) | Required |
| 8 | **State migration consistency**: Old state.db → new path migration is reversible. export/import + rollback toggle | `harness doctor --migration` PASS | Required |
| 9 | **Script migration rate 80%+**: 100+ of 127 scripts absorbed into Go subcommands | `harness doctor` migration report | Recommended |
| 10 | **Cross-platform build**: darwin-arm64, darwin-amd64, linux-amd64 | CI cross-compilation success | Recommended |
| 11 | **Binary size < 10MB**: stripped + optimized | `ls -lh bin/harness` | Recommended |

**Minimum completion condition**: If all 8 required criteria (1-8) are met, v4.0.0 can be released. Criteria 9-11 can be achieved in v4.1.0 and later.

Detailed specification: [go/SPEC.md](go/SPEC.md)

---

### Phase 35.0: Protocol + Guardrails [P0]

Purpose: Replace the `pre-tool.sh` → `node core/dist/index.js` call chain with the single binary `harness hook pre-tool`, achieving p99 < 10ms

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.0.1 | Go module initialization + full compliance type definitions for official stdin/stdout JSON schema in `pkg/protocol/types.go` | `transcript_path`, `permission_mode`, `hook_event_name`, `defer`, `updatedInput`, `additionalContext` included in types | - | cc:done |
| 35.0.2 | Port all R01-R13 rules 1:1 into `internal/guard/rules.go` + stdin parser implementation in `internal/hook/codec.go` | All 58 tests PASS | 35.0.1 | cc:done |
| 35.0.3 | `cmd/harness/main.go` CLI + PreToolUse/PostToolUse/PermissionRequest handlers + `bin/harness` build | E2E 8 scenarios PASS, p99 5ms | 35.0.2 | cc:done |
| 35.0.4 | Rewrite 3 hook shims (`pre-tool.sh`, `post-tool.sh`, `permission.sh`) to call Go binary directly | Node.js fallback removed, clear error when binary not found | 35.0.3 | cc:done |

---

### Phase 35.1: SQLite State Layer [P0]

Purpose: Go port of `core/src/state/` + leveraging `${CLAUDE_PLUGIN_DATA}` persistent storage

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.1.1 | `internal/state/schema.go` — Existing DDL (sessions, signals, task_failures, work_states, schema_meta) + add assumptions table | Schema initialization with migrations works | 35.0.3 | cc:done |
| 35.1.2 | `internal/state/store.go` — Go port of HarnessStore (WAL mode, busy timeout 5s) | No deadlocks with 3 goroutines running parallel INSERT/SELECT | 35.1.1 | cc:done |
| 35.1.3 | Integrate SQLite work_states lookup into BuildContext in `pre_tool.go` | Retrieve codexMode/workMode from DB with session_id in input | 35.1.2 | cc:done |

---

### Phase 35.2: Unified Configuration [P0]

Purpose: Root-fix the dual hooks.json sync problem through auto-generation of `harness.toml` → CC required files

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.2.1 | `pkg/config/toml.go` — harness.toml parser ([project], [safety], [agent], [env], [hooks], [telemetry] sections). Mapping Table conforms to SPEC.md §5 | TOML parse + validation + unsupported key rejection | 35.0.3 | cc:done |
| 35.2.2 | `harness sync` — auto-generate harness.toml → hooks.json + settings.json (permissions/sandbox/env/agent) + plugin.json | Generated files functionally equivalent to current files | 35.2.1 | cc:done |
| 35.2.3 | `harness init` subcommand — project initialization (generate harness.toml template) | `harness init && harness sync` works in new project | 35.2.2 | cc:done |
| 35.2.4 | Integrate dual hooks.json sync script (`sync-plugin-cache.sh`) into `harness sync` | `sync-plugin-cache.sh` becomes a wrapper for `harness sync` | 35.2.2 | cc:done |

---

### Phase 35.3: Handler Integration [P1]

Purpose: Incrementally absorb 127 scripts into Go subcommands by category

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.3.1 | Port 5 hook-handlers to Go (session-env, post-tool-failure, post-compact, notification, permission-denied) | 25 tests PASS, symlink check included | 35.2.2 | cc:done |
| 35.3.2 | Port 4 session-* scripts to Go (init, cleanup, monitor, summary) | 30 tests PASS | 35.3.1 | cc:done |
| 35.3.3 | codex-companion.sh is **excluded** from Go integration (SPEC.md decision). Maintain shell wrapper | No change (policy documented in SPEC.md) | - | cc:done |
| 35.3.4 | Port ci-status-checker + evidence collector to Go | 15 tests PASS | 35.3.2 | cc:done |
| 35.3.5 | `harness doctor` + `--migration` list hook migration status. Mixed-mode warnings, hooks.json divergence detection | `harness doctor` 11 tests PASS | 35.3.1 | cc:done |

---

### Phase 35.4: Agent Lifecycle [P1]

Purpose: SPAWNING→RUNNING→REVIEWING→APPROVED→COMMITTED state machine + 4-stage recovery

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.4.1 | `internal/lifecycle/state.go` — State machine definition + transition rules. Full states from SPEC.md §8 (including FAILED/CANCELLED/STALE/RECOVERING/ABORTED) | Invalid transitions prevented at type level, all anomalous states defined | 35.3.2 | cc:done |
| 35.4.2 | `internal/lifecycle/recovery.go` — 4-stage recovery (self-repair → peer-repair → commander intervention → stop) | Trigger conditions and behavior for each stage defined | 35.4.1 | cc:done |
| 35.4.3 | Integration with SubagentStart/Stop hooks | State transitions persisted to SQLite, displayed via `harness status` | 35.4.2, 35.1.2 | cc:done |

---

### Phase 35.5: Skill Validation [P1]

Purpose: Type-safe validation of SKILL.md frontmatter

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.5.1 | `harness validate skills` — Match all SKILL.md frontmatter against official schema. Regex-based (no external YAML dependency) | name, description required fields + optional type validation | 35.2.1 | cc:done |
| 35.5.2 | `harness validate agents` — Validate frontmatter in agents/*.md (tools, disallowedTools, isolation, background, maxTurns) | 22 tests PASS | 35.5.1 | cc:done |

---

### Phase 35.6: Breezing Concurrency [P2]

Purpose: Safe parallel task execution via goroutines + worktrees

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.6.1 | `internal/breezing/orchestrator.go` — Worker/Reviewer goroutine management | Max concurrency control, graceful shutdown | 35.4.3 | cc:done |
| 35.6.2 | Go implementation of automatic worktree creation/cleanup | Integrates with CC WorktreeCreate/Remove hooks | 35.6.1 | cc:done |
| 35.6.3 | Automatic task dependency resolution + file-lock claiming | Auto-unblock of dependent tasks works | 35.6.2 | cc:done |

---

### Phase 35.7: npm Distribution + Cross-Compilation [P2]

Purpose: Go binary distribution using `bin/` directory

| Task | Description | DoD | Depends | Status |
|------|------|-----|---------|--------|
| 35.7.1 | Cross-compilation (darwin-arm64/amd64, linux-amd64). CGO_ENABLED=0 + modernc.org/sqlite | All 3 binaries at 6.6-6.8MB | 35.3.1 | cc:done |
| 35.7.2 | npm package configuration + postinstall places platform-specific binary | `npm install` places `bin/harness` in PATH | 35.7.1 | cc:done |
| 35.7.3 | Migration notice to old package + GitHub Release automation | Release workflow includes Go binary | 35.7.2 | cc:done |

---


## Phase 35: Repository Rebrand + Structure Consolidation

Created: 2026-04-12
Purpose: Rebrand repository to `tim-hub/powerball-harness`, eliminate redundant v3 directories, and unify all skills/agents under a single directory structure

### Phase 35.0: Repository URL Rebrand

Purpose: Update all references from old repository URLs to `tim-hub/powerball-harness`

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 35.0.1 | Replace all `Chachamaru127/claude-code-harness` and `tim-hub/claude-code-harness` URLs with `tim-hub/powerball-harness` across README, marketplace.json, CONTRIBUTING, install scripts, CI scripts, social posts, and benchmark docs | `grep -r 'Chachamaru127\|tim-hub/claude-code-harness'` returns only intentional attribution in README Origin section | - | cc:done |
| 35.0.2 | Update marketplace.json owner, author, homepage, and repository fields | `marketplace.json` owner is `tim-hub`, all URLs point to `tim-hub/powerball-harness` | 35.0.1 | cc:done |
| 35.0.3 | Add Origin section to README crediting the original upstream repository | README bottom contains attribution link to `Chachamaru127/claude-code-harness` | 35.0.1 | cc:done |

### Phase 35.1: v3 Directory Consolidation

Purpose: Eliminate redundant `skills-v3/` and `agents-v3/` directories by merging into `skills/` and `agents/`

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 35.1.1 | Copy `agents-v3/` files (worker.md, reviewer.md, scaffolder.md, team-composition.md) into `agents/` | All 4 files exist in `agents/` | - | cc:done |
| 35.1.2 | Remove `skills-v3/` directory (core skills are duplicates of `skills/`, extensions are symlinks back to `skills/`) | `skills-v3/` does not exist | 35.1.1 | cc:done |
| 35.1.3 | Remove `agents-v3/` directory | `agents-v3/` does not exist | 35.1.1 | cc:done |
| 35.1.4 | Update all `skills-v3` and `agents-v3` references across docs, scripts, rules, CI tests, and CLAUDE.md to point to `skills/` and `agents/` | `grep -r 'skills-v3\|agents-v3'` returns only CHANGELOG.md and Plans.md (historical records) | 35.1.2 | cc:done |

### Phase 35.2: Skill Configuration

Purpose: Improve planning quality by routing to the best model

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 35.2.1 | Add `model: opus` to `harness-plan` SKILL.md frontmatter | `harness-plan/SKILL.md` contains `model: opus` | - | cc:done |

---


## Upstream Phase History (v4.0.x — merged from upstream/main)

> These phases were completed in upstream/main (Hokage branch) and are included here as historical record.

