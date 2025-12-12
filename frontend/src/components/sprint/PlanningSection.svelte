<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    title: string;
    cardCount: number;
    storyPoints?: number;
    expanded?: boolean;
    variant?: 'active' | 'future' | 'backlog' | 'closed';
    badge?: string;
    onToggle?: (expanded: boolean) => void;
    headerAction?: Snippet;
    children: Snippet;
  }

  let {
    title,
    cardCount,
    storyPoints = 0,
    expanded = $bindable(false),
    variant = 'future',
    badge,
    onToggle,
    headerAction,
    children,
  }: Props = $props();

  function toggleExpanded() {
    expanded = !expanded;
    onToggle?.(expanded);
  }

  const variantStyles = {
    active: 'border-green-200 bg-green-50/50',
    future: 'border-gray-200 bg-white',
    backlog: 'border-amber-200 bg-amber-50/50',
    closed: 'border-gray-200 bg-gray-50/50',
  };

  const badgeStyles = {
    active: 'bg-green-100 text-green-800',
    future: 'bg-blue-100 text-blue-800',
    backlog: 'bg-amber-100 text-amber-800',
    closed: 'bg-gray-100 text-gray-600',
  };
</script>

<div class="border rounded-lg {variantStyles[variant]} overflow-hidden">
  <!-- Header -->
  <button
    type="button"
    onclick={toggleExpanded}
    class="w-full px-4 py-3 flex items-center justify-between gap-3 hover:bg-black/5 transition-colors"
  >
    <div class="flex items-center gap-3 min-w-0">
      <!-- Expand/Collapse icon -->
      <svg
        class="w-4 h-4 text-gray-500 flex-shrink-0 transition-transform {expanded ? 'rotate-90' : ''}"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
      </svg>

      <!-- Title -->
      <span class="font-medium text-gray-900 truncate">{title}</span>

      <!-- Badge (optional) -->
      {#if badge}
        <span class="px-2 py-0.5 text-xs font-medium rounded-full {badgeStyles[variant]}">
          {badge}
        </span>
      {/if}
    </div>

    <div class="flex items-center gap-4 flex-shrink-0">
      <!-- Stats -->
      <div class="flex items-center gap-3 text-sm text-gray-500">
        <span>{cardCount} card{cardCount === 1 ? '' : 's'}</span>
        {#if storyPoints > 0}
          <span class="text-gray-300">|</span>
          <span>{storyPoints} pts</span>
        {/if}
      </div>

      <!-- Header action slot (rendered outside of button click) -->
      {#if headerAction}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div onclick={(e) => e.stopPropagation()}>
          {@render headerAction()}
        </div>
      {/if}
    </div>
  </button>

  <!-- Content -->
  {#if expanded}
    <div class="border-t border-gray-200">
      {@render children()}
    </div>
  {/if}
</div>
