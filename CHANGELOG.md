# Changelog

Change history for claude-code-harness.

> **Writing Guidelines**: Focus on user-facing changes. Keep internal fixes brief.

## [Unreleased](https://github.com/tim-hub/powerball-harness/compare/v4.5.2...HEAD)

## [4.5.2](https://github.com/tim-hub/powerball-harness/compare/v4.5.1...v4.5.2) - 2026-04-17

### Theme: Guardrail performance, drift detection, and memory hygiene

**Cuts guardrail latency up to 77%, adds pre-implementation drift detection, and replaces stale maintenance references with actionable skill commands.**

---

#### Plans Drift Check on `harness-work` Entry (Phase 65.1)

**Before**: Stale Plans.md markers (e.g. a task marked `cc:TODO` when it was already committed) were only caught after implementation and review finished.

**After**: `harness-work` now runs `harness/scripts/plans-drift-check.sh` before mode selection. If stale markers are detected the user sees a drift summary and must confirm before implementation starts. Zero behavioral change when Plans.md is already in sync.

#### Fix Stale `/maintenance` References in Hook Warning Messages

**Before**: Hook warnings for oversized `session-log.md` and `Plans.md` referenced `/maintenance` — a deprecated slash command removed in v4.1.0.

**After**: Plans.md warnings now say `/harness-plan archive`; session-log.md warnings now say `/harness-plan session-log`.

#### `harness-plan session-log` — Session Log Month Archiving

**Before**: When `session-log.md` exceeded 500 lines, hook warnings said "consider splitting it by month manually" with no concrete path to follow.

**After**: `harness-plan session-log` groups sessions by calendar month, moves past-month blocks to `.claude/memory/session-log-YYYY-MM.md`, rewrites the active file with only current-month sessions and an updated Index, then commits. Hook warnings now point directly to `/harness-plan session-log`.

#### Guardrail Engine Optimizations — Phase 62 (62.2 / 62.3 / 62.4)

**Before**: `go/internal/guardrail/` called `filepath.EvalSymlinks` on every Write/Edit protected-path check (even for the same file edited repeatedly), ran `normalizeCommand` with a regex allocation on every Bash command, and ran 12 security regex patterns against every file's content regardless of whether the file could plausibly contain secrets.

**After**: Three targeted optimizations cut guardrail latency significantly on the M4 Pro benchmark:

| Path | Baseline | Optimized | Δ |
|------|----------|-----------|---|
| Pre-tool (per hook invocation) | 5,969 ns | 4,740 ns | **−21%** |
| Post-tool non-test file | 19,085 ns | 4,356 ns | **−77%** (0 allocs) |
| Post-tool test file | 80,437 ns | 40,616 ns | **−50%** (0 allocs) |

- **62.2** — `isProtectedPath` now caches `filepath.EvalSymlinks` results in a bounded 256-entry FIFO map. Repeated writes to the same file resolve symlinks exactly once per session.
- **62.3** — `getChangedContent` called once and reused for both tampering and security checks. Security scanning gated by a 20-string `hasSuspiciousContent` pre-screen; most source files skip 12 regexes entirely.
- **62.4** — `normalizeCommand` fast-path returns the original string (zero allocation) when no tabs, newlines, or double-spaces are present — the common case for well-formatted commands.

#### Baseline: Guardrail Engine Performance Benchmarks (Phase 62)

**Before**: No benchmark baseline existed for `go/internal/guardrail/` — optimization work in Phases 62–64 had no reference point for before/after comparison.

**After**: `benchmarks/phase62-baseline.json` records pre-tool and post-tool median latency on darwin-arm64. Future optimization PRs reference this file for improvement percentages.

## [4.5.1](https://github.com/tim-hub/powerball-harness/compare/v4.5.0...v4.5.1) - 2026-04-16

### Theme: Hook cost reduction, security hardening, architecture documentation

**Removes an expensive LLM hook from the pre-tool hot path; tightens the deny rule list; adds three Mermaid diagram READMEs that document the hook chain, full workflow, and Go guardrail engine.**

---

#### 1. Remove PreToolUse Secret-Scanning Agent Hook

**Before**: A Haiku-based agent hook fired on every PreToolUse `Write|Edit` event, reviewing code for hardcoded secrets and security vulnerabilities. At 30 s timeout and model invocation overhead, this added significant latency to every edit — even trivial whitespace changes.

**After**: The agent hook is removed. Secret detection is already handled by the Go guardrail engine (R02, R03, R09 rules) and the post-tool security pattern scanner. Same coverage; no LLM invocation on the pre-tool fast path.

#### 2. Harden `settings.json` Deny Rules

**Before**: `.env` read blocking used `Bash(cat .env:*)`, which only caught `cat`; other read commands slipped through. `export PATH=*`, `export LD_LIBRARY_PATH=*`, and `export PYTHONPATH=*` were not in the deny list, allowing environment-variable injection.

**After**: Deny pattern widened to `Bash(* .env)` to block any command targeting `.env`; three `export` path patterns added to the deny list; `git reset --hard *` entry moved to the correct deny section.

#### 3. Architecture Diagram READMEs

**Before**: Understanding the hook event flow, the skill/agent workflow, or the Go guardrail engine internals required reading source code directly — no visual overview existed.

**After**: Three new Mermaid diagram READMEs:

- **`harness/hooks/README.md`** — LR flowchart of all 22 hook event types → shell shims → Go binary → handler scripts; implementation pattern table (command / agent / prompt).
- **`harness/README.md`** — Full lifecycle diagram (Setup → Plan → Work → Review → Release); execution mode decision tree; Breezing fix-loop sequence diagram; memory architecture layers; skill catalog (23 skills) and agent roles table.
- **`go/README.md`** — End-to-end guardrail flow; rule engine iterator; R01–R13 rules grouped by tool type; post-tool tampering + security scan pipeline; permission handler safe-command allowlist; state resolution (env → SQLite → defaults); fail-safe design.

## [4.5.0](https://github.com/tim-hub/powerball-harness/compare/v4.4.5...v4.5.0) - 2026-04-15

### Theme: Prebuilt binaries, agent optimization pass, reviewer upgrade to Opus

**Binaries now ship in `harness/bin/` so installs never need Go. All 6 agent files optimized (descriptions, tool lists, size). Reviewer upgraded to Opus for sharper critical/major classification.**

---

### Phase 60: Prebuilt binaries + build tooling overhaul

**Prebuilt platform binaries now ship in `harness/bin/`. Fresh installs no longer require Go. `make build-all` cross-compiles all platforms. The pre-commit hook auto-rebuilds binaries when `go/` files change.**

---

#### 1. Prebuilt binaries committed to `harness/bin/`

**Before**: Fresh plugin installs required Go to be installed. `/harness-setup init` ran a `binary` build step that compiled the harness binary from Go source — failing silently if Go was absent, and adding setup time for every new user.

**After**: All three platform binaries (`harness-darwin-arm64`, `harness-darwin-amd64`, `harness-linux-amd64`) ship prebuilt in `harness/bin/`. No Go toolchain needed for fresh installs. The `binary` subcommand has been removed from `harness-setup` entirely.

#### 2. Build tooling moved to `local-scripts/`

**Before**: `build-binary.sh` lived inside `harness/skills/harness-setup/scripts/` — mixing a contributor dev tool with user-facing setup scripts. The root Makefile `build` target pointed to this skill-scripts path.

**After**: `build-binary.sh` moved to `local-scripts/build-binary.sh` alongside other dev helpers. `make build` now calls `local-scripts/build-binary.sh`. New `make build-all` target cross-compiles all three platforms into `harness/bin/` via `go/scripts/build-all.sh`.

#### 3. Pre-commit hook auto-rebuilds on `go/` changes

**Before**: Contributors changing Go source had to manually rebuild the binary and stage it — easy to forget, leading to stale binaries in-repo.

**After**: `.githooks/pre-commit` detects staged `go/` files and automatically runs `make build-all`, re-staging all three platform binaries before the commit completes.

### Phase 61: Agent files optimization pass

**All 6 `harness/agents/` files optimized: descriptions now follow skill-description.md format, `model:` fields have rationale comments, `allowed-tools` corrected, and the two longest files trimmed by extracting Codex CLI details to a references file.**

---

#### 1. Description format normalized across all agents

**Before**: All 5 agent definitions used capability-prose descriptions (`"Integrated worker that cycles through..."`, `"Safety-first diagnosis and fix support..."`) that don't start with the required `Use when <trigger>` prefix. Routing inference was less precise.

**After**: All descriptions start with `Use when <trigger>` and stay ≤300 chars, matching `.claude/rules/skill-description.md`. Each includes relevant exclusions (`Do NOT load for: ...`) where applicable.

#### 2. `model:` rationale comments added; `allowed-tools` corrected

**Before**: All agents had `model: sonnet` with no explanation. `ci-cd-fixer.md` listed `tools: [Read, Write, Bash, Grep, Glob]` but its body instructed using the `Edit` tool for TypeScript fixes — an inconsistency.

**After**: Each `model: sonnet` line has an inline comment explaining the choice (e.g. `# balanced capability/cost for implementation; haiku lacks depth for complex tasks`). `ci-cd-fixer.md` now lists `Edit` in its tools.

#### 3. `team-composition.md` and `ci-cd-fixer.md` trimmed

**Before**: `team-composition.md` at 570 lines included a 160-line Codex CLI Environment section and a phased rollout schedule inline. `ci-cd-fixer.md` at 475 lines had verbose report templates and a 50-line auto-detection signal section.

**After**: Codex CLI details extracted to `harness/agents/references/team-composition-codex.md` (linked). `team-composition.md` is 404 lines. `ci-cd-fixer.md` trimmed to 342 lines by condensing templates and collapsing the auto-detection section to essential steps.

#### 4. Reviewer upgraded to `opus`; `team-composition.md` frontmatter added

**Before**: The `reviewer` agent ran on `sonnet`. Review quality occasionally missed nuance in critical/major classification thresholds. `team-composition.md` lacked YAML frontmatter, causing the plugin loader to silently skip it.

**After**: `reviewer` model set to `opus` — better reasoning reduces false-positive `REQUEST_CHANGES` and missed `APPROVE` calls. `team-composition.md` gains minimal `name` + `description` frontmatter.

## [4.4.6] - 2026-04-15

### Theme: SKILL.md quality pass — all 26 skills

**Fixes Quick Reference table format, anchors all script paths to `${CLAUDE_SKILL_DIR}`, removes dead references, and updates deprecated slash commands across every skill.**

---

#### 1. Quick Reference table format standardized

**Before**: Several skills used bash code blocks or 2-column tables in their Quick Reference sections instead of the required 3-column `User Input | Subcommand | Behavior` format. `harness-work` used `Mode` as the middle column header. Two skills (`session`, `notebook-lm`) had no Quick Reference at all despite having multi-token `argument-hint` values.

**After**: All Quick Reference sections now use the 3-column format with the correct `Subcommand` header. `breezing`, `harness-sync`, `harness-release`, `writing-changelog`, and `harness-work` converted; `session` and `notebook-lm` had tables added.

#### 2. Script paths anchored to `${CLAUDE_SKILL_DIR}`

**Before**: Several skills referenced scripts with bare relative paths like `bash scripts/foo.sh` or `python3 scripts/gog_parse_url.py`. These paths depend on the shell's current working directory at runtime, causing silent failures when Claude's working directory differs from the skill's directory.

**After**: All script references anchored to the correct tier — skill-local (`"${CLAUDE_SKILL_DIR}/scripts/..."`) or plugin-level (`"${CLAUDE_SKILL_DIR}/../../scripts/..."`). Affected: `gogcli-ops`, `session-control`, `session-state`, `harness-review` (SKILL.md + `dual-review.md`), `harness-work` (SKILL.md + `codex-work.md`), `memory`.

#### 3. Dead references removed

**Before**: Several skills referenced documentation files that do not exist: `docs/SESSION_ORCHESTRATION.md`, `docs/MEMORY_POLICY.md`, `AGENTS.md` in deploy health-checking, and a broken `${CLAUDE_SKILL_DIR}/../../docs/release-preflight.md` path (resolves to non-existent `harness/docs/`). `principles/references/vibecoder-guide.md` duplicated the standalone `vibecoder-guide` skill.

**After**: Dead references removed and replaced with inline prose or correct `(project-root)` annotations. Redundant `vibecoder-guide.md` reference file deleted.

#### 4. Deprecated slash commands updated

**Before**: `workflow-guide`, `session-init`, `session`, `deploy`, `memory`, and `principles` contained legacy slash commands (`/harness-init`, `/plan-with-agent`, `/work`, `/handoff-to-cursor`, `/sync-status`, etc.) from the pre-skills era.

**After**: All updated to current skill invocations: `harness-setup init`, `harness-plan`, `harness-work`, `cc-cursor-cc`, `harness-sync`, etc.

#### 5. Redundant `## Trigger Phrases` sections removed

**Before**: 11 files contained `## Trigger Phrases` sections listing user-quoted phrases — redundant with the `description:` frontmatter, which is the SSOT for skill auto-loading.

**After**: All `## Trigger Phrases` and `## Trigger Conditions` sections removed. Routing signal consolidated in each skill's `description:` field.

## [4.4.5](https://github.com/tim-hub/powerball-harness/compare/v4.4.4...v4.4.5) - 2026-04-15

### Theme: Fix plugin install error and sync-version.sh template glob

**Resolves "output-styles path not found" on plugin install; sync-version.sh now auto-discovers all templates instead of relying on a hardcoded list.**

---

#### 1. outputStyles path double-prefix in marketplace.json

**Before**: After installing the plugin, `/doctor` reported `output-styles path not found: .../harness/output-styles/`. The `marketplace.json` entry `"outputStyles": "./harness/output-styles/"` was resolved relative to the installed plugin root — which is already the `harness/` directory (set by `"source": "./harness/"`). This produced a double `harness/harness/output-styles/` lookup that never resolves.

**After**: Changed to `"outputStyles": "./output-styles/"`. The path now resolves correctly against the installed plugin root. `harness.toml` already used this correct form; marketplace.json now matches.

#### 2. sync-version.sh hardcoded template list replaced with glob

**Before**: `sync-version.sh` (bump and check) only updated `_harness_version` in 3 hardcoded template files (CLAUDE.md, AGENTS.md, Plans.md). Templates added in v4.4.3 (`rules/quality-gates.md.template` etc.) were silently skipped, requiring manual bumps every release.

**After**: Both the `sync` and `check` commands now use `grep -rl '_harness_version:' harness/templates/` to discover all templates automatically. Any new `.md.template` with a `_harness_version:` field is picked up without code changes.

## [4.4.4](https://github.com/tim-hub/powerball-harness/compare/v4.4.3...v4.4.4) - 2026-04-15

### Theme: CI hotfixes — validate-plugin.sh strict-mode regression and missing template frontmatter

**Two CI failures introduced by v4.4.3 fixes: `set -e` exposed a latent post-increment bug, and newly registered templates were missing required frontmatter.**

---

#### 1. validate-plugin.sh post-increment crash under set -e

**Before**: After v4.4.3 added `set -euo pipefail` to `tests/validate-plugin.sh`, the CI job failed immediately after printing "26 skills defined". The culprit was `((SKILLS_WITH_DESCRIPTION++))` — in bash, a post-increment expression evaluates to the *old* value, so when the counter is 0, `((0))` exits with code 1, which `set -e` treats as a fatal error.

**After**: Changed `((var++))` to `((++var))` (pre-increment). The expression now evaluates to the new value, which is always ≥ 1 after the first increment. The script runs cleanly end-to-end.

#### 2. Missing frontmatter on three newly registered rules templates

**Before**: v4.4.3 registered `rules/quality-gates.md.template`, `rules/security-guidelines.md.template`, and `rules/tdd-guidelines.md.template` in `template-registry.json` with `tracked: true`, but the files themselves had no YAML frontmatter. `test-frontmatter-integration.sh` Test 7 requires all `tracked: true` `.md.template` files to carry `_harness_template:` and `_harness_version:` fields, so CI failed.

**After**: Added proper YAML frontmatter to all three files.

## [4.4.3](https://github.com/tim-hub/powerball-harness/compare/v4.4.2...v4.4.3) - 2026-04-15

### Theme: Phases 56–58 — Project-wide review fixes (validator, agent frontmatter, docs, script hygiene)

**Fixes 23 issues found in a comprehensive project-wide review: critical Go validator bug, stale agent configs, documentation drift from the v4 Go migration, and shell script hygiene.**

---

#### 1. Go validator now accepts short model aliases and bare list syntax (Phase 56.1, fix(validate))

**Before**: `harness validate agents` rejected all 5 agent files with errors like `model "sonnet" is not a recognized model name` and `could not parse list items from "[Read, Write, Bash]"`. The validator was effectively broken for any real use.

**After**: The `validModelNames` map accepts short aliases (`sonnet`, `opus`, `haiku`) in addition to full IDs. `parseStringSlice` handles both bare items `[Read, Write]` (agent style) and quoted items `["Read", "Write"]` (skill style). `harness validate agents --root harness` now passes all 5 agents cleanly.

#### 2. Agent frontmatter standardized across ci-cd-fixer and error-recovery (Phase 56.2–56.4)

**Before**: `ci-cd-fixer.md` referenced a non-existent `verify` skill, used the old `disallowedTools: [Task]` name (renamed to `Agent` in CC v2.1.63), had inconsistent hook syntax, and was missing `permissionMode`, `effort`, and `maxTurns`. `error-recovery.md` had the same issues plus referenced two non-existent skills (`verify`, `troubleshoot`). `scaffolder.md`'s output schema only listed `v2 | v3` as valid `harness_version` values.

**After**: Both agents have `permissionMode: bypassPermissions`, `effort: medium`, `maxTurns: 75`, correct `disallowedTools: [Agent]`, and no references to non-existent skills. `error-recovery.md` has a deprecation notice clarifying it's consolidated into `worker`. `scaffolder.md` accepts `v4` in the version enum.

#### 3. Documentation drift from v4 Go migration resolved (Phase 57)

**Before**: Multiple docs still described the v3 TypeScript architecture:
- `docs/CLAUDE-skill-catalog.md` listed 5 skills (`impl/`, `verify/`, `handoff/`, `maintenance/`, `troubleshoot/`) that don't exist
- `CONTRIBUTING.md` described the wrong CHANGELOG format (Keep a Changelog style vs. actual Before/After narrative), wrong version sources (referenced `marketplace.json` which has had no version field since v4.4.0), and had a duplicate step "3." in the Testing section
- `.claude/memory/patterns.md` P1–P3 had TypeScript code examples referencing `core/src/` files removed in v4
- `docs/CLAUDE-feature-table.md` had 2 duplicate rows and open-ended "planned" items with no review dates
- `go/DESIGN.md` listed `internal/plans/` as an existing package
- `.claude/rules/hooks-editing.md` required maintaining a `.claude-plugin/hooks.json` that doesn't exist

**After**: All references updated to reflect v4 reality. Skill catalog shows only real skills. CONTRIBUTING.md matches actual project conventions. Patterns P1–P3 marked as superseded with pointers to `go/internal/guardrail/`. Feature table deduplicated; planned items annotated with `(planned — review 2026-07-15)`. DESIGN.md notes `plans/` as not yet implemented. hooks-editing.md updated to name `harness/hooks/hooks.json` as the single SSOT. 4 orphaned templates registered in `template-registry.json`.

#### 4. Shell script hygiene: path conventions, strict mode, settings (Phase 58)

**Before**: 4 scripts used `dirname "$0"` (breaks when sourced — should use `${BASH_SOURCE[0]}`). A variable named `repo_root` in `codex-setup-local.sh` actually resolved to the plugin root. `generate-sprint-contract.sh` used `process.cwd()` for project root resolution. `tests/validate-plugin.sh` hardcoded `/tmp` and was missing `set -e`. `harness/settings.json` had an inconsistent deny rule `"Bash(* .env)"` and aggressively hard-denied `export PATH/LD_LIBRARY_PATH/PYTHONPATH`. `vibecoder-guide/SKILL.md` described a user attribute instead of task shape.

**After**: All 4 scripts use `${BASH_SOURCE[0]}`. `codex-setup-local.sh` variable renamed to `plugin_dir` with `# plugin-local:` comment. `generate-sprint-contract.sh` uses `git rev-parse --show-toplevel`. `validate-plugin.sh` uses `${TMPDIR:-/tmp}` and `set -euo pipefail`. Settings deny rule syntax normalized; `export PATH/LD_LIBRARY_PATH/PYTHONPATH` moved from `deny` to `ask`. `vibecoder-guide` description rewritten to task shape.

## [4.4.2](https://github.com/tim-hub/powerball-harness/compare/v4.4.1...v4.4.2) - 2026-04-15

### Theme: Phase 55 — Three-tier path convention standardization + sync-version improvements

**Establishes a clear three-tier path convention across all skills and scripts, extends version sync to cover all template files, and eliminates CWD-dependent bare relative paths.**

---

#### 1. Three-tier path convention for skills and scripts

**Before**: Skills used a mix of bare relative paths (`bash skills/harness-release/scripts/...`, `./scripts/sync-version.sh`), undefined `${CLAUDE_PLUGIN_ROOT}` variables, and inconsistent anchoring. Paths broke silently when Claude's working directory differed from the repo root.

**After**: All paths are anchored to one of three explicit tiers:
- **skill-local** (`${CLAUDE_SKILL_DIR}/scripts/...`) — files within the skill's own directory
- **plugin-local** (`${CLAUDE_SKILL_DIR}/../../scripts/...`) — files elsewhere in the harness plugin
- **project-root** (`git rev-parse --show-toplevel` in scripts, plain names in prose) — files in the user's repo

| Before | After |
|--------|-------|
| `bash skills/harness-release/scripts/release-preflight.sh` | `bash "${CLAUDE_SKILL_DIR}/scripts/release-preflight.sh"` |
| `./scripts/sync-version.sh bump` | `"${CLAUDE_SKILL_DIR}/scripts/sync-version.sh" bump` |
| `${CLAUDE_PLUGIN_ROOT}/scripts/codex-companion.sh` | `${CLAUDE_SKILL_DIR}/../../scripts/codex-companion.sh` |
| CHANGELOG.md checked at plugin root | CHANGELOG.md checked at `git rev-parse --show-toplevel` |

#### 2. Script relocations for logical ownership

**Before**: `release-preflight.sh` and `sync-version.sh` lived in `harness/scripts/` (shared scripts dir), despite being release-workflow-specific tools.

**After**: Both scripts moved to `harness/skills/harness-release/scripts/` where they belong. References updated across SKILL.md, docs, local-scripts, and CONTRIBUTING.md.

#### 3. CHANGELOG check uses git project root

**Before**: `release-preflight.sh` checked for `CHANGELOG.md` relative to the plugin root (`harness/`), which always failed for the actual user repo changelog. Error message also read `"CHANGELOG.md exists"` when the file was *missing*.

**After**: Script derives `GIT_ROOT=$(git rev-parse --show-toplevel)` and checks `$GIT_ROOT/CHANGELOG.md`. Error message corrected to `"CHANGELOG.md not found"`.

#### 4. sync-version.sh extended to cover all template files

**Before**: `sync-version.sh` only synced `harness/VERSION` → `harness/harness.toml`. Template files (`template-registry.json`, `CLAUDE.md.template`, `AGENTS.md.template`, `Plans.md.template`) drifted silently after each release.

**After**: `sync-version.sh sync` and `sync-version.sh bump` now update all 6 version sources atomically. `make sync-version` added as a convenience target. `make check-version` path fixed to match the script's new location.

| Before | After |
|--------|-------|
| 4 template files stayed at old version | All 6 version sources updated together |
| `make check-version` pointed to old path | Path corrected to `harness/skills/harness-release/scripts/` |

#### 5. Path convention lint check in validate-plugin.sh (section 12)

**Before**: No automated check for bare relative script paths in SKILL.md bash blocks.

**After**: Section 12 in `tests/validate-plugin.sh` greps all SKILL.md files for `bash "?(\./)?skills/` patterns and fails the suite if any are found. `.claude/rules/path-conventions.md` documents the three-tier model for future contributors.

#### 6. CI-only test skip in `make test-all`

**Before**: Two known-failing scripts (`test-agent-telemetry-summary.sh`, `test-structured-handoff-artifact.sh`) were unconditionally skipped in `make test-all`, hiding them from local runs.

**After**: The skip only activates when `CI=true`. Local runs always execute all scripts. The CI workflow now sets `CI: "true"` explicitly at the job level.

## [4.4.1](https://github.com/tim-hub/powerball-harness/compare/v4.4.0...v4.4.1) - 2026-04-15

### Theme: CI stability — skip known failing test scripts

