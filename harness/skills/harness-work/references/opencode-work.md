# OpenCode Work — harness-work

Load this reference only when **both** conditions are true:
1. `command -v opencode` succeeds (OpenCode CLI is installed)
2. The user explicitly passes `--opencode` or asks to use OpenCode for task execution

---

## OpenCode Mode (`--opencode` explicit only)

Delegate tasks to OpenCode via the official plugin `opencode-plugin-cc` companion.

```bash
# Task delegation
bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" task "task content"

# Via stdin (for large prompts)
OPENCODE_PROMPT=$(mktemp /tmp/opencode-prompt-XXXXXX.md)
# Write task content to $OPENCODE_PROMPT, then:
cat "$OPENCODE_PROMPT" | bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" task
rm -f "$OPENCODE_PROMPT"

# Resume previous thread
bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" task --resume-last "continue where we left off"
```

The companion communicates with OpenCode via its HTTP REST API (`opencode serve` headless server),
providing job management, thread resume, and structured output.
Results are verified, and if quality standards are not met, fixes are applied independently.

### Combining with Other Modes

`--opencode` can be combined with other flags:
- `--opencode --breezing` → OpenCode + Breezing (OpenCode handles implementation; Lead/Reviewer structure applies)

In OpenCode + Breezing, implementation is delegated to OpenCode sessions.
OpenCode uses HTTP REST sessions rather than native `spawn_agent`/`send_input` APIs.
See `opencode-orchestration.md` (Phase 2, deferred) for parallel session management details.

---

## OpenCode Exec Review (via official plugin)

When OpenCode is available, the review loop can use OpenCode as the review path.

Record the HEAD at task start as `BASE_REF` and review the diff against that ref.

```bash
# Record base ref at task start (execute before cc:WIP update in Step 2)
BASE_REF=$(git rev-parse HEAD)

# ... after implementation completion ...

# Execute structured review via official plugin
bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" review --base "${BASE_REF}"
REVIEW_EXIT=$?
```

**Verdict Mapping** (opencode-plugin-cc → Harness format):

The official plugin returns structured output conforming to `review-output.schema.json`.

| opencode-plugin-cc | Harness | Verdict Impact |
|---|---|---|
| `approve` | `APPROVE` | - |
| `needs-attention` | `REQUEST_CHANGES` | - |
| `findings[].severity: critical` | `critical_issues[]` | 1 item → REQUEST_CHANGES |
| `findings[].severity: high` | `major_issues[]` | 1 item → REQUEST_CHANGES |
| `findings[].severity: medium/low` | `recommendations[]` | No impact on verdict |

AI Residuals scan runs in parallel with companion review:

```bash
AI_RESIDUALS_JSON="$(bash "${CLAUDE_SKILL_DIR}/../../scripts/review-ai-residuals.sh" --base-ref "${BASE_REF}" 2>/dev/null || echo '{"tool":"review-ai-residuals","scan_mode":"diff","base_ref":null,"files_scanned":[],"summary":{"verdict":"APPROVE","major":0,"minor":0,"recommendation":0,"total":0},"observations":[]}')"
```

The final verdict combines both results.
