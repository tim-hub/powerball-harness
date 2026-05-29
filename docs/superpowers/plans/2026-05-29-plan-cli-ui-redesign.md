# plan-cli Web UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers-extended-cc:subagent-driven-development (recommended) or superpowers-extended-cc:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the `harness plan-cli serve` web UI with a 4-column Kanban board, Neo4j-style dependency map, working comments, port 8888, and a CLI reference doc.

**Architecture:** Complete rewrite of `App.svelte` (Board + Map views, modal popup), minor Go changes (port default, comment hardening), Cytoscape.js added as a dependency, and a new CLI reference markdown file for the harness-plan skill.

**Tech Stack:** Svelte 5 (runes), TailwindCSS v4, Cytoscape.js, Go 1.21+

**Spec:** `docs/superpowers/specs/2026-05-29-plan-cli-ui-redesign-design.md`

---

### Task 1: Backend — port 8888 + comment endpoint hardening

**Goal:** Change default serve port from 8080 to 8888 and make `apiPostComment` return a structured JSON error on decode failure.

**Files:**
- Modify: `go/cmd/harness/plan_serve.go`

**Acceptance Criteria:**
- [ ] `runPlanServe` defaults to port 8888
- [ ] `apiPostComment` returns `{"error":"..."}` JSON on malformed body (not a plain-text 500)
- [ ] `go build ./go/cmd/harness/...` exits 0
- [ ] Existing plan tests pass: `go test ./go/cmd/harness/...`

**Verify:** `go build -o /tmp/h ./go/cmd/harness/... && grep 8888 go/cmd/harness/plan_serve.go` → line with `port = 8888`

**Steps:**

- [ ] **Step 1: Change the default port**

In `go/cmd/harness/plan_serve.go`, line 28, change:
```go
port := 8080
```
to:
```go
port := 8888
```

- [ ] **Step 2: Harden apiPostComment JSON error response**

Find `apiPostComment` in `plan_serve.go`. The current code calls `http.Error(w, "invalid JSON", ...)` which returns plain text. Replace the error path with a structured JSON response so the frontend can parse it. Locate:
```go
if err := decodeBody(r, &req); err != nil {
    http.Error(w, "invalid JSON", http.StatusBadRequest)
    return
}
```
Replace with:
```go
if err := decodeBody(r, &req); err != nil {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON: " + err.Error()})
    return
}
```

- [ ] **Step 3: Run build + tests**

```bash
go build -o /tmp/harness-serve-test ./go/cmd/harness/...
go test ./go/cmd/harness/... -run TestPlan
```

Expected: build exits 0, tests pass.

- [ ] **Step 4: Commit**

```bash
git add go/cmd/harness/plan_serve.go
git commit -m "fix: change plan-cli serve default port to 8888, harden comment JSON error"
```

---

### Task 2: Add Cytoscape.js dependency

**Goal:** Add `cytoscape` as a production dependency in `plan_web/package.json` and verify the build still works.

**Files:**
- Modify: `go/cmd/harness/plan_web/package.json`

**Acceptance Criteria:**
- [ ] `cytoscape` appears in `dependencies` in `package.json`
- [ ] `npm run build` in `plan_web/` exits 0
- [ ] `plan_web/dist/` is regenerated

**Verify:** `cd go/cmd/harness/plan_web && npm run build 2>&1 | tail -5` → exits 0 with no errors

**Steps:**

- [ ] **Step 1: Install cytoscape**

```bash
cd go/cmd/harness/plan_web
npm install cytoscape
```

Expected: `package.json` gains `"cytoscape": "^3.x.x"` in `dependencies`.

- [ ] **Step 2: Verify build**

```bash
npm run build
```

Expected: exits 0. `dist/` directory is updated.

- [ ] **Step 3: Commit**

```bash
git add go/cmd/harness/plan_web/package.json go/cmd/harness/plan_web/package-lock.json
git commit -m "feat: add cytoscape.js dependency for dependency map view"
```

---

### Task 3: Rewrite App.svelte — Board view + modal

**Goal:** Complete rewrite of `App.svelte` with the new 4-column Kanban board, phase selector, and a 3-tab modal popup (Details / Edit / Comments). No Map view yet — that's Task 4.

**Files:**
- Modify: `go/cmd/harness/plan_web/src/App.svelte`

