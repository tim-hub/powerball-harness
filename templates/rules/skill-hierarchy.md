---
_harness_template: rules/skill-hierarchy.md
_harness_version: 2.6.1
---

# Skill Hierarchy Guidelines

## Overview

Skills in claude-code-harness follow a 2-tier structure of **parent skills (categories)** and **child skills (specific features)**.

```
skills/
├── impl/                      # Parent skill (SKILL.md)
│   ├── SKILL.md              # Category overview and routing
│   └── work-impl-feature/    # Child skill
│       └── doc.md            # Specific instructions
├── harness-review/
│   ├── SKILL.md
│   ├── code-review/
│   │   └── doc.md
│   └── security-review/
│       └── doc.md
...
```

## Required Rules

### 1. After Reading the Parent Skill, Also Read the Child Skill

After launching a parent skill with the Skill tool, **you must also Read the child skill (doc.md) that matches the user's intent**.

```
Correct flow:
1. Launch "impl" with the Skill tool -> Get SKILL.md content
2. Determine the user's intent (e.g., feature implementation)
3. Read work-impl-feature/doc.md with the Read tool
4. Follow the instructions in doc.md

Wrong:
1. Launch "impl" with the Skill tool
2. Start working after reading only SKILL.md (ignoring child skills)
```

### 2. Choosing the Right Child Skill

| User's Intent | Skill to Launch | Child Skill to Read |
|--------------|----------------|-------------------|
| "Implement a feature" | impl | work-impl-feature/doc.md |
| "Review the code" | harness-review | code-review/doc.md |
| "Security check" | harness-review | security-review/doc.md |
| "Build it" | verify | build-verify/doc.md |

### 3. When Multiple Child Skills Apply

Ask the user for clarification, or pick the most relevant one to start with.

---

## Why This Matters

- The parent SKILL.md provides only "overview and routing"
- Child doc.md files contain "specific instructions, checklists, and pattern collections"
- Without reading the child skill, work will be incomplete

---

## PostToolUse Hook Integration

A reminder is automatically displayed after using the Skill tool.
From the displayed list of child skills, Read the one that applies.
