<script lang="ts">
  import { dndzone } from 'svelte-dnd-action';
  import KanbanCard from './KanbanCard.svelte';
  import ColumnSettingsMenu from './ColumnSettingsMenu.svelte';
  import type { BoardColumn, BoardCard } from '../../lib/api/boards';

  interface Props {
    column: BoardColumn;
    cards: BoardCard[];
    onCardMove: (cardId: string, columnId: string, afterCardId: string | null) => void;
    onCardClick?: (card: BoardCard) => void;
    onAddCard?: (columnId: string) => void;
    onRename?: () => void;
    onEditColor?: () => void;
    onEditWipLimit?: () => void;
    onToggleVisibility?: () => void;
    onDelete?: () => void;
    onQuickDelete?: (card: BoardCard) => void;
    priorityStyle?: 'border' | 'badge';
  }

  let {
    column,
    cards,
    onCardMove,
    onCardClick,
    onAddCard,
    onRename,
    onEditColor,
    onEditWipLimit,
    onToggleVisibility,
    onDelete,
    onQuickDelete,
    priorityStyle = 'badge'
  }: Props = $props();

  let items = $state(cards.map(card => ({ ...card, id: card.id })));

  // Sync items when cards prop changes
  $effect(() => {
    items = cards.map(card => ({ ...card, id: card.id }));
  });

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
</script>

<div class="flex-shrink-0 w-72 bg-gray-100 rounded-lg flex flex-col max-h-full {column.isHidden ? 'opacity-60' : ''}">
  <div class="p-4 flex items-center justify-between border-b border-gray-200">
    <div class="flex items-center gap-2 min-w-0 flex-1">
      {#if column.color}
        <span class="w-3 h-3 rounded-full flex-shrink-0" style="background-color: {column.color};"></span>
      {/if}
      <h3 class="font-medium text-gray-900 truncate">{column.name}</h3>
      <span class="text-sm text-gray-500 flex-shrink-0">({cards.length}{column.wipLimit ? `/${column.wipLimit}` : ''})</span>
      {#if column.isHidden}
        <span class="text-xs text-gray-400 flex-shrink-0">(hidden)</span>
      {/if}
    </div>
    <ColumnSettingsMenu
      {column}
      onRename={onRename ?? (() => {})}
      onEditColor={onEditColor ?? (() => {})}
      onEditWipLimit={onEditWipLimit ?? (() => {})}
      onToggleVisibility={onToggleVisibility ?? (() => {})}
      onDelete={onDelete ?? (() => {})}
    />
  </div>

  <div class="flex-1 flex flex-col min-h-0">
    <div
      class="flex-1 p-2 overflow-y-auto min-h-[60px]"
      use:dndzone={{
        items,
        flipDurationMs: 200,
        dropTargetStyle: { outline: '2px dashed #6366f1', outlineOffset: '-2px' },
        dragDisabled: false,
        type: 'cards',
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
