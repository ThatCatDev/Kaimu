<script lang="ts">
  import { dndzone } from 'svelte-dnd-action';
  import KanbanCard from './KanbanCard.svelte';
  import type { BoardColumn, BoardCard } from '../../lib/api/boards';

  interface Props {
    column: BoardColumn;
    cards: BoardCard[];
    onCardMove: (cardId: string, columnId: string, afterCardId: string | null) => void;
    onCardClick?: (card: BoardCard) => void;
    onAddCard?: (columnId: string) => void;
    onColumnSettings?: (column: BoardColumn) => void;
    onQuickDelete?: (card: BoardCard) => void;
    priorityStyle?: 'border' | 'badge';
  }

  let { column, cards, onCardMove, onCardClick, onAddCard, onColumnSettings, onQuickDelete, priorityStyle = 'badge' }: Props = $props();

  let dragDisabled = $state(true);
  let items = $derived(cards.map(card => ({ ...card, id: card.id })));

  function handleConsider(e: CustomEvent<{ items: BoardCard[] }>) {
    items = e.detail.items;
  }

  function handleFinalize(e: CustomEvent<{ items: BoardCard[] }>) {
    const newItems = e.detail.items;
    const movedCard = newItems.find((item, index) => {
      const originalIndex = cards.findIndex(c => c.id === item.id);
      return originalIndex !== index || cards[originalIndex]?.id !== newItems[index]?.id;
    });

    if (movedCard) {
      const newIndex = newItems.findIndex(item => item.id === movedCard.id);
      const afterCardId = newIndex > 0 ? newItems[newIndex - 1].id : null;
      onCardMove(movedCard.id, column.id, afterCardId);
    }

    items = newItems;
  }

  function handleAddCard() {
    if (onAddCard) {
      onAddCard(column.id);
    }
  }

  function handleColumnSettings() {
    if (onColumnSettings) {
      onColumnSettings(column);
    }
  }
</script>

<div class="flex-shrink-0 w-72 bg-gray-100 rounded-lg flex flex-col max-h-full">
  <div class="p-4 flex items-center justify-between border-b border-gray-200">
    <div class="flex items-center gap-2">
      {#if column.color}
        <span class="w-4 h-4 rounded-full" style="background-color: {column.color};"></span>
      {/if}
      <h3 class="font-medium text-gray-900">{column.name}</h3>
      <span class="text-sm text-gray-500">({cards.length})</span>
    </div>
    <button
      type="button"
      class="p-1 text-gray-400 hover:text-gray-600 rounded"
      onclick={handleColumnSettings}
      title="Column settings"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
      </svg>
    </button>
  </div>

  <div class="flex-1 flex flex-col min-h-0">
    <div
      class="flex-1 p-2 overflow-y-auto min-h-[60px]"
      use:dndzone={{
        items,
        flipDurationMs: 200,
        dropTargetStyle: { outline: '2px dashed #6366f1', outlineOffset: '-2px' },
        dragDisabled: false,
      }}
      onconsider={handleConsider}
      onfinalize={handleFinalize}
    >
      {#each items as card (card.id)}
        <div class="mb-2">
          <KanbanCard {card} {onCardClick} {onQuickDelete} {priorityStyle} />
        </div>
      {/each}
    </div>

    <!-- Add card button - sticky at bottom -->
    <div class="p-2 pt-0 sticky bottom-0 bg-gray-100">
      <button
        type="button"
        class="w-full py-2 px-4 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-200 rounded-lg flex items-center justify-center gap-1.5 transition-colors"
        onclick={handleAddCard}
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
        Add card
      </button>
    </div>
  </div>

  {#if column.wipLimit && cards.length >= column.wipLimit}
    <div class="px-4 py-2 bg-yellow-50 border-t border-yellow-200 text-xs text-yellow-700">
      WIP limit reached ({cards.length}/{column.wipLimit})
    </div>
  {/if}
</div>
