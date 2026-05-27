# harness-releaser Agent Design Spec

**Date**: 2026-05-27  
**Phase**: 106  
**Status**: Spike → **GO** (see recommendation)

---

## Background

The `harness-release` skill runs on Sonnet. Of its 8 execution units (Phases 0–6 plus two gates), 5 are deterministic bash invocations with no drafting requirement:

| Unit | Type |
|------|------|
| Phase 0 — Pre-flight script | Pure bash |
| Phase 1 — Read VERSION | `cat VERSION` |
| Phase 2 — Calculate + write new version | `sync-version.sh bump` |
| Phase 3 — CHANGELOG transform | **Drafting — Sonnet** |
| Phase 4 — Commit + tag | Deterministic git commands |
| Phase 5 — Push | Deterministic git command |
| Phase 6 — GitHub Release notes | **Drafting — Sonnet** |
| Review Gate / Work Commit Gate | **Multi-source judgment — Sonnet** |

The `harness-planner` pattern (Phase 105) showed that Haiku + `[Read, Bash]` tools is the right split for deterministic file operations. This design applies the same pattern to the release bash phases.

---

## Design Question A — Schemas

### Refined: Two-invocation Batched Design

The original plan spec proposed one `operation` per request (`preflight`, `bump-version`, `commit-tag`, `push`). After analysis, a **two-invocation batched design** is superior:

| Invocation | Operations included | When |
|------------|---------------------|------|
| **setup** | preflight + bump-version | Before skill's CHANGELOG/Release notes drafting (Phases 0–2) |
| **finalize** | commit-tag + push | After skill has written CHANGELOG.md (Phases 4–5) |

**Why batched over per-operation**: Each Agent tool call carries initialization overhead. Five individual calls for `cat VERSION`, `sync-version.sh bump`, etc. would spend more on scaffolding than on execution. Two batched invocations amortize the overhead — `setup` runs preflight + version bump in one agent turn; `finalize` runs commit/tag/push in one agent turn.

### `releaser-request.v1` schema

```json
{
  "schema_version": "releaser-request.v1",
  "invocation": "setup | finalize",
  "bump_type": "patch | minor | major  — required for setup",
  "version":   "string — required for finalize (the new version string from setup response)",
  "dry_run":   "boolean — optional, default false"
}
```

### `releaser-response.v1` schema

```json
{
  "schema_version": "releaser-response.v1",
  "invocation": "setup | finalize",
  "status":      "applied | skipped | error",
  "new_version": "string — present on setup; the bumped version string",
  "git_hash":    "string — present on finalize; short hash of the release commit",
  "changes":     ["one-line description per action taken"],
  "error":       "string — present only when status=error"
}
```

### `setup` invocation — operations

```bash
# Step 1: Run preflight
bash "${SKILL_DIR}/scripts/release-preflight.sh"
# exits non-zero on any failure → return status: "error" immediately

# Step 2: Read current version
CURRENT=$(cat VERSION)

# Step 3: Bump version + sync manifests + update CHANGELOG compare links
HARNESS_RELEASE_EXTRA_VERSION_FILES="harness/harness.toml" \
  bash "${SKILL_DIR}/scripts/sync-version.sh" bump "${bump_type}"

# Step 4: Read new version
NEW=$(cat VERSION)
```

Returns: `{ "new_version": "$NEW", "changes": ["preflight OK", "bumped $CURRENT → $NEW"] }`

### `finalize` invocation — operations

```bash
# Step 1: Stage release files
git add VERSION CHANGELOG.md harness/harness.toml

# Step 2: Commit (idempotent — skip if commit already exists)
EXISTING=$(git log --oneline | grep "^[0-9a-f]* chore: release v${version}$" | head -1)
if [ -z "$EXISTING" ]; then
  git commit -m "chore: release v${version}"
fi

# Step 3: Tag (idempotent — skip if tag already exists)
if ! git tag -l "v${version}" | grep -q "v${version}"; then
  git tag -a "v${version}" -m "Release v${version}"
fi

# Step 4: Push
git push origin HEAD --tags
```

Returns: `{ "git_hash": "$(git rev-parse --short HEAD)", "changes": [...] }`

---

## Design Question B — Skill ↔ Agent Threading

The skill owns CHANGELOG.md authoring. The sequence:

```
Skill (Sonnet):
  ├─ Review Gate / Work Commit Gate  (skill-only)
  ├─ Agent("harness-releaser", setup, bump_type)
  │    └─ runs preflight + sync-version.sh bump
  │    └─ returns new_version
  ├─ [Phase 3] Sonnet drafts CHANGELOG section, writes CHANGELOG.md
  ├─ [Phase 6] Sonnet drafts GitHub Release notes (held in memory)
  ├─ Confirmation gate: show user the version + CHANGELOG draft + release notes draft
  ├─ Agent("harness-releaser", finalize, version=new_version)
  │    └─ runs commit-tag + push
  │    └─ returns git_hash
  └─ [Phase 6] Skill creates GitHub Release (uses gh CLI, not agent)
```

Key insight: the `finalize` agent only needs the `version` string. CHANGELOG.md is already written by the skill before the agent is invoked. No content passes through the agent — it just runs git commands.

The GitHub Release creation (`gh release create`) stays in the skill because:
1. It requires the drafted release notes (a string the skill holds in memory)
2. `gh` CLI invocation is a single command the skill can run directly with Bash

---

## Design Question C — `--dry-run` Propagation

**Decision: Skill controls dry-run; agent is never invoked for write operations in dry-run mode.**

