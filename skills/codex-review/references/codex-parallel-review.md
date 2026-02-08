# Codex Parallel Review Execution Guide

Orchestration steps for calling multiple experts in parallel via Codex CLI in Codex mode.

## Overview

In Codex mode, Claude acts as an orchestrator, calling **4 experts** in parallel via Codex CLI (`codex exec`). Different experts are used depending on the review type.

## Review Types and 4 Experts

| Review Type | 4 Experts | Expert Files |
|-------------|-----------|--------------|
| **Code** | Security, Performance, Quality, Accessibility | `security-expert.md`, `performance-expert.md`, `quality-expert.md`, `accessibility-expert.md` |
| **Plan** | Clarity, Feasibility, Dependencies, Acceptance | `clarity-expert.md`, `feasibility-expert.md`, `dependencies-expert.md`, `acceptance-expert.md` |
| **Scope** | Scope-creep, Priority, Feasibility, Impact | `scope-creep-expert.md`, `priority-expert.md`, `scope-feasibility-expert.md`, `impact-expert.md` |

```
Claude (Orchestrator)
    ↓
Determine review type
    ↓
Parallel CLI calls (4 experts)
    ├── Expert 1
    ├── Expert 2
    ├── Expert 3
    └── Expert 4
    ↓
Aggregate results → Judgment
```

---

## ⚠️ Mandatory Parallel Invocation Rules (MANDATORY)

**These rules MUST be followed. Violations will significantly degrade review quality.**

### Prohibited Practices

