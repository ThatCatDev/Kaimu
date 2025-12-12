<script lang="ts">
  import { updateCard, deleteCard, type BoardCard, type Tag } from '../../lib/api/boards';
  import { CardPriority, SprintStatus } from '../../lib/graphql/generated';
  import { Button, ConfirmModal, AssigneeCombobox } from '../ui';
  import CardForm from './CardForm.svelte';
  import { getActiveSprint, getFutureSprints, getClosedSprints, setCardSprints, type SprintData } from '../../lib/api/sprints';

  interface Props {
    card: BoardCard | null;
    projectId: string;
    organizationId: string;
    boardId: string;
    tags: Tag[];
    isOpen: boolean;
    onClose: () => void;
    onUpdated: () => void;
    onCardDataChanged?: (cardId: string, updates: Partial<BoardCard>) => void;
    onTagsChanged?: () => void;
    viewMode: 'modal' | 'panel';
    onViewModeChange: (mode: 'modal' | 'panel') => void;
    // Permission props
    canEditCard?: boolean;
    canDeleteCard?: boolean;
  }

  let { card, projectId, organizationId, boardId, tags, isOpen, onClose, onUpdated, onCardDataChanged, onTagsChanged, viewMode, onViewModeChange, canEditCard = true, canDeleteCard = true }: Props = $props();


  let title = $state('');
  let description = $state('');
  let priority = $state<CardPriority>(CardPriority.None);
  let selectedTagIds = $state<string[]>([]);
  let dueDate = $state('');
  let saving = $state(false);
  let deleting = $state(false);
  let error = $state<string | null>(null);
  let saveTimeout: ReturnType<typeof setTimeout> | null = null;
  let lastSavedData = $state<string>('');
  let showDeleteConfirm = $state(false);

  // Sprint state
  let activeSprints = $state<SprintData[]>([]);
  let futureSprints = $state<SprintData[]>([]);
  let closedSprints = $state<SprintData[]>([]);
  let closedSprintsPageInfo = $state<{ hasNextPage: boolean; endCursor: string | null; totalCount: number } | null>(null);
  let selectedSprintIds = $state<string[]>([]);
  let loadingSprints = $state(false);
  let loadingMoreClosed = $state(false);
  let savingSprints = $state(false);

  // Assignee state
  let selectedAssigneeId = $state<string | null>(null);
  let savingAssignee = $state(false);

  let currentCardId = $state<string | null>(null);
  let isEditing = $state(false); // Track if user has started editing

  // Use derived instead of function call in template to avoid re-render triggers
  const currentDataHash = $derived(JSON.stringify({ title, description, priority, selectedTagIds, dueDate }));
  const isSaved = $derived(currentDataHash === lastSavedData);

  // Sprint UI state
  let sprintSearch = $state('');
  let showAllClosedSprints = $state(false);
  const CLOSED_SPRINTS_PREVIEW_COUNT = 5;
  const CLOSED_SPRINTS_PAGE_SIZE = 10;

  // Combine all sprints for operations
  const availableSprints = $derived([...activeSprints, ...futureSprints, ...closedSprints]);

  // Filter sprints by search
  const filteredActiveSprints = $derived(
    sprintSearch ? activeSprints.filter(s => s.name.toLowerCase().includes(sprintSearch.toLowerCase())) : activeSprints
  );
  const filteredFutureSprints = $derived(
    sprintSearch ? futureSprints.filter(s => s.name.toLowerCase().includes(sprintSearch.toLowerCase())) : futureSprints
  );
  const filteredClosedSprints = $derived(
    sprintSearch ? closedSprints.filter(s => s.name.toLowerCase().includes(sprintSearch.toLowerCase())) : closedSprints
  );

  // Limit closed sprints display unless searching or expanded
  const displayedClosedSprints = $derived(
    sprintSearch || showAllClosedSprints ? filteredClosedSprints : filteredClosedSprints.slice(0, CLOSED_SPRINTS_PREVIEW_COUNT)
  );
  const hasMoreClosedSprints = $derived(filteredClosedSprints.length > CLOSED_SPRINTS_PREVIEW_COUNT);
  const totalSprintCount = $derived(availableSprints.length);

  // Reset state when panel closes
  $effect(() => {
    if (!isOpen) {
      currentCardId = null;
      isEditing = false;
    }
  });

  // Load sprints and members when panel opens
  $effect(() => {
    if (isOpen && boardId) {
      loadSprints();
    }
  });

  async function loadSprints() {
    try {
      loadingSprints = true;
      const [active, future, closedResult] = await Promise.all([
        getActiveSprint(boardId),
        getFutureSprints(boardId),
        getClosedSprints(boardId, CLOSED_SPRINTS_PAGE_SIZE),
      ]);
      activeSprints = active ? [active] : [];
      futureSprints = future;
      closedSprints = closedResult.sprints;
      closedSprintsPageInfo = closedResult.pageInfo;
    } catch (e) {
      console.error('Failed to load sprints:', e);
    } finally {
      loadingSprints = false;
    }
  }

  async function loadMoreClosedSprints() {
    if (!closedSprintsPageInfo?.hasNextPage || loadingMoreClosed) return;
    try {
      loadingMoreClosed = true;
      const result = await getClosedSprints(boardId, CLOSED_SPRINTS_PAGE_SIZE, closedSprintsPageInfo.endCursor ?? undefined);
      closedSprints = [...closedSprints, ...result.sprints];
      closedSprintsPageInfo = result.pageInfo;
    } catch (e) {
      console.error('Failed to load more closed sprints:', e);
    } finally {
      loadingMoreClosed = false;
    }
  }

  // Load card data ONLY when a DIFFERENT card is selected
  // Never reload while editing the same card (prevents form reset during auto-save)
  $effect(() => {
    if (isOpen && card && card.id !== currentCardId) {
      // New card selected - load its data
      currentCardId = card.id;
      title = card.title;
      description = card.description ?? '';
      priority = card.priority;
      selectedTagIds = card.tags?.map(t => t.id) ?? [];
      selectedSprintIds = card.sprints?.map(s => s.id) ?? [];
      selectedAssigneeId = card.assignee?.id ?? null;
      dueDate = card.dueDate ? card.dueDate.split('T')[0] : '';
      error = null;
      isEditing = false;
      // Set lastSavedData to current values so isSaved shows correctly
      lastSavedData = JSON.stringify({ title, description, priority, selectedTagIds, dueDate });
    }
  });

  // Auto-save effect
  $effect(() => {
    if (card && lastSavedData && currentDataHash !== lastSavedData && title.trim()) {
      // User has made changes - mark as editing
      isEditing = true;
      if (saveTimeout) clearTimeout(saveTimeout);
      saveTimeout = setTimeout(() => {
        autoSave();
      }, 800);
    }
    return () => {
      if (saveTimeout) clearTimeout(saveTimeout);
    };
  });

  async function autoSave() {
    if (!card || !title.trim() || saving || !canEditCard) return;

    try {
      saving = true;
      error = null;
      const dueDateRfc3339 = dueDate ? new Date(dueDate + 'T00:00:00Z').toISOString() : null;
      await updateCard(
        card.id,
        title.trim(),
        description.trim() || undefined,
        priority,
        undefined,
        selectedTagIds,
        dueDateRfc3339
      );
      lastSavedData = currentDataHash;

      // Update card display on board without resetting form
      onCardDataChanged?.(card.id, {
        title: title.trim(),
        description: description.trim() || undefined,
        priority,
        dueDate: dueDateRfc3339,
        tags: tags.filter(t => selectedTagIds.includes(t.id))
      });
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to save';
    } finally {
      saving = false;
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    await handleClose();
  }

  async function handleClose() {
    if (saveTimeout) {
      clearTimeout(saveTimeout);
      await autoSave();
    }
    onUpdated();
  }

  function handleDeleteClick() {
    showDeleteConfirm = true;
  }

  async function confirmDelete() {
    if (!card) return;

    try {
      deleting = true;
      showDeleteConfirm = false;
      error = null;
      await deleteCard(card.id);
      onUpdated();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete card';
    } finally {
      deleting = false;
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  // Stable callback functions to prevent re-renders
  function handleTitleChange(v: string) { title = v; }
  function handleDescriptionChange(v: string) { description = v; }
  function handlePriorityChange(v: typeof priority) { priority = v; }
  function handleDueDateChange(v: string) { dueDate = v; }
  function handleTagSelectionChange(ids: string[]) { selectedTagIds = ids; }

  // Handle assignee selection changes
  async function handleAssigneeChange(userId: string | null, displayName?: string) {
    console.log('CardDetailPanel handleAssigneeChange called with:', userId, displayName);
    console.log('card:', card?.id, 'canEditCard:', canEditCard);

    if (!card || !canEditCard) {
      console.log('Early return - card or canEditCard is falsy');
      return;
    }

    try {
      savingAssignee = true;
      console.log('Calling updateCard with assigneeId:', userId);
      // Pass null to clear assignee, undefined to not change it
      await updateCard(card.id, undefined, undefined, undefined, userId, undefined, undefined);
      console.log('updateCard completed successfully');
      selectedAssigneeId = userId;

      // Update the board display with the displayName for proper avatar rendering
      onCardDataChanged?.(card.id, {
        assignee: userId ? { id: userId, username: displayName ?? '', displayName: displayName ?? '' } : undefined
      });
    } catch (e) {
      console.error('updateCard failed:', e);
      error = e instanceof Error ? e.message : 'Failed to update assignee';
    } finally {
      savingAssignee = false;
    }
  }

  // Handle sprint selection changes
  async function handleSprintToggle(sprintId: string) {
    if (!card || !canEditCard) return;

    const isSelected = selectedSprintIds.includes(sprintId);
    const newSprintIds = isSelected
      ? selectedSprintIds.filter(id => id !== sprintId)
      : [...selectedSprintIds, sprintId];

    try {
      savingSprints = true;
      await setCardSprints(card.id, newSprintIds);
      selectedSprintIds = newSprintIds;

      // Update the board display
      onCardDataChanged?.(card.id, {
        sprints: availableSprints.filter(s => newSprintIds.includes(s.id)).map(s => ({
          id: s.id,
          name: s.name,
          status: s.status
        }))
      });
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to update sprints';
    } finally {
      savingSprints = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!isOpen) return;

    if (e.key === 'Escape') {
      handleClose();
    }
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      handleSubmit(e);
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- Panel -->
<div
  class="fixed inset-y-0 right-0 w-[420px] bg-white shadow-2xl z-50 flex flex-col transition-transform duration-300 ease-out {isOpen ? 'translate-x-0' : 'translate-x-full'}"
>
  {#if card}
    <form onsubmit={handleSubmit} class="flex flex-col h-full">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between flex-shrink-0">
        <h2 class="text-lg font-semibold text-gray-900">Card Details</h2>
        <div class="flex items-center gap-1">
          <button
            type="button"
            class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
            onclick={() => onViewModeChange('modal')}
            title="Switch to modal view"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h12a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V6z" />
            </svg>
          </button>
          <button
            type="button"
            class="p-1 text-gray-400 hover:text-gray-600 rounded-md hover:bg-gray-100 transition-colors"
            onclick={handleClose}
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto px-6 py-4">
        <CardForm
          {title}
          {description}
          {priority}
          {dueDate}
          {selectedTagIds}
          {projectId}
          {tags}
          onTitleChange={handleTitleChange}
          onDescriptionChange={handleDescriptionChange}
          onPriorityChange={handlePriorityChange}
          onDueDateChange={handleDueDateChange}
          onTagSelectionChange={handleTagSelectionChange}
          {onTagsChanged}
          {error}
          disabled={deleting}
          readOnly={!canEditCard}
          descriptionRows={5}
          idPrefix="detail-"
        />

        <!-- Assignee Selection -->
        <div class="mt-4 pt-4 border-t border-gray-200">
          <div class="flex items-center justify-between mb-2">
            <span class="block text-sm font-medium text-gray-700">Assignee</span>
            {#if savingAssignee}
              <span class="text-xs text-gray-400">Saving...</span>
            {/if}
          </div>
          <AssigneeCombobox
            value={selectedAssigneeId}
            {organizationId}
            initialUserName={card?.assignee?.displayName ?? card?.assignee?.username}
            placeholder="Search for a user..."
            disabled={!canEditCard || savingAssignee}
            readOnly={!canEditCard}
            onValueChange={handleAssigneeChange}
          />
        </div>

        <!-- Sprint Selection -->
        {#if availableSprints.length > 0 || selectedSprintIds.length > 0}
          <div class="mt-4 pt-4 border-t border-gray-200">
            <div class="flex items-center justify-between mb-2">
              <label class="block text-sm font-medium text-gray-700">Sprints</label>
              {#if savingSprints}
                <span class="text-xs text-gray-400">Saving...</span>
              {/if}
            </div>
            {#if loadingSprints}
              <div class="text-sm text-gray-400">Loading sprints...</div>
            {:else if availableSprints.length === 0}
              <div class="text-sm text-gray-500">No sprints available</div>
            {:else}
              <!-- Search input for large sprint lists -->
              {#if totalSprintCount > 10}
                <div class="mb-3">
                  <input
                    type="text"
                    bind:value={sprintSearch}
                    placeholder="Search sprints..."
                    class="w-full px-3 py-1.5 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>
              {/if}

              <div class="space-y-3 max-h-64 overflow-y-auto">
                <!-- Active Sprints -->
                {#if filteredActiveSprints.length > 0}
                  <div>
                    <div class="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">Active</div>
                    <div class="space-y-1">
                      {#each filteredActiveSprints as sprint (sprint.id)}
                        {@const isSelected = selectedSprintIds.includes(sprint.id)}
                        <button
                          type="button"
                          onclick={() => handleSprintToggle(sprint.id)}
                          disabled={!canEditCard || savingSprints}
                          class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-left transition-colors {isSelected ? 'bg-indigo-50 border border-indigo-200' : 'bg-gray-50 border border-gray-200 hover:bg-gray-100'} {!canEditCard ? 'cursor-not-allowed opacity-60' : ''}"
                        >
                          <span class="w-4 h-4 flex items-center justify-center rounded {isSelected ? 'bg-indigo-600 text-white' : 'border border-gray-300'}">
                            {#if isSelected}
                              <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                              </svg>
                            {/if}
                          </span>
                          <span class="flex-1 text-sm {isSelected ? 'text-indigo-700 font-medium' : 'text-gray-700'}">{sprint.name}</span>
                          <span class="text-xs px-1.5 py-0.5 rounded bg-green-100 text-green-700">Active</span>
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Future Sprints -->
                {#if filteredFutureSprints.length > 0}
                  <div>
                    <div class="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">Future</div>
                    <div class="space-y-1">
                      {#each filteredFutureSprints as sprint (sprint.id)}
                        {@const isSelected = selectedSprintIds.includes(sprint.id)}
                        <button
                          type="button"
                          onclick={() => handleSprintToggle(sprint.id)}
                          disabled={!canEditCard || savingSprints}
                          class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-left transition-colors {isSelected ? 'bg-indigo-50 border border-indigo-200' : 'bg-gray-50 border border-gray-200 hover:bg-gray-100'} {!canEditCard ? 'cursor-not-allowed opacity-60' : ''}"
                        >
                          <span class="w-4 h-4 flex items-center justify-center rounded {isSelected ? 'bg-indigo-600 text-white' : 'border border-gray-300'}">
                            {#if isSelected}
                              <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                              </svg>
                            {/if}
                          </span>
                          <span class="flex-1 text-sm {isSelected ? 'text-indigo-700 font-medium' : 'text-gray-700'}">{sprint.name}</span>
                          <span class="text-xs px-1.5 py-0.5 rounded bg-blue-100 text-blue-700">Future</span>
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Closed Sprints -->
                {#if filteredClosedSprints.length > 0 || closedSprintsPageInfo?.hasNextPage}
                  <div>
                    <div class="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">
                      Closed {#if closedSprintsPageInfo}({closedSprintsPageInfo.totalCount}){/if}
                    </div>
                    <div class="space-y-1">
                      {#each displayedClosedSprints as sprint (sprint.id)}
                        {@const isSelected = selectedSprintIds.includes(sprint.id)}
                        <button
                          type="button"
                          onclick={() => handleSprintToggle(sprint.id)}
                          disabled={!canEditCard || savingSprints}
                          class="w-full flex items-center gap-2 px-3 py-2 rounded-md text-left transition-colors {isSelected ? 'bg-indigo-50 border border-indigo-200' : 'bg-gray-50 border border-gray-200 hover:bg-gray-100'} {!canEditCard ? 'cursor-not-allowed opacity-60' : ''}"
                        >
                          <span class="w-4 h-4 flex items-center justify-center rounded {isSelected ? 'bg-indigo-600 text-white' : 'border border-gray-300'}">
                            {#if isSelected}
                              <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                              </svg>
                            {/if}
                          </span>
                          <span class="flex-1 text-sm {isSelected ? 'text-indigo-700 font-medium' : 'text-gray-600'}">{sprint.name}</span>
                          <span class="text-xs px-1.5 py-0.5 rounded bg-gray-100 text-gray-500">Closed</span>
                        </button>
                      {/each}
                    </div>
                    <!-- Show more/less toggle for local preview -->
                    {#if hasMoreClosedSprints && !sprintSearch && !showAllClosedSprints}
                      <button
                        type="button"
                        onclick={() => showAllClosedSprints = true}
                        class="w-full mt-1 px-3 py-1.5 text-xs text-indigo-600 hover:text-indigo-800 hover:bg-indigo-50 rounded-md transition-colors"
                      >
                        Show {filteredClosedSprints.length - CLOSED_SPRINTS_PREVIEW_COUNT} more loaded
                      </button>
                    {/if}
                    <!-- Load more from server -->
                    {#if closedSprintsPageInfo?.hasNextPage && (showAllClosedSprints || sprintSearch || !hasMoreClosedSprints)}
                      <button
                        type="button"
                        onclick={loadMoreClosedSprints}
                        disabled={loadingMoreClosed}
                        class="w-full mt-1 px-3 py-1.5 text-xs text-indigo-600 hover:text-indigo-800 hover:bg-indigo-50 rounded-md transition-colors disabled:opacity-50"
                      >
                        {loadingMoreClosed ? 'Loading...' : `Load more (${closedSprintsPageInfo.totalCount - closedSprints.length} remaining)`}
                      </button>
                    {/if}
                    <!-- Show less when expanded -->
                    {#if showAllClosedSprints && hasMoreClosedSprints && !sprintSearch}
                      <button
                        type="button"
                        onclick={() => showAllClosedSprints = false}
                        class="w-full mt-1 px-3 py-1.5 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-50 rounded-md transition-colors"
                      >
                        Show less
                      </button>
                    {/if}
                  </div>
                {/if}

                <!-- No results -->
                {#if sprintSearch && filteredActiveSprints.length === 0 && filteredFutureSprints.length === 0 && filteredClosedSprints.length === 0}
                  <div class="text-sm text-gray-500 text-center py-2">No sprints match "{sprintSearch}"</div>
                {/if}
              </div>
            {/if}
          </div>
        {/if}

        <div class="pt-4 mt-4 border-t border-gray-200 text-xs text-gray-500">
          <p>Created: {formatDate(card.createdAt)}</p>
          {#if card.updatedAt !== card.createdAt}
            <p>Updated: {formatDate(card.updatedAt)}</p>
          {/if}
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-4 border-t border-gray-200 flex-shrink-0">
        <div class="flex items-center justify-between mb-4">
          {#if canDeleteCard}
            <Button variant="danger" size="sm" onclick={handleDeleteClick} disabled={deleting || saving}>
              {deleting ? 'Deleting...' : 'Delete'}
            </Button>
          {:else}
            <div></div>
          {/if}
          <div class="flex items-center gap-4">
            {#if canEditCard}
              {#if saving}
                <span class="text-xs text-gray-400">Saving...</span>
              {:else if isSaved}
                <span class="text-xs text-green-600">Saved</span>
              {/if}
            {/if}
            <Button variant="secondary" size="sm" onclick={handleClose} disabled={deleting}>
              Close
            </Button>
          </div>
        </div>
        <div class="text-xs text-gray-400 text-center">
          <kbd class="px-1.5 py-0.5 bg-gray-100 border border-gray-300 rounded text-gray-600">Esc</kbd> to close
          {#if canEditCard}
            <span class="mx-2">·</span>
            Auto-saves as you type
          {/if}
        </div>
      </div>
    </form>
  {:else}
    <div class="flex-1 flex items-center justify-center text-gray-400">
      <p>Select a card to view details</p>
    </div>
  {/if}
</div>

<ConfirmModal
  isOpen={showDeleteConfirm}
  title="Delete Card"
  message="Are you sure you want to delete this card? This action cannot be undone."
  confirmText="Delete"
  cancelText="Cancel"
  variant="danger"
  onConfirm={confirmDelete}
  onCancel={() => showDeleteConfirm = false}
/>
