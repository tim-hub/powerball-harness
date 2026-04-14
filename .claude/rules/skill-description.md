# Skill Description Format Rule

Rule for the `description:` field in every `SKILL.md` frontmatter.

Applies to all skills under `skills/` and `templates/codex-skills/`.

## Required Format

```
description: "Use when <trigger>. [Optional: Do NOT load for: <exclusions>.]"
```

## Rules

### 1. Must start with the literal prefix `Use when `

Nothing comes before it. No `Use this skill...`, no `This skill...`, no `the user mentions`, no `the user asks`.

**Why**: The first words of the description are what the auto-loader reads first when deciding "is this skill relevant?". Starting with `Use when ` puts the trigger condition in the highest-signal position. Self-describing prose ("This skill...") wastes that position.

### 2. Trigger describes task shape or invocation — not user phrases

- Good: `Use when reviewing code, plans, or scope`
- Bad: `Use when the user says "review" or "check"`

**Why**: Auto-loading matches semantic task shape, not keywords. Enumerating user phrases (`the user mentions review, check, audit, inspect, look at...`) bloats the description without adding routing signal the model wasn't already going to pick up from the task shape.

### 3. Skill introduction / summary / capability list belongs in the SKILL.md body

Drop trailing sentences like `Implements X using Y` or `A collection of skills for Z` from the description. Move them into the first paragraph of the SKILL.md body (or a `## Overview` section).

**Why**: The description is always loaded into context (routing signal). The body is loaded only after routing succeeds (operating manual). Mixing introduction into the description wastes the always-loaded token budget on content that's only relevant once the skill has already been selected.

### 4. Exclusions are optional but recommended when another skill covers adjacent territory

Format: `Do NOT load for: <exclusion 1>, <exclusion 2>.`

Point to the sibling skill in parens if helpful: `Do NOT load for: planning (use harness-plan).`

**Why**: Exclusions prevent the auto-loader from picking this skill when an adjacent skill is a better fit. Skip exclusions for standalone skills with no near-neighbors.

### 5. Length: target ≤ 200 characters, ceiling 300 characters

Hard-enforced by `.claude/scripts/audit-skill-descriptions.sh`: anything > 300 chars fails.

**Why**: The description field is evaluated on every session. Keeping it tight preserves context budget for everything else. If a description doesn't fit within 300 chars, the summary sentence is the first thing to cut — never the trigger or the exclusions.

## Good Examples

### Trigger + exclusions, both present

```yaml
---
name: harness-review
description: "Use when reviewing code, plans, or scope — pre-merge quality gate, security audit, or scope check. Do NOT load for: implementation (harness-work), planning (harness-plan), release (harness-release)."
---
# Harness Review

Multi-angle review skill covering code, plans, and scope with optional dual-reviewer and security profiles.
```

Length: ~205 chars. Body opens with the introduction that used to be in the description.

### Trigger only (standalone skill, no close siblings)

```yaml
---
name: memory
description: "Use when recording decisions, managing SSOT, searching memory, or promoting learnings to decisions.md / patterns.md."
---
# Memory Skills

SSOT (Single Source of Truth) and cross-tool memory management for Harness.
```

Length: ~117 chars.

## Bad Example With Fix

Current `skills/auth/SKILL.md`:

```yaml
description: "Use this skill whenever the user mentions login, signup, authentication, OAuth, session management, payments, subscriptions, billing, Stripe integration, or checkout flows. Also use when the user needs to protect routes, add role-based access, or implement payment webhooks. Do NOT load for: general UI components, database schema design, non-auth API endpoints, or business logic unrelated to auth/payments. Implements authentication and payment features using Clerk, Supabase Auth, or Stripe."
```

Problems:
- Starts with `Use this skill whenever the user mentions` — violates Rule 1.
- Enumerates 10+ user phrases — violates Rule 2.
- Trailing `Implements authentication and payment features using Clerk, Supabase Auth, or Stripe.` is body intro — violates Rule 3.
- 500+ characters — violates Rule 5.

Fixed description:

```yaml
description: "Use when implementing authentication, OAuth, sessions, payments, subscriptions, or billing — including route protection, RBAC, and payment webhooks. Do NOT load for: general UI, schema design, or non-auth API endpoints."
```

And move the capability description into the body:

```markdown
# Auth Skills

Implements authentication and payment features using Clerk, Supabase Auth, or Stripe.

## Feature Details
...
```

Length after fix: ~223 chars.

## Migration Steps

When updating an existing SKILL.md to conform:

1. Identify the skill's core trigger shape — what task makes this skill relevant?
2. Rewrite `description` starting with `Use when <trigger>`.
3. Move any introduction/summary sentence into the SKILL.md body's first paragraph.
4. Keep exclusions if another skill covers adjacent territory; drop them if the skill is unique.
5. Run `.claude/scripts/audit-skill-descriptions.sh <skill-dir>` to verify conformance.

## Supersedes

`.claude/rules/skill-editing.md` Section 5 ("Description Best Practices") — the older "Use when user mentions..." recommendation there is replaced by this rule. That section will point here once both rules exist.

## Related

- `.claude/scripts/audit-skill-descriptions.sh` — automated enforcement (wired into `tests/validate-plugin.sh` Section 10 once Phase 44 rewrites are complete).
- `.claude/rules/skill-editing.md` — other SKILL.md editing rules (frontmatter fields, file size, `references/` layout).
- `.claude/rules/skill-quick-reference.md` — Quick Reference table format for multi-subcommand skills.
