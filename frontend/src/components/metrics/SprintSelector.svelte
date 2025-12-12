<script lang="ts">
  import { onMount } from 'svelte';
  import { toast } from 'svelte-sonner';
  import { Select } from 'bits-ui';
  import { getActiveSprint, getFutureSprints, getClosedSprints, type SprintData } from '../../lib/api/sprints';

  interface Props {
    boardId: string;
    selectedSprintId: string | null;
    onSprintChange: (sprintId: string | null) => void;
  }

  let { boardId, selectedSprintId, onSprintChange }: Props = $props();

  let sprints = $state<SprintData[]>([]);
  let loading = $state(true);

  const selectedSprint = $derived(sprints.find(s => s.id === selectedSprintId) ?? null);

  onMount(async () => {
    await loadSprints();
  });

  async function loadSprints() {
    try {
      const [active, future, closedResult] = await Promise.all([
        getActiveSprint(boardId),
        getFutureSprints(boardId),
        getClosedSprints(boardId, 10),
      ]);

      // Combine all sprints, putting active first, then future, then closed
      const allSprints: SprintData[] = [];
      if (active) allSprints.push(active);
      allSprints.push(...future);
      allSprints.push(...closedResult.sprints);

      sprints = allSprints;

      // Auto-select the active sprint if no sprint is selected
      if (!selectedSprintId && active) {
        onSprintChange(active.id);
      }
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load sprints';
      toast.error(message);
    } finally {
      loading = false;
    }
  }

  function handleValueChange(value: string | undefined) {
    onSprintChange(value || null);
  }

  function formatDate(dateString: string | null | undefined): string {
    if (!dateString) return '';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function getStatusBadge(status: string): { text: string; class: string } {
    switch (status) {
      case 'ACTIVE':
        return { text: 'Active', class: 'bg-green-100 text-green-700' };
      case 'CLOSED':
        return { text: 'Closed', class: 'bg-gray-100 text-gray-600' };
      case 'FUTURE':
        return { text: 'Future', class: 'bg-blue-100 text-blue-700' };
      default:
        return { text: status, class: 'bg-gray-100 text-gray-600' };
    }
  }
</script>

<div class="flex items-center gap-2">
  <span class="text-sm font-medium text-gray-700">Sprint:</span>

  <Select.Root
    type="single"
    value={selectedSprintId ?? undefined}
    onValueChange={handleValueChange}
    disabled={loading}
  >
    <Select.Trigger
      class="inline-flex items-center justify-between gap-2 rounded-md border border-gray-300 bg-white py-1.5 px-3 text-sm shadow-sm hover:bg-gray-50 focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500 disabled:bg-gray-50 disabled:text-gray-500 min-w-[220px]"
      aria-label="Select a sprint"
    >
      {#if loading}
        <div class="h-4 w-28 bg-gray-200 rounded animate-pulse"></div>
      {:else if selectedSprint}
        <span class="flex items-center gap-2">
          <span class="font-medium">{selectedSprint.name}</span>
          {#if selectedSprint.status === 'ACTIVE'}
            <span class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-green-100 text-green-700">
              Active
            </span>
          {/if}
        </span>
      {:else}
        <span class="text-gray-500">Select a sprint</span>
      {/if}
      <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l4-4 4 4m0 6l-4 4-4-4" />
      </svg>
    </Select.Trigger>

    <Select.Portal>
      <Select.Content
        class="z-50 min-w-[280px] rounded-lg border border-gray-200 bg-white shadow-lg"
        sideOffset={4}
      >
        <Select.Viewport class="p-1">
          {#if sprints.length === 0}
            <div class="px-3 py-2 text-sm text-gray-500">No sprints available</div>
          {:else}
            {#each sprints as sprint (sprint.id)}
              {@const badge = getStatusBadge(sprint.status)}
              <Select.Item
                value={sprint.id}
                label={sprint.name}
                class="relative flex cursor-pointer select-none items-center rounded-md px-3 py-2 text-sm outline-none hover:bg-gray-100 focus:bg-gray-100 data-[highlighted]:bg-gray-100 data-[selected]:bg-indigo-50"
              >
                {#snippet children({ selected })}
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <span class="font-medium text-gray-900 truncate">{sprint.name}</span>
                      <span class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium {badge.class}">
                        {badge.text}
                      </span>
                    </div>
                    {#if sprint.startDate || sprint.endDate}
                      <div class="text-xs text-gray-500 mt-0.5">
                        {#if sprint.startDate && sprint.endDate}
                          {formatDate(sprint.startDate)} - {formatDate(sprint.endDate)}
                        {:else if sprint.startDate}
                          Starts {formatDate(sprint.startDate)}
                        {:else if sprint.endDate}
                          Ends {formatDate(sprint.endDate)}
                        {/if}
                      </div>
                    {/if}
                  </div>
                  {#if selected}
                    <svg class="w-4 h-4 text-indigo-600 flex-shrink-0 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                  {/if}
                {/snippet}
              </Select.Item>
            {/each}
          {/if}
        </Select.Viewport>
      </Select.Content>
    </Select.Portal>
  </Select.Root>
</div>
