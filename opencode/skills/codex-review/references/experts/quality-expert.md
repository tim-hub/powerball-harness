# Quality Expert Prompt for Codex

Code quality review prompt for Codex MCP.

## 7-Section Format

### TASK

Analyze code quality (readability, maintainability, best practices) and detect areas needing improvement.

### EXPECTED OUTCOME

Report quality issues in the following format:
- Issue list (Severity: Critical/High/Medium/Low)
- Specific fix proposals
- Quality score (A-F)

### CONTEXT

Review target:
- Changed files: {files}
- Tech stack: {tech_stack}
- Focus: Naming, structure, duplication, error handling

### CONSTRAINTS

- Respect existing project style
- Avoid excessive improvement suggestions

### MUST DO

1. **Readability**:
   - Ambiguous naming (x, tmp, data)
   - Functions longer than 50 lines
   - Deep nesting (4+ levels)
   - Magic numbers

2. **Maintainability**:
   - Duplicate code
   - Tight coupling
   - Excessive global state
   - Unused code

3. **Best practices**:
   - Empty catch blocks
   - Overuse of `any` type
   - Callback hell
   - Hard-to-test structures

4. **Cross-platform**:
   - Missing responsive design
   - 100vw scrollbar overflow issue
   - Touch targets too small

### MUST NOT DO

- Do not report style/format issues as High/Critical
- Do not flag auto-generated code quality
- Do not report duplication in test files as DRY violations

### OUTPUT FORMAT

```markdown
## Quality Review Results

**Score**: [A-F]

### Findings

| # | Severity | File | Line | Issue | Fix |
|---|----------|------|------|-------|-----|
| 1 | Medium | services/user.ts | 45 | Function too long (78 lines) | Split into smaller functions |
```
