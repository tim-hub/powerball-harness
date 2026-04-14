# Claude Code Harness — Plans.md

Last archive: 2026-04-15 (Phase 35–48 → `.claude/memory/archive/Plans-2026-04-15-phase35-48.md`)
Last release: v4.2.0 on 2026-04-14 (Phase 49)

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
- `skills-codex/`: drop entirely (deprecated, codex-companion.sh bridges the gap)
- AGENTS.md: static template pointing to CLAUDE.md + agent role table (not generated)
- CI: replace mirror sync checks with template existence + setup idempotency tests

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 51.1 | Create `templates/codex/` with `config.toml`, `rules/harness.rules`, `.codexignore`, `AGENTS.md`, `README.md` from `codex/` | Files exist in `templates/codex/`; content matches originals | - | cc:TODO |
| 51.2 | Move opencode config to `templates/opencode/`: add `opencode.json`, `AGENTS.md`, `README.md`; commands already exist there | `templates/opencode/` has all config + commands | - | cc:TODO |
| 51.3 | Delete `codex/` directory | `codex/` no longer exists; git rm clean | 51.1 | cc:TODO |
| 51.4 | Delete `opencode/` directory | `opencode/` no longer exists | 51.2 | cc:TODO |
| 51.5 | Delete `skills-codex/` directory | `skills-codex/` no longer exists | - | cc:TODO |
| 51.6 | Implement `harness-setup codex` subcommand in SKILL.md | Checks codex installed; copies `templates/codex/*` → `.codex/`; copies `skills/` → `.codex/skills/` with `disable-model-invocation: true` patch; creates AGENTS.md pointing to CLAUDE.md | 51.1, 51.3 | cc:TODO |
| 51.7 | Implement `harness-setup opencode` subcommand in SKILL.md | Checks opencode installed; copies `templates/opencode/*` → `.opencode/`; copies `skills/` → `.opencode/skills/` as-is | 51.2, 51.4 | cc:TODO |
| 51.8 | Implement `harness-setup duo` subcommand | Runs both codex + opencode setup | 51.6, 51.7 | cc:TODO |
| 51.9 | Remove mirror sync scripts: `sync-skill-mirrors.mjs`, `build-opencode.mjs`, `sync-skills.mjs`, `validate-opencode.mjs` | Scripts deleted; no remaining references | 51.3, 51.4 | cc:TODO |
| 51.10 | Remove mirror sync CI: update `compatibility-check.yml`, `check-consistency.sh` mirror section, `validate-plugin.sh` opencode refs | CI passes without mirror checks | 51.9 | cc:TODO |
| 51.11 | Add template existence check to `validate-plugin.sh` | `templates/codex/config.toml` and `templates/opencode/opencode.json` verified in CI | 51.10 | cc:TODO |
| 51.12 | Add `codex/`, `opencode/`, `skills-codex/` to `deleted-concepts.yaml` | `check-residue.sh` 0 detections on HEAD | 51.3–51.5 | cc:TODO |
| 51.13 | Update `tests/test-codex-package.sh` — remove refs to deleted paths | Test passes; no references to `codex/.codex/skills/` | 51.3 | cc:TODO |
| 51.14 | Update CHANGELOG [Unreleased] and run full validation | `validate-plugin.sh` + `check-consistency.sh` + `check-residue.sh` all pass | 51.1–51.13 | cc:TODO |

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