**Acceptance Criteria:**
- [ ] Header shows: project name left, phase selector center, view toggle (Board/Map) right
- [ ] Board has 4 columns: TODO / In Progress / Done / Archive
- [ ] Archive column only visible when phase selector = "All phases"
- [ ] In Progress column shows both `cc:WIP` and `blocked` tasks; blocked tasks show amber warning
- [ ] Done column shows `cc:done`, `pm:confirmed`, `pm:requested` tasks
- [ ] Cards show `#id name` format with status badge and blocked warning
- [ ] Phase label on card hidden when a specific phase is selected
- [ ] Click any card → centered modal overlay opens
- [ ] Modal has Details / Edit / Comments tabs
- [ ] Details tab: name, description, DoD, depends, quality markers (read-only)
- [ ] Edit tab: status dropdown, blocked reason field (only when status=blocked), urgency/importance, Save button
- [ ] Comments tab: list newest-first, compose box, Cmd+Enter or Post button submits
- [ ] `npm run build` exits 0

**Verify:** `cd go/cmd/harness/plan_web && npm run build 2>&1 | tail -3` → exits 0

**Steps:**

- [ ] **Step 1: Replace App.svelte with the complete new implementation**

Overwrite `go/cmd/harness/plan_web/src/App.svelte` with:

```svelte
<script>
  import { onMount } from 'svelte';

  // ── State ─────────────────────────────────────────────────────────────────
  let phases = $state([]);
  let view = $state('board');        // 'board' | 'map'
  let phaseFilter = $state('all');   // 'all' | String(phase.id)
  let modalTask = $state(null);
  let modalTab = $state('details');  // 'details' | 'edit' | 'comments'
  let commentText = $state('');
  let editStatus = $state('');
  let editBlockedReason = $state('');
  let editUrgency = $state('');
  let editImportance = $state('');
  let loading = $state(true);
  let error = $state('');
  let mapContainer;
  let cy = null;

  const STATUSES = ['cc:TODO', 'cc:WIP', 'cc:done', 'pm:confirmed', 'pm:requested', 'blocked'];

  // ── Derived ───────────────────────────────────────────────────────────────
  const activePhases = $derived(phases.filter(ph => ph.status !== 'archived'));
  const archivedPhases = $derived(phases.filter(ph => ph.status === 'archived'));

  const filteredActivePhases = $derived(
    phaseFilter === 'all'
      ? activePhases
      : activePhases.filter(ph => String(ph.id) === phaseFilter)
  );

  const allActiveTasks = $derived(
    filteredActivePhases.flatMap(ph =>
      (ph.tasks || []).map(t => ({ ...t, _phase: ph, _archived: false }))
    )
  );

  const allArchivedTasks = $derived(
    archivedPhases.flatMap(ph =>
      (ph.tasks || []).map(t => ({ ...t, _phase: ph, _archived: true }))
    )
  );

  const todoTasks    = $derived(allActiveTasks.filter(t => t.status === 'cc:TODO'));
  const wipTasks     = $derived(allActiveTasks.filter(t => t.status === 'cc:WIP' || t.status === 'blocked'));
  const doneTasks    = $derived(allActiveTasks.filter(t => t.status === 'cc:done' || t.status === 'pm:confirmed' || t.status === 'pm:requested'));
  const archiveTasks = $derived(allArchivedTasks);
  const showArchiveColumn = $derived(phaseFilter === 'all');
  const showPhaseLabel    = $derived(phaseFilter === 'all');

  // ── Helpers ───────────────────────────────────────────────────────────────
  function statusColor(s) {
    const m = {
      'cc:TODO':      'bg-gray-100 text-gray-600',
      'cc:WIP':       'bg-blue-100 text-blue-700',
      'cc:done':      'bg-green-100 text-green-700',
      'pm:confirmed': 'bg-green-100 text-green-800',
      'pm:requested': 'bg-yellow-100 text-yellow-700',
      'blocked':      'bg-red-100 text-red-700',
    };
    return m[s] || 'bg-gray-100 text-gray-600';
  }

  function allTasks() {
    return [...allActiveTasks, ...allArchivedTasks];
  }

  // ── API ───────────────────────────────────────────────────────────────────
  async function fetchPhases() {
    loading = true; error = '';
    try {
      const res = await fetch('/api/phases');
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      phases = await res.json();
    } catch (e) {
      error = e.message;
    } finally {
      loading = false;
    }
  }

  async function patchTask(id, fields) {
    const res = await fetch(`/api/tasks/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(fields),
    });
    if (!res.ok) { error = `Save failed: HTTP ${res.status}`; return; }
    await fetchPhases();
    if (modalTask && modalTask.id === id) {
      modalTask = allTasks().find(t => t.id === id) || null;
    }
  }

  async function postComment(targetId) {
    const text = commentText.trim();
    if (!text) return;
    const res = await fetch(`/api/comments/${targetId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text, author: 'user' }),
    });
    if (!res.ok) { error = `Comment failed: HTTP ${res.status}`; return; }
    commentText = '';
    await fetchPhases();
    if (modalTask && modalTask.id === targetId) {
      modalTask = allTasks().find(t => t.id === targetId) || null;
    }
  }

  // ── Modal ─────────────────────────────────────────────────────────────────
  function openTask(task) {
    modalTask = task;
    modalTab = 'details';
    commentText = '';
    editStatus = task.status;
    editBlockedReason = task.blockedReason || '';
    editUrgency = task.urgency || 'medium';
    editImportance = task.importance || 'medium';
  }

  function closeModal() { modalTask = null; }

  async function saveEdit() {
    const fields = { status: editStatus, urgency: editUrgency, importance: editImportance };
    if (editStatus === 'blocked') fields.blockedReason = editBlockedReason;
    await patchTask(modalTask.id, fields);
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────
  onMount(async () => {
    await fetchPhases();
    window.addEventListener('keydown', e => { if (e.key === 'Escape') closeModal(); });
  });
</script>

<div class="min-h-screen bg-gray-50 text-gray-900 flex flex-col">

  <!-- Header -->
  <header class="border-b border-gray-200 bg-white px-4 flex items-center gap-3 sticky top-0 z-30 h-[57px]">
    <span class="text-base font-semibold text-gray-800 mr-2">harness plan-cli</span>

    <select bind:value={phaseFilter}
            class="border border-gray-200 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400">
      <option value="all">All phases</option>
      {#each activePhases as ph}
        <option value={String(ph.id)}>Phase {ph.id} — {ph.title}</option>
      {/each}
    </select>

    <div class="flex-1"></div>

    <div class="flex rounded border border-gray-200 overflow-hidden text-sm">
      <button class="px-3 py-1 {view==='board' ? 'bg-gray-900 text-white' : 'bg-white hover:bg-gray-50'}"
              onclick={() => view = 'board'}>Board</button>
      <button class="px-3 py-1 {view==='map' ? 'bg-gray-900 text-white' : 'bg-white hover:bg-gray-50'}"
              onclick={() => view = 'map'}>Map</button>
    </div>

    <button onclick={fetchPhases}
            class="text-xs text-gray-400 hover:text-gray-700 border border-gray-200 rounded px-2 py-1">↻</button>
  </header>

  {#if error}
    <div class="bg-red-50 text-red-700 text-sm px-4 py-2 border-b border-red-200 flex items-center justify-between">
      <span>{error}</span>
      <button onclick={() => error = ''} class="text-red-400 hover:text-red-700">✕</button>
    </div>
  {/if}

  {#if loading}
    <div class="flex-1 flex items-center justify-center text-gray-400 text-sm">Loading…</div>

  {:else if view === 'board'}
    <!-- Kanban board -->
    <div class="flex gap-4 p-4 overflow-hidden" style="height: calc(100vh - 57px);">

      <!-- TODO -->
      <div class="flex flex-col flex-1 min-w-[200px] bg-white rounded-lg border border-gray-200 overflow-hidden">
        <div class="px-3 py-2 border-b border-gray-200 flex items-center justify-between">
          <span class="text-sm font-semibold text-gray-700">TODO</span>
          <span class="text-xs text-gray-400">{todoTasks.length}</span>
        </div>
        <div class="flex-1 overflow-y-auto p-2 flex flex-col gap-2">
          {#each todoTasks as task (task.id)}
            <button class="text-left bg-white border border-gray-100 rounded-lg p-3 shadow-sm hover:shadow-md transition-shadow w-full"
                    onclick={() => openTask(task)}>
              <div class="flex items-baseline gap-1.5 mb-0.5">
                <span class="text-[10px] text-gray-400 font-mono shrink-0">#{task.id}</span>
                <span class="text-sm font-medium text-gray-900 line-clamp-2 leading-snug">{task.name}</span>
              </div>
              {#if showPhaseLabel}
                <span class="text-xs text-gray-400">Phase {task._phase.id} · {task.status}</span>
              {:else}
                <span class="text-xs px-1.5 py-0.5 rounded-full {statusColor(task.status)}">{task.status}</span>
              {/if}
            </button>
          {/each}
        </div>
      </div>

      <!-- In Progress -->
      <div class="flex flex-col flex-1 min-w-[200px] bg-white rounded-lg border border-blue-200 overflow-hidden">
        <div class="px-3 py-2 border-b border-blue-200 flex items-center justify-between bg-blue-50">
          <span class="text-sm font-semibold text-blue-700">In Progress</span>
          <span class="text-xs text-blue-400">{wipTasks.length}</span>
        </div>
        <div class="flex-1 overflow-y-auto p-2 flex flex-col gap-2">
          {#each wipTasks as task (task.id)}
            <button class="text-left bg-white border border-gray-100 rounded-lg p-3 shadow-sm hover:shadow-md transition-shadow w-full"
                    onclick={() => openTask(task)}>
              <div class="flex items-baseline gap-1.5 mb-0.5">
                <span class="text-[10px] text-gray-400 font-mono shrink-0">#{task.id}</span>
                <span class="text-sm font-medium text-gray-900 line-clamp-2 leading-snug">{task.name}</span>
              </div>
              {#if task.status === 'blocked'}
                <span class="text-xs text-amber-600">⚠ {task.blockedReason || 'blocked'}</span>
              {:else if showPhaseLabel}
                <span class="text-xs text-gray-400">Phase {task._phase.id} · {task.status}</span>
              {:else}
                <span class="text-xs px-1.5 py-0.5 rounded-full {statusColor(task.status)}">{task.status}</span>
              {/if}
            </button>
          {/each}
        </div>
      </div>

      <!-- Done -->
      <div class="flex flex-col flex-1 min-w-[200px] bg-white rounded-lg border border-green-200 overflow-hidden">
        <div class="px-3 py-2 border-b border-green-200 flex items-center justify-between bg-green-50">
          <span class="text-sm font-semibold text-green-700">Done</span>
          <span class="text-xs text-green-400">{doneTasks.length}</span>
        </div>
        <div class="flex-1 overflow-y-auto p-2 flex flex-col gap-2">
          {#each doneTasks as task (task.id)}
            <button class="text-left bg-white border border-gray-100 rounded-lg p-3 shadow-sm hover:shadow-md transition-shadow w-full"
                    onclick={() => openTask(task)}>
              <div class="flex items-baseline gap-1.5 mb-0.5">
                <span class="text-[10px] text-gray-400 font-mono shrink-0">#{task.id}</span>
                <span class="text-sm font-medium text-gray-900 line-clamp-2 leading-snug">{task.name}</span>
              </div>
              {#if showPhaseLabel}
                <span class="text-xs text-gray-400">Phase {task._phase.id}</span>
              {:else}
                <span class="text-xs px-1.5 py-0.5 rounded-full {statusColor(task.status)}">{task.status}</span>
              {/if}
            </button>
          {/each}
        </div>
      </div>

      <!-- Archive (all-phases view only) -->
      {#if showArchiveColumn}
        <div class="flex flex-col flex-1 min-w-[200px] bg-gray-50 rounded-lg border border-gray-200 overflow-hidden opacity-75">
          <div class="px-3 py-2 border-b border-gray-200 flex items-center justify-between">
            <span class="text-sm font-semibold text-gray-400">Archive</span>
            <span class="text-xs text-gray-400">{archiveTasks.length}</span>
          </div>
          <div class="flex-1 overflow-y-auto p-2 flex flex-col gap-2">
            {#each archiveTasks as task (task.id)}
              <button class="text-left bg-gray-50 border border-gray-200 rounded-lg p-3 hover:bg-white transition-colors w-full"
                      onclick={() => openTask(task)}>
                <div class="flex items-baseline gap-1.5 mb-0.5">
                  <span class="text-[10px] text-gray-400 font-mono shrink-0">#{task.id}</span>
                  <span class="text-sm font-medium text-gray-400 line-clamp-2 leading-snug">{task.name}</span>
                </div>
                <span class="text-xs text-gray-400">Phase {task._phase.id}</span>
              </button>
            {/each}
          </div>
        </div>
      {/if}

    </div>

  {:else}
    <!-- Map view container — wired in Task 4 -->
    <div bind:this={mapContainer} class="flex-1 bg-white" style="height: calc(100vh - 57px);"></div>
  {/if}

  <!-- Modal overlay -->
  {#if modalTask}
    <div class="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4"
         onclick={closeModal}
         role="dialog" aria-modal="true">
      <div class="bg-white rounded-xl shadow-2xl w-[600px] max-h-[80vh] flex flex-col overflow-hidden"
           onclick={(e) => e.stopPropagation()}>

        <!-- Modal header -->
        <div class="flex items-center gap-2 px-5 py-4 border-b border-gray-100">
          <span class="text-[11px] font-mono text-gray-400 shrink-0">#{modalTask.id}</span>
          <h2 class="text-base font-semibold text-gray-900 flex-1 leading-snug">{modalTask.name}</h2>
          <button onclick={closeModal} class="text-gray-400 hover:text-gray-700 text-lg leading-none ml-2">✕</button>
        </div>

        <!-- Tabs -->
        <div class="flex border-b border-gray-100 px-5">
          {#each ['details', 'edit', 'comments'] as tab}
            <button class="px-3 py-2 text-sm capitalize mr-1
                           {modalTab === tab
                             ? 'border-b-2 border-gray-900 font-medium text-gray-900'
                             : 'text-gray-400 hover:text-gray-600'}"
                    onclick={() => modalTab = tab}>{tab}</button>
          {/each}
          {#if (modalTask.comments || []).length > 0 && modalTab !== 'comments'}
            <span class="self-center text-xs text-gray-400 ml-1">{(modalTask.comments || []).length}</span>
          {/if}
        </div>

        <!-- Tab body -->
        <div class="flex-1 overflow-y-auto px-5 py-4">

          {#if modalTab === 'details'}
            <div class="flex flex-col gap-4">
              {#if modalTask._phase}
                <p class="text-xs text-gray-400">Phase {modalTask._phase.id} — {modalTask._phase.title}</p>
              {/if}
              <div class="flex flex-wrap gap-2">
                <span class="text-xs px-2 py-0.5 rounded-full {statusColor(modalTask.status)}">{modalTask.status}</span>
                <span class="text-xs text-gray-400">urgency: {modalTask.urgency}</span>
                <span class="text-xs text-gray-400">importance: {modalTask.importance}</span>
              </div>
              {#if modalTask.description}
                <div>
                  <p class="text-xs font-semibold text-gray-500 mb-1">Description</p>
                  <p class="text-sm text-gray-700 leading-relaxed whitespace-pre-wrap">{modalTask.description}</p>
                </div>
              {/if}
              {#if modalTask.dod}
                <div class="bg-green-50 border border-green-200 rounded-lg p-3">
                  <p class="text-xs font-semibold text-green-700 mb-1">Definition of Done</p>
                  <p class="text-sm text-green-800 leading-relaxed">{modalTask.dod}</p>
                </div>
              {/if}
              {#if (modalTask.depends || []).length}
                <div>
                  <p class="text-xs font-semibold text-gray-500 mb-1">Depends on</p>
                  <div class="flex flex-wrap gap-1">
                    {#each modalTask.depends as d}
                      <button class="text-xs bg-gray-100 rounded px-2 py-0.5 hover:bg-gray-200 font-mono"
                              onclick={() => { const t = allTasks().find(x => x.id === d); if (t) openTask(t); }}>{d}</button>
                    {/each}
                  </div>
                </div>
              {/if}
              {#if (modalTask.qualityMarkers || []).length}
                <div>
                  <p class="text-xs font-semibold text-gray-500 mb-1">Quality markers</p>
                  <div class="flex flex-wrap gap-1">
                    {#each modalTask.qualityMarkers as m}
                      <span class="text-xs bg-indigo-50 text-indigo-600 rounded px-2 py-0.5">{m}</span>
                    {/each}
                  </div>
                </div>
              {/if}
            </div>

          {:else if modalTab === 'edit'}
            <div class="flex flex-col gap-4">
              <div>
                <label class="text-xs font-semibold text-gray-500 block mb-1">Status</label>
                <select bind:value={editStatus}
                        class="border border-gray-200 rounded px-2 py-1.5 text-sm w-full focus:outline-none focus:ring-1 focus:ring-blue-400">
                  {#each STATUSES as s}<option value={s}>{s}</option>{/each}
                </select>
              </div>
              {#if editStatus === 'blocked'}
                <div>
                  <label class="text-xs font-semibold text-gray-500 block mb-1">Blocked reason</label>
                  <input bind:value={editBlockedReason}
                         class="border border-gray-200 rounded px-2 py-1.5 text-sm w-full focus:outline-none focus:ring-1 focus:ring-blue-400"
                         placeholder="What is blocking this task?" />
                </div>
              {/if}
              <div class="flex gap-4">
                <div class="flex-1">
                  <label class="text-xs font-semibold text-gray-500 block mb-1">Urgency</label>
                  <select bind:value={editUrgency}
                          class="border border-gray-200 rounded px-2 py-1.5 text-sm w-full focus:outline-none focus:ring-1 focus:ring-blue-400">
                    <option value="high">High</option>
                    <option value="medium">Medium</option>
                    <option value="low">Low</option>
                  </select>
                </div>
                <div class="flex-1">
                  <label class="text-xs font-semibold text-gray-500 block mb-1">Importance</label>
                  <select bind:value={editImportance}
                          class="border border-gray-200 rounded px-2 py-1.5 text-sm w-full focus:outline-none focus:ring-1 focus:ring-blue-400">
                    <option value="high">High</option>
                    <option value="medium">Medium</option>
                    <option value="low">Low</option>
                  </select>
                </div>
              </div>
              <button onclick={saveEdit}
                      class="bg-gray-900 text-white text-sm rounded-lg px-4 py-2 hover:bg-gray-700 self-start">Save</button>
            </div>

          {:else}
            <!-- Comments tab -->
            <div class="flex flex-col gap-3">
              <div class="flex flex-col gap-2">
                {#each [...(modalTask.comments || [])].reverse() as c (c.id)}
                  <div class="bg-gray-50 rounded-lg p-3">
                    <div class="flex items-center gap-2 mb-1">
                      <span class="text-xs font-semibold text-gray-700">{c.authorName || c.author}</span>
                      <span class="text-xs text-gray-400">{c.at?.slice(0, 10)}</span>
                    </div>
                    <p class="text-sm text-gray-700 whitespace-pre-wrap">{c.text}</p>
                  </div>
                {/each}
                {#if !(modalTask.comments || []).length}
                  <p class="text-sm text-gray-400 py-2">No comments yet.</p>
                {/if}
              </div>
              <div class="border-t border-gray-100 pt-3 flex gap-2 items-end">
                <textarea bind:value={commentText}
                          class="flex-1 border border-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400 resize-none"
                          rows="3"
                          placeholder="Add a comment… (⌘Enter to post)"
                          onkeydown={e => { if (e.key === 'Enter' && e.metaKey && commentText.trim()) postComment(modalTask.id); }}></textarea>
                <button onclick={() => postComment(modalTask.id)}
                        class="bg-gray-900 text-white text-sm rounded-lg px-3 py-2 hover:bg-gray-700 shrink-0">Post</button>
              </div>
            </div>
          {/if}

        </div>
      </div>
    </div>
  {/if}

</div>
```

- [ ] **Step 2: Build and verify**

```bash
cd go/cmd/harness/plan_web
npm run build 2>&1
```

Expected: exits 0, `dist/` updated. Fix any Svelte compile errors before proceeding.

- [ ] **Step 3: Commit**

```bash
git add go/cmd/harness/plan_web/src/App.svelte go/cmd/harness/plan_web/dist
git commit -m "feat: rewrite plan-cli web UI — 4-col Kanban, modal, comments fix"
```

---

### Task 4: Add Dependency Map view (Cytoscape.js)

**Goal:** Wire up the Map view in `App.svelte` using Cytoscape.js. Phase nodes by default; click a phase to expand into task nodes. Task click opens the modal.

**Files:**
- Modify: `go/cmd/harness/plan_web/src/App.svelte`

**Acceptance Criteria:**
- [ ] Map view renders phase nodes as large rounded-rectangle nodes
- [ ] Clicking a phase node expands it into task nodes with dependency edges
- [ ] Task node color reflects status (gray/blue/green/red)
- [ ] Clicking a task node opens the modal
- [ ] Double-clicking background resets the layout (fit to view)
- [ ] Switching from Map back to Board destroys the Cytoscape instance cleanly

**Verify:** `cd go/cmd/harness/plan_web && npm run build 2>&1 | tail -3` → exits 0

**Steps:**

- [ ] **Step 1: Add Cytoscape import and map state to script section**

At the top of the `<script>` block (after the existing `import { onMount } from 'svelte';`), add:

```javascript
import cytoscape from 'cytoscape';
```

Add to the state declarations section (after `let mapContainer;`):

```javascript
let expandedPhases = new Set();
```

- [ ] **Step 2: Add initMap and destroyMap functions**

Add these functions to the script section (after the `saveEdit` function):

```javascript
function buildMapElements() {
  const elements = [];
  const targetPhases = phaseFilter === 'all' ? phases : phases.filter(ph => String(ph.id) === phaseFilter);

  for (const ph of targetPhases) {
    if (!expandedPhases.has(ph.id)) {
      elements.push({
        data: { id: `ph-${ph.id}`, label: `Phase ${ph.id}\n${ph.title}`, type: 'phase', phaseId: ph.id }
      });
    } else {
      for (const t of (ph.tasks || [])) {
        const color = { 'cc:TODO': '#9ca3af', 'cc:WIP': '#3b82f6', 'cc:done': '#22c55e',
                        'pm:confirmed': '#16a34a', 'pm:requested': '#eab308', 'blocked': '#ef4444' }[t.status] || '#9ca3af';
        elements.push({
          data: { id: `task-${t.id}`, label: `#${t.id}\n${t.name.slice(0, 30)}`, type: 'task', taskId: t.id, color, phaseId: ph.id }
        });
      }
    }
  }

  for (const ph of targetPhases) {
    if (expandedPhases.has(ph.id)) {
      for (const t of (ph.tasks || [])) {
        for (const dep of (t.depends || [])) {
          const depTask = allTasks().find(x => x.id === dep);
          if (!depTask) continue;
          const depPhaseExpanded = expandedPhases.has(depTask._phase?.id);
          const srcId = `task-${t.id}`;
          const tgtId = depPhaseExpanded ? `task-${dep}` : `ph-${depTask._phase?.id}`;
          if (elements.find(e => e.data.id === tgtId)) {
            elements.push({ data: { id: `edge-${t.id}-${dep}`, source: srcId, target: tgtId, dashed: !depPhaseExpanded } });
          }
        }
      }
    }
  }

  return elements;
}

