---
name: ci-analyze-failures
description: "Analyze CI failure logs and identify the cause. Use when tests or builds fail in a CI/CD pipeline."
allowed-tools: ["Read", "Grep", "Bash"]
---

# CI Analyze Failures

A skill for analyzing CI/CD pipeline failures and identifying causes.
Interprets logs from GitHub Actions, GitLab CI, and similar platforms.

---

## Input

- **CI logs**: Logs from the failed job
- **run_id**: CI run identifier (if available)
- **Repository context**: CI configuration files

---

## Output

- **Failure cause identification**: Specific cause
- **Fix suggestions**: Proposed remediation steps

---

## Execution Steps

### Step 1: Check CI Status

```bash
# For GitHub Actions
gh run list --limit 5

# Check latest failure
gh run view --log-failed
```

### Step 2: Retrieve Failure Logs

```bash
# Logs for a specific run
gh run view {{run_id}} --log

# Failed steps only
gh run view {{run_id}} --log-failed
```

### Step 3: Analyze Error Patterns

#### Build Errors

```
Pattern: "error TS\d+:" or "Build failed"
Possible causes:
- TypeScript type errors
- Missing dependencies
- Syntax errors
```

#### Test Errors

```
Pattern: "FAIL" or "✕" or "AssertionError"
Possible causes:
- Test failures
- Test timeouts
- Mock mismatches
```

#### Dependency Errors

```
Pattern: "npm ERR!" or "Could not resolve"
Possible causes:
- package.json inconsistencies
- Private package authentication
- Version conflicts
```

#### Environment Errors

```
Pattern: "not found" or "undefined"
Possible causes:
- Unset environment variables
- Missing secrets
- Path issues
```

### Step 4: Output Analysis Results

```markdown
## 🔍 CI Failure Analysis

**Run ID**: {{run_id}}
**Failure time**: {{timestamp}}
**Failed step**: {{step_name}}

### Cause Identification

**Error type**: {{build / test / dependency / environment}}

**Error message**:
```
{{core error message}}
```

**Cause analysis**:
{{specific cause explanation}}

### Related Files

| File | Relevance |
|------|-----------|
| `{{path}}` | {{relevance details}} |

### Fix Suggestions

1. {{specific fix step 1}}
2. {{specific fix step 2}}

### Auto-Fix Feasibility

- Auto-fix: {{possible / not possible}}
- Reason: {{reason}}
```

---

## Error Pattern Dictionary

### TypeScript Errors

| Error Code | Meaning | Typical Fix |
|-----------|---------|-------------|
| TS2304 | Name not found | Add import |
| TS2322 | Type mismatch | Fix type |
| TS2345 | Argument type mismatch | Fix argument |
| TS7006 | Implicit any | Add type annotation |

### npm Errors

| Error | Meaning | Typical Fix |
|-------|---------|-------------|
| ERESOLVE | Dependency resolution failure | Delete package-lock & reinstall |
| ENOENT | File not found | Check path |
| EACCES | Permission error | Check CI configuration |

### Jest/Vitest Errors

| Error | Meaning | Typical Fix |
|-------|---------|-------------|
| Timeout | Test timeout | Extend timeout or fix async |
| Snapshot | Snapshot mismatch | `npm test -- -u` |

---

## Multiple Error Priority

1. **Build errors**: Fix first with highest priority
2. **Dependency errors**: Must resolve before build
3. **Test errors**: Address after build succeeds
4. **Lint errors**: Address last

---

## Connecting to Next Actions

After analysis is complete:

> 📊 **Analysis Complete**
>
> **Cause**: {{cause summary}}
>
> **Next actions**:
> - "Fix it" -> Attempt auto-fix
> - "More details" -> Deeper analysis
> - "Skip" -> Switch to manual resolution

---

## Notes

- **Logs are large**: Extract the important parts
- **Watch for cascading errors**: Find the first error
- **Environment differences**: Consider differences between local and CI
