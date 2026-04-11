# Checklist: Translate Skills to English

## Content Quality

- [x] All 4 SSOT skill bodies (`harness-work`, `harness-review`, `harness-setup`, `harness-sync`) are in English
- [x] No Japanese characters remain in any translated SKILL.md body content
- [x] All `${VARIABLE}` template placeholders are preserved intact
- [x] All code blocks and shell commands are unchanged
- [x] All table structures and formatting are preserved

## Frontmatter Integrity

- [x] `description` fields are in English in all translated skills
- [x] `description-ja` fields (if present) are still present and in Japanese
- [x] `name` fields are unchanged

## Codex Variants

- [x] `skills-v3-codex/` directory has been removed
- [x] No references to removed codex files remain broken in other files

## Mirror Sync

- [x] `skills/harness-work/SKILL.md` is English (synced from v3)
- [x] `skills/harness-review/SKILL.md` is English (synced from v3)
- [x] `skills/harness-setup/SKILL.md` is English (synced from v3)
- [x] `skills/harness-sync/SKILL.md` is English (synced from v3)
- [x] `skills/harness-plan/SKILL.md` is English (re-synced from v3)
- [x] `skills/harness-release/SKILL.md` is English (re-synced from v3)

## Validation

- [x] `./tests/validate-plugin.sh` passes with no errors
