# Triage Framework — Upstream Porting Decisions

Reference for the three-bucket decision matrix used in Step 3 of the upstream-clone skill.

---

## The Three Buckets

### ✅ PORTABLE — Can Port

The change is a clean improvement that applies to our architecture without major rework.

**Criteria for PORTABLE**:
- Logic is language/framework-agnostic (e.g., a config key naming convention, a defensive check, a monitoring threshold)
- The improvement is self-contained — it doesn't depend on upstream-only types, schemas, or infrastructure
- Applying it to our codebase requires ≤ ~100 lines of new code, OR it's a pure documentation/config change
- The upstream version may be basic, but the concept is valid and we can implement it correctly for our architecture

**Examples from Phase 77**:
- `filepath.Clean` on config path reads — pure defensive improvement, zero coupling
- Plans.md WIP threshold monitoring — concept is portable; we implement it in Go where upstream did it in Node
- Ring-buffer last-200-lines scan for session.events.jsonl — upstream *deferred* the optimization; we adopt it from day one

**Examples from Phase 75**:
- Worker NG rules (Plans.md markers are Lead-owned; no embedded git; no nested teammates) — behavioral contract additions, zero coupling
- `scripts/enable-1h-cache.sh` — a 10-line shell script, self-contained

---

### ❌ SKIP — Cannot Port

The change is tied to upstream-specific infrastructure that doesn't exist locally, or the architectural assumptions are incompatible.

**Criteria for SKIP**:
- Upstream references files, packages, or schemas that we deleted (e.g., `core/src/guardrails/`, Node.js package.json)
- The change migrates a format we already migrated differently (e.g., upstream doing `v3 → v4`; we did `v4 → v4.9` via a different path)
- The CC/Codex version-specific wiring targets a CC version we've already surpassed with a different approach
- Applying it would require adding a whole new dependency or runtime that we deliberately removed

**Examples**:
- `SKIP — upstream's TypeScript guardrail engine update; we use Go-native bin/harness`
- `SKIP — upstream migrating agents from v3 format; our agents are already in v4 format`
- `SKIP — upstream's `plugins-reference` plugin schema migration; our .claude-plugin/ is already past this`
- `SKIP — upstream's Node.js binary shim; we compile a Go binary directly`

---

### 🔄 HAVE_IT — Already Done (or Done Better)

We have an equivalent implementation, often stronger than upstream's version.

**Criteria for HAVE_IT**:
- A grep search confirms we have a function, config key, or behavior that covers the same ground
- Our version may be more robust (e.g., upstream's was a stub, ours is fully implemented)
- The upstream PR is solving a problem we already solved a different way

**Examples from Phase 77**:
- `HAVE_IT — our UserPromptInjectPolicyHandler already emits additionalContext via consumeResumeContext()`; upstream's Go handler was a stub; wiring the shell script would double-inject
- `HAVE_IT — our harness mem subcommand uses os.Executable() for binary resolution`; upstream's initial version had the projectRoot/bin/harness security issue that CodeRabbit flagged

**When HAVE_IT is close but not identical**:
Sometimes we have the feature but at a lower quality threshold. In this case, consider:
- `HAVE_IT (but weaker)` — create a task to upgrade our version, referencing the upstream improvement as inspiration
- `PORTABLE (upgrade)` — treat it as a new task to strengthen what we have

---

## Decision Flowchart

```
Does the change touch files that don't exist locally?
├─ Yes → SKIP (architecture mismatch)
└─ No → Do we already have this behavior?
    ├─ Yes, fully covered → HAVE_IT
    ├─ Yes, but weaker → PORTABLE (upgrade) or HAVE_IT (weaker)
    └─ No → Is the concept applicable to our architecture?
        ├─ Yes → PORTABLE
        └─ No → SKIP
```

---

## Scoring a PR (Phase-level aggregation)

After triaging all changes in all PRs, count by bucket:
- If PORTABLE > 0: create a harness-plan phase
- If PORTABLE == 0 and HAVE_IT > 0: document the audit in session-log.md; no new phase needed
- If all SKIP: document why; consider a brief note in decisions.md

A phase should not be created for a single trivial PORTABLE item (e.g., one config comment fix).
Bundle trivial items together as a single `chore:` task.

---

## Handling Upstream CodeRabbit / Review Comments

If an upstream PR was amended based on review feedback (e.g., a CodeRabbit security flag), always adopt the *reviewed* version, not the original PR content:
- Read the PR's comment thread (or inspect the diff of the final commit vs the first commit)
- The reviewed version has caught issues the original author missed
- Example: upstream's initial `bin/harness` path used a broad `bin/` allowlist entry; CodeRabbit flagged it as too broad; we adopted the narrower allowlist from the start
