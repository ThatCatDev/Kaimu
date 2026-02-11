<script lang="ts">
  import { toast } from 'svelte-sonner';
  import { bulkSelection } from '../../lib/stores/bulkSelection.svelte';
  import { ConfirmModal } from '../ui';
  import type { BoardColumn, Tag } from '../../lib/api/boards';
  import type { SprintData } from '../../lib/api/sprints';
  import { CardPriority } from '../../lib/graphql/generated';

  interface Props {
    columns: BoardColumn[];
    tags: Tag[];
    sprints: SprintData[];
    activeSprint: SprintData | null;
    projectMembers: { id: string; username: string; displayName?: string | null }[];
    onBulkActionComplete: () => void;
  }

  let { columns, tags, sprints, activeSprint, projectMembers, onBulkActionComplete }: Props = $props();

  let showDeleteConfirm = $state(false);
  let showActionDropdown = $state<string | null>(null);

  let selectedCount = $derived(bulkSelection.selectedCount);
  let isLoading = $derived(bulkSelection.isLoading);

  function handleClickOutside() {
    showActionDropdown = null;
  }

  async function handleMoveToColumn(columnId: string) {
    showActionDropdown = null;
    try {
      await bulkSelection.bulkUpdate({ columnId });
      toast.success(`Moved ${selectedCount} cards`);
      onBulkActionComplete();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to move cards');
    }
  }

  async function handleAssign(userId: string | null) {
    showActionDropdown = null;
    try {
      if (userId === null) {
        await bulkSelection.bulkUpdate({ clearAssignee: true });
        toast.success(`Unassigned ${selectedCount} cards`);
      } else {
        await bulkSelection.bulkUpdate({ assigneeId: userId });
        toast.success(`Assigned ${selectedCount} cards`);
      }
      onBulkActionComplete();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to assign cards');
    }
  }

  async function handleSetPriority(priority: CardPriority) {
    showActionDropdown = null;
    try {
      await bulkSelection.bulkUpdate({ priority });
      toast.success(`Updated priority for ${selectedCount} cards`);
      onBulkActionComplete();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to update priority');
    }
  }

  async function handleAddToSprint(sprintId: string) {
    showActionDropdown = null;
    try {
      await bulkSelection.bulkAddToSprint(sprintId);
      toast.success(`Added ${selectedCount} cards to sprint`);
      onBulkActionComplete();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to add cards to sprint');
    }
  }

  async function handleRemoveFromSprint(sprintId: string) {
    showActionDropdown = null;
    try {
      await bulkSelection.bulkRemoveFromSprint(sprintId);
      toast.success(`Removed ${selectedCount} cards from sprint`);
      onBulkActionComplete();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to remove cards from sprint');
    }
  }

  async function handleDelete() {
    showDeleteConfirm = false;
    try {
      const count = await bulkSelection.bulkDelete();
      toast.success(`Deleted ${count} cards`);
      onBulkActionComplete();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Failed to delete cards');
    }
  }

  function toggleDropdown(dropdownName: string) {
    if (showActionDropdown === dropdownName) {
      showActionDropdown = null;
    } else {
      showActionDropdown = dropdownName;
    }
  }

  const priorityOptions = [
    { value: CardPriority.None, label: 'None', color: 'gray' },
    { value: CardPriority.Low, label: 'Low', color: 'blue' },
    { value: CardPriority.Medium, label: 'Medium', color: 'yellow' },
    { value: CardPriority.High, label: 'High', color: 'orange' },
    { value: CardPriority.Urgent, label: 'Urgent', color: 'red' },
  ];
</script>

<svelte:window onclick={handleClickOutside} />

