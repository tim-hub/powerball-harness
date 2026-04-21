# Phase 81 Migration Manifest
# Audit & Catalog: harness-release → release-this

**Generated**: 2026-04-21
**Auditor**: Harness Worker (Task 81.1)
**Scope**: `harness/skills/harness-release/SKILL.md` + all 5 scripts

---

## Classification Legend

| Tag | Meaning |
|-----|---------|
| **GENERIC** | Works for any project with VERSION, CHANGELOG.md, git repo, GitHub |
| **PLUGIN-SPECIFIC** | References harness-specific constructs: marketplace.json, harness.toml, codex symlinks, check-consistency.sh, /start-task, ccp-*, skills/ directory structure, SVG EN/JA sync, or claude-code-harness-specific paths |
| **MIXED** | Contains both generic logic and plugin-specific checks; must be split on migration |

---

## Section 1: SKILL.md Phase-by-Phase Classification

| Phase | Title | Classification | Rationale |
|-------|-------|---------------|-----------|
| Frontmatter | `name`, `description`, `when_to_use`, `allowed-tools`, `argument-hint` | GENERIC | These fields are standard skill frontmatter applicable to any release skill |
| Quick Reference table | Subcommand table (patch/minor/major/--dry-run/--complete) | GENERIC | SemVer bump subcommands work for any project |
| Release-only policy | Normal PRs vs release commits rule | GENERIC | VERSION / CHANGELOG discipline is project-agnostic |
| Branch Policy | Solo vs collaborative push rules | GENERIC | git branching policy applies to any repo |
| Version Determination Criteria | SemVer flowchart + table | GENERIC | Standard SemVer semantics, no harness specifics |
| Version Distribution | "Only 2 files subject to version management: VERSION + `.claude-plugin/marketplace.json`" | **PLUGIN-SPECIFIC** | References `.claude-plugin/marketplace.json` which is a harness plugin manifest; a generic project would reference its own manifest (e.g. `package.json`) |
| Distribution Surfaces / Mirror Sync | `skills/` SSOT table, Codex CLI symlinks `codex/.codex/skills/` | **PLUGIN-SPECIFIC** | References Codex CLI symlink structure specific to claude-code-harness |
| **Phase 0** | Pre-flight Checks | **MIXED** | See detail below |
| Phase 0 — tool check | `gh` / `jq` check | GENERIC | Applies to any repo using GitHub Releases |
| Phase 0 — preflight script | `release-preflight.sh` | GENERIC (script itself is MIXED — see Section 2) | Invocation is generic; internal checks are MIXED |
| Phase 0 — plugin structure | `bash tests/validate-plugin.sh` | **PLUGIN-SPECIFIC** | `validate-plugin.sh` is a claude-code-harness-specific test harness; generic project would skip or substitute |
| Phase 0 — consistency | `bash check-consistency.sh` | **PLUGIN-SPECIFIC** | `check-consistency.sh` is entirely plugin-specific (see Section 2) |
| Phase 0 — codex symlinks | `ls -la codex/.codex/skills/` | **PLUGIN-SPECIFIC** | Codex symlink structure is harness-specific |
| Phase 0 — env vars table | `HARNESS_RELEASE_PLUGIN_ROOT`, `HARNESS_RELEASE_HEALTHCHECK_CMD`, `HARNESS_RELEASE_CI_STATUS_CMD` | GENERIC | These env vars are generic override hooks applicable to any project |
| **Phase 1** | Get Current Version | GENERIC | `cat VERSION` works for any project with a VERSION file |
| **Phase 2** | Calculate New Version | **MIXED** | `sync-version.sh bump` is MIXED (see Section 2); `sync` syncs to `marketplace.json` which is PLUGIN-SPECIFIC; manual minor/major bump logic is GENERIC |
| Phase 2 — sync-version.sh sync | Applies VERSION to `.claude-plugin/marketplace.json` | **PLUGIN-SPECIFIC** | Marketplace.json is harness-specific; generic project syncs to their own manifest |
| **Phase 3** | CHANGELOG Update | GENERIC | `[Unreleased]` → versioned entry pattern is standard Keep a Changelog format |
| **Phase 4** | Update Version Files | **MIXED** | VERSION update is GENERIC; `sync-version.sh sync` to marketplace.json is PLUGIN-SPECIFIC |
| **Phase 5** | Verify Codex Symlinks | **PLUGIN-SPECIFIC** | Codex CLI symlink check is harness-specific; a generic project would omit this phase |
| **Phase 6** | Commit & Tag | **MIXED** | `git commit` / `git tag` are GENERIC; the `git add harness/VERSION harness/harness.toml CHANGELOG.md` line is PLUGIN-SPECIFIC (hardcoded harness paths) |
| **Phase 7** | Push | GENERIC | `git push origin {main/master} --tags` applies to any repo |
| Phase 7 — safety net note | `.github/workflows/release.yml` tag-push detection | GENERIC | Any project can implement this CI safety net pattern |
| **Phase 8** | Create GitHub Release | GENERIC | `gh release create` with Before/After template is applicable to any GitHub project |
| Phase 8 — validate-release-notes.sh | Post-creation validation | GENERIC | The validation rules (heading, Before/After, bold summary) are content-agnostic |
| **Phase 9** | Release Completion Marking | GENERIC | Empty commit marker pattern works for any project |
| `--dry-run` Mode | Steps 1-5 preview | GENERIC | Dry-run logic applies to any project |
| `--complete` Mode | Phase 9 only | GENERIC | Completion marker applies to any project |
| Regression Checklist | Full checklist | **MIXED** | Several items are GENERIC (`git tag --sort`, `gh release`, git clean); several are PLUGIN-SPECIFIC (`validate-plugin.sh`, `check-consistency.sh`, `R01-R13 in go/internal/guardrail/rules.go`, `check-residue.py`) |
| CI Safety Net | Workflow pattern | GENERIC | Any project can use this tag-based CI safety net |
| PM Handoff | Completion report template | GENERIC | Generic project handoff pattern |
| Prohibited Actions | Tag immutability, minor bump limits, force push prohibition | GENERIC | Standard release hygiene rules |
| Related Skills | `harness-review`, `harness-work`, `harness-plan`, `harness-setup` | **PLUGIN-SPECIFIC** | These are harness-specific skill references |
| Related Rules | `.claude/rules/versioning.md`, `.claude/rules/github-release.md`, `.claude/rules/cc-update-policy.md` | **PLUGIN-SPECIFIC** | These reference harness-internal rule files |

