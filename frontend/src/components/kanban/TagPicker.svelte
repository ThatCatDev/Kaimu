<script lang="ts">
  import { createTag, updateTag, type Tag } from '../../lib/api/boards';

  interface Props {
    projectId: string;
    tags: Tag[];
    selectedTagIds: string[];
    onSelectionChange: (ids: string[]) => void;
    onTagsChanged?: () => void;
    disabled?: boolean;
    readOnly?: boolean;
  }

  let { projectId, tags, selectedTagIds, onSelectionChange, onTagsChanged, disabled = false, readOnly = false }: Props = $props();

  // Computed selected tags for readOnly display
  const selectedTags = $derived(tags.filter(t => selectedTagIds.includes(t.id)));

  const presetColors = [
    '#ef4444', '#f97316', '#f59e0b', '#eab308',
    '#84cc16', '#22c55e', '#14b8a6', '#06b6d4',
    '#0ea5e9', '#3b82f6', '#6366f1', '#8b5cf6',
    '#a855f7', '#d946ef', '#ec4899', '#f43f5e',
  ];

  let tagSearch = $state('');
  let showTagDropdown = $state(false);
  let creatingTag = $state(false);
  let error = $state<string | null>(null);

  // Color picker state
  let showColorPicker = $state(false);
  let newTagName = $state('');
  let selectedColor = $state(presetColors[0]);
  let editingTag = $state<Tag | null>(null);

  let filteredTags = $derived(
    tagSearch.trim()
      ? tags.filter(t => t.name.toLowerCase().includes(tagSearch.toLowerCase()))
      : tags
  );

  let exactMatch = $derived(
    tags.find(t => t.name.toLowerCase() === tagSearch.trim().toLowerCase())
  );

  function toggleTag(tagId: string) {
    if (selectedTagIds.includes(tagId)) {
      onSelectionChange(selectedTagIds.filter(id => id !== tagId));
    } else {
      onSelectionChange([...selectedTagIds, tagId]);
    }
  }

  function selectTag(tagId: string) {
    if (!selectedTagIds.includes(tagId)) {
      onSelectionChange([...selectedTagIds, tagId]);
    }
    tagSearch = '';
    showTagDropdown = false;
  }

  function openColorPicker(name: string) {
    newTagName = name;
    selectedColor = presetColors[Math.floor(Math.random() * presetColors.length)];
    editingTag = null;
    showColorPicker = true;
    showTagDropdown = false;
  }

  function openEditTagColor(tag: Tag, e: Event) {
    e.stopPropagation();
    editingTag = tag;
    newTagName = tag.name;
    selectedColor = tag.color;
    showColorPicker = true;
    showTagDropdown = false;
  }

  async function saveTag() {
    if (!newTagName.trim()) return;
    try {
      creatingTag = true;
      error = null;
      if (editingTag) {
        await updateTag(editingTag.id, undefined, selectedColor);
        onTagsChanged?.();
      } else {
        const newTag = await createTag(projectId, newTagName.trim(), selectedColor);
        onSelectionChange([...selectedTagIds, newTag.id]);
        onTagsChanged?.();
      }
      tagSearch = '';
      newTagName = '';
      editingTag = null;
      showColorPicker = false;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to save tag';
    } finally {
      creatingTag = false;
    }
  }

  function cancelColorPicker() {
    showColorPicker = false;
    newTagName = '';
    editingTag = null;
    tagSearch = '';
  }

  function handleTagKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      e.stopPropagation();
      if (exactMatch) {
        if (!selectedTagIds.includes(exactMatch.id)) {
          onSelectionChange([...selectedTagIds, exactMatch.id]);
        }
        tagSearch = '';
        showTagDropdown = false;
      } else if (tagSearch.trim()) {
        openColorPicker(tagSearch.trim());
      }
    } else if (e.key === 'Escape') {
      e.stopPropagation();
      if (showColorPicker) {
        cancelColorPicker();
      } else {
        showTagDropdown = false;
        tagSearch = '';
      }
    }
  }

  function handleBlur(e: FocusEvent) {
    // Delay to allow click events on dropdown items to fire first
    setTimeout(() => {
      const relatedTarget = e.relatedTarget as HTMLElement | null;
      if (!relatedTarget?.closest('.tag-picker-dropdown')) {
        showTagDropdown = false;
      }
    }, 150);
  }
