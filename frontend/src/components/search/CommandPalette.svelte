<script lang="ts">
  import { Dialog } from 'bits-ui';
  import { fly, fade } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { search, type SearchResult, type SearchScope, SearchEntityType } from '../../lib/api/search';

  interface Props {
    open: boolean;
    onOpenChange?: (open: boolean) => void;
    scope?: SearchScope;
  }

  let {
    open = $bindable(false),
    onOpenChange,
    scope
  }: Props = $props();

  let query = $state('');
  let results = $state<SearchResult[]>([]);
  let loading = $state(false);
  let selectedIndex = $state(0);
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;
  let inputEl: HTMLInputElement | undefined = $state();
  let activeFilter = $state<SearchEntityType | 'all'>('all');

  const filterOptions: { value: SearchEntityType | 'all'; label: string }[] = [
    { value: 'all', label: 'All' },
    { value: SearchEntityType.Card, label: 'Cards' },
    { value: SearchEntityType.Project, label: 'Projects' },
    { value: SearchEntityType.Board, label: 'Boards' },
    { value: SearchEntityType.Organization, label: 'Orgs' },
    { value: SearchEntityType.User, label: 'Users' },
  ];

  // Filtered results based on active filter
  let filteredResults = $derived(
    activeFilter === 'all'
      ? results
      : results.filter(r => r.type === activeFilter)
  );

  // Reset state when modal opens/closes
  $effect(() => {
    if (open) {
      query = '';
      results = [];
      selectedIndex = 0;
      activeFilter = 'all';
      // Focus input after a short delay
      setTimeout(() => inputEl?.focus(), 50);
    }
  });

  // Reset selected index when filter changes
  $effect(() => {
    activeFilter; // track this
    selectedIndex = 0;
  });

  // Debounced search
  $effect(() => {
    if (!open) return;

    if (searchTimeout) clearTimeout(searchTimeout);

    if (query.trim().length < 2) {
      results = [];
      return;
    }

    loading = true;
    searchTimeout = setTimeout(async () => {
      try {
        const searchResults = await search(query, scope, 10);
        results = searchResults.results;
        selectedIndex = 0;
      } catch (e) {
        console.error('Search error:', e);
        results = [];
      } finally {
        loading = false;
      }
    }, 200);
  });

  function handleOpenChange(newOpen: boolean) {
    open = newOpen;
    onOpenChange?.(newOpen);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, filteredResults.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, 0);
    } else if (e.key === 'Enter' && filteredResults.length > 0) {
      e.preventDefault();
      navigateToResult(filteredResults[selectedIndex]);
    }
  }

  function navigateToResult(result: SearchResult) {
    open = false;
    window.location.href = result.url;
  }

  function getEntityIcon(type: SearchEntityType): string {
    switch (type) {
      case SearchEntityType.Card:
        return 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2';
      case SearchEntityType.Project:
        return 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z';
      case SearchEntityType.Board:
        return 'M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2';
      case SearchEntityType.Organization:
        return 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4';
      case SearchEntityType.User:
        return 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z';
      default:
        return 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z';
    }
  }

  function getEntityLabel(type: SearchEntityType): string {
    switch (type) {
      case SearchEntityType.Card:
        return 'Card';
      case SearchEntityType.Project:
        return 'Project';
      case SearchEntityType.Board:
        return 'Board';
      case SearchEntityType.Organization:
        return 'Organization';
      case SearchEntityType.User:
        return 'User';
      default:
        return 'Unknown';
    }
  }

  function getEntityColor(type: SearchEntityType): string {
    switch (type) {
      case SearchEntityType.Card:
        return 'bg-blue-100 text-blue-700';
      case SearchEntityType.Project:
        return 'bg-purple-100 text-purple-700';
      case SearchEntityType.Board:
        return 'bg-green-100 text-green-700';
      case SearchEntityType.Organization:
        return 'bg-amber-100 text-amber-700';
      case SearchEntityType.User:
        return 'bg-pink-100 text-pink-700';
      default:
        return 'bg-gray-100 text-gray-700';
    }
  }
</script>

