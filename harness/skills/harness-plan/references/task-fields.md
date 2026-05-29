# Task Fields and Markers

Semantic definitions for fields and markers used in `.claude/harness/plans.json`.

---

## Field Definitions

### DoD (Definition of Done)

Must be a verifiable yes/no condition. Banned phrases: "looks good", "works properly", "is done", "is complete".

### Depends

Use `-` for no dependency, `N.1` for a single task, `N.1, N.2` for multiple, `Phase N` for a full phase dependency.

---

## Status Markers

| Marker | `--status` value | Meaning |
|--------|-----------------|---------|
| `cc:TODO` | `cc:TODO` | Not started |
| `cc:WIP` | `cc:WIP` | In progress |
| `cc:done [hash]` | `cc:done` | Worker completed — include a short git hash via `--hash` when a commit can be attributed |
| `pm:requested` | `pm:requested` | Requested by PM |
| `pm:confirmed` | `pm:confirmed` | PM review confirmed |
| `blocked` | `blocked` | Blocked — **always supply a one-line reason via `--reason`** |

---

## Quality Markers

Quality markers appear inline in the task `description` field at creation time.

| Marker | Meaning |
|--------|---------|
| `[needs-spike]` | High Impact × High Risk — spike task required before implementation |
| `[skip:tdd]` | Skip TDD phase (docs, config, style, trivial changes) |
| `[feature:security]` | Auth, login, payment tasks — security review required |
| `[feature:a11y]` | UI/screen/component tasks — accessibility review required |
| `[bugfix:reproduce-first]` | Bug tasks — must reproduce before fixing |
| `[ralph]` | Iterative loop task — executed by `harness-ralph-loop` (see below) |

---

## `[ralph]` Task Format

`[ralph]` tasks carry two extra fields in the `ralph` sub-object in plans.json:

- **`verify`** (required) — bash command that must exit 0 for the loop to succeed. Auto-inferred from project type if omitted at creation.
- **`maxIter`** (optional) — max worker iterations. Default: 10.

See [ralph-tasks.md](${CLAUDE_SKILL_DIR}/references/ralph-tasks.md) for full format and serialization rules.

---

## Task Status SSOT

`.claude/harness/plans.json` is the sole source of truth for task status. The native Claude Code task list is not mirrored.
