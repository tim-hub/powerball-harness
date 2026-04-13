---
name: merge-plans
description: "Skill for merge-updating Plans.md (preserving user tasks). Use when multiple Plans.md files need to be consolidated."
allowed-tools: ["Read", "Write", "Edit"]
---

# Merge Plans Skill

A skill that applies the template structure while preserving user task data
when updating an existing Plans.md.

---

## Purpose

- Preserve user tasks (🔴🟡🟢📦 sections)
- Update template structure and marker definitions
- Update last-modified information

---

## Plans.md Structure

```markdown
# Plans.md - Task Management

> **Project**: {{PROJECT_NAME}}
> **Last Updated**: {{DATE}}
> **Updated By**: Claude Code

---

## 🔴 In-Progress Tasks         <- User data (preserved)

## 🟡 Not-Started Tasks         <- User data (preserved)

## 🟢 Completed Tasks           <- User data (preserved)

## 📦 Archive                   <- User data (preserved)

## Marker Legend                 <- Updated from template

## Last Update Info              <- Date updated
```

---

## Merge Algorithm

### Step 1: Section Splitting

```
Split the existing Plans.md into the following sections:

1. Header section (# Plans.md ... ---)
2. 🔴 In-Progress Tasks (until next section)
3. 🟡 Not-Started Tasks (until next section)
4. 🟢 Completed Tasks (until next section)
5. 📦 Archive (until next section)
6. Marker Legend (until next section)
7. Last Update Info (until end of file)
```

### Step 2: Task Section Extraction

```bash
extract_section() {
  local file="$1"
  local start_marker="$2"
  local end_markers="$3"  # Pipe-delimited end markers

  awk -v start="$start_marker" -v ends="$end_markers" '
    BEGIN { in_section = 0; split(ends, end_arr, "|") }
    $0 ~ start { in_section = 1; next }
    in_section {
      for (i in end_arr) {
        if ($0 ~ end_arr[i]) { in_section = 0; exit }
      }
      if (in_section) print
    }
  ' "$file"
}

# Extract each section
TASKS_WIP=$(extract_section "$PLANS_FILE" "## 🔴" "## 🟡|## 🟢|## 📦|## Marker|---")
TASKS_TODO=$(extract_section "$PLANS_FILE" "## 🟡" "## 🔴|## 🟢|## 📦|## Marker|---")
TASKS_DONE=$(extract_section "$PLANS_FILE" "## 🟢" "## 🔴|## 🟡|## 📦|## Marker|---")
TASKS_ARCHIVE=$(extract_section "$PLANS_FILE" "## 📦" "## 🔴|## 🟡|## 🟢|## Marker|---")
```

### Step 3: Task Validation

```bash
# Verify non-empty
count_tasks() {
  echo "$1" | grep -c "^\s*- \[" || echo "0"
}

WIP_COUNT=$(count_tasks "$TASKS_WIP")
TODO_COUNT=$(count_tasks "$TASKS_TODO")
DONE_COUNT=$(count_tasks "$TASKS_DONE")
ARCHIVE_COUNT=$(count_tasks "$TASKS_ARCHIVE")

echo "Tasks to be preserved:"
echo "  In-progress: $WIP_COUNT"
echo "  Not started: $TODO_COUNT"
echo "  Completed: $DONE_COUNT"
echo "  Archived: $ARCHIVE_COUNT"
```

### Step 4: Generate New Plans.md

```markdown
# Plans.md - Task Management

> **Project**: {{PROJECT_NAME}}
> **Last Updated**: {{DATE}}
> **Updated By**: Claude Code

---

## 🔴 In-Progress Tasks

<!-- List cc:WIP tasks here -->

{{TASKS_WIP}}

---

## 🟡 Not-Started Tasks

<!-- List cc:TODO, pm:requested (compatible: cursor:requested) tasks here -->

{{TASKS_TODO}}

---

## 🟢 Completed Tasks

<!-- List cc:done, pm:confirmed (compatible: cursor:confirmed) tasks here -->

{{TASKS_DONE}}

---

## 📦 Archive

<!-- Move old completed tasks here -->

{{TASKS_ARCHIVE}}

---

## Marker Legend

| Marker | Meaning |
|--------|---------|
| `pm:requested` | Task requested by PM (compatible: cursor:requested) |
| `cc:TODO` | Claude Code not started |
| `cc:WIP` | Claude Code in progress |
| `cc:done` | Claude Code completed (awaiting confirmation) |
| `pm:confirmed` | PM confirmed complete (compatible: cursor:confirmed) |
| `cursor:requested` | (Compatible) Synonym for pm:requested |
| `cursor:confirmed` | (Compatible) Synonym for pm:confirmed |
| `blocked` | Blocked (include reason) |

---

## Last Update Info

- **Updated**: {{DATE}}
- **Last Session By**: Claude Code
- **Branch**: main
- **Update Type**: Plugin update
```

---

## Empty Section Handling

When a task section is empty, insert default text:

```markdown
## 🔴 In-Progress Tasks

<!-- List cc:WIP tasks here -->

(None currently)
```

---

## Error Handling

### When Plans.md Cannot Be Parsed

```bash
if ! validate_plans_structure "$PLANS_FILE"; then
  echo "Warning: Could not parse Plans.md structure"
  echo "Keeping backup and using new template instead"

  # Backup
  cp "$PLANS_FILE" "${PLANS_FILE}.bak.$(date +%Y%m%d%H%M%S)"

  # Use template
  use_template_instead=true
fi
```

### When Required Sections Are Missing

Missing sections are filled with template defaults.

---

## Output

| Field | Description |
|-------|-------------|
| `merge_successful` | Merge success flag |
| `tasks_wip_count` | In-progress task count |
| `tasks_todo_count` | Not-started task count |
| `tasks_done_count` | Completed task count |
| `tasks_archive_count` | Archived task count |
| `backup_created` | Whether backup was created |

---

## Usage Example

```bash
# Invoke the skill
merge_plans \
  --existing "./Plans.md" \
  --template "$PLUGIN_PATH/templates/Plans.md.template" \
  --output "./Plans.md" \
  --project-name "my-project" \
  --date "$(date +%Y-%m-%d)"
```

---

## Related Skills

- `update-2agent-files` - Overall update flow
- `generate-workflow-files` - New file generation
