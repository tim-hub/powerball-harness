# Claude Code Harness — Plans.md

Last archive: 2026-04-12 (Phase 25–34 → `.claude/memory/archive/Plans-2026-04-12-phase25-34.md`)

---

## Future Considerations

- **Codex CLI support**: Codex platform dirs were removed in 3.17.2. If needed later, restore `codex/.codex/skills/` symlinks, `scripts/codex-companion.sh`, and `--codex` flags in harness-work/breezing. The skill routing rules in `skills/routing-rules.md` still reference Codex patterns.

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
