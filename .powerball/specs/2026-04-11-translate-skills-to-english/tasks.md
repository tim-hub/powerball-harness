# Tasks: Translate Skills to English

## Phase 1: Translate SSOT Skills

| # | Task | Files | Done |
|---|------|-------|------|
| 1 | Translate `skills-v3/harness-work/SKILL.md` body from Japanese to English | `skills-v3/harness-work/SKILL.md` | [x] |
| 2 | Translate `skills-v3/harness-review/SKILL.md` body from Japanese to English | `skills-v3/harness-review/SKILL.md` | [x] |
| 3 | Translate `skills-v3/harness-setup/SKILL.md` body from Japanese to English | `skills-v3/harness-setup/SKILL.md` | [x] |
| 4 | Translate `skills-v3/harness-sync/SKILL.md` body from Japanese to English | `skills-v3/harness-sync/SKILL.md` | [x] |
| 4a | Translate `skills-v3/harness-plan/SKILL.md` body from Japanese to English | `skills-v3/harness-plan/SKILL.md` | [x] |
| 4b | Translate `skills-v3/harness-release/SKILL.md` body from Japanese to English | `skills-v3/harness-release/SKILL.md` | [x] |

## Phase 2: Remove Codex Variants

| # | Task | Files | Done |
|---|------|-------|------|
| 5 | Remove `skills-v3-codex/` directory entirely | `skills-v3-codex/` | [x] |

## Phase 3: Mirror Re-sync

| # | Task | Files | Done |
|---|------|-------|------|
| 6 | Run `./scripts/sync-v3-skill-mirrors.sh` to propagate SSOT to `skills/` mirrors | `skills/harness-*/SKILL.md` | [x] |

## Phase 4: Validate

| # | Task | Files | Done |
|---|------|-------|------|
| 7 | Run `./tests/validate-plugin.sh` and confirm all checks pass | — | [x] |
