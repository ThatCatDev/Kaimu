<script lang="ts">
  import { createCard, type Tag } from '../../lib/api/boards';
  import { CardPriority } from '../../lib/graphql/generated';
  import { Button } from '../ui';
  import CardForm from './CardForm.svelte';

  interface Props {
    columnId: string;
    projectId: string;
    tags: Tag[];
    onClose: () => void;
    onCreated: () => void;
    onTagsChanged?: () => void;
  }

  let { columnId, projectId, tags, onClose, onCreated, onTagsChanged }: Props = $props();

  let title = $state('');
  let description = $state('');
  let priority = $state<CardPriority>(CardPriority.None);
  let selectedTagIds = $state<string[]>([]);
  let dueDate = $state('');
  let loading = $state(false);
  let error = $state<string | null>(null);

  async function handleSubmit(e: Event) {
    e.preventDefault();

    if (!title.trim()) {
      error = 'Title is required';
      return;
    }

    try {
      loading = true;
      error = null;
      const dueDateRfc3339 = dueDate ? new Date(dueDate + 'T00:00:00Z').toISOString() : undefined;
      await createCard(
        columnId,
        title.trim(),
        description.trim() || undefined,
        priority,
        undefined,
        selectedTagIds.length > 0 ? selectedTagIds : undefined,
        dueDateRfc3339
      );
      onCreated();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create card';
    } finally {
      loading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onClose();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- Backdrop -->
<div class="fixed inset-0 bg-gray-900/60 backdrop-blur-sm z-50 animate-fade-in">
  <!-- Modal -->
  <div class="fixed inset-0 flex items-center justify-center p-4">
    <div class="bg-white rounded-xl shadow-2xl max-w-lg w-full max-h-[90vh] overflow-y-auto animate-scale-in">
      <form onsubmit={handleSubmit}>
        <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900">Create Card</h2>
          <button
            type="button"
            class="text-gray-400 hover:text-gray-600 transition-colors"
            onclick={onClose}
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

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
            disabled={loading}
          />
        </div>

        <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
          <Button variant="secondary" onclick={onClose} disabled={loading}>
            Cancel
          </Button>
          <Button type="submit" {loading}>
            {loading ? 'Creating...' : 'Create Card'}
          </Button>
        </div>
      </form>
    </div>
  </div>
</div>

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
