# Claude Code Harness — Plans.md

Last archive: 2026-04-12 (Phase 25–34 → `.claude/memory/archive/Plans-2026-04-12-phase25-34.md`)

---

## Future Considerations

(none currently)

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
