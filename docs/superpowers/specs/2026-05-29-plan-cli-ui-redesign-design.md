# harness plan-cli Web UI Redesign

**Date**: 2026-05-29  
**Status**: Approved  
**Scope**: `go/cmd/harness/plan_serve.go`, `go/cmd/harness/plan_web/src/App.svelte`, `harness/skills/harness-plan/`

---

## Summary

Complete redesign of the `harness plan-cli serve` web UI. The current UI has an unusable layout, broken comments, and no board-style task visibility. This spec replaces it with a minimal Kanban board, a Neo4j-style dependency map, and a working comment system — all served from the same embedded Go binary on port 8888.

---

## 1. Page Layout

A single-page app with a slim top bar and a full-height content area.

```
┌──────────────────────────────────────────────────────────────┐
│ harness plan-cli         [Phase: All phases ▾]  [Board | Map]│
├──────────────────────────────────────────────────────────────┤
│                     (view content)                           │
└──────────────────────────────────────────────────────────────┘
```

**Top bar elements:**
- Left: project title (from `plans.json` project name)
- Center: phase selector dropdown
- Right: view toggle (Board / Map)

**Phase selector:**
- Default: "All phases"
- Options: each phase listed by title (e.g., "Phase 108 — JSON Plans System")
- Filtering applies to both Board and Map views

**Views:** Board (Kanban) and Map (Dependency DAG). No separate phases tab.

---

## 2. Board View (Kanban)

Four fixed columns:

| Column | Status mapping |
|--------|---------------|
| **TODO** | `cc:TODO` |
| **In Progress** | `cc:WIP` + `blocked` (with badge) |
| **Done** | `cc:done`, `pm:confirmed` |
| **Archive** | Tasks from archived phases (read-only) |

**Column behavior:**
- Archive column only visible when "All phases" is selected; hidden in single-phase view
- Columns scroll independently; board does not scroll horizontally

### Card Design

```
┌─────────────────────────┐
│ #108.3 Task name        │
│ Phase 108 · cc:WIP      │  ← phase label hidden in single-phase view
│ ⚠ blocked: waiting on X │  ← only shown if blocked
└─────────────────────────┘
```

- **Task ID**: short human-readable prefix (`#<phaseNumber>.<taskIndex>`) rendered as muted text — phaseNumber is the numeric portion of the phase title (e.g., `108` from "Phase 108"); taskIndex is the 1-based position in the phase task array
- **Status badge**: color-coded pill — gray (TODO), blue (WIP), green (done/confirmed), red (blocked)
- **Blocked warning**: amber warning line shown only when `status === "blocked"`
- Click anywhere on card → opens modal

### Modal (Task Detail / Edit / Comment)

Three tabs:

**Details tab:**
- Task name, description, DoD
- Depends-on list (linked — click jumps to that task's modal)
- Quality markers, urgency/importance indicators
- Read-only

**Edit tab:**
- Status dropdown (cc:TODO / cc:WIP / cc:done / pm:confirmed / blocked)
- Blocked reason field (shown only when status = blocked)
- Urgency and importance selectors
- Depends-on multi-select
- Save button → PATCH `/api/tasks/:id`

**Comments tab:**
- Threaded list, newest-first, with author + ISO timestamp
- Compose box at bottom: textarea + Submit button
- On submit: POST `/api/comments/:targetID` with `{"text": "...", "author": "user"}`

---

## 3. Dependency Map View

Neo4j Browser-style interactive graph.

**Library:** Cytoscape.js (handles force layout, pan/zoom, click events natively)

### Default state

Phase nodes rendered as large rounded-rectangle nodes, labeled by phase title. A directed edge between two phase nodes indicates that at least one task in the source phase depends on a task in the target phase.

### Expand interaction

Click a phase node → the node fans out in place into individual task nodes arranged around it. Dependency edges between tasks within the phase appear. Cross-phase dependency edges run from a task node back to the unexpanded phase node (or to the specific task node if that phase is also expanded).

### Visual encoding

| Element | Style |
|---------|-------|
| Phase node | Large rounded-rectangle, dark border |
| Task node (TODO) | Circle, gray fill |
| Task node (WIP) | Circle, blue fill |
| Task node (done) | Circle, green fill |
| Task node (blocked) | Circle, red fill + amber ring |
| Dependency edge | Directed arrow, gray |
| Cross-phase edge | Directed arrow, dashed |

**Task node click:** opens the same Detail / Edit / Comments modal as the board cards.

**Controls:** pan (drag), zoom (scroll), reset (double-click canvas background to re-fit).

---

## 4. Comments Fix

**Root cause:** the frontend POSTs to `/api/comments/:targetID` without `Content-Type: application/json`, causing Go's `json.NewDecoder` to fail silently or return a 500.

**Fix — frontend (`App.svelte`):**
```javascript
await fetch(`/api/comments/${taskId}`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ text: commentText, author: 'user' })
});
```

**Fix — backend (`plan_serve.go`, `apiPostComment`):**
- Return a structured JSON error (not a blank 500) on malformed body
- Default `Author` to `"user"` if the field is empty in the decoded payload
- Append the comment to the correct phase or task `Comments` slice and atomic-save

---

## 5. Port Change

Change default port from `8080` to `8888` in `runPlanServe()` in `plan_serve.go`.

```go
// before
port := flag.Int("port", 8080, "port to listen on")

// after
port := flag.Int("port", 8888, "port to listen on")
```

---

## 6. CLI Reference (harness-plan skill)

New file: `harness/skills/harness-plan/references/cli-reference.md`

Documents every `harness plan-cli` subcommand: flags, argument shapes, output format, exit codes, and agent-usage examples. Covers:

- `list` — list active phases/tasks (JSON output with `--json`)
- `get <id>` — get a single phase or task by ID
- `add-phase` — `--title`, `--goal`, `--urgency`, `--importance`
- `add-task <phaseID>` — `--name`, `--dod`, `--description`, `--depends`
- `update <taskID>` — `--status`, `--urgency`, `--importance`, `--reason`, `--hash`
- `archive <phaseID>`
- `comment <targetID>` — `--text`, `--author`
- `migrate` — convert Plans.md → plans.json
- `serve` — `--port` (default 8888), `--open`

The `harness-plan/SKILL.md` references table gains a row pointing to this file.

---

## 7. Architecture

No new files beyond the CLI reference doc. All changes are in:

| File | Change |
|------|--------|
| `go/cmd/harness/plan_serve.go` | Port default 8080→8888; comments endpoint defensive decode |
| `go/cmd/harness/plan_web/src/App.svelte` | Complete rewrite: Kanban board, Map view, modal, comments fix |
| `harness/skills/harness-plan/references/cli-reference.md` | New file — CLI reference |
| `harness/skills/harness-plan/SKILL.md` | Add cli-reference.md row to references table |

The Svelte build (`npm run build` in `plan_web/`) produces `dist/` which is embedded via `go:embed` in the binary. Build pipeline is unchanged.

**Cytoscape.js** added as a dev + prod dependency in `plan_web/package.json`.

---

## 8. Out of Scope

- Drag-and-drop between columns (deferred — modal Edit tab covers status changes)
- Authentication / multi-user (CLI is local-only)
- Real-time updates / WebSocket (polling or manual refresh is sufficient)
- Dark mode