---

## Section 2: Script-by-Script Classification

### 2.1 `sync-version.sh`

**Overall classification**: **MIXED** (mostly PLUGIN-SPECIFIC)

| Check / Function | Classification | Rationale |
|-----------------|---------------|-----------|
| VERSION file path (`harness/VERSION`) | **PLUGIN-SPECIFIC** | Hardcoded path under `harness/`; generic project would use `./VERSION` at repo root |
| `HARNESS_TOML` (`harness/harness.toml`) | **PLUGIN-SPECIFIC** | `harness.toml` is a harness-specific runtime manifest; generic projects would target their own manifest (e.g. `package.json`, `pyproject.toml`, `Cargo.toml`) |
| `check_version()` — VERSION vs harness.toml | **PLUGIN-SPECIFIC** | Checks harness.toml specifically |
| `sync_version()` — updates harness.toml | **PLUGIN-SPECIFIC** | Updates harness.toml specifically |
| `update_changelog_compare_links()` | **GENERIC** | Rewrites `[Unreleased]` compare link in CHANGELOG.md using git tag URLs; works for any GitHub project |
| `bump_version()` — patch/minor/major math | **GENERIC** | SemVer bump arithmetic is project-agnostic |
| `bump_version()` — calls `sync_version()` | **PLUGIN-SPECIFIC** | Because `sync_version()` itself targets harness.toml |
| CLI dispatch (`check`/`sync`/`bump`) | **GENERIC** (structure) | The command interface pattern is reusable |
| Comment "`.claude-plugin/marketplace.json` no longer carries a version field" | **PLUGIN-SPECIFIC** | Refers to a past harness-specific artifact |

