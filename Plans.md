# Powerball Harness — Plans.md

Last archive: 2026-04-15 (Phase 35–48 → `.claude/memory/archive/Plans-2026-04-15-phase35-48.md`)
Last release: v4.3.0 on 2026-04-15 (Phase 50+51)

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
| 58.11 | Run full validation suite: `validate-plugin.sh` + `check-consistency.sh` + `check-residue.sh` + `harness validate all` | All pass with 0 failures | 56.1-56.5, 57.1-57.9, 58.1-58.10 | cc:TODO |
| 58.12 | Record all changes under `[Unreleased]` in CHANGELOG.md in Before/After format | CHANGELOG entry added | 58.11 | cc:TODO |

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

