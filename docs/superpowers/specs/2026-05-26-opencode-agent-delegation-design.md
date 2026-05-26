# OpenCode Agent Delegation — Design Spec

**Date:** 2026-05-26  
**Status:** Approved for implementation  
**Scope:** Phase 1 — Foundation (single-session delegation)

---

## Goal

Add OpenCode as a first-class delegate target alongside Codex in the harness plugin. Claude Code remains the primary orchestrator; OpenCode can be used for task implementation and second-opinion review via `--opencode` flags, mirroring the existing `--codex` integration.

**Motivations (all confirmed):**
- Second-opinion diversity across models
- Cost/model flexibility (OpenCode can run local or cheaper models)
- Parallel throughput (per-task backend choice)
- Capability arbitrage (route task classes to best-fit model)

**Out of scope (Phase 2):** Breezing orchestration with OpenCode workers in parallel multi-agent teams.

---

## Prerequisites

- `tasict/opencode-plugin-cc` installed (`plugin marketplace add tasict/opencode-plugin-cc`)
- OpenCode CLI installed (`opencode` on PATH)
- Existing `harness/scripts/codex-companion.sh` unchanged

---

## Architecture

Symmetric proxy pattern — mirrors the existing Codex integration layer exactly.

```
harness/
├── scripts/
│   ├── codex-companion.sh          (existing)
│   └── opencode-companion.sh       NEW — proxy to opencode-plugin-cc companion
├── rules/
│   ├── codex-cli-only.md           (existing)
│   └── opencode-cli-only.md        NEW — invocation policy
└── skills/
    ├── harness-work/
    │   ├── SKILL.md                UPDATE — add --opencode to flags table
    │   └── references/
    │       ├── codex-work.md       (existing)
    │       └── opencode-work.md    NEW — delegation reference
    └── harness-review/
        ├── SKILL.md                UPDATE — add opencode-review.md row
        └── references/
            ├── codex-review.md     (existing)
            └── opencode-review.md  NEW — review reference + verdict mapping

harness/agents/references/
├── codex-env.md                    (existing)
└── opencode-env.md                 NEW — tool fallbacks inside OpenCode
```

**Dependency chain:** `opencode-companion.sh` is the foundation. Both `opencode-work.md` and `opencode-review.md` invoke it via:
```bash
bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" <subcommand> [args]
```

---

## File Designs

### 1. `harness/scripts/opencode-companion.sh`

Mirror of `codex-companion.sh` with three differences:

**Discovery** — searches for `opencode-companion.mjs` from `opencode-plugin-cc`:
```bash
COMPANION=$(find "${PLUGIN_DIRS[@]}" -name "opencode-companion.mjs" \
  \( -path "*/tasict-opencode-plugin-cc/*" \
     -o -path "*/opencode-plugin-cc/*" \
     -o -path "*/plugins/opencode/*" \) \
  2>/dev/null \
  | awk -F/ '{version="0.0.0"; for(i=1;i<=NF;i++){if($i~/^[0-9]+\.[0-9]+(\.[0-9]+)?$/){version=$i}} print version,$0}' \
  | sort -t. -k1,1n -k2,2n -k3,3n \
  | tail -1 \
  | cut -d' ' -f2-)
```
Only searches `~/.claude/plugins` (no `~/.opencode/` equivalent).

**Error message:**
```
ERROR: opencode-plugin-cc not found.
Install: plugin marketplace add tasict/opencode-plugin-cc
Or run: /opencode:setup
```

**Effort fallback env var:** `OPENCODE_EFFORT` (instead of `CODEX_EFFORT`).

Everything else — effort calculation via `calculate-effort.sh`, flag parsing, stdin piping, `exec node "$COMPANION" "$@"` — is identical to `codex-companion.sh`.

Supported subcommands (from `opencode-companion.mjs`):
`task`, `review`, `adversarial-review`, `setup`, `status`, `result`, `cancel`, `task-worker`, `task-resume-candidate`

---

### 2. `harness/rules/opencode-cli-only.md`

Invocation policy mirroring `codex-cli-only.md`:

- **Prohibited:** Direct `opencode run` calls from harness skills and agents
- **Permitted method 1:** `bash scripts/opencode-companion.sh <subcommand>` — from skills and agents
- **Permitted method 2:** `/opencode:*` commands — ad-hoc user interaction (provided by the plugin)
- **No MCP equivalent:** OpenCode uses HTTP REST, not an MCP server; no `deny` rules needed

---

### 3. `harness/skills/harness-work/references/opencode-work.md`

**Load condition** (same gate as `codex-work.md`):
```
Load only when BOTH:
1. opencode CLI is installed (command -v opencode succeeds)
2. User explicitly passes --opencode or asks to use OpenCode for task execution
```