function initMap() {
  if (!mapContainer || cy) return;
  cy = cytoscape({
    container: mapContainer,
    elements: buildMapElements(),
    layout: { name: 'cose', animate: false, padding: 40 },
    style: [
      { selector: 'node[type="phase"]', style: {
        'shape': 'round-rectangle', 'width': 160, 'height': 60,
        'background-color': '#1f2937', 'color': '#fff',
        'label': 'data(label)', 'text-valign': 'center', 'text-halign': 'center',
        'font-size': '11px', 'text-wrap': 'wrap', 'text-max-width': '140px',
        'border-width': 2, 'border-color': '#374151', 'cursor': 'pointer',
      }},
      { selector: 'node[type="task"]', style: {
        'shape': 'ellipse', 'width': 80, 'height': 80,
        'background-color': 'data(color)', 'color': '#fff',
        'label': 'data(label)', 'text-valign': 'center', 'text-halign': 'center',
        'font-size': '9px', 'text-wrap': 'wrap', 'text-max-width': '70px',
        'cursor': 'pointer',
      }},
      { selector: 'edge', style: {
        'curve-style': 'bezier', 'target-arrow-shape': 'triangle',
        'arrow-scale': 1.2, 'line-color': '#9ca3af', 'target-arrow-color': '#9ca3af',
        'line-style': 'data(dashed)',
      }},
    ],
  });

  cy.on('tap', 'node[type="phase"]', evt => {
    const phaseId = evt.target.data('phaseId');
    expandedPhases = new Set([...expandedPhases, phaseId]);
    cy.destroy(); cy = null;
    initMap();
  });

  cy.on('tap', 'node[type="task"]', evt => {
    const taskId = evt.target.data('taskId');
    const t = allTasks().find(x => x.id === taskId);
    if (t) openTask(t);
  });

  cy.on('dbltap', evt => {
    if (evt.target === cy) cy.fit(undefined, 40);
  });
}

