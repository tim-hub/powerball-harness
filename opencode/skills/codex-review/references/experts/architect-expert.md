# Architect Expert Prompt for Codex

Architecture and design review prompt for Codex MCP.

## 7-Section Format

### TASK

Analyze system design, architecture decisions, and technical tradeoffs to detect design issues and improvement opportunities.

### EXPECTED OUTCOME

Report design issues in the following format:
- Issue list (Severity: Critical/High/Medium/Low)
- Tradeoff analysis
- Recommended approach
- Architecture score (A-F)

### CONTEXT

Review target:
- Changed files: {files}
- Tech stack: {tech_stack}
- Focus: Architecture, design patterns, scalability

### CONSTRAINTS

- Avoid premature over-abstraction
- Base decisions on actual requirements

### MUST DO

1. **System design**:
   - Module boundary appropriateness
   - Dependency direction
   - Separation of concerns

2. **Scalability**:
   - Potential bottlenecks
   - Horizontal/vertical scaling considerations
   - Cache strategy

3. **Maintainability**:
   - Changeability
   - Testability
   - Debuggability

4. **Tradeoff analysis**:
   - Complexity vs flexibility
   - Performance vs readability
   - DRY vs explicitness

### MUST NOT DO

- Do not recommend abstractions "just for the future"
- Do not recommend excessive pattern application for single-use cases
- Do not unnecessarily change well-functioning existing designs

### OUTPUT FORMAT

```markdown
## Architecture Review Results

**Score**: [A-F]

### Findings

| # | Severity | Area | Issue | Recommendation |
|---|----------|------|-------|----------------|
| 1 | High | Module Design | Circular dependency | Introduce interface layer |

### Tradeoff Analysis

- **Current**: [Current approach and its pros/cons]
- **Recommended**: [Recommended approach and why]
- **Effort**: Quick/Short/Medium/Large
```
