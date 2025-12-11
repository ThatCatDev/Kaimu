<script lang="ts">
  import { updateColumn } from '../../lib/api/boards';
  import type { BoardColumn } from '../../lib/api/boards';
  import { Input, Button, Modal } from '../ui';

  type EditMode = 'rename' | 'color' | 'wipLimit';

  interface Props {
    open: boolean;
    column: BoardColumn | null;
    mode: EditMode;
    onClose: () => void;
    onUpdated: () => void;
  }

  let { open, column, mode, onClose, onUpdated }: Props = $props();

  let name = $state('');
  let color = $state('');
  let wipLimit = $state<number | undefined>(undefined);
  let loading = $state(false);
  let error = $state<string | null>(null);

  const presetColors = [
    '#ef4444', '#f97316', '#f59e0b', '#eab308',
    '#84cc16', '#22c55e', '#14b8a6', '#06b6d4',
    '#0ea5e9', '#3b82f6', '#6366f1', '#8b5cf6',
  ];

  const titles: Record<EditMode, string> = {
    rename: 'Rename Column',
    color: 'Change Column Color',
    wipLimit: 'Set WIP Limit',
  };

  // Initialize form when modal opens or column changes
  $effect(() => {
    if (open && column) {
      name = column.name;
      color = column.color ?? '';
      wipLimit = column.wipLimit ?? undefined;
      error = null;
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!column) return;

    try {
      loading = true;
      error = null;

      if (mode === 'rename') {
        if (!name.trim()) return;
        await updateColumn(column.id, name.trim(), undefined, undefined);
      } else if (mode === 'color') {
        await updateColumn(column.id, undefined, color || undefined, undefined);
      } else if (mode === 'wipLimit') {
        const shouldClear = wipLimit === undefined;
        await updateColumn(column.id, undefined, undefined, wipLimit, shouldClear);
      }

      onUpdated();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to update column';
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

<Modal {open} onOpenChange={handleOpenChange} title={titles[mode]} size="md">
  {#snippet children()}
    <form id="edit-column-form" onsubmit={handleSubmit}>
      <div class="px-6 py-4 space-y-4">
        {#if error}
          <div class="rounded-md bg-red-50 p-3">
            <p class="text-sm text-red-700">{error}</p>
          </div>
        {/if}

        {#if mode === 'rename'}
          <Input
            id="column-name"
            label="Column Name"
            bind:value={name}
            placeholder="e.g., In Review"
            required
            disabled={loading}
          />
        {:else if mode === 'color'}
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">Select Color</label>
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
            {#if color && column}
              <div class="mt-3 flex items-center gap-2">
                <span class="text-sm text-gray-600">Preview:</span>
                <span class="w-4 h-4 rounded-full" style="background-color: {color};"></span>
                <span class="font-medium text-gray-900">{column.name}</span>
              </div>
            {/if}
          </div>
        {:else if mode === 'wipLimit'}
          <div>
            <label for="wip-limit" class="block text-sm font-medium text-gray-700 mb-1">
              WIP Limit
            </label>
            <div class="flex items-center gap-2">
              <input
                id="wip-limit"
                type="number"
                min="1"
                bind:value={wipLimit}
                placeholder="No limit"
                disabled={loading}
                class="block w-32 px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm disabled:bg-gray-50 disabled:text-gray-500"
              />
              {#if wipLimit !== undefined}
                <button
                  type="button"
                  onclick={() => wipLimit = undefined}
                  disabled={loading}
                  class="px-3 py-2 text-sm text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-md transition-colors disabled:opacity-50"
                >
                  Clear
                </button>
              {/if}
            </div>
            <p class="mt-1 text-xs text-gray-500">
              Maximum number of cards allowed in this column. Leave empty for no limit.
            </p>
            {#if wipLimit && column}
              <p class="mt-2 text-sm text-gray-600">
                Current cards: <span class="font-medium">{column.cards.length}</span> / {wipLimit}
              </p>
            {/if}
          </div>
        {/if}
      </div>
    </form>
  {/snippet}

  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <Button type="button" variant="secondary" onclick={onClose} disabled={loading}>
        Cancel
      </Button>
      <Button
        type="submit"
        form="edit-column-form"
        disabled={loading || (mode === 'rename' && !name.trim())}
      >
        {loading ? 'Saving...' : 'Save'}
      </Button>
    </div>
  {/snippet}
</Modal>
