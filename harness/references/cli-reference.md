# harness plan-cli Reference

Machine-readable CLI reference for `harness plan-cli`. All subcommands exit 0 on success, non-zero on error.

## Global flags

| Flag | Description |
|------|-------------|
| `--help` | Print usage and exit |

---

## `list`

List active phases and their tasks.

```
harness plan-cli list [--json] [--phase <phaseID>] [--status <status>]
```

| Flag | Description | Default |
|------|-------------|---------|
| `--json` | Output JSON array instead of human-readable text | false |
| `--phase <id>` | Filter to a single phase | all |
| `--status <status>` | Filter by task status | all |

**Exit codes:** 0 success · 1 plans.json not found or unreadable

**Agent examples:**
```bash
# List all active tasks as JSON
harness plan-cli list --json

# List tasks in phase 108
harness plan-cli list --phase 108 --json
```

---

## `get <id>`

Get a single phase or task by ID.

```
harness plan-cli get <id>
```

`<id>` is a phase number (e.g. `108`) or task ID (e.g. `108.3`).

**Exit codes:** 0 success · 1 not found · 2 plans.json unreadable

**Agent examples:**
```bash
harness plan-cli get 108
harness plan-cli get 108.3
```

---

## `add-phase`

Create a new phase at the top of the phase list.

```
harness plan-cli add-phase --title <title> --goal <goal> [--urgency <u>] [--importance <i>]
```

| Flag | Required | Default |
|------|----------|---------|
| `--title <title>` | Yes | — |
| `--goal <goal>` | Yes | — |
| `--urgency high\|medium\|low` | No | medium |
| `--importance high\|medium\|low` | No | medium |

**Exit codes:** 0 success · 1 missing required flag · 2 save error

**Agent examples:**
```bash
harness plan-cli add-phase --title "Phase 109 — Auth Redesign" --goal "Replace JWT with session cookies" --urgency high
```

---

## `add-task <phaseID>`

Add a task to an existing phase.

```
harness plan-cli add-task <phaseID> --name <name> --dod <dod> [--description <desc>] [--depends <id,...>]
```

| Flag | Required | Default |
|------|----------|---------|
| `--name <name>` | Yes | — |
| `--dod <dod>` | Yes | — |
| `--description <desc>` | No | "" |
| `--depends <id,...>` | No | [] |
| `--urgency high\|medium\|low` | No | medium |
| `--importance high\|medium\|low` | No | medium |

**Exit codes:** 0 success · 1 phase not found · 2 missing required flag

**Agent examples:**
```bash
harness plan-cli add-task 109 \
  --name "Implement session store" \
  --dod "POST /auth/login returns Set-Cookie header with session ID; go test passes" \
  --description "Use Redis-backed session store" \
  --depends "108.3"
```

---

## `update <taskID>`

Update one or more fields of a task.

```
harness plan-cli update <taskID> [--status <s>] [--urgency <u>] [--importance <i>] [--reason <r>] [--hash <h>]
```

| Flag | Description |
|------|-------------|
| `--status <status>` | New status: cc:TODO / cc:WIP / cc:done / pm:confirmed / pm:requested / blocked |
| `--urgency <u>` | high / medium / low |
| `--importance <i>` | high / medium / low |
| `--reason <reason>` | Blocked reason (required when --status blocked) |
| `--hash <hash>` | Git commit hash for statusHash field |

**Exit codes:** 0 success · 1 task not found · 2 invalid status value

**Agent examples:**
```bash
# Mark task in progress
harness plan-cli update 108.3 --status cc:WIP

# Mark blocked with reason
harness plan-cli update 108.3 --status blocked --reason "waiting on PR #42 to merge"

# Mark done with commit hash
harness plan-cli update 108.3 --status cc:done --hash abc1234
```

---

## `archive <phaseID>`

Mark a phase as archived. Tasks remain readable; they appear in the Archive column of the web UI.

```
harness plan-cli archive <phaseID>
```

No flags — only the positional `<phaseID>` argument.

**Exit codes:** 0 success · 1 phase not found

**Agent examples:**
```bash
harness plan-cli archive 107
```

---

## `comment <targetID>`

Add a comment to a phase or task.

```
harness plan-cli comment <targetID> --text <text> [--author <author>]
```

`<targetID>` is a phase number or task ID. `--author` defaults to `"agent"`.

| Flag | Required |
|------|----------|
| `--text <text>` | Yes |
| `--author <author>` | No (default: "agent") |

**Exit codes:** 0 success · 1 target not found · 2 missing required flag

**Agent examples:**
```bash
harness plan-cli comment 108.3 --text "Verified: go test passes, binary starts on 8888"
harness plan-cli comment 108 --text "Phase 108 complete — all 6 tasks cc:done"
```

---

## `migrate`

Convert `Plans.md` (legacy) to `.claude/harness/plans.json`. Non-destructive — Plans.md is not deleted.

```
harness plan-cli migrate [--dry-run]
```

| Flag | Description |
|------|-------------|
| `--dry-run` | Print what would be written; do not write |

**Exit codes:** 0 success · 1 Plans.md not found · 2 parse error

**Agent examples:**
```bash
# Preview migration without writing
harness plan-cli migrate --dry-run

# Migrate Plans.md to plans.json
harness plan-cli migrate
```

---

## `serve`

Start a local HTTP server serving the Kanban web UI and REST API.

```
harness plan-cli serve [--port <port>] [--open]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--port <port>` | 8888 | TCP port to listen on |
| `--open` | false | Open the browser automatically |

**Exit codes:** 0 on SIGINT · 1 on listen error

**Agent examples:**
```bash
harness plan-cli serve --open
harness plan-cli serve --port 9000
```
