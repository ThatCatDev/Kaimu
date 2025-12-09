<script lang="ts">
  import { updateCard, deleteCard, createLabel, type BoardCard, type Label } from '../../lib/api/boards';
  import { CardPriority } from '../../lib/graphql/generated';
  import { Input, Textarea, Select, Button, ConfirmModal } from '../ui';

  interface Props {
    card: BoardCard | null;
    projectId: string;
    labels: Label[];
    isOpen: boolean;
    onClose: () => void;
    onUpdated: () => void;
    onLabelsChanged?: () => void;
    viewMode: 'modal' | 'panel';
    onViewModeChange: (mode: 'modal' | 'panel') => void;
  }

  let { card, projectId, labels, isOpen, onClose, onUpdated, onLabelsChanged, viewMode, onViewModeChange }: Props = $props();

  let title = $state('');
  let description = $state('');
  let priority = $state<CardPriority>(CardPriority.None);
  let selectedLabelIds = $state<string[]>([]);
  let dueDate = $state('');
  let saving = $state(false);
  let deleting = $state(false);
  let error = $state<string | null>(null);
  let saveTimeout: ReturnType<typeof setTimeout> | null = null;
  let lastSavedData = $state<string>('');
  let showDeleteConfirm = $state(false);

  // Label search/create state
  let labelSearch = $state('');
  let showLabelDropdown = $state(false);
  let creatingLabel = $state(false);

  const presetColors = [
    '#ef4444', '#f97316', '#f59e0b', '#eab308',
    '#84cc16', '#22c55e', '#14b8a6', '#06b6d4',
    '#0ea5e9', '#3b82f6', '#6366f1', '#8b5cf6',
    '#a855f7', '#d946ef', '#ec4899', '#f43f5e',
  ];

  let filteredLabels = $derived(
    labelSearch.trim()
      ? labels.filter(l => l.name.toLowerCase().includes(labelSearch.toLowerCase()))
      : labels
  );

  let exactMatch = $derived(
    labels.find(l => l.name.toLowerCase() === labelSearch.trim().toLowerCase())
  );

  function getCurrentDataHash(): string {
    return JSON.stringify({ title, description, priority, selectedLabelIds, dueDate });
  }

  // Update form when card changes
  $effect(() => {
    if (card) {
      title = card.title;
      description = card.description ?? '';
      priority = card.priority;
      selectedLabelIds = card.labels?.map(l => l.id) ?? [];
      dueDate = card.dueDate ? card.dueDate.split('T')[0] : '';
      error = null;
      // Set initial saved state
      lastSavedData = JSON.stringify({
        title: card.title,
        description: card.description ?? '',
        priority: card.priority,
        selectedLabelIds: card.labels?.map(l => l.id) ?? [],
        dueDate: card.dueDate ? card.dueDate.split('T')[0] : ''
      });
    }
  });

  // Auto-save effect
  $effect(() => {
    const currentData = getCurrentDataHash();
    if (card && lastSavedData && currentData !== lastSavedData && title.trim()) {
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
    if (!card || !title.trim() || saving) return;

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
        selectedLabelIds,
        dueDateRfc3339
      );
      lastSavedData = getCurrentDataHash();
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
    // Flush any pending auto-save
    if (saveTimeout) {
      clearTimeout(saveTimeout);
      await autoSave();
    }
    // Refresh the board to show updated card data
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

  function toggleLabel(labelId: string) {
    if (selectedLabelIds.includes(labelId)) {
      selectedLabelIds = selectedLabelIds.filter(id => id !== labelId);
    } else {
      selectedLabelIds = [...selectedLabelIds, labelId];
    }
  }

  function getRandomColor(): string {
    return presetColors[Math.floor(Math.random() * presetColors.length)];
  }

  async function handleLabelKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      e.stopPropagation();
      if (exactMatch) {
        if (!selectedLabelIds.includes(exactMatch.id)) {
          selectedLabelIds = [...selectedLabelIds, exactMatch.id];
        }
        labelSearch = '';
        showLabelDropdown = false;
      } else if (labelSearch.trim()) {
        await createNewLabel(labelSearch.trim());
      }
    } else if (e.key === 'Escape') {
      showLabelDropdown = false;
      labelSearch = '';
    }
  }

  async function createNewLabel(name: string) {
    if (!projectId) return;
    try {
      creatingLabel = true;
      const newLabel = await createLabel(projectId, name, getRandomColor());
      selectedLabelIds = [...selectedLabelIds, newLabel.id];
      labelSearch = '';
      showLabelDropdown = false;
      onLabelsChanged?.();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create label';
    } finally {
      creatingLabel = false;
    }
  }

  function selectLabel(labelId: string) {
    if (!selectedLabelIds.includes(labelId)) {
      selectedLabelIds = [...selectedLabelIds, labelId];
    }
    labelSearch = '';
    showLabelDropdown = false;
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
      <div class="px-5 py-4 border-b border-gray-200 flex items-center justify-between flex-shrink-0">
        <h2 class="text-lg font-semibold text-gray-900">Card Details</h2>
        <div class="flex items-center gap-1">
          <!-- View mode toggle -->
          <button
            type="button"
            class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
            onclick={() => onViewModeChange('modal')}
            title="Switch to modal view"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h12a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V6z" />
            </svg>
          </button>
          <button
            type="button"
            class="p-1 text-gray-400 hover:text-gray-600 rounded-md hover:bg-gray-100 transition-colors"
            onclick={handleClose}
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto px-5 py-4 space-y-4">
        {#if error}
          <div class="rounded-md bg-red-50 p-3">
            <p class="text-sm text-red-700">{error}</p>
          </div>
        {/if}

        <Input
          id="panel-title"
          label="Title"
          bind:value={title}
          required
        />

        <Textarea
          id="panel-description"
          label="Description"
          bind:value={description}
          rows={5}
          placeholder="Add a description"
        />

        <div class="grid grid-cols-2 gap-4">
          <Select id="panel-priority" label="Priority" bind:value={priority}>
            <option value={CardPriority.None}>None</option>
            <option value={CardPriority.Low}>Low</option>
            <option value={CardPriority.Medium}>Medium</option>
            <option value={CardPriority.High}>High</option>
            <option value={CardPriority.Urgent}>Urgent</option>
          </Select>

          <Input
            type="date"
            id="panel-dueDate"
            label="Due Date"
            bind:value={dueDate}
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">Labels</label>

          <!-- Selected labels -->
          {#if selectedLabelIds.length > 0}
            <div class="flex flex-wrap gap-1.5 mb-2">
              {#each labels.filter(l => selectedLabelIds.includes(l.id)) as label}
                <span
                  class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-sm font-medium"
                  style="background-color: {label.color}25; color: {label.color};"
                >
                  {label.name}
                  <button
                    type="button"
                    class="hover:bg-black/10 rounded-full p-0.5"
                    onclick={() => toggleLabel(label.id)}
                  >
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </span>
              {/each}
            </div>
          {/if}

          <!-- Label input with dropdown -->
          <div class="relative">
            <input
              type="text"
              bind:value={labelSearch}
              onfocus={() => showLabelDropdown = true}
              onkeydown={handleLabelKeydown}
              placeholder="Type to search or create labels..."
              disabled={creatingLabel}
              class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
            />

            {#if showLabelDropdown && (filteredLabels.length > 0 || (labelSearch.trim() && !exactMatch))}
              <div class="absolute z-10 mt-1 w-full bg-white border border-gray-200 rounded-md shadow-lg max-h-48 overflow-y-auto">
                {#each filteredLabels as label}
                  <button
                    type="button"
                    class="w-full px-3 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2 {selectedLabelIds.includes(label.id) ? 'bg-gray-50' : ''}"
                    onclick={() => selectLabel(label.id)}
                  >
                    <span
                      class="w-3 h-3 rounded-full flex-shrink-0"
                      style="background-color: {label.color};"
                    ></span>
                    <span class="flex-1">{label.name}</span>
                    {#if selectedLabelIds.includes(label.id)}
                      <svg class="w-4 h-4 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                      </svg>
                    {/if}
                  </button>
                {/each}

                {#if labelSearch.trim() && !exactMatch && projectId}
                  <button
                    type="button"
                    class="w-full px-3 py-2 text-left text-sm hover:bg-indigo-50 text-indigo-600 border-t border-gray-100 flex items-center gap-2"
                    onclick={() => createNewLabel(labelSearch.trim())}
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                    Create "{labelSearch.trim()}"
                  </button>
                {/if}
              </div>
            {/if}
          </div>
        </div>

        <div class="pt-4 border-t border-gray-200 text-xs text-gray-500">
          <p>Created: {formatDate(card.createdAt)}</p>
          {#if card.updatedAt !== card.createdAt}
            <p>Updated: {formatDate(card.updatedAt)}</p>
          {/if}
        </div>
      </div>

      <!-- Footer -->
      <div class="px-5 py-4 border-t border-gray-200 flex-shrink-0">
        <div class="flex items-center justify-between mb-3">
          <Button variant="danger" size="sm" onclick={handleDeleteClick} disabled={deleting || saving}>
            {deleting ? 'Deleting...' : 'Delete'}
          </Button>
          <div class="flex items-center gap-3">
            {#if saving}
              <span class="text-xs text-gray-400">Saving...</span>
            {:else if getCurrentDataHash() === lastSavedData}
              <span class="text-xs text-green-600">Saved</span>
            {/if}
            <Button variant="secondary" size="sm" onclick={handleClose} disabled={deleting}>
              Close
            </Button>
          </div>
        </div>
        <div class="text-xs text-gray-400 text-center">
          <kbd class="px-1.5 py-0.5 bg-gray-100 border border-gray-300 rounded text-gray-600">Esc</kbd> to close
          <span class="mx-2">·</span>
          Auto-saves as you type
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