</script>

<div class="relative">
  <label class="block text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">Tags</label>

  {#if readOnly}
    <!-- Read-only display: plain tags without X buttons -->
    {#if selectedTags.length > 0}
      <div class="flex flex-wrap gap-1.5">
        {#each selectedTags as tag}
          <span
            class="inline-flex items-center px-2 py-0.5 rounded-full text-sm font-medium"
            style="background-color: {tag.color}25; color: {tag.color};"
          >
            {tag.name}
          </span>
        {/each}
      </div>
    {:else}
      <p class="text-sm text-gray-900">—</p>
    {/if}
  {:else}
    {#if error}
      <div class="rounded-md bg-red-50 p-2 mb-2">
        <p class="text-xs text-red-700">{error}</p>
      </div>
    {/if}

    <!-- Selected tags -->
    {#if selectedTagIds.length > 0}
      <div class="flex flex-wrap gap-1.5 mb-2">
        {#each tags.filter(t => selectedTagIds.includes(t.id)) as tag}
          <span
            class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-sm font-medium cursor-pointer hover:opacity-80 transition-opacity"
            style="background-color: {tag.color}25; color: {tag.color};"
            onclick={(e) => openEditTagColor(tag, e)}
            title="Click to change color"
          >
            {tag.name}
            <button
              type="button"
              class="hover:bg-black/10 rounded-full p-0.5"
              onclick={(e) => { e.stopPropagation(); toggleTag(tag.id); }}
              {disabled}
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </span>
        {/each}
      </div>
    {/if}

    <!-- Tag input with dropdown -->
    <div class="relative">
    <input
      type="text"
      bind:value={tagSearch}
      onfocus={() => showTagDropdown = true}
      onblur={handleBlur}
      onkeydown={handleTagKeydown}
      placeholder="Type to search or create tags..."
      disabled={disabled || creatingTag}
      class="block w-full px-4 py-2 border border-gray-300 rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-sm disabled:bg-gray-50 disabled:text-gray-500"
    />

    {#if showTagDropdown && (filteredTags.length > 0 || (tagSearch.trim() && !exactMatch))}
      <div class="tag-picker-dropdown absolute z-10 mt-1 w-full bg-white border border-gray-200 rounded-md shadow-lg max-h-48 overflow-y-auto animate-in fade-in-0 zoom-in-95">
        {#each filteredTags as tag}
          <button
            type="button"
            class="w-full px-4 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2 {selectedTagIds.includes(tag.id) ? 'bg-gray-50' : ''}"
            onclick={() => selectTag(tag.id)}
          >
            <span
              class="w-4 h-4 rounded-full flex-shrink-0"
              style="background-color: {tag.color};"
            ></span>
            <span class="flex-1">{tag.name}</span>
            {#if selectedTagIds.includes(tag.id)}
              <svg class="w-4 h-4 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            {/if}
          </button>
        {/each}

        {#if tagSearch.trim() && !exactMatch && projectId}
          <button
            type="button"
            class="w-full px-4 py-2 text-left text-sm hover:bg-indigo-50 text-indigo-600 border-t border-gray-100 flex items-center gap-2"
            onclick={() => openColorPicker(tagSearch.trim())}
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Create "{tagSearch.trim()}"
          </button>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Color Picker -->
  {#if showColorPicker}
    <div class="absolute z-20 mt-1 left-0 right-0 bg-white border border-gray-200 rounded-lg shadow-lg p-4 animate-in fade-in-0 zoom-in-95">
      <div class="flex items-center justify-between mb-4">
        <span class="text-sm font-medium text-gray-700">
          {editingTag ? 'Edit color for' : 'Choose color for'} "{newTagName}"
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
      <div class="mb-4">
        <span
          class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium"
          style="background-color: {selectedColor}25; color: {selectedColor};"
        >
          {newTagName}
        </span>
      </div>

      <!-- Color grid -->
      <div class="grid grid-cols-8 gap-2 mb-4">
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
          onclick={saveTag}
          disabled={creatingTag}
        >
          {#if creatingTag}
            {editingTag ? 'Saving...' : 'Creating...'}
          {:else}
            {editingTag ? 'Save Color' : 'Create Tag'}
          {/if}
        </button>
      </div>
    </div>
  {/if}
  {/if}
</div>