**Summary**: The core bump/compare-link logic is generic. The sync targets (harness.toml) are plugin-specific. A generic `release-this` version would target `package.json` or a configurable manifest.

---

### 2.2 `validate-release-notes.sh`

**Overall classification**: **GENERIC** with one PLUGIN-SPECIFIC reference

| Check / Section | Classification | Rationale |
|----------------|---------------|-----------|
| `gh release view` — fetch release notes | GENERIC | Any GitHub project uses `gh` |
| Check 1: `## What's Changed` heading | GENERIC | Standard section heading |
| Check 2: `Before.*After` table | GENERIC | Before/After is a harness convention but expressed as a content check that works for any project following this convention |
| Check 3: `Generated with [Claude Code]` footer | GENERIC | Optional footer check, not harness-specific |
| Check 4: Section presence (Added/Changed/Fixed/Security) | GENERIC | Standard Keep a Changelog sections |
| Check 6: Bold summary in first 10 lines | GENERIC | Applies to any GitHub Release using this style guide |
| Error reference: `.claude/rules/github-release.md` | **PLUGIN-SPECIFIC** | References an internal harness rule file path; a generic project would reference its own style guide |

**Summary**: Nearly entirely GENERIC. The only plugin-specific element is the error message pointing to a harness-internal rule file. Easy to parameterize.

---

### 2.3 `release-preflight.sh`

**Overall classification**: **MIXED** (generic structure, PLUGIN-SPECIFIC sub-checks)

| Check / Function | Classification | Rationale |
|----------------|---------------|-----------|
| `PLUGIN_ROOT` derivation (3 levels up from script) | **PLUGIN-SPECIFIC** | Assumes script lives at `harness/skills/harness-release/scripts/`; generic version would use configurable root |
| `HARNESS_RELEASE_PLUGIN_ROOT` env override | GENERIC | Override mechanism is reusable |
| `check_git_clean()` | **GENERIC** | `git status --porcelain` clean check applies to any repo |
| `check_changelog()` — `[Unreleased]` section | **GENERIC** | Standard Keep a Changelog check |
| `check_env_and_healthcheck()` — `.env` / `.env.example` parity | **GENERIC** | Env var parity check applies to any project with `.env.example` |
| `check_env_and_healthcheck()` — `npm run healthcheck/preflight` | **GENERIC** | npm script detection is project-agnostic |
| `check_env_and_healthcheck()` — `HARNESS_RELEASE_HEALTHCHECK_CMD` | **GENERIC** | Configurable healthcheck command is reusable |
| `check_runtime_residuals()` — scans `agents/`, `core/`, `hooks/`, `scripts/` | **PLUGIN-SPECIFIC** | The hardcoded directory list (`agents/`, `core/`, `hooks/`, `scripts/`) reflects the claude-code-harness plugin layout; a generic project would need to configure its shipped directories |
| `check_runtime_residuals()` — patterns (mockData, dummy, localhost, TODO, FIXME, test.skip) | **GENERIC** | The residual patterns themselves are generic best practices |
| `check_runtime_residuals()` — `HARNESS_RELEASE_RESIDUAL_PATTERNS` override | **GENERIC** | Configurable pattern override is reusable |
| `check_sprint_contract_schema()` — reads `.claude/state/contracts/*.sprint-contract.json` | **PLUGIN-SPECIFIC** | Sprint contracts are a harness-specific state format; generic projects would not have this |
| `check_sprint_contract_schema()` — validates `reviewer_profile`, `loop_pacing`, `browser_verdict` | **PLUGIN-SPECIFIC** | These schema fields are harness-specific |
| `check_ci_status()` — `gh run list` + pass/fail | **GENERIC** | GitHub Actions CI status check applies to any project using GH Actions |
| `check_ci_status()` — `HARNESS_RELEASE_CI_STATUS_CMD` | **GENERIC** | Configurable CI command is reusable |
| Summary counters (PASS/WARN/FAIL) | **GENERIC** | Output format pattern is reusable |

