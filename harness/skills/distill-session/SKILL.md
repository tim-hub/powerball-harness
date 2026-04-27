---
name: distill-session
description: Distills the current Claude Code session into a reusable project skill. Use this whenever the user says "save this", "distill this", "turn this into a skill", or at the end of a session that solved a non-trivial, repeatable problem. Trigger even without explicit ask if the session contained a clear reusable workflow worth capturing for this project.
model: opus
effort: xhigh
---

# Distill Session into a Project Skill

When invoked:

1. Review what happened this session. Identify the *workflow* — the repeatable pattern — not the specific files, names, or one-off context.

2. Decide whether it's worth saving. Skip and say so if:
   - It was Q&A or pure exploration
   - The same outcome would come from a one-line prompt

   If a similar skill already exists in `.claude/skills/`, **invoke `update-skill` instead** of creating a new one.

3. If worth saving, draft a SKILL.md with:
   - A short kebab-case `name`
   - A "pushy" description listing trigger phrases (skills under-trigger by default — be explicit about when to fire)
   - Instructions focused on what Claude wouldn't do correctly on its own. Skip generic advice.
   - Keep under 200 lines. Link out to `references/*.md` for anything longer.

4. Show me the draft. Wait for approval before writing.

5. On approval, write to `.claude/skills/<name>/SKILL.md` (relative to the project root). The skill is live in the current session immediately — no restart needed.

6. Remind me to commit it so the team gets it too.

## What not to do

- Don't write to `~/.claude/` — project skills only.
- Don't create one skill per session. If the pattern looks like an extension of an existing skill, edit that one instead.
- Don't include user-specific paths, credentials, or transcripts in the skill body.