<script lang="ts">
  import { onMount } from 'svelte';
  import { Popover } from 'bits-ui';
  import { toast } from 'svelte-sonner';
  import { Button } from '../ui';
  import CreateSprintModal from './CreateSprintModal.svelte';
  import EditSprintModal from './EditSprintModal.svelte';
  import CompleteSprintModal from './CompleteSprintModal.svelte';
  import {
    getActiveSprint,
    getFutureSprints,
    getClosedSprints,
    getSprintCards,
    startSprint,
    reopenSprint,
    updateSprint,
    type SprintData,
  } from '../../lib/api/sprints';
  import { getBoard } from '../../lib/api/boards';

  interface Props {
    boardId: string;
    onSprintChange?: () => void;
  }

  let { boardId, onSprintChange }: Props = $props();

  let activeSprint = $state<SprintData | null>(null);
  let futureSprints = $state<SprintData[]>([]);
  let closedSprints = $state<SprintData[]>([]);
  let closedSprintsPageInfo = $state<{ hasNextPage: boolean; endCursor: string | null; totalCount: number } | null>(null);
  let loadingMoreClosed = $state(false);
  let loading = $state(true);
  let open = $state(false);
  let showCreateModal = $state(false);
  let showEditModal = $state(false);
  let showCompleteModal = $state(false);
  let incompleteCardCount = $state(0);
  let editingSprint = $state<SprintData | null>(null);
  let actionLoading = $state(false);
  let editingSprintId = $state<string | null>(null);
  let editingName = $state('');

  onMount(async () => {
    await loadSprints();
  });

  export async function refresh() {
    await loadSprints();
  }

  async function loadSprints() {
    try {
      loading = true;
      const [active, future, closedResult] = await Promise.all([
        getActiveSprint(boardId),
        getFutureSprints(boardId),
        getClosedSprints(boardId, 10),
      ]);
      activeSprint = active;
      futureSprints = future;
      closedSprints = closedResult.sprints;
      closedSprintsPageInfo = closedResult.pageInfo;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load sprints';
      toast.error(message);
    } finally {
      loading = false;
    }
  }

  async function loadMoreClosedSprints() {
    if (!closedSprintsPageInfo?.hasNextPage || loadingMoreClosed) return;
    try {
      loadingMoreClosed = true;
      const result = await getClosedSprints(boardId, 10, closedSprintsPageInfo.endCursor ?? undefined);
      closedSprints = [...closedSprints, ...result.sprints];
      closedSprintsPageInfo = result.pageInfo;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load more sprints';
      toast.error(message);
    } finally {
      loadingMoreClosed = false;
    }
  }

  async function handleStartSprint(sprint: SprintData) {
    try {
      actionLoading = true;
      await startSprint(sprint.id);
      toast.success(`Started ${sprint.name}`);
      open = false;
      await loadSprints();
      onSprintChange?.();
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to start sprint';
      toast.error(message);
    } finally {
      actionLoading = false;
    }
  }

  async function handleCompleteSprint() {
    if (!activeSprint) return;

    try {
      actionLoading = true;

      // Calculate incomplete card count
      const [sprintCards, board] = await Promise.all([
        getSprintCards(activeSprint.id),
        getBoard(boardId),
      ]);

      if (board) {
        // Get done column IDs
        const doneColumnIds = new Set(
          board.columns.filter(col => col.isDone).map(col => col.id)
        );

        // Count cards not in done columns
        incompleteCardCount = sprintCards.filter(card =>
          card.column && !doneColumnIds.has(card.column.id)
        ).length;
      } else {
        incompleteCardCount = 0;
      }

      // Open the complete sprint modal
      open = false;
      showCompleteModal = true;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load sprint data';
      toast.error(message);
    } finally {
      actionLoading = false;
    }
  }

  async function handleSprintCompleted() {
    showCompleteModal = false;
    await loadSprints();
    onSprintChange?.();
  }

  async function handleReopenSprint(sprint: SprintData) {
    try {
      actionLoading = true;
      await reopenSprint(sprint.id);
      toast.success(`Reopened ${sprint.name}`);
      await loadSprints();
      onSprintChange?.();
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to reopen sprint';
      toast.error(message);
    } finally {
      actionLoading = false;
    }
  }

  async function handleSprintCreated() {
    showCreateModal = false;
    await loadSprints();
    onSprintChange?.();
  }

  function startEditingSprint(sprint: SprintData, e: Event) {
    e.stopPropagation();
    editingSprintId = sprint.id;
    editingName = sprint.name;
  }

  function cancelEditing() {
    editingSprintId = null;
    editingName = '';
  }

  async function saveSprintName(sprintId: string) {
    if (!editingName.trim()) {
      cancelEditing();
      return;
    }
    try {
      actionLoading = true;
      await updateSprint(sprintId, { name: editingName.trim() });
      toast.success('Sprint renamed');
      editingSprintId = null;
      editingName = '';
      await loadSprints();
      onSprintChange?.();
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to rename sprint';
      toast.error(message);
    } finally {
      actionLoading = false;
    }
  }

  function handleEditKeydown(e: KeyboardEvent, sprintId: string) {
    if (e.key === 'Enter') {
      e.preventDefault();
      saveSprintName(sprintId);
    } else if (e.key === 'Escape') {
      cancelEditing();
    }
  }

  function openEditModal(sprint: SprintData, e: Event) {
    e.stopPropagation();
    editingSprint = sprint;
    showEditModal = true;
  }

  function handleSprintUpdated(updated: SprintData) {
    // Update the sprint in the appropriate list
    if (activeSprint?.id === updated.id) {
      activeSprint = updated;
    }
    futureSprints = futureSprints.map(s => s.id === updated.id ? updated : s);
    closedSprints = closedSprints.map(s => s.id === updated.id ? updated : s);
    showEditModal = false;
    editingSprint = null;
    onSprintChange?.();
  }

  function formatDateRange(sprint: SprintData): string {
    const format = (dateStr: string | null | undefined): string => {
      if (!dateStr) return '';
      return new Date(dateStr).toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
      });
    };
    const start = format(sprint.startDate);
    const end = format(sprint.endDate);
    if (start && end) return `${start} - ${end}`;
    if (start) return `From ${start}`;
    if (end) return `Until ${end}`;
    return '';
  }