function destroyMap() {
  if (cy) { cy.destroy(); cy = null; }
  expandedPhases = new Set();
}
```

- [ ] **Step 3: Wire initMap/destroyMap to view changes**

Add a `$effect` at the bottom of the script section (before the closing `</script>`):

```javascript
$effect(() => {
  if (view === 'map') {
    // Use setTimeout to ensure mapContainer DOM element is mounted
    setTimeout(initMap, 0);
  } else {
    destroyMap();
  }
});
```

- [ ] **Step 4: Handle phase filter changes in map view**

Add another `$effect` to rebuild the map when `phaseFilter` changes while in map view:

```javascript
$effect(() => {
  const _filter = phaseFilter; // track dependency
  if (view === 'map' && cy) {
    destroyMap();
    setTimeout(initMap, 0);
  }
});
```

- [ ] **Step 5: Build and verify**

```bash
cd go/cmd/harness/plan_web
npm run build 2>&1
```

Expected: exits 0. Fix any compile errors.

- [ ] **Step 6: Commit**

```bash
git add go/cmd/harness/plan_web/src/App.svelte go/cmd/harness/plan_web/dist
git commit -m "feat: add Cytoscape.js dependency map view with Neo4j-style expand"
```

---

### Task 5: CLI reference doc + harness-plan SKILL.md link

**Goal:** Create `harness/skills/harness-plan/references/cli-reference.md` documenting every `harness plan-cli` subcommand, and add a link to it in SKILL.md.

**Files:**
- Create: `harness/skills/harness-plan/references/cli-reference.md`
- Modify: `harness/skills/harness-plan/SKILL.md`

**Acceptance Criteria:**
- [ ] `cli-reference.md` exists with all 9 subcommands documented (list, get, add-phase, add-task, update, archive, comment, migrate, serve)
- [ ] Each subcommand has: synopsis, flags table, exit codes, and 1-2 agent-usage examples
- [ ] `SKILL.md` references table contains a row pointing to `${CLAUDE_SKILL_DIR}/references/cli-reference.md`

**Verify:** `ls harness/skills/harness-plan/references/cli-reference.md && grep cli-reference harness/skills/harness-plan/SKILL.md` → both lines print

**Steps:**

- [ ] **Step 1: Create cli-reference.md**

Create `harness/skills/harness-plan/references/cli-reference.md` with this content:

````markdown
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

**Examples:**
```bash
harness plan-cli serve --open
harness plan-cli serve --port 9000
```
````

- [ ] **Step 2: Add CLI reference row to SKILL.md**

In `harness/skills/harness-plan/SKILL.md`, find the Subcommand Details table that lists references. It looks like:

```markdown
| Subcommand | Reference |
|------------|-----------|
```

Add a new row **above** the `_(quality gate)_` row:

```markdown
| _(CLI reference)_ | [cli-reference.md](${CLAUDE_SKILL_DIR}/references/cli-reference.md) — all subcommands, flags, exit codes, agent examples |
```

- [ ] **Step 3: Verify**

```bash
ls harness/skills/harness-plan/references/cli-reference.md
grep cli-reference harness/skills/harness-plan/SKILL.md
```

Both should print.

- [ ] **Step 4: Commit**

```bash
git add harness/skills/harness-plan/references/cli-reference.md harness/skills/harness-plan/SKILL.md
git commit -m "docs: add harness plan-cli CLI reference for harness-plan skill"
```

---

### Task 6: Full build + end-to-end verification

**Goal:** Verify the complete pipeline: Svelte build → Go embed → binary starts on 8888.

**Files:** None (verification only)

**Acceptance Criteria:**
- [ ] `npm run build` in `plan_web/` exits 0
- [ ] `go build -o /tmp/harness-final ./go/cmd/harness/...` exits 0
- [ ] `/tmp/harness-final plan-cli serve &` starts and the server responds on port 8888
- [ ] `curl -s http://localhost:8888/api/phases` returns a JSON array (not an error)
- [ ] `go test ./go/cmd/harness/... -run TestPlan` passes

