# Patterns (SSOT)

This file is the Single Source of Truth (SSOT) for reusable solutions (patterns).
Record the **problem, solution, and applicability conditions** so the same decision can be quickly replicated next time.

## Index

- P1: Declarative rule table pattern #guardrails #rules
- P2: stdin -> route -> stdout pipeline #hooks #architecture
- P3: Synchronizing test assertions with output language #testing #i18n
- P4: Symlink + CI verification pattern #monorepo #consistency
- P5: Native module management via optionalDependencies #node #dependencies

---

## P1: Declarative rule table pattern #guardrails #rules

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

- Automatically verify symlink health in CI with `scripts/ci/check-consistency.sh`
- The check validates that each symlink in `codex/.codex/skills/` resolves to a valid directory in `skills/`
- OpenCode platform was retired in Phase 36; `sync-v3-skill-mirrors.sh` was removed (symlinks replaced file-copy mirrors)

### When to Apply

- Monorepo structures where a single source is referenced from multiple directories

### Notes

- On Windows (`core.symlinks=false`), symlinks become regular files; `codex/README.md` documents a `setup-codex.sh` workaround

### Related

- decisions: D1
- files: `scripts/ci/check-consistency.sh`

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