<Dialog.Root {open} onOpenChange={handleOpenChange}>
  <Dialog.Portal>
    <Dialog.Overlay forceMount>
      {#snippet child({ props, open: isOpen })}
        {#if isOpen}
          <div
            {...props}
            class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50"
            transition:fade={{ duration: 150 }}
          ></div>
        {/if}
      {/snippet}
    </Dialog.Overlay>

    <Dialog.Content forceMount>
      {#snippet child({ props, open: isOpen })}
        {#if isOpen}
          <div
            {...props}
            class="fixed inset-0 z-50 flex items-start justify-center pt-[15vh] p-4"
            transition:fade={{ duration: 150 }}
          >
            <div
              class="bg-white rounded-xl shadow-2xl w-full max-w-xl overflow-hidden"
              transition:fly={{ y: -10, duration: 200, easing: cubicOut }}
              onclick={(e) => e.stopPropagation()}
              onkeydown={handleKeydown}
            >
              <!-- Search Input -->
              <div class="flex items-center gap-3 px-4 py-4">
                <svg class="w-5 h-5 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <input
                  bind:this={inputEl}
                  type="text"
                  bind:value={query}
                  placeholder="Search cards, projects, boards..."
                  class="flex-1 bg-transparent border-none outline-none focus:outline-none focus:ring-0 text-gray-900 placeholder-gray-500 text-base"
                  style="box-shadow: none;"
                />
                {#if loading}
                  <svg class="w-5 h-5 text-gray-400 animate-spin flex-shrink-0" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                {:else}
                  <kbd class="hidden sm:inline-flex items-center px-2 py-1 text-xs font-medium text-gray-400 bg-gray-100 rounded border border-gray-200 flex-shrink-0">
                    Esc
                  </kbd>
                {/if}
              </div>

              <!-- Filter Bar -->
              <div class="flex items-center gap-1 px-4 py-2 border-y border-gray-100 bg-gray-50/50">
                {#each filterOptions as opt (opt.value)}
                  <button
                    type="button"
                    onclick={() => activeFilter = opt.value}
                    class="px-3 py-1 text-xs font-medium rounded-full transition-colors {activeFilter === opt.value ? 'bg-indigo-100 text-indigo-700' : 'text-gray-500 hover:bg-gray-100 hover:text-gray-700'}"
                  >
                    {opt.label}
                  </button>
                {/each}
              </div>

              <!-- Results -->
              <div class="max-h-80 overflow-y-auto">
                {#if filteredResults.length === 0 && query.length >= 2 && !loading}
                  <div class="px-4 py-8 text-center text-gray-500">
                    <svg class="w-12 h-12 mx-auto text-gray-300 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <p class="text-sm">No results found for "{query}"</p>
                  </div>
                {:else if filteredResults.length === 0 && query.length < 2}
                  <div class="px-4 py-8 text-center text-gray-500">
                    <p class="text-sm">Type at least 2 characters to search</p>
                  </div>
                {:else}
                  <div class="py-2">
                    {#each filteredResults as result, index (result.id)}
                      <button
                        type="button"
                        class="w-full px-4 py-2.5 flex items-start gap-3 text-left hover:bg-gray-50 transition-colors {index === selectedIndex ? 'bg-indigo-50' : ''}"
                        onclick={() => navigateToResult(result)}
                        onmouseenter={() => selectedIndex = index}
                      >
                        <div class="flex-shrink-0 mt-0.5">
                          <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={getEntityIcon(result.type)} />
                          </svg>
                        </div>
                        <div class="flex-1 min-w-0">
                          <div class="flex items-center gap-2">
                            <span class="font-medium text-gray-900 truncate">{result.title}</span>
                            <span class="text-xs px-1.5 py-0.5 rounded {getEntityColor(result.type)}">
                              {getEntityLabel(result.type)}
                            </span>
                          </div>
                          {#if result.description}
                            <p class="text-sm text-gray-500 truncate mt-0.5">{result.description}</p>
                          {/if}
                          <p class="text-xs text-gray-400 mt-1">
                            {result.organizationName}
                            {#if result.projectName}
                              <span class="mx-1">/</span>
                              {result.projectName}
                            {/if}
                          </p>
                        </div>
                        {#if index === selectedIndex}
                          <kbd class="hidden sm:flex items-center px-1.5 py-0.5 text-xs text-gray-400 bg-gray-100 rounded border border-gray-200 self-center">
                            Enter
                          </kbd>
                        {/if}
                      </button>
                    {/each}
                  </div>
                {/if}
              </div>

              <!-- Footer -->
              <div class="px-4 py-2 border-t border-gray-200 bg-gray-50 text-xs text-gray-500 flex items-center gap-4">
                <span class="flex items-center gap-1">
                  <kbd class="px-1 py-0.5 bg-white rounded border border-gray-200">↑↓</kbd>
                  navigate
                </span>
                <span class="flex items-center gap-1">
                  <kbd class="px-1 py-0.5 bg-white rounded border border-gray-200">Enter</kbd>
                  select
                </span>
                <span class="flex items-center gap-1">
                  <kbd class="px-1 py-0.5 bg-white rounded border border-gray-200">Esc</kbd>
                  close
                </span>
              </div>
            </div>
          </div>
        {/if}
      {/snippet}
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
