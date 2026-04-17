# trace.v1 — Per-Task Execution Trace Schema

Structured append-only log of what a Worker *actually tried* while executing a Plans.md task. Complements `decisions.md` (why) and `patterns.md` (how) with a third layer: **attempts** (what happened, including failed ones).

Consumers:
- Phase 73 advisor (reads trace for repeated-failure reasoning)
- Phase 74 code-space proposer (reads trace to understand why a skill variant failed)
- Humans inspecting past work

Non-goals:
- Replacement for `.claude/state/session-events.jsonl` (that is session-level; traces are task-level)
- Performance monitoring (no latency histograms, no sampling)
- Audit log for permissions (that's the Go guardrail engine's domain)

---

## Format

UTF-8 **JSONL** (newline-delimited JSON). One event per line. Append-only.

### Storage

| Purpose | Path |
|---------|------|
| Active traces | `.claude/state/traces/<task_id>.jsonl` |
| Archived traces (post 30d of `cc:done`) | `.claude/memory/archive/traces/YYYY-MM/<task_id>.jsonl` |
| Rotated files (when file > 50 MB) | `.claude/state/traces/<task_id>.<N>.jsonl` where `N` = 1, 2, … |

One file per Plans.md task. `<task_id>` matches the task number column exactly (e.g. `72.1`, `72.1.fix`).

### Size caps

| Limit | Value | Behavior |
|-------|-------|----------|
| Soft cap | 10 MB | Writer emits a `decision` event `{rationale: "trace soft-cap reached"}` once per file |
| Hard cap | 50 MB | Writer rotates to next `.N.jsonl` suffix; appends `outcome` event to old file noting rotation |
| Retention | 30 days after task marker flips to `cc:done` or `pm:confirmed` | Maintenance script moves to archive (see 72.4) |

---

## Event Envelope

Every line is a JSON object with these required fields:

| Field | Type | Description |
|-------|------|-------------|
| `schema` | string | Always `"trace.v1"` for files under this schema version |
| `ts` | string | ISO8601 UTC timestamp, e.g. `"2026-04-17T10:30:00Z"` |
| `task_id` | string | Plans.md task identifier, e.g. `"72.1"` |
| `event_type` | enum | One of: `task_start`, `tool_call`, `decision`, `error`, `fix_attempt`, `outcome` |
| `payload` | object | Event-specific, shape defined below |

Optional top-level fields (populated when known):

| Field | Type | Description |
|-------|------|-------------|
| `agent` | string | Agent role that emitted the event: `"worker"`, `"reviewer"`, `"advisor"`, `"lead"`, `"scaffolder"` |
| `attempt_n` | integer | Monotonic attempt counter within this task; starts at 1, increments on each `fix_attempt` |

---

## Event Types

### `task_start`

Emitted once at the start of Worker execution for a task.

```json
{"schema":"trace.v1","ts":"2026-04-17T10:30:00Z","task_id":"72.1","event_type":"task_start","agent":"worker","attempt_n":1,"payload":{"description":"Define trace schema + storage layout","dod":"schema doc exists; includes ≥1 concrete example per event type","phase":72}}
```

Payload:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `description` | string | yes | From Plans.md Description column |
| `dod` | string | yes | From Plans.md DoD column |
| `phase` | integer | yes | Plans.md phase number |

### `tool_call`

Emitted after any tool invocation by an agent. Captures **what was called, not the contents**.

```json
{"schema":"trace.v1","ts":"2026-04-17T10:30:14Z","task_id":"72.1","event_type":"tool_call","agent":"worker","attempt_n":1,"payload":{"tool":"Edit","args_summary":"file_path=.claude/memory/schemas/trace.v1.md (new)","duration_ms":42,"exit_code":0}}
```

Payload:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `tool` | string | yes | Tool name: `Read`, `Edit`, `Write`, `Bash`, `Glob`, `Grep`, agent types, etc. |
| `args_summary` | string | yes | **Short** summary (≤500 chars) — paths + flags, not contents. Bash: first 500 chars of command, truncated with `…` |
| `duration_ms` | integer | no | Wall-clock from invocation to result |
| `exit_code` | integer | no | 0 on success, non-zero on failure (Bash), nullable |
| `error_signature` | string | no | Present only if tool failed; see normalization below |