**Invocation pattern:**
```bash
# Task delegation
bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" task "<task content>"

# Via stdin (large prompts)
OPENCODE_PROMPT=$(mktemp /tmp/opencode-prompt-XXXXXX.md)
cat "$OPENCODE_PROMPT" | bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" task
rm -f "$OPENCODE_PROMPT"

# Resume previous thread
bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" task --resume-last "continue"
```

Combines with `--breezing` the same way `--codex` does. OpenCode uses HTTP REST sessions (via `opencode serve` headless server) rather than Codex's App Server Protocol — the companion abstracts this difference.

---

### 4. `harness/skills/harness-review/references/opencode-review.md`

**Load condition:**
```
Load only when BOTH:
1. opencode CLI is installed (command -v opencode succeeds)
2. User explicitly requests OpenCode review (--opencode or "use opencode")
```

**Review invocation:**
```bash
BASE_REF=$(git rev-parse HEAD)   # record before implementation starts

# After implementation:
bash "${CLAUDE_SKILL_DIR}/../../scripts/opencode-companion.sh" review --base "${BASE_REF}"
REVIEW_EXIT=$?
```

**Verdict mapping** (from `opencode-plugin-cc`'s `review-output.schema.json` — identical schema to `codex-plugin-cc`):

| opencode-plugin-cc output | Harness format | Verdict impact |
|---|---|---|
| `approve` | `APPROVE` | — |
| `needs-attention` | `REQUEST_CHANGES` | — |
| `findings[].severity: critical` | `critical_issues[]` | → REQUEST_CHANGES |
| `findings[].severity: high` | `major_issues[]` | → REQUEST_CHANGES |
| `findings[].severity: medium/low` | `recommendations[]` | no impact |

AI Residuals scan runs in parallel (same as codex review path).

---

### 5. `harness/agents/references/opencode-env.md`

Tool fallbacks when `harness-review` runs *inside* OpenCode CLI (where Claude Code Task tools are unavailable):

| Normal Environment | OpenCode Fallback |
|---|---|
| `TaskList` | Read Plans.md, check WIP/TODO markers |
| `TaskUpdate` | Edit Plans.md markers directly |
| Write review result to Task | Output to stdout (markdown) |

**Detection:** `OPENCODE_CLI` follows the same convention as `CODEX_CLI` — it is not auto-set by the plugin; the caller (harness skill or agent) sets it before delegating if needed. Implementation should verify the best detection mechanism (env var vs presence check) and document in the reference file.
```bash
if [ "${OPENCODE_CLI:-}" = "1" ]; then
  # OpenCode environment: Plans.md-based fallback
fi
```

---

### 6. SKILL.md updates

**`harness-work/SKILL.md`** — add `--opencode` row to the flags reference table alongside `--codex`:

| Flag | Effect |
|---|---|
| `--codex` | Delegate task to Codex via codex-plugin-cc |
| `--opencode` | Delegate task to OpenCode via opencode-plugin-cc |

**`harness-review/SKILL.md`** — add `opencode-review.md` reference row alongside `codex-review.md`.

---

## Data Flow: `harness-work --opencode`

```
User: /harness-work --opencode "implement X"
  │
  ├─ harness-work SKILL.md loads opencode-work.md
  ├─ opencode-work.md calls opencode-companion.sh task "implement X"
  │     │
  │     ├─ companion.sh discovers opencode-companion.mjs
  │     ├─ calculate-effort.sh determines effort level
  │     └─ exec node opencode-companion.mjs task --effort <level> "implement X"
  │           │
  │           └─ opencode-companion.mjs starts/connects OpenCode HTTP server
  │              spawns OpenCode session, returns job-id + result
  │
  └─ harness-work runs review loop:
        opencode-companion.sh review --base <BASE_REF>
        maps verdict → APPROVE / REQUEST_CHANGES
        proceeds per standard review loop
```

---

## Error Handling

| Failure | Behavior |
|---|---|
| `opencode-plugin-cc` not installed | `opencode-companion.sh` exits 1 with install instructions |
| `opencode` CLI not on PATH | Load condition check fails; skill falls back to Claude-native path |
| OpenCode server fails to start | `opencode-companion.mjs` exits non-zero; harness-work surfaces error to user |
| Review returns `needs-attention` | Standard REQUEST_CHANGES flow (same as codex path) |

---

## Testing

- `tests/validate-plugin.sh` — existing validation covers path conventions and skill format; no new test sections needed
- Manual smoke test: `bash harness/scripts/opencode-companion.sh setup --json`
- Manual smoke test: `harness-work --opencode` on a trivial task
- `check-consistency.sh` — verify no broken symlinks or missing references

---

## Phase 2 (deferred)

Breezing orchestration with OpenCode workers: `breezing --opencode-workers N`. OpenCode uses HTTP REST sessions (`opencode serve`) rather than Codex's native `spawn_agent`/`wait`/`send_input` API. A separate `opencode-orchestration.md` will document the parallel session management pattern when this is scoped.
