# Record to SSOT

Append a decision or pattern to the SSOT — **only if the input is genuinely SSOT-worthy**. The validation step is the whole point of this subcommand; skipping it turns SSOT into a dumping ground.

## Step 1: Validate SSOT-worthiness

Before writing anything, classify the input. If it's not a durable decision or a reusable pattern, reject the record and suggest the right destination.

### Decision criteria (goes to `decisions.md`)

A decision is SSOT-worthy when **all** of the following hold:

- Expresses a **Why** — rationale, constraint, or trade-off. Not just a What.
- Affects future choices — if we reversed it, other code or docs would have to change.
- Stable across sessions and contributors — not a one-off task item.

Good examples:
- "We use the Go native binary instead of TypeScript because cold-start latency was breaking hooks."
- "Hooks ship inside the plugin, not in user projects, because Claude Code loads them from `.claude-plugin/` automatically."

### Pattern criteria (goes to `patterns.md`)

A pattern is SSOT-worthy when **all** of the following hold:

- Describes a **reusable recipe** (problem → solution → when to apply → when not to).
- Likely to be applied again in a different context.
- Generalized — not tied to one specific file change.

Good examples:
- "Marker-block idempotent merge for user-owned config files."
- "Optional-tool extraction via conditional reference pointer."

### Rejection cases (NOT SSOT)

If the input matches any of these, reject the record and redirect:

| Input type | Better destination |
|---|---|
| A task to do | `Plans.md` (via `harness-plan add`) |
| A change that was just made | Git commit message / `CHANGELOG.md` |
| A transient session finding | Claude Code auto memory (Layer 1) — no action needed |
| A one-off fix with no broader applicability | Commit message, not SSOT |
| A personal preference with no project-wide impact | User's `~/.claude/CLAUDE.md`, not project SSOT |
| A duplicate of an existing D-entry or P-entry | Point the user at the existing entry; offer to update it |

When rejecting, say so directly and suggest where it belongs:

> This looks like a **task** — better tracked in `Plans.md` than in SSOT. If you believe this is a decision, please add the **Why** (what constraint or trade-off drove it) and the long-term impact.

## Step 2: Determine target and ID

Once validated:

1. Classify: decision → `.claude/memory/decisions.md`, pattern → `.claude/memory/patterns.md`
2. Read the tail of the target file to find the last ID (`D<N>` or `P<N>`)
3. Assign the next ID (e.g. last was `P7` → new is `P8`)
4. Check the index block at the top of the file — the new entry needs a matching index row

## Step 3: Propose the entry

Use the same heading + body format already in the target file. Do **not** invent a new format. Typical structure:

For a decision (`decisions.md`):

```markdown
## D<N>: <short title>

### Context / Problem
<What situation triggered this decision>

### Decision
<What we decided to do>

### Consequences
<What this enables, forbids, or forces>

### Related
- patterns: <if any>
- files: <if any>
```

For a pattern (`patterns.md`):

```markdown
## P<N>: <short title> #tag1 #tag2

### Problem
- <one-line problem statement>

### Solution
- <solution steps>

### When to Apply
- <conditions>

### When NOT to Apply
- <conditions>

### Related
- decisions: <if any>
- files: <if any>
```

## Step 4: Get user approval, then write

Show the proposed entry to the user and ask for approval before writing. Example interaction:

```
User: /memory record We decided to use Opus for high-effort tasks because
                    Haiku couldn't handle the deeper reasoning paths.

Claude:
  Validating SSOT worthiness...
    ✓ Contains a Why ("because Haiku couldn't handle...")
    ✓ Affects future choices (which model to pick per task)
    ✓ Stable across sessions
  → Classifying as: Decision
  → Target: .claude/memory/decisions.md
  → Next ID: D23

  Proposed entry:
  ---
  ## D23: Opus for high-effort tasks

  ### Context / Problem
  ...

  ### Decision
  ...

  ### Consequences
  ...
  ---

  Append to decisions.md and update the index? [yes / edit / no]
```

On approval:

1. Append the entry to the target file
2. Add a matching row to the index block at the top of the file (keep TOC in sync)
3. Stage the change with `git add <file>` — but do **not** auto-commit. The user commits with their own Conventional Commits message per `CONTRIBUTING.md`.

## Step 5: Mirror to MCP shared DB (optional, extension layer)

Local SSOT is authoritative; MCP is an extension that makes the new entry discoverable by other agents (Claude Code, Codex, OpenCode) via the shared DB.

### Conditions

Run this step **only when all of the following are true**:

- The local write in Step 4 succeeded (the entry is in the file and staged).
- The `mcp__harness__harness_mem_record_event` tool is available in the current session (MCP server connected).

If MCP is unavailable, skip this step silently from the tool side but **mention in the Step 6 report** that the mirror did not run. The user should know the cross-tool layer wasn't reached.

### How to mirror

Record a single event that points back to the SSOT entry — don't duplicate the full content, just enough metadata that another agent searching the shared DB can find the committed file.

```
mcp__harness__harness_mem_record_event(
  kind          = "ssot_entry",
  agent         = "<current agent id>",
  payload       = {
    "id":        "<D23 or P8>",
    "type":      "decision" | "pattern",
    "file":      ".claude/memory/decisions.md" | ".claude/memory/patterns.md",
    "title":     "<short title from the heading>",
    "summary":   "<1-2 line summary so search hits make sense>",
    "commit":    "pending"  // user will commit separately
  }
)
```

Failure handling: if the MCP call errors, log the failure locally (print to stdout, don't throw). The SSOT entry is already safe on disk — an MCP hiccup must not cascade into appearing-to-fail from the user's perspective.

## Step 6: Report

After writing, report:

1. **Local**: ID, file, one-line summary.
2. **MCP mirror**: one of
   - `mirrored to shared DB (event id: <...>)` — success
   - `MCP unavailable — entry is local-only; it will sync to shared DB on the next MCP-enabled session` — MCP absent
   - `MCP mirror failed: <reason>` — error, with the local write still successful

If the user wants to commit, chain into `/commit` — but don't do it implicitly.

## Related

- `skills/memory/SKILL.md` — entry point for the `record` subcommand
- `references/ssot-initialization.md` — bootstrap files when they don't exist yet
- `references/sync-ssot-from-memory.md` — promote Layer 1 observations into SSOT
- `.claude/memory/decisions.md` / `.claude/memory/patterns.md` — the target files