</script>

<Popover.Root bind:open onOpenChange={(isOpen) => { if (isOpen) loadSprints(); }}>
  <Popover.Trigger
    class="inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium rounded-md border transition-colors
      {activeSprint
        ? 'bg-green-50 border-green-200 text-green-800 hover:bg-green-100'
        : 'bg-gray-50 border-gray-200 text-gray-700 hover:bg-gray-100'}"
    disabled={loading}
  >
    {#if loading}
      <div class="h-4 w-24 bg-gray-200 rounded animate-pulse"></div>
    {:else if activeSprint}
      <svg class="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
      </svg>
      <span>{activeSprint.name}</span>
    {:else}
      <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>No active sprint</span>
    {/if}
    <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
    </svg>
  </Popover.Trigger>

  <Popover.Content
    side="bottom"
    align="start"
    sideOffset={8}
    class="z-50 w-72 bg-white border border-gray-200 rounded-lg shadow-lg animate-in fade-in-0 zoom-in-95"
  >
    <div class="p-3 border-b border-gray-100">
      <h3 class="text-sm font-semibold text-gray-900">Sprints</h3>
    </div>

    <div class="max-h-80 overflow-y-auto">
      <!-- Active Sprint -->
      {#if activeSprint}
        {@const sprint = activeSprint}
        <div class="p-3 bg-green-50 border-b border-gray-100">
          <div class="flex items-start justify-between gap-2">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-1.5">
                <span class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                  Active
                </span>
                {#if editingSprintId === sprint.id}
                  <input
                    type="text"
                    bind:value={editingName}
                    onkeydown={(e) => handleEditKeydown(e, sprint.id)}
                    onblur={() => saveSprintName(sprint.id)}
                    class="text-sm font-medium text-gray-900 border border-gray-300 rounded px-1.5 py-0.5 w-full focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    autofocus
                  />
                {:else}
                  <button
                    type="button"
                    onclick={(e) => startEditingSprint(sprint, e)}
                    class="text-sm font-medium text-gray-900 truncate hover:text-indigo-600 text-left"
                    title="Click to rename"
                  >
                    {sprint.name}
                  </button>
                {/if}
                <button
                  type="button"
                  onclick={(e) => openEditModal(sprint, e)}
                  class="p-0.5 text-gray-400 hover:text-gray-600 rounded"
                  title="Edit sprint details"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
              </div>
              {#if sprint.goal}
                <p class="text-xs text-gray-600 mt-1 line-clamp-2">{sprint.goal}</p>
              {/if}
              {#if formatDateRange(sprint)}
                <p class="text-xs text-gray-500 mt-1">{formatDateRange(sprint)}</p>
              {/if}
            </div>
            <Button
              size="sm"
              variant="secondary"
              onclick={handleCompleteSprint}
              disabled={actionLoading}
            >
              Complete
            </Button>
          </div>
        </div>
      {/if}

      <!-- Future Sprints -->
      {#if futureSprints.length > 0}
        <div class="py-1">
          <div class="px-3 py-1.5 text-xs font-medium text-gray-500 uppercase tracking-wide">
            Upcoming
          </div>
          {#each futureSprints as sprint (sprint.id)}
            <div class="px-3 py-2 hover:bg-gray-50">
              <div class="flex items-start justify-between gap-2">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-1.5">
                    {#if editingSprintId === sprint.id}
                      <input
                        type="text"
                        bind:value={editingName}
                        onkeydown={(e) => handleEditKeydown(e, sprint.id)}
                        onblur={() => saveSprintName(sprint.id)}
                        class="text-sm font-medium text-gray-900 border border-gray-300 rounded px-1.5 py-0.5 w-full focus:outline-none focus:ring-1 focus:ring-indigo-500"
                        autofocus
                      />
                    {:else}
                      <button
                        type="button"
                        onclick={(e) => startEditingSprint(sprint, e)}
                        class="text-sm font-medium text-gray-900 truncate hover:text-indigo-600 text-left"
                        title="Click to rename"
                      >
                        {sprint.name}
                      </button>
                    {/if}
                    <button
                      type="button"
                      onclick={(e) => openEditModal(sprint, e)}
                      class="p-0.5 text-gray-400 hover:text-gray-600 rounded"
                      title="Edit sprint details"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                      </svg>
                    </button>
                  </div>
                  {#if sprint.goal}
                    <p class="text-xs text-gray-500 mt-0.5 line-clamp-1">{sprint.goal}</p>
                  {/if}
                  {#if formatDateRange(sprint)}
                    <p class="text-xs text-gray-400 mt-0.5">{formatDateRange(sprint)}</p>
                  {/if}
                </div>
                {#if !activeSprint}
                  <Button
                    size="sm"
                    variant="ghost"
                    onclick={() => handleStartSprint(sprint)}
                    disabled={actionLoading}
                  >
                    Start
                  </Button>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}

      <!-- Closed Sprints -->
      {#if closedSprints.length > 0}
        <div class="py-1 border-t border-gray-100">
          <div class="px-3 py-1.5 text-xs font-medium text-gray-500 uppercase tracking-wide">
            Closed
          </div>
          {#each closedSprints as sprint (sprint.id)}
            <div class="px-3 py-2 hover:bg-gray-50">
              <div class="flex items-start justify-between gap-2">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-1.5">
                    {#if editingSprintId === sprint.id}
                      <input
                        type="text"
                        bind:value={editingName}
                        onkeydown={(e) => handleEditKeydown(e, sprint.id)}
                        onblur={() => saveSprintName(sprint.id)}
                        class="text-sm font-medium text-gray-600 border border-gray-300 rounded px-1.5 py-0.5 w-full focus:outline-none focus:ring-1 focus:ring-indigo-500"
                        autofocus
                      />
                    {:else}
                      <button
                        type="button"
                        onclick={(e) => startEditingSprint(sprint, e)}
                        class="text-sm font-medium text-gray-600 truncate hover:text-indigo-600 text-left"
                        title="Click to rename"
                      >
                        {sprint.name}
                      </button>
                    {/if}
                    <button
                      type="button"
                      onclick={(e) => openEditModal(sprint, e)}
                      class="p-0.5 text-gray-400 hover:text-gray-600 rounded"
                      title="Edit sprint details"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                      </svg>
                    </button>
                    <span class="text-xs px-1.5 py-0.5 rounded bg-gray-100 text-gray-500 flex-shrink-0">Closed</span>
                  </div>
                  {#if sprint.goal}
                    <p class="text-xs text-gray-400 mt-0.5 line-clamp-1">{sprint.goal}</p>
                  {/if}
                  {#if formatDateRange(sprint)}
                    <p class="text-xs text-gray-400 mt-0.5">{formatDateRange(sprint)}</p>
                  {/if}
                </div>
                <Button
                  size="sm"
                  variant="ghost"
                  onclick={() => handleReopenSprint(sprint)}
                  disabled={actionLoading}
                >
                  Reopen
                </Button>
              </div>
            </div>
          {/each}
          {#if closedSprintsPageInfo?.hasNextPage}
            <button
              type="button"
              class="w-full px-3 py-2 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-50 text-center"
              onclick={loadMoreClosedSprints}
              disabled={loadingMoreClosed}
            >
              {#if loadingMoreClosed}
                <svg class="w-3 h-3 mr-1 animate-spin inline" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Loading...
              {:else}
                Load more ({closedSprintsPageInfo.totalCount - closedSprints.length} remaining)
              {/if}
            </button>
          {/if}
        </div>
      {/if}

      <!-- Empty state -->
      {#if !activeSprint && futureSprints.length === 0 && closedSprints.length === 0}
        <div class="p-4 text-center text-sm text-gray-500">
          No sprints yet. Create your first sprint to get started.
        </div>
      {/if}
    </div>

    <!-- Footer -->
    <div class="p-2 border-t border-gray-100 bg-gray-50 rounded-b-lg">
      <button
        type="button"
        class="w-full px-3 py-2 text-sm text-left text-gray-700 hover:bg-gray-100 rounded-md flex items-center gap-2"
        onclick={() => { open = false; showCreateModal = true; }}
      >
        <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
        Create Sprint
      </button>
    </div>
  </Popover.Content>
</Popover.Root>

<CreateSprintModal
  open={showCreateModal}
  {boardId}
  onClose={() => showCreateModal = false}
  onCreated={handleSprintCreated}
/>

<EditSprintModal
  open={showEditModal}
  sprint={editingSprint}
  onClose={() => { showEditModal = false; editingSprint = null; }}
  onUpdated={handleSprintUpdated}
/>

<CompleteSprintModal
  open={showCompleteModal}
  sprint={activeSprint}
  {boardId}
  {incompleteCardCount}
  onClose={() => showCompleteModal = false}
  onCompleted={handleSprintCompleted}
/>