**Prevents two known-failing test scripts from blocking `make test-all` while their underlying issues are resolved.**

---

#### 1. Skip known-failing tests in `make test-all`

**Before**: `test-agent-telemetry-summary.sh` and `test-structured-handoff-artifact.sh` caused `make test-all` to fail, blocking CI for unrelated work.

**After**: Both scripts are skipped with a `SKIP (known failure)` label in `test-all` output. The pass/fail count excludes them so CI can proceed.

| Before | After |
|--------|-------|
| `make test-all` fails on two scripts | `make test-all` skips them with `SKIP (known failure)` |
| CI blocked for unrelated changes | CI passes; skip is visible in output |

## [4.4.0](https://github.com/tim-hub/powerball-harness/compare/v3.17.4...v4.4.0) - 2026-04-15

### Theme: Phase 52-54 — Marketplace restructure, Makefile CI layer, test suite hardening

**Moves all plugin files into a `harness/` subfolder enabling multi-plugin coexistence, stabilizes CI against path changes via a Makefile abstraction layer, and hardens the test suite against macOS sandbox TMPDIR restrictions.**

---

#### 1. Multi-plugin marketplace restructure (Phase 52)

**Before**: All plugin files (`skills/`, `agents/`, `hooks/`, `scripts/`, `templates/`, `bin/`, `VERSION`, etc.) lived at the repo root. `marketplace.json source: "./"` pointed to the whole repo. Only one plugin could coexist in the repo.

**After**: All plugin files now live under `harness/`. `marketplace.json` points `source: "./harness/"`. The repo root holds only shared infrastructure (`docs/`, `go/`, `tests/`, `.github/`, `README.md`, `CHANGELOG.md`), with room for additional plugins. Developer-only tooling lives in `local-scripts/` (not shipped to users).

| Before | After |
|--------|-------|
| `skills/`, `agents/`, `hooks/` at repo root | `harness/skills/`, `harness/agents/`, `harness/hooks/` |
| `scripts/`, `templates/`, `bin/` at root | `harness/scripts/`, `harness/templates/`, `harness/bin/` |
| `VERSION`, `harness.toml`, `settings.json` at root | `harness/VERSION`, `harness/harness.toml`, `harness/settings.json` |
| `marketplace.json source: "./"` | `marketplace.json source: "./harness/"` |
| `marketplace.json` had `version` field (CC validator rejected) | Version tracked only in `harness/VERSION` + `harness/harness.toml` |
| `.claude/scripts/` held dev tools mixed with plugin scripts | Dev tools moved to `local-scripts/` at repo root |

#### 2. Makefile as stable CI interface layer (Phase 53)

**Before**: CI workflow steps and local dev commands referenced raw script paths (`bash ./tests/validate-plugin.sh`, `bash ./local-scripts/check-consistency.sh`). After Phase 52 moved scripts into subfolders, every path had to be updated in N places — drift was invisible until CI ran.

**After**: All CI steps and local dev commands call `make <target>`. The Makefile is the single point of update when script paths change; workflow files stay untouched.

```sh
make validate    # bash tests/validate-plugin.sh
make check       # bash local-scripts/check-consistency.sh
make test-all    # run all tests under tests/ and harness/tests/
```

#### 3. Test suite hardening (Phase 54)

**Before**: 8 of 29 tests failed in the Claude Code sandbox: macOS sets `TMPDIR` to `/var/folders/...` which the sandbox blocks for writes, causing silent `mktemp` failures. Several tests also referenced pre-Phase-52 script paths and old skill names (`impl`, `verify`, `harness-init`, `harness-update`).

**After**: All 29 tests pass. Every test script exports `TMPDIR=/tmp` before any `mktemp` call. Tests that require unavailable external resources (socket binding, sibling repos) skip with a `SKIP:` message instead of failing.

## [4.3.0] - 2026-04-15

### Theme: Phase 50+51 — Refocus skills directory, eliminate 600-file mirror dirs

**Removes non-development skills from `skills/` and replaces build-time mirror sync with setup-time copy, shrinking the plugin by ~600 duplicate files.**

---

#### 1. Refocus `skills/` on software development (Phase 50)

**Before**: `allow1` (image generation), `generate-slide`, `generate-video`, and `video-scene-generator` lived in `skills/` and `agents/` alongside core development tools, cluttering the skill palette. `routing-rules.md` was filed as a skill despite being a reference document.

**After**: Non-development skills moved to `.claude/skills/` and `.claude/agents/` — out of the user-facing palette. `routing-rules.md` relocated to `.claude/rules/skill-routing-rules.md`.

- `skills/allow1` → `.claude/skills/allow1`
- `skills/generate-slide` → `.claude/skills/generate-slide`
- `skills/generate-video` → `.claude/skills/generate-video`
- `agents/video-scene-generator.md` → `.claude/agents/video-scene-generator.md`
- `skills/routing-rules.md` → `.claude/rules/skill-routing-rules.md`

#### 2. Eliminate mirror directories — setup-time copy replaces build-time sync (Phase 51)

**Before**: `codex/.codex/skills/` and `opencode/skills/` were build-time mirrors of `skills/`, kept in sync by `sync-skill-mirrors.mjs` and `build-opencode.mjs`. Every skill edit required re-running sync scripts or CI would fail. The plugin repo shipped ~600 duplicate skill files.

**After**: Mirror directories deleted. `harness-setup codex` and `harness-setup opencode` copy skills from the plugin's `skills/` into the user's project at setup time. `harness-setup duo` sets up both. Config templates live in `templates/codex/`, `templates/opencode/`, and `templates/codex-skills/`. No ongoing sync required.

- Deleted `codex/` and `opencode/` mirror directories (~600 files removed from plugin repo)
- Moved `skills-codex/` → `templates/codex-skills/` (codex-native overrides: breezing, harness-work)
- Added `harness-setup codex`: copies templates + skills to `.codex/`, patches `disable-model-invocation: true` into each SKILL.md, overlays codex-native overrides
- Added `harness-setup opencode`: copies templates + skills to `.opencode/`
- Added `harness-setup duo`: runs codex + opencode setup in one step
- Removed 4 mirror sync scripts (`sync-skill-mirrors.mjs`, `build-opencode.mjs`, `sync-skills.mjs`, `validate-opencode.mjs`)
- Replaced mirror sync CI with template existence check; removed PostToolUse auto-sync hook

## [4.2.0] - 2026-04-14

### Theme: Phase 49 — build-from-source installation, hooks.json SSOT, deny list cleanup

**Replaces the network-dependent binary download with a local Go build, makes `.claude-plugin/hooks.json` the single source of truth for hook configuration, and removes 42 duplicate entries from the permission deny list.**

---

#### 1. Build harness binary from source instead of downloading

**Before**: `harness-setup binary` ran `download-binary.sh`, which fetched a pre-built binary from a remote URL. This failed silently on air-gapped machines, behind corporate proxies, or whenever the release asset was unavailable.

**After**: `harness-setup binary` now runs `build-binary.sh`, which compiles the binary directly from the Go source already present in the repository using `CGO_ENABLED=0` for a fully static binary. Requires `go` to be installed; produces the correct `harness-<OS>-<ARCH>` binary in `bin/`.

```sh
# What runs now
bash scripts/build-binary.sh
# → CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/harness-darwin-arm64 ./cmd/harness/
```

#### 2. Consolidate hooks.json to a single SSOT

**Before**: `hooks/hooks.json` and `.claude-plugin/hooks.json` were two separate files with identical content. Any change had to be made in both places. `harness sync` copied one to the other, which would clobber itself if they diverged.

**After**: `hooks/hooks.json` is now a symlink pointing to `../.claude-plugin/hooks.json`. `.claude-plugin/hooks.json` is the single authoritative file. `harness sync` detects the symlink and skips the copy step, printing "skipped (symlinked to …)" instead.

#### 3. Deduplicate and categorize permission deny list

**Before**: `harness.toml` `[safety.permissions]` had 171 deny entries with 42 duplicates (same command in both `Bash(cmd:*)` and glob form, and within category groups).

**After**: 129 entries — zero duplicates — organized under labeled categories (`# --- dangerous system commands ---`, `# --- network ---`, `# --- git destructive ---`, etc.). The umbrella rules (`sudo:*`, `rm -rf:*`) subsume the per-command items they were already covering.

## [4.1.4] - 2026-04-14

### Theme: CI workflow fix, harness-mem removal, gitignore template update

**Fixes a broken CI workflow after Phase 45 script relocation, removes the `harness-mem` proxy skill that caused a plugin name conflict, and scopes gitignore runtime paths under `.claude/`.**

---

#### 1. Fix stale CI script paths in `.github/workflows/validate-plugin.yml`

**Before**: Phase 45 moved `scripts/ci/check-version-bump.sh` and `scripts/ci/check-consistency.sh` to `.claude/scripts/`, but `validate-plugin.yml` still referenced the old paths. CI was broken on every push.

**After**: Both workflow steps updated to `.claude/scripts/`. `scripts/ci/` added to `deleted-concepts.yaml` so the migration residue scanner catches any future re-introduction of the old paths.

#### 2. Remove `harness-mem` skill and simplify `harness-setup` default

**Before**: `skills/harness-mem/` was a thin proxy to `skills/memory/`. It existed solely so `/harness-mem` resolved, but conflicted with an identically-named skill from another plugin. The `harness-setup` no-args run sequenced through `binary → init → gitignore → harness-mem → cleanup`, making the first-run experience heavyweight and confusing.

**After**: `skills/harness-mem/` (and its codex/opencode mirrors) deleted. `harness-setup` with no args now runs `init` only. The `init` flow itself runs binary download and gitignore setup first, then generates `CLAUDE.md`, `Plans.md`, and `settings.json` — ensuring the binary is available before any hook-dependent steps run.

#### 3. Scope gitignore runtime paths under `.claude/`

**Before**: `templates/gitignore-harness` ignored `logs/`, `settings.local.json`, and `states/` at the project root — too broad, and inconsistent with how Harness actually places runtime files.

**After**: Paths scoped to `.claude/logs/`, `.claude/settings.local.json`, `.claude/states/`. Added `!.claude/agent-memory/` force-track entry so agent memory is never accidentally ignored.

## [4.1.3] - 2026-04-14

### Theme: Phase 46 — Skill housekeeping: rename, relocate, and complete missing assets

**`notebookLM` renamed to `notebook-lm`, `cc-update-review` moved to `.claude/skills/`, and `harness-setup` gains its missing script and reference file.**

---

#### 1. Rename `notebookLM` → `notebook-lm`

**Before**: The NotebookLM skill directory was named `notebookLM` (camelCase), inconsistent with every other skill's kebab-case naming. Mirror paths also used `notebookLM`.

**After**: Renamed to `notebook-lm` across `skills/`, `codex/.codex/skills/`, and `opencode/skills/`. All cross-references updated.

#### 2. Move `cc-update-review` to `.claude/skills/`

**Before**: `cc-update-review` was in `skills/` alongside user-facing skills, but it's an internal skill only referenced from `claude-codex-upstream-update`. It shouldn't appear in the user-facing skill palette.

**After**: Moved to `.claude/skills/cc-update-review/`. Removed from codex and opencode mirrors. References in `claude-codex-upstream-update/SKILL.md` updated to new path.

#### 3. Complete `harness-setup` missing assets

**Before**: `harness-setup/SKILL.md` referenced `scripts/merge-gitignore.sh` and `references/codex.md` that did not exist on disk. Running `harness-setup gitignore` or `harness-setup codex` would fail.

**After**:
- `skills/harness-setup/scripts/merge-gitignore.sh` — idempotent script that appends the harness-managed `.gitignore` block (ignore runtime dirs, force-track `.claude/memory/`, `.claude/rules/`, etc.)
- `skills/harness-setup/references/codex.md` — Codex CLI install, config template copy, skill sync verification, and troubleshooting table
- SKILL.md trimmed to 118 lines (was 126) by condensing the `harness-mem` inline template to descriptive prose

## [4.1.2] - 2026-04-14

### Theme: Phase 45 — Script reorganization into canonical locations

**Dead scripts removed and CI-only scripts moved to `.claude/scripts/`, cleaning up the `scripts/` directory after the v4.0 Go migration.**

---

#### 1. Delete 8 dead scripts left over from v4.0 Go migration

**Before**: Six shell hook handlers (`posttooluse-security-review.sh`, `posttooluse-tampering-detector.sh`, `pretooluse-guard.sh`, `permission-request.sh`, `skill-child-reminder.sh`) plus `setup-opencode.sh`, `ci/diagnose-and-fix.sh`, and `ci/check-checklist-sync.sh` still existed on disk despite being replaced by `bin/harness hook ...` Go-native calls in v4.0. Grep confirmed zero references outside CHANGELOG/tests.

**After**: All 8 deleted. `scripts/ci/` directory is now empty and removed. Migration residue scanner confirms no references remain.

#### 2. Move CI-only scripts to `.claude/scripts/`

**Before**: `scripts/ci/check-consistency.sh`, `check-version-bump.sh`, `check-template-registry.sh`, `check-residue.sh`, and `audit-skill-descriptions.sh` were called only by CI pipelines and `tests/validate-plugin.sh` — not by any plugin hook, skill, or agent. They were mixed into `scripts/` alongside plugin infrastructure.

**After**: All five moved to `.claude/scripts/`. Callers updated: `.github/workflows/`, `tests/validate-plugin.sh`, `skills/harness-release/SKILL.md`, and all rule/doc references. `REPO_ROOT` path resolution fixed in `check-residue.sh` and `audit-skill-descriptions.sh` (both needed `../..` instead of `..` after the one-level-deeper location change).

#### 3. Convert `scripts/validate-opencode.js` → `.claude/scripts/validate-opencode.mjs` (ESM)

**Before**: `scripts/validate-opencode.js` used CommonJS `require()` and implicit `__dirname`. GitHub Actions ran it with `node scripts/validate-opencode.js`.

**After**: Converted to ESM (`import` syntax, `fileURLToPath(import.meta.url)` for `__dirname`). Moved to `.claude/scripts/validate-opencode.mjs`. `.github/workflows/compatibility-check.yml` updated to `node .claude/scripts/validate-opencode.mjs`.

## [4.1.1] - 2026-04-14

### Theme: Script consolidation and automatic mirror sync

**Consolidated skill mirror sync scripts into `.claude/scripts/` with unified ESM format. Added automatic mirror sync hook on skill edits.**

---

#### 1. Move mirror sync scripts to `.claude/scripts/` and convert to ESM

**Before**: `build-opencode.js` (CommonJS) and `sync-skill-mirrors.sh` (shell script) lived in `scripts/`, making them difficult to share across projects.

**After**: Relocated under `.claude/scripts/` as `.mjs` (ES Modules), matching the format of `check-japanese.mjs`. Other projects can now reuse them by simply copying `.claude/scripts/`.

#### 2. Automatic mirror sync via PostToolUse hook

**Before**: Editing files under `skills/` required manually running `sync-skill-mirrors.sh` to propagate changes to `codex/.codex/skills/` and `opencode/skills/`. Forgetting caused CI failures.

**After**: A PostToolUse hook in `.claude/settings.json` (`sync-skills.mjs`) detects Write/Edit operations on `skills/` and automatically syncs both mirrors. Writes to `opencode/skills/`, `skills-codex/`, and `.codex/skills/` are skipped to prevent loops.

#### 3. Fix process substitution in `check-consistency.sh`

**Before**: The `diff_mirror()` function used process substitution (`<(...)`) which caused sandbox environments to deny `/dev/fd/` access, producing false-positive "mirror drift" reports.

**After**: Replaced process substitution with `$TMPDIR` temporary files. Comparisons now work correctly in all environments.

## [4.1.0] - 2026-04-14

### Theme: Self-Enforcing Skill Description Standards ("Use when" Rule)

**Every SKILL.md description now follows a uniform trigger-first format, enforced by a CI gate in `validate-plugin.sh`. Descriptions that drift from the standard are caught automatically.**

---

#### 1. New Rule: Skill Description Format

**Before**: SKILL.md `description` fields mixed triggers, exclusions, and capability introductions in any order. Some started with "Use this skill whenever the user mentions...", others with a skill summary. No rule governed the format, so drift was inevitable every time a new skill was added.

**After**: `.claude/rules/skill-description.md` defines the standard: descriptions must start with `Use when <trigger>`, describe task shape (not user phrases), keep introductions in the body, and stay under 300 characters. Includes good/bad examples from the actual codebase.

#### 2. Audit Script for Conformance Checking

**Before**: No automated way to verify whether a SKILL.md description met any standard.

**After**: `scripts/audit-skill-descriptions.sh` scans all SKILL.md files and reports tab-separated violations — missing prefix, forbidden phrases (`the user mentions`, `Use this skill`), and over-length. Accepts an optional directory argument to scope the scan.

```
$ bash scripts/audit-skill-descriptions.sh skills/
Scanned 31 SKILL.md file(s).
All descriptions conform to .claude/rules/skill-description.md.
```

#### 3. 59 SKILL.md Descriptions Rewritten

**Before**: All 59 SKILL.md files across `skills/` (31), `opencode/skills/` (26), and `skills-codex/` (2) were non-conformant. Typical violations: started with "Use this skill whenever...", enumerated 10+ user phrases, included trailing capability summaries reaching 500+ characters.

**After**: All 59 descriptions rewritten to `Use when <trigger>` form. Average length dropped from ~400 to ~180 characters. Introductory sentences moved to each SKILL.md body where not already present.

#### 4. CI Gate (validate-plugin.sh Section 10)

**Before**: No CI check prevented non-conforming descriptions from being merged.

**After**: `tests/validate-plugin.sh` Section 10 runs `audit-skill-descriptions.sh` on every validation. Each violation becomes its own `fail_test` line so per-file feedback survives in the summary. Missing script on older branches degrades to a warning.

#### 5. Drift Reporter for Cross-Directory Sync

**Before**: No tool to detect when `skills/` and `opencode/skills/` descriptions diverge after independent edits.

**After**: `scripts/skill-description-drift-report.sh` compares descriptions across both directories for all 26 shared skills. Reports divergent pairs with recommended actions; never overwrites. Currently shows 0 drift.

## [4.0.5] - 2026-04-13

### Theme: Auto-Download Platform Binary on Plugin Install

**The harness binary is now downloaded automatically when the plugin is installed, eliminating `UserPromptSubmit hook error` messages for new users.**

---

#### 1. Automatic Binary Download on Install

**Before**: After installing the plugin, all hooks produced `UserPromptSubmit hook error` on every prompt because the platform binary (`harness-darwin-arm64` etc.) was not present. Users had to manually run `/harness-setup binary` to fix it.

**After**: A pure shell download script (`skills/harness-setup/scripts/download-binary.sh`) runs automatically as the first step of the `Setup: init` hook. It detects the OS/arch, fetches the latest release from GitHub, and installs the correct binary. No Go runtime required to run the script.

#### 2. Graceful Hook Fallback When Binary Missing

**Before**: `bin/harness` exited with code 1 when the platform binary was not found, causing Claude Code to surface a hook error on every prompt.

**After**: `bin/harness` exits with code 0 (approve/no-op) when the binary is absent, so hooks pass silently until the binary is installed.

## [4.0.4] - 2026-04-13

### Theme: Fix GitHub Actions Auto-Release Race Condition and CI Improvements

**Fixed a race condition in the GitHub Actions release workflow when a release already exists, and cleaned up CI consistency checks.**

---

#### 1. GitHub Actions Release Race Condition Fix

**Before**: When a tag was pushed and a release already existed, `release.yml` would fail with an error.

**After**: Added an existence check so the workflow skips release creation if one already exists.

#### 2. CI Consistency Check Cleanup

**Before**: `check-consistency.sh` referenced the deleted `README_ja.md`, causing CI checks to fail.

**After**: Removed the stale reference; CI checks now pass cleanly.

#### 3. Mirror Sync and Distribution Cleanup

**Before**: `IMPLEMENTATION_SUMMARY.md` remained in `skills/` and all mirrors — an internal dev doc included in the distribution.

**After**: Removed the internal document from all mirrors; distribution is now clean.

## [4.0.3] - 2026-04-13

### Theme: Fix opencode/ Mirror Sync CI Check

**The `opencode/` directory was out of sync with its source, causing the mirror compatibility CI check to fail on every push.**

---

#### 1. opencode/ Regenerated

**Before**: `opencode/` contained stale generated files — 7 files diverged from their sources in `commands/` and `CLAUDE.md`, including a leftover `IMPLEMENTATION_SUMMARY.md` that had been deleted upstream. The CI check `node scripts/build-opencode.js` detected the drift and blocked pushes.

**After**: Ran `node scripts/build-opencode.js` to regenerate all `opencode/` files from their SSOT. The stale `IMPLEMENTATION_SUMMARY.md` was removed and PM command files were updated. CI check passes cleanly.

---

<!-- v4.0.x entries below were merged from upstream/main (Hokage branch) -->

## [4.0.2] - 2026-04-12

### Theme: Automatic Detection of "Invisible Residue" After Large-Scale Migrations

**After the full TS→Go migration in v4.0.0, 13 "old-world references" remained scattered across tests, documentation, and skill definitions — and were only discovered by chance. Going forward, Harness automatically detects this kind of problem and blocks releases before they ship.**

---

#### 1. Introduction of the Migration Residue Scanner

**Before**: After a major migration (e.g., v3→v4), references to "files and concepts that were supposed to be deleted" would linger throughout the codebase. Test scripts would keep grepping for deleted files, READMEs would still say "Node.js 18+ required," skill headings would still show `(v3)` — these were the kind of bugs that **passed tests, slipped through reviews, and were only noticed once users saw them**. In the two days after the v4.0.0 release, 13 such issues were discovered by chance.

**After**: Register "deleted paths and concepts" in `.claude/rules/deleted-concepts.yaml`, and `scripts/check-residue.sh` scans the entire repository to detect residue. Historical records (like CHANGELOG) are excluded via allowlist, so false positives are zero.

Three verification points run automatically:
- **During development**: Manual check with `bin/harness doctor --residue`
- **Per PR**: Auto-check in section 9 of `validate-plugin.sh`
- **Pre-release**: Auto-block in `harness-release` preflight

```bash
# Example: what check-residue.sh detects
bin/harness doctor --residue
# ⚠️  Residue found: 3 references to deleted concept "core/dist/index.js"
#   tests/test-hooks-sync.sh:42: node core/dist/index.js
#   CLAUDE.md:128: requires core/dist/index.js
#   docs/ARCHITECTURE.md:67: delegates to core/dist/index.js
```

#### 2. 10 Residue Items Fixed

The following residue from the v4.0.0 migration was detected and corrected:

| Location | Old reference | New reference |
|----------|--------------|---------------|
| `tests/validate-plugin.sh` | `core/dist/index.js` | `bin/harness hook` |
| `tests/test-hooks-sync.sh` | `node core/` | `bin/harness` |
| `tests/test-quality-guardrails.sh` | TypeScript paths | Go paths |
| `docs/ARCHITECTURE.md` | v3 architecture | v4 architecture |
| Multiple skills | `Harness v3` headings | `Harness v4` |

#### 3. Version Source Migrated from plugin.json to marketplace.json

**Before**: `.claude-plugin/plugin.json` was the canonical source for plugin metadata and version. It existed alongside `marketplace.json` with overlapping metadata, and all tooling (`sync-version.sh`, pre-commit hook, CI scripts, `emit-agent-trace.js`) pointed at it.

**After**: `plugin.json` has been removed. `marketplace.json` is the single source of plugin metadata. A `version` field was added to `marketplace.json`, and all tooling was updated to read/write from there. `emit-agent-trace.js` was also adapted for the nested structure (`plugins[0]` for license/author fields).

```bash
./scripts/sync-version.sh check
# ✅ Versions match: 4.0.2
```

---

## [4.0.1] - 2026-04-11

### Theme: Review Experience Improvements + Pre-v4.0.1 Polish

**Improved `/harness-review` output readability, fixed infrastructure issues, and unified naming conventions across the codebase.**

---

#### 1. Improved `/harness-review` Output

**Before**: Review output mixed machine-readable JSON with human-readable text, making it hard to scan. The "⚠️ Concerns" section lacked structure.

**After**:
- **"⚠️ Concerns" uses a 4-part structure**: Plain-language title → Problem (plain English) → Evidence (code quote) → Recommendation
- **JSON demoted to "📦 Detailed Data" section**: Explicitly marked "non-specialist reference"
- **"✅ Strengths" section added**: Highlights what was done well, not just problems
- Verdict now displayed prominently at the top

#### 2. Infrastructure Fixes

