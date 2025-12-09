<script lang="ts">
  import { createLabel, updateLabel, type Label } from '../../lib/api/boards';

  interface Props {
    projectId: string;
    labels: Label[];
    selectedLabelIds: string[];
    onSelectionChange: (ids: string[]) => void;
    onLabelsChanged?: () => void;
    disabled?: boolean;
  }

  let { projectId, labels, selectedLabelIds, onSelectionChange, onLabelsChanged, disabled = false }: Props = $props();

  const presetColors = [
    '#ef4444', '#f97316', '#f59e0b', '#eab308',
    '#84cc16', '#22c55e', '#14b8a6', '#06b6d4',
    '#0ea5e9', '#3b82f6', '#6366f1', '#8b5cf6',
    '#a855f7', '#d946ef', '#ec4899', '#f43f5e',
  ];

  let labelSearch = $state('');
  let showLabelDropdown = $state(false);
  let creatingLabel = $state(false);
  let error = $state<string | null>(null);

  // Color picker state
  let showColorPicker = $state(false);
  let newLabelName = $state('');
  let selectedColor = $state(presetColors[0]);
  let editingLabel = $state<Label | null>(null);

  let filteredLabels = $derived(
    labelSearch.trim()
      ? labels.filter(l => l.name.toLowerCase().includes(labelSearch.toLowerCase()))
      : labels
  );

  let exactMatch = $derived(
    labels.find(l => l.name.toLowerCase() === labelSearch.trim().toLowerCase())
  );

  function toggleLabel(labelId: string) {
    if (selectedLabelIds.includes(labelId)) {
      onSelectionChange(selectedLabelIds.filter(id => id !== labelId));
    } else {
      onSelectionChange([...selectedLabelIds, labelId]);
    }
  }

  function selectLabel(labelId: string) {
    if (!selectedLabelIds.includes(labelId)) {
      onSelectionChange([...selectedLabelIds, labelId]);
    }
    labelSearch = '';
    showLabelDropdown = false;
  }

  function openColorPicker(name: string) {
    newLabelName = name;
    selectedColor = presetColors[Math.floor(Math.random() * presetColors.length)];
    editingLabel = null;
    showColorPicker = true;
    showLabelDropdown = false;
  }

  function openEditLabelColor(label: Label, e: Event) {
    e.stopPropagation();
    editingLabel = label;
    newLabelName = label.name;
    selectedColor = label.color;
    showColorPicker = true;
    showLabelDropdown = false;
  }

  async function saveLabel() {
    if (!newLabelName.trim()) return;
    try {
      creatingLabel = true;
      error = null;
      if (editingLabel) {
        await updateLabel(editingLabel.id, undefined, selectedColor);
        onLabelsChanged?.();
      } else {
        const newLabel = await createLabel(projectId, newLabelName.trim(), selectedColor);
        onSelectionChange([...selectedLabelIds, newLabel.id]);
        onLabelsChanged?.();
      }
      labelSearch = '';
      newLabelName = '';
      editingLabel = null;
      showColorPicker = false;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to save label';
    } finally {
      creatingLabel = false;
    }
  }

  function cancelColorPicker() {
    showColorPicker = false;
    newLabelName = '';
    editingLabel = null;
    labelSearch = '';
  }

  function handleLabelKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      e.stopPropagation();
      if (exactMatch) {
        if (!selectedLabelIds.includes(exactMatch.id)) {
          onSelectionChange([...selectedLabelIds, exactMatch.id]);
        }
        labelSearch = '';
        showLabelDropdown = false;
      } else if (labelSearch.trim()) {
        openColorPicker(labelSearch.trim());
      }
    } else if (e.key === 'Escape') {
      e.stopPropagation();
      if (showColorPicker) {
        cancelColorPicker();
      } else {
        showLabelDropdown = false;
        labelSearch = '';
      }
    }
  }
</script>

<div class="relative">
  <label class="block text-sm font-medium text-gray-700 mb-2">Labels</label>

  {#if error}
    <div class="rounded-md bg-red-50 p-2 mb-2">
      <p class="text-xs text-red-700">{error}</p>
    </div>
  {/if}

  <!-- Selected labels -->
  {#if selectedLabelIds.length > 0}
    <div class="flex flex-wrap gap-1.5 mb-2">
      {#each labels.filter(l => selectedLabelIds.includes(l.id)) as label}
        <span
          class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-sm font-medium cursor-pointer hover:opacity-80 transition-opacity"
          style="background-color: {label.color}25; color: {label.color};"
          onclick={(e) => openEditLabelColor(label, e)}
          title="Click to change color"
        >
          {label.name}
          <button
            type="button"
            class="hover:bg-black/10 rounded-full p-0.5"
            onclick={(e) => { e.stopPropagation(); toggleLabel(label.id); }}
            {disabled}
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
      disabled={disabled || creatingLabel}
      class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm disabled:bg-gray-50 disabled:text-gray-500"
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
            onclick={() => openColorPicker(labelSearch.trim())}
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

  <!-- Color Picker -->
  {#if showColorPicker}
    <div class="absolute z-20 mt-1 left-0 right-0 bg-white border border-gray-200 rounded-lg shadow-lg p-4">
      <div class="flex items-center justify-between mb-3">
        <span class="text-sm font-medium text-gray-700">
          {editingLabel ? 'Edit color for' : 'Choose color for'} "{newLabelName}"
        </span>
        <button
          type="button"
          class="text-gray-400 hover:text-gray-600"
          onclick={cancelColorPicker}
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Preview -->
      <div class="mb-3">
        <span
          class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium"
          style="background-color: {selectedColor}25; color: {selectedColor};"
        >
          {newLabelName}
        </span>
      </div>

      <!-- Color grid -->
      <div class="grid grid-cols-8 gap-2 mb-3">
        {#each presetColors as color}
          <button
            type="button"
            class="w-6 h-6 rounded-full border-2 transition-transform hover:scale-110 {selectedColor === color ? 'border-gray-800 ring-2 ring-offset-1 ring-gray-400' : 'border-transparent'}"
            style="background-color: {color};"
            onclick={() => selectedColor = color}
          ></button>
        {/each}
      </div>

      <!-- Actions -->
      <div class="flex justify-end gap-2">
        <button
          type="button"
          class="px-3 py-1.5 text-sm text-gray-600 hover:text-gray-800"
          onclick={cancelColorPicker}
        >
          Cancel
        </button>
        <button
          type="button"
          class="px-3 py-1.5 text-sm bg-indigo-600 text-white rounded-md hover:bg-indigo-700 disabled:opacity-50"
          onclick={saveLabel}
          disabled={creatingLabel}
        >
          {#if creatingLabel}
            {editingLabel ? 'Saving...' : 'Creating...'}
          {:else}
            {editingLabel ? 'Save Color' : 'Create Label'}
          {/if}
        </button>
      </div>
    </div>
  {/if}
</div>
