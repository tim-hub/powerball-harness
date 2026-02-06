# Shared Expert Constraints

Common constraints and output rules for all Codex expert prompts.
The orchestrator should inject this as `base-instructions` when calling experts via MCP.

## CONSTRAINTS (Common)

- **English only, max 2500 chars** (increased for thorough analysis)
- Critical/High: report all, **Medium: max 5**, Low: max 3
- No issues → `Score: A / No issues.`
- **Consider project SSOT (decisions.md, patterns.md) when reviewing**
- Do not flag test files unless explicitly reviewing test quality
- Do not flag auto-generated code

## OUTPUT RULES

### Score

Rate A-F based on findings:
- **A**: No Critical/High issues
- **B**: No Critical, 1-2 High
- **C**: No Critical, 3+ High
- **D**: 1 Critical
- **F**: 2+ Critical

### Summary (append to every response)

```
### Summary

- Critical: X
- High: X
- Medium: X
- Low: X
```
