# Patterns (SSOT)

This file is the Single Source of Truth (SSOT) for reusable solutions (patterns).
Record the **problem, solution, and applicability conditions** so the same decision can be quickly replicated next time.

## Index

- P1: Declarative rule table pattern #guardrails #rules _(superseded — see P8 for Go equivalent)_
- P2: stdin -> route -> stdout pipeline #hooks #architecture _(superseded — see P8 for Go equivalent)_
- P3: Synchronizing test assertions with output language #testing #i18n _(superseded — TypeScript era)_
- P4: Symlink + CI verification pattern #monorepo #consistency
- P5: Native module management via optionalDependencies #node #dependencies _(TypeScript era — Go binary has no runtime deps)_
- P6: Idempotent managed-block template merge #templates #setup #idempotency
- P7: Optional-tool extraction for skills #skills #conditional-load #codex
- P8: Go declarative guardrail rule table #guardrails #go
- P9: Concurrent hook fan-out with deny-wins merge #hooks #concurrency

---

## P1: Declarative rule table pattern #guardrails #rules

> _(Superseded by D9/Go migration — TypeScript core removed in v4 Hokage. See `go/internal/guardrail/` for current Go implementation.)_

### Problem

- Guardrail rules written as if-else chains make adding new rules, testing, and priority management difficult

### Solution

- Declare rules as an array of `{id, toolPattern, evaluate(ctx) -> RuleResult | null}`
- Array order = priority. Return the result of the first matching rule
- Returning `null` means "this rule does not apply" -> proceed to the next rule

```typescript
const GUARD_RULES: GuardRule[] = [
  { id: "R01:no-sudo", toolPattern: /^Bash$/, evaluate: (ctx) => { ... } },
  { id: "R02:no-write-protected", toolPattern: /^(?:Write|Edit)$/, evaluate: (ctx) => { ... } },
  // ...
];
```

### When to Apply

- When there are 5 or more rules with complex priority or conditional branching
- When individual rule unit testing is needed

### When NOT to Apply

- When there are only 2-3 simple rules (if-else is sufficient)
- When rules have complex interdependencies (DAG-based evaluation is needed)

### Notes

- Filter by tool type first using toolPattern to ensure performance
- Control rule skipping via context flags such as workMode / codexMode

### Related

- decisions: D2
- files: `core/src/guardrails/rules.ts`

---

## P2: stdin -> route -> stdout pipeline #hooks #architecture

> _(Superseded by D9/Go migration — TypeScript core removed in v4 Hokage. See `go/internal/guardrail/` for current Go implementation.)_

### Problem

- Claude Code hooks receive JSON via stdin and return JSON via stdout
- When logic is scattered across bash scripts, error handling and testing become difficult

### Solution

- Read stdin from a single TypeScript entry point (index.ts)
- Route by `hook_event_name` (PreToolUse -> pre-tool.ts, PostToolUse -> post-tool.ts, etc.)
- Output results as JSON to stdout

```
hooks/pre-tool.sh -> stdin -> node core/dist/index.js -> stdout -> Claude Code
```

### When to Apply

- When implementing complex logic in Claude Code hooks
- When you want to handle multiple hook events in a unified way

### When NOT to Apply

- For simple single-line bash processing (when echo/exit is sufficient)

### Notes

- Keep the bash shim in hooks/ to just `cat | node core/dist/index.js`
- Handle JSON parse errors properly on the TypeScript side

### Related

- decisions: D7
- files: `core/src/index.ts`, `hooks/pre-tool.sh`

---

## P3: Synchronizing test assertions with output language #testing #i18n

> _(Superseded by D9/Go migration — TypeScript core removed in v4 Hokage. See `go/internal/guardrail/` for current Go implementation.)_

### Problem

- When translating output strings in source code, tests break if assertions still use the old language
- Example: After changing `"Warning"` (from Japanese) in `rules.ts`, `toContain("old-Japanese-string")` in `integration.test.ts` fails

### Solution

- When translating output strings, always grep all test files for matching assertions
- Update assertion strings simultaneously to match the output
- Recommended: Set up CI to detect mismatches between source language and test assertions

