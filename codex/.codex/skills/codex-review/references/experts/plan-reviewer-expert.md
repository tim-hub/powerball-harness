# Plan Reviewer Expert Prompt for Codex

Plan review prompt for Codex MCP.

## 7-Section Format

### TASK

Analyze work plans (Plans.md) and detect gaps, ambiguities, and missing context that could block implementation.

### EXPECTED OUTCOME

Report plan issues in the following format:
- **[APPROVE / REJECT]** verdict
- Issue list (Severity: Critical/High/Medium/Low)
- Improvement suggestions
- Plan score (A-F)

### CONTEXT

Review target:
- Plan file: {plan_content}
- Focus: Task definitions, acceptance criteria, dependencies, context

### CONSTRAINTS

- Evaluate from "can this actually be implemented?" perspective
- Avoid overly strict criteria

### MUST DO

1. **Task clarity**:
   - Does each task specify WHERE to look?
   - Can reference materials provide 90%+ confidence?

2. **Verification / acceptance criteria**:
   - Is there a concrete way to confirm completion?
   - Are acceptance criteria measurable/observable?

3. **Context completeness**:
   - Is there missing information that causes 10%+ uncertainty?
   - Are implicit assumptions made explicit?

4. **Big picture / workflow**:
   - Is the purpose clear?
   - Are inter-task dependencies defined?
   - Is "done" defined?

### MUST NOT DO

- Do not over-criticize simple single-task plans
- Do not demand re-explanation of obvious context
- Do not flag "missing details" when reference files exist

### OUTPUT FORMAT

```markdown
## Plan Review Results

**Verdict**: [APPROVE / REJECT]

**Score**: [A-F]

### Evaluation Summary

| Criteria | Assessment |
|----------|------------|
| Clarity | [Brief assessment] |
| Verifiability | [Brief assessment] |
| Completeness | [Brief assessment] |
| Big Picture | [Brief assessment] |

### Findings (if REJECT)

| # | Severity | Area | Issue | Suggestion |
|---|----------|------|-------|------------|
| 1 | High | Task Definition | Missing reference file | Add link to existing implementation |
```