| Prohibited | Reason |
|------|------|
| ❌ Combining multiple experts in a single CLI call | Each expert's specialization is diluted |
| ❌ Requesting "check security, performance, and quality" in one call | Aspects are mixed, preventing deep analysis |
| ❌ Sending generic prompts without reading experts/*.md | Expert prompt insights are not utilized |

### Required Practices

| Required | Method |
|------|------|
| ✅ Execute each expert via **individual CLI calls** | Call `codex exec` 4 times based on review type |
| ✅ Read prompts **individually from experts/*.md** | `security-expert.md` → Security call → `performance-expert.md` → Performance call... |
| ✅ **Execute 4 CLI calls in parallel via Bash background processes** | Use `&` + `wait` for parallel execution |

### Correct Execution Pattern

```bash
# 0. Detect timeout command (macOS: brew install coreutils)
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")

# 1. Determine enabled experts based on config and project type
# 2. Write each expert prompt to temp file with shared constraints prepended
# 3. Execute only enabled experts in parallel:

for expert in security performance quality accessibility; do
  $TIMEOUT 120 codex exec "$(cat /tmp/expert-${expert}-prompt.md)" \
    > /tmp/expert-${expert}-result.txt 2>/dev/null &
done
wait

# 4. Collect results
for expert in security performance quality accessibility; do
  echo "=== ${expert} ==="
  cat /tmp/expert-${expert}-result.txt
done

# → Parallel execution of only necessary experts
# → Skip irrelevant aspects to reduce cost
```

### Why Separate Them?

| Combined in one call | Separated into 4 calls |
|------------------|-----------------|
| Each aspect gets 2-3 lines, shallow | Each aspect is analyzed in detail |
| Important issues are missed | Expert perspective catches everything |
| Tends to end with "no issues" | Provides specific improvement suggestions |

---

## Execution Flow

### Step 0: Check Remaining Context

Before Codex parallel review, **execute /compact if remaining context is ≤ 30%**.

> **Note**: If context is still tight after /compact, proceed to Step 1 anyway.

### Step 1: Confirm Configuration

```yaml
# Read from .claude-code-harness.config.yaml
review:
  mode: codex
  codex:
    enabled: true
    # Code Review experts (used when review type = code)
    code_experts:
      security: true
      accessibility: true
      performance: true
      quality: true
    # Plan Review experts (used when review type = plan)
    plan_experts:
      clarity: true
      feasibility: true
      dependencies: true
      acceptance: true
    # Scope Review experts (used when review type = scope)
    scope_experts:
      scope_creep: true
      priority: true
      feasibility: true
      impact: true
```

> **Note**: Legacy configuration (directly under `experts:`) is also supported for backward compatibility.

### Step 2: Collect Changed Files

```bash
# Get changed files via git diff
git diff --name-only HEAD~1
```

### Step 3: Determine Which Experts to Call (Filtering)

**Do not call all experts every time; select only necessary ones.**

#### 3.1 Configuration-Based Filtering

Exclude experts set to `false` in `.claude-code-harness.config.yaml`:

```yaml
experts:
  security: true       # ✅ Call
  accessibility: false # ❌ Skip
  performance: true    # ✅ Call
  ...
```

#### 3.2 Auto-Exclusion by Project Type

| Project Type | Auto-Excluded Experts |
|-----------------|------------------------|
| CLI / Backend API | Accessibility, SEO |
| Library / SDK | Accessibility, SEO |
| Web Frontend | (All enabled) |
| Mobile App | SEO |
| Plan/Review only | Security, Performance, Quality (when no code changes) |

**Detection method**:
```
1. Check changed file paths:
   - src/components/, pages/, app/ → Web frontend → Accessibility, SEO enabled
   - src/api/, server/, cli/ → Backend/CLI → Accessibility, SEO excluded
   - *.md only → Documentation changes → Prioritize Quality, Architect, Plan Reviewer, Scope Analyst

2. Check package.json / pyproject.toml:
   - react, vue, next → Web frontend
   - express, fastify, flask → Backend
   - commander, yargs → CLI
```

#### 3.3 Exclusion by Change Content

| Change Content | Prioritized Experts | Excludable |
|---------|---------------------|---------|
| Plans.md only | Plan Reviewer, Scope Analyst | Security, Performance, Quality, Accessibility, SEO, Architect |
| Test files only | Security, Quality, Performance | Architect, Plan Reviewer, Scope Analyst |
| README / Docs only | Quality, Architect, Plan Reviewer, Scope Analyst | Security, Performance, Accessibility |

#### 3.4 Final Call List Decision

```
Determine review type → Call corresponding 4 experts

Example 1: After /work with code changes → Code Review
→ Security, Performance, Quality, Accessibility
→ 4 experts in parallel

Example 2: After /plan-with-agent → Plan Review
→ Clarity, Feasibility, Dependencies, Acceptance
→ 4 experts in parallel

Example 3: After task addition → Scope Review
→ Scope-creep, Priority, Feasibility, Impact
→ 4 experts in parallel
```

### Step 4: Prepare Expert Prompts (Extended Version)

**Only for experts determined in Step 3**, read prompts from `experts/*.md` and expand the following variables:

| Variable | Content | Retrieval Method | Note |
|------|------|----------|------|
| `{files}` | Changed files list | `git diff --name-only HEAD~1` | Required |
| `{tech_stack}` | **Detailed tech stack** | Detected from package.json/pyproject.toml | Required |
| `{plan_content}` | Plans.md content | For Plan Reviewer | When applicable |
| `{requirements}` | Requirements content | For Scope Analyst | When applicable |

**Read `experts/_shared-constraints.md` and prepare it as `base-instructions` for all MCP calls.**

> **Note**: Diff and SSOT context are **injected directly to Codex by the orchestrator**. No need to add placeholders in expert templates.

#### 4.1 Get Diffs (Important for Context Provision)

To provide sufficient context to Codex, pass not only file names but also **diff content**:

```bash
# Changed files list
FILES=$(git diff --name-only HEAD~1)

# Diff content (max 200 lines per file)
for file in $FILES; do
  echo "=== $file ==="
  git diff HEAD~1 -- "$file" | head -200
done
```

> **Note**: If diff is too large, truncate to 200 lines to control prompt size.

#### 4.2 Detailed Tech Stack Detection

Extract major dependencies from package.json or pyproject.toml:

```bash
# Detect major frameworks from package.json
jq -r '.dependencies // {} | keys | map(select(
  test("react|vue|next|nuxt|express|fastify|nest|prisma|drizzle|trpc|zod|typescript|tailwind")
)) | join(", ")' package.json
```

**Example output**: `"react, next, prisma, zod, typescript, tailwind"`

#### 4.3 SSOT Context Injection (Prevent Irrelevant Feedback)

Communicate project decisions and reusable patterns to Codex:

```markdown
## Project Context (SSOT)

### Recent Decisions (decisions.md)
- D-12: OAuth authentication adopted, no additional auth features needed
- D-15: Standardized eager loading with Prisma

### Relevant Patterns (patterns.md)
- P-8: Error handling uses Result type
- P-12: API responses use unified format

Review with these project conventions in mind.
```

> **Effect**: By referencing SSOT, reviewers understand "why this implementation" and prevent off-target feedback.

### Step 5: Parallel CLI Calls

**Execution mode**: Parallel experts use **Bash background processes** (`&` + `wait`)

**Important**: Call only enabled experts determined in Step 3

Prepend shared constraints to each expert prompt file:

```bash
# Read shared constraints
SHARED=$(cat experts/_shared-constraints.md)

# Prepare prompt files for each enabled expert
for expert in "${ENABLED_EXPERTS[@]}"; do
  cat <<PROMPT > /tmp/expert-${expert}-prompt.md
${SHARED}

---

$(cat experts/${expert}-expert.md)

## Review Target
Files: ${FILES}
Tech Stack: ${TECH_STACK}

$(git diff HEAD~1)
PROMPT
done

# Detect timeout command (macOS: brew install coreutils)
TIMEOUT=$(command -v timeout || command -v gtimeout || echo "")

# Parallel execution (timeout 120s each)
for expert in "${ENABLED_EXPERTS[@]}"; do
  $TIMEOUT 120 codex exec "$(cat /tmp/expert-${expert}-prompt.md)" \
    > /tmp/expert-${expert}-result.txt 2>/dev/null &
done
wait

# Collect results
for expert in "${ENABLED_EXPERTS[@]}"; do
  echo "=== ${expert} ==="
  cat /tmp/expert-${expert}-result.txt
done
```

### Step 5.1: Output Limitation Rules (Prevent Context Overflow + Sufficient Analysis)

Each expert's response follows these constraints:

**Code Review Experts** (security/performance/quality/accessibility-expert.md):

| Constraint | Content |
|------|------|
| Language | **English only** (save tokens, Claude translates to Japanese during aggregation) |
| Max characters | **2500 chars** (increased for thorough analysis) |
| Score format | A-F |
| Count limit | Critical/High: all, **Medium: max 5**, Low: 3 |
| No issues | `Score: A / No issues.` only |
| **SSOT consideration** | Review considering project decisions |

**Plan/Scope Review Experts**:

| Constraint | Content |
|------|------|
| Language | **English only** |
| Max characters | **2500 chars** |
| Score format | **A-F** (aligned with Code Review) |
| No issues | `Score: A / No issues.` only |
| **SSOT consideration** | Review considering project decisions |

> **Reason**: Even with 4 experts in parallel, 2500 chars × 4 = 10,000 chars ≈ 3,300 tokens is acceptable. Increased from 1500 → 2500 for thorough analysis.

### Step 6: Aggregate Results

Consolidate results from each expert:

```markdown
## 📊 Codex Parallel Review Results

### Expert Summary

| Expert | Score | Critical | High | Medium | Low |
|--------|-------|----------|------|--------|-----|
| Security | B | 0 | 1 | 2 | 3 |
| Accessibility | A | 0 | 0 | 1 | 2 |
| Performance | C | 0 | 2 | 3 | 1 |
| Quality | B | 0 | 0 | 4 | 5 |
| SEO | A | 0 | 0 | 0 | 2 |
| Architect | B | 0 | 1 | 1 | 0 |
| Plan Reviewer | APPROVE | - | - | - | - |
| Scope Analyst | Proceed | - | - | - | - |

### Aggregated Findings

| # | Expert | Severity | File | Issue |
|---|--------|----------|------|-------|
| 1 | Security | High | src/api/auth.ts:45 | SQL Injection |
| 2 | Performance | High | src/api/posts.ts:23 | N+1 Query |
| 3 | Architect | High | src/services/ | Circular dependency |
```

### Step 7: Commit Judgment (Clarified Exit Conditions)

Calculate final judgment from aggregated results:

| Aggregate | Judgment | Action |
|------|------|-----------|
| Critical ≥ 1 | REJECT | Manual intervention required, exit loop |
| High ≥ 1 | REQUEST CHANGES | Auto-fix loop (max 3 times) |
| **Critical = 0 AND High = 0** | **APPROVE** | **Exit loop, ready to commit** |

#### Clarified Exit Conditions (Important)

**Review loop exit conditions** (loop ends when any of these is met):

1. ✅ **APPROVE**: Critical = 0 AND High = 0
2. ❌ **REJECT**: Critical ≥ 1 (manual intervention required)
3. ⏹️ **STOP**: Verification failed (lint/test errors)
4. 🔄 **Retry limit**: High issues remain after 3 auto-fix attempts

#### Handling "Nitpicking" Problem

**Low/Medium-only findings are treated as "nitpicking" and result in APPROVE if Critical/High = 0.**

| Situation | Judgment | Reason |
|------|------|------|
| Medium: 5, Low: 10 | **APPROVE** | Acceptable since Critical/High = 0 |
| High: 1, Medium: 0 | REQUEST CHANGES | Fix needed due to High |
| Critical: 1, others: 0 | REJECT | Manual handling required |

> **Principle**: To prevent endless reviews, **Low/Medium are acceptable for next iteration**.

### Step 7.1: Judgment Output Template (Unified)

Judgment output aligns with **commit-judgment-logic unified template**.

#### APPROVE

```markdown
### 🎯 How to Achieve A

**Decision**: APPROVE
**Grade**: A
**A Grade Criteria**:
- Critical: 0 ✅
- High: 0 ✅
- Medium: ≤5 ✅
**Required fixes**: None
**Next action**: Ready to `git commit`
```

#### REQUEST CHANGES

```markdown
### 🎯 How to Achieve A

**Decision**: REQUEST CHANGES
**Grade**: [grade]
**A Grade Criteria**:
- Critical: 0 [status]
- High: 0 [status]
- Medium: ≤5 [status]
**Required fixes**:
1. [file:line] - [issue] → [fix]
```

#### REJECT

```markdown
### Manual Intervention Required

**Decision**: REJECT
**Grade**: F
**Reason**: Critical issues require manual review and fix
**Critical issues**: ...
```

#### STOP

```markdown
### Verification Failed

**Decision**: STOP
**Grade**: N/A (blocked)
**Failure Type**: [lint_failure | test_failure | environment_error]
**Failed command**: ...
**Required fixes**: ...
```

## Error Handling

### When Some Experts Fail

```markdown
⚠️ Some experts encountered errors

| Expert | Status |
|--------|--------|
| Security | ✅ Success |
| Performance | ❌ Timeout |
| Quality | ✅ Success |

Skip failed experts and continue judgment?
```

### When All Experts Fail

```markdown
❌ Failed to communicate with Codex experts

Cause: MCP server connection error

Fallback: Execute review with Claude alone?
```

## Auto-Fix Loop

Auto-fix flow when REQUEST CHANGES judgment occurs:

```
REQUEST CHANGES judgment
    ↓
Extract fix targets (High/Medium findings)
    ↓
Claude executes fixes
    ↓
Re-run Codex parallel review
    │
    ├── APPROVE → Complete
    ├── REQUEST CHANGES → Loop (Retry: ${current}/${max_retries})
    └── REJECT → Manual handling needed
```

### Retry Limit

- Default: max 3 times
- Configuration: `review.judgment.max_retries`

### When Retry Limit Exceeded

```markdown
## ⚠️ Auto-Fix Limit Reached

Attempted ${max_retries} auto-fixes, but the following issues remain:

| # | Severity | File | Issue |
|---|----------|------|-------|
| 1 | High | src/api/users.ts | N+1 Query |

**Recommended actions**:
1. Manually fix the above
2. Re-run `/harness-review`
```

## Related Files

### Code Review Experts (4 types)

| File | Role |
|---------|------|
| `experts/security-expert.md` | Security expert prompt |
| `experts/performance-expert.md` | Performance expert prompt |
| `experts/quality-expert.md` | Quality expert prompt |
| `experts/accessibility-expert.md` | a11y expert prompt |

### Plan Review Experts (4 types)

| File | Role |
|---------|------|
| `experts/clarity-expert.md` | Clarity expert prompt |
| `experts/feasibility-expert.md` | Feasibility expert prompt |
| `experts/dependencies-expert.md` | Dependencies expert prompt |
| `experts/acceptance-expert.md` | Acceptance criteria expert prompt |

### Scope Review Experts (4 types)

| File | Role |
|---------|------|
| `experts/scope-creep-expert.md` | Scope creep detection expert prompt |
| `experts/priority-expert.md` | Priority analysis expert prompt |
| `experts/scope-feasibility-expert.md` | Scope feasibility expert prompt |
| `experts/impact-expert.md` | Impact analysis expert prompt |

### Other Experts (4 types)

| File | Role |
|---------|------|
| `experts/seo-expert.md` | SEO expert prompt |
| `experts/architect-expert.md` | Architecture expert prompt |
| `experts/plan-reviewer-expert.md` | Plan review expert prompt |
| `experts/scope-analyst-expert.md` | Requirements analysis expert prompt |

### Shared Constraints

| File | Role |
|---------|------|
| `experts/_shared-constraints.md` | Common constraints and output rules for all experts |

### Judgment Logic

| File | Role |
|---------|------|
| `commit-judgment-logic.md` | Commit judgment logic |
