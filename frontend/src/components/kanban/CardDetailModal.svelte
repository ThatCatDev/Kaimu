<script lang="ts">
  import { updateCard, deleteCard, type BoardCard, type Tag } from '../../lib/api/boards';
  import { CardPriority } from '../../lib/graphql/generated';
  import { Button, Modal, ConfirmModal } from '../ui';
  import CardForm from './CardForm.svelte';

  interface Props {
    open: boolean;
    card: BoardCard | null;
    projectId: string;
    tags: Tag[];
    onClose: () => void;
    onUpdated: () => void;
    onTagsChanged?: () => void;
    viewMode: 'modal' | 'panel';
    onViewModeChange: (mode: 'modal' | 'panel') => void;
  }

  let { open, card, projectId, tags, onClose, onUpdated, onTagsChanged, viewMode, onViewModeChange }: Props = $props();

  let title = $state('');
  let description = $state('');
  let priority = $state<CardPriority>(CardPriority.None);
  let selectedTagIds = $state<string[]>([]);
  let dueDate = $state('');
  let saving = $state(false);
  let deleting = $state(false);
  let error = $state<string | null>(null);
  let saveTimeout: ReturnType<typeof setTimeout> | null = null;
  let showDeleteConfirm = $state(false);
  let lastSavedData = $state<string>('');

  // Initialize form when card changes or modal opens
  $effect(() => {
    if (open && card) {
      title = card.title;
      description = card.description ?? '';
      priority = card.priority;
      selectedTagIds = card.tags?.map(t => t.id) ?? [];
      dueDate = card.dueDate ? card.dueDate.split('T')[0] : '';
      error = null;
      lastSavedData = JSON.stringify({
        title: card.title,
        description: card.description ?? '',
        priority: card.priority,
        selectedTagIds: card.tags?.map(t => t.id) ?? [],
        dueDate: card.dueDate ? card.dueDate.split('T')[0] : ''
      });
    }
  });

  function getCurrentDataHash(): string {
    return JSON.stringify({ title, description, priority, selectedTagIds, dueDate });
  }

  // Auto-save effect
  $effect(() => {
    if (!open || !card) return;

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
        selectedTagIds,
        dueDateRfc3339
      );
      lastSavedData = getCurrentDataHash();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to save';
    } finally {
      saving = false;
    }
  }

  async function handleClose() {
    if (saveTimeout) {
      clearTimeout(saveTimeout);
      await autoSave();
    }
    onUpdated();
  }

  async function handleOpenChange(newOpen: boolean) {
    if (!newOpen) {
      await handleClose();
    }
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
</script>

<Modal {open} onOpenChange={handleOpenChange} title="Card Details" size="2xl">
  {#snippet headerActions()}
    <button
      type="button"
      class="p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
      onclick={() => onViewModeChange('panel')}
      title="Switch to side panel view"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3h12v18H9M3 9h6M3 15h6" />
      </svg>
    </button>
  {/snippet}

  {#snippet children()}
    {#if card}
      <div class="px-6 py-4">
        <CardForm
          {title}
          {description}
          {priority}
          {dueDate}
          {selectedTagIds}
          {projectId}
          {tags}
          onTitleChange={(v) => title = v}
          onDescriptionChange={(v) => description = v}
          onPriorityChange={(v) => priority = v}
          onDueDateChange={(v) => dueDate = v}
          onTagSelectionChange={(ids) => selectedTagIds = ids}
          {onTagsChanged}
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
    {/if}
  {/snippet}

  {#snippet footer()}
    <div class="flex items-center justify-between">
      <Button variant="danger" onclick={handleDeleteClick} disabled={deleting || saving}>
        {deleting ? 'Deleting...' : 'Delete Card'}
      </Button>
      <div class="flex items-center gap-4">
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
    <div class="text-xs text-gray-400 text-center mt-3">
      <kbd class="px-1.5 py-0.5 bg-gray-100 border border-gray-300 rounded text-gray-600">Esc</kbd> to close
      <span class="mx-2">·</span>
      Auto-saves as you type
    </div>
  {/snippet}
</Modal>

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