```bash
# Post-translation verification
grep -rn '[Japanese-characters]' core/src/ | wc -l  # Should be 0
```

### When to Apply

- When translating user-facing output strings (warnings, error messages, etc.)
- When tests verify output strings via partial matching

### When NOT to Apply

- When output is consolidated in constant files or resource bundles (only the translation file needs to be swapped)

### Notes

- Partial match tests like `toContain()` are particularly easy to overlook
- `toEqual()` is safer because translation omissions immediately cause errors

### Related

- decisions: D3

---

## P4: Symlink + CI verification pattern #monorepo #consistency

### Problem

- In a structure where skills/ is the source of truth and codex/ references it via symlinks, broken symlinks are hard to notice

### Solution

- Automatically verify symlink health in CI with `local-scripts/check-consistency.sh`
- The check validates that each symlink in `codex/.codex/skills/` resolves to a valid directory in `skills/`
- OpenCode platform was retired in Phase 36; `sync-v3-skill-mirrors.sh` was removed (symlinks replaced file-copy mirrors)

### When to Apply

- Monorepo structures where a single source is referenced from multiple directories

### Notes

- On Windows (`core.symlinks=false`), symlinks become regular files; `codex/README.md` documents a `setup-codex.sh` workaround

### Related

- decisions: D1
- files: `local-scripts/check-consistency.sh`

---

## P5: Native module management via optionalDependencies #node #dependencies

### Problem

- Native modules like better-sqlite3 fail to build on certain Node versions, blocking the entire installation

### Solution

- Place them in `optionalDependencies` in package.json
- Allow installation to succeed even with `npm install --ignore-scripts`
- Use `createRequire` + try-catch at runtime for graceful degradation

### When to Apply

- When a native module is only needed for a subset of features
- When portability across different Node versions and operating systems is required

### When NOT to Apply

- When the native module is essential for core functionality with no alternatives

### Related

- decisions: D4

---

## P6: Idempotent managed-block template merge #templates #setup #idempotency

### Problem

- A setup/init step needs to add a known block of content to a user-owned file (`.gitignore`, `.editorconfig`, CI config, etc.) without duplicating on re-run, and without overwriting the user's other content

### Solution

- Wrap the managed content in unique sentinel markers: `# >>> harness-managed >>>` ... `# <<< harness-managed <<<`
- Before appending, `grep -qF` for the open marker — if present, skip silently; otherwise append the template (with a leading blank line)
- The close marker is reserved for future "replace-block" tooling (read between markers, swap content) — not used today but cheap to leave in place

### Example

```bash
MARKER="# >>> harness-managed >>>"
if grep -qF "$MARKER" .gitignore 2>/dev/null; then
  echo ".gitignore already contains harness-managed block — skipping"
else
  echo "" >> .gitignore
  cat "${CLAUDE_PLUGIN_ROOT}/templates/gitignore-harness" >> .gitignore
  echo "Appended harness-managed gitignore block"
fi
```

### When to Apply

- Setup steps that need to merge a known content block into a user-owned config file
- Any "managed snippet" that should remain stable across re-runs of the same setup command

### When NOT to Apply

- Files Harness fully owns (just overwrite)
- Content that needs interleaving with user content (use a real config-merge tool instead)

### Notes

- Use `${CLAUDE_PLUGIN_ROOT}/...` for the template source, not a CWD-relative path — `harness-setup init` runs from the user's project, not from the plugin install dir
- Choose marker comment syntax appropriate to the file (`#` for shell-style, `//` for JS/TS-style, `;` for ini-style)

### Related

- files: `templates/gitignore-harness`, `skills/harness-setup/SKILL.md` (init subcommand)

---

## P7: Optional-tool extraction for skills #skills #conditional-load #codex

### Problem

- A skill's main flow defaults to one tool path (e.g., Claude tools), but documents an alternative path that requires an optional external CLI (e.g., Codex). Inlining the alternative path bloats the skill, adds noise for users without the tool, and increases the token budget on every load.

### Solution

