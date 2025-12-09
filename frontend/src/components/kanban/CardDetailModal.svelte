<script lang="ts">
  import { updateCard, deleteCard, type BoardCard, type Label } from '../../lib/api/boards';
  import { CardPriority } from '../../lib/graphql/generated';
  import { Button, ConfirmModal } from '../ui';
  import CardForm from './CardForm.svelte';

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

      <div class="px-6 py-4">
        <CardForm
          {title}
          {description}
          {priority}
          {dueDate}
          {selectedLabelIds}
          {projectId}
          {labels}
          onTitleChange={(v) => title = v}
          onDescriptionChange={(v) => description = v}
          onPriorityChange={(v) => priority = v}
          onDueDateChange={(v) => dueDate = v}
          onLabelSelectionChange={(ids) => selectedLabelIds = ids}
          {onLabelsChanged}
          {error}
          disabled={saving || deleting}
          descriptionRows={4}
        />

        <div class="pt-4 mt-4 border-t border-gray-200 text-xs text-gray-500">
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