{#if selectedCount > 0}
  <div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-50">
    <div class="flex items-center gap-2 bg-white rounded-lg shadow-lg border border-gray-200 px-4 py-2">
      <!-- Selection count -->
      <span class="text-sm font-medium text-gray-700 mr-2">
        {selectedCount} card{selectedCount !== 1 ? 's' : ''} selected
      </span>

      <div class="h-4 w-px bg-gray-300"></div>

      <!-- Move to column -->
      <div class="relative">
        <button
          type="button"
          onclick={(e) => { e.stopPropagation(); toggleDropdown('column'); }}
          disabled={isLoading}
          class="px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md transition-colors disabled:opacity-50"
        >
          Move
        </button>
        {#if showActionDropdown === 'column'}
          <div
            class="absolute bottom-full left-0 mb-1 w-48 bg-white rounded-md shadow-lg border border-gray-200 py-1 max-h-64 overflow-y-auto"
            onclick={(e) => e.stopPropagation()}
          >
            {#each columns.filter(c => !c.isHidden) as column}
              <button
                type="button"
                onclick={() => handleMoveToColumn(column.id)}
                class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100"
              >
                {column.name}
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Assign -->
      <div class="relative">
        <button
          type="button"
          onclick={(e) => { e.stopPropagation(); toggleDropdown('assign'); }}
          disabled={isLoading}
          class="px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md transition-colors disabled:opacity-50"
        >
          Assign
        </button>
        {#if showActionDropdown === 'assign'}
          <div
            class="absolute bottom-full left-0 mb-1 w-48 bg-white rounded-md shadow-lg border border-gray-200 py-1 max-h-64 overflow-y-auto"
            onclick={(e) => e.stopPropagation()}
          >
            <button
              type="button"
              onclick={() => handleAssign(null)}
              class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 text-gray-500 italic"
            >
              Unassigned
            </button>
            {#each projectMembers as member}
              <button
                type="button"
                onclick={() => handleAssign(member.id)}
                class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100"
              >
                {member.displayName || member.username}
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Priority -->
      <div class="relative">
        <button
          type="button"
          onclick={(e) => { e.stopPropagation(); toggleDropdown('priority'); }}
          disabled={isLoading}
          class="px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md transition-colors disabled:opacity-50"
        >
          Priority
        </button>
        {#if showActionDropdown === 'priority'}
          <div
            class="absolute bottom-full left-0 mb-1 w-40 bg-white rounded-md shadow-lg border border-gray-200 py-1"
            onclick={(e) => e.stopPropagation()}
          >
            {#each priorityOptions as option}
              <button
                type="button"
                onclick={() => handleSetPriority(option.value)}
                class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 flex items-center gap-2"
              >
                <span class="w-2 h-2 rounded-full bg-{option.color}-400"></span>
                {option.label}
              </button>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Sprint -->
      {#if sprints.length > 0 || activeSprint}
        <div class="relative">
          <button
            type="button"
            onclick={(e) => { e.stopPropagation(); toggleDropdown('sprint'); }}
            disabled={isLoading}
            class="px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-100 rounded-md transition-colors disabled:opacity-50"
          >
            Sprint
          </button>
          {#if showActionDropdown === 'sprint'}
            <div
              class="absolute bottom-full left-0 mb-1 w-56 bg-white rounded-md shadow-lg border border-gray-200 py-1 max-h-64 overflow-y-auto"
              onclick={(e) => e.stopPropagation()}
            >
              <div class="px-3 py-1 text-xs text-gray-500 uppercase tracking-wide">Add to Sprint</div>
              {#each sprints.filter(s => s.status !== 'CLOSED') as sprint}
                <button
                  type="button"
                  onclick={() => handleAddToSprint(sprint.id)}
                  class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 flex items-center justify-between"
                >
                  <span>{sprint.name}</span>
                  {#if sprint.status === 'ACTIVE'}
                    <span class="text-xs text-green-600 font-medium">Active</span>
                  {/if}
                </button>
              {/each}
              {#if activeSprint}
                <div class="border-t border-gray-200 mt-1 pt-1">
                  <div class="px-3 py-1 text-xs text-gray-500 uppercase tracking-wide">Remove from Sprint</div>
                  <button
                    type="button"
                    onclick={() => handleRemoveFromSprint(activeSprint.id)}
                    class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100"
                  >
                    {activeSprint.name}
                  </button>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {/if}

      <div class="h-4 w-px bg-gray-300"></div>

      <!-- Delete -->
      <button
        type="button"
        onclick={() => showDeleteConfirm = true}
        disabled={isLoading}
        class="px-3 py-1.5 text-sm font-medium text-red-600 hover:bg-red-50 rounded-md transition-colors disabled:opacity-50"
      >
        Delete
      </button>

      <div class="h-4 w-px bg-gray-300"></div>

      <!-- Clear selection -->
      <button
        type="button"
        onclick={() => bulkSelection.disableSelectionMode()}
        disabled={isLoading}
        class="px-3 py-1.5 text-sm font-medium text-gray-500 hover:bg-gray-100 rounded-md transition-colors disabled:opacity-50"
      >
        Cancel
      </button>

      {#if isLoading}
        <div class="ml-2">
          <svg class="animate-spin h-4 w-4 text-indigo-600" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>
      {/if}
    </div>
  </div>
{/if}

<ConfirmModal
  isOpen={showDeleteConfirm}
  title="Delete Cards"
  message="Are you sure you want to delete {selectedCount} card{selectedCount !== 1 ? 's' : ''}? This action cannot be undone."
  confirmText="Delete"
  cancelText="Cancel"
  variant="danger"
  onConfirm={handleDelete}
  onCancel={() => showDeleteConfirm = false}
/>
