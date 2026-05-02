# Universal Violations Session Injection

Cross-task violation injection — Reviewer findings scoped as `universal` are prepended to subsequent Worker briefings in the same breezing session to prevent recurring mistakes.

---

In breezing mode, Reviewer findings can carry a `scope` field that controls whether the violation is injected into subsequent Workers in the same session.

| `scope` value | Meaning | Effect |
|---------------|---------|--------|
| `task-specific` (default) | Violation is relevant only to this task | Not propagated |
| `universal` | Violation is a pattern that must never recur in any task | Prepended to all subsequent Worker briefings in the session |

**When to mark `universal`**: Use for structural violations that are likely to recur — e.g., "always use parameterized queries", "never hardcode API keys", "always validate input at system boundaries". Avoid over-labeling; only patterns with cross-task recurrence risk qualify.

**Injection format** (prepended to Worker spawn prompt when `universal_violations` is non-empty):

```
Universal violations from prior tasks in this session — do NOT repeat these:
- [Security] Raw SQL string concatenation used — always use parameterized queries
- [AI Residuals] localhost hardcoded as connection target — use environment variable
```

**Lead behavior**:
- Initialise `universal_violations = []` before Phase B begins
- After each B-3 review, extract findings with `scope: universal` and append to the list
- Before spawning each Worker in B-2, prepend the accumulated list as a preamble if non-empty
- The list resets when the breezing session ends (not persisted across sessions)
