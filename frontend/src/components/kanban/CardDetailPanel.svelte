<script lang="ts">
  import { updateCard, deleteCard, type BoardCard, type Tag } from '../../lib/api/boards';
  import { CardPriority } from '../../lib/graphql/generated';
  import { Button, ConfirmModal } from '../ui';
  import CardForm from './CardForm.svelte';

  interface Props {
    card: BoardCard | null;
    projectId: string;
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

  let { card, projectId, tags, isOpen, onClose, onUpdated, onCardDataChanged, onTagsChanged, viewMode, onViewModeChange, canEditCard = true, canDeleteCard = true }: Props = $props();


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

  let currentCardId = $state<string | null>(null);
  let isEditing = $state(false); // Track if user has started editing

  // Use derived instead of function call in template to avoid re-render triggers
  const currentDataHash = $derived(JSON.stringify({ title, description, priority, selectedTagIds, dueDate }));
  const isSaved = $derived(currentDataHash === lastSavedData);

  // Reset state when panel closes
  $effect(() => {
    if (!isOpen) {
      currentCardId = null;
      isEditing = false;
    }
  });

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