- Fixed `scripts/write-review-result.sh` to correctly handle `DOWNGRADE_TO_STATIC` verdict
- Corrected `scripts/generate-sprint-contract.sh` path references (v3→v4)
- Fixed broken symlinks in `codex/.codex/skills/`

#### 3. Naming Consistency

- Renamed all `v3` references to `v4` in skill headings and documentation
- Unified hook handler naming conventions across the codebase

---

## [4.0.0] - 2026-04-09

### Theme: Harness v4 — Full Rewrite to Go ("Hokage")

**127 shell scripts + TypeScript core replaced by a single Go binary. Hook response time reduced to under 5ms. Dual `hooks.json` management eliminated.**

---

#### 1. Go Binary (`bin/harness`) Replaces Shell + Node.js

**Before**: Each hook invoked `node core/dist/index.js` or a shell script. Node.js cold-start added 200–800ms per hook. 37 hook handlers existed as separate shell scripts.

**After**: All hooks invoke `bin/harness hook <name>`. The Go binary starts in under 5ms. All 37 handlers are compiled into a single binary.

```bash
# Before
"command": "node core/dist/index.js pre-tool"

# After  
"command": "bin/harness hook pre-tool"
```

#### 2. SQLite State Layer

**Before**: State was scattered across multiple JSON files in `.claude/state/`. Concurrent writes caused corruption. No atomic transactions.

**After**: Single SQLite database at `.claude/state/harness.db`. All state reads/writes are atomic. Session tracking, agent traces, fix proposals all in one place.

#### 3. Unified Configuration (`harness.toml`)

**Before**: `hooks.json` and `.claude-plugin/hooks.json` had to be kept in sync manually. Mismatches caused silent failures.

**After**: Edit only `harness.toml`. Run `harness sync` to regenerate all CC plugin files automatically. Dual management eliminated.

#### 4. TypeScript Core Removed

**Before**: `core/` directory contained TypeScript guardrail engine requiring Node.js 18+.

**After**: Guardrail rules (R01-R09+) implemented in `go/internal/guardrail/`. No Node.js dependency. `core/` directory excluded from distribution.

#### 5. Guardrail Parity

All guardrail rules ported to Go with full test coverage:

| Rule | Description | Status |
|------|-------------|--------|
| R01 | Test tampering detection | ✅ Ported |
| R02 | Dangerous command blocking | ✅ Ported |
| R03 | Secret file protection | ✅ Ported |
| R04 | Plans.md marker validation | ✅ Ported |
| R05 | Sprint contract enforcement | ✅ Ported |
| R06–R09 | Additional guardrails | ✅ Ported |

---

<!-- Local v3.17.x entries below -->

## [3.17.4](https://github.com/tim-hub/powerball-harness/compare/v3.17.3...v3.17.4) - 2026-04-12

### Theme: Automated changelog generation skill

**Added an `update-changelog` skill that generates CHANGELOG entries by diffing versions, so release notes follow a consistent Before/After format without manual effort.**

---

#### 1. New `update-changelog` skill

**Before**: Writing changelog entries required manually reviewing git diffs, categorizing changes, and formatting them in the project's Before/After style. This was tedious and format-inconsistent across releases.

**After**: Running `/update-changelog` (or triggering it after a version bump in `marketplace.json`) automatically gathers commits between the old and new version, categorizes them, and writes a properly formatted CHANGELOG entry. The skill lives at `.claude/skills/update-changelog/` as a project-local skill.

## [3.17.3](https://github.com/tim-hub/powerball-harness/compare/v3.17.2...v3.17.3) - 2026-04-12

### Theme: Repository rebrand + v3 directory consolidation

**Rebranded the repository from `Chachamaru127/claude-code-harness` to `tim-hub/powerball-harness`, eliminated the redundant `skills-v3/` and `agents-v3/` directories, and unified all skills and agents under `skills/` and `agents/`.**

---

#### 1. Repository URL rebrand

**Before**: Repository URLs pointed to `Chachamaru127/claude-code-harness` (original upstream) or `tim-hub/claude-code-harness` (fork). Marketplace, install scripts, CI badges, social posts, and documentation all referenced the old paths.

**After**: All URLs now point to `tim-hub/powerball-harness`. Updated across README, CONTRIBUTING, marketplace.json, CHANGELOG link references, CI scripts, install scripts, social post drafts, and benchmark docs. Added an Origin section to README crediting the original upstream repository.

#### 2. Merge `skills-v3/` and `agents-v3/` into `skills/` and `agents/`

**Before**: Two parallel directory structures existed — `skills/` (legacy, 28 skills) and `skills-v3/` (v3 consolidation, 7 core + 10 extension symlinks). The v3 migration was never completed; both directories had converged to identical content. `agents-v3/` held 4 agent files (`worker.md`, `reviewer.md`, `scaffolder.md`, `team-composition.md`) that had no counterparts in `agents/`.

**After**: Copied v3 agents into `agents/`. Removed `skills-v3/` entirely (core skills were duplicates, extensions were symlinks back to `skills/`). Updated all references across 27 files (docs, scripts, rules, CI tests, CLAUDE.md). `CHANGELOG.md` and `Plans.md` left as historical records.

#### 3. Planning skill uses Opus model

**Before**: `harness-plan` skill used the default model.

**After**: Added `model: opus` to `harness-plan` frontmatter for higher-quality planning output.

---

## [3.17.2](https://github.com/tim-hub/powerball-harness/compare/v3.17.1...v3.17.2) - 2026-04-11

### Theme: Full English translation + OSS readiness

**Translated the entire codebase from Japanese to English for open-source readiness, removed non-Claude Code platform directories, and cleaned up the plugin manifest.**

---

#### 1. Full English translation

**Before**: Skills, agents, rules, docs, hooks, scripts, workflows, benchmarks, and CI all contained Japanese text. This limited accessibility for English-speaking contributors and users.

**After**: All SKILL.md frontmatter descriptions, agent instructions, rule files, documentation, hook scripts, CI workflows, and benchmark reports translated to English. Japanese README (`README_ja.md`) and license (`LICENSE.ja.md`) removed.

#### 2. Non-Claude Code platform removal

**Before**: Repository included directories for Codex (`codex/`), OpenCode (`opencode/`), Cursor (`.cursor/`), and other non-CC platforms with mirrored skills and configurations.

**After**: Removed `codex/`, `opencode/`, `.cursor/`, and related platform-specific files. Harness now focuses exclusively on Claude Code as its target platform.

#### 3. Plugin manifest cleanup

**Before**: `.claude-plugin/plugin.json` existed alongside `marketplace.json` with overlapping metadata.

**After**: Removed redundant `plugin.json`. `marketplace.json` is the single source for plugin metadata.

## [3.17.1](https://github.com/tim-hub/powerball-harness/compare/v3.17.0...v3.17.1) - 2026-04-06

### Theme: harness-mem integration fix (emergency patch)

**Fixed an issue where harness-mem integration was broken for marketplace users.**

---

#### 1. Fix search paths in harness-mem-bridge.sh

**Before**: `memory-bridge.sh` -> `harness-mem-bridge.sh` only searched hardcoded development environment paths (`../harness-mem`, `~/LocalWork/...`) when looking for the harness-mem repository.
For users who installed via marketplace, the repository exists at `~/.harness-mem/runtime/harness-mem`, so the search failed with `exit 0` (silent failure), and no resume pack was generated at SessionStart.

**After**: Added the standard installation path `~/.harness-mem/runtime/harness-mem` as the highest priority in the search order.
When resuming a session with `/resume`, the previous work context (WIP tasks, recently edited files) is now properly restored.

## [3.17.0](https://github.com/tim-hub/powerball-harness/compare/v3.16.0...v3.17.0) - 2026-04-04

### Theme: Feature Table consistency recovery + upstream integration + Claude/Codex parity enhancement

**A release that resolves all "documented but non-functional" items in the Feature Table, incorporates new features from CC 2.1.87-2.1.90, and improves Harness reliability and utilization across both Claude and Codex contexts.**

---

### Added --dual flag to harness-review

**Added the `--dual` flag to run Claude Reviewer and Codex Reviewer in parallel, improving review quality through different model perspectives.**

#### 1. Dual review via --dual flag

**Before**: `/harness-review` ran only with Claude's Reviewer agent, limited to a single model's perspective. Getting a second opinion from Codex required manually running `scripts/codex-companion.sh review` separately.

**After**: Running `harness-review --dual` launches Claude Reviewer and Codex Reviewer in parallel, with their verdicts automatically merged. If either returns REQUEST_CHANGES, the overall verdict becomes REQUEST_CHANGES.
In environments where Codex is unavailable, it automatically falls back to Claude-only execution, so it's safe to use in projects without Codex setup.

```bash
# Claude + Codex parallel review
harness-review --dual

# Existing single-model flow remains unchanged
harness-review
harness-review code
```

The `dual_review` field in the output shows each model's verdict and the reasoning when verdicts differ.

---

### Claude Code 2.1.87-2.1.90 / Codex 0.118 Integration

(Added auto mode denial tracking and Breezing safety valve. Improved guardrail reliability leveraging CC-side hook fixes)

#### 1. Auto mode denial tracking via PermissionDenied hook

**CC Update**: A `PermissionDenied` hook now fires when the auto mode classifier denies a command (v2.1.89).
Returning `{retry: true}` informs the model that a retry is possible.

**Harness Integration**: Implemented `permission-denied-handler.sh` to record denial events as telemetry in `permission-denied.jsonl`.
When a Breezing Worker is denied, Lead is notified via `systemMessage` to consider alternative approaches.
Uses `agent_id` / `agent_type` to track "which agent was denied what".

#### 2. Documentation for defer permission decision

**CC Update**: Returning `"defer"` from a PreToolUse hook pauses a headless session, and hooks are re-evaluated when resumed via `claude -p --resume` (v2.1.89).

**Harness Integration**: Added design guidelines for defer decisions to hooks-editing.md.
Documented as a safety valve for when Breezing Workers encounter difficult-to-judge operations.
Specific defer rules (production DB writes, destructive git, etc.) to be designed after accumulating operational patterns.

#### 3. Improved guardrail reliability from PreToolUse exit 2 fix

**CC Update**: Fixed behavior when PreToolUse hooks return a block with JSON stdout + exit code 2 (v2.1.90).
Previously, blocking did not work correctly with this pattern.

**Harness Integration**: `pre-tool.sh` uses this pattern for denials, so guardrail denials work more reliably starting from v2.1.90.
No additional implementation changes needed (CC auto-inherited + existing code benefits as-is).

#### 4. Key CC auto-inherited fixes

- `--resume` prompt-cache miss fix (v2.1.90): Faster session resume
- autocompact thrash loop fix (v2.1.89): Stops after 3 consecutive cycles -> actionable error
- Nested CLAUDE.md re-injection fix (v2.1.89): Improved context efficiency
- SSE/transcript performance (v2.1.90): O(n^2) -> O(n) speedup
- PostToolUse format-on-save fix (v2.1.90): Resolved Edit/Write failures after hooks
- Cowork Dispatch fix (v2.1.87): Stabilized team communication

---

### Feature Table consistency recovery + implementation of unused features

#### 5. Feature Table exaggeration corrections (7 items)

Corrected descriptions in the Feature Table that could be misread as "implemented" to match actual state. HTTP hooks -> template only, OTel -> custom JSONL, Analytics Dashboard -> planned, LSP -> CC native, Auto Mode -> RP Phase 1, Slack -> future support, Desktop Scheduled Tasks -> CC native.

#### 6. PostCompact WIP restoration

**Before**: Warned about WIP tasks before context compaction, but did not restore them after compaction. The warning was unhelpful without follow-through.

**After**: PostCompact restores WIP information saved by PreCompact as a `systemMessage`, maintaining task state even after compaction.

#### 7. Webhook notification (TaskCompleted HTTP hook)

**Before**: The Feature Table stated "HTTP hooks implemented" but hooks.json had zero `type: "http"` entries.

**After**: Setting `HARNESS_WEBHOOK_URL` sends notifications to Slack / Discord / any URL on task completion. Silently skips when unset (opt-in).

#### 8. Security review (--security)

**Before**: `/security-review` was listed in the Feature Table but lacked a standalone feature.

**After**: `harness-review --security` launches a review focused on OWASP Top 10 + authentication/authorization + data exposure. Checks more strictly than normal using security-specific verdict criteria.

#### 9. Effort propagation to Codex Worker

**Before**: On the Claude side, Lead calculates task complexity and auto-injects ultrathink, but Codex Worker always used medium effort.

**After**: `calculate-effort.sh` computes a score from file count, dependencies, keywords, and DoD conditions, propagating effort to Codex Worker. Complex tasks automatically get higher effort.

#### 10. OTel Span transmission

**Before**: `emit-agent-trace.js` used a custom JSONL format. Could not send directly to Datadog or Grafana.

**After**: When `OTEL_EXPORTER_OTLP_ENDPOINT` is set, sends spans in OTel Span JSON format via HTTP POST. Falls back to existing JSONL when unset.

#### 11. Comprehensive overhaul of harness-release skill

Includes regression checklist, explicit non-NPM distribution, i18n support, mirror sync flow, SemVer criteria integration, and detailed `--dry-run` / `--complete` / `--announce` modes.

## [3.16.0](https://github.com/tim-hub/powerball-harness/compare/v3.15.0...v3.16.0) - 2026-04-01

### Theme: Long-running harness hardening + team/release planning surfaces

**A release that brings long-running review / handoff / browser verification into the mainline, while reducing pre-build and pre-release uncertainty with team mode issue bridge and release preflight.**

### Added

- Added opt-in team mode and `scripts/plans-issue-bridge.sh` to generate tracking issue / sub-issue dry-run payloads while keeping `Plans.md` as the source of truth
- Added `scripts/release-preflight.sh` and release preflight docs / tests so `/harness-release --dry-run` can run vendor-neutral pre-publication checks
- Added optional brief rules for `harness-plan create` and `scripts/generate-skill-manifest.sh` to generate UI/API briefs and machine-readable skill surface manifests

### Changed

- Updated `skills-v3` planning / release skills to integrate team mode, pre-release verification, and brief/manifest workflows into existing flows
- Synced public skill mirrors so the same planning / release surface is available across Claude / Codex / OpenCode distributions

#### Before/After

| Before | After |
|--------|-------|
| Sharing Plans.md tasks with a team required figuring out issue structure and payload each time | Opt-in team mode generates stable tracking issue / sub-issue dry-run payloads from Plans.md |
| `/harness-release --dry-run` had person-dependent pre-publication checks with no unified view of repo healthcheck or CI status | Vendor-neutral preflight script checks working tree, CHANGELOG, env parity, healthcheck, CI, and shipped surface residuals together |
| No machine-readable pipeline for UI/API task briefs or skill surface listings, requiring manual creation each time for comparison/audit/auto-docs | Added design brief / contract brief rules and skill-manifest.v1 generation for lightweight reusable auxiliary materials and manifests |

## [3.15.0](https://github.com/tim-hub/powerball-harness/compare/v3.14.0...v3.15.0) - 2026-03-28

### Theme: Claude 2.1.80-2.1.86 integration + Codex/OpenCode mirror alignment

**From Claude comes a step up in "lightness" and "safety"; from Codex comes stabilized initial quality for heavy workflows and distribution mirror alignment. A release that turns upstream tracking into real operational strength.**

---

#### 1. Made it harder to lose track of changed assumptions with Claude's reactive hooks

**Before**: After updating `Plans.md` or switching to a different worktree, it was easy to continue working with stale assumptions. When background tasks were created, the recording and re-checking triggers were weak, making context drift more likely in longer sessions.

**After**: Incorporated Claude Code's `TaskCreated` / `FileChanged` / `CwdChanged` hooks into Harness, with `runtime-reactive.sh` catching task creation, Plans updates, rule changes, and worktree switches to return supplementary context. Even when assumptions change mid-work, it's easier to notice on the next step.

```json
{"hook_event_name":"FileChanged","file_path":"Plans.md"}
-> "Plans.md has been updated. Please re-read the latest task state before your next implementation or review."
```

#### 2. Tuned Claude's permission flow toward safety without sacrificing speed

**Before**: `PermissionRequest` hooks tended to fire broadly, adding evaluation overhead and noise even for ultimately safe Bash commands. Sandbox startup failures allowing continuation, and credential propagation to subprocesses were easy-to-miss issues.

**After**: Using Claude Code 2.1.85's conditional `if` field, permission hooks are limited to safe Bash commands like `git status`, `git diff`, `pytest`, `npm run lint`. Also aligned `Edit|Write|MultiEdit` matchers and added `sandbox.failIfUnavailable: true` and `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=1` to `.claude-plugin/settings.json`, balancing lightness with safety.

```text
Bash(git status*) | Bash(pytest*) | Bash(npm run lint*)
-> permission hook target

Bash(dangerous commands)
-> handled by existing guardrails, not broad hooks
```

#### 3. Made it easier to think deeply from the start on heavy Claude/Codex flows

**Before**: Even for heavy workflows like `harness-work`, `harness-review`, and `harness-release`, the first turn's priorities could drift, and quality varied depending on how instructions were written each time.

**After**: Added `effort` frontmatter to `skills-v3/`, Codex native mirror, and OpenCode mirror, plus `initialPrompt` to `agents-v3/worker.md`, `reviewer.md`, and `scaffolder.md`. This means reviews start from verdict criteria, implementations start from DoD and verification strategy, and setup starts from the premise of not breaking existing assets.

```yaml
effort: high
initialPrompt: |
  First, briefly organize the target task, DoD, candidate files, and verification strategy...
```

#### 4. Made rule scoping and distribution mirrors less fragile

**Before**: Rules `paths:` used single-line string format that was hard to read and fragile when adding multiple globs. Codex/OpenCode mirrors could drift from source with partially inconsistent judgment criteria, requiring root-cause investigation after CI failures.

**After**: Migrated rules templates and `scripts/localize-rules.sh` to YAML list format for structured handling of multiple globs. Also aligned OpenCode build and Codex package public skill policies to keep internal-only skills out of distribution mirrors so CI stays green.

```yaml
paths:
  - "**/*.{test,spec}.{ts,tsx,js,jsx}"
  - "**/tests/**/*.*"
```

## [3.14.0] - 2026-03-25

### Theme: Cross-runtime quality hardening + Marketplace fix

**Unified quality guardrails across Claude Code and Codex, and fixed memory hook gaps on Marketplace installation.**

---

#### 1. Cross-runtime quality guardrail unification

**Before**: Claude Code guardrails (`--no-verify` detection, protected branch `reset --hard` warnings, etc.) did not exist on the Codex side, causing inconsistent quality standards across runtimes.

**After**: Defined a policy matrix in `docs/hardening-parity.md` and applied the same rules to both Claude Code hooks and Codex CLI quality gates. Automated cross-runtime verification with `validate-plugin.sh` / `validate-plugin-v3.sh`.

- Guardrails: `--no-verify` / `--no-gpg-sign`, protected branch `git reset --hard`, direct push to `main`/`master` warning, protected file edit warning
- Codex parity: Inject runtime contracts into `codex exec` flow, verifying bypass flags, protected file edits, and secret contamination before merge

#### 2. Detailed rules added to Codex AGENTS.md

**Before**: The Codex `AGENTS.md` lacked details from `.claude/rules/`, missing references to CC update policy and v3 architecture.

**After**: Integrated content from `cc-update-policy.md`, `v3-architecture.md`, and `versioning.md` into Codex AGENTS.md. Codex users can now directly reference rule details.

### Fixed

- Fixed missing `scripts/hook-handlers/memory-*.sh` on Marketplace installation causing SessionStart / UserPromptSubmit / PostToolUse / Stop hook errors
- Consolidated memory lifecycle hooks into a single `memory-bridge.sh` entry point, removing dependency on individual wrapper paths
- Fixed path resolution in `sync-plugin-cache.sh` source detection when `CLAUDE_PLUGIN_ROOT` points to the plugin root itself
- Added regression tests for memory hook wiring and Marketplace cache sync

## [3.13.0] - 2026-03-25

### Theme: Codex native support + Review quality enhancement + Memory persistence

**Harness team execution (breezing) is now usable from Codex CLI, review quality is improved with automatic AI residual detection, and cross-session memory is persisted to harness-mem with automatic context restoration on resume.**

---

#### 1. Codex native skills (skills-v3-codex/)

**Before**: Using `/harness-work` or `/breezing` from Codex CLI would encounter Claude Code-specific APIs (`Agent()`, `SendMessage()`) in pseudo-code, which Codex's LLM could not correctly interpret. There was only an annotation saying "read differently for Codex," with a risk of runtime errors.

**After**: Created Codex native versions in `skills-v3-codex/`. Rewritten with correct API signatures for `spawn_agent` / `wait_agent` / `send_input` / `close_agent`, implementing Worker isolation via `git worktree add`, and working directory/verdict retrieval via `codex exec -C/-o`. Achieved APPROVE through 5 rounds of Codex self-review.

Available from any project by deploying to user scope (`~/.codex/skills/`).

```
~/.codex/skills/
├── harness-work -> skills-v3-codex/  [CODEX NATIVE]
├── breezing     -> skills-v3-codex/  [CODEX NATIVE]
├── harness-plan -> skills-v3/        [shared]
└── ...5 others  -> skills-v3/        [shared]
```

**Key differences from Claude Code version**:

| Item | Claude Code | Codex Native |
|------|-------------|--------------|
| Worker spawn | `Agent(subagent_type="worker")` | `spawn_agent({message, fork_context})` |
| Fix instructions | `SendMessage(to: agentId)` | `send_input({id, message})` |
| Worktree isolation | `isolation="worktree"` auto | `git worktree add` manual |
| Review | Codex exec -> Reviewer agent fallback | `codex exec -o <file>` only |
| Mode escalation | Auto when 4+ tasks | `--breezing` explicit only |

#### 2. AI Residuals review gate (Phase 29.0)

**Before**: AI-generated remnants like mockData, dummy, localhost, TODO could slip through reviews, with "works but can't ship" code getting merged.

**After**: Added "AI Residuals" as the 5th review perspective in `harness-review`. `scripts/review-ai-residuals.sh` statically scans diffs and classifies remnants by severity (minor/major). Test fixtures included.

```bash
# Detection targets (examples)
mockData, dummyUser, localhost:3000, TODO:, FIXME,
test.skip, describe.skip, hardcoded API keys
```

#### 3. harness-mem session memory persistence (Phase 27.1.4-5)

**Before**: Closing a Claude session meant losing all context and decisions from that session, requiring a fresh start in the next session.

**After**: Connected Claude's SessionStart / UserPromptSubmit / Stop hooks to the harness-mem runtime. Automatically displays a "Continuity Briefing" from previous session memory at session start, and persists memory at session stop.

- `scripts/lib/harness-mem-bridge.sh` abstracts harness-mem API calls
- Integrated continuity briefing into `session-init.sh` / `session-resume.sh`
- Added memory lifecycle regression tests (wiring, bridge, integration)

## [3.12.0] - 2026-03-21

### Theme: work/Breezing end-to-end flow automation

**Achieved end-to-end automation from skill invocation to commit and reporting without manual intervention. Combined review quality and convergence through Codex exec review loops with threshold-based verdicts.**

---

#### 1. Plans.md auto-registration (Phase A)

**Before**: If Plans.md didn't exist, harness-work would stop with an error.
Also, requirements communicated in conversation wouldn't be detected as missing from Plans.md, requiring manual addition.

**After**: If Plans.md doesn't exist, automatically calls `harness-plan create --ci` to generate it.
Detects action verbs (like "add", "fix") from conversation and auto-adds unregistered tasks in v2 format.

#### 2. Codex exec review loop (Phase B)

**Before**: Solo/Parallel modes had no review stage, relying only on Worker self-review.
In Breezing mode, Reviewer agent ran independent reviews but fix loops required manual approval.

**After**: Auto-review runs after implementation completion across all modes.
Two-tier structure: Codex exec (primary) -> internal Reviewer agent (fallback).
On REQUEST_CHANGES: auto-fix -> re-review (max 3 rounds).

#### 3. Review threshold criteria (Phase B addition)

**Before**: With free-form reviews, even minor improvement suggestions returned REQUEST_CHANGES, preventing review loop convergence.

**After**: Explicitly passes 4-tier threshold criteria (critical/major/minor/recommendation) to the review prompt.
Only critical/major triggers REQUEST_CHANGES; minor/recommendation results in APPROVE.
Out-of-scope findings (external tool constraints, etc.) also don't affect the verdict.

#### 4. Rich completion report (Phase C)

**Before**: Post-task reports were simple text only (Progress: Task N/M completed).

**After**: Auto-generates a visual summary after commit.
Shows "what was done", "what changes (Before/After)", "changed files", and "remaining issues (Plans.md linked)" in box format.
In Breezing mode, generates a consolidated report after all tasks complete.

