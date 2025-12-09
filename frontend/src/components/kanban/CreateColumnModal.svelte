<script lang="ts">
  import { createColumn } from '../../lib/api/boards';
  import { Input, Button, Modal } from '../ui';

  interface Props {
    open: boolean;
    boardId: string;
    onClose: () => void;
    onCreated: () => void;
  }

  let { open, boardId, onClose, onCreated }: Props = $props();

  let name = $state('');
  let color = $state('');
  let wipLimit = $state<number | undefined>(undefined);
  let isBacklog = $state(false);
  let loading = $state(false);
  let error = $state<string | null>(null);

  const presetColors = [
    '#ef4444', '#f97316', '#f59e0b', '#eab308',
    '#84cc16', '#22c55e', '#14b8a6', '#06b6d4',
    '#0ea5e9', '#3b82f6', '#6366f1', '#8b5cf6',
  ];

  // Reset form when modal opens
  $effect(() => {
    if (open) {
      name = '';
      color = '';
      wipLimit = undefined;
      isBacklog = false;
      error = null;
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!name.trim()) return;

    try {
      loading = true;
      error = null;
      await createColumn(boardId, name.trim(), isBacklog);
      onCreated();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create column';
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

<Modal {open} onOpenChange={handleOpenChange} title="Add Column" size="md">
  {#snippet children()}
    <form id="create-column-form" onsubmit={handleSubmit}>
      <div class="px-6 py-4 space-y-4">
        {#if error}
          <div class="rounded-md bg-red-50 p-3">
            <p class="text-sm text-red-700">{error}</p>
          </div>
        {/if}

        <Input
          id="column-name"
          label="Column Name"
          bind:value={name}
          placeholder="e.g., In Review"
          required
          disabled={loading}
        />

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">Color (optional)</label>
          <div class="flex flex-wrap gap-2">
            <button
              type="button"
              class="w-8 h-8 rounded-full border-2 transition-transform hover:scale-110 {!color ? 'border-gray-400 ring-2 ring-offset-1 ring-gray-300' : 'border-gray-200'}"
              style="background-color: #f3f4f6;"
              onclick={() => color = ''}
              title="No color"
            >
              {#if !color}
                <svg class="w-4 h-4 mx-auto text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              {/if}
            </button>
            {#each presetColors as presetColor}
              <button
                type="button"
                class="w-8 h-8 rounded-full border-2 transition-transform hover:scale-110 {color === presetColor ? 'border-gray-800 ring-2 ring-offset-1 ring-gray-400' : 'border-transparent'}"
                style="background-color: {presetColor};"
                onclick={() => color = presetColor}
              ></button>
            {/each}
          </div>
        </div>

        <div>
          <label for="wip-limit" class="block text-sm font-medium text-gray-700 mb-1">
            WIP Limit (optional)
          </label>
          <input
            id="wip-limit"
            type="number"
            min="1"
            bind:value={wipLimit}
            placeholder="No limit"
            disabled={loading}
            class="block w-24 px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm disabled:bg-gray-50 disabled:text-gray-500"
          />
          <p class="mt-1 text-xs text-gray-500">Maximum number of cards in this column</p>
        </div>

        <label class="flex items-center gap-2 text-sm text-gray-700">
          <input
            type="checkbox"
            bind:checked={isBacklog}
            disabled={loading}
            class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
          />
          Mark as backlog column
        </label>
      </div>
    </form>
  {/snippet}

  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <Button type="button" variant="secondary" onclick={onClose} disabled={loading}>
        Cancel
      </Button>
      <Button type="submit" form="create-column-form" disabled={loading || !name.trim()}>
        {loading ? 'Creating...' : 'Create Column'}
      </Button>
    </div>
  {/snippet}
</Modal>
