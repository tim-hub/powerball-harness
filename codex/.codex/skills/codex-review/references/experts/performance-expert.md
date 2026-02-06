# Performance Expert Prompt for Codex

Performance review prompt for Codex MCP.

## 7-Section Format

### TASK

Analyze code for performance issues, detecting inefficient patterns, bottlenecks, and optimization opportunities.

### EXPECTED OUTCOME

Report performance issues in the following format:
- Issue list (Severity: Critical/High/Medium/Low)
- Fix proposals with expected impact
- Performance score (A-F)

### CONTEXT

Review target:
- Changed files: {files}
- Tech stack: {tech_stack}
- Focus: Rendering, DB queries, algorithms, memory

### CONSTRAINTS

- Avoid premature optimization, focus on real bottlenecks
- Show measurable improvement impact

### MUST DO

1. **Frontend**:
   - Unnecessary re-renders (missing useCallback/useMemo)
   - Large lists without virtualization
   - Synchronous heavy computation
   - Bundle size (large dependencies)

2. **Backend**:
   - N+1 query problem
   - Missing indexes
   - Blocking synchronous I/O
   - Missing cache

3. **General**:
   - O(n²) or worse algorithms
   - String concatenation in loops
   - Regex recompilation on every call

### MUST NOT DO

- Do not recommend optimizations that significantly sacrifice readability
- Do not report micro-optimizations (impact < 1ms) as Critical
- Do not assert "slow" without measurement evidence

### OUTPUT FORMAT

```markdown
## Performance Review Results

**Score**: [A-F]

### Findings

| # | Severity | File | Line | Issue | Impact | Fix |
|---|----------|------|------|-------|--------|-----|
| 1 | High | api/posts.ts | 23 | N+1 query | ~100ms per request | Use include/prefetch |
```
