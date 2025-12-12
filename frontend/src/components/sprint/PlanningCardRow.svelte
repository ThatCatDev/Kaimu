<script lang="ts">
  import { DropdownMenu } from 'bits-ui';
  import type { SprintCard, SprintData } from '../../lib/api/sprints';
  import { CardPriority } from '../../lib/graphql/generated';

  interface Props {
    card: SprintCard;
    availableSprints?: SprintData[];
    onCardClick?: (card: SprintCard) => void;
    onMoveToSprint?: (cardId: string, sprintId: string) => void;
    onMoveToBacklog?: (cardId: string) => void;
  }

  let {
    card,
    availableSprints = [],
    onCardClick,
    onMoveToSprint,
    onMoveToBacklog,
  }: Props = $props();

  const priorityColors: Record<CardPriority, string> = {
    [CardPriority.Urgent]: 'bg-red-100 text-red-800',
    [CardPriority.High]: 'bg-orange-100 text-orange-800',
    [CardPriority.Medium]: 'bg-yellow-100 text-yellow-800',
    [CardPriority.Low]: 'bg-green-100 text-green-800',
    [CardPriority.None]: 'bg-gray-100 text-gray-600',
  };

  const priorityLabels: Record<CardPriority, string> = {
    [CardPriority.Urgent]: 'Urgent',
    [CardPriority.High]: 'High',
    [CardPriority.Medium]: 'Med',
    [CardPriority.Low]: 'Low',
    [CardPriority.None]: 'None',
  };

  function formatDueDate(dateStr: string | null | undefined): string {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    const now = new Date();
    const diffTime = date.getTime() - now.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays < 0) return 'Overdue';
    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Tomorrow';

    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function getDueDateClass(dateStr: string | null | undefined): string {
    if (!dateStr) return 'text-gray-400';
    const date = new Date(dateStr);
    const now = new Date();
    const diffTime = date.getTime() - now.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays < 0) return 'text-red-600 font-medium';
    if (diffDays <= 2) return 'text-orange-600';
    return 'text-gray-500';
  }

  function handleRowClick(e: MouseEvent) {
    // Don't trigger card click if clicking dropdown
    if ((e.target as HTMLElement).closest('[data-dropdown]')) return;
    onCardClick?.(card);
  }
</script>

<div
  role="button"
  tabindex="0"
  class="px-4 py-3 flex items-center gap-4 hover:bg-gray-50 cursor-pointer transition-colors group"
  onclick={handleRowClick}
  onkeydown={(e) => e.key === 'Enter' && onCardClick?.(card)}
>
  <!-- Title -->
  <div class="flex-1 min-w-0">
    <p class="text-sm font-medium text-gray-900 truncate" title={card.title}>
      {card.title}
    </p>
    {#if card.column}
      <p class="text-xs text-gray-400 mt-0.5">{card.column.name}</p>
    {/if}
  </div>

  <!-- Tags (show first 2) -->
  {#if card.tags && card.tags.length > 0}
    <div class="hidden sm:flex items-center gap-1 flex-shrink-0">
      {#each card.tags.slice(0, 2) as tag}
        <span
          class="px-1.5 py-0.5 text-xs font-medium rounded"
          style="background-color: {tag.color}20; color: {tag.color}"
        >
          {tag.name}
        </span>
      {/each}
      {#if card.tags.length > 2}
        <span class="text-xs text-gray-400">+{card.tags.length - 2}</span>
      {/if}
    </div>
  {/if}

  <!-- Assignee -->
  <div class="w-8 flex-shrink-0">
    {#if card.assignee}
      {#if card.assignee.avatarUrl}
        <img
          src={card.assignee.avatarUrl}
          alt={card.assignee.displayName || card.assignee.username}
          class="w-6 h-6 rounded-full"
          title={card.assignee.displayName || card.assignee.username}
        />
      {:else}
        <div
          class="w-6 h-6 rounded-full bg-indigo-100 text-indigo-700 flex items-center justify-center text-xs font-medium"
          title={card.assignee.displayName || card.assignee.username}
        >
          {(card.assignee.displayName || card.assignee.username).charAt(0).toUpperCase()}
        </div>
      {/if}
    {:else}
      <div class="w-6 h-6 rounded-full bg-gray-100 border border-dashed border-gray-300" title="Unassigned"></div>
    {/if}
  </div>

  <!-- Priority -->
  <span class="px-2 py-0.5 text-xs font-medium rounded-full flex-shrink-0 {priorityColors[card.priority]}">
    {priorityLabels[card.priority]}
  </span>

  <!-- Story Points -->
  <div class="w-12 text-right flex-shrink-0">
    {#if card.storyPoints != null}
      <span class="text-sm text-gray-600">{card.storyPoints} pts</span>
    {:else}
      <span class="text-sm text-gray-300">-</span>
    {/if}
  </div>

  <!-- Due Date -->
  <div class="w-20 text-right flex-shrink-0">
    {#if card.dueDate}
      <span class="text-sm {getDueDateClass(card.dueDate)}">{formatDueDate(card.dueDate)}</span>
    {:else}
      <span class="text-sm text-gray-300">-</span>
    {/if}
  </div>

  <!-- Actions Dropdown -->
  <div class="flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" data-dropdown>
    <DropdownMenu.Root>
      <DropdownMenu.Trigger
        class="p-1 rounded hover:bg-gray-200 text-gray-400 hover:text-gray-600"
        onclick={(e) => e.stopPropagation()}
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
        </svg>
      </DropdownMenu.Trigger>

      <DropdownMenu.Content
        class="z-50 min-w-[180px] bg-white border border-gray-200 rounded-lg shadow-lg py-1"
        side="bottom"
        align="end"
      >
        <DropdownMenu.Item
          class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer flex items-center gap-2"
          onclick={() => onCardClick?.(card)}
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
          View Details
        </DropdownMenu.Item>

        {#if availableSprints.length > 0}
          <DropdownMenu.Sub>
            <DropdownMenu.SubTrigger class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer flex items-center gap-2 justify-between w-full">
              <div class="flex items-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 9l3 3m0 0l-3 3m3-3H8m13 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Move to Sprint
              </div>
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </DropdownMenu.SubTrigger>
            <DropdownMenu.SubContent class="z-50 min-w-[160px] bg-white border border-gray-200 rounded-lg shadow-lg py-1">
              {#each availableSprints as sprint}
                <DropdownMenu.Item
                  class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer"
                  onclick={() => onMoveToSprint?.(card.id, sprint.id)}
                >
                  {sprint.name}
                </DropdownMenu.Item>
              {/each}
            </DropdownMenu.SubContent>
          </DropdownMenu.Sub>
        {/if}

        {#if onMoveToBacklog}
          <DropdownMenu.Item
            class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 cursor-pointer flex items-center gap-2"
            onclick={() => onMoveToBacklog?.(card.id)}
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            Move to Backlog
          </DropdownMenu.Item>
        {/if}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  </div>
</div>
