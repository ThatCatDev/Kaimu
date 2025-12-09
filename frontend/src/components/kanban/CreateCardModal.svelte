<script lang="ts">
  import { createCard, type Tag } from '../../lib/api/boards';
  import { CardPriority } from '../../lib/graphql/generated';
  import { Button, Modal } from '../ui';
  import CardForm from './CardForm.svelte';

  interface Props {
    open: boolean;
    columnId: string;
    projectId: string;
    tags: Tag[];
    onClose: () => void;
    onCreated: () => void;
    onTagsChanged?: () => void;
  }

  let { open, columnId, projectId, tags, onClose, onCreated, onTagsChanged }: Props = $props();

  let title = $state('');
  let description = $state('');
  let priority = $state<CardPriority>(CardPriority.None);
  let selectedTagIds = $state<string[]>([]);
  let dueDate = $state('');
  let loading = $state(false);
  let error = $state<string | null>(null);

  // Reset form when modal opens
  $effect(() => {
    if (open) {
      title = '';
      description = '';
      priority = CardPriority.None;
      selectedTagIds = [];
      dueDate = '';
      error = null;
    }
  });

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

  function handleOpenChange(newOpen: boolean) {
    if (!newOpen) {
      onClose();
    }
  }
</script>

<Modal {open} onOpenChange={handleOpenChange} title="Create Card" size="lg">
  {#snippet children()}
    <form id="create-card-form" onsubmit={handleSubmit}>
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
    </form>
  {/snippet}

  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <Button variant="secondary" onclick={onClose} disabled={loading}>
        Cancel
      </Button>
      <Button type="submit" form="create-card-form" {loading}>
        {loading ? 'Creating...' : 'Create Card'}
      </Button>
    </div>
  {/snippet}
</Modal>
