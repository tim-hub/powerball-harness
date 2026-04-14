# Codex Work — harness-work

Load this reference only when **both** conditions are true:
1. `command -v codex` succeeds (Codex CLI is installed)
2. The user explicitly passes `--codex` or asks to use Codex for task execution

---

## Codex Mode (`--codex` explicit only)

Delegate tasks to Codex CLI via the official plugin `codex-plugin-cc` companion.

```bash
# Task delegation (writable)
bash scripts/codex-companion.sh task --write "task content"

# Via stdin (for large prompts)
CODEX_PROMPT=$(mktemp /tmp/codex-prompt-XXXXXX.md)
# Write task content
cat "$CODEX_PROMPT" | bash scripts/codex-companion.sh task --write
rm -f "$CODEX_PROMPT"

# Resume previous thread
bash scripts/codex-companion.sh task --resume-last --write "continue where we left off"
```

The companion communicates with Codex via the App Server Protocol,
providing job management, thread resume, and structured output.
Results are verified, and if quality standards are not met, fixes are applied independently.

### Combining with Other Modes

`--codex` can be combined with other flags:
- `--codex --breezing` → Codex + Breezing (Codex handles implementation; Lead/Reviewer structure applies)

In Codex + Breezing, the Breezing mode uses native Codex subagent orchestration:
`spawn_agent`, `wait`, `send_input`, `resume_agent`, `close_agent`
(not the Claude Code Agent/SendMessage API).
See the API mapping table in `team-composition.md` for details.

---

## Codex Exec Review (via official plugin)

When Codex is available, the review loop uses Codex exec as the priority path.

Record the HEAD at task start as `BASE_REF` and review the diff against that ref.

```bash
# Record base ref at task start (execute before cc:WIP update in Step 2)
BASE_REF=$(git rev-parse HEAD)

# ... after implementation completion ...

# Execute structured review via official plugin
bash scripts/codex-companion.sh review --base "${BASE_REF}"
REVIEW_EXIT=$?
```

**Verdict Mapping** (official plugin → Harness format):

The official plugin returns structured output conforming to `review-output.schema.json`.

| Official Plugin | Harness | Verdict Impact |
|---|---|---|
| `approve` | `APPROVE` | - |
| `needs-attention` | `REQUEST_CHANGES` | - |
| `findings[].severity: critical` | `critical_issues[]` | 1 item → REQUEST_CHANGES |
| `findings[].severity: high` | `major_issues[]` | 1 item → REQUEST_CHANGES |
| `findings[].severity: medium/low` | `recommendations[]` | No impact on verdict |

AI Residuals scan runs in parallel with companion review:

```bash
AI_RESIDUALS_JSON="$(bash scripts/review-ai-residuals.sh --base-ref "${BASE_REF}" 2>/dev/null || echo '{"tool":"review-ai-residuals","scan_mode":"diff","base_ref":null,"files_scanned":[],"summary":{"verdict":"APPROVE","major":0,"minor":0,"recommendation":0,"total":0},"observations":[]}')"
```

The final verdict combines both results.