**Summary**: The generic checks (git clean, changelog, env parity, CI status) form the reusable core. The plugin-specific items are: hardcoded directory list in residuals scan, and sprint-contract schema validation.

---

### 2.4 `check-consistency.sh`

**Overall classification**: **PLUGIN-SPECIFIC**

| Check | Classification | Rationale |
|-------|---------------|-----------|
| **[1/13] Template file existence** | **PLUGIN-SPECIFIC** | Checks for `harness/templates/AGENTS.md.template`, cursor commands, claude settings templates — all harness-specific paths |
| **[2/13] Command ↔ Skill consistency** | **PLUGIN-SPECIFIC** | Validates `commands/` → `skills/` migration, which is harness-specific |
| **[3/13] Version number consistency** | **PLUGIN-SPECIFIC** | Checks `harness/VERSION` vs `harness/harness.toml`; includes hardcoded GitHub badge URLs for `tim-hub/powerball-harness` |
| **[4/13] Expected skill definition file structure** | **PLUGIN-SPECIFIC** | Checks for `harness/skills/harness-setup/SKILL.md` — harness skill layout |
| **[5/13] Hooks configuration consistency** | **PLUGIN-SPECIFIC** | Validates `harness/hooks/hooks.json` script references — harness hooks system |
| **[6/13] /start-task deprecation regression** | **PLUGIN-SPECIFIC** | `/start-task` was a harness-specific command that was deprecated |
| **[7/13] docs/ normalization regression** | **PLUGIN-SPECIFIC** | Checks for `proposal.md`, `technical-spec.md`, `priority_matrix.md` without `docs/` prefix — harness-specific doc layout enforcement |
| **[8/13] bypassPermissions assumption** | **PLUGIN-SPECIFIC** | Validates `harness/templates/claude/settings.security.json.template` and `settings.local.json.template` — harness template format |
| **[9/13] ccp-* skill deprecation regression** | **PLUGIN-SPECIFIC** | `ccp-*` skills were a harness-specific naming convention that was deprecated |
| **[10/13] Template existence check** | **PLUGIN-SPECIFIC** | Checks for codex, opencode, and harness-specific templates and setup scripts |
| **[11/13] CHANGELOG format validation** | **MIXED** | The `[Unreleased]` and ISO 8601 date checks are GENERIC; checking `CHANGELOG_ja.md` is PLUGIN-SPECIFIC (harness maintains bilingual CHANGELOG) |
| **[12/13] README claim drift** | **PLUGIN-SPECIFIC** | Checks for specific badge URLs (`tim-hub/powerball-harness`), `Go-native guardrail engine` claim, `distribution-scope.md`, `benchmark-rubric.md`, `positioning-notes.md` — all harness-specific documents |
| **[13/13] EN/JA visual sync** | **PLUGIN-SPECIFIC** | Checks SVG files in `docs/assets/readme-visuals-en/` and `readme-visuals-ja/` — harness bilingual documentation |

**Summary**: This script is entirely PLUGIN-SPECIFIC. It encodes deep knowledge of the harness repository structure, its specific migration history (commands→skills, /start-task, ccp-*), and its bilingual documentation system. It should NOT move to `release-this`.

---

### 2.5 `check-residue.py`

**Overall classification**: **PLUGIN-SPECIFIC**