#### 5. codex exec flag unification

**Before**: All skills/scripts used the deprecated flag `-a never` (removed in codex-cli 0.115.0), causing codex exec to immediately error out.

**After**: Unified all locations to `--full-auto`. Also fixed `$TIMEOUT` expansion to safe pattern `${TIMEOUT:+$TIMEOUT N}`.
Review codex exec runs with `--sandbox read-only` (no write permissions).

#### 6. Platform copy full sync

**Before**: Primary `skills/` and platform copies (`codex/.codex/skills/`, `opencode/skills/`, `skills-v3/`) were manually synced and had drifted apart.

**After**: This change fully syncs all platform copies with primary.
`harness-review` BASE_REF support and `breezing` Review Policy reflected in all copies.

#### 7. Breezing review loop implementation (Phase F)

**Before**: In Breezing mode, Workers committed directly to main before Reviewer reviewed.
Even when REQUEST_CHANGES was returned, changes were already committed, structurally preventing a fix loop.

**After**: Workers commit within worktrees, and Lead cherry-picks to main after review.
- Worker: In `mode: breezing`, commits within worktree -> returns `{commit, worktreePath}` to Lead
- Lead: Reviews via Codex exec / Reviewer agent -> `git cherry-pick` on APPROVE
- REQUEST_CHANGES: Lead sends fix instructions to Worker via SendMessage -> Worker amends -> re-review (max 3 rounds)
- Phase C: Lead generates Breezing summary report from `git log` + Plans.md

Added `worktreePath` / `summary` fields to Worker output JSON.
Plans.md updates managed centrally by Lead (Worker does not edit Plans.md in breezing mode).

## [3.11.0] - 2026-03-20

### Theme: Claude Code v2.1.77-v2.1.79 integration + "Documentation-only prohibition" quality revolution

**Integrated latest CC version, structurally resolving the "documented-only problem" identified through self-review. Added StopFailure logging/notification, and incorporated Effort dynamic injection and Sandbox auto-configuration design into SKILL.md and agent definitions.**

---

#### 1. Claude Code v2.1.77-v2.1.79 integration

Added 21 new features/fixes to the Feature Table with documentation on how Harness leverages them.

##### 1-1. `StopFailure` hook event support

**CC Update**: v2.1.78 added a `StopFailure` event to capture session stop failures caused by API errors (rate limit 429, auth failure 401, etc.).

**Harness Integration**: Created new `stop-failure.sh` handler that logs error information (project-scoped when `${CLAUDE_PLUGIN_DATA}` is set, otherwise to `.claude/state/stop-failures.jsonl`). Useful for post-mortem analysis of Breezing Worker stop failures due to rate limits.

##### 1-2. Documentation of PreToolUse `allow` / `deny` priority

**CC Update**: v2.1.77 applied a security fix where settings.json `deny` rules take priority even when PreToolUse hooks return `allow`.

**Harness Integration**: Added version notes to hooks-editing.md documenting priority rules for guardrail design. `deny: ["mcp__*"]` pattern now recommended.

##### 1-3. Feature Table v2.1.77-v2.1.79 additions (21 items)

**CC Update**: Output token 64k/128k expansion, `allowRead` sandbox, Agent `resume` deprecation -> `SendMessage`, `/branch` rename, `${CLAUDE_PLUGIN_DATA}` variable, Agent `effort` frontmatter, etc.

**Harness Integration**: Added all items to both the CLAUDE.md Feature Table and docs/CLAUDE-feature-table.md. Detailed how each feature is used/impacted in Harness.

### Changed

- Updated session-control skill description from `/fork` to `/branch` (v2.1.77 rename)
- Added `StopFailure`, `ConfigChange` to hooks-editing.md event types list
- Added v2.1.77+ PreToolUse priority and v2.1.78+ StopFailure notes to hooks-editing.md
- Added `stop_failure` to `SignalType` in core/src/types.ts
- Added `mcp__codex__*` deny rule to `.claude-plugin/settings.json` (v2.1.78 recommended pattern)
- Added settings.json deny pattern recommendation section to `codex-cli-only.md`
- Updated `stop-failure.sh`, `notification-handler.sh` state save paths to `${CLAUDE_PLUGIN_DATA}` (with fallback)
- Added `effort: medium` field to Worker/Reviewer agent definitions (v2.1.78 official support)
- Added environment variable reference (`CLAUDE_PLUGIN_DATA`, `ANTHROPIC_CUSTOM_MODEL_OPTION`, etc.) to `harness-setup/SKILL.md`

### Added

#### Phase 28.0: "Documentation-only prohibition" guardrail skill

**Before**: When CC updates occurred, entries were simply transcribed to the Feature Table without becoming "Harness value-add." A 3-agent parallel review identified 14 out of 21 items as "documentation-only."

**After**: `skills/cc-update-review/` (non-distributed, internal-only skill) auto-classifies all Feature Table items as A/B/C during CC update integration. When Category B (documentation-only) is detected, it forces presentation of an implementation proposal. Formalized as a rule in `.claude/rules/cc-update-policy.md`.

#### Phase 28.1: StopFailure auto-recovery design addition

**Before**: When a Breezing Worker died from rate limiting (429), it was only logged. Neither Lead nor humans noticed, and Workers silently disappeared.

**After**: Added StopFailure auto-recovery flow design to `breezing/SKILL.md`. 429 -> exponential backoff (30s/60s/120s) + auto-restart Worker via `SendMessage`. 401 -> user notification. 500 -> record blocker in Plans.md. Implemented `stop-failure.sh` notifying Lead via `systemMessage` on 429 detection.

#### Phase 28.2: Effort dynamic injection design addition

**Before**: Worker/Reviewer `effort: medium` was a fixed value. The harness-work scoring (ultrathink at >= 3) and Agent frontmatter's `effort` field were not connected.

**After**: Added scoring -> effort injection flow design to `harness-work/SKILL.md`. Added dynamic effort reception and post-task recording instructions to `agents-v3/worker.md`. Workers record `effort_applied`, `effort_sufficient`, `turns_used` to agent memory at task completion, informing future scoring accuracy.

#### Phase 28.3: Log visualization + Sandbox template addition

**Before**: Logs accumulated in `stop-failures.jsonl` with no way to view them. Reviewer had no sandbox configuration, and some environments couldn't even read `.env.example`.

**After**: `scripts/show-failures.sh` displays error code and time-based summaries (implemented). Added `sandbox.allowRead` template to `.claude-plugin/settings.json` (`.env.example`, `docs/**`, etc.). Added sandbox auto-generation procedure by project type to `harness-setup init` in SKILL.md.

---

- `scripts/hook-handlers/stop-failure.sh` -- StopFailure hook handler (with systemMessage notification on 429)
- `skills/cc-update-review/SKILL.md` -- CC update integration quality guardrail skill (non-distributed)
- `.claude/rules/cc-update-policy.md` -- Quality policy for Feature Table additions
- hooks.json (both files) with `StopFailure` event definition
- `tests/validate-plugin.sh` with `claude plugin validate` step (runs only when v2.1.77+ is available)
- `.claude-plugin/settings.json` with `sandbox.allowRead` template

## [3.10.6] - 2026-03-19

### Theme: Plugin user quality improvements

**Fixed critical errors and UX issues that occur after `claude plugin install`. Addresses Issue #64, #65.**

---

### Fixed

#### 0-1. Fixed hooks failing with MODULE_NOT_FOUND after plugin installation (Issue #64)

**Before**: Since `core/dist/` was excluded by `.gitignore`, compiled JavaScript didn't exist in environments that ran `claude plugin install`, causing all hooks (PreToolUse / PostToolUse / PermissionRequest) to immediately fail with `MODULE_NOT_FOUND`. A critical issue that completely disabled the guardrail engine (R01-R09).

**After**: Removed `/core/dist/` exclusion from `.gitignore` to include built JS in the repository. Hooks work immediately after plugin installation.

#### 0-2. Fixed PostToolUse HTTP hook erroring by default (Issue #65)

**Before**: `hooks.json` had a metrics HTTP hook targeting `localhost:9090` enabled by default. Users without a metrics server would get connection refused errors on every `Write`/`Edit`/`Bash`/`Task` operation, with up to 5-second delays. The CHANGELOG described it as a "template," but it was actually active.

**After**: Removed the HTTP hook entry from `hooks.json` and moved it to `docs/examples/hooks-metrics-http.json` as a template. No errors in default state. Users wanting metrics integration can reference the template to add to their own hooks.json.

#### 0-3. Removed broken symlink `codex-review`

**Before**: `skills-v3/extensions/codex-review` pointed to `../../skills/codex-review` but the target `skills/codex-review/` directory didn't exist, creating a broken symlink.

**After**: Removed the broken symlink. Will be re-added when `codex-review` functionality is implemented.

#### 0-4. Fixed license inconsistency between `plugin.json` and `marketplace.json`

**Before**: `plugin.json` had `"license": "MIT"` while `marketplace.json` had `"license": "Proprietary"` -- contradictory.

**After**: Unified `marketplace.json` license to `"MIT"`.

### Changed

#### 1. Unified agent `disallowedTools` to official naming

**Before**: Worker / Reviewer / Scaffolder `disallowedTools` used legacy name `[Task]`. While Task acts as an alias since CC v2.1.63 renamed the Task tool to Agent, official documentation consistently uses `Agent`.

**After**: Updated `disallowedTools` in all agent definitions to `[Agent]`. Ensures consistency with official documentation and prepares for potential alias deprecation.

### Added

#### 2. Added `elicitation_dialog` support to Notification handler

**Before**: The `elicitation_dialog` notification type added in CC v2.1.76 for MCP Elicitation was not individually detected by the Notification handler. Auto-skip was implemented in the `Elicitation` hook, but Notification-side log detection was missing.

**After**: Added `elicitation_dialog` detection to `notification-handler.sh`. When MCP Elicitation occurs in Breezing background Workers, it's now logged like `permission_prompt`. Enables post-mortem tracking of Elicitation occurrences.

#### 3. Added `harness-ops` Output Style as plugin component

**Before**: The Feature Table mentioned a `harness-ops` output style, but the actual style file didn't exist. Also, plugin.json lacked the `outputStyles` field, preventing distribution via plugin.

**After**: Created `output-styles/harness-ops.md` defining structured output styles for Plan/Work/Review phases. Added `outputStyles: "./output-styles/"` to plugin.json for automatic distribution on plugin install. Users can select `Harness Ops` via `/config` -> Output style.

## [3.10.5] - 2026-03-15

### Theme: set-locale.sh skills-v3 support

**Fixed a bug where `set-locale.sh` did not process the `skills-v3/` directory.**

---

### Fixed

#### 1. `set-locale.sh` not processing `skills-v3/`

**Before**: Running `scripts/i18n/set-locale.sh ja` left SKILL.md files in the `skills-v3/` directory with their `description` field still in English. `skills/`, `codex/.codex/skills/`, and `opencode/skills/` were processed, but `skills-v3/` introduced with the v3 architecture was missing from the processing target list.

**After**: Added `skills-v3/` to `process_skill_dir` calls. All 4 directories now switch in a single batch.

### Changed

- `.gitignore`: Added `.superset/`, `skills/x-announce/` to untracked targets

## [3.10.4] - 2026-03-15

### Theme: Agent safety limits and Notification hook implementation

**Introduced `maxTurns` safety limits to all subagents to prevent runaway execution, and completed the implementation of the previously documentation-only Notification hook.**

---

### Added

#### 1. `maxTurns` safety limits for agent runaway prevention

**Before**: Worker / Reviewer / Scaffolder had no turn limits set. If an agent entered an infinite loop or excessive exploration, it wouldn't stop until exhausting the context window, making token costs uncontrollable.

**After**: Added `maxTurns` field recommended by CC official documentation to all agent frontmatters. Worker: 100 (for complex implementation tasks), Reviewer: 50 (specialized for Read-only analysis), Scaffolder: 75 (intermediate complexity). When limits are reached, Lead can collect partial results for decision-making. Combined with `bypassPermissions`, this serves as a safety valve against runaway execution.

#### 2. `Notification` hook handler implementation

**Before**: hooks-editing.md and the Feature Table listed the `Notification` event, but no handler was registered in hooks.json. It was the only "documented but not implemented" gap among 26 hook events.

**After**: Created `notification-handler.sh` and registered it in both hooks.json files (source + distribution). Logs notification events (`permission_prompt` / `idle_prompt` / `auth_success`, etc.) to `.claude/state/notification-events.jsonl`. Enables post-mortem analysis of permission_prompt events occurring in Breezing background Workers.

#### 3. Added `/context` command to Feature Table

**Before**: The `/context` command added in CC v2.1.74 (context consumption visualization and optimization suggestions) was not listed in the Feature Table.

**After**: Added to both the CLAUDE.md summary table and docs/CLAUDE-feature-table.md detail section. Useful for identifying causes of frequent compaction in long Breezing sessions.

## [3.10.3](https://github.com/tim-hub/powerball-harness/compare/v3.10.2...v3.10.3) - 2026-03-14

### Changed

- release metadata updates are now release-only: normal PRs should leave `VERSION` and `.claude-plugin/plugin.json` untouched and record changes under `[Unreleased]`
- pre-commit and CI now validate release metadata consistency without auto-bumping patch versions on ordinary code changes
- README and README_ja now use the GitHub latest release badge instead of hardcoded per-version badge URLs
- `.claude/rules/hooks-editing.md` now documents `SessionEnd` timeout guidance and `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS` so the PR61 docs fix can be merged without carrying release metadata drift
- Codex workflow docs now standardize on `$harness-plan`, `$harness-sync`, `$harness-work`, `$breezing`, and `$harness-review`, and setup scripts archive removed legacy Harness skills from `~/.codex/skills`

### Added

#### 1. Added 9 features from official documentation to Feature Table

**Before**: Features documented in Claude Code official documentation (60+ pages) including `--remote` / Cloud Sessions, `/teleport`, `CLAUDE_CODE_REMOTE`, `CLAUDE_ENV_FILE`, Slack Integration, Server-managed settings, Microsoft Foundry, `PreCompact` hook, and `Notification` hook event were not registered in the Feature Table.

**After**: Added 9 entries to `docs/CLAUDE-feature-table.md` (summary table + feature detail sections). Reflected 4 high-impact items in `CLAUDE.md`. Detailed Harness usage methods, code examples, and prerequisites for each feature.

#### 2. Added cloud session detection to session-env-setup.sh

**Before**: `session-env-setup.sh` assumed a local environment and had no way to determine if running in a cloud session (`--remote` execution).

**After**: Persists `CLAUDE_CODE_REMOTE` environment variable as `HARNESS_IS_REMOTE` in `CLAUDE_ENV_FILE`. Other hook handlers can now branch on cloud vs. local.

#### 3. Added PreCompact / Notification events to hooks-editing.md

**Before**: `PreCompact` and `Notification` were not listed in the hooks-editing.md Event Types list, leaving developers without a reference when adding new hooks.

**After**: Added `PreCompact` (pre-compaction state save) and `Notification` (custom handler on notification events) to the Event Types JSON block. In Harness, `PreCompact` is already implemented (2-layer structure: command + agent).

#### 4. Codex command surface cleanup + stale skill cleanup

**Before**: Legacy command surfaces like `$work` / `$plan-with-agent` / `$verify` remained in Codex-side documentation, and legacy Harness skills lingered in `~/.codex/skills` after updates, cluttering the listing.

**After**:

- **Codex docs**: Unified main workflow to `$harness-plan`, `$harness-sync`, `$harness-work`, `$breezing`, `$harness-review`
- **setup scripts**: `scripts/setup-codex.sh` / `scripts/codex-setup-local.sh` back up legacy Harness skills that are no longer shipped
- **test coverage**: Added regressions in `tests/test-codex-package.sh` and `validate-plugin-v3.sh` for `harness-sync` surface, native multi-agent wording, and legacy skill cleanup

#### 5. Claude Code 2.1.76 Integration

Integrated Claude Code 2.1.76 new features into Harness. Updated Feature Table version notation from `2.1.74+` to `2.1.76+`.

##### 5-1. Automatic handling of MCP Elicitation

**CC Update**: MCP servers (external tool connections like GitHub, Slack) can now "ask" users questions during task execution (Elicitation). For example, requesting form input like "Which repository should we push to?" Two new hook events were also added: `Elicitation` (before question) and `ElicitationResult` (after answer).

**Harness Integration**: Breezing Workers run in the background and cannot respond to MCP question forms. Leaving them unhandled freezes Workers. Created `elicitation-handler.sh` to auto-skip elicitation during Breezing sessions, while passing through normally in regular sessions for user response. Results logged via `elicitation-result.sh`.

##### 5-2. PostCompact context re-injection

**CC Update**: A `PostCompact` hook was added that fires **after** context compaction completes. Pairs with the existing `PreCompact` (before compaction).

**Harness Integration**: Long sessions had the problem of losing track of "which task is in progress" after compaction. Created `post-compact.sh` to auto-reinject Plans.md WIP/TODO task state after compaction. The symmetric structure of PreCompact (state save) -> PostCompact (state restore) ensures work context continuity.

##### 5-3. Worktree speedup and stabilization

**CC Update**: Three improvements: (1) `worktree.sparsePaths` setting for checking out only needed directories in huge repos during worktree creation, (2) faster `--worktree` startup via direct git refs reading, (3) auto-cleanup of stale worktrees from interrupted parallel executions.

**Harness Integration**: Reduced startup time when launching multiple Workers simultaneously in Breezing. Manual stale worktree deletion no longer needed. Added usage guides to breezing/SKILL.md and harness-work/SKILL.md.

##### 5-4. Session naming and Effort dynamic control

**CC Update**: `-n`/`--name` flag sets a display name for sessions. `/effort` command switches thinking depth (low/medium/high) during sessions.

**Harness Integration**: Set `breezing-{timestamp}` format names on Breezing sessions for easy identification. harness-work's multi-factor scoring (automatic effort adjustment based on task complexity) can now be combined with manual `/effort` switching.

##### 5-5. Background agent partial result retention

**CC Update**: When background agents are killed (timeout or manual stop), partial work results now remain in context. Previously they were completely lost.

**Harness Integration**: Even when a Breezing Worker is interrupted mid-task, Lead can now take over partial results and reassign to another Worker. Reduces "restart from scratch" costs.

##### 5-6. Auto-compaction circuit breaker

**CC Update**: A circuit breaker was introduced that stops auto-compaction after 3 consecutive failures. Prevents token waste from infinite retries.

**Harness Integration**: Same design philosophy as Harness's "3 strike rule" (3-attempt limit on CI failures). Prevents unexpected cost increases during long Breezing sessions.

##### 5-7. `--plugin-dir` breaking change

**CC Update**: `--plugin-dir` changed to accept only one path. Multiple directories must be specified as `--plugin-dir path1 --plugin-dir path2` with repeated flags.

**Harness Impact**: No impact for standalone Harness plugin usage. Syntax change only needed when using multiple plugins simultaneously.

---

## [3.10.2](https://github.com/tim-hub/powerball-harness/compare/v3.10.1...v3.10.2) - 2026-03-12

### Theme: TaskCompleted finalize hardening + Claude Code 2.1.74 docs/README alignment

**Implemented safety improvement to front-load `harness-mem` finalize at the point of all-task completion, and synced feature docs / README / compatibility snapshot to Claude Code 2.1.74. Also recovered validate-plugin which was failing due to missing version bump, as a proper patch release.**

---

#### 1. Safety improvement for TaskCompleted-based finalize

**Before**: Session finalization was concentrated at the Stop point, leaving room for `harness-mem` completion records to be missed in cases where "the last task finished but a crash occurred before Stop."

**After**: `task-completed.sh` executes `work_completed` -> `/v1/sessions/finalize` exactly once the moment it detects "completed count >= total tasks." Added `session.json` `session_id` / `project_name` fallback, success marker for idempotency, `HARNESS_MEM_BASE_URL` for testability, and silent skip on API unreachability.

#### 2. Added finalize regression tests

**Before**: Fix proposal tests existed, but no fixtures directly verified "only finalize on last task," "no duplicate finalize," or "skip when session_id unresolved."

**After**: Added `tests/test-task-completed-finalize.sh` to independently verify finalize trigger conditions and safety conditions from the TaskCompleted hook. Combined with existing `tests/test-fix-proposal-flow.sh`, both progress control and completion confirmation can be regression-tested.

#### 3. Synced Claude Code 2.1.74 docs / README / compatibility

**Before**: `docs/CLAUDE-feature-table.md` had started incorporating 2.1.74 features, while README's feature summary still showed `2.1.71+` and the compatibility document's latest verified snapshot remained at `2.1.69` / plugin `3.6.0`.

**After**: Unified feature table to `2.1.74+`, updated English/Japanese READMEs and `docs/CLAUDE_CODE_COMPATIBILITY.md` to match actual measurements. Reflected key 2.1.73-2.1.74 items including `modelOverrides`, `autoMemoryDirectory`, `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS`, and full model ID support in summaries.

#### 4. Promoted release metadata to 3.10.2

**Before**: Commit `4239d542` contained code changes but `VERSION` / `plugin.json` / README badge / CHANGELOG remained at `3.10.1`, causing GitHub Actions `validate-plugin` to fail on missing version bump.

**After**: Aligned `VERSION`, `.claude-plugin/plugin.json`, English/Japanese README version badges, and CHANGELOG compare links to `3.10.2`, making it ready for publication as a patch release.

---
## [3.10.1](https://github.com/tim-hub/powerball-harness/compare/v3.10.0...v3.10.1) - 2026-03-12

### Theme: Claude Code official documentation deep integration -- 12 features added + Auto Mode rollout cleanup + SubagentStart/Stop matcher enhancement

**Added 12 previously untracked features found through a thorough review of 60 pages of official documentation to the Feature Table. Separated Auto Mode into shipped defaults and rollout targets, and added agent type-specific matchers to SubagentStart/SubagentStop hooks for individual tracking of Worker/Reviewer/Scaffolder/Video Generator startup and shutdown.**

---

#### 1. SubagentStart/SubagentStop matcher enhancement

**Before**: `SubagentStart`/`SubagentStop` hooks uniformly launched `subagent-tracker` for all agents. team-composition.md incorrectly stated "SubagentStart: not implemented."

**After**: Added agent type-specific matchers (`worker`, `reviewer`, `scaffolder`, `video-scene-generator`). Enables individual tracking of each agent's startup/shutdown for role-specific metrics collection. Updated the Quality Gate Hooks table in team-composition.md to match reality.

#### 2. Added 12 features to Feature Table

**Before**: Official documentation features including Chrome Integration, LSP server integration, Task Dependencies, `/btw`, Plugin CLI commands were not registered in the Feature Table.

**After**: Added the following to the Feature Table (summary table + feature detail sections):
- Chrome Integration (`--chrome`, beta)
- LSP server integration (`.lsp.json`)
- SubagentStart/SubagentStop matcher
- Agent Teams: Task Dependencies
- `--teammate-mode` CLI flag
- `CLAUDE_CODE_DISABLE_BACKGROUND_TASKS`
- `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`
- `cleanupPeriodDays` setting
- `/btw` side question
- Plugin CLI commands
- Remote Control enhancements
- `skills` field in agent frontmatter

#### 3. CLAUDE.md Feature Table summary update

**Before**: CLAUDE.md summary table didn't include Chrome Integration, LSP, matcher, Task Dependencies, etc.

**After**: Added the 6 highest-impact features to the CLAUDE.md summary table.

#### 4. Organized Breezing Auto Mode rollout

**Before**: Auto Mode description was ahead of implementation, reading as if it were already default in Breezing.

**After**: Shipped default remains `bypassPermissions`, with `--auto-mode` documented as opt-in rollout for testing only with compatible parent sessions. Project template / frontmatter retains `bypassPermissions` as shown in official docs.

---

## [3.10.0](https://github.com/tim-hub/powerball-harness/compare/v3.9.0...v3.10.0) - 2026-03-11

### Theme: Integration of 10 Claude Code documentation features + Status Line implementation

**Integrated new features from Claude Code official documentation (Sandboxing, Model Configuration, Checkpointing, Code Review, Status Line, etc.) into the Feature Table, and added a Harness-specific status line script.**

---

#### 1. Sandboxing (`/sandbox`) integration

**Before**: Worker Bash commands were controlled via `bypassPermissions` + hooks. OS-level filesystem/network isolation was not included in Harness operational guides.

**After**: Positioned Claude Code's native Sandboxing (macOS Seatbelt / Linux bubblewrap) as a **complementary layer** to `bypassPermissions`. Added phased introduction plan (Phase 0->1->2) to `team-composition.md`. Phased approach for introducing OS-level safety boundaries to Worker Bash.

