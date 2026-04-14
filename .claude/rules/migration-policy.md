# Migration Residue Policy

Policy for operating Harness **exclusion-based verification** (residue checks for deleted concepts).
Defines the operational rules for `deleted-concepts.yaml` + `check-residue.sh` introduced in Phase 40 (v4.1.0).

## Why This Rule Is Needed

Immediately after the v4.0.0 "Hokage" release, the full migration from TypeScript to Go was supposed to be "complete."
However, within 2 days of the release, 13 "relics of the old era" were discovered one after another.
File paths that should have been removed still referenced in test scripts, old version names lingering in documentation,
READMEs still stating Node.js was required — none of these could be caught by individual reviews or
"does X exist?" style checks.

To verify that nothing old remains after a major migration,
a reverse-direction check — exclusion-based verification, asking "does anything we deleted still remain?" — is required.
Following this rule will prevent the same failures from recurring in future major migrations.

## 5 Rules

### Rule 1: Always update deleted-concepts.yaml during major version migrations

The PR that deletes X and the PR that adds X to `deleted-concepts.yaml` must be submitted
simultaneously. Delays are prohibited.

**Why**: If the yaml update is deferred after deletion, another PR may introduce references to X in the meantime,
and those references can be merged without anyone noticing. By bundling the yaml update into the deletion PR,
"deletion = making it a scan target" becomes an indivisible single transaction.

### Rule 2: Update timing is "simultaneous with the deletion PR"

The strong form of Rule 1. Example: if you submit a PR to delete the TypeScript guardrail engine,
add `"TypeScript guardrail engine"` to `deleted_concepts` in the same PR.

"Deleted" and "made a scan target" must always be completed as a set. Either one alone means the job is only half done.

### Rule 3: Operate the allowlist under 3 principles

The `allowlist` field in deleted-concepts.yaml may include the following:

- **Historical records**: CHANGELOG.md and `.claude/memory/archive/` are always allowlisted.
  Recording "this thing existed in the past" is a legitimate mention, not a residue.
- **Migration guides**: Documents like `docs/MIGRATION-*.md` that describe old → new comparisons.
  Mentioning old names in a comparison table is intentional writing.
- **Individual context**: Cases where a reference to an old concept in a specific document is **intentionally legitimate**.
  Example: `.claude/rules/v3-architecture.md` is a historical record of the v3 architecture,
  so it naturally contains `"Harness v3"`.

The allowlist is applied using prefix matching. Keep entry **granularity to a minimum**.
Adding all of `CHANGELOG.md` to the allowlist is legitimate, but
adding the entire `docs/` directory is excessive and renders the scanner meaningless.

### Rule 4: Always perform retroactive validation (verification going back to past commits)

After adding a new deleted-concepts.yaml entry, **go back to past commits, run the scanner,
and confirm that residues are detected as expected**:

```bash
git checkout <past-commit>
bash .claude/scripts/check-residue.sh
# → Expected number of detections (1 or more)
git checkout -
```

This verifies "whether the yaml can actually detect the problem."
If nothing is detected, the allowlist may be too broad, or the pattern may be incorrect.
The goal is to catch false allowlists that accidentally pass early.

### Rule 5: Keep false positives at zero (current HEAD is always 0 detections)

When running the scanner on the current HEAD, **the detection count must always be 0**.
If detections occur, handle them with one of the following:

1. **True residue** — fix immediately (modify the file to remove the old reference)
2. **Should be allowlisted as a historical record, etc.** — update the yaml
3. **Misclassification** (the yaml pattern is matching unintentionally) — remove from yaml

Both CI (section 9 of validate-plugin.sh) and release preflight (Phase 0 of harness-release)
run automatic checks, so **0 detections before merge is guaranteed**.

## Appendix: 13 v3 Residue Cases from This Session (v4.0.0 → v4.0.1)

The cases that motivated Phase 40. **The story of why this feature was born.**

### How They Were Discovered

The v4.0.0 "Hokage" release (2026-04-09) was a full migration from TypeScript implementation to Go native implementation.
The migration itself was completed, but **references from the TypeScript era remained as residues scattered
throughout test scripts, documentation, and SKILL.md files**.
These were discovered accidentally via the following channels:

1. Test execution failures → validate-plugin.sh / check-consistency.sh failing
2. User noticing in the slash palette → "Harness v3" in SKILL.md frontmatter
3. Found during code review → v3 narrative in agents/*.md
4. Found during documentation review → `core/` engine mentions in README.md

The problem is that they were "found by accident." Without a system in place, the same thing will happen in the next release.

### Classification of 13 Cases

| Category | Count | Representative Example |
|---------|------|--------|
| Deleted path references | 2 | `core/src/guardrails/rules.ts` |
| Deleted concept terms | 3 | "TypeScript guardrail engine" |
| SKILL.md version suffixes | 2 | `# Harness Work (v3)` |
| Old runtime requirements | 1 | "Node.js 18+ is installed" |
| Historical tables | 1 | `core/` in README file tree |
| Other (individual formatting bugs) | 4 | README duplicate lines, Japanese/English drift |

### Lessons Learned

All 13 cases were undetectable by **inclusion-based verification** ("does X exist?" style checks).
This is because verifying "X does not remain" requires the prior knowledge that "X was deleted."

The perspective of **exclusion-based verification** ("does deleted X still remain?" — a reverse-direction check)
is required. Phase 40 was born to embed that perspective into the Harness verification layer.

## Related Files

- `.claude/rules/deleted-concepts.yaml` — SSOT catalog of deleted paths/concepts
- `.claude/scripts/check-residue.sh` — Scanner implementation (keep false positives immediately at 0)
- `go/cmd/harness/doctor.go` — `bin/harness doctor --residue` flag
- `tests/validate-plugin.sh` — Section 9: Migration residue check (CI gate)
- `skills/harness-release/SKILL.md` — Phase 0 preflight step 2 (release gate)
