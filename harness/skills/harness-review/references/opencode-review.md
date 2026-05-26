# OpenCode Review — harness-review

Load this reference only when **both** conditions are true:
1. `command -v opencode` succeeds (OpenCode CLI is installed)
2. The user explicitly requests OpenCode review (e.g., `--opencode`, "use opencode", "opencode review")

For environment fallbacks when harness-review runs **inside** OpenCode CLI, see
[`opencode-env.md`](${CLAUDE_SKILL_DIR}/../../agents/references/opencode-env.md).

---

## OpenCode Review Invocation

Record the HEAD before implementation starts as `BASE_REF`, then invoke after implementation completes.

```bash
# Record base ref at task start (execute before cc:WIP update in Step 2)
BASE_REF=$(git rev-parse HEAD)

# ... after implementation completion ...

# Execute structured review via official plugin
bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" review --base "${BASE_REF}"
REVIEW_EXIT=$?
```

For adversarial review (challenge design decisions):

```bash
bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" adversarial-review
```

---

## Verdict Mapping (opencode-plugin-cc → Harness)

The official plugin returns structured output conforming to `review-output.schema.json`.

| opencode-plugin-cc | Harness | Verdict Impact |
|---|---|---|
| `approve` | `APPROVE` | - |
| `needs-attention` | `REQUEST_CHANGES` | - |
| `findings[].severity: critical` | `critical_issues[]` | 1 item → REQUEST_CHANGES |
| `findings[].severity: high` | `major_issues[]` | 1 item → REQUEST_CHANGES |
| `findings[].severity: medium/low` | `recommendations[]` | No impact on verdict |

---

## AI Residuals Parallel Scan

AI Residuals scan runs in parallel with the companion review:

```bash
AI_RESIDUALS_JSON="$(bash "${CLAUDE_SKILL_DIR}/../../scripts/review-ai-residuals.sh" --base-ref "${BASE_REF}" 2>/dev/null || echo '{"tool":"review-ai-residuals","scan_mode":"diff","base_ref":null,"files_scanned":[],"summary":{"verdict":"APPROVE","major":0,"minor":0,"recommendation":0,"total":0},"observations":[]}')"
```

The final verdict combines both the companion review result and the AI Residuals scan.

---

## OpenCode CLI Environment Fallbacks

When harness-review runs **inside** an OpenCode session, Claude Code Task tools are unavailable.
See [`opencode-env.md`](${CLAUDE_SKILL_DIR}/../../agents/references/opencode-env.md) for fallback patterns.
