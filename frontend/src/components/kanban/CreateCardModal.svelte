<script lang="ts">
  import { createCard, createLabel, type Label } from '../../lib/api/boards';
  import { CardPriority } from '../../lib/graphql/generated';
  import { Input, Textarea, Select, Button } from '../ui';

  interface Props {
    columnId: string;
    projectId: string;
    labels: Label[];
    onClose: () => void;
    onCreated: () => void;
    onLabelsChanged?: () => void;
  }

  let { columnId, projectId, labels, onClose, onCreated, onLabelsChanged }: Props = $props();

  let title = $state('');
  let description = $state('');
  let priority = $state<CardPriority>(CardPriority.None);
  let selectedLabelIds = $state<string[]>([]);
  let dueDate = $state('');
  let loading = $state(false);
  let error = $state<string | null>(null);

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

    if (!title.trim()) {
      error = 'Title is required';
      return;
    }

    try {
      loading = true;
      error = null;
      // Convert date input (YYYY-MM-DD) to RFC3339 format for the API
      const dueDateRfc3339 = dueDate ? new Date(dueDate + 'T00:00:00Z').toISOString() : undefined;
      await createCard(
        columnId,
        title.trim(),
        description.trim() || undefined,
        priority,
        undefined, // assigneeId
        selectedLabelIds.length > 0 ? selectedLabelIds : undefined,
        dueDateRfc3339
      );
      onCreated();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create card';
    } finally {
      loading = false;
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
        // Select existing label
        if (!selectedLabelIds.includes(exactMatch.id)) {
          selectedLabelIds = [...selectedLabelIds, exactMatch.id];
        }
        labelSearch = '';
        showLabelDropdown = false;
      } else if (labelSearch.trim()) {
        // Create new label
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

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      // If label dropdown is open, just close it first
      if (showLabelDropdown) {
        showLabelDropdown = false;
        labelSearch = '';
      } else {
        // Otherwise close the modal
        onClose();
      }
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
          placeholder="Enter card title"
          required
        />

        <Textarea
          id="description"
          label="Description"
          bind:value={description}
          rows={3}
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
