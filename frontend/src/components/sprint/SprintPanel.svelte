<script lang="ts">
  import { onMount } from 'svelte';
  import { toast } from 'svelte-sonner';
  import { Button } from '../ui';
  import CreateSprintModal from './CreateSprintModal.svelte';
  import SprintSection from './SprintSection.svelte';
  import {
    getActiveSprint,
    getFutureSprints,
    getClosedSprints,
    getBacklogCards,
    startSprint,
    completeSprint,
    addCardToSprint,
    moveCardToBacklog,
    type SprintData,
    type BacklogCard,
    SprintStatus,
  } from '../../lib/api/sprints';

  interface Props {
    boardId: string;
    onBoardRefresh?: () => void;
  }

  let { boardId, onBoardRefresh }: Props = $props();

  let activeSprint = $state<SprintData | null>(null);
  let futureSprints = $state<SprintData[]>([]);
  let closedSprints = $state<SprintData[]>([]);
  let backlogCards = $state<BacklogCard[]>([]);
  let loading = $state(true);
  let showCreateModal = $state(false);
  let isExpanded = $state(true);

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      const [active, future, closedResult, backlog] = await Promise.all([
        getActiveSprint(boardId),
        getFutureSprints(boardId),
        getClosedSprints(boardId),
        getBacklogCards(boardId),
      ]);
      activeSprint = active;
      futureSprints = future;
      closedSprints = closedResult.sprints;
      backlogCards = backlog;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load sprint data';
      toast.error(message);
    } finally {
      loading = false;
    }
  }

  async function handleStartSprint(sprintId: string) {
    try {
      await startSprint(sprintId);
      toast.success('Sprint started');
      await loadData();
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to start sprint';
      toast.error(message);
    }
  }

  async function handleCompleteSprint(sprintId: string) {
    try {
      await completeSprint(sprintId, true);
      toast.success('Sprint completed');
      await loadData();
      onBoardRefresh?.();
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to complete sprint';
      toast.error(message);
    }
  }

  async function handleAddToSprint(cardId: string, sprintId: string) {
    try {
      await addCardToSprint(cardId, sprintId);
      toast.success('Card added to sprint');
      await loadData();
      onBoardRefresh?.();
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to add card to sprint';
      toast.error(message);
    }
  }

  async function handleMoveToBacklog(cardId: string) {
    try {
      await moveCardToBacklog(cardId);
      toast.success('Card moved to backlog');
      await loadData();
      onBoardRefresh?.();
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to move card';
      toast.error(message);
    }
  }

  async function handleSprintCreated() {
    showCreateModal = false;
    await loadData();
  }

  function formatDate(dateStr: string | null | undefined): string {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
    });
  }

  function getSprintDateRange(sprint: SprintData): string {
    const start = formatDate(sprint.startDate);
    const end = formatDate(sprint.endDate);
    if (start && end) return `${start} - ${end}`;
    if (start) return `Starts ${start}`;
    if (end) return `Ends ${end}`;
    return '';
  }
</script>