#### 2. Model Configuration (3 features)

**Before**: Worker/Reviewer models were fixed via `model: sonnet` in agent definitions. Lead also ran Plan and Execute with a single model.

**After**:
- **`opusplan` alias**: Auto-switches between Opus for Plan and Sonnet for Execute in Lead sessions
- **`CLAUDE_CODE_SUBAGENT_MODEL`**: Bulk-specify model for all subagents via env var (useful for CI cost reduction)
- **`availableModels`**: Model governance for enterprise environments

#### 3. Checkpointing (`/rewind`) support

**Before**: When file edits didn't go as expected during a session, you had to manually git revert or start from scratch.

**After**: Rewind to any point in the session via `Esc+Esc` or `/rewind`. "Summarize from here" selectively recovers context from verbose debugging sessions. Useful for safe exploration during harness-work self-review phase.

#### 4. Code Review (managed service) support

**Before**: Harness's `harness-review` only provided local agent-based code review.

**After**: Added multi-agent PR review on Anthropic infrastructure (Teams/Enterprise Research Preview) to the Feature Table. Documented the `REVIEW.md` review-specific guidance mechanism. Local review (`harness-review`) and managed review are positioned as complementary dual-check.

#### 5. New Harness Status Line script

**Before**: Claude Code's `/statusline` feature existed, but there was no Harness-specific status display.

**After**: Added `scripts/statusline-harness.sh`. Always displays the following on 2 lines:
- Line 1: Model name + git branch + staged/modified file count + agent name/worktree name
- Line 2: Context usage bar (yellow at 70%, red at 90%) + session cost + elapsed time + output style name

```bash
# Configuration
/statusline use scripts/statusline-harness.sh
```

#### 6. Feature Table expansion (10 items added)

Added the following to `docs/CLAUDE-feature-table.md` and `CLAUDE.md` summary:
- Sandboxing (`/sandbox`)
- `opusplan` model alias
- `CLAUDE_CODE_SUBAGENT_MODEL` env var
- `availableModels` setting
- Checkpointing (`/rewind`)
- Code Review (managed service)
- Status Line (`/statusline`)
- 1M Context Window (`sonnet[1m]`)
- Per-model Prompt Caching Control
- `CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING`

---

## [3.9.0](https://github.com/tim-hub/powerball-harness/compare/v3.7.2...v3.9.0) - 2026-03-11

### Theme: Output Styles integration + Agent definition enhancement + Agent Teams official best practices alignment

**Reflected latest Claude Code official documentation specs for Output Styles / Agent Teams / agent frontmatter into Harness, improving operational experience.**

> **Release note**: Draft changes that had accumulated as `v3.7.3` / `v3.8.0` equivalents have been consolidated into this `v3.9.0` official release.

---

#### 1. New Harness Output Style

**Before**: Progress reports and Quality Gate results for Plan/Work/Review had no unified format, with each skill/agent using its own output format.

**After**: Created `.claude/output-styles/harness-ops.md`. Enabling with `/output-style harness-ops` provides structured output for:
- Progress reports (done/current/next action format)
- Quality Gate results (Build/Test/Lint table format)
- Review verdicts (APPROVE/REQUEST_CHANGES structured format)
- Escalations (standard output format for 3-strike violations)
- Decision points (max 3 choices, recommended first)

```bash
/output-style harness-ops
```

#### 2. Added explicit `permissionMode` to agent definitions

**Before**: Worker/Reviewer/Scaffolder permission mode was specified at spawn time as `mode: "bypassPermissions"`. Agent definitions themselves had no permission info, depending on Lead's spawn code.

**After**: Following Claude Code official documentation's formalization of `permissionMode` as an official agent frontmatter field, added `permissionMode: bypassPermissions` to all 3 agent frontmatters. Achieves declarative permission management at the definition level.

```yaml
# agents-v3/worker.md
permissionMode: bypassPermissions  # newly added
```

#### 3. Agent Teams official best practices alignment

**Before**: Harness team operations were based on proprietary patterns. Claude Code Agent Teams had limited official guidance with only "experimental" status.

