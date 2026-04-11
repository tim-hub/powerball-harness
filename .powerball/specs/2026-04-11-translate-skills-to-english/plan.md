# Plan: Translate Skills from Japanese to English

## Goal

Translate all Japanese-language skill body content to English. `skills-v3/` is the SSOT (Single Source of Truth); `skills/` mirrors from it via `sync-v3-skill-mirrors.sh`.

## Scope

### Files to Translate (Phase 1: SSOT)
- `skills-v3/harness-work/SKILL.md` — fully Japanese body
- `skills-v3/harness-review/SKILL.md` — fully Japanese body
- `skills-v3/harness-setup/SKILL.md` — fully Japanese body
- `skills-v3/harness-sync/SKILL.md` — fully Japanese body

### Files to Remove (Phase 2: Codex Variants)
- `skills-v3-codex/harness-work/SKILL.md` — remove entire directory
- `skills-v3-codex/breezing/SKILL.md` — remove entire directory
- `skills-v3-codex/` — remove if empty

### Mirror Re-sync (Phase 3)
- Run `./scripts/sync-v3-skill-mirrors.sh` to propagate translated SSOT to `skills/`
- This also fixes the broken mirrors: harness-plan and harness-release (v3 is English but skills/ is still Japanese)

## Translation Rules

1. Translate body content to English; preserve all template variables (`${VARIABLE}`, `{{placeholder}}`)
2. Preserve all code blocks, shell commands, and file paths exactly
3. Keep frontmatter `description` in English (already maintained as `description-en` in many files)
4. Preserve `description-ja` field if it exists (don't remove the Japanese backup)
5. Preserve marker strings and structured format (tables, checklists, phase headers)
6. Do NOT add new content or restructure — translate only

## Architecture Notes

- `skills-v3/` → `skills/` sync is done via `./scripts/sync-v3-skill-mirrors.sh`
- The `harness-release` skill documents i18n mechanism: `description` is the active locale field
- Codex variants (`skills-v3-codex/`) are being removed, not translated
