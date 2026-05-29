<script>
  import { onMount } from 'svelte';
  import cytoscape from 'cytoscape';

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
  let expandedPhases = new Set();

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
    try {
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
    } catch (e) {
      error = `Save failed: ${e.message}`;
    }
  }

  async function postComment(targetId) {
    const text = commentText.trim();
    if (!text) return;
    try {
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
    } catch (e) {
      error = `Comment failed: ${e.message}`;
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
    if (!modalTask) return;
    const fields = { status: editStatus, urgency: editUrgency, importance: editImportance };
    if (editStatus === 'blocked') fields.blockedReason = editBlockedReason;
    await patchTask(modalTask.id, fields);
  }

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
              elements.push({ data: { id: `edge-${t.id}-${dep}`, source: srcId, target: tgtId, dashed: !depPhaseExpanded ? 'dashed' : 'solid' } });
            }
          }
        }
      }
    }

    return elements;
  }

  function initMap() {
    if (!mapContainer || cy) return;
    try {
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
    } catch (e) {
      cy = null;
      error = `Map failed to initialize: ${e.message}`;
      return;
    }

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

  // ── Lifecycle ─────────────────────────────────────────────────────────────
  onMount(async () => {
    await fetchPhases();
    const handler = (e) => { if (e.key === 'Escape') closeModal(); };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  });

  $effect(() => {
    const _view = view;
    const _filter = phaseFilter; // track both — rebuild map when either changes
    if (_view === 'map') {
      const timer = setTimeout(initMap, 0);
      return () => { clearTimeout(timer); destroyMap(); };
    } else {
      destroyMap();
    }
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
