<script lang="ts">
  import { dndzone } from 'svelte-dnd-action';
  import KanbanCard from './KanbanCard.svelte';
  import ColumnSettingsMenu from './ColumnSettingsMenu.svelte';
  import type { BoardColumn, BoardCard } from '../../lib/api/boards';

  interface Props {
    column: BoardColumn;
    cards: BoardCard[];
    onCardMove: (cardId: string, columnId: string, afterCardId: string | null) => void;
    onBulkCardMove?: (cardIds: string[], columnId: string) => void;
    onCardClick?: (card: BoardCard) => void;
    onAddCard?: (columnId: string) => void;
    onRename?: () => void;
    onEditColor?: () => void;
    onEditWipLimit?: () => void;
    onEditIsDone?: () => void;
    onToggleVisibility?: () => void;
    onDelete?: () => void;
    onQuickDelete?: (card: BoardCard) => void;
    priorityStyle?: 'border' | 'badge';
    // Permission props
    canManageBoard?: boolean;
    canEditCard?: boolean;
    canMoveCard?: boolean;
    canDeleteCard?: boolean;
    // Selection mode props
    isSelectionMode?: boolean;
    selectedCardIds?: Set<string>;
    onToggleSelect?: (card: BoardCard, shiftKey?: boolean) => void;
  }

  let {
    column,
    cards,
    onCardMove,
    onBulkCardMove,
    onCardClick,
    onAddCard,
    onRename,
    onEditColor,
    onEditWipLimit,
    onEditIsDone,
    onToggleVisibility,
    onDelete,
    onQuickDelete,
    priorityStyle = 'badge',
    canManageBoard = true,
    canEditCard = true,
    canMoveCard = true,
    canDeleteCard = true,
    isSelectionMode = false,
    selectedCardIds = new Set(),
    onToggleSelect,
  }: Props = $props();

  let items = $state(cards.map(card => ({ ...card, id: card.id })));
  let bulkDragActive = $state(false);
  let bulkDragId = $state<string | null>(null);

  // Sync items when cards prop changes (e.g. after API refresh)
  $effect(() => {
    items = cards.map(card => ({ ...card, id: card.id }));
    bulkDragActive = false;
    bulkDragId = null;
  });

  function handleConsider(e: CustomEvent<{ items: BoardCard[]; info: { id: string } }>) {
    items = e.detail.items;
    const dragId = e.detail.info.id;

    // Track bulk drag state (animation handles hiding via CSS)
    if (isSelectionMode && selectedCardIds.has(dragId) && selectedCardIds.size > 1) {
      bulkDragActive = true;
      bulkDragId = dragId;
    }
  }

  function handleFinalize(e: CustomEvent<{ items: BoardCard[]; info: { id: string } }>) {
    const newItems = e.detail.items;
    const draggedCardId = e.detail.info.id;

    const isFromThisColumn = cards.some(c => c.id === draggedCardId);
    const isBulkDrag = isSelectionMode && selectedCardIds.has(draggedCardId) && selectedCardIds.size > 1;

    // Target column: card dropped here from another column
    if (isBulkDrag && onBulkCardMove && !isFromThisColumn) {
      items = newItems;
      onBulkCardMove(Array.from(selectedCardIds), column.id);
      return; // bulkDragActive stays true until cards prop refreshes
    }

    // Source column: card dragged out
    if (isBulkDrag && isFromThisColumn && !newItems.some(item => item.id === draggedCardId)) {
      items = newItems;
      return; // bulkDragActive stays true until cards prop refreshes
    }

    // Cancelled or within-column: reset animation and restore
    bulkDragActive = false;
    bulkDragId = null;

    // Normal single-card move
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

  function transformDraggedElement(element: HTMLElement, data: any) {
    if (isSelectionMode && selectedCardIds.has(data.id) && selectedCardIds.size > 1) {
      const count = selectedCardIds.size;
      element.style.overflow = 'visible';

      // Stacked card layers peeking out behind the dragged card
      for (let i = 2; i >= 1; i--) {
        const layer = document.createElement('div');
        Object.assign(layer.style, {
          position: 'absolute',
          top: `${i * 4}px`,
          left: `${i * 4}px`,
          width: '100%',
          height: '100%',
          backgroundColor: 'white',
          borderRadius: '0.5rem',
          border: '1px solid #e5e7eb',
          boxShadow: '0 1px 2px 0 rgba(0,0,0,0.05)',
          pointerEvents: 'none',
        });
        element.insertBefore(layer, element.firstChild);
      }

      // Count badge
      const badge = document.createElement('div');
      badge.textContent = String(count);
      Object.assign(badge.style, {
        position: 'absolute',
        top: '-8px',
        right: '-8px',
        width: '24px',
        height: '24px',
        borderRadius: '50%',
        backgroundColor: '#6366f1',
        color: 'white',
        fontSize: '12px',
        fontWeight: '600',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: '10',
        boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
      });
      element.appendChild(badge);
    }
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
    {#if canManageBoard}
      <ColumnSettingsMenu
        {column}
        onRename={onRename ?? (() => {})}
        onEditColor={onEditColor ?? (() => {})}
        onEditWipLimit={onEditWipLimit ?? (() => {})}
        onEditIsDone={onEditIsDone ?? (() => {})}
        onToggleVisibility={onToggleVisibility ?? (() => {})}
        onDelete={onDelete ?? (() => {})}
      />
    {/if}
  </div>

  <div class="flex-1 flex flex-col min-h-0">
    <div
      class="flex-1 p-2 overflow-y-auto min-h-[60px]"
      use:dndzone={{
        items,
        flipDurationMs: 200,
        dropTargetStyle: canMoveCard ? { outline: '2px dashed #6366f1', outlineOffset: '-2px' } : {},
        dragDisabled: !canMoveCard,
        type: 'cards',
        transformDraggedElement,
      }}
      onconsider={handleConsider}
      onfinalize={handleFinalize}
    >
      {#each items as card (card.id)}
        {@const isScrunch = bulkDragActive && selectedCardIds.has(card.id) && card.id !== bulkDragId}
        <div class="mb-2" class:bulk-scrunch={isScrunch}>
          <KanbanCard
            {card}
            {onCardClick}
            onQuickDelete={canDeleteCard ? onQuickDelete : undefined}
            {priorityStyle}
            {canEditCard}
            {canDeleteCard}
            {isSelectionMode}
            isSelected={selectedCardIds.has(card.id)}
            {onToggleSelect}
          />
        </div>
      {/each}
    </div>

    <!-- Add card button - sticky at bottom, only show if user can add cards -->
    {#if onAddCard}
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
    {/if}
  </div>

  {#if column.wipLimit && cards.length >= column.wipLimit}
    <div class="px-4 py-2 bg-yellow-50 border-t border-yellow-200 text-xs text-yellow-700">
      WIP limit reached ({cards.length}/{column.wipLimit})
    </div>
  {/if}
</div>

<style>
  .bulk-scrunch {
    animation: bulk-scrunch 200ms ease-out forwards;
    pointer-events: none;
    overflow: hidden;
  }

  @keyframes bulk-scrunch {
    0% {
      transform: scale(1);
      opacity: 1;
      max-height: 200px;
      margin-bottom: 0.5rem;
    }
    40% {
      transform: scale(0.7);
      opacity: 0;
      max-height: 200px;
      margin-bottom: 0.5rem;
    }
    100% {
      transform: scale(0);
      opacity: 0;
      max-height: 0;
      margin-bottom: 0;
    }
  }
</style>