| Component | Classification | Rationale |
|-----------|---------------|-----------|
| REPO_ROOT derivation (4 levels up) | **PLUGIN-SPECIFIC** | Hardcoded to 4-level depth specific to `harness/skills/harness-release/scripts/` layout |
| YAML_PATH (`.claude/rules/deleted-concepts.yaml`) | **PLUGIN-SPECIFIC** | This config file catalogs harness-specific deleted concepts (TypeScript guardrail engine, `core/`, Harness v3 terminology, etc.) |
| `_load_yaml()` fallback parser | GENERIC (logic) | The parser itself is reusable, but it's purpose-built for this harness-specific YAML |
| `grep_files()` / `grep_line_numbers()` | **GENERIC** | Generic grep wrappers, reusable |
| `grep_h1_v3_files()` | **PLUGIN-SPECIFIC** | Scans for `(v3)` in H1 titles — a harness migration artifact from the v3→v4 TypeScript→Go migration |
| `deleted_paths` scan | **PLUGIN-SPECIFIC** | Data comes from `deleted-concepts.yaml` which catalogs harness-specific deleted paths |
| `deleted_concepts` scan | **PLUGIN-SPECIFIC** | Data comes from `deleted-concepts.yaml`; concept terms are harness-specific (e.g. "TypeScript guardrail engine") |
| Default allowlist (CHANGELOG.md, `.claude/memory/archive/`, etc.) | **PLUGIN-SPECIFIC** | Allowlist entries are harness-specific paths |
| H1 (v3) suffix scan | **PLUGIN-SPECIFIC** | v3 suffix was a harness-specific versioning pattern from the pre-Go era |
| Summary / exit code contract | **GENERIC** | Exit 0 = clean, exit 1 = residue found is a reusable convention |

**Summary**: The scanning mechanism is reusable, but the entire configuration (what to scan for, allowlists, H1 (v3) suffix) is harness-specific. A generic project would need its own `deleted-concepts.yaml` with different content. The script cannot be directly reused without its config file.

---

## Section 3: Migration Decision Table

What stays in `harness/skills/harness-release/` vs what moves to `.claude/skills/release-this/`

