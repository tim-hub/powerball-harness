# Claude Code Harness — Plans.md

Last archive: 2026-04-15 (Phase 35–48 → `.claude/memory/archive/Plans-2026-04-15-phase35-48.md`)
Last release: v4.2.0 on 2026-04-14 (Phase 49)

---

## Future Considerations

(none currently)

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