- Move the optional-tool sections (flag dispatch, command examples, fallback behavior) into a sibling reference file: `skills/<skill>/references/<tool>-<skill>.md`
- Replace the inline section in the main `SKILL.md` with a 2-line conditional pointer:

  ```markdown
  > Load [`${CLAUDE_SKILL_DIR}/references/<tool>-<skill>.md`](${CLAUDE_SKILL_DIR}/references/<tool>-<skill>.md)
  > only when `command -v <tool>` succeeds **and** the user passes `--<tool>` or explicitly asks to use <tool>.
  ```

- Keep flag names and option-table rows in `SKILL.md` (they're discoverability surface) — only the *detailed mechanics* go to the reference

### Example

Phase 41 applied this twice:

| Skill | Section moved | Reference file | SKILL.md size |
|-------|---------------|----------------|---------------|
| `harness-review` | "Codex Environment" (~22 lines) | `references/codex-review.md` | 236 → 218 |
| `harness-work` | "Codex Mode" + "Codex Exec Review" (~60 lines) | `references/codex-work.md` | 520 → 471 |

### When to Apply

- The optional tool is genuinely opt-in (not used by most users)
- The optional-tool section is >15 lines of inline detail
- A `command -v <tool>` check can reliably gate the load

### When NOT to Apply

- The "optional" tool is actually the default path most users take
- The detail fits in a few lines (just inline it)
- Loading the reference unconditionally is cheap and the section is small

### Notes

- Use `${CLAUDE_SKILL_DIR}` (CC v2.1.69+) for the reference path, not a relative path
- Keep `argument-hint` listing the flag (e.g., `--codex`) — the flag exists; only the docs moved
- Pseudocode function calls referencing the tool (e.g., `codex_exec_review()` inside a review-loop pseudocode block) can stay in the main `SKILL.md` — they're internal references, not user-facing docs

### Related

- files: `skills/harness-review/references/codex-review.md`, `skills/harness-work/references/codex-work.md`

---

## P8: Go declarative guardrail rule table #guardrails #go

> _(Go successor to the TypeScript P1/P2 patterns. Current implementation since v4.0.0.)_

### Problem

- Guardrail rules written as if-else chains make adding new rules, testing, and priority management difficult
- TypeScript core/ required Node.js runtime at hook execution time

### Solution

- Declare rules as a slice of `GuardRule` structs in `go/internal/guardrail/rules.go`
- Each rule has `{id, toolPattern, evaluate(ctx) -> *RuleResult}`; slice order = priority
- Return `nil` means "rule does not apply" → next rule is evaluated
- Context carries `WorkMode`, `BreezingRole`, `CodexMode` flags to enable rule skipping

```go
var guardRules = []GuardRule{
  {id: "R01:no-sudo", toolPattern: regexp.MustCompile(`^Bash$`), evaluate: evalNoSudo},
  {id: "R02:no-write-protected-paths", toolPattern: regexp.MustCompile(`^Write|Edit|MultiEdit$`), evaluate: evalProtectedPaths},
  // ...
}
```

- Binary compiled CGO_ENABLED=0 (static, no runtime deps) for darwin-arm64/amd64/linux-amd64
- hooks/ shims call `"${CLAUDE_PLUGIN_ROOT}/bin/harness" hook <name>` — no Node.js needed

### When to Apply

- Adding a new guardrail rule: add one entry to the slice in `rules.go`, add a test in `rules_test.go`
- Rules with complex priority or context-dependent skipping

### When NOT to Apply

- Simple one-off permission checks that don't need priority ordering

### Notes

- `deny-wins` semantics: if any rule returns Deny, the entire evaluation denies even if a later rule would approve
- `WorkMode` bypasses R04 and R05 (write outside project / rm -rf) — these are safe in breezing workers but risky in interactive mode

### Related

- decisions: D9
- files: `go/internal/guardrail/rules.go`, `go/internal/guardrail/rules_test.go`, `harness/bin/harness`

---

## P9: Concurrent hook fan-out with deny-wins merge #hooks #concurrency

### Problem

- `PostToolUse` triggers many independent side-effect scripts sequentially (9 subprocess forks), adding latency after every tool call
- `PreToolUse` needs to run multiple guardrail checks but must deny if any single check denies

### Solution

- Introduce fan-out orchestrator scripts (`post-tool-batch`, `pre-tool-batch`) that launch all handlers as goroutines and merge results
- **PostToolUse fan-out**: fire-and-forget goroutines — side effects run concurrently, result is always approve (no blocking)
- **PreToolUse fan-out**: collect all results; if any result is `deny`, return deny (deny-wins semantics)

```
PostToolUse → post-tool-batch → [goroutine: memory-bridge] [goroutine: log-toolname] [goroutine: usage-tracker] ...
                                  all concurrent, result = approve
PreToolUse  → pre-tool-batch  → [goroutine: guardrail] [goroutine: browser-guide] ...
                                  deny-wins merge
```

### When to Apply

- Multiple independent side-effect hooks that don't need to block the response
- Multiple guard checks where any-deny should block

### When NOT to Apply

- Hooks with data dependencies between them (must run sequentially)
- Hooks where ordering of side effects matters

### Notes

- Use `flock` for shared file access (Plans.md) within fan-out goroutines to prevent race conditions
- Fan-out reduced PostToolUse from 9 sequential forks to 2 concurrent — measurable latency drop

### Related

- decisions: D15
- files: `go/internal/hookhandler/`, `harness/hooks/hooks.json` (batch entries)

---

## P10: Per-task execution trace as causal-history layer #observability #traces

### Problem

- Summary-only memory (`decisions.md`, `patterns.md`) captures *why* and *how* but loses the causal chain of "attempted X → failed with Y → fixed via Z"
- Downstream consumers (Phase 73 advisor, Phase 74 code-space proposer) need causal reasoning about *why* past attempts failed, not just that they did
- Session-level `.claude/state/agent-trace.jsonl` aggregates across all tasks in a session, making per-task replay expensive (have to filter across everything)

### Solution

- Add a per-task trace layer between raw tool calls and human-curated summaries: JSONL files at `.claude/state/traces/<task_id>.jsonl`
- Schema (`trace.v1`) is flat: one JSONL line per event, six event types (`task_start`, `tool_call`, `decision`, `error`, `fix_attempt`, `outcome`)
- Writer is concurrency-safe (flock + fsync) and caps marshaled line size at 1 MiB as a privacy guard against accidental file-content leaks
- Capture is automatic via PostToolUse hook scoped to the active `cc:WIP` task in Plans.md — no LLM instruction needed
- Retention: 30 days post `cc:Done`, then moved to `.claude/memory/archive/traces/YYYY-MM/` by `/maintenance --archive-traces`

### When to Apply

- Downstream agents need causal reasoning ("why did this fail?" not just "did this fail?")
- Work is structured into discrete tasks with clear start/end markers (Plans.md `cc:WIP` → `cc:Done`)
- The cost of losing attempt history exceeds the cost of writing + archiving JSONL

### When NOT to Apply

- For session-level monitoring — `.claude/state/agent-trace.jsonl` already exists for that
- For metrics/telemetry — this is a replay log, not a metrics system; events aren't numeric
- For audit logging of permissions — that belongs to the Go guardrail engine

### Notes

- `error_signature` normalization (lowercase, strip numerics/hex/UUIDs/paths, collapse whitespace) makes "same logical error" detectable across runs and is shared with `harness/agents/advisor.md`'s duplicate-suppression cache
- Flat schema chosen over nested-per-attempt for append simplicity; attempt boundaries preserved via monotonic `attempt_n` field on each event
- Privacy defaults favour less data over more: `args_summary` preserves paths but never contents; Bash commands truncated to 500 chars; environment variables are never captured

### Related

- schema: `.claude/memory/schemas/trace.v1.md`
- files: `go/internal/trace/writer.go`, `go/internal/trace/errsig.go`, `go/internal/hookhandler/posttooluse_trace.go`, `harness/skills/maintenance/scripts/archive-traces.sh`
- roadmap: Phase 73 (advisor consumes traces), Phase 74 (code-space proposer consumes traces)