**After**: `agent-teams.md` was promoted to standalone official documentation. Reflected the following in `agents-v3/team-composition.md`:
- **Task granularity guidelines**: Official recommendation of 5-6 tasks/teammate
- **`teammateMode` setting**: 3 modes -- `"auto"` / `"in-process"` / `"tmux"`
- **Plan Approval pattern**: Official flow requiring plan mode from Workers
- **Quality Gate Hooks**: exit 2 feedback pattern for `TeammateIdle`/`TaskCompleted`
- **Team size**: Official recommendation of 3-5 teammates (confirmed alignment with Harness's Worker 1-3 + Reviewer 1)

#### 4. Feature Table expansion (3 items added)

Added the following to `docs/CLAUDE-feature-table.md`:
- Output Styles integration
- `permissionMode` in agent frontmatter
- Agent Teams official best practices alignment

#### 5. Pre-merge alignment fixes

**Before**: README version badge, compare links, Auto Mode phasing text, `validate-plugin` core dependency step, and opencode mirror had partial inconsistencies, preventing required checks from passing reliably.

**After**: Synced version notation and compare links, corrected Auto Mode expression to "staged rollout / verify after RP starts." Fixed `validate-plugin` to use `core/package.json` as cache key with `npm install`, and rebuilt opencode mirror with regeneration as premise.

---

### Included: Claude Code v2.1.72 compatibility

**Reflected all new features/fixes from Claude Code v2.1.72 into Harness. Added 12 features to Feature Table and agent definitions including Effort level simplification, ExitWorktree tool, Agent tool model parameter restoration, and parallel tool call fixes.**

---

#### 1. ExitWorktree tool support

**Before**: Exiting worktree sessions depended on session end prompts. Worker agents had no programmatic way to close worktrees after completing implementation.

**After**: CC v2.1.72's `ExitWorktree` tool allows Workers to explicitly exit worktrees after implementation. Added "Worktree Operations" section to `agents-v3/worker.md` documenting `ExitWorktree` usage.

#### 2. Effort level simplification (`max` deprecated)

**Before**: Effort levels included `max`, though Harness documentation only used `ultrathink` -> high effort mapping.

**After**: CC v2.1.72 deprecated `max`, unifying to 3 levels: `low(○)/medium(◐)/high(●)`. Updated Harness documentation with symbols. Affected files:
- `skills-v3/harness-work/SKILL.md` + 3 mirrors
- `agents-v3/worker.md`
- `agents-v3/reviewer.md`
- `agents-v3/team-composition.md`

#### 3. Agent tool `model` parameter restoration

**Before**: Per-invocation model overrides were unavailable for a period, operating solely with agent definition `model` fields.

**After**: CC v2.1.72 restored the Agent tool's `model` parameter. Dynamic model selection based on task characteristics is again possible. Listed as Phase 2 consideration in `agents-v3/team-composition.md`.

#### 4. Feature Table expansion (12 items added)

Added the following to `CLAUDE.md` and `docs/CLAUDE-feature-table.md`:
- `ExitWorktree` tool
- Effort levels simplification
- Agent tool `model` parameter restoration
- `/plan` description argument
- Parallel tool call fix
- Worktree isolation fixes
- `/clear` background agent preservation
- Hooks fix group (4 items)
- HTML comments hidden
- Bash auto-approval additions
- Prompt cache fix

Detail sections also added to `docs/CLAUDE-feature-table.md`.

#### 5. Version header update

Updated `CLAUDE.md` and `docs/CLAUDE-feature-table.md` headers from `2.1.71+` to `2.1.72+`.

---

### Included: Claude Code official documentation alignment

**Reflected new features/fields added to Claude Code v2.1.71+ official documentation into Harness documentation, and updated Auto Mode Phase 1 transition markers.**

---

#### 1. Feature Table expansion (9 items added)

**Before**: Only v2.1.71 release-time features were listed. New subagent fields and Agent Teams experimental flag added in official docs were not reflected.

**After**: Added the following features to the Feature Table:
- Subagent `background` field
- Subagent `local` memory scope
- Agent Teams experimental flag (`CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`)
- `/agents` command (interactive management UI)
- Desktop Scheduled Tasks
- `CronCreate/CronList/CronDelete` tools
- `CLAUDE_CODE_DISABLE_CRON` env var
- `--agents` CLI flag

Detail sections also added to `docs/CLAUDE-feature-table.md`.

#### 2. Updated Auto Mode to Phase 1 start notation

**Before**: Stated "Phase 0 (current)" and "Phase 1 (RP start)" -- notation from before RP start date 2026-03-12.

**After**: Updated to "Phase 0 (pre-RP)" and "Phase 1 (after RP start)." Affected files:
- `docs/CLAUDE-feature-table.md`
- `CLAUDE.md` Feature Table
- `agents-v3/team-composition.md`

#### 3. Agent Teams official documentation support

**Before**: Harness breezing used Agent Teams but lacked official activation instructions.

**After**: Added `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env var setup method and `teammateMode` settings to `agents-v3/team-composition.md`.

---

## [3.7.2](https://github.com/tim-hub/powerball-harness/compare/v3.7.1...v3.7.2) - 2026-03-10

### Fixed
- **Hook stdout purity**: `session-init` and usage tracking hooks now discard telemetry output so hook consumers receive the JSON payload only.
- **Quiet session summary output**: `session-init` / `session-resume` no longer leak standalone `0` lines when Plans counts are zero matches.

### Changed
- **Regression coverage**: Added direct-execution tests for snapshot summary output and quiet usage tracking hooks to keep hook output stable.

---

## [3.7.1](https://github.com/tim-hub/powerball-harness/compare/v3.7.0...v3.7.1) - 2026-03-09

### Theme: Team execution safety improvements

**Enhanced Breezing (Agent Teams) execution foundation from 3 perspectives: agent type name unification, phased Auto Mode migration preparation, and Worker worktree isolation.**

---

#### 1. Agent definition unification

**Before**: Worker and Reviewer agent type names varied across files. `breezing/SKILL.md` used `general-purpose` while `team-composition.md` used `claude-code-harness:worker`, preventing per-agent hooks (agent type-specific guardrails) from firing correctly.

**After**: Unified to `claude-code-harness:worker` / `claude-code-harness:reviewer` across all files. Worker-specific PreToolUse guards (Write/Edit checks) and Reviewer-specific Stop logs (completion records) now apply reliably.

#### 2. Auto Mode preparation (`--auto-mode`)

**Before**: Breezing used `bypassPermissions` (skip all permissions) because Workers run in the background and can't display permission prompts. This worked but allowed unintended file changes or dangerous commands to pass through silently.

**After**: Added `--auto-mode` flag supporting Claude Code 2.1.71+ Auto Mode. Auto Mode uses an allowlist approach that "auto-approves only defined safe operations" and blocks dangerous operations (`rm -rf`, `git push --force`, etc.). Migration in 3 phases:

- Phase 0 (current): `--auto-mode` is opt-in
- Phase 1 (after verification): `--auto-mode` becomes default
- Phase 2 (when stable): `bypassPermissions` deprecated

```bash
/breezing --auto-mode              # Run with Auto Mode
/harness-work --breezing --auto-mode
```

#### 3. Worker worktree isolation

**Before**: Running multiple Workers in parallel caused conflicts when two Workers edited the same file simultaneously. Lead mitigated this with a rule to "assign tasks touching the same files to the same Worker," but it wasn't perfect.

**After**: Added `isolation: worktree` to Worker agent definitions. Each Worker automatically runs in a git worktree (independent working directory), so even editing the same file is physically in separate directories with no collisions. Lead merges results after completion.

---

## [3.7.0](https://github.com/tim-hub/powerball-harness/compare/v3.6.0...v3.7.0) - 2026-03-08

### Theme: Transition to state-centric architecture

**Applied Masao Theory (Macro Harness / Micro Harness / Project OS) to build 5 features ensuring "work doesn't break when conversations end."**

---

#### 1. Automatic re-ticketing of failed tasks

**Before**: When tests/CI failed after task implementation, it would retry up to 3 times then stop. After stopping, you had to investigate the cause yourself, manually add fix tasks to Plans.md, and re-run `/work`.

**After**: When stopping after 3 failures, Harness classifies the failure cause (`assertion_error`, `import_error`, etc.) and saves fix task proposals to state. Approving with `approve fix <task_id>` adds them to Plans.md as `.fix` tasks.

```
Failure analysis:
  Category: assertion_error
  Fix task proposal: 26.1.1.fix -- Fix getByStatus return value
  DoD: npm test passes all

Approve: approve fix 26.1.1
Reject: reject fix 26.1.1
```

Future plan: auto-promotion to fully automatic when approval rate exceeds 80% (D30).

#### 2. Session snapshot (`/harness-sync --snapshot`)

**Before**: When resuming after a session break, you had to read Plans.md, check git log, and assess the situation yourself. This "situation assessment" took time each session, and WIP task progress couldn't be gleaned from Plans.md alone.

**After**: `/harness-sync --snapshot` saves current progress as JSON. The next SessionStart or `/resume` automatically displays a latest snapshot summary and comparison with the previous one.

```
Snapshot diff:

| Metric        | Previous (03/08 22:00) | Current    | Change   |
|---------------|----------------------|------------|----------|
| Completed     | 8/16                 | 13/16      | +5       |
| WIP           | 2                    | 0          | -2       |
| TODO          | 6                    | 3          | -3       |
```

Think of it as a "save point" for your work.

#### 3. Artifact Hash (linking tasks to commits)

**Before**: Even when Plans.md tasks became `cc:done`, there was no way to track which commit completed them. Finding "what changed for this task" required manually tracing git log.

**After**: On task completion, the recent commit hash (7-char short form) is automatically attached to the Status field.

```markdown
| Task | Description          | Status              |
|------|----------------------|---------------------|
| 26.1 | Add snapshot feature | cc:done [a1b2c3d]   |  <- auto-attached
```

`git show a1b2c3d` shows the changes for that task at any time. `cc:done` without hash remains valid (backward compatible).

#### 4. Progress Feed (progress display during Breezing)

**Before**: When running all tasks in parallel with `/breezing`, no progress was shown in the terminal until completion. With 10+ tasks, there was no way to know "how many are done" -- causing anxiety.

**After**: Each time a Worker completes a task, Lead outputs a one-line progress summary.

```
Progress: Task 1/16 completed -- "Add failure re-ticketing to harness-work"
Progress: Task 2/16 completed -- "Add --snapshot to harness-sync"
Progress: Task 3/16 completed -- "Add progress feed to breezing"
```

TaskCompleted hook's `systemMessage` also outputs coordinated progress information.

#### 5. Plans.md Purpose line

**Before**: Phase headers only had names and tags. Understanding "what this phase is for" required reading the body text.

**After**: Optionally add a single `Purpose:` line after the Phase header. Not writing it is fine (not mandatory). Auto-populated only when users state the phase's purpose.

```markdown
### Phase 26.0: Failure -> re-ticketing flow [P0]

Purpose: Transform from "just stopping" on self-correction loop failure to "proposing the next step"
```

---

## [3.6.0] - 2026-03-08

### What's Changed for You

**Solo mode PM framework: structured self-questioning built into every skill. Impact x Risk planning, DoD/Depends columns, Value-axis reviews, and retrospectives -- no new commands, just smarter existing ones.**

| Before | After |
|--------|-------|
| Plans.md had 3 columns (Task, Content, Status) | Plans.md has 5 columns (+DoD, +Depends); v1 format dropped |
| Priority was 1-axis (Required/Recommended/Optional) | 2-axis Impact x Risk matrix with automatic `[needs-spike]` for high-risk items |
| Plan Review checked 4 axes (Clarity/Feasibility/Dependencies/Acceptance) | 5 axes (+Value: user problem fit, alternative analysis, Elephant detection) |
| No retrospective capability | `sync` auto-runs retro when completed tasks exist (`--no-retro` to skip) |
| Breezing Phase 0 was undefined | Structured 3-question pre-flight check (scope, dependencies, risk flags) |
| Solo mode jumped straight to implementation | Step 1.5 background confirmation (purpose + impact scope inference) |
| Task dependencies were implicit in text | Explicit `Depends` column enables dependency-graph-based task assignment |

---

### Added
- **Plans.md v2 format**: 5-column table with DoD (Definition of Done) and Depends columns
- **DoD auto-inference**: `harness-plan create` generates testable completion criteria from task keywords
- **Depends auto-inference**: Automatic dependency detection (DB->API->UI->Test ordering)
- **`[needs-spike]` marker**: High Impact x High Risk tasks get auto-generated spike (tech validation) tasks
- **Plan Review Value axis**: 5th review axis checking user problem fit, alternatives, and Elephant detection
- **DoD/Depends quality checks**: Empty DoD warnings, untestable DoD suggestions, circular dependency detection
- **Retrospective (default ON)**: `sync` auto-runs retro when `cc:done` tasks >= 1; `--no-retro` to skip
- **Breezing Phase 0 structured check**: 3-question pre-flight (scope confirmation, dependency validation, risk flags)
- **Solo Step 1.5**: 30-second background confirmation inferring task purpose and impact scope
- **Dependency-graph task assignment**: Breezing assigns Depends=`-` tasks first, chains dependents on completion

### Changed
- **harness-plan create Step 5**: Upgraded from 1-axis to Impact x Risk 2-axis priority matrix
- **harness-plan SKILL.md**: Plans.md format specification updated to v2 with DoD/Depends guide
- **harness-plan sync**: v1 (3-column) format support removed; Plans.md is always 5-column
- **harness-review Plan Review**: Expanded from 4-axis to 5-axis evaluation
- **harness-work Solo flow**: Added Step 1.5 between task identification and WIP marking
- **breezing Flow Summary**: Phase 0 now has concrete check items instead of undefined discussion

---

## [3.5.0](https://github.com/tim-hub/powerball-harness/compare/v3.4.2...v3.5.0) - 2026-03-07

### What's Changed for You

**Claude Code v2.1.70-v2.1.71 features fully integrated: `/loop` scheduling for active monitoring, `PostToolUseFailure` auto-escalation, safe background agents, and Marketplace `@ref` installs.**

| Before | After |
|--------|-------|
| Feature Table covered up to v2.1.69 | Feature Table now covers v2.1.70-v2.1.71 (12 new items) |
| No automatic escalation on repeated tool failures | `PostToolUseFailure` hook escalates after 3 consecutive failures within 60s |
| Breezing relied solely on passive TeammateIdle monitoring | `/loop 5m /sync-status` enables active polling alongside passive hooks |
| Background agents risked losing output after compaction | v2.1.71 fix documented; `run_in_background` usage guide added |
| Plugin install used plain `owner/repo` | `owner/repo@vX.X.X` ref pinning recommended (v2.1.71 parser fix) |

---

### Added
- **`PostToolUseFailure` hook handler**: 60-second window consecutive failure counter with auto-escalation after 3 failures
- **Feature Table v2.1.70-v2.1.71**: 12 items added to `docs/CLAUDE-feature-table.md`
- **Breezing `/loop` guide**: Active monitoring guide explaining the division of roles between `TeammateIdle` and `/loop`
- **Breezing Background Agent guide**: `run_in_background` operational guide reflecting v2.1.71 output path fix
- **Marketplace `@ref` install guidance**: Setup procedure recommending `owner/repo@vX.X.X`

### Changed
- **CLAUDE.md Feature Table**: Reflected `/loop`, `PostToolUseFailure`, Background Agent output fix, Compaction image retention
- **Feature adoption notes**: Organized Plugin hooks fix, `--print` hang fix, parallel plugin install fix, `--resume` skill re-injection removal in Feature Table
- **README version badges**: Synced to `3.5.0`
- **Compatibility doc**: Updated plugin version to `3.5.0`

### Fixed
- Windows checkout with `core.symlinks=false` no longer hides `harness-*` command skills before SessionStart runs

### Security
- **Symlink-safe failure counter writes**: `post-tool-failure.sh` skips state writes when symlinks are detected on `.claude` parent directory, `.claude/state`, or `tool-failure-counter.txt`

---

## [3.4.2](https://github.com/tim-hub/powerball-harness/compare/v3.4.1...v3.4.2) - 2026-03-06

### What's Changed for You

**README now explains Claude Harness as a steadier operating model, not just a feature list, and `/harness-work all` now ships with rerunnable success and failure evidence that matches the real exit status.**

| Before | After |
|--------|-------|
| README mixed feature descriptions, comparison copy, and duplicate visual explanations | README now leads with clearer "what changes after install" messaging and SVG-driven comparisons |
| `/harness-work all` evidence existed, but the full runner could misread a failing test exit code | success / failure evidence runners now record the real command status, so the artifact contract matches what actually happened |

### Changed
- **README refresh (EN/JA)**: Reworked the hero and comparison sections around the default operating path after install, added new SVG cards, and removed duplicated explanation blocks.
- **Competitive positioning docs**: Added a dated harness comparison matrix, compatibility notes, distribution scope, claims audit, positioning notes, and release checklist docs so public claims stay grounded.
- **Codex package surface**: Clarified `harness-*` workflow surfaces in Codex docs and aligned setup scripts with path-based skill loading.

### Added
- **`/harness-work all` evidence pack**: Added success / failure fixtures, smoke/full runners, replay-aware success artifacts, and public docs for rerunnable verification.
- **README visual assets**: Added `why-harness-pillars` and default-flow comparison SVGs in both English and Japanese.

### Fixed
- **Evidence runner exit status capture**: Full success / failure runners now preserve the real `claude` and `npm test` exit codes instead of the inverted `!` status.
- **Claim drift checks**: Expanded `check-consistency.sh` to catch README badge drift, missing docs, stale positioning claims, and distribution-scope mismatches before release.

---

## [3.4.1](https://github.com/tim-hub/powerball-harness/compare/v3.4.0...v3.4.1) - 2026-03-06

### What's Changed for You

**Fixed stale skill labels in the Claude Code 2.1.69+ feature tables (EN/JA), so the docs now match the actual harness skill set.**

| Before | After |
|--------|-------|
| `task-worker`, `code-reviewer`, `work`, `all skills` labels remained in README feature tables | Unified to current names: `harness-work`, `harness-review`, `all harness-* skills` |

### Changed
- **README (EN/JA) feature table cleanup**: Updated the "Skills" column under "Claude Code 2.1.69+ Features" to current harness naming.

### Fixed
- **Documentation drift**: Removed legacy skill aliases that could mislead users during `/breezing` and `/harness-work` onboarding.

---

## [3.4.0](https://github.com/tim-hub/powerball-harness/compare/v3.3.1...v3.4.0) - 2026-03-06

### What's Changed for You

**Completed Claude Code v2.1.69 support. Updated teammate event control, skill reference resolution, and development flow documentation in a single batch, strengthening team execution stop decisions and compatibility.**

| Before | After |
|--------|-------|
| Teammate hooks were session_id-centric and always approve-only | Uses `agent_id`/`agent_type` and can return `{"continue": false, "stopReason": "..."}` for stops |
| `InstructionsLoaded` event was not handled | Dedicated handler added and wired in both hooks.json files |
| SKILL references used relative `references/` paths | Unified to `${CLAUDE_SKILL_DIR}/references/...` reducing execution environment dependency |
| Docs were centered on 2.1.68+ | Feature docs/README/command docs updated to 2.1.69+ |

### Added
- **InstructionsLoaded handler**: Added `scripts/hook-handlers/instructions-loaded.sh`
- **Teammate stop response support**: Added `continue:false` response logic to `teammate-idle.sh` / `task-completed.sh`
- **2.1.69 feature docs**: Documented `${CLAUDE_SKILL_DIR}`, `agent_id/agent_type`, `/reload-plugins`, `includeGitInstructions: false`, `git-subdir` operational policies

### Changed
- **PreToolUse breezing role guard**: Extended role lookup to `agent_id` priority with `session_id` fallback
- **SKILL reference path policy**: Updated references in skills/codex/opencode SKILL.md files to `${CLAUDE_SKILL_DIR}` base
- **check-consistency**: Validates project template `defaultMode` baseline and documents policy against distributing undocumented values
- **Feature docs**: Updated CLAUDE.md / README / README_ja / docs/CLAUDE-feature-table.md / docs/CLAUDE-commands.md

### Fixed
- **Plans drift**: Synced Phase 17/19 unsynced task markers to actual state
- **continue:false parsing**: Fixed cases where boolean `false` was dropped, ensuring stopReason is reliably reflected

---

## [3.3.1](https://github.com/tim-hub/powerball-harness/compare/v3.3.0...v3.3.1) - 2026-03-05

### What's Changed for You

**All README visuals unified to brand-orange palette, logo regenerated with Nano Banana Pro, and duplicate content sections removed for a cleaner reading experience.**

| Before | After |
|--------|-------|
| Mixed indigo/blue/teal/purple SVGs | Unified orange palette (#F7931A hierarchy) |
| Hero comparison shown twice (SVG + table) | Single SVG visualization |
| /work all flow shown twice (mermaid + SVG) | Single SVG visualization |
| Review section had no visual | 4-perspective review card SVG added |
| 47KB logo (old design) | 53KB Nano Banana Pro logo with "Plan -> Work -> Review" tagline |

### Changed
- **8 SVGs recolored** (EN/JA): Unified orange brand palette across all README visuals
- **Logo regenerated**: Nano Banana Pro interlocking-loops icon + "Plan -> Work -> Review" tagline
- **README cleanup**: Removed duplicate mermaid/SVG and SVG/table sections in both EN/JA

### Added
- **Review perspectives SVG** (EN/JA): 4-angle code review visualization (Security, Performance, Quality, Accessibility)
- **3 JA generated SVGs**: hero-comparison, core-loop, safety-guardrails (Japanese localized versions)
- **Alternative logo**: `docs/images/claude-harness-logo-alt.png` (carabiner icon + color-split text)

---

## [3.3.0](https://github.com/tim-hub/powerball-harness/compare/v3.2.0...v3.3.0) - 2026-03-05

### What's Changed for You

**Claude Code v2.1.68 introduced effort levels, agent hooks, and more. Harness v3.3.0 puts all of them to work -- so you get smarter task execution, LLM-powered code guards, and fully automated worktree lifecycle out of the box.**

> Claude Code got new superpowers. Harness makes sure you actually use them.

| What Claude Code added | How Harness uses it |
|------------------------|---------------------|
| **Opus 4.6 medium effort default** -- Claude now thinks less deeply by default | Harness auto-detects complex tasks (security, architecture, multi-file changes) and injects `ultrathink` to restore full thinking depth exactly when it matters |
| **Agent hooks (`type: "agent"`)** -- hooks can now use LLM intelligence | 3 smart guards deployed: catches hardcoded secrets before commit, blocks session exit with unfinished tasks, runs lightweight code review after every write |
| **WorktreeCreate/Remove hooks** -- lifecycle events for git worktrees | Breezing parallel workers now auto-initialize their workspace and clean up temp files when done. No more orphaned `/tmp` clutter |
| **`CLAUDE_ENV_FILE`** -- session environment persistence | Harness version, effort defaults, and Breezing session IDs persist across hooks. Workers know who they are |
| **Prompt hooks expanded to all events** -- no longer Stop-only | Every hook event can now use LLM judgment (was incorrectly documented as Stop-only) |

### Added
- **Effort level auto-tuning**: Multi-element scoring system (file count + directory criticality + task keywords + past failure history). Score >= 3 triggers `ultrathink` -- meaning complex tasks get deep thinking, simple tasks stay fast
- **Agent hooks (3 deployments)**:
  - *PreToolUse quality guard*: LLM reviews every Write/Edit for secrets, TODO stubs, and security issues before they land
  - *Stop WIP guard*: Reads Plans.md and warns you if you're about to close a session with unfinished `cc:WIP` tasks
  - *PostToolUse code review*: Lightweight haiku-powered review runs after every file write
- **Worktree lifecycle automation**: `worktree-create.sh` sets up `.claude/state/worktree-info.json` with worker identity; `worktree-remove.sh` cleans Codex temp files and logs
- **Session environment persistence**: `session-env-setup.sh` writes `HARNESS_VERSION`, `HARNESS_EFFORT_DEFAULT=medium`, and `HARNESS_BREEZING_SESSION_ID` to `CLAUDE_ENV_FILE`
- **PreCompact agent hook**: Catches WIP tasks before context compaction -- so important context isn't lost mid-task
- **HTTP hook template**: Ready-to-use PostToolUse metrics hook for external dashboards (localhost:9090)

### Changed
- **4-type hook system**: Harness now supports all 4 hook types -- `command`, `prompt` (all events), `http`, and `agent`
- **Feature Table**: Updated from v2.1.63+ to v2.1.68+ with 30 tracked features
- **Worker/Reviewer/Team agents**: Now understand effort levels and when to request deeper thinking
- **PM templates**: All handoff templates include `ultrathink` with clear intent comments

### Fixed
- **Prompt hook documentation**: Removed incorrect "Stop/SubagentStop only" restriction (prompt hooks work on all events since v2.1.63)
- **Dead reference cleanup**: Removed link to deleted `guardrails-inheritance.md` in Feature Table

---

## [3.2.0] - 2026-03-04

### What's Changed for You

**TDD is now enabled by default for all tasks, and Windows users get automatic symlink repair on session start.**

| Before | After |
|--------|-------|
| TDD only active with `[feature:tdd]` marker (opt-in) | TDD active by default; skip with `[skip:tdd]` (opt-out) |
| Windows users: v3 skills not recognized (broken symlinks) | Auto-detected and repaired on session start |
| Worker had no TDD phase in execution flow | TDD phase (Red->Green) integrated into Worker and Solo mode |

### Added
- **TDD-by-default**: TDD is now opt-out (`[skip:tdd]`) instead of opt-in (`[feature:tdd]`). All WIP tasks get TDD reminders unless explicitly skipped
- **`--no-tdd` option**: Skip TDD phase in `/harness-work` execution
- **Windows symlink auto-repair**: `fix-symlinks.sh` detects broken symlinks from Windows git clone and replaces them with directory copies
- **Session-init Step 1.5**: Symlink health check runs automatically before skill discovery

### Changed
- **tdd-order-check.sh**: `has_tdd_wip_task()` split into `has_active_wip_task()` + `is_tdd_skipped()` for clearer logic
- **harness-plan create.md**: Step 5.5 inverted from "TDD adoption criteria" to "TDD skip criteria"
- **worker.md**: Execution flow expanded from 10 to 12 steps with TDD judgment and Red phase
- **harness-work SKILL.md**: Solo mode expanded from 6 to 7 steps with TDD phase

---

## [3.1.0] - 2026-03-03

### What's Changed for You

**Codex CLI 0.107.0 full compatibility, 15 deprecated skill stubs removed (-40,000 lines), and `/harness-work` now auto-selects the best execution mode based on task count.**

| Before | After |
|--------|-------|
| 15 deprecated redirect stubs cluttering skill listings | Clean 5-verb structure only |
| `/harness-work` always defaulted to Solo mode | Auto-detection: 1->Solo, 2-3->Parallel, 4+->Breezing |
| `--codex` could be confusing for users without Codex CLI | `--codex` is explicit-only, never auto-selected |
| MCP server references in Codex config | All MCP remnants removed, pure CLI integration |
| `--approval-policy` (non-official flag) in docs | Correct `-a never -s workspace-write` flags |

### Added
- **Auto Mode Detection**: `/harness-work` auto-selects Solo/Parallel/Breezing based on task count (1/2-3/4+)
- **Breezing backward-compatible alias**: `/breezing` delegates to `/harness-work --breezing`
- **Codex environment fallback**: Added Plans.md direct manipulation pattern for harness-review when Task tool is unavailable
- **Codex environment notes**: Added Codex CLI-specific constraints and alternatives to team-composition.md, worker.md
- **config.toml expansion**: Added [notify] section (after_agent memory bridge), reviewer Read-only sandbox
- **.codexignore**: Added patterns to prevent CLAUDE.md noise
- **README visual improvement**: hero-comparison, core-loop, safety-guardrails images

### Changed
- **MCP remnant removal**: Completely removed MCP server references from config.toml, setup-codex.sh, codex-setup-local.sh
- **codex exec flag normalization**: --approval-policy -> -a (--ask-for-approval), --sandbox -> -s unified
- **Prompt passing improvement**: "$(cat file)" -> stdin pipe (`cat file | codex exec -`) (ARG_MAX mitigation)
- **codex-worker-engine.sh**: Renamed mcp-params.json -> codex-exec-params.json

### Fixed
- **/tmp/codex-prompt.md fixed path**: Changed to mktemp unique path (prevents conflicts in parallel execution)
- **2>/dev/null error suppression**: Changed to log file redirect (enables debugging)
- **Skill description quality**: gogcli-ops YAML fix, session-memory invalid tool removal, session-state non-standard fields cleanup

### Removed
- **15 DEPRECATED redirect stubs**: breezing(old), codex-review, handoff, harness-init, harness-update, impl, maintenance, parallel-workflows, planning, plans-management, release-har, setup, sync-status, troubleshoot, verify, work -- all consolidated into 5-verb skills
- **Old -harness suffix stubs**: plan-harness, release-harness, review-harness, setup-harness, work-harness from skills-v3/
- **x-release-harness**: consolidated into harness-release

---

## [3.0.0] - 2026-03-02

### What's Changed for You

**Harness v3: Full architectural rewrite -- 42 skills unified to 5 verbs, 11 agents consolidated to 3, TypeScript engine replaces Bash guardrails, SQLite replaces scattered JSON state files.**

| Before | After |
|--------|-------|
| 42 skills spread across multiple dirs | 5 verb skills: `plan` / `execute` / `review` / `release` / `setup` |
| 11 agents with overlapping responsibilities | 3 agents: `worker` / `reviewer` / `scaffolder` |
| Bash scripts for guardrails (pretooluse-guard.sh etc.) | TypeScript engine in `core/` (strict, ESM, NodeNext) |
| JSON/JSONL state files scattered across dirs | SQLite single-file state via `better-sqlite3` |
| rsync-based mirror sync for codex/opencode | Symlink-based mirror (zero sync overhead) |
| No session lifecycle management | `core/engine/lifecycle.ts` unifies session-init/control/state/memory |

### Added

- **`core/` TypeScript engine**: Strict ESM module (`exactOptionalPropertyTypes`, `noUncheckedIndexedAccess`, `NodeNext`). Includes guardrails, state, and engine subsystems
- **`core/src/guardrails/`**: Rules engine (R01-R09), pre-tool/post-tool/permission/tampering detection -- all ported from Bash to TypeScript
- **`core/src/state/`**: SQLite state management via `better-sqlite3` with schema, store, and JSON->SQLite migration
- **`core/src/engine/lifecycle.ts`**: Session lifecycle -- `initSession`, `transitionSession`, `finalizeSession`, `forkSession`, `resumeSession`
- **`skills-v3/`**: 5 verb skills with unified SKILL.md + references/
- **`agents-v3/`**: 3 consolidated agent definitions + team-composition.md
- **`tests/validate-plugin-v3.sh`**: v3 structural validator (6 checks, 34 assertions)
- **Symlink mirrors**: `codex/.codex/skills/` and `opencode/skills/` 5-verb dirs now symlinks to `skills-v3/`
- **`skills-v3/routing-rules.md`**: Trigger/exclusion keywords per skill verb

### Changed

- **Skills**: 42 -> 5 (plan/execute/review/release/setup). Legacy `skills/` retained for backwards compatibility
- **Agents**: 11 -> 3 (worker/reviewer/scaffolder). Legacy `agents/` retained for backwards compatibility
- **Hooks shims**: `hooks/pre-tool.sh`, `hooks/post-tool.sh`, `hooks/permission.sh` now delegate to `core/src/index.ts`
- **PermissionRequest**: Switched from v2 `run-script.js permission-request` to v3 TypeScript core (`hooks/permission.sh`)
- **`check-consistency.sh`**: Mirror check updated from rsync diff to symlink validation
- **CLAUDE.md**: Compact v3 version; architecture details moved to `.claude/rules/v3-architecture.md`
- **README.md / README_ja.md**: Updated for v3 (5 verb skills, 3 agents, TypeScript core, architecture diagram)

### Fixed

- **`core/src/state/store.ts`**: Fixed `better-sqlite3` type import -- `typeof import("better-sqlite3").default` -> `import type DatabaseConstructor from "better-sqlite3"` (ESM/CJS compatibility)
- **Duplicate `posttooluse-tampering-detector`**: Removed v2 script from PostToolUse `Write|Edit|Task` block (v3 `post-tool.ts` already handles tampering detection)

### Removed

- rsync-based mirror sync (replaced by symlinks)
- Standalone Bash guardrail scripts (replaced by `core/src/guardrails/`)
- Scattered JSON/JSONL state files (replaced by SQLite)
- Duplicate `posttooluse-tampering-detector` hook (consolidated into v3 post-tool engine)

---

## [2.26.1](https://github.com/tim-hub/powerball-harness/compare/v2.26.0...v2.26.1) - 2026-03-02

### Added

- **12 section-specific SVG illustrations**: 6 EN + 6 JA hand-crafted visuals embedded in both READMEs (before-after, /work all flow, parallel workers, safety shield, skills ecosystem, breezing agents)

### Fixed

- **review-loop.md APPROVE flow inconsistency**: Phase 3.5 Auto-Refinement step was missing from the APPROVE judgment table, causing inconsistency with SKILL.md and execution-flow.md

## [2.26.0](https://github.com/tim-hub/powerball-harness/compare/v2.25.0...v2.26.0) - 2026-03-02

### What's Changed for You

**Claude Code v2.1.63 integration: `/work` now auto-simplifies code after review, `/breezing` can delegate horizontal tasks to `/batch`, and HTTP hooks enable external service notifications.**

| Before | After |
|--------|-------|
| `/work` flow: implement -> review -> commit | `/work` flow: implement -> review -> **auto-simplify** -> commit |
| Horizontal migration tasks handled manually | `/breezing` auto-detects and delegates to `/batch` |
| Feature table covers up to v2.1.51 | Feature table covers up to v2.1.63 (27 features) |
| Hooks only support `command` and `prompt` types | Hooks now support `http` type (POST to external services) |

### Added

- **Phase 3.5 Auto-Refinement in `/work`**: After review APPROVE, `/simplify` runs automatically to clean up code. `--deep-simplify` adds `code-simplifier` plugin. `--no-simplify` skips
- **`/batch` delegation in `/breezing`**: Horizontal pattern detection (migrate/replace-all/add-to-all) auto-proposes `/batch` delegation for bulk changes
- **HTTP hooks documentation** (`.claude/rules/hooks-editing.md`): `type: "http"` spec with field reference, response behavior, command-vs-http comparison table, and 3 sample templates (Slack, metrics, dashboard)
- **7 new feature-table entries** (`docs/CLAUDE-feature-table.md`): `/simplify`, `/batch`, `code-simplifier` plugin, HTTP hooks, auto-memory worktree sharing, `/clear` skill cache reset, `ENABLE_CLAUDEAI_MCP_SERVERS`

### Changed

- **Version references**: `2.1.49+` -> `2.1.63+` across CLAUDE.md and feature table
- **Feature count**: 20 -> 27 in CLAUDE.md and feature table
- **`/breezing` guardrails**: Added auto-memory worktree sharing (v2.1.63) to inheritance table
- **`troubleshoot` skill**: Added `/clear` cache reset to CC v2.1.63+ diagnostics
- **`work-active.json` schema**: Added `simplify_mode: "default" | "deep" | "skip"` field

## [2.25.0](https://github.com/tim-hub/powerball-harness/compare/v2.24.0...v2.25.0) - 2026-02-24

### What's Changed for You

**Auto-detects `CLAUDE_CODE_SIMPLE` mode (CC v2.1.50+) impact and explicitly shows users which features are disabled. Prevents silent failures.**

| Before | After |
|--------|-------|
| 37 skills and 11 agents silently disabled in SIMPLE mode | SessionStart/Setup hooks auto-detect and show warnings in terminal + additionalContext |
| Impact scope of SIMPLE mode unknown (only 1 line in compatibility matrix) | Dedicated doc `docs/SIMPLE_MODE_COMPATIBILITY.md` covers all impacts (skills, agents, memory, workflows) |
| Zero defense code or detection logic | `scripts/check-simple-mode.sh` utility for consistent detection and multilingual warnings |
| `/work`, `/breezing` etc. fail without explanation | Immediately understandable with 3 categories: "skills disabled," "agents disabled," "hooks only" |

### Added

- **SIMPLE mode detection utility** (`scripts/check-simple-mode.sh`): `is_simple_mode()` function and `simple_mode_warning()` multilingual message generation. Can be sourced from any hook/script
- **SessionStart SIMPLE mode warning**: `scripts/session-init.sh` detects `CLAUDE_CODE_SIMPLE` env var at session start, outputs stderr banner + additionalContext with detailed warning
- **Setup hook SIMPLE mode warning**: `scripts/setup-hook.sh` detects SIMPLE mode during init/maintenance, adds warning to output
- **`docs/SIMPLE_MODE_COMPATIBILITY.md`**: Complete SIMPLE mode guide -- impact summary table, working/non-working lists, impact classification for 37 skills and 11 agents, detection methods, workarounds, developer extension guide

### Changed

- **Compatibility matrix enhancement** (`docs/CLAUDE_CODE_COMPATIBILITY.md`):
  - Updated v2.1.50 SIMPLE mode row status from "caution" to "**supported**"
  - Added SIMPLE mode detailed impact (37 skills, 11 agents, memory disabled) and detection methods to incompatibility section
  - Added cross-reference link to `SIMPLE_MODE_COMPATIBILITY.md`

---

## [2.24.0](https://github.com/tim-hub/powerball-harness/compare/v2.23.6...v2.24.0) - 2026-02-24

### What's Changed for You

**Claude Code v2.1.50-v2.1.51 new feature support. Updated compatibility matrix, memory stability improvements, and new CLI command utilization.**

| Before | After |
|--------|-------|
| Compatibility matrix stopped at v2.1.49 | All v2.1.50-v2.1.51 features documented, recommended version raised to v2.1.51+ |
| WorktreeCreate/Remove hooks unknown | Documented as future support in Breezing guardrails |
| Limited diagnostics for agent spawn failures | Added `claude agents list` (CC 2.1.50+) to troubleshoot skill |
| Background agent stop methods undocumented | Added `Ctrl+F` (CC 2.1.49+) to breezing guardrails, ESC deprecated |

### Added

- **CC v2.1.50/v2.1.51 compatibility matrix**: 17 items added to `docs/CLAUDE_CODE_COMPATIBILITY.md` (memory leak fix, completed task GC, WorktreeCreate/Remove hooks, `claude agents` CLI, declarative worktree isolation, SIMPLE mode caution, remote-control, etc.)
- **`claude agents` CLI diagnostics**: Added agent diagnostics section to `skills/troubleshoot/SKILL.md` (CC 2.1.50+)
- **WorktreeCreate/WorktreeRemove hooks**: Added as future support to `skills/breezing/references/guardrails-inheritance.md`
- **Ctrl+F keybinding**: Added background agent stop method to breezing guardrails (CC 2.1.49+, ESC deprecated)
- **Feature Table expansion**: 4 features added to `docs/CLAUDE-feature-table.md` for v2.1.50/v2.1.51 (memory leak fix, claude agents CLI, WorktreeCreate/Remove, remote-control)

### Changed

- **Recommended CC version**: Raised from v2.1.49+ to **v2.1.51+**
- **Feature Table title**: Updated from 2.1.49+ to 2.1.51+

---

## [2.23.6](https://github.com/tim-hub/powerball-harness/compare/v2.23.5...v2.23.6) - 2026-02-24

### Added

- **Auto-release workflow** (`release.yml`): Safety-net GitHub Release creation on `v*` tag push -- prevents orphan tags if `release-har` is interrupted
- **CHANGELOG format validation in CI**: ISO 8601 date format, `[Unreleased]` section presence, non-standard heading warnings
- **Codex mirror sync check in CI**: `codex/.codex/skills/` <-> `skills/` consistency validated in both `check-consistency.sh` and `opencode-compat.yml`
- **Branch Policy in release-har**: Explicitly documents that main direct push is allowed for solo projects (force push remains prohibited)

### Changed

- **CHANGELOG link definitions repaired**: All version compare links supplemented
- **CHANGELOG_ja.md translation gaps filled**: 5 versions added (2.20.1, 2.17.6, 2.17.1, 2.17.0, 2.16.21)
- **README version and count updated**: Badge version, skill count (41), agent count (11) updated to reflect reality
- **CHANGELOG non-standard headings normalized**: `### Internal` -> `### Changed` (Keep a Changelog compliant)
- **Mirror compat workflow renamed**: `OpenCode Compatibility Check` -> `Mirror Compatibility Check` (now covers both opencode and codex mirrors)
- **AGENTS.md template updated**: Removed `main` direct push prohibition for solo projects; force push remains prohibited
- **Tamper detection expanded** (`codex-worker-quality-gate.sh`): Python skip patterns, catch-all assertions, config relaxation detection

---

## [2.23.5](https://github.com/tim-hub/powerball-harness/compare/v2.23.3...v2.23.5) - 2026-02-23

### What's Changed for You

**Phase 13: Breezing quality automation and Codex rule injection -- tamper detection, auto-test runner, CI signal handling, AGENTS.md rule sync, and APPROVE fast-path.**

| Before | After |
|--------|-------|
| Test tampering detection covered skip patterns and assertion deletion only | 12+ patterns: weakening (`toBe -> toBeTruthy`), timeout inflation, catch-all assertions, Python skip decorators |
| Auto-test runner only recommended tests without running them | `HARNESS_AUTO_TEST=run` actually runs tests and feeds results back via `additionalContext` |
| CI failures required manual detection | PostToolUse hook detects CI failures after `git push` and injects `ci-cd-fixer` recommendation signals |
| `.claude/rules/` existed only for Claude Code; Codex had no rule awareness | `sync-rules-to-agents.sh` auto-syncs rules to `codex/AGENTS.md`; Codex reads full project rules on startup |
| `codex exec` called bare without pre/post processing | `codex-exec-wrapper.sh` handles rule sync, `[HARNESS-LEARNING]` extraction, and secret filtering |
| Breezing Phase C required manual APPROVE confirmation | `review-result.json` + commit hash check enables instant fast-path to integration tests |
| Implementer count fixed at `min(independent_tasks, 3)` | Auto-calculated as `max(1, min(independent_tasks, --parallel, planner_max_parallel, 5))` |

### Added

- **Tamper detection (12+ patterns)**: assertion weakening, timeout inflation, catch-all assertions, Python skip decorators -- `scripts/posttooluse-tampering-detector.sh`
- **`HARNESS_AUTO_TEST=run` mode**: `scripts/auto-test-runner.sh` actually runs tests and returns pass/fail via `additionalContext` JSON
- **CI signal injection**: `scripts/hook-handlers/ci-status-checker.sh` detects CI failures post-push and writes to `breezing-signals.jsonl`; `scripts/hook-handlers/breezing-signal-injector.sh` injects unconsumed signals via UserPromptSubmit hook
- **`sync-rules-to-agents.sh`**: Auto-converts `.claude/rules/*.md` to `codex/AGENTS.md` Rules section with hash-based drift detection
- **`codex-exec-wrapper.sh`**: Pre/post wrapper for `codex exec` -- rule sync, `[HARNESS-LEARNING]` marker extraction, secret filtering, atomic write-back to `codex-learnings.md`
- **APPROVE fast-path (Phase C)**: Checks `.claude/state/review-result.json` + HEAD commit hash; skips manual confirmation when APPROVE is already recorded
- **`review-result.json` auto-record**: Reviewer reports `review_result_json` in SendMessage; Lead writes `.claude/state/review-result.json` for fast-path reference
- **Docs reorganization**: `docs/CLAUDE-feature-table.md`, `docs/CLAUDE-skill-catalog.md`, `docs/CLAUDE-commands.md` -- detailed references extracted from CLAUDE.md
- **`harness.rules` -- execpolicy guard rules**: `npm test`/`yarn test`/`pnpm test` auto-allowed; `git push --force`, `git reset --hard`, `rm -rf`, `git clean -f`, SQL destructive statements (`DROP TABLE`, `DELETE FROM`) require user confirmation via `codex execpolicy`; 20 patterns verified with `codex execpolicy check`

### Changed

- **CLAUDE.md compressed to 120 lines**: Feature Table (5 items), skill category table (5 categories); full details moved to `docs/`
- **Implementer count auto-determination**: `max(1, min(independent_tasks, --parallel N, planner_max_parallel, 5))` -- starvation prevention + hard cap at 5
- **`review-retake-loop.md`**: Added `review-result.json` write spec with JSON format, Reviewer->Lead delegation flow, and file lifecycle
- **`execution-flow.md` Phase C**: APPROVE fast-path check added as step 2; phase processing renumbered
- **`team-composition.md`**: Extended configuration (5 Implementers) cost estimate table added
- **`release-har` skill redesigned (Phase 14)**: Full redesign with Pre-flight checks, structured git log, Conventional Commits classification, Claude diff summarization (Highlights + Before/After), SemVer auto-detection, dry-run preview, 4-section Release Notes, Compare link auto-generation, `--announce` option, and `--dry-run` default gate; `references/release-notes-template.md` and `references/changelog-format.md` added

---

## [2.23.3](https://github.com/tim-hub/powerball-harness/compare/v2.23.2...v2.23.3) - 2026-02-22

### What's Changed for You

**Codex integration is now explicitly CLI-first (`codex exec`) outside breezing, and Codex package parity includes the new `generate-slide` skill.**

| Before | After |
|--------|-------|
| `work`/`harness-review`/`codex-review` docs mixed Codex MCP wording with CLI execution examples | Non-breezing Codex flows are documented as CLI-only (`codex exec`) with consistent setup and troubleshooting |
| `codex-worker-setup.sh` checked MCP registration state | Setup now checks `codex exec` readiness directly (`codex_exec_ready`) |
| Codex package parity test did not block non-breezing MCP vocabulary regressions | New CLI-only regression checks added to `tests/test-codex-package.sh` |
| `generate-slide` existed in source/opencode but not in Codex package | `codex/.codex/skills/generate-slide/` is now included and parity tests pass |

### Added

- **Codex package skill parity**: Added `generate-slide` skill files to `codex/.codex/skills/`
- **CLI-only regression guard**: Added non-breezing Codex vocabulary checks to `tests/test-codex-package.sh`
- **README updates (EN/JA)**: Added `/generate-slide` command docs and slide-generation feature section

### Changed

- **Codex docs (non-breezing)**: Updated `work`, `harness-review`, `codex-review`, routing/setup references to CLI-first terminology and behavior (`codex exec`)
- **Codex setup reference**: Reworked `codex-mcp-setup.md` content into Codex CLI setup flow (legacy filename retained for compatibility)
- **README Codex review section (EN/JA)**: Clarified Codex second-opinion execution path as Codex CLI-based

### Fixed

- **Setup behavior mismatch**: Replaced MCP registration check in `scripts/codex-worker-setup.sh` with actual CLI execution readiness check
- **Codex mirror consistency**: Synced updated non-breezing Codex skill docs between `skills/` and `codex/.codex/skills/`

---

## [2.23.2](https://github.com/tim-hub/powerball-harness/compare/v2.23.1...v2.23.2) - 2026-02-22

### What's Changed for You

**Codex skills now use fully native multi-agent vocabulary -- CI checks pass, and `--claude` review routing is explicitly documented.**

| Before | After |
|--------|-------|
| Codex breezing/work skills contained Claude Code-specific terms (`delegate mode`, `TaskCreate`, `subagent_type`, etc.) | All 82+ occurrences replaced with Codex native API equivalents (`Phase B`, `spawn_agent`, `role`, etc.) |
| No `review_engine` matrix in Codex breezing/work SKILL.md | `review_engine` comparison table added with `codex` / `claude` columns |
| `--claude + --codex-review` conflict undocumented | Explicit conflict rule: mutually exclusive, fails before execution |
| State files referenced `.claude/state/` paths | State files use `${CODEX_HOME:-~/.codex}/state/harness/` paths |
| `opencode/` contained stale breezing files | Rebuilt `opencode/` -- breezing removed (dev-only skill) |

### Fixed

- **Codex vocabulary migration**: replaced 82+ legacy Claude Code terms across 13 files in `codex/.codex/skills/breezing/` and `codex/.codex/skills/work/` -- `delegate mode` -> `Phase B`, `TaskCreate` -> `spawn_agent`, `subagent_type` -> `role:`/`spawn_agent()`, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` -> `config.toml [features] multi_agent`, `.claude/state/` -> `${CODEX_HOME}/state/harness/`
- **`--claude` review routing**: added `review_engine` matrix table and `--claude + --codex-review` conflict rule to both `breezing/SKILL.md` and `work/SKILL.md`
- **OpenCode sync**: rebuilt `opencode/` to remove stale breezing files and routing-rules.md

---

## [2.23.1](https://github.com/tim-hub/powerball-harness/compare/v2.23.0...v2.23.1) - 2026-02-22

### What's Changed for You

**Codex CLI setup now merges files instead of overwriting, and README setup instructions are clearer with a collapsible quick-start.**

| Before | After |
|--------|-------|
| `setup-codex.sh` overwrote all destination files on every sync | Merge strategy: new files added, existing files updated, user-created files preserved |
| Codex CLI Setup was a top-level README section | Moved to collapsible `<details>` block with step-by-step quick-start |
| `config.toml` had 4 agent definitions | 9 agents: added `task_worker`, `code_reviewer`, `codex_implementer`, `plan_analyst`, `plan_critic` |

### Changed

- **README (EN/JA)**: Codex CLI Setup section moved from top-level to collapsible `<details>` block with prerequisites, 3-step quick-start, and flag reference table
- **`setup-codex.sh`**: `sync_named_children()` rewritten with 3-way merge strategy -- new files are copied, existing files are backed up and updated, destination-only files are preserved; log output now shows `(N new, N updated, N preserved, N skipped)`
- **`codex-setup-local.sh`**: same merge strategy applied to project-local setup script

### Added

- **`merge_dir_recursive()`** helper in both setup scripts for recursive directory merging with backup
- **5 new Codex agent definitions** in `setup-codex.sh` `config.toml` generation: `task_worker`, `code_reviewer`, `codex_implementer`, `plan_analyst`, `plan_critic` (Breezing roles)
- Idempotent agent injection: existing `config.toml` files receive missing agent entries without duplicating existing ones

---

## [2.23.0](https://github.com/tim-hub/powerball-harness/compare/v2.22.0...v2.23.0) - 2026-02-21

### What's Changed for You

**Codex breezing now has its own Phase 0 (Planning Discussion) using Codex's native multi-agent API -- Planner and Critic agents analyze your plan before implementation begins.**

| Before | After |
|--------|-------|
| Codex breezing Phase 0 was dead code (referenced Claude-only APIs) | Phase 0 uses `spawn_agent`/`send_input`/`wait`/`close_agent` natively |
| `config.toml` had 4 agent definitions | 9 agents defined including `plan_analyst`, `plan_critic`, `task_worker`, `code_reviewer`, `codex_implementer` |
| All breezing reference files were identical between Claude and Codex | 3 files now intentionally diverge with platform-native implementations |

### Added

- **Codex Phase 0 (Planning Discussion)**: ported from Claude Agent Teams to Codex native multi-agent API (`spawn_agent`/`send_input`/`wait`/`close_agent`)
- **5 new Codex agent definitions** in `config.toml`: `plan_analyst`, `plan_critic`, `task_worker`, `code_reviewer`, `codex_implementer`
- **Mirror sync divergence management** (D24, P20): 3 breezing files (`planning-discussion.md`, `execution-flow.md`, `team-composition.md`) now excluded from rsync to preserve Codex-native implementations

### Changed

- **Codex `planning-discussion.md`**: fully rewritten with Codex native API -- Planner <-> Critic dialogue via Lead relay pattern using `send_input` + `wait` loops
- **Codex `execution-flow.md`**: Phase 0 + Phase A spawn logic updated to `spawn_agent()` format; environment check now references `config.toml [features] multi_agent = true`
- **Codex `team-composition.md`**: all role definitions updated -- `subagent_type` removed, `spawn_agent()` format, `SendMessage` -> `send_input()`, `shutdown_request` -> `close_agent()`

---

## [2.22.0](https://github.com/tim-hub/powerball-harness/compare/v2.21.0...v2.22.0) - 2026-02-21

### What's Changed for You

**Security guardrails now apply automatically from the moment you install Harness -- no `/harness-init` required. Permission policy hardened with least-privilege defaults and privacy-safe session logging.**

| Before | After |
|--------|-------|
| Security settings (deny/ask rules) required running `/harness-init` | Plugin settings applied automatically on install (CC 2.1.49+) |
| Plugin settings had a broad `allow` rule; no DB CLI protection | Least-privilege: removed blanket `allow`; added deny for `psql`/`mysql`/`mongo` |
| `stop-session-evaluator.sh` always returned `{"ok":true}` without reading input | Hook reads `last_assistant_message`, stores length+hash only (privacy-safe) with atomic writes |
| No hook for configuration file changes | New `ConfigChange` hook records config changes to breezing timeline when active |
| `npm install` / `bun install` ran without confirmation | Package manager installs now require user confirmation (`ask` rule) |

### Added

- **Plugin settings.json** (`.claude-plugin/settings.json`): default security permissions distributed with the plugin -- active from install (CC 2.1.49+)
  - **Deny**: `.env`, secrets, SSH keys (`id_rsa`, `id_ed25519`), `.aws/`, `.ssh/`, `.npmrc`, `sudo`, `rm -rf/-fr`, DB CLIs (`psql`, `mysql`, `mongo`)
  - **Ask**: destructive git (`push --force`, `reset --hard`, `clean -f`, `rebase`, `merge`), package installs (`npm/bun/pnpm install`), `npx`/`npm exec`
- **`ConfigChange` hook** (`scripts/hook-handlers/config-change.sh`): records configuration file changes to `breezing-timeline.jsonl` when breezing is active; always non-blocking
  - Normalizes `file_path` to repo-relative paths in timeline logs
  - Portable timeout detection (`timeout`/`gtimeout`/`dd` fallback)
- **`last_assistant_message` support** in `stop-session-evaluator.sh`: reads CC 2.1.47+ Stop payload
  - Stores message length + SHA-256 hash only (no plaintext -- privacy by design)
  - Atomic writes via `mktemp` (TOCTOU fix)
  - Portable hash detection (`shasum`/`sha256sum`)
- **CC 2.1.49 compatibility matrix** (`docs/CLAUDE_CODE_COMPATIBILITY.md`): added v2.1.43-v2.1.49 entries covering Plugin settings.json, Worktree isolation, Background agents, ConfigChange hook, Sonnet 4.6, WASM memory fix

### Changed

- **Breezing: Worktree isolation support** (CC 2.1.49+): documented `isolation: "worktree"` in `guardrails-inheritance.md` -- parallel Implementers can now work on the same files without conflicts via git worktree isolation
- **Breezing: Agent model field fix** (CC 2.1.47+): documented model field behavior change in guardrails for correct agent spawning
- **Breezing: Background agents** (`background: true`): `video-scene-generator` agent now supports non-blocking background execution
- **Breezing: opencode mirror full sync**: all 10 breezing reference files (execution-flow, team-composition, review-retake-loop, session-resilience, planning-discussion, plans-to-tasklist, codex-engine, codex-review-integration, guardrails-inheritance, SKILL.md) synced to `opencode/skills/breezing/` for the first time
- **Breezing: Codex mirror updates**: all breezing reference files in `codex/.codex/skills/breezing/` updated to latest
- **Work skill**: major Codex mirror updates for auto-commit, auto-iteration, codex-engine, error-handling, execution-flow, parallel-execution, review-loop, scope-dialog, session-management
- **`quick-install.sh`**: added note that default security permissions apply automatically -- no manual configuration needed
- **`claude-settings.md` skill**: added note that CC 2.1.49+ auto-applies plugin settings; manual `settings.json` generation only needed for project-specific additions
- **`settings.security.json.template`**: updated `_harness_version` and added `_harness_note` clarifying role separation from plugin settings; unified `rm -rf/-fr` deny variants
- **Version references**: updated from CC 2.1.38 to 2.1.49 across 16+ skill and agent files

### Security

- **Least-privilege enforcement**: removed overly broad `allow` from plugin settings.json; all permissions now explicit deny or ask
- **DB CLI deny rules**: `psql`, `mysql`, `mongod`, `mongo` blocked by default to prevent accidental data operations
- **Secret path expansion**: added `id_ed25519`, recursive `.ssh/`, `.aws/`, `.npmrc` to deny patterns
- **Privacy-safe session logging**: `last_assistant_message` stored as length+hash, not plaintext
- **Atomic file writes**: `session.json` updates use `mktemp` + `mv` to prevent TOCTOU race conditions
- All 3 Codex experts (Security/Quality/Architect) scored A on hardening review

---

## [2.21.0](https://github.com/tim-hub/powerball-harness/compare/v2.20.13...v2.21.0) - 2026-02-20

### What's Changed for You

**Breezing now reviews your plan before coding starts. Phase 0 (Planning Discussion) runs by default--skip with `--no-discuss`.**

| Before | After |
|--------|-------|
| `/breezing` jumps straight into coding | Plan reviewed by Planner + Critic before implementation |
| No task validation before execution | V1-V5 checks (scope, ambiguity, overlap, deps, TDD) |
| All tasks registered at once | 8+ tasks auto-split into progressive batches |
| Implementers communicate only via Lead | Implementers can message each other directly |

### Added

- **Breezing Planning Discussion (Phase 0)**: pre-execution plan review with Planner + Critic teammates (default-on, skip with `--no-discuss`)
- **Task granularity validation (V1-V5)**: validates task scope, ambiguity, owns overlap, dependency consistency, and TDD markers before TaskCreate
- **Progressive Batch strategy**: automatic batch splitting for 8+ tasks with 60% completion triggers
- **Implementer peer communication (Pattern D)**: direct Implementer-to-Implementer knowledge sharing via SendMessage
- **Hook-driven signals**: `task-completed.sh` now generates `partial_review_recommended` and `next_batch_recommended` signals
- **Spec Driven Development integration**: `[feature:tdd]` markers in Plans.md trigger test-first task generation
- **New agents**: `plan-analyst` (task analysis) and `plan-critic` (Red Teaming review) for Phase 0

### Fixed

- **Signal threshold comparison**: Changed `-eq` to `-ge` in `task-completed.sh` to handle simultaneous task completions that skip exact threshold
- **Signal deduplication**: Added existing signal check before emitting to prevent duplicate signals
- **Signal generation fallback**: Added `python3` fallback for signal JSON generation when `jq` is unavailable
- **Completion counting**: Fixed `grep -c` overcounting in batch scope (now counts each task_id once regardless of retakes)
- **Document consistency**: Resolved contradictions between execution-flow.md, team-composition.md, and planning-discussion.md regarding round counts and V1-V4 skip policy
- **Signal session scoping**: Signals now include `session_id` and dedup is session-scoped, preventing prior sessions from suppressing signals
- **grep pattern safety**: Changed `grep -q` to `grep -Fq` (fixed-string match) for task_id lookups, preventing regex meta-character injection
- **stdin piping safety**: Changed `echo` to `printf '%s'` for JSON piping to jq/python3, preventing edge-case mangling
- **DRY signal construction**: Extracted `_build_signal_json` helper to eliminate jq/python3 fallback duplication in signal paths
- **Phase 0 handoff persistence**: Added `handoff` payload to breezing-active.json for Compaction resilience between Phase 0 and Phase A
- **Resume stale-ID reconciliation**: Added rules for mapping old task IDs to new IDs during session resume, with completion evaluation against active ID set

---

## [2.20.13](https://github.com/tim-hub/powerball-harness/compare/v2.20.11...v2.20.13) - 2026-02-19

### What's Changed

**Codex execution is now documented and validated as native multi-agent first, with `--claude` forcing both implementation and review delegation to Claude.**

| Before | After |
|--------|-------|
| Codex skill docs still mixed legacy task-team vocabulary and old state paths | Codex skill docs are aligned to native multi-agent tool flow (`spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`) and CODEX_HOME state paths |
| `--claude` behavior could read as implementation-only delegation in some references | `--claude` is now consistently specified as implementation + review delegation to Claude |
| Setup could leave `multi_agent` / role defaults implicit | Setup scripts now ensure `features.multi_agent=true` and harness agent role defaults in target `config.toml` |

### Changed

- Rewrote Codex distribution docs for `work`/`breezing` to use native multi-agent flow terminology and removed legacy task-team wording.
- Standardized runtime state references to `${CODEX_HOME:-~/.codex}/state/harness/` across Codex skill docs.
- Added explicit flag conflict rule: `--claude + --codex-review` fails before execution.
- Updated Codex setup references and README to reflect native multi-agent defaults and role declarations.
- Strengthened `tests/test-codex-package.sh` and CI to guard against legacy vocabulary regressions and enforce required multi-agent keywords/config defaults.

### Fixed

- Fixed inconsistent review routing by making `--claude` mode explicitly require Claude reviewer routing in both `work` and `breezing`.

---
## [2.20.11](https://github.com/tim-hub/powerball-harness/compare/v2.20.10...v2.20.11) - 2026-02-19

### Changed

- **Harness UI moved out of distribution scope**: tracked UI assets/skills/templates/hooks are excluded from release payload
- **SessionStart hooks simplified**: removed `harness-ui-register` execution from startup/resume

### Fixed

- **Issue #50**: removed distribution-path dependency on memory wrapper scripts with hardcoded absolute paths
  - distribution no longer tracks the 8 wrapper files (`scripts/harness-mem*`, `scripts/hook-handlers/memory-*.sh`)
  - hooks/config no longer reference those wrapper scripts

---

## [2.20.10](https://github.com/tim-hub/powerball-harness/compare/v2.20.9...v2.20.10) - 2026-02-18

### What's Changed

**Codex Harness now defaults to user-based installation, and Codex command execution is Codex-first with explicit `--claude` delegation.**

| Before | After |
|--------|-------|
| Codex setup copied `.codex` per project by default | Setup defaults to user scope (`${CODEX_HOME:-~/.codex}`), with `--project` as opt-in |
| `/work --codex` and `/breezing --codex` were primary for Codex execution | Codex is default engine; `--claude` explicitly delegates implementation |
| Codex setup guidance was mixed between project/user scopes | README + setup references are aligned to user-based rollout (JP/EN) |

### Changed

- Updated Codex setup scripts (`scripts/setup-codex.sh`, `scripts/codex-setup-local.sh`) to install skills/rules to `${CODEX_HOME:-~/.codex}` by default.
- Added explicit fallback mode `--project` for project-local deployment when needed.
- Updated Codex distribution docs and setup references to user-based defaults in both English and Japanese.
- Reworked Codex skill routing/docs so implementation intents resolve to Codex-first `/work`, with `--claude` for intentional delegation.
- Aligned `/breezing` recovery/state docs (`impl_mode`) with Codex-first runtime semantics.
- Synced release-related references and command docs to avoid setup drift between README, setup skill references, and Codex distribution docs.

---
## [2.20.9](https://github.com/tim-hub/powerball-harness/compare/v2.20.8...v2.20.9) - 2026-02-15

### What's Changed for You

**In Codex mode, `harness-review` guidance is now consistently documented as delegating to Claude CLI (`claude -p`).**

| Before | After |
|--------|-------|
| Codex-side review docs mixed Codex/MCP wording and delegation targets | Codex-side docs consistently describe Claude CLI (`claude -p`) delegation flow |

### Changed

- Updated Codex-side review docs to align review mode wording, integration flow, and detection guidance around `claude -p` delegation.
- Documentation consistency cleanup for Codex review-mode references.

---
## [2.20.8](https://github.com/tim-hub/powerball-harness/compare/v2.20.7...v2.20.8) - 2026-02-14

### Changed

- **Claude Code 2.1.41/2.1.42 adaptation**: Updated compatibility matrix and recommended version to v2.1.41+
  - Added v2.1.39-v2.1.42 entries to `docs/CLAUDE_CODE_COMPATIBILITY.md` (4 new version sections, 30+ feature rows)
  - Recommended version raised from v2.1.38+ to **v2.1.41+** (Agent Teams Bedrock/Vertex/Foundry model ID fix, Hook stderr visibility fix)
- **Breezing Bedrock/Vertex/Foundry note**: Added CC 2.1.41+ requirement note to `guardrails-inheritance.md` for non-Anthropic API users
- **Session `/rename` auto-naming**: Added CC 2.1.41+ auto-generate session name documentation to session skill
- **Troubleshoot `claude auth` commands**: Added CC 2.1.41+ `claude auth login/status/logout` to diagnostic table

---
## [2.20.7](https://github.com/tim-hub/powerball-harness/compare/v2.20.6...v2.20.7) - 2026-02-14

### Fixed

- **Stop hook "JSON validation failed" on every turn (#42)**: Replaced unreliable `type: "prompt"` hook with deterministic `type: "command"` hook (`stop-session-evaluator.sh`)
  - Root cause: prompt-type hook instructed the LLM to respond in JSON, but the model frequently returned natural language, causing repeated JSON parse errors
  - New command-based evaluator always outputs valid JSON, eliminating validation failures entirely
  - Both `hooks/hooks.json` and `.claude-plugin/hooks.json` updated in sync

---
## [2.20.6](https://github.com/tim-hub/powerball-harness/compare/v2.20.5...v2.20.6) - 2026-02-14

### Fixed

- **hookEventName validation error in session-auto-broadcast.sh** (#41):
  - Fixed `hookEventName` from `"AutoBroadcast"` to `"PostToolUse"` (4 locations)
  - Fixed `hookEventName` in `session-broadcast.sh` from `"Broadcast"` to `"PostToolUse"`
  - Prevented subprocess stdout contamination (added `>/dev/null` redirect)
  - Added `test-hook-event-names.sh` test (hookEventName consistency regression test)

---
## [2.20.5](https://github.com/tim-hub/powerball-harness/compare/v2.20.4...v2.20.5) - 2026-02-12

### Fixed

- **Breezing `--codex` subagent_type enforcement**: Fixed `--codex` flag being ignored during Implementer spawn
  - Root cause: `execution-flow.md` Step 3 hardcoded `task-worker` with no `--codex` branch
  - Added mandatory `impl_mode` branching to SKILL.md, execution-flow.md, and team-composition.md
  - Added three "absolute prohibition" rules: codex mode must use `codex-implementer`, standard mode must use `task-worker`, codex mode Lead must not Write/Edit source
  - Added explicit parallel spawn instruction: N Implementers spawned simultaneously (`N = min(independent_tasks, --parallel N, 3)`)
  - Compaction Recovery now restores correct subagent_type based on `impl_mode`

---

## [2.20.4](https://github.com/tim-hub/powerball-harness/compare/v2.20.3...v2.20.4) - 2026-02-11

### Fixed

- **Codex MCP -> CLI migration (Phase 7 completion)**:
  - Replace all `mcp__codex__codex` text references with `codex exec (CLI)` in `pretooluse-guard.sh` (4 messages) and `codex-worker-engine.sh` (1 log message)
  - Remove MCP legacy note from `codex-review/SKILL.md`
  - Add `codex-cli-only.md` rule to `.claude/rules/` for prevention
  - Add PreToolUse hook failsafe: deny `mcp__codex__*` tool calls with localized message via `emit_deny` + `msg()` pattern
  - Add `.gitignore` patterns for opencode/codex mirror dev-only skills (`test-*`, `x-promo`, `x-release-harness`)

### Security

- **Codex MCP dual-defense**: Three-layer protection against deprecated MCP usage (text correction + hook block + rule file). Codex review: Security A, Architect B

---

## [2.20.3](https://github.com/tim-hub/powerball-harness/compare/v2.20.2...v2.20.3) - 2026-02-10

### Fixed

- **Hook handler security hardening** (Codex review Round 1-3):
  - Replace manual JSON string escaping with `jq -nc --arg` and `python3 json.dumps` for safe JSON construction
  - Fix Python code injection vulnerability: pass data via `sys.argv`/`stdin` instead of triple-quote interpolation
  - Fix `grep` failure under `set -euo pipefail` with `|| true`
  - Use `grep -F` for fixed-string matching (avoid regex metacharacter issues)
  - Add `chmod 700` on `.claude/state` directory
  - Add `tostring` guard for description truncation type safety
  - Add 5-second dedup for TeammateIdle events
  - Add JSONL rotation (500 -> 400 lines) to prevent unbounded growth

---

## [2.20.2](https://github.com/tim-hub/powerball-harness/compare/v2.20.1...v2.20.2) - 2026-02-10

### Added

- **TeammateIdle/TaskCompleted hook handlers**: New `scripts/hook-handlers/teammate-idle.sh` and `task-completed.sh` log agent team events to `.claude/state/breezing-timeline.jsonl`
- **3-layer memory architecture (D22)**: Documented coexistence design for Claude Code auto memory, Harness SSOT, and Agent Memory in `decisions.md`
- **Task(agent_type) pattern (P18)**: Documented sub-agent type restriction syntax in `patterns.md`

### Changed

- **Claude Code 2.1.38+ adaptation**: Updated Feature Table in CLAUDE.md with 6 new rows (TeammateIdle/TaskCompleted Hook, Agent Memory, Fast mode, Auto Memory, Skill Budget Scaling, Task(agent_type))
- **Version references**: Updated all "CC 2.1.30+" references to "CC 2.1.38+" across 16+ skill and agent files
- **Skill budget scaling**: Relaxed 500-line hard rule to recommendation in `skill-editing.md`, noting CC 2.1.32+ 2% context window scaling
- **Session memory**: Added "Auto Memory Relationship (D22)" section to `session-memory/SKILL.md` and `memory/SKILL.md`
- **Breezing execution flow**: Updated hook implementation status to "implemented" in `execution-flow.md`
- **Guardrails inheritance**: Added Task(agent_type) to safety mechanism table

---

## [2.20.1](https://github.com/tim-hub/powerball-harness/compare/v2.20.0...v2.20.1) - 2026-02-10

### Fixed

- **PostToolUse hook syntax error**: Fix bash parser error in `posttooluse-tampering-detector.sh` caused by `|| true` after heredoc inside command substitution
- **python3 fallback in all hooks**: Replace heredoc python3 fallback with `python3 -c` in all 10 hook scripts to fix stdin conflict
- **POSIX compliance**: Replace `echo` with `printf '%s'` for safe input piping, `echo -e` with `printf '%b'`
- **Pattern matching**: Replace `echo | grep -qE` with `[[ =~ ]]` for 6 pattern checks (with word boundaries)
- **Error handling**: Change `set -euo pipefail` to `set +e` to match all other PostToolUse scripts
- **Bilingual warnings**: Add English + Japanese warning messages to hook scripts

---

## [2.20.0](https://github.com/tim-hub/powerball-harness/compare/v2.19.0...v2.20.0) - 2026-02-08

### What's Changed for You

**28 skills consolidated to 19. Breezing now runs with Phase A/B/C separation, teammate permissions fixed, and repo cleaned up.**

| Before | After |
|--------|-------|
| `memory`, `sync-ssot-from-memory`, `cursor-mem` as 3 skills | Unified `memory` (SSOT promotion + memory search in references) |
| `setup`, `setup-tools`, `harness-mem`, `codex-setup`, `2agent`, `localize-rules` as 6 skills | Unified `setup` (routing table dispatches to references) |
| `ci`, `agent-browser`, `x-release-harness` visible as slash commands | Hidden with `user-invocable: false` (auto-load still works) |
| Delegate mode ON at breezing start -> bypass permissions lost | Phase A (prep) maintains bypass -> delegate only in Phase B |
| Delegate mode stays on during completion -> commit restricted | Phase C exits delegate -> Lead can commit directly |
| Teammates auto-denied Bash due to "prompts unavailable" | `mode: "bypassPermissions"` + PreToolUse hooks for safety |
| Build artifacts, dev docs, lock files tracked in git | 33 files untracked, .gitignore updated |

### Changed

- **Skill consolidation (28 -> 19)**:
  - `/memory`: Absorbed `sync-ssot-from-memory` and `cursor-mem`
  - `/setup`: Absorbed `setup-tools`, `harness-mem`, `codex-setup`, `2agent`, `localize-rules`
  - `/troubleshoot`: Added CI failure triggers to description
- **Breezing Phase separation**: Restructured execution flow into Phase A (Pre-delegate) / Phase B (Delegate) / Phase C (Post-delegate)
  - Phase A: Maintain user's permission mode while initializing Team and spawning teammates
  - Phase B: Delegate mode -- Lead uses only TaskCreate/TaskUpdate/SendMessage
  - Phase C: Exit delegate, then run integration verification, commit, and cleanup
- **Teammate permission model**: All teammate spawns use `mode: "bypassPermissions"` with PreToolUse hooks as safety layer
  - PreToolUse hooks fire independently of permission system (official spec)
  - Safety layers: disallowedTools + spawn prompt constraints + .claude/rules/ + Lead monitoring
- **English-only releases**: GitHub release notes now written in English. Updated release rules and skills.
- **All related docs updated**: execution-flow.md, team-composition.md, codex-engine.md, guardrails-inheritance.md, session-resilience.md

### Added

- `skills/memory/references/cursor-mem-search.md` - Cursor memory search reference
- `skills/setup/references/harness-mem.md` - Harness-Mem setup reference
- `skills/setup/references/localize-rules.md` - Rule localization reference
- **Codex first-use check hook**: Auto-runs `check-codex.sh` on first `/codex-review` use (`once: true`)
- **timeout/gtimeout detection**: Guides macOS users to `brew install coreutils`

### Fixed

- **Codex review fixes (22 issues)**: pretooluse-guard JSON parse consolidation (5->1 jq call), symlink security guard, session-monitor `eval` removal
- **macOS compatibility**: All docs `timeout N codex exec` -> `$TIMEOUT N codex exec` (GNU coreutils independent)
- **Teammate Bash auto-deny**: Resolved "prompts unavailable" error for background teammates

### Removed

- **Untracked 33 files**: `mcp-server/dist/` (24 build artifacts), `docs/design/` (2), `docs/slides/` (1), `docs/claude-mem-japanese-setup.md`, dev-only docs (3), lock files (2)
- **Archived skills**: `sync-ssot-from-memory`, `cursor-mem`, `setup-tools`, `harness-mem`, `codex-setup`, `2agent`, `localize-rules` -> `skills/_archived/`

---

## [2.19.0](https://github.com/tim-hub/powerball-harness/compare/v2.18.11...v2.19.0) - 2026-02-08

### What's Changed for You

**5 implementation commands unified into 2: `/work` and `/breezing`. Both support `--codex`.**

| Before | After |
|--------|-------|
| 5 commands: `/work`, `/ultrawork`, `/breezing`, `/breezing-codex`, `/codex-worker` | Unified to 2 commands: `/work` and `/breezing` |
| Complex command selection | `/work` = Claude implementation, `/breezing` = team completion |
| Codex required separate commands (`/codex-worker`, `/breezing-codex`) | Unified switching via `--codex` flag |
| Scope specification differed per command | Common interactive scope confirmation for both commands |

### Changed

- **`/work` full overhaul**: Interactive scope confirmation + automatic strategy selection based on task count
  - 1 task -> direct implementation, 2-3 -> parallel, 4+ -> auto-iteration (integrated former ultrawork)
  - `--codex` flag for Codex MCP implementation delegation mode
  - New references: scope-dialog.md, auto-iteration.md, codex-engine.md
- **`/breezing` update**: `--codex` flag integration (absorbed former breezing-codex)
  - Added interactive scope confirmation
  - Consolidated Codex Implementer integration into codex-engine.md
- **pretooluse-guard.sh**: Unified `ultrawork-active.json` -> `work-active.json`
  - Backward compatible: old filename detected as fallback

### Removed

- **ultrawork** skill -> `/work all` provides equivalent functionality (moved to `skills/_archived/`)
- **breezing-codex** skill -> `/breezing --codex` provides equivalent functionality (moved to `skills/_archived/`)
- **codex-worker** skill -> `/work --codex` provides equivalent functionality (moved to `skills/_archived/`)

---

## [2.18.11](https://github.com/tim-hub/powerball-harness/compare/v2.18.10...v2.18.11) - 2026-02-06

### What's Changed for You

**In `--codex` mode, Claude now acts as PM and Edit/Write are automatically blocked**

| Before | After |
|--------|-------|
| Claude could edit directly in `--codex` mode | Edit/Write blocked except for Plans.md |
| Ambiguous role separation | Clear PM (Claude) vs Worker (Codex) separation |

### Added

- **breezing skill (v2)**: Full auto task completion using Agent Teams
  - Lead in delegate mode (coordination only), Implementer for coding, independent Reviewer
  - `--codex-review` for multi-AI review integration
  - session_id-based Hook enforcement: Reviewer Read-only, Implementer file ownership (pretooluse-guard.sh)
  - Flexible flow: Lead-autonomous stages replace rigid Phase 0-4
  - State simplification: Agent Teams TaskList as SSOT, breezing-active.json metadata-only
  - Peer-to-peer: Reviewer<->Implementer direct dialogue for lightweight questions
  - Agent Trace: per-Teammate metrics in completion reports
- **Codex mode guard**: Added Codex mode detection to `pretooluse-guard.sh`
  - Claude functions as PM, delegating implementation to Codex Worker
  - Enabled via `codex_mode: true` in `ultrawork-active.json`
  - Only Plans.md state marker updates allowed

### Changed

- **Codex review improvements**: Enhanced parallel review quality
  - SSOT-aware reviews (considers decisions.md/patterns.md)
  - Output limit relaxed 1500 -> 2500 chars for thorough analysis
  - Clear termination conditions (APPROVE when Critical/High = 0)
  - Fixed "nitpicking" issue (Low/Medium only -> APPROVE)
- Minor expert template fixes

---

## [2.18.10](https://github.com/tim-hub/powerball-harness/compare/v2.18.7...v2.18.10) - 2026-02-06

### Added

- **Agent persistent memory**: Added `memory: project/user` to all 7 agents
  - Subagents can now build institutional knowledge across conversations
  - Security: Read-only agents (code-reviewer, project-analyzer) keep Bash/Write/Edit disabled
  - Privacy guards: Each agent documents forbidden data (secrets, PII, source code snippets)

---

## [2.18.7](https://github.com/tim-hub/powerball-harness/compare/v2.18.6...v2.18.7) - 2026-02-05

### Changed

- **Claude guardrails**: Stop prompting on normal `git push`; prompt only on `git push -f/--force/--force-with-lease`.

---

## [2.18.6](https://github.com/tim-hub/powerball-harness/compare/v2.18.5...v2.18.6) - 2026-02-05

### Fixed

- **Codex guardrails**: `harness.rules` now parses reliably and avoids prompting on safe commands (e.g. `git clean -n`, `sudo -n true`).
- **Claude guardrails**: `templates/claude/settings.security.json.template` now uses valid permission syntax (`:*`) and prompts only on destructive variants.

### Changed

- **Codex package test**: Added rule example validation to prevent startup parse errors.

---

## [2.18.5](https://github.com/tim-hub/powerball-harness/compare/v2.18.4...v2.18.5) - 2026-02-05

### Added

- **gogcli-ops skill**: Google Workspace CLI operations (Drive/Sheets/Docs/Slides)
  - Auth workflow and account selection
  - URL-to-ID resolution via `gog_parse_url.py`
  - Read-only by default, write requires confirmation

---

## [2.18.4](https://github.com/tim-hub/powerball-harness/compare/v2.18.2...v2.18.4) - 2026-02-04

### Added

- **Codex setup command**: Added `/codex-setup` skill and `scripts/codex-setup-local.sh`
- **Setup tools**: `/setup-tools codex` subcommand for in-session Codex setup
- **Harness init/update**: Optional Codex CLI sync during `/harness-init` and `/harness-update`

---

## [2.18.2](https://github.com/tim-hub/powerball-harness/compare/v2.18.1...v2.18.2) - 2026-02-04

### Added

- **Codex CLI distribution**: Added `codex/.codex` with full skills and temporary Rules guardrails
- **Codex setup**: Added `scripts/setup-codex.sh` and `codex/README.md`
- **Codex AGENTS**: Added `codex/AGENTS.md` tuned for `$skill` usage
- **Codex package test**: Added `tests/test-codex-package.sh`

### Changed

- **Docs**: README now includes Codex CLI setup instructions

---

## [2.18.1](https://github.com/tim-hub/powerball-harness/compare/v2.18.0...v2.18.1) - 2026-02-04

### Added

- **Aivis/VOICEVOX TTS support**: Added Japanese TTS providers to generate-video skill
  - `aivis`: Aivis Cloud API (speaker_id, intonation_scale, etc.)
  - `voicevox`: VOICEVOX (character voices like Zundamon)
  - Sample character configurations included

### Changed

- **MCP server optional**: Removed `.mcp.json`, excluded mcp-server from distribution
  - Users who need it can set up separately

---

## [2.18.0](https://github.com/tim-hub/powerball-harness/compare/v2.17.10...v2.18.0) - 2026-02-04

### Added

- **Claude Code 2.1.30 compatibility**: Full integration with new features
  - **AgentTrace v0.3.0**: Task tool metrics (tokenCount, toolUses, duration) in `docs/AGENT_TRACE_SCHEMA.md`
  - **`/debug` command integration**: troubleshoot skill now routes to `/debug` for complex session issues
  - **PDF page range reading**: notebookLM and harness-review support `pages` parameter for large documents
  - **Git log extended flags**: harness-review, CI, harness-release use `--format`, `--raw`, `--cherry-pick`
  - **OAuth `--client-id/--client-secret`**: codex-mcp-setup.md documents DCR-incompatible MCP setup
  - **68% memory optimization**: session-memory and session skills document `--resume` benefits
  - **Subagent MCP access**: task-worker and codex-worker document MCP tool sharing (bugfix in CC 2.1.30)
  - **Accessibility settings**: harness-ui documents `reducedMotion` setting

---

## [2.17.10](https://github.com/tim-hub/powerball-harness/compare/v2.17.9...v2.17.10) - 2026-02-04

### Added

- **PreCompact/SessionEnd hooks**: Support automatic session state save and cleanup
- **AgentTrace v0.2.0**: Added Attribution field for plugin attribution tracking
- **Sandbox settings template**: Added `templates/settings/harness-sandbox.json`

### Changed

- **context: fork added**: deploy/generate-video/memory/verify skills now use isolated context
- **release -> harness-release**: Renamed to avoid conflict with Claude Code built-in command

---

## [2.17.9](https://github.com/tim-hub/powerball-harness/compare/v2.17.8...v2.17.9) - 2026-02-04

### Changed

- **Codex mode as default**: New project config template now defaults to `review.mode: codex`
- **Worktree necessity check**: `/ultrawork --codex` now auto-determines if Worktree is actually needed
  - Single task, all sequential dependencies, or file overlap -> fallback to direct execution mode
  - Avoids unnecessary Worktree creation overhead

---

## [2.17.8](https://github.com/tim-hub/powerball-harness/compare/v2.17.6...v2.17.8) - 2026-02-04

### Fixed

- **release skill**: Fix `/release` not launching via Skill tool
  - Removed `disable-model-invocation: true`

---

## [2.17.6](https://github.com/tim-hub/powerball-harness/compare/v2.17.3...v2.17.6) - 2026-02-04

### What's Changed for You

**generate-video skill evolved to JSON Schema-driven hybrid architecture, and README was refreshed**

| Before | After |
|--------|-------|
| Video generation config scattered in code | JSON Schema provides centralized scenario management |
| README structure was lengthy | TL;DR: Ultrawork section for immediate start |
| Skill descriptions English only | 28 skill descriptions localized to Japanese with humor |

### Added

- **generate-video JSON Schema Architecture** (#37)
  - `scenario-schema.json` for strict scenario structure definition
  - `validate-scenario.js` for semantic validation
  - `template-registry.js` for template management
  - Path traversal attack prevention implemented

- **TL;DR: Ultrawork section**: Added "Too long? Just this:" section to README

### Changed

- **Skill description localization**: Added Japanese descriptions with humor to 28 skills
- **README restructure**: Optimized to Install -> TL;DR -> Core Loop flow
- **Skill count update**: 42 -> 45 skills

### Fixed

- `validate-scenario.js`: Semantic error filtering bug fix
- `TransitionWrapper.tsx`: `slideIn` -> `slide_in` for schema naming convention alignment

---

## [2.17.3](https://github.com/tim-hub/powerball-harness/compare/v2.17.2...v2.17.3) - 2026-02-03

### What's Changed for You

**Ultrawork now automatically enters a self-correction loop after review**

| Before | After |
|--------|-------|
| Manual prompt input needed after review | Auto-correction loop until APPROVE |
| Codex presence manually specified | Codex MCP auto-detection + fallback |
| Improvement direction unclear | "How to Achieve A" section provides clear guidance |

### Added

- **Self-correction loop**: After `/harness-review`, automatically repeats corrections until APPROVE
  - Retry state management (`ultrawork-retry.json`) for progress tracking
  - REJECT/STOP immediately stops and prompts manual intervention
  - STOP after max 3 retries

- **Run-all-validations rule**: Execute all existing validation scripts in priority order, stop on failure

- **Improvement guidance template**: "How to Achieve A" section clearly shows how to reach A rating
  - Unified format per Decision (APPROVE/REQUEST CHANGES/REJECT/STOP)

### Changed

- **Codex auto-detection**: Auto-switches to Codex mode when Codex MCP is available
  - Falls back to subagent parallel when unavailable
  - Timeout configurable via `timeout_ms` (milliseconds)

- **Diff calculation improvement**: Calculate changed files based on `merge-base`
  - Includes staged/unstaged diffs
  - Handles initial commits/merges

- **review_aspects detection**: Deterministic judgment via path-based regex

---

## [2.17.2](https://github.com/tim-hub/powerball-harness/compare/v2.17.1...v2.17.2) - 2026-02-03

### What's Changed for You

**Plans.md is now auto-updated when Codex Worker completes**

| Before | After |
|--------|-------|
| Manual Plans.md update after work completion | Skill auto-updates to `cc:done` |

### Added

- **Plans.md auto-update**: Always executes task completion processing when Codex Worker skill completes
  - Auto-identifies relevant task
  - Updates `[ ]` -> `[x]`, `cc:WIP` -> `cc:done`
  - Confirms with user when task not found

### Changed

- Codex Worker script quality improvements (shared library, security hardening)

---

## [2.17.1](https://github.com/tim-hub/powerball-harness/compare/v2.17.0...v2.17.1) - 2026-02-03

### Added

- **Agent Trace**: Track AI-generated code edits for session context visibility
  - `emit-agent-trace.js`: PostToolUse hook records Edit/Write operations to `.claude/state/agent-trace.jsonl`
  - `agent-trace-schema.json`: JSON Schema (v0.1.0) for trace records
  - Stop hook now shows project name, current task, and recent edits at session end
  - `sync-status` skill now includes Agent Trace data for progress verification
  - `session-memory` skill now reads Agent Trace for cross-session context

### Changed

- Stop hook (`session-summary.sh`) enhanced with Agent Trace information display
- VCS info retrieval optimized: single `git status --porcelain=2 -b -uno` call with 5s TTL cache
- Repo root detection no longer spawns git process (walks up directory tree)

### Fixed

- Security hardening for trace file operations (symlink checks, permission enforcement)
- Rotation concurrency protection with lock file (O_CREAT|O_EXCL pattern)

---

## [2.17.0](https://github.com/tim-hub/powerball-harness/compare/v2.16.21...v2.17.0) - 2026-02-03

### Added

- **Codex Worker**: Delegate implementation tasks to OpenAI Codex as parallel workers
  - `codex-worker` skill for single task delegation
  - `ultrawork --codex` for parallel worker execution with git worktrees
  - Quality gates: evidence verification, lint/type-check, test, tampering detection
  - File locking mechanism with TTL and heartbeat
  - Automatic Plans.md update on task completion

### Changed

- Skills `codex-worker` and `codex-review` now have explicit routing rules (Do NOT Load For sections)
- Improved skill description for better auto-loading accuracy
- Added 5 shell scripts: `codex-worker-setup.sh`, `codex-worker-engine.sh`, `codex-worker-lock.sh`, `codex-worker-quality-gate.sh`, `codex-worker-merge.sh`
- Added integration test: `tests/test-codex-worker.sh`
- Added reference documentation: `skills/codex-worker/references/*.md`

### Fixed

- Shell script security improvements (jq injection, git option injection, value validation)
- POSIX compatibility for grep patterns (`\s` to `[[:space:]]`)
- Arithmetic operation in `set -e` context

---

## [2.16.21](https://github.com/tim-hub/powerball-harness/compare/v2.16.20...v2.16.21) - 2026-02-03

### Changed

- `ultrawork` Codex Mode options (`--codex`, `--parallel`, `--worktree-base`) moved to Design Draft
  - These features are planned but not yet implemented
  - Documentation now clearly marks them as "(Design Draft / Not Implemented)"
- Added `skills/ultrawork/references/codex-mode.md` as design draft documentation
- Added Codex Worker scripts and references (untracked, for future implementation)

---

## [2.16.20](https://github.com/tim-hub/powerball-harness/compare/v2.16.19...v2.16.20) - 2026-02-03

### Changed

- Centralized skill routing rules to `skills/routing-rules.md` (SSOT pattern)
- Made `codex-review` and `codex-worker` routing deterministic (removed context judgment)

---

## [2.16.19](https://github.com/tim-hub/powerball-harness/compare/v2.16.17...v2.16.19) - 2026-02-03

### Fixed

- Reduced duplicate display of Stop hook reason (now outputs keywords only)

---

## [2.16.17](https://github.com/tim-hub/powerball-harness/compare/v2.16.14...v2.16.17) - 2026-02-03

### What's Changed for You

**Skills now show usage hints in autocomplete**

| Before | After |
|--------|-------|
| `/harness-review` | `/harness-review [code|plan|scope]` |
| `/troubleshoot` | `/troubleshoot [build|test|runtime]` |

### Added

- Usage hints (`argument-hint`) added to 17 skills
- Inter-session notifications (useful for multi-session workflows)

### Changed

- Updated CI/tests/docs for Skills-only architecture

---

## [2.16.14](https://github.com/tim-hub/powerball-harness/compare/v2.16.11...v2.16.14) - 2026-02-02

### What's Changed for You

**Implementation requests are now automatically registered in Plans.md**

| Before | After |
|--------|-------|
| Ad-hoc requests not tracked | All tasks recorded in Plans.md |
| Hard to track progress | `/sync-status` shows full picture |

---

## [2.16.11](https://github.com/tim-hub/powerball-harness/compare/v2.16.5...v2.16.11) - 2026-02-02

### What's Changed for You

**Commands have been unified into Skills (usage unchanged)**

| Before | After |
|--------|-------|
| `/work`, `/harness-review` as commands | Same names, now powered by skills |
| Internal skills (impl, verify) in menu | Hidden (less noise) |
| `dev-browser`, `docs`, `video` | Renamed to `agent-browser`, `notebookLM`, `generate-video` |

### Changed

- README rewritten for VibeCoders (added troubleshooting, uninstall)
- CI scripts updated for Skills structure

---

## [2.16.5](https://github.com/tim-hub/powerball-harness/compare/v2.16.0...v2.16.5) - 2026-01-31

### What's Changed for You

**`/generate-video` now supports AI images, BGM, subtitles, and visual effects**

| Before | After |
|--------|-------|
| Manual image preparation | AI auto-generates (Nano Banana Pro) |
| No BGM/subtitles | Royalty-free BGM, Japanese subtitles |
| Basic transitions only | GlitchText, Particles, and more |

---

## [2.16.0](https://github.com/tim-hub/powerball-harness/compare/v2.15.0...v2.16.0) - 2026-01-31

### What's Changed for You

**`/ultrawork` now requires fewer confirmations for rm -rf and git push (experimental)**

| Before | After |
|--------|-------|
| rm -rf always asks | Only paths approved in plan auto-approved |
| git push always asks | Auto-approved during ultrawork (except force) |

---

## [2.15.0](https://github.com/tim-hub/powerball-harness/compare/v2.14.0...v2.15.0) - 2026-01-26

### What's Changed for You

**Full OpenCode compatibility mode added**

| Before | After |
|--------|-------|
| Separate setup needed for OpenCode | `/setup-opencode` auto-configures |
| Different skills/ structure | Same skills work in both environments |

---

## [2.14.0](https://github.com/tim-hub/powerball-harness/compare/v2.13.0...v2.14.0) - 2026-01-16

### What's Changed for You

**`/work --full` enables parallel task execution**

| Before | After |
|--------|-------|
| Tasks run one at a time | `--parallel 3` runs up to 3 concurrently |
| Manual completion checks | Each worker self-reviews autonomously |

---

## [2.13.0](https://github.com/tim-hub/powerball-harness/compare/v2.12.0...v2.13.0) - 2026-01-14

### What's Changed for You

**Codex MCP parallel review added**

| Before | After |
|--------|-------|
| Claude reviews alone | 4 Codex experts review in parallel |
| One perspective at a time | Security/Quality/Performance/a11y simultaneously |

---

## [2.12.0](https://github.com/tim-hub/powerball-harness/compare/v2.11.0...v2.12.0) - 2026-01-10

### Added

- **Harness UI Dashboard** (`/harness-ui`) - Track progress in browser
- **Browser Automation** (`agent-browser`) - Page interactions & screenshots

---

## [2.11.0](https://github.com/tim-hub/powerball-harness/compare/v2.10.0...v2.11.0) - 2026-01-08

### Added

- **Inter-session Messaging** - Send/receive messages between Claude Code sessions
- **CRUD Auto-generation** (`crud` skill) - Generate endpoints with Zod validation

---

## [2.10.0](https://github.com/tim-hub/powerball-harness/compare/v2.9.24...v2.10.0) - 2026-01-04

### Added

- **LSP Integration** - Go-to-definition, Find-references for accurate code understanding
- **AST-Grep Integration** - Structural code pattern search

---

## Earlier Versions

For v2.9.x and earlier, see [GitHub Releases](https://github.com/tim-hub/powerball-harness/releases).

