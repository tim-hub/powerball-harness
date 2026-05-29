<script>
  import { onMount } from 'svelte';
  import { Dialog, Tabs } from 'bits-ui';

  // ── State ─────────────────────────────────────────────────────────────────
  let phases = $state([]);
  let phaseFilter = $state('all');     // 'all' | String(phase.id)
  let searchQuery = $state('');
  let modalTask = $state(null);
  let modalTab = $state('details');    // 'details' | 'edit' | 'comments'
  let commentText = $state('');
  let editStatus = $state('');
  let editBlockedReason = $state('');
  let editUrgency = $state('');
  let editImportance = $state('');
  let loading = $state(true);
  let error = $state('');

  const STATUSES = ['cc:TODO', 'cc:WIP', 'cc:done', 'pm:confirmed', 'pm:requested', 'blocked'];
  const DONE_STATUSES = new Set(['cc:done', 'pm:confirmed']);

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

  const q = $derived(searchQuery.toLowerCase().trim());

  function matchesSearch(t) {
    if (!q) return true;
    return String(t.id).includes(q) || t.name.toLowerCase().includes(q);
  }

  // Status of every task by id (across all phases, including archived) so we can
  // resolve whether a task's dependencies are satisfied.
  const statusById = $derived.by(() => {
    const m = new Map();
    for (const ph of phases) for (const t of (ph.tasks || [])) m.set(t.id, t.status);
    return m;
  });

  // A dependency is satisfied when the depended-on task is done/confirmed.
  // Unknown ids (dep points at a task that no longer exists) count as satisfied
  // so a dangling reference never blocks a task forever.
  function depSatisfied(depId) {
    const s = statusById.get(depId);
    return s === undefined || DONE_STATUSES.has(s);
  }

  // "Ready to go": a not-started task whose every dependency is satisfied.
  function isReady(task) {
    return task.status === 'cc:TODO' && (task.depends || []).every(depSatisfied);
  }

  // TODO column: ready-to-go tasks float to the top. Array.sort is stable, so
  // ordering within the ready / not-ready groups is preserved (phase order).
  const todoTasks = $derived.by(() => {
    const list = allActiveTasks.filter(t => t.status === 'cc:TODO' && matchesSearch(t));
    return list.sort((a, b) => (isReady(a) ? 0 : 1) - (isReady(b) ? 0 : 1));
  });
  const wipTasks     = $derived(allActiveTasks.filter(t => (t.status === 'cc:WIP' || t.status === 'blocked') && matchesSearch(t)));
  const doneTasks    = $derived(allActiveTasks.filter(t => (t.status === 'cc:done' || t.status === 'pm:confirmed' || t.status === 'pm:requested') && matchesSearch(t)));
  const archiveTasks = $derived(allArchivedTasks.filter(t => matchesSearch(t)));
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

  // ── Lifecycle ─────────────────────────────────────────────────────────────
  onMount(async () => {
    await fetchPhases();
    const handler = (e) => { if (e.key === 'Escape') closeModal(); };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  });
</script>

