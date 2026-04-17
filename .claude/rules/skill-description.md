# Skill Description Format Rule

Rule for the `description:` field in every `SKILL.md` frontmatter.

Applies to all skills under `skills/` and `templates/codex-skills/`.

## Required Format

```
description: "<Capability summary>. Use when <trigger>."
when_to_use: "<trigger phrase 1>, <trigger phrase 2>, ..."
```

## Rules

### 1. Description must contain `Use when `

The description must include the phrase `Use when ` followed by the trigger condition.
A one-sentence capability summary may precede it.
Forbidden at the start: `Use this skill...`, `This skill...`, `the user mentions`, `the user asks`.

**Why**: `Use when` is the routing signal the auto-loader looks for. The capability summary preceding it gives humans and the model a quick "what is this?" before the trigger condition.

### 2. Trigger describes task shape — not user phrases

- Good: `Use when reviewing code, plans, or scope`
- Bad: `Use when the user says "review" or "check"`

**Why**: Auto-loading matches semantic task shape. Enumerating user phrases bloats the description without adding routing signal.

### 3. Capability summary leads the description (one sentence, ≤ 120 chars)

Open with what the skill does: `"<Does X for Y>. Use when <trigger>."`

Avoid starting with "This skill..." or "A skill that...". Use an active verb: "Manages", "Builds", "Runs", "Detects".

### 4. Exclusions: use `when_to_use` for routing signal, not `Do NOT load for`

Positive trigger phrases in `when_to_use` provide better routing signal than negative exclusions in the description. Drop `Do NOT load for` clauses entirely.

### 5. Length: `description` ≤ 300 chars; `len(description) + len(when_to_use)` ≤ 1536 chars (hard)

Hard-enforced by `local-scripts/audit-skill-descriptions.sh` (description > 300 chars fails).
Character budget check for combined length is validated in `tests/validate-plugin.sh`.

**Why**: The description is evaluated on every session. The combined budget keeps the routing overhead bounded.

## Good Examples

### Capability summary + trigger + when_to_use

```yaml
---
name: harness-review
description: "Multi-angle code and plan review with security, scope, and UI profiles. Use when reviewing code, plans, PRs, or running pre-merge quality gates."
when_to_use: "review code, review plan, review PR, security audit, pre-merge check, scope analysis, quality gate"
---
# Harness Review

...
```

description length: ~155 chars. when_to_use: ~88 chars. Combined: ~243 chars.

### Minimal (internal/auto-triggered skill)

```yaml
---
name: session-control
description: "Internal skill for --resume and --fork workflow boundaries in harness-work. Auto-triggered by orchestration."
when_to_use: "resume session, fork session"
---
```

description length: ~108 chars. when_to_use: ~28 chars. Combined: ~136 chars.

## Bad Example With Fix

Old `skills/auth/SKILL.md`:

```yaml
description: "Use this skill whenever the user mentions login, signup, authentication, OAuth, session management, payments..."
```

Problems:
- Starts with `Use this skill whenever` — violates Rule 1 (forbidden phrasing).
- Enumerates 10+ user phrases — violates Rule 2.
- 500+ characters — violates Rule 5.

Fixed:

```yaml
description: "Implements authentication, OAuth, sessions, payments, and billing. Use when adding auth flows, route protection, RBAC, or payment webhooks."
when_to_use: "authentication, OAuth, login, signup, payments, Stripe, billing, subscriptions, route protection, RBAC"
```

Combined: ~241 chars.

## Migration Steps

When updating an existing SKILL.md to conform:

1. Write a one-sentence capability summary: what does this skill do?
2. Append `Use when <trigger shape>` — describe the task, not user phrases.
3. Drop `Do NOT load for` clauses; add positive trigger phrases to `when_to_use` instead.
4. Run `local-scripts/audit-skill-descriptions.sh <skill-dir>` to verify conformance.

## Supersedes

`.claude/rules/skill-editing.md` Section 5 ("Description Best Practices") — the older "Use when user mentions..." recommendation there is replaced by this rule. That section will point here once both rules exist.

## Related

- `local-scripts/audit-skill-descriptions.sh` — automated enforcement (wired into `tests/validate-plugin.sh` Section 10 once Phase 44 rewrites are complete).
- `.claude/rules/skill-editing.md` — other SKILL.md editing rules (frontmatter fields, file size, `references/` layout).
- `.claude/rules/skill-quick-reference.md` — Quick Reference table format for multi-subcommand skills.