**Verify:** `curl -s http://localhost:8888/api/phases | head -1` → `[` or `{` (valid JSON)

**Steps:**

- [ ] **Step 1: Full Svelte build**

```bash
cd go/cmd/harness/plan_web
npm run build 2>&1
```

Expected: exits 0. `dist/assets/` contains hashed JS/CSS bundles.

- [ ] **Step 2: Go build with embedded dist**

```bash
go build -o /tmp/harness-final ./go/cmd/harness/...
echo "Binary size: $(du -sh /tmp/harness-final | cut -f1)"
```

Expected: exits 0. Binary is typically 15-25MB.

- [ ] **Step 3: Smoke test the server**

```bash
/tmp/harness-final plan-cli serve &
SERVER_PID=$!
sleep 1
STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8888/api/phases)
kill $SERVER_PID 2>/dev/null
echo "API status: $STATUS"
```

Expected: `API status: 200`

- [ ] **Step 4: Run plan tests**

```bash
go test ./go/cmd/harness/... -run TestPlan -v 2>&1 | tail -10
```

Expected: all TestPlan* tests PASS.

- [ ] **Step 5: Commit dist if not already tracked**

```bash
git status go/cmd/harness/plan_web/dist/
git add go/cmd/harness/plan_web/dist/
git commit -m "chore: rebuild plan_web dist for v* release" 2>/dev/null || true
```

- [ ] **Step 6: Final status**

```bash
echo "All tasks complete. harness plan-cli serve now runs on port 8888."
```