**Privacy guards — never include in `args_summary` or any trace field**:
- File contents (only paths + sizes)
- Environment variable values (names ok)
- Secrets, tokens, credentials (redact with `[REDACTED]` if they appear in a Bash command)

### `decision`

Emitted when an agent makes a non-obvious choice worth replaying later. Not every line of code gets a decision event — only branches where another reasonable agent might have chosen differently.

```json
{"schema":"trace.v1","ts":"2026-04-17T10:31:02Z","task_id":"72.1","event_type":"decision","agent":"worker","attempt_n":1,"payload":{"rationale":"Flat JSONL chosen over nested-per-attempt; append-simplicity outweighs native attempt grouping since attempt_n field preserves the distinction","alternatives_considered":["nested per-attempt JSONL","single append-log across all tasks"]}}
```

Payload:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `rationale` | string | yes | 1-2 sentences; why this choice |
| `alternatives_considered` | array[string] | no | Short labels, ≤3 items |

### `error`

Emitted when a tool call fails, a test fails, or an unexpected runtime condition occurs. Distinct from `tool_call` with nonzero `exit_code` — use `error` for anything that triggers a retry or blocks progress.

```json
{"schema":"trace.v1","ts":"2026-04-17T10:32:15Z","task_id":"72.2","event_type":"error","agent":"worker","attempt_n":1,"payload":{"error_signature":"go test flock: device or resource busy","raw_error":"go test ./go/internal/trace/...\nFAIL\ntrace_test.go:42: flock: device or resource busy","tool":"Bash"}}
```

Payload:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `error_signature` | string | yes | Normalized — see rule below |
| `raw_error` | string | yes | First 2000 chars of actual error output, truncated with `…` |
| `tool` | string | no | Tool that produced the error |

**`error_signature` normalization** (must match advisor.md line 51):

1. Lowercase
2. Remove all numeric sequences (`:42:` → `::`, `0x7fff...` → ``)
3. Remove run-specific paths (`/tmp/xyz-abc123/` → `<tmp>/`)
4. Remove UUIDs, hashes, commit SHAs
5. Collapse whitespace to single space
6. Trim to first 200 chars

The same logical error must produce the same signature across runs. This enables duplicate suppression in the advisor (`.claude/state/advisor/history.jsonl`).

### `fix_attempt`

Emitted to mark that following events belong to a retry after an `error`. Increments `attempt_n`.

```json
{"schema":"trace.v1","ts":"2026-04-17T10:32:45Z","task_id":"72.2","event_type":"fix_attempt","agent":"worker","attempt_n":2,"payload":{"prior_error_signature":"go test flock: device or resource busy","approach":"Switch to per-file flock instead of single global lock"}}
```

Payload:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `prior_error_signature` | string | yes | Must match the `error_signature` of the triggering error |
| `approach` | string | yes | 1-2 sentence description of the fix strategy |

### `outcome`

Emitted once at task termination (success, fail, or blocked). Closes the trace.

```json
{"schema":"trace.v1","ts":"2026-04-17T10:35:00Z","task_id":"72.1","event_type":"outcome","agent":"worker","attempt_n":1,"payload":{"status":"success","commit":"a1b2c3d","notes":"Schema doc created with 6 event types + concrete examples"}}
```

Payload:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `status` | enum | yes | `"success"`, `"fail"`, or `"blocked"` |
| `commit` | string | no | 7-char commit hash if work was committed |
| `notes` | string | no | ≤500 chars summary |

---

## Invariants

1. **Append-only.** Never rewrite or delete lines. Corrections are new events (typically a `decision` explaining the correction).
2. **Monotonic `ts`.** Timestamps in a single file must be non-decreasing. If the clock jumps backward, writer must coerce to `previous_ts + 1ms`.
3. **Monotonic `attempt_n`.** Starts at 1, increases only on `fix_attempt` events.
4. **Single-task per file.** Cross-task correlation happens at read time, not write time.
5. **UTF-8, LF line endings.** No BOM, no CRLF.
6. **Schema version locked per file.** All lines in a `trace.v1` file have `schema: "trace.v1"`. Future `trace.v2` files are separate.

