<script lang="ts">
  import { updateCard, deleteCard, createLabel, type BoardCard, type Label } from '../../lib/api/boards';
  import { CardPriority } from '../../lib/graphql/generated';
  import { Input, Textarea, Select, Button, ConfirmModal } from '../ui';

  interface Props {
    card: BoardCard;
    projectId: string;
    labels: Label[];
    onClose: () => void;
    onUpdated: () => void;
    onLabelsChanged?: () => void;
    viewMode: 'modal' | 'panel';
    onViewModeChange: (mode: 'modal' | 'panel') => void;
  }

  let { card, projectId, labels, onClose, onUpdated, onLabelsChanged, viewMode, onViewModeChange }: Props = $props();

  let title = $state(card.title);
  let description = $state(card.description ?? '');
  let priority = $state<CardPriority>(card.priority);
  let selectedLabelIds = $state<string[]>(card.labels?.map(l => l.id) ?? []);
  let dueDate = $state(card.dueDate ? card.dueDate.split('T')[0] : '');
  let saving = $state(false);
  let deleting = $state(false);
  let error = $state<string | null>(null);
  let saveTimeout: ReturnType<typeof setTimeout> | null = null;
  let showDeleteConfirm = $state(false);
  let lastSavedData = $state<string>(JSON.stringify({
    title: card.title,
    description: card.description ?? '',
    priority: card.priority,
    selectedLabelIds: card.labels?.map(l => l.id) ?? [],
    dueDate: card.dueDate ? card.dueDate.split('T')[0] : ''
  }));

  function getCurrentDataHash(): string {
    return JSON.stringify({ title, description, priority, selectedLabelIds, dueDate });
  }

  // Auto-save effect
  $effect(() => {
    const currentData = getCurrentDataHash();
    if (lastSavedData && currentData !== lastSavedData && title.trim()) {
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
    if (!title.trim() || saving) return;

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
    if (e.key === 'Escape') {
      handleClose();
    }
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      handleSubmit(e);
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      handleClose();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- Backdrop -->
<div
  class="fixed inset-0 bg-gray-900/60 backdrop-blur-sm z-50 animate-fade-in"
  onclick={handleBackdropClick}
  role="dialog"
  aria-modal="true"
>
  <!-- Modal -->
  <div class="fixed inset-0 flex items-center justify-center p-4 pointer-events-none">
    <div
      class="bg-white rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto pointer-events-auto animate-scale-in"
      onclick={(e) => e.stopPropagation()}
    >
    <form onsubmit={handleSubmit}>
      <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
        <h2 class="text-lg font-semibold text-gray-900">Card Details</h2>
        <div class="flex items-center gap-2">
          <!-- View mode toggle -->
          <button
            type="button"
            class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
            onclick={() => onViewModeChange('panel')}
            title="Switch to side panel view"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3h12v18H9M3 9h6M3 15h6" />
            </svg>
          </button>
          <button
            type="button"
            class="text-gray-400 hover:text-gray-600"
            onclick={handleClose}
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <div class="px-6 py-4 space-y-4">
        {#if error}
          <div class="rounded-md bg-red-50 p-3">
            <p class="text-sm text-red-700">{error}</p>
          </div>
        {/if}

        <Input
          id="title"
          label="Title"
          bind:value={title}
          required
        />

        <Textarea
          id="description"
          label="Description"
          bind:value={description}
          rows={4}
          placeholder="Add a description"
        />

        <div class="grid grid-cols-2 gap-4">
          <Select id="priority" label="Priority" bind:value={priority}>
            <option value={CardPriority.None}>None</option>
            <option value={CardPriority.Low}>Low</option>
            <option value={CardPriority.Medium}>Medium</option>
            <option value={CardPriority.High}>High</option>
            <option value={CardPriority.Urgent}>Urgent</option>
          </Select>

          <Input
            type="date"
            id="dueDate"
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

                {#if labelSearch.trim() && !exactMatch}
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

      <div class="px-6 py-4 border-t border-gray-200">
        <div class="flex items-center justify-between mb-2">
          <Button variant="danger" onclick={handleDeleteClick} disabled={deleting || saving}>
            {deleting ? 'Deleting...' : 'Delete Card'}
          </Button>
          <div class="flex items-center gap-3">
            {#if saving}
              <span class="text-xs text-gray-400">Saving...</span>
            {:else if getCurrentDataHash() === lastSavedData}
              <span class="text-xs text-green-600">Saved</span>
            {/if}
            <Button variant="secondary" onclick={handleClose} disabled={deleting}>
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
    </div>
  </div>
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

<style>
  @keyframes fade-in {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes scale-in {
    from {
      opacity: 0;
      transform: scale(0.95) translateY(10px);
    }
    to {
      opacity: 1;
      transform: scale(1) translateY(0);
    }
  }

  :global(.animate-fade-in) {
    animation: fade-in 0.15s ease-out;
  }

  :global(.animate-scale-in) {
    animation: scale-in 0.2s ease-out;
  }
</style>