| Item | Stays in `harness-release` | Moves to `release-this` | Notes |
|------|---------------------------|------------------------|-------|
| **SKILL.md phases** | | | |
| Phase 0: tool checks (gh/jq) | — | ✅ | Verbatim copy |
| Phase 0: `release-preflight.sh` invocation | — | ✅ | Invoke a `release-this` version of preflight |
| Phase 0: `validate-plugin.sh` | ✅ KEEP (harness CI) | — | Harness-only; omit from release-this |
| Phase 0: `check-consistency.sh` | ✅ KEEP (harness CI) | — | 100% plugin-specific |
| Phase 0: Codex symlink check | ✅ KEEP | — | Harness distribution-specific |
| Phase 0: env var overrides | — | ✅ | Generic override pattern; copy |
| Phase 1: Get Current Version | — | ✅ | Verbatim `cat VERSION` |
| Phase 2: bump / sync VERSION | — | ✅ (adapted) | Remove harness.toml sync; add configurable manifest target |
| Phase 3: CHANGELOG Update | — | ✅ | Verbatim |
| Phase 4: Update Version Files | — | ✅ (adapted) | Only VERSION; manifest sync via configurable hook |
| Phase 5: Codex symlink verify | ✅ KEEP | — | Harness-specific |
| Phase 6: Commit & Tag | — | ✅ (adapted) | Remove hardcoded `harness/VERSION harness/harness.toml`; use `VERSION CHANGELOG.md` |
| Phase 7: Push | — | ✅ | Verbatim |
| Phase 8: GitHub Release | — | ✅ | Verbatim |
| Phase 9: Completion marker | — | ✅ | Verbatim |
| `--dry-run` mode | — | ✅ | Verbatim |
| `--complete` mode | — | ✅ | Verbatim |
| Regression checklist (generic items) | — | ✅ (subset) | git tag continuity, CI status, gh release check |
| Regression checklist (plugin items) | ✅ KEEP | — | validate-plugin.sh, check-consistency.sh, go guardrails |
| PM Handoff template | — | ✅ | Verbatim |
| Prohibited Actions | — | ✅ | Verbatim |
| Related Skills (harness-specific) | ✅ KEEP | — | Harness skill cross-references |
| Related Rules (harness-specific) | ✅ KEEP | — | Harness rule file references |
| **Scripts** | | | |
| `sync-version.sh` — bump_version() | — | ✅ (adapted) | Port SemVer arithmetic; remove harness.toml sync |
| `sync-version.sh` — update_changelog_compare_links() | — | ✅ | Verbatim; generic |
| `sync-version.sh` — sync to harness.toml | ✅ KEEP | ✗ OMIT | Replace with configurable manifest-sync hook |
| `sync-version.sh` — check harness.toml | ✅ KEEP | ✗ OMIT | Replace with configurable check |
| `validate-release-notes.sh` | — | ✅ (minor adaptation) | Change error reference from `.claude/rules/github-release.md` to a configurable path |
| `release-preflight.sh` — check_git_clean() | — | ✅ | Verbatim |
| `release-preflight.sh` — check_changelog() | — | ✅ | Verbatim |
| `release-preflight.sh` — check_env_and_healthcheck() | — | ✅ | Verbatim |
| `release-preflight.sh` — check_runtime_residuals() (patterns + override) | — | ✅ (adapted) | Parameterize directory list; default to empty or auto-detect |
| `release-preflight.sh` — check_sprint_contract_schema() | ✅ KEEP | ✗ OMIT | Harness-specific sprint contract format |
| `release-preflight.sh` — check_ci_status() | — | ✅ | Verbatim |
| `check-consistency.sh` — [1/13] through [13/13] | ✅ KEEP (all) | ✗ OMIT (all) | 100% plugin-specific; no portable content |
| `check-residue.py` — grep utilities | — | ✅ (optional) | Portable scanner; needs generic config |
| `check-residue.py` — YAML config load | — | ✅ (optional) | Portable; needs user-supplied `deleted-concepts.yaml` |
| `check-residue.py` — harness allowlist defaults | ✅ KEEP | ✗ ADAPT | Generic version omits harness-specific allowlist entries |
| `check-residue.py` — H1 (v3) suffix scan | ✅ KEEP | ✗ OMIT | Harness migration artifact |
| `check-residue.py` — YAML_PATH (`.claude/rules/deleted-concepts.yaml`) | ✅ KEEP | ✗ ADAPT | Generic version uses project's own `deleted-concepts.yaml` if it exists, else skip |

---

## Section 4: Edge Cases and Items Requiring Human Decision

### 4.1 `check-residue.py` portability

**Question**: Should `release-this` include a variant of `check-residue.py`?

**Analysis**: The scanning mechanism is genuinely useful for any project that undergoes major migrations. However, its value depends entirely on the user populating `.claude/rules/deleted-concepts.yaml` with their own deleted concepts. Without that file, it skips cleanly (exits 2 if the yaml is absent — currently an error).

**Decision needed**: Should `release-this` include `check-residue.py` as an optional scanner that gracefully skips when `deleted-concepts.yaml` is absent? Or omit entirely and let projects adopt it independently?

**Recommendation**: Include with a graceful-skip (`exit 0` when YAML absent, with a hint). Change the `REPO_ROOT` derivation to use `git rev-parse --show-toplevel` instead of counting directory levels.

---

### 4.2 `sync-version.sh` manifest sync target

**Question**: What should `release-this`'s `sync-version.sh` sync to, instead of `harness.toml`?

