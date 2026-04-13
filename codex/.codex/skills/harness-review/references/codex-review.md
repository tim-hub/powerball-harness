# Codex Environment — harness-review

Load this reference only when **both** conditions are true:
1. `command -v codex` succeeds (Codex CLI is installed)
2. The user explicitly requests Codex review or duo review (e.g., `--dual`, "use codex", "duo review")

The `--dual` flag is handled separately via [`dual-review.md`](${CLAUDE_SKILL_DIR}/references/dual-review.md).
This file covers the Codex environment fallback behavior when harness-review runs **inside** Codex CLI.

---

## Codex CLI Environment Fallbacks

In Codex CLI environments (`CODEX_CLI=1`), some tools are unavailable. Use these fallbacks:

| Normal Environment | Codex Fallback |
|-------------------|----------------|
| Get task list with `TaskList` | Read Plans.md and check WIP/TODO tasks |
| Update status with `TaskUpdate` | Directly update Plans.md markers with `Edit` (e.g., `cc:WIP` → `cc:Done`) |
| Write review result to Task | Output review result to stdout |

### Detection

```bash
if [ "${CODEX_CLI:-}" = "1" ]; then
  # Codex environment: Plans.md-based fallback
fi
```

### Review Output in Codex Environment

Since Task tools are not supported, review results are output to stdout in markdown format.
The Lead agent or user reads the results and decides the next action.
