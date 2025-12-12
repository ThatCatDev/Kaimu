<script lang="ts">
  import { Combobox } from 'bits-ui';
  import { search, SearchEntityType, type SearchResult } from '../../lib/api/search';

  interface Props {
    value: string | null;
    organizationId: string;
    initialUserName?: string | null;
    placeholder?: string;
    disabled?: boolean;
    readOnly?: boolean;
    onValueChange?: (userId: string | null, displayName?: string) => void;
  }

  let {
    value = $bindable(null),
    organizationId,
    initialUserName = null,
    placeholder = 'Search for a user...',
    disabled = false,
    readOnly = false,
    onValueChange
  }: Props = $props();

  let inputValue = $state('');
  let searchResults = $state<SearchResult[]>([]);
  let loading = $state(false);
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;
  let open = $state(false);
  let inputRef = $state<HTMLInputElement | null>(null);

  // Selected user display info
  let selectedUserName = $state<string | null>(initialUserName);
  let hasUserSelected = $state(false); // Track if user made a selection in this session

  // Only use initialUserName on mount or when it changes from parent (card changed)
  $effect(() => {
    // If user hasn't selected anything yet, use the initial value from props
    if (!hasUserSelected && initialUserName !== selectedUserName) {
      selectedUserName = initialUserName;
    }
  });

  // Reset when value becomes null (unassigned)
  $effect(() => {
    if (value === null) {
      selectedUserName = null;
      hasUserSelected = false;
    }
  });

  // Debounced search
  $effect(() => {
    // Track inputValue explicitly
    const query = inputValue;

    if (searchTimeout) clearTimeout(searchTimeout);

    if (query.trim().length < 1) {
      searchResults = [];
      loading = false;
      return;
    }

    loading = true;
    searchTimeout = setTimeout(async () => {
      try {
        const results = await search(query, { organizationId }, 20);
        // Filter to only user results
        searchResults = results.results.filter(r => r.type === SearchEntityType.User);
      } catch (e) {
        console.error('Search error:', e);
        searchResults = [];
      } finally {
        loading = false;
      }
    }, 200);
  });

  function handleSelect(selected: string) {
    if (selected === '__unassigned__') {
      value = null;
      selectedUserName = null;
      hasUserSelected = true;
      inputValue = '';
      searchResults = [];
      if (inputRef) inputRef.value = '';
      onValueChange?.(null);
    } else {
      const user = searchResults.find(r => r.id === selected);
      // Backend already returns displayName (with username fallback) as title
      const displayName = user?.title || 'User';
      value = selected;
      selectedUserName = displayName;
      hasUserSelected = true;
      inputValue = '';
      searchResults = [];
      if (inputRef) inputRef.value = '';
      onValueChange?.(selected, displayName);
    }
    open = false;
  }

  function handleClear() {
    value = null;
    selectedUserName = null;
    hasUserSelected = true;
    inputValue = '';
    searchResults = [];
    if (inputRef) inputRef.value = '';
    onValueChange?.(null);
  }

  function handleInputFocus() {
    open = true;
  }
</script>

{#if readOnly}
  <p class="text-sm text-gray-900">
    {selectedUserName ?? 'Unassigned'}
  </p>
{:else}
  <div class="relative">
    <Combobox.Root
      type="single"
      {disabled}
      bind:open
      onValueChange={(v: string) => {
        handleSelect(v || '__unassigned__');
      }}
    >
      <div class="relative">
        <Combobox.Input
          bind:ref={inputRef}
          class="w-full px-3 py-2 pr-16 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 disabled:bg-gray-50 disabled:text-gray-500"
          placeholder={value ? (selectedUserName ?? 'Selected user') : placeholder}
          oninput={(e: Event & { currentTarget: HTMLInputElement }) => { inputValue = e.currentTarget.value; }}
          onfocus={handleInputFocus}
        />
        <div class="absolute inset-y-0 right-0 flex items-center pr-2 gap-1">
          {#if value}
            <button
              type="button"
              class="p-1 text-gray-400 hover:text-gray-600 rounded"
              onclick={handleClear}
              title="Clear assignee"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          {/if}
          <Combobox.Trigger class="p-1 text-gray-400 hover:text-gray-600">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </Combobox.Trigger>
        </div>
      </div>

      <Combobox.Portal>
        <Combobox.Content
          class="z-50 w-[var(--bits-combobox-anchor-width)] rounded-md border border-gray-200 bg-white shadow-lg animate-in fade-in-0 zoom-in-95"
          sideOffset={4}
        >
          <Combobox.Viewport class="p-1 max-h-60 overflow-y-auto">
            {#if loading}
              <div class="px-3 py-2 text-sm text-gray-500">Searching...</div>
            {:else if inputValue.trim().length === 0}
              <div class="px-3 py-2 text-sm text-gray-500">Type to search users...</div>
              {#if value}
                <Combobox.Item
                  value="__unassigned__"
                  label="Unassigned"
                  class="relative flex w-full cursor-pointer select-none items-center rounded-md px-3 py-2 text-sm text-gray-700 outline-none transition-colors data-[highlighted]:bg-red-50 data-[highlighted]:text-red-700"
                >
                  <svg class="w-4 h-4 mr-2 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  Remove assignee
                </Combobox.Item>
              {/if}
            {:else if searchResults.length === 0}
              <div class="px-3 py-2 text-sm text-gray-500">No users found</div>
            {:else}
              {#if value}
                <Combobox.Item
                  value="__unassigned__"
                  label="Unassigned"
                  class="relative flex w-full cursor-pointer select-none items-center rounded-md px-3 py-2 text-sm text-gray-700 outline-none transition-colors data-[highlighted]:bg-red-50 data-[highlighted]:text-red-700 border-b border-gray-100 mb-1"
                >
                  <svg class="w-4 h-4 mr-2 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  Remove assignee
                </Combobox.Item>
              {/if}
              {#each searchResults as user (user.id)}
                {@const displayName = user.title || 'Unknown'}
                {@const email = user.description}
                <Combobox.Item
                  value={user.id}
                  label={displayName}
                  class="relative flex w-full cursor-pointer select-none items-center rounded-md px-3 py-2 text-sm text-gray-900 outline-none transition-colors data-[highlighted]:bg-indigo-50 data-[highlighted]:text-indigo-900 data-[selected]:bg-indigo-100"
                >
                  {#snippet children({ selected })}
                    <span class="inline-flex items-center justify-center w-6 h-6 rounded-full bg-indigo-100 text-indigo-600 text-xs font-medium mr-2">
                      {displayName.charAt(0).toUpperCase()}
                    </span>
                    <div class="flex-1 min-w-0">
                      <div class="truncate">{displayName}</div>
                      {#if email}
                        <div class="text-xs text-gray-500 truncate">{email}</div>
                      {/if}
                    </div>
                    {#if selected || user.id === value}
                      <svg class="h-4 w-4 text-indigo-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                      </svg>
                    {/if}
                  {/snippet}
                </Combobox.Item>
              {/each}
            {/if}
          </Combobox.Viewport>
        </Combobox.Content>
      </Combobox.Portal>
    </Combobox.Root>

    <!-- Show current selection below -->
    {#if value && selectedUserName}
      <div class="mt-2 flex items-center gap-2 text-sm text-gray-600">
        <span class="inline-flex items-center justify-center w-6 h-6 rounded-full bg-indigo-100 text-indigo-600 text-xs font-medium">
          {selectedUserName.charAt(0).toUpperCase()}
        </span>
        <span>{selectedUserName}</span>
      </div>
    {/if}
  </div>
{/if}
