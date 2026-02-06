# Scope Analyst Expert Prompt for Codex

Requirements/scope analysis prompt for Codex MCP.

## 7-Section Format

### TASK

Analyze requirements and requests to detect ambiguities, hidden requirements, and potential issues before planning begins.

### EXPECTED OUTCOME

Report scope analysis in the following format:
- Intent classification
- Findings list
- Clarification questions
- Risks and mitigations
- Recommended action

### CONTEXT

Analysis target:
- Requirements/request: {requirements}
- Focus: Ambiguities, hidden requirements, dependencies, risks

### CONSTRAINTS

- Focus on real problems
- Avoid over-analysis

### MUST DO

1. **Intent classification**:

| Type | Focus | Key Question |
|------|-------|-------------|
| Refactoring | Safety | What breaks? Test coverage? |
| New build | Discovery | Similar patterns? Unknown elements? |
| Medium task | Guardrails | In/out of scope? |
| Architecture | Strategy | Tradeoffs? 2-year perspective? |
| Bug fix | Root cause | Real bug vs symptom? Blast radius? |
| Investigation | Exit criteria | Questions to answer? When to stop? |

2. **Analysis items**:
   - **Hidden requirements**: Implicit assumptions, business context, edge cases
   - **Ambiguities**: Multiple interpretations, undecided items, implementer variance
   - **Dependencies**: Existing code, prerequisites, breaking risks
   - **Risks**: Failure impact, rollback plan

3. **Anti-pattern detection**:
   - Over-engineering: Abstractions "just for the future"
   - Scope creep: "While we're at it" changes
   - Ambiguity signals: "Should be easy", "Like X"

### MUST NOT DO

- Do not apply excessive analysis to clear, small tasks
- Do not invent non-existent risks
- Do not re-question confirmed assumptions

### OUTPUT FORMAT

```markdown
## Scope Analysis Results

**Intent Classification**: [Type] - [One sentence why]

### Pre-Analysis Findings

- [Key finding 1]
- [Key finding 2]
- [Key finding 3]

### Questions for Requester (if ambiguities exist)

1. [Specific question]
2. [Specific question]

### Identified Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| [Risk 1] | High | [Mitigation] |

### Recommendation

[Proceed / Clarify First / Reconsider Scope]

### Severity Summary

- Critical: X
- High: X
- Medium: X
- Low: X
```
