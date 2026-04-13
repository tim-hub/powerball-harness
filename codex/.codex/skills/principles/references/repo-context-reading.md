---
name: core-read-repo-context
description: "Read and understand repository context (README, Plans.md, existing code). Use at session start, before starting a new task, or when understanding the project structure is needed."
allowed-tools: ["Read", "Grep", "Glob"]
---

# Read Repository Context

A skill for understanding the repository structure and context.
Use before starting work or before implementing a new feature.

---

## Input

- **Required**: Access to the repository root directory
- **Optional**: Focus on specific files or directories

---

## Output

Structured context information including repository understanding

---

## Execution Steps

### Step 1: Understand Basic Structure

```bash
# Directory structure
ls -la
find . -maxdepth 2 -type d | head -20

# Check main files
cat README.md 2>/dev/null | head -50
cat package.json 2>/dev/null | head -20
```

### Step 2: Check Workflow Files

```bash
# Plans.md status
cat Plans.md 2>/dev/null || echo "Plans.md not found"

# AGENTS.md role assignments
cat AGENTS.md 2>/dev/null | head -100 || echo "AGENTS.md not found"

# CLAUDE.md configuration
cat CLAUDE.md 2>/dev/null | head -50 || echo "CLAUDE.md not found"
```

### Step 3: Identify the Tech Stack

```bash
# Frontend
[ -f package.json ] && cat package.json | grep -E '"(react|vue|angular|next|nuxt)"'

# Backend
[ -f requirements.txt ] && head -10 requirements.txt
[ -f Gemfile ] && head -10 Gemfile
[ -f go.mod ] && head -10 go.mod

# Configuration files
[ -f tsconfig.json ] && echo "TypeScript project"
[ -f .eslintrc* ] && echo "ESLint configured"
[ -f tailwind.config.* ] && echo "Tailwind CSS"
```

### Step 4: Check Git Status

```bash
git status -sb
git log --oneline -5
git branch -a | head -10
```

---

## Output Format

```markdown
## Repository Context

### Basic Information
- **Project name**: {{name}}
- **Tech stack**: {{framework}} + {{language}}
- **Current branch**: {{branch}}

### Workflow State
- **Plans.md**: {{exists/does not exist, task count}}
- **AGENTS.md**: {{exists/does not exist}}
- **CLAUDE.md**: {{exists/does not exist}}

### Recent Changes
{{last 3 commits}}

### Important Files
{{list of key files to be aware of}}
```

---

## When to Use

1. **At session start**: Understand the current state
2. **Before implementing a new feature**: Check consistency with existing code
3. **During error investigation**: Identify related files
4. **During review**: Understand the scope of changes

---

## Notes

- **Large repositories**: Focus on important sections when there are many files
- **Sensitive information**: Do not read contents of .env or secrets/
- **Cache utilization**: Minimize re-reading within the same session