<div class="h-screen bg-gray-50 text-gray-900 flex flex-col overflow-hidden">

  <!-- Header -->
  <header class="border-b border-gray-200 bg-white px-4 flex items-center gap-3 sticky top-0 z-30 h-[57px]">
    <span class="text-base font-semibold text-gray-800 mr-2">powerball-harness</span>

    <select bind:value={phaseFilter}
            class="border border-gray-200 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-400">
      <option value="all">All phases</option>
      {#each activePhases as ph}
        <option value={String(ph.id)}>Phase {ph.id} — {ph.title}</option>
      {/each}
    </select>

    <div class="flex-1"></div>

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

  {:else}
    <!-- Search bar -->
    <div class="px-4 py-2 bg-white border-b border-gray-200">
      <input
        bind:value={searchQuery}
        type="search"
        placeholder="Search tasks by name or ID…"
        class="w-full max-w-sm border border-gray-200 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
      />
    </div>

    <!-- Card snippet shared by every column. Shows a "Ready" badge for
         ready-to-go tasks and a "needs" line listing each dependency, colored
         green when satisfied and amber when still pending. -->
    {#snippet taskCard(task)}
      <button class="text-left rounded-lg p-3 w-full transition-shadow
                     {task._archived
                       ? 'bg-gray-50 border border-gray-200 hover:bg-white'
                       : 'bg-white border border-gray-100 shadow-sm hover:shadow-md'}"
              onclick={() => openTask(task)}>
        <div class="flex items-baseline gap-1.5 mb-0.5">
          <span class="text-[10px] text-gray-400 font-mono shrink-0">#{task.id}</span>
          <span class="text-sm font-medium line-clamp-2 leading-snug {task._archived ? 'text-gray-400' : 'text-gray-900'}">{task.name}</span>
        </div>

        <div class="flex flex-wrap items-center gap-1.5 mt-1">
          {#if isReady(task)}
            <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-emerald-100 text-emerald-700 font-medium">● Ready</span>
          {/if}
          {#if task.status === 'blocked'}
            <span class="text-xs text-amber-600">⚠ {task.blockedReason || 'blocked'}</span>
          {:else if showPhaseLabel}
            <span class="text-xs text-gray-400">Phase {task._phase.id} · {task.status}</span>
          {:else}
            <span class="text-xs px-1.5 py-0.5 rounded-full {statusColor(task.status)}">{task.status}</span>
          {/if}
        </div>

        {#if (task.depends || []).length}
          <div class="mt-1.5 flex flex-wrap items-center gap-1">
            <span class="text-[10px] text-gray-400 uppercase tracking-wide mr-0.5">needs</span>
            {#each task.depends as d}
              <span class="text-[10px] font-mono px-1 py-0.5 rounded {depSatisfied(d) ? 'bg-green-50 text-green-600' : 'bg-amber-50 text-amber-700'}">#{d}{depSatisfied(d) ? ' ✓' : ''}</span>
            {/each}
          </div>
        {/if}
      </button>
    {/snippet}

    <!-- Kanban board — fixed-width columns, horizontal scroll -->
    <div class="flex gap-4 p-4 overflow-x-auto flex-1">

      <!-- TODO (ready-to-go tasks float to the top) -->
      <div class="flex flex-col w-72 shrink-0 bg-white rounded-lg border border-gray-200 overflow-hidden">
        <div class="px-3 py-2 border-b border-gray-200 flex items-center justify-between">
          <span class="text-sm font-semibold text-gray-700">TODO</span>
          <span class="text-xs text-gray-400">{todoTasks.length}</span>
        </div>
        <div class="flex-1 overflow-y-auto p-2 flex flex-col gap-2">
          {#each todoTasks as task (task.id)}
            {@render taskCard(task)}
          {/each}
        </div>
      </div>

      <!-- In Progress -->
      <div class="flex flex-col w-72 shrink-0 bg-white rounded-lg border border-blue-200 overflow-hidden">
        <div class="px-3 py-2 border-b border-blue-200 flex items-center justify-between bg-blue-50">
          <span class="text-sm font-semibold text-blue-700">In Progress</span>
          <span class="text-xs text-blue-400">{wipTasks.length}</span>
        </div>
        <div class="flex-1 overflow-y-auto p-2 flex flex-col gap-2">
          {#each wipTasks as task (task.id)}
            {@render taskCard(task)}
          {/each}
        </div>
      </div>

      <!-- Done -->
      <div class="flex flex-col w-72 shrink-0 bg-white rounded-lg border border-green-200 overflow-hidden">
        <div class="px-3 py-2 border-b border-green-200 flex items-center justify-between bg-green-50">
          <span class="text-sm font-semibold text-green-700">Done</span>
          <span class="text-xs text-green-400">{doneTasks.length}</span>
        </div>
        <div class="flex-1 overflow-y-auto p-2 flex flex-col gap-2">
          {#each doneTasks as task (task.id)}
            {@render taskCard(task)}
          {/each}
        </div>
      </div>

      <!-- Archive (all-phases view only) -->
      {#if showArchiveColumn}
        <div class="flex flex-col w-72 shrink-0 bg-gray-50 rounded-lg border border-gray-200 overflow-hidden opacity-75">
          <div class="px-3 py-2 border-b border-gray-200 flex items-center justify-between">
            <span class="text-sm font-semibold text-gray-400">Archive</span>
            <span class="text-xs text-gray-400">{archiveTasks.length}</span>
          </div>
          <div class="flex-1 overflow-y-auto p-2 flex flex-col gap-2">
            {#each archiveTasks as task (task.id)}
              {@render taskCard(task)}
            {/each}
          </div>
        </div>
      {/if}

    </div>
  {/if}

  <!-- Task detail modal — bits-ui Dialog for proper a11y -->
  <Dialog.Root
    open={!!modalTask}
    onOpenChange={(open) => { if (!open) closeModal(); }}
  >
    <Dialog.Portal>
      <Dialog.Overlay class="fixed inset-0 bg-black/40 z-50" />
      <Dialog.Content
        class="fixed left-[50%] top-[50%] -translate-x-1/2 -translate-y-1/2 z-50
               bg-white rounded-xl shadow-2xl w-[600px] max-h-[80vh] flex flex-col overflow-hidden"
      >
        {#if modalTask}
          <!-- Modal header -->
          <div class="flex items-center gap-2 px-5 py-4 border-b border-gray-100">
            <span class="text-[11px] font-mono text-gray-400 shrink-0">#{modalTask.id}</span>
            <Dialog.Title class="text-base font-semibold text-gray-900 flex-1 leading-snug">
              {modalTask.name}
            </Dialog.Title>
            <Dialog.Close
              class="text-gray-400 hover:text-gray-700 text-lg leading-none ml-2"
              aria-label="Close">✕</Dialog.Close>
          </div>

          <!-- Tabs — bits-ui -->
          <Tabs.Root
            value={modalTab}
            onValueChange={(v) => modalTab = v}
            class="flex flex-col flex-1 overflow-hidden"
          >
            <Tabs.List class="flex border-b border-gray-100 px-5 shrink-0">
              {#each ['details', 'edit', 'comments'] as tab}
                <Tabs.Trigger
                  value={tab}
                  class="px-3 py-2 text-sm capitalize mr-1
                         data-[state=active]:border-b-2 data-[state=active]:border-gray-900
                         data-[state=active]:font-medium data-[state=active]:text-gray-900
                         data-[state=inactive]:text-gray-400 data-[state=inactive]:hover:text-gray-600"
                >{tab}</Tabs.Trigger>
              {/each}
              {#if (modalTask.comments || []).length > 0 && modalTab !== 'comments'}
                <span class="self-center text-xs text-gray-400 ml-1">{(modalTask.comments || []).length}</span>
              {/if}
            </Tabs.List>

            <!-- Details -->
            <Tabs.Content value="details" class="flex-1 overflow-y-auto px-5 py-4">
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
            </Tabs.Content>

            <!-- Edit -->
            <Tabs.Content value="edit" class="flex-1 overflow-y-auto px-5 py-4">
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
            </Tabs.Content>

            <!-- Comments -->
            <Tabs.Content value="comments" class="flex-1 overflow-y-auto px-5 py-4">
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
            </Tabs.Content>

          </Tabs.Root>
        {/if}
      </Dialog.Content>
    </Dialog.Portal>
  </Dialog.Root>

</div>
