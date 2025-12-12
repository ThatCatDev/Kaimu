<script lang="ts">
  import type { BoardCard } from '../../lib/api/boards';
  import { CardPriority } from '../../lib/graphql/generated';
  import { ConfirmModal } from '../ui';

  interface Props {
    card: BoardCard;
    onCardClick?: (card: BoardCard) => void;
    onQuickDelete?: (card: BoardCard) => void;
    priorityStyle?: 'border' | 'badge';
    // Permission props
    canEditCard?: boolean;
    canDeleteCard?: boolean;
  }

  let {
    card,
    onCardClick,
    onQuickDelete,
    priorityStyle = 'badge',
    canEditCard = true,
    canDeleteCard = true
  }: Props = $props();
  let showDeleteConfirm = $state(false);


  const priorityColors: Record<CardPriority, string> = {
    [CardPriority.None]: '',
    [CardPriority.Low]: 'border-l-blue-400',
    [CardPriority.Medium]: 'border-l-yellow-400',
    [CardPriority.High]: 'border-l-orange-400',
    [CardPriority.Urgent]: 'border-l-red-500',
  };

  const priorityBadgeStyles: Record<CardPriority, { bg: string; text: string; label: string }> = {
    [CardPriority.None]: { bg: '', text: '', label: '' },
    [CardPriority.Low]: { bg: 'bg-blue-100', text: 'text-blue-700', label: 'Low' },
    [CardPriority.Medium]: { bg: 'bg-yellow-100', text: 'text-yellow-700', label: 'Medium' },
    [CardPriority.High]: { bg: 'bg-orange-100', text: 'text-orange-700', label: 'High' },
    [CardPriority.Urgent]: { bg: 'bg-red-100', text: 'text-red-700', label: 'Urgent' },
  };

  function formatDueDate(dateStr: string): string {
    // Parse date string - extract just the date part to avoid timezone issues
    const datePart = dateStr.split('T')[0];
    const [year, month, day] = datePart.split('-').map(Number);
    const date = new Date(year, month - 1, day); // Create in local time

    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const tomorrow = new Date(today);
    tomorrow.setDate(tomorrow.getDate() + 1);

    if (date.getTime() === today.getTime()) {
      return 'Today';
    }
    if (date.getTime() === tomorrow.getTime()) {
      return 'Tomorrow';
    }
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  }

  function isOverdue(dateStr: string): boolean {
    // Parse date string - extract just the date part to avoid timezone issues
    const datePart = dateStr.split('T')[0];
    const [year, month, day] = datePart.split('-').map(Number);
    const date = new Date(year, month - 1, day);
    date.setHours(0, 0, 0, 0);

    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return date < today;
  }

  function handleClick() {
    if (onCardClick) {
      onCardClick(card);
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick();
    }
  }

  function handleQuickDeleteClick(e: Event) {
    e.stopPropagation();
    showDeleteConfirm = true;
  }

  function confirmDelete() {
    showDeleteConfirm = false;
    if (onQuickDelete) {
      onQuickDelete(card);
    }
  }

  function handleEdit(e: Event) {
    e.stopPropagation();
    if (onCardClick) {
      onCardClick(card);
    }
  }
</script>

<div
  class="group relative w-full text-left rounded-lg shadow-sm border p-4 transition-shadow bg-white border-gray-200 hover:shadow-md cursor-pointer {priorityStyle === 'border' && card.priority !== CardPriority.None ? `border-l-4 ${priorityColors[card.priority]}` : ''}"
  onclick={handleClick}
  onkeydown={handleKeydown}
  role="button"
  tabindex="0"
>

  <!-- Quick actions - simple icons that appear on hover, permission-gated -->
  {#if canEditCard || canDeleteCard}
    <div class="absolute top-3 right-3 hidden group-hover:flex gap-1 z-10">
      {#if canEditCard}
        <button
          type="button"
          onclick={handleEdit}
          title="Edit card"
          class="p-1 text-gray-400 hover:text-indigo-600 rounded transition-colors"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
        </button>
      {/if}
      {#if canDeleteCard && onQuickDelete}
        <button
          type="button"
          onclick={handleQuickDeleteClick}
          title="Delete card"
          class="p-1 text-gray-400 hover:text-red-600 rounded transition-colors"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      {/if}
    </div>
  {/if}

  <h4 class="text-sm font-medium text-gray-900 mb-1 pr-16">{card.title}</h4>

  {#if card.tags && card.tags.length > 0}
    <div class="flex flex-wrap gap-1 mb-2">
      {#each card.tags as tag}
        <span
          class="inline-block px-2 py-0.5 text-xs font-medium rounded"
          style="background-color: {tag.color}20; color: {tag.color};"
        >
          {tag.name}
        </span>
      {/each}
    </div>
  {/if}

  <div class="flex items-center justify-between text-xs text-gray-500">
    <div class="flex items-center gap-2">
      {#if priorityStyle === 'badge' && card.priority !== CardPriority.None}
        <span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs font-medium {priorityBadgeStyles[card.priority].bg} {priorityBadgeStyles[card.priority].text}">
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M3 6a3 3 0 013-3h10a1 1 0 01.8 1.6L14.25 8l2.55 3.4A1 1 0 0116 13H6a1 1 0 00-1 1v3a1 1 0 11-2 0V6z" clip-rule="evenodd" />
          </svg>
          {priorityBadgeStyles[card.priority].label}
        </span>
      {/if}
      {#if card.dueDate}
        <span class="flex items-center gap-1 {isOverdue(card.dueDate) ? 'text-red-600' : ''}">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          {formatDueDate(card.dueDate)}
        </span>
      {/if}
    </div>

    {#if card.assignee}
      <span class="inline-flex items-center justify-center w-6 h-6 rounded-full bg-indigo-100 text-indigo-600 text-xs font-medium">
        {(card.assignee.displayName || card.assignee.username).charAt(0).toUpperCase()}
      </span>
    {/if}
  </div>
</div>

<ConfirmModal
  isOpen={showDeleteConfirm}
  title="Delete Card"
  message="Are you sure you want to delete this card? This action cannot be undone."
  confirmText="Delete"
  cancelText="Cancel"
  variant="danger"
  onConfirm={confirmDelete}
  onCancel={() => showDeleteConfirm = false}
/>