<div class="bg-white border-l border-gray-200 h-full flex flex-col {isExpanded ? 'w-80' : 'w-12'}">
  <!-- Toggle button -->
  <button
    type="button"
    onclick={() => isExpanded = !isExpanded}
    class="flex items-center justify-center h-10 border-b border-gray-200 hover:bg-gray-50"
    title={isExpanded ? 'Collapse panel' : 'Expand panel'}
  >
    <svg
      class="w-5 h-5 text-gray-500 transition-transform {isExpanded ? '' : 'rotate-180'}"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
    </svg>
  </button>

  {#if isExpanded}
    <div class="flex-1 overflow-y-auto">
      {#if loading}
        <div class="p-4 text-center text-gray-500">Loading...</div>
      {:else}
        <!-- Active Sprint Section -->
        <div class="border-b border-gray-200">
          <div class="p-3 bg-green-50">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-semibold text-green-800">Active Sprint</h3>
              {#if activeSprint}
                <Button
                  size="sm"
                  variant="secondary"
                  onclick={() => handleCompleteSprint(activeSprint!.id)}
                >
                  Complete
                </Button>
              {/if}
            </div>
            {#if activeSprint}
              <div class="mt-2">
                <p class="font-medium text-green-900">{activeSprint.name}</p>
                {#if activeSprint.goal}
                  <p class="text-xs text-green-700 mt-1">{activeSprint.goal}</p>
                {/if}
                {#if getSprintDateRange(activeSprint)}
                  <p class="text-xs text-green-600 mt-1">{getSprintDateRange(activeSprint)}</p>
                {/if}
              </div>
            {:else}
              <p class="text-sm text-green-700 mt-2">No active sprint</p>
              {#if futureSprints.length > 0}
                <Button
                  size="sm"
                  class="mt-2"
                  onclick={() => handleStartSprint(futureSprints[0].id)}
                >
                  Start {futureSprints[0].name}
                </Button>
              {/if}
            {/if}
          </div>
        </div>

        <!-- Backlog Section -->
        <SprintSection
          title="Backlog"
          cardCount={backlogCards.length}
          expanded={true}
        >
          {#if backlogCards.length === 0}
            <p class="text-sm text-gray-500 p-2">No cards in backlog</p>
          {:else}
            <ul class="divide-y divide-gray-100">
              {#each backlogCards as card (card.id)}
                <li class="p-2 hover:bg-gray-50 group">
                  <div class="flex items-start justify-between">
                    <span class="text-sm text-gray-900 flex-1">{card.title}</span>
                    {#if activeSprint}
                      <button
                        type="button"
                        onclick={() => handleAddToSprint(card.id, activeSprint!.id)}
                        class="text-xs text-indigo-600 hover:text-indigo-800 opacity-0 group-hover:opacity-100 transition-opacity ml-2"
                        title="Add to active sprint"
                      >
                        Add to sprint
                      </button>
                    {/if}
                  </div>
                  {#if card.column}
                    <span class="text-xs text-gray-500">{card.column.name}</span>
                  {/if}
                </li>
              {/each}
            </ul>
          {/if}
        </SprintSection>

        <!-- Future Sprints Section -->
        <SprintSection
          title="Future Sprints"
          cardCount={futureSprints.length}
          expanded={futureSprints.length > 0}
        >
          {#if futureSprints.length === 0}
            <p class="text-sm text-gray-500 p-2">No future sprints planned</p>
          {:else}
            <ul class="divide-y divide-gray-100">
              {#each futureSprints as sprint (sprint.id)}
                <li class="p-2 hover:bg-gray-50">
                  <div class="flex items-start justify-between">
                    <div class="flex-1">
                      <p class="text-sm font-medium text-gray-900">{sprint.name}</p>
                      {#if sprint.goal}
                        <p class="text-xs text-gray-500 mt-0.5">{sprint.goal}</p>
                      {/if}
                      {#if getSprintDateRange(sprint)}
                        <p class="text-xs text-gray-400 mt-0.5">{getSprintDateRange(sprint)}</p>
                      {/if}
                    </div>
                    {#if !activeSprint}
                      <button
                        type="button"
                        onclick={() => handleStartSprint(sprint.id)}
                        class="text-xs text-indigo-600 hover:text-indigo-800 ml-2"
                      >
                        Start
                      </button>
                    {/if}
                  </div>
                </li>
              {/each}
            </ul>
          {/if}
        </SprintSection>

        <!-- Closed Sprints Section (Sprint History) -->
        {#if closedSprints.length > 0}
          <SprintSection
            title="Closed Sprints"
            cardCount={closedSprints.length}
            expanded={false}
          >
            <ul class="divide-y divide-gray-100">
              {#each closedSprints as sprint (sprint.id)}
                <li class="p-2 hover:bg-gray-50">
                  <div class="flex-1">
                    <div class="flex items-center gap-2">
                      <p class="text-sm font-medium text-gray-700">{sprint.name}</p>
                      <span class="text-xs px-1.5 py-0.5 rounded bg-gray-100 text-gray-500">Closed</span>
                    </div>
                    {#if sprint.goal}
                      <p class="text-xs text-gray-500 mt-0.5">{sprint.goal}</p>
                    {/if}
                    {#if getSprintDateRange(sprint)}
                      <p class="text-xs text-gray-400 mt-0.5">{getSprintDateRange(sprint)}</p>
                    {/if}
                  </div>
                </li>
              {/each}
            </ul>
          </SprintSection>
        {/if}

        <!-- Create Sprint Button -->
        <div class="p-3">
          <Button
            variant="secondary"
            class="w-full"
            onclick={() => showCreateModal = true}
          >
            <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            Create Sprint
          </Button>
        </div>
      {/if}
    </div>
  {/if}
</div>

<CreateSprintModal
  open={showCreateModal}
  {boardId}
  onClose={() => showCreateModal = false}
  onCreated={handleSprintCreated}
/>
