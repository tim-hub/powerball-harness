---
name: core-diff-aware-editing
description: "Edit files with minimal diffs, minimizing impact on existing code."
allowed-tools: ["Read", "Edit"]
---

# Diff-Aware Editing

A skill for making changes with minimal diffs when editing files.
Prevents destruction of existing code and produces review-friendly changes.

---

## Fundamental Principles

### 1. Read Before Edit

**Always read the target file before editing**

```
Bad example: Overwrite the entire file with the Write tool
Good example: Read -> Verify contents -> Change only necessary parts with Edit
```

### 2. Prefer Minimal Diffs

Keep changes to the bare minimum:

- Preserve existing indentation and formatting
- Keep existing comments
- Match the existing style

### 3. Change in Meaningful Units

```typescript
// Bad example: Mix unrelated changes
// Function addition + formatting changes + import cleanup

// Good example: Focus on one change
// Function addition only
```

---

## How to Use the Edit Tool

### Pattern 1: Simple Replacement

```
old_string: "const value = 1"
new_string: "const value = 2"
```

### Pattern 2: Adding a Code Block

```
old_string: "// TODO: implement feature"
new_string: "// Feature implemented
const feature = () => {
  // implementation
}"
```

### Pattern 3: Modifying a Function

```
old_string: "function getData() {
  return []
}"
new_string: "function getData() {
  const data = fetchData()
  return data
}"
```

---

## Patterns to Avoid

### 1. Rewriting the Entire File

```
Bad: Rewrite all 100 lines of a file with the Write tool
Good: Fix only the 5 lines that need changing with the Edit tool
```

### 2. Mixing in Formatting Changes

```
Bad: Change indentation at the same time as adding a feature
Good: Add the feature only. Handle formatting in a separate commit
```

### 3. Adding Unnecessary Blank Lines or Comments

```
Bad: Impose your own style
Good: Follow the existing style
```

---

## Pre-Edit Checklist

1. [ ] Verified the target file with Read
2. [ ] Identified the sections that need changing
3. [ ] Understood the existing style (indentation, naming conventions)
4. [ ] Confirmed the change is within paths.allowed_modify
5. [ ] Can visualize the behavior after the change

---

## Post-Edit Verification

```bash
# Verify the diff
git diff

# Check the number of changed lines (not too large?)
git diff --stat

# Check for syntax errors
npm run build 2>&1 | head -20
# or
npx tsc --noEmit
```

---

## Editing Multiple Files

When editing multiple files:

1. **Dependency order**: Type definitions -> Implementation -> Tests
2. **Ensure consistency**: Make related changes together
3. **Keep intermediate states working**: Maintain a buildable state after each edit

---

## Error Handling

When an edit produces an error:

1. **Re-check the original code**: Verify current state with Read
2. **Verify old_string match**: Ensure whitespace and line breaks are exact
3. **Try splitting**: Break large changes into smaller pieces
