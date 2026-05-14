# GitHub Release Notes Rules

Formatting rules applied when creating GitHub Release notes.

## Required Format

### Structure

```markdown
## What's Changed

**One-line description of the change's value**

### Before / After

| Before | After |
|--------|-------|
| Previous state | New state |
| ... | ... |

---

## Added

- **Feature name**: Description
  - Detail 1
  - Detail 2

## Changed

- **Change**: Description

## Fixed

- **Fix**: Description

## Requirements (if applicable)

- **Claude Code vX.X.X+** (recommended)
- Link: [Documentation](URL)
```

### Required Elements

| Element | Required | Description |
|---------|----------|-------------|
| `## What's Changed` | Yes | Section heading |
| **Bold summary** | Yes | One-line value description |
| `Before / After` table | Yes | User-facing changes |
| `Added/Changed/Fixed` | When applicable | Detailed changes |

### Language

- **GitHub Release**: English required (public repository)
- **CHANGELOG.md**: Detailed Before/After format
- Keep descriptions user-focused

## CHANGELOG Format (Detailed Before/After)

CHANGELOG describes each feature concretely using a "Before / After" format:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Theme: [One-line summary of the overall change]

**[Value to the user in 1-2 sentences]**

---

#### 1. [Feature Name]

**Before**: [Previous behavior. Concretely describe the inconvenience the user experienced]

**After**: [New behavior. What is resolved + concrete examples]

```Example output or command examples```

#### 2. [Next Feature Name]

**Before**: ...
**After**: ...
```

**Writing Rules**:
- Give each feature its own section with `#### N. Feature Name`
- "Before" should **describe the problem** (use the pattern "users had to...")
- "After" should provide a **concrete picture of the solution** (include command and output examples)
- Longer is OK. Readability is the top priority
- Keep technical details (file names, step numbers) to a minimum as supplementary notes in "After"

## Prohibited

- No skipping the Before / After (CHANGELOG) or Before / After table (GitHub Release)
- No technical-only descriptions (user perspective required)
- No bare change lists without value explanation

## Good Example (GitHub Release — English)

```markdown
## What's Changed

**`/work --full` now automates implement -> self-review -> improve -> commit in parallel**

### Before / After

| Before | After |
|--------|-------|
| `/work` executes tasks one at a time | `/work --full --parallel 3` runs in parallel |
| Reviews required separate manual step | Each task-worker self-reviews autonomously |
```

## Good Example (CHANGELOG)

```markdown
#### 1. Automatic Re-Ticketing of Failed Tasks

**Before**: When tests/CI failed, it would just retry 3 times and stop.
After stopping, you had to manually investigate the root cause and add fix tasks to Plans.md yourself.

**After**: When stopping after 3 failures, Harness classifies the failure cause and auto-generates fix task proposals.
Once approved, they are automatically added to Plans.md as `.fix` tasks.
```

## Bad Example

```markdown
## What's New

### Added
- Added task-worker.md
- Added --full option
```

-> Doesn't communicate user value

## Release Creation Command

```bash
gh release create vX.X.X \
  --title "vX.X.X - Title" \
  --notes "$(cat <<'EOF'
## What's Changed
...
EOF
)"
```

## Editing Past Releases

```bash
gh release edit vX.X.X --notes "$(cat <<'EOF'
...
EOF
)"
```

## CHANGELOG Pattern for CC Version Integration

For releases that include new Claude Code version integration, use the
**"CC Update -> Harness Usage" format** instead of the usual "Before / After" format.
By explaining from the upstream (CC) change rationale, readers can understand in context why the change is relevant to them.

### When to Apply

Apply this pattern when any of the following conditions are met:

- The Feature Table version labels have been updated
- New CC-derived events have been added to hooks.json
- Usage guides for new CC features have been added to skills

### Structure

```markdown
#### N. Claude Code X.Y.Z Integration

(One-line overview)

##### N-1. Feature Name

**CC Update**: What changed in Claude Code. Explain from the user's perspective so they understand what the feature does.

**Harness Usage**: How Harness leverages this change. Include specific mechanisms (script names, flow).

##### N-2. Next Feature Name

**CC Update**: ...
**Harness Usage**: ...
```

### Writing Rules

- Give each feature its own section with `##### N-X.`
- "CC Update" should describe **changes in user experience**, not file changes
- "Harness Usage" should describe **specific mechanisms** (what runs, what is prevented)
- Avoid listing file names. Write "Prevents Worker freeze" instead of "Updated hooks.json"
- Documentation-only changes (Feature Table updates, detail section additions) should not be separate entries; include them in the one-line overview at the top

### Good Example

```markdown
##### 5-1. Automatic Handling of MCP Elicitation

**CC Update**: MCP servers can now ask users "questions" during task execution (Elicitation).
For example, they may request form input like "Which repository should I push to?"

**Harness Usage**: Breezing Workers run in the background and cannot respond to question forms.
If left unhandled, the Worker freezes. Created elicitation-handler.sh to
auto-skip during Breezing sessions while passing through normally for user-interactive sessions.
```

### Bad Example

```markdown
#### CC 2.1.76 Integration

- Added Elicitation to hooks.json
- Created elicitation-handler.sh
- Updated CLAUDE.md
```

-> A list of file changes that fails to communicate why the change was needed or what changes for the user

## Reference

- Good examples: v2.8.0, v2.8.2, v2.9.1, v3.10.3 (CC integration pattern)
- Keep consistent with CHANGELOG
