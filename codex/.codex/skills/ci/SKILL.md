---
name: ci
description: "Use this skill whenever the user mentions CI failures, GitHub Actions errors, build pipeline issues, test failures in CI, red builds, or needs to debug why a pipeline broke. Also trigger when the user pastes CI logs or error output from a remote build. Do NOT load for: local builds, local test runs, standard implementation work, code reviews, or project setup. Diagnoses and fixes CI/CD pipeline failures — analyzes logs, identifies root causes, and applies fixes."
allowed-tools: ["Read", "Grep", "Bash", "Task"]
user-invocable: false
context: fork
argument-hint: "[analyze|fix|run]"
---

# CI/CD Skills

A collection of skills for resolving CI/CD pipeline issues.

---

## Trigger Conditions

- "CI is failing", "GitHub Actions failed"
- "Build error", "Tests aren't passing"
- "Fix the pipeline"

---

## Feature Details

| Feature | Details | Trigger |
|---------|--------|---------|
| **Failure Analysis** | See [references/analyzing-failures.md](${CLAUDE_SKILL_DIR}/references/analyzing-failures.md) | "Check the logs", "Investigate the cause" |
| **Test Fixes** | See [references/fixing-tests.md](${CLAUDE_SKILL_DIR}/references/fixing-tests.md) | "Fix the tests", "Suggest a fix" |

---

## Execution Steps

1. **Test vs Implementation assessment** (Step 0)
2. Classify the user's intent (analysis or fix)
3. Assess complexity (see below)
4. Read the appropriate reference file from "Feature Details" above, or launch the ci-cd-fixer sub-agent
5. Verify results and re-run if necessary

### Step 0: Test vs Implementation Assessment (Quality Gate)

When CI fails, first triage the root cause:

```
CI Failure Report
    ↓
┌─────────────────────────────────────────┐
│       Test vs Implementation Assessment │
├─────────────────────────────────────────┤
│  Analyze the cause of the error:        │
│  ├── Implementation is wrong → Fix the implementation │
│  ├── Tests are outdated → Confirm with user │
│  └── Environment issue → Fix the environment │
└─────────────────────────────────────────┘
```

#### Prohibited Actions (tampering prevention)

```markdown
⚠️ Prohibited Actions on CI Failure

The following "solutions" are prohibited:

| Prohibited | Example | Correct Response |
|-----------|---------|-----------------|
| Skipping tests | `it.skip(...)` | Fix the implementation |
| Removing assertions | Deleting `expect()` | Verify expected values |
| Bypassing CI checks | `continue-on-error` | Fix the root cause |
| Relaxing lint rules | `eslint-disable` | Fix the code |
```

#### Decision Flow

```markdown
🔴 CI is failing

**A decision is needed**:

1. **Implementation is wrong** → Fix the implementation ✅
2. **Test expectations are outdated** → Ask the user for confirmation
3. **Environment issue** → Fix environment settings

⚠️ Test tampering (skipping, removing assertions) is prohibited

Which case applies?
```

#### When Approval is Required

When test/config changes are unavoidable:

```markdown
## 🚨 Test/Config Change Approval Request

### Reason
[Why this change is necessary]

### Changes
[Diff]

### Alternative Considerations
- [ ] Confirmed that fixing the implementation cannot resolve this

Awaiting explicit user approval
```

### Using Git Log Extended Flags (CC 2.1.49+)

Use structured logs to identify the commit that caused a CI failure.

#### Identifying the Causal Commit

```bash
# Analyze commits in structured format
git log --format="%h|%s|%an|%ad" --date=short -10

# Chronological analysis in topological order
git log --topo-order --oneline -20

# Correlate changed files with the cause
git log --raw --oneline -5
```

#### Key Use Cases

| Use Case | Flag | Effect |
|----------|------|--------|
| **Identifying failure cause** | `--format="%h|%s"` | Structured commit listing |
| **Chronological tracking** | `--topo-order` | Tracking with merge order considered |
| **Understanding change impact** | `--raw` | Detailed file change display |
| **Merge-excluded analysis** | `--cherry-pick --no-merges` | Extract actual commits only |

#### Example Output

```markdown
🔍 CI Failure Root Cause Analysis

Recent commits (structured):
| Hash | Subject | Author | Date |
|------|---------|--------|------|
| a1b2c3d | feat: update API | Alice | 2026-02-04 |
| e4f5g6h | test: add tests | Bob | 2026-02-03 |

Changed files (--raw):
├── src/api/endpoint.ts (Modified) ← Type error occurred
├── tests/api.test.ts (Modified)
└── package.json (Modified)

→ Commit a1b2c3d is likely the cause
  Type error: src/api/endpoint.ts:42
```

## Sub-agent Integration

Launch the ci-cd-fixer via the Task tool when the following conditions are met:

- The fix → re-run → failure loop has occurred **2 or more times**
- Or, the error spans multiple files in a complex case

**Launch pattern:**

```
Task tool:
  subagent_type="ci-cd-fixer"
  prompt="Diagnose and fix the CI failure. Error log: {error_log}"
```

ci-cd-fixer operates safety-first (default dry-run mode).
See `agents/ci-cd-fixer.md` for details.

---

## For VibeCoders

```markdown
🔧 How to talk about CI failures

1. **"CI is down" / "It's red"**
   - Automated tests are failing

2. **"Why is it failing?"**
   - Investigate the cause

3. **"Fix it"**
   - Attempt an automatic fix

💡 Important: "Faking" a fix by tampering with tests is prohibited
   - ❌ Deleting or skipping tests
   - ⭕ Fixing the code properly

If you suspect the test itself is wrong,
verify first before deciding on a course of action
```
