---
name: ci
description: "Diagnoses and fixes CI/CD failures from GitHub Actions logs. Use when a build is broken, CI is red, or a pipeline has failed."
when_to_use: "CI failed, build broken, GitHub Actions error, red build, pipeline failure, fix CI"
allowed-tools: ["Read", "Grep", "Bash", "Task"]
user-invocable: false
context: fork
argument-hint: "[analyze|fix|run]"
---

# CI/CD Skills

A collection of skills for resolving CI/CD pipeline issues — failure analysis, test fixes, and pipeline repair using structured git log and sub-agent delegation.

---

## Feature Details

| Feature | Details |
|---------|--------|
| **Failure Analysis** | See [references/analyzing-failures.md](${CLAUDE_SKILL_DIR}/references/analyzing-failures.md) |
| **Test Fixes** | See [references/fixing-tests.md](${CLAUDE_SKILL_DIR}/references/fixing-tests.md) |

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

Structured log commands to identify causal commits: see [references/analyzing-failures.md](${CLAUDE_SKILL_DIR}/references/analyzing-failures.md) (Git Log Extended Flags section).

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