**Analysis**: Different projects have different version manifest files: `package.json` (Node.js), `pyproject.toml` (Python), `Cargo.toml` (Rust), `go.mod` (Go), or nothing. The harness.toml sync is genuinely useful but platform-specific.

**Decision needed**: Should `release-this` use a configurable `RELEASE_MANIFEST_SYNC_CMD` env var that users override for their platform? Or attempt auto-detection (`if [ -f package.json ]; then ... elif [ -f pyproject.toml ]; then ...`)?

**Recommendation**: Use `RELEASE_MANIFEST_SYNC_CMD` env override with a default of "no-op with warning." Auto-detection is fragile and would be misleading for projects with multiple manifest files.

---

### 4.3 `check_runtime_residuals()` directory list

**Question**: In `release-preflight.sh`, the residuals scan hardcodes `agents/`, `core/`, `hooks/`, `scripts/` as the directories to scan. What should `release-this` scan?

**Analysis**: These directories are the harness plugin's "shipped surfaces." A generic project's shipped surfaces might be `src/`, `lib/`, `bin/`, or simply the entire repo.

**Decision needed**: Should `release-this` use `RELEASE_SHIPPED_DIRS` env var (space-separated list), or scan the entire git-tracked file set by default?

**Recommendation**: Default to scanning all git-tracked files (i.e., remove the directory filter) and let `HARNESS_RELEASE_RESIDUAL_PATTERNS` still be overridable. This is the safest generic behavior.

---

### 4.4 Phase 6 commit staging

**Question**: `Phase 6` in SKILL.md hardcodes `git add harness/VERSION harness/harness.toml CHANGELOG.md`. What should `release-this` stage?

**Analysis**: A generic project would stage `VERSION CHANGELOG.md` plus whatever manifest was synced (determined by `RELEASE_MANIFEST_SYNC_CMD`).

**Decision needed**: Should `release-this` document `git add VERSION CHANGELOG.md` as the default, with a note to add the synced manifest file?

**Recommendation**: Yes. Document `git add VERSION CHANGELOG.md` as the baseline and instruct users to add their manifest file. This is safe and honest.

---

### 4.5 `check-consistency.sh` [11/13] CHANGELOG bilingual check

**Question**: Check [11/13] validates both `CHANGELOG.md` and `CHANGELOG_ja.md`. The `CHANGELOG_ja.md` check is plugin-specific, but the format validation logic is generic.

**Decision needed**: Should `release-this` include a generic CHANGELOG format validator (ISO 8601 dates, `[Unreleased]` section)?

**Recommendation**: Yes — extract the generic CHANGELOG format checks into `release-preflight.sh` for `release-this`. This avoids duplicating `check-consistency.sh`. The bilingual `CHANGELOG_ja.md` check stays in harness-release only.

---

### 4.6 Version Distribution section references marketplace.json

The SKILL.md `Version Distribution` section says "Only 2 files subject to version management: VERSION + `.claude-plugin/marketplace.json`". This is factually incorrect for a generic project.

**Decision needed**: Should `release-this` omit this section entirely, or replace it with a generic "Version Distribution" section listing `VERSION` + "your project manifest" with examples?

**Recommendation**: Replace with a generic section. The concept (canonicalize version to as few files as possible) is sound; only the specific file names are wrong.

---

## Appendix: File Locations

| File | Path |
|------|------|
| SKILL.md | `harness/skills/harness-release/SKILL.md` |
| sync-version.sh | `harness/skills/harness-release/scripts/sync-version.sh` |
| validate-release-notes.sh | `harness/skills/harness-release/scripts/validate-release-notes.sh` |
| release-preflight.sh | `harness/skills/harness-release/scripts/release-preflight.sh` |
| check-consistency.sh | `harness/skills/harness-release/scripts/check-consistency.sh` |
| check-residue.py | `harness/skills/harness-release/scripts/check-residue.py` |
| This manifest | `.claude/state/phase-81-manifest.md` |
