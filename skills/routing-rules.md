# Skill Routing Rules (Reference)

Reference document for routing rules between skills.

> **SSOT location**: Each skill's `description` field is the SSOT for routing.
> This file is a reference providing detailed explanations and examples; actual routing depends on each skill's description.
>
> **Important**: Each skill's description and the "Do NOT Load For" table in its body must be in complete agreement.

## Codex-Related Routing

### harness-review (includes Codex review functionality)

**Purpose**: Provide second-opinion reviews via Codex CLI (`codex exec`) (integrated from `codex-review` in v3)

**Trigger keywords** (quoted from description):
- "review", "code review", "plan review"
- "scope analysis", "security", "performance"
- "quality checks", "PRs", "diffs"
- "/harness-review"

**Exclusion keywords** (quoted from description):
- "implementation", "new features", "bug fixes"
- "setup", "release"

### harness-work --codex (includes Codex implementation functionality)

**Purpose**: Use Codex as the implementation engine (integrated in v3)

**Trigger keywords**:
- "implement", "execute", "/work"
- "breezing", "team run"
- "--codex", "--parallel"

**Exclusion keywords** (quoted from description):
- "planning", "code review", "release"
- "setup", "initialization"

**Invocation**: Run with `/harness-work --codex`

## Routing Decision Flow (Reference)

> This section explains Claude Code's internal behavior and does not define additional keywords.
> Actual routing is determined solely by the keywords specified in each skill's description.

```
User input
    |
    |-- Matches trigger keywords in description -> Load the matching skill
    |-- Matches exclusion keywords in description -> Exclude the matching skill
    |-- Neither -> Standard skill matching
```

## Priority Rules (Reference)

Priority when keywords match multiple skills:

1. **Exclusions take highest priority**: Skills matching exclusion keywords are never loaded
2. **More specific keywords take priority**: Exact match > partial match

> **Note**: "Contextual inference" is not used to avoid ambiguity. Routing is deterministically decided by description keywords.

## Update Rules

1. **description = SSOT**: Each skill's `description` field is the authoritative routing definition
2. **Body consistency**: Each skill's "Do NOT Load For" table must exactly match its description
3. **Role of this file**: Reference for detailed explanations and decision flow (not the SSOT)
4. **Maintain complete lists**: Use specific keywords instead of generic expressions (e.g., avoid "all of ~")
