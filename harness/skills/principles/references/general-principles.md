---
name: core-general-principles
description: "Provides fundamental development principles and safety rules. Basic guidelines applied to all tasks."
---

# General Principles

Fundamental principles for using claude-code-harness. Applied to all workflows.

---

## Safety Principles

### 1. Verify Before Modifying

Before editing a file, always verify the following:

- **Check contents with the Read tool**: Understand existing code before making changes
- **Understand the scope of impact**: Consider how changes affect other files
- **Consider backups**: Check git status before making significant changes

### 2. Edit with Minimal Diffs

```
Bad example: Rewrite the entire file
Good example: Change only the necessary parts with the Edit tool
```

### 3. Respect Configuration Files

Follow the settings in `claude-code-harness.config.json`:

- `safety.mode`: dry-run / apply-local / apply-and-push
- `paths.protected`: Do not modify protected paths
- `paths.allowed_modify`: Only modify allowed paths

---

## Work Principles

### 1. Always Update Plans.md

- At task start: `cc:TODO` -> `cc:WIP`
- At task completion: `cc:WIP` -> `cc:done`
- When blocked: Add `blocked` with a reason

### 2. Proceed Incrementally

```
1. Research & Understand -> 2. Plan -> 3. Implement -> 4. Verify -> 5. Report
```

### 3. Error Handling

- Up to 3 automatic retries
- Escalate (report) if unresolved
- Clearly state the error and attempted fixes

---

## Communication Principles

### VibeCoder Support

To ensure understanding even without technical knowledge:

- **Avoid jargon**: Or add explanations alongside
- **Suggest next actions**: "Next, please say [something]"
- **Visualize progress**: Clearly show what's done and what remains

### Collaboration with PM (Cursor)

- **Share state via Plans.md**: Maintain as a single source of truth
- **Use `/handoff-to-cursor` for completion reports**: Follow the format
- **Stay within scope**: Get confirmation before working outside the requested scope

---

## Prohibited Actions

1. **Direct deployment to production** (staging only)
2. **Hardcoding sensitive information** (use .env)
3. **Modifying protected paths** (.github/, secrets/, etc.)
4. **Destructive operations without user confirmation** (rm -rf, etc.)
