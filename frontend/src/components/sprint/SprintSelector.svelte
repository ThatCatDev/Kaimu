<script lang="ts">
  import { onMount } from 'svelte';
  import { Popover } from 'bits-ui';
  import { toast } from 'svelte-sonner';
  import { Button } from '../ui';
  import CreateSprintModal from './CreateSprintModal.svelte';
  import {
    getActiveSprint,
    getFutureSprints,
    getClosedSprints,
    startSprint,
    completeSprint,
    updateSprint,
    type SprintData,
  } from '../../lib/api/sprints';

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
      await completeSprint(activeSprint.id, true);
      toast.success(`Completed ${activeSprint.name}`);
      open = false;
      await loadSprints();
      onSprintChange?.();
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to complete sprint';
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

<Popover.Root bind:open>
  <Popover.Trigger
    class="inline-flex items-center gap-2 px-3 py-1.5 text-sm font-medium rounded-md border transition-colors
      {activeSprint
        ? 'bg-green-50 border-green-200 text-green-800 hover:bg-green-100'
        : 'bg-gray-50 border-gray-200 text-gray-700 hover:bg-gray-100'}"
    disabled={loading}
  >
    {#if loading}
      <span class="text-gray-500">Loading...</span>
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
        <div class="p-3 bg-green-50 border-b border-gray-100">
          <div class="flex items-start justify-between gap-2">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-1.5">
                <span class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                  Active
                </span>
                {#if editingSprintId === activeSprint.id}
                  <input
                    type="text"
                    bind:value={editingName}
                    onkeydown={(e) => handleEditKeydown(e, activeSprint.id)}
                    onblur={() => saveSprintName(activeSprint.id)}
                    class="text-sm font-medium text-gray-900 border border-gray-300 rounded px-1.5 py-0.5 w-full focus:outline-none focus:ring-1 focus:ring-indigo-500"
                    autofocus
                  />
                {:else}
                  <button
                    type="button"
                    onclick={(e) => startEditingSprint(activeSprint, e)}
                    class="text-sm font-medium text-gray-900 truncate hover:text-indigo-600 text-left"
                    title="Click to rename"
                  >
                    {activeSprint.name}
                  </button>
                {/if}
              </div>
              {#if activeSprint.goal}
                <p class="text-xs text-gray-600 mt-1 line-clamp-2">{activeSprint.goal}</p>
              {/if}
              {#if formatDateRange(activeSprint)}
                <p class="text-xs text-gray-500 mt-1">{formatDateRange(activeSprint)}</p>
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
                    <span class="text-xs px-1.5 py-0.5 rounded bg-gray-100 text-gray-500 flex-shrink-0">Closed</span>
                  </div>
                  {#if sprint.goal}
                    <p class="text-xs text-gray-400 mt-0.5 line-clamp-1">{sprint.goal}</p>
                  {/if}
                  {#if formatDateRange(sprint)}
                    <p class="text-xs text-gray-400 mt-0.5">{formatDateRange(sprint)}</p>
                  {/if}
                </div>
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
