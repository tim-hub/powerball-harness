# Search Memory

Keyword or regex search across the memory layers. Local-first, MCP-extended when available. Mirrors the same layering contract as `record`: the authoritative reads are local; MCP extends reach across tools.

## Step 1: Local stage (always runs first)

Ripgrep is preferred for speed; grep is the portable fallback. Search both the SSOT files and per-agent memory.

```bash
# Primary: ripgrep
rg -n --ignore-case "<term>" .claude/memory/ .claude/agent-memory/ 2>/dev/null \
  || grep -rn -i "<term>" .claude/memory/ .claude/agent-memory/ 2>/dev/null
```

### Scope

| Path | Contains |
|------|----------|
| `.claude/memory/decisions.md` | SSOT decisions (the primary target) |
| `.claude/memory/patterns.md` | SSOT patterns (the primary target) |
| `.claude/memory/archive/**` | Older SSOT snapshots — still relevant for "when did we decide X" |
| `.claude/agent-memory/**/MEMORY.md` | Per-agent memories (Worker, Reviewer, etc.) |

### Presenting results

- Output as `filename:line:content` for each hit.
- If the query is a regex (contains `|`, `[`, `(`, `.*`, etc.), mention that the local stage treats it as a regex — MCP may not.
- If no hits, say "no local matches found" and suggest one or two synonyms the user could try before failing. Don't loop forever — one synonym attempt is enough.

## Step 2: MCP extension (only if the server is connected)

After the local stage reports back, check whether `mcp__harness__harness_mem_search` is available in the current session.

### Conditions

Run the MCP stage only when **both** are true:

- The local stage completed (even with zero hits — MCP can still add cross-tool context).
- The `mcp__harness__harness_mem_search` tool is present.

### How

```
mcp__harness__harness_mem_search(
  query = "<term>",
  limit = 20              // generous default; trim if the user asked for a quick check
)
```

Related MCP tools to consider chaining when useful:

| Tool | When to chain |
|------|----------------|
| `harness_mem_timeline` | User asked "when did we…?" — the timeline view gives chronological context |
| `harness_mem_get_observations` | You got an observation id from search and need the full record |

### Merging results

Present local and MCP hits under clearly separated headers so the user can tell which layer each hit came from — the layers have different trust properties (local is committed to git; MCP is shared-DB session state).

```
## Local SSOT hits (authoritative, committed)
.claude/memory/patterns.md:42:  ...
.claude/memory/patterns.md:87:  ...

## Cross-tool hits (shared DB — session-scoped, may be stale)
codex/session-2026-04-09 #checkpoint: ...
codex/session-2026-04-12 #event: ...
```

If either layer returns nothing, still include its header with "no matches" so the user sees the coverage explicitly.

## Step 3: Report honestly

Tell the user which layers actually ran and how many hits each returned. Examples:

- `Local: 3 hits in patterns.md. MCP: 2 cross-tool hits.` — both layers ran
- `Local: 0 hits. MCP unavailable — coverage is local-only this session.` — MCP absent
- `Local: 3 hits. MCP search failed: <reason>. Results include local only.` — MCP errored but local still succeeded

Never silently degrade. If MCP didn't run, the user should know — otherwise they might assume cross-tool coverage they didn't get.

## Failure handling

- **Local grep returns nonzero but no hits**: That's a no-match, not an error. Report cleanly.
- **`rg` not installed**: The `||` fallback to `grep` covers it. If both fail, the PATH is broken — report to the user and stop.
- **MCP call errors**: Log the error locally, continue with local-only results. Don't let an MCP hiccup erase the successful local results.

## When to prefer MCP over local

Use your judgement — the user's intent should drive it:

| Question shape | Prefer |
|----------------|--------|
| Literal word or code symbol | **Local** (regex-capable, deterministic) |
| Conceptual question ("when did we decide on Go") | **MCP** (cross-tool, richer context) |
| Cross-tool recall ("what did Codex do last week") | **MCP** (only MCP sees non-Claude sessions) |
| Resume context ("pick up where we left off") | `harness_mem_resume_pack` (not search — see MCP reference) |

## Related

- `skills/memory/SKILL.md` — entry point
- `references/harness-mem-mcp.md` — full MCP tool catalog
- `references/record.md` — writes follow the same local-first / MCP-extension contract
