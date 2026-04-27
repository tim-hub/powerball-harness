---
name: update-skill
description: Updates an existing project skill in .claude/skills/ based on new learnings from the current session. Use whenever the user says "update the X skill", "improve this skill", "the X skill should also handle Y", or when distill-session identifies that a session's learnings extend an existing skill rather than warranting a new one. Prefer this over creating a new skill when the pattern overlaps with something already in .claude/skills/.
model: opus
effort: medium
---

# Update an Existing Project Skill

When invoked:

1. **Identify the target skill.**
   - If the user named one, use that.
   - Otherwise, list skills in `.claude/skills/` and propose the closest match based on this session's content. Confirm before proceeding.
   - If no existing skill is a good fit, stop and suggest running `distill-session` instead.

2. **Read the current SKILL.md fully** before proposing changes. Don't guess at its contents.

3. **Diagnose what kind of update this is.** Pick one:
   - **Description tightening** — skill exists but didn't trigger when it should have, or triggered when it shouldn't. Fix the description, not the body.
   - **New case / branch** — skill handles workflow A, this session revealed workflow B that belongs in the same skill. Add a section.
   - **Correction** — skill has instructions that produced a wrong result this session. Fix the specific instruction; don't rewrite around it.
   - **Reference extraction** — skill is getting long (>200 lines) and a chunk should move to `references/*.md` with a pointer from SKILL.md.

   State which kind out loud. The kind dictates the diff.

4. **Propose a minimal diff.** Show:
   - The exact lines being changed (before/after), not a rewrite of the whole file.
   - One-sentence rationale per change.
   - If touching the description, explain what trigger case it fixes.

5. **Wait for approval.** Then apply with targeted edits to `.claude/skills/<name>/SKILL.md`. The change is live immediately.

6. **Sanity check the result:**
   - SKILL.md still under 200 lines? If not, suggest extracting to `references/`.
   - Frontmatter still valid YAML with `name` and `description`?
   - Description still concrete and "pushy" with trigger phrases?

7. **Remind me to commit.** Mention what changed in one line so the commit message writes itself.

## What not to do

- Don't rewrite a skill from scratch when a few lines would do. Skills accumulate trust through stable instructions; churn breaks that.
- Don't expand scope. If the new learning isn't really the same workflow, stop and recommend a new skill via `distill-session`.
- Don't touch `~/.claude/` — project skills only.
- Don't add generic advice ("be careful", "think step by step"). Only add things Claude wouldn't do correctly on its own.
- Don't silently delete sections. If something should go, call it out and explain why.

## When in doubt

If the update would more than double the skill's size, or change its core purpose, that's a signal it's actually a *new* skill. Stop and say so.