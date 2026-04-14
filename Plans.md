# Powerball Harness — Plans.md

Last archive: 2026-04-15 (Phase 35–48 → `.claude/memory/archive/Plans-2026-04-15-phase35-48.md`)
Last release: v4.3.0 on 2026-04-15 (Phase 50+51)

---

## Phase 52: Marketplace restructure — move plugin files to `harness/` subfolder

Created: 2026-04-15

Goal: Restructure the repo from single-plugin (`source: "./"`) to multi-plugin marketplace (`source: "./harness/"`). Move all harness-plugin-specific directories into a `harness/` subfolder. Keep repo-level files (docs, tests, Go source, CI, README, CHANGELOG, Plans.md) at root. Move `assets/` under `docs/`.

Design decisions (confirmed with Opus agent):
- `.claude-plugin/` at root keeps ONLY `marketplace.json`; plugin-specific `plugin.json`, `hooks.json`, `settings.json` move to `harness/.claude-plugin/`
- `.claude/rules/` stays at root (Claude Code reads rules from project root; SSOT for both dev and distribution)
- `.claude/scripts/` stays at root (dev/CI scripts for this repo)
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
| 52.1 | Move `skills/` → `harness/skills/` | `harness/skills/` has all 25+ skill dirs; root `skills/` removed | - | cc:TODO |
| 52.2 | Move `agents/` → `harness/agents/` | `harness/agents/` has all agent .md files; root `agents/` removed | - | cc:TODO |
| 52.3 | Move `templates/` → `harness/templates/` | `harness/templates/` has codex/, opencode/, codex-skills/, modes/ | - | cc:TODO |
| 52.4 | Move `scripts/` → `harness/scripts/` | `harness/scripts/` has all script files; root `scripts/` removed | - | cc:TODO |
| 52.5 | Move `bin/` → `harness/bin/` | `harness/bin/` exists; root `bin/` removed | - | cc:TODO |
| 52.6 | Move `output-styles/` → `harness/output-styles/` | `harness/output-styles/` has all style files; root removed | - | cc:TODO |
| 52.7 | Move `workflows/` → `harness/workflows/` | `harness/workflows/` has all workflow files; root removed | - | cc:TODO |
| 52.8 | Move `VERSION`, `harness.toml`, `claude-code-harness.config.*` → `harness/` | Files in `harness/`; root copies removed | - | cc:TODO |

### Batch 2: Restructure .claude-plugin/ and marketplace config

Per official plugin docs: `settings.json` at plugin root; `hooks/hooks.json` at plugin root; only `plugin.json` inside `.claude-plugin/` — but we don't use plugin.json so it gets deleted.

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 52.9 | Delete `.claude-plugin/plugin.json` (unused) | File no longer exists; root `.claude-plugin/` keeps only `marketplace.json` | - | cc:TODO |
| 52.10 | Move entire `hooks/` folder → `harness/hooks/` (hooks.json, BEST_PRACTICES.md, *.sh scripts) | `harness/hooks/` has all files; root `hooks/` removed | - | cc:TODO |
| 52.11 | Move `.claude-plugin/settings.json` → `harness/settings.json` (plugin root) | `harness/settings.json` exists; `.claude-plugin/settings.json` removed | - | cc:TODO |
| 52.12 | Update `marketplace.json`: `source: "./"` → `source: "./harness/"` and `outputStyles: "./harness/output-styles/"` | marketplace.json points to `./harness/` | 52.1–52.11 | cc:TODO |

### Batch 3: Move assets and clean up empty dirs

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 52.13 | Move `assets/` → `docs/assets/` | `docs/assets/` has all SVGs; root `assets/` removed | - | cc:TODO |
| 52.14 | Delete empty `codex/` and `opencode/` dirs if still present | No empty ghost directories | - | cc:TODO |

### Batch 4: Update all path references

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 52.15 | Update `CLAUDE.md` path references | All paths in CLAUDE.md resolve correctly | 52.1–52.14 | cc:TODO |
| 52.16 | Update `harness.toml` internal paths | `harness sync` produces correct output | 52.8 | cc:TODO |
| 52.17 | Update `scripts/sync-version.sh` — read `VERSION` from `harness/VERSION`, write `harness/.claude-plugin/marketplace.json` | Version sync works from new locations | 52.4, 52.8 | cc:TODO |
| 52.18 | Update `build-binary.sh` — change output from `bin/` → `harness/bin/`; no hook path changes needed (`CLAUDE_PLUGIN_ROOT` = harness/) | Binary lands in `harness/bin/harness-*` | 52.5 | cc:TODO |
| 52.19 | Update `sync.go`: (1) read `harness.toml` from `harness/harness.toml`; (2) remove `syncHooksJSON` (hooks/ is now inside plugin dir, nothing to sync) | `harness sync` runs cleanly; `go test ./cmd/harness/` passes | 52.8, 52.10 | cc:TODO |
| 52.20 | Update CI workflows (`.github/workflows/*.yml`) paths | CI passes | 52.1–52.14 | cc:TODO |
| 52.21 | Update test files (`tests/validate-plugin.sh`, `tests/test-codex-package.sh`, etc.) paths | All tests pass with new paths | 52.1–52.14 | cc:TODO |
| 52.22 | Update `.claude/scripts/check-consistency.sh` and `.claude/scripts/check-residue.sh` paths | Consistency check passes; residue check clean | 52.1–52.14 | cc:TODO |
| 52.23 | Update README.md, CONTRIBUTING.md, docs/ path references | All doc links resolve | 52.1–52.14 | cc:TODO |
| 52.24 | Update `docs/repository-structure.md` to reflect new layout | Matches actual directory tree | 52.1–52.14 | cc:TODO |
| 52.25 | Update skill SKILL.md files that reference `${CLAUDE_PLUGIN_ROOT}/scripts/` or sibling paths | Skills resolve correct paths | 52.1, 52.4 | cc:TODO |
| 52.26 | Update `deleted-concepts.yaml` with old root-level paths | `check-residue.sh` 0 detections on HEAD | 52.1–52.14 | cc:TODO |
| 52.27 | Update `.gitignore` — replace `bin/harness-*` with `harness/bin/harness-*` | gitignore covers new binary path | 52.5 | cc:TODO |

### Batch 5: Validation

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 52.28 | Full validation: `validate-plugin.sh` + `check-consistency.sh` + `check-residue.sh` + `test-codex-package.sh` | All pass (existing sandbox failures excepted) | 52.15–52.27 | cc:TODO |
| 52.29 | Update CHANGELOG `[Unreleased]` with Phase 52 entry | Before/After documented | 52.28 | cc:TODO |

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