---

## Concrete Full-Lifecycle Example

Task 72.2 (Go emitter implementation) — one error, one fix, success:

```
{"schema":"trace.v1","ts":"2026-04-17T11:00:00Z","task_id":"72.2","event_type":"task_start","agent":"worker","attempt_n":1,"payload":{"description":"Implement Go emitter go/internal/trace/writer.go","dod":"concurrent test with 10 goroutines produces 10 valid JSONL files","phase":72}}
{"schema":"trace.v1","ts":"2026-04-17T11:00:30Z","task_id":"72.2","event_type":"tool_call","agent":"worker","attempt_n":1,"payload":{"tool":"Write","args_summary":"file_path=go/internal/trace/writer.go (new, 120 lines)","exit_code":0}}
{"schema":"trace.v1","ts":"2026-04-17T11:01:10Z","task_id":"72.2","event_type":"tool_call","agent":"worker","attempt_n":1,"payload":{"tool":"Bash","args_summary":"go test ./go/internal/trace/...","duration_ms":3200,"exit_code":1,"error_signature":"go test flock: device or resource busy"}}
{"schema":"trace.v1","ts":"2026-04-17T11:01:10Z","task_id":"72.2","event_type":"error","agent":"worker","attempt_n":1,"payload":{"error_signature":"go test flock: device or resource busy","raw_error":"...","tool":"Bash"}}
{"schema":"trace.v1","ts":"2026-04-17T11:01:30Z","task_id":"72.2","event_type":"fix_attempt","agent":"worker","attempt_n":2,"payload":{"prior_error_signature":"go test flock: device or resource busy","approach":"Switch from package-global mutex to per-file flock"}}
{"schema":"trace.v1","ts":"2026-04-17T11:02:15Z","task_id":"72.2","event_type":"tool_call","agent":"worker","attempt_n":2,"payload":{"tool":"Edit","args_summary":"file_path=go/internal/trace/writer.go","exit_code":0}}
{"schema":"trace.v1","ts":"2026-04-17T11:02:45Z","task_id":"72.2","event_type":"tool_call","agent":"worker","attempt_n":2,"payload":{"tool":"Bash","args_summary":"go test ./go/internal/trace/...","duration_ms":2800,"exit_code":0}}
{"schema":"trace.v1","ts":"2026-04-17T11:03:00Z","task_id":"72.2","event_type":"outcome","agent":"worker","attempt_n":2,"payload":{"status":"success","commit":"e5f6g7h","notes":"Per-file flock resolved concurrent-write race"}}
```

Reading this lifecycle in order tells a causal story: attempted, failed with a specific error, reasoned about the fix, retried, succeeded. That causal chain is the point of the whole system — it's what a summary cannot preserve.

---

## Design Decisions (for future maintainers)

| Decision | Alternative considered | Why this choice |
|----------|------------------------|-----------------|
| Flat JSONL events | Nested per-attempt records | Simpler append; attempt boundaries preserved via `attempt_n` field; tools like `tail -f`, `grep`, `jq -c` work naturally |
| One file per task | Single append-log across all tasks | Concurrent writers never contend on the same file; archival per-task aligns with Plans.md lifecycle |
| `schema` field on every line | Schema implicit from file name | Future `trace.v2` events can coexist in the same dir if needed; parsers can reject mismatches without filename parsing |
| `error_signature` normalization matches `advisor.md` | Independent normalization | Single source of truth for "same logical error" — enables zero-translation advisor integration |
| Args summary, not full args | Full tool args | Privacy (secrets in Bash args), size (file contents would dominate) |

---

## Related

- `harness/agents/advisor.md` — Consumer; reads traces in Phase 73
- `.claude/rules/migration-policy.md` — If schema ever changes incompatibly, add `trace.v1` to deleted-concepts.yaml
- Plans.md Phase 72 — Implementation tasks that depend on this schema