| Dry-run step | Who handles | Agent call? |
|--------------|-------------|-------------|
| Preflight check | Skill calls `setup` with `dry_run: true` | Yes (preflight is read-only and useful to show) |
| Version calculation | Skill computes arithmetic locally | No agent call |
| CHANGELOG draft display | Skill generates and shows, does NOT write | No agent call |
| Release notes draft display | Skill generates and shows | No agent call |
| Confirmation gate | Skill shows "dry-run complete" | No agent call |

For `setup` with `dry_run: true`:
- Run preflight (read-only)
- Compute `new_version` arithmetically from `cat VERSION`
- Do NOT run `sync-version.sh bump`
- Return `{ "new_version": "$COMPUTED", "changes": ["preflight OK", "would bump to $COMPUTED (dry-run: VERSION not written)"] }`

`finalize` is never called in dry-run. The skill shows a dry-run summary of what the git commands would do.

---

## Design Question D — Failure Recovery

### `setup` failures

Preflight fails (non-zero exit):
- Agent returns `status: "error"`, `error: "preflight failed: <stderr>"`
- Skill presents the preflight output to user — same behavior as today
- VERSION has NOT been modified (preflight runs before bump)

Version bump fails (e.g., `harness.toml` has unexpected format):
- Agent returns `status: "error"`, `error: "sync-version.sh failed: <stderr>"`
- VERSION may be partially modified — agent should include `"current_version"` in the error response so skill can detect drift
- Skill runs `sync-version.sh check` to diagnose

### `finalize` failures — idempotency is the recovery mechanism

The `finalize` flow is idempotent by design (see Question A). If it fails mid-way and is re-invoked, it skips already-completed steps:

| Failure point | State | Re-invocation result |
|---------------|-------|----------------------|
| `git add` fails | Nothing staged | Re-invocation restarts cleanly |
| `git commit` fails (hook rejected) | Changes staged but no commit | Re-invocation: staging is re-attempted (idempotent), commit retried after fix |
| `git commit` succeeds, `git tag` fails | Commit exists, no tag | Re-invocation: commit step skipped (already exists), tag retried |
| `git tag` succeeds, `git push` fails | Commit + tag local only | Re-invocation: commit + tag skipped, push retried |
| `git push` fails (remote conflict) | All local operations done | Agent returns error with current state; skill surfaces to user |

The idempotency check for commit uses `git log --oneline | grep "chore: release v${version}"`. For the tag: `git tag -l "v${version}"`.

### Error response contract

On ANY non-zero bash exit, the agent MUST:
1. Return `status: "error"` immediately — no swallowing
2. Include the actual stderr in the `error` field
3. Include `changes` listing which steps were completed before failure
4. Never return `status: "applied"` unless ALL steps in the invocation succeeded

---

## Design Question E — Explicit Non-Goals

The `harness-releaser` agent **never**:
- Drafts CHANGELOG content (Before/After tables, user value framing)
- Drafts GitHub Release notes
- Makes Review Gate or Work Commit Gate decisions
- Calls `AskUserQuestion`
- Reads `.claude/state/review-result.json` or `.claude/state/review-approved.json`
- Creates or edits Markdown files (tools exclude `Write` and `Edit`)
- Invokes `gh release create` (skill runs this directly after the finalize response)
- Makes version classification decisions (`patch` vs `minor` vs `major`) — caller supplies `bump_type`

---

## Tool Restriction Rationale

```yaml
tools: [Read, Bash]
disallowedTools: [Write, Edit, Agent]
```

| Restriction | Why |
|-------------|-----|
| No `Write` | Agent cannot author CHANGELOG, release notes, or any content file |
| No `Edit` | Same — all file mutations go through bash scripts (`sync-version.sh`) |
| No `Agent` | No nested spawning; caller is the orchestrator |
| `Read` allowed | Agent needs to read VERSION, git status output |
| `Bash` allowed | Agent's entire value is running bash scripts and git commands |

The `sync-version.sh` script handles all file writes internally (VERSION, harness.toml, CHANGELOG compare links). The agent invokes it via Bash — no direct file authoring needed.

---

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Haiku misreads bash stderr as success | High | Error contract: any non-zero exit = `status: "error"`, full stderr in `error` field |
| `commit-tag` partial failure corrupts state | Medium | Idempotency design; re-invocation recovers cleanly |
| `sync-version.sh bump` updates CHANGELOG compare links unexpectedly | Low | Script is well-tested and idempotent; agent just invokes it |
| Agent overhead exceeds savings | Low | Two-invocation batched design amortizes overhead; savings on each invocation are multiple bash commands |
| Regression blocks a real release | High | 106.4 smoke test with `--dry-run` is mandatory before any real release uses this path; 106.3 wiring must be behind a feature check |

---

## Go / No-Go Recommendation

**GO — with two constraints:**

1. **Hard error contract enforced in agent body**: The agent MUST document that any non-zero bash exit triggers `status: "error"` immediately. This prevents Haiku from treating a failed `git push` as a soft warning and returning `status: "applied"`.

2. **Smoke test (106.4) gates real usage**: The `harness-release` SKILL.md wiring in 106.3 must include a comment that the releaser agent path is available but requires the 106.4 dry-run validation before use in a live release. This is a documentation guard, not a code flag.

**Why go rather than no-go**:
- The two-invocation batched design addresses the per-operation overhead concern
- Tool restriction to `[Read, Bash]` is a hard architectural guarantee — agent cannot corrupt files
- Idempotency design means partial failures are recoverable without manual intervention
- The bash phases being delegated are genuinely trivial; the judgment phases (CHANGELOG, Release notes, gates) are not being moved

**Deferred to Phase 106 v2 (if Phase 106 succeeds)**:
- `release-this` skill adoption of the releaser agent
- `--dry-run` calling `setup` with `dry_run: true` vs. skipping entirely (current design skips — simpler)
