<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    title: string;
    cardCount?: number;
    expanded?: boolean;
    children: Snippet;
  }

  let { title, cardCount = 0, expanded = false, children }: Props = $props();

  let isExpanded = $state(expanded);
</script>

<div class="border-b border-gray-200">
  <button
    type="button"
    onclick={() => isExpanded = !isExpanded}
    class="w-full p-3 flex items-center justify-between hover:bg-gray-50 transition-colors"
  >
    <div class="flex items-center gap-2">
      <svg
        class="w-4 h-4 text-gray-400 transition-transform {isExpanded ? 'rotate-90' : ''}"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>
      <span class="text-sm font-medium text-gray-700">{title}</span>
    </div>
    {#if cardCount > 0}
      <span class="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full">{cardCount}</span>
    {/if}
  </button>

  {#if isExpanded}
    <div class="bg-gray-50">
      {@render children()}
    </div>
  {/if}
</div>
