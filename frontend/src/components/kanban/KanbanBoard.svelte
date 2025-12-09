<script lang="ts">
  import { onMount } from 'svelte';
  import { dndzone } from 'svelte-dnd-action';
  import KanbanColumn from './KanbanColumn.svelte';
  import CreateCardModal from './CreateCardModal.svelte';
  import CreateColumnModal from './CreateColumnModal.svelte';
  import EditColumnModal from './EditColumnModal.svelte';
  import CardDetailModal from './CardDetailModal.svelte';
  import CardDetailPanel from './CardDetailPanel.svelte';
  import { ConfirmModal } from '../ui';
  import type { BoardWithColumns, BoardColumn, BoardCard, Tag } from '../../lib/api/boards';
  import { getBoard, moveCard, getTags, deleteCard, reorderColumns, toggleColumnVisibility, deleteColumn, updateBoard } from '../../lib/api/boards';
  import EditableTitle from '../EditableTitle.svelte';

  interface Props {
    boardId: string;
  }

  let { boardId }: Props = $props();

  let board = $state<BoardWithColumns | null>(null);
  let tags = $state<Tag[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let showHiddenColumns = $state(false);

  // Modal states
  let showCreateCardModal = $state(false);
  let createCardColumnId = $state<string | null>(null);
  let showCardDetailModal = $state(false);
  let selectedCard = $state<BoardCard | null>(null);

  // Column modal states
  let showCreateColumnModal = $state(false);
  let showEditColumnModal = $state(false);
  let editColumnMode = $state<'rename' | 'color' | 'wipLimit'>('rename');
  let selectedColumn = $state<BoardColumn | null>(null);
  let showDeleteColumnConfirm = $state(false);
  let columnToDelete = $state<BoardColumn | null>(null);

  let deleteColumnMessage = $derived(() => {
    if (!columnToDelete) return '';
    const cardCount = columnToDelete.cards.length;
    const base = `Are you sure you want to delete the column "${columnToDelete.name}"?`;
    return cardCount > 0
      ? `${base} This will also delete ${cardCount} card(s) in this column.`
      : base;
  });

  // Card view mode: modal or panel
  let cardViewMode = $state<'modal' | 'panel'>('modal');

  // Priority display style: 'border' (left border) or 'badge' (text badge)
  let priorityStyle = $state<'border' | 'badge'>('badge');

  // Column drag state
  let columnItems = $state<BoardColumn[]>([]);
  let isDraggingColumn = $state(false);

  let visibleColumns = $derived(
    board?.columns.filter(col => showHiddenColumns || !col.isHidden).sort((a, b) => a.position - b.position) ?? []
  );

  // Sync columnItems with visibleColumns
  $effect(() => {
    if (!isDraggingColumn) {
      columnItems = visibleColumns.map(col => ({ ...col }));
    }
  });

  onMount(async () => {
    // Load card view mode preference from localStorage
    const savedMode = localStorage.getItem('cardViewMode');
    if (savedMode === 'panel' || savedMode === 'modal') {
      cardViewMode = savedMode;
    }
    // Load priority style preference from localStorage
    const savedPriorityStyle = localStorage.getItem('priorityStyle');
    if (savedPriorityStyle === 'border' || savedPriorityStyle === 'badge') {
      priorityStyle = savedPriorityStyle;
    }
    await loadBoard();
  });

  // Save card view mode preference when changed
  function setCardViewMode(mode: 'modal' | 'panel') {
    cardViewMode = mode;
    localStorage.setItem('cardViewMode', mode);
  }

  // Save priority style preference when changed
  function setPriorityStyle(style: 'border' | 'badge') {
    priorityStyle = style;
    localStorage.setItem('priorityStyle', style);
  }

  async function loadBoard() {
    try {
      loading = true;
      error = null;
      board = await getBoard(boardId);
      if (board) {
        tags = await getTags(board.project.id);
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load board';
    } finally {
      loading = false;
    }
  }

  async function handleRenameBoard(newName: string) {
    if (!board) return;
    const updated = await updateBoard(boardId, newName);
    board = { ...board, name: updated.name };
  }

  async function refreshBoard() {
    try {
      board = await getBoard(boardId);
      if (board) {
        tags = await getTags(board.project.id);
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to refresh board';
    }
  }

  async function reloadTags() {
    if (board) {
      tags = await getTags(board.project.id);
    }
  }

  async function handleCardMove(cardId: string, columnId: string, afterCardId: string | null) {
    try {
      await moveCard(cardId, columnId, afterCardId ?? undefined);
      // Don't refresh - the UI is already updated optimistically by svelte-dnd-action
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to move card';
      // Revert on error by refetching
      await refreshBoard();
    }
  }

  function handleCardClick(card: BoardCard) {
    selectedCard = card;
    showCardDetailModal = true;
  }

  function handleAddCard(columnId: string) {
    createCardColumnId = columnId;
    showCreateCardModal = true;
  }

  // Column settings handlers
  function handleColumnRename(column: BoardColumn) {
    selectedColumn = column;
    editColumnMode = 'rename';
    showEditColumnModal = true;
  }

  function handleColumnEditColor(column: BoardColumn) {
    selectedColumn = column;
    editColumnMode = 'color';
    showEditColumnModal = true;
  }

  function handleColumnEditWipLimit(column: BoardColumn) {
    selectedColumn = column;
    editColumnMode = 'wipLimit';
    showEditColumnModal = true;
  }

  async function handleColumnToggleVisibility(column: BoardColumn) {
    try {
      await toggleColumnVisibility(column.id);
      await refreshBoard();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to toggle column visibility';
    }
  }

  function handleColumnDelete(column: BoardColumn) {
    columnToDelete = column;
    showDeleteColumnConfirm = true;
  }

  async function confirmDeleteColumn() {
    if (!columnToDelete) return;
    try {
      await deleteColumn(columnToDelete.id);
      showDeleteColumnConfirm = false;
      columnToDelete = null;
      await refreshBoard();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete column';
    }
  }

  // Column drag and drop handlers
  function handleColumnConsider(e: CustomEvent<{ items: BoardColumn[] }>) {
    isDraggingColumn = true;
    columnItems = e.detail.items;
  }

  async function handleColumnFinalize(e: CustomEvent<{ items: BoardColumn[] }>) {
    const newItems = e.detail.items;
    columnItems = newItems;
    isDraggingColumn = false;

    // Check if order actually changed
    const originalOrder = visibleColumns.map(c => c.id);
    const newOrder = newItems.map(c => c.id);
    const orderChanged = originalOrder.some((id, index) => id !== newOrder[index]);

    if (orderChanged && board) {
      // Optimistically update the board state to prevent flash
      const updatedColumns = newItems.map((col, index) => ({
        ...col,
        position: index
      }));
      board = { ...board, columns: updatedColumns };

      try {
        await reorderColumns(board.id, newOrder);
        // No need to refresh - optimistic update is sufficient
      } catch (e) {
        error = e instanceof Error ? e.message : 'Failed to reorder columns';
        // Revert on error by refetching
        await refreshBoard();
      }
    }
  }

  async function handleCardCreated() {
    showCreateCardModal = false;
    createCardColumnId = null;
    await refreshBoard();
  }

  async function handleCardUpdated() {
    showCardDetailModal = false;
    selectedCard = null;
    await refreshBoard();
  }

  async function handleColumnCreated() {
    showCreateColumnModal = false;
    await refreshBoard();
  }

  async function handleColumnUpdated() {
    showEditColumnModal = false;
    selectedColumn = null;
    await refreshBoard();
  }

  async function handleQuickDelete(card: BoardCard) {
    try {
      await deleteCard(card.id);
      await refreshBoard();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete card';
    }
  }

  function closeCreateCardModal() {
    showCreateCardModal = false;
    createCardColumnId = null;
  }

  function closeCardDetailModal() {
    showCardDetailModal = false;
    selectedCard = null;
  }
</script>

{#if loading}
  <div class="flex items-center justify-center h-64">
    <div class="text-gray-500">Loading board...</div>
  </div>
{:else if error}
  <div class="rounded-md bg-red-50 p-4">
    <p class="text-sm text-red-700">{error}</p>
    <button
      type="button"
      class="mt-2 text-sm text-red-600 hover:text-red-700 underline"
      onclick={loadBoard}
    >
      Retry
    </button>
  </div>
{:else if board}
  <div class="h-full flex flex-col">
    <!-- Board header -->
    <div class="flex items-center justify-between mb-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">
          <EditableTitle value={board.name} onSave={handleRenameBoard} />
        </h1>
        {#if board.description}
          <p class="text-sm text-gray-500">{board.description}</p>
        {/if}
      </div>
      <div class="flex items-center gap-4">
        <!-- Priority style toggle -->
        <div class="flex items-center gap-2 text-sm text-gray-600">
          <span>Priority:</span>
          <div class="flex rounded-md shadow-sm">
            <button
              type="button"
              onclick={() => setPriorityStyle('border')}
              class="px-2 py-1 text-xs font-medium rounded-l-md border {priorityStyle === 'border' ? 'bg-indigo-100 text-indigo-700 border-indigo-300' : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'}"
              title="Show priority as colored left border"
            >
              Border
            </button>
            <button
              type="button"
              onclick={() => setPriorityStyle('badge')}
              class="px-2 py-1 text-xs font-medium rounded-r-md border-t border-r border-b -ml-px {priorityStyle === 'badge' ? 'bg-indigo-100 text-indigo-700 border-indigo-300' : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'}"
              title="Show priority as text badge"
            >
              Badge
            </button>
          </div>
        </div>
        <label class="flex items-center gap-2 text-sm text-gray-600">
          <input
            type="checkbox"
            bind:checked={showHiddenColumns}
            class="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
          />
          Show hidden columns
        </label>
      </div>
    </div>

    <!-- Columns container with drag and drop -->
    <div class="flex-1 overflow-x-auto">
      <div
        class="flex gap-4 h-full pb-4"
        use:dndzone={{
          items: columnItems,
          flipDurationMs: 200,
          type: 'columns',
          dropTargetStyle: {},
        }}
        onconsider={handleColumnConsider}
        onfinalize={handleColumnFinalize}
      >
        {#each columnItems as column (column.id)}
          <KanbanColumn
            {column}
            cards={column.cards}
            onCardMove={handleCardMove}
            onCardClick={handleCardClick}
            onAddCard={handleAddCard}
            onRename={() => handleColumnRename(column)}
            onEditColor={() => handleColumnEditColor(column)}
            onEditWipLimit={() => handleColumnEditWipLimit(column)}
            onToggleVisibility={() => handleColumnToggleVisibility(column)}
            onDelete={() => handleColumnDelete(column)}
            onQuickDelete={handleQuickDelete}
            {priorityStyle}
          />
        {/each}

        <!-- Add column button -->
        <button
          type="button"
          class="flex-shrink-0 w-72 h-32 bg-gray-50 border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center text-gray-500 hover:text-gray-700 hover:border-gray-400 hover:bg-gray-100 transition-colors"
          onclick={() => showCreateColumnModal = true}
        >
          <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Add Column
        </button>
      </div>
    </div>
  </div>

  <!-- Create Card Modal -->
  {#if board}
    <CreateCardModal
      open={showCreateCardModal && createCardColumnId !== null}
      columnId={createCardColumnId ?? ''}
      projectId={board.project.id}
      {tags}
      onClose={closeCreateCardModal}
      onCreated={handleCardCreated}
      onTagsChanged={reloadTags}
    />

    <!-- Create Column Modal -->
    <CreateColumnModal
      open={showCreateColumnModal}
      boardId={board.id}
      onClose={() => showCreateColumnModal = false}
      onCreated={handleColumnCreated}
    />

    <!-- Edit Column Modal -->
    <EditColumnModal
      open={showEditColumnModal}
      column={selectedColumn}
      mode={editColumnMode}
      onClose={() => { showEditColumnModal = false; selectedColumn = null; }}
      onUpdated={handleColumnUpdated}
    />

    <!-- Delete Column Confirmation -->
    <ConfirmModal
      isOpen={showDeleteColumnConfirm && columnToDelete !== null}
      title="Delete Column"
      message={deleteColumnMessage()}
      confirmText="Delete"
      variant="danger"
      onConfirm={confirmDeleteColumn}
      onCancel={() => { showDeleteColumnConfirm = false; columnToDelete = null; }}
    />

    <!-- Card Detail Modal (when in modal mode) -->
    <CardDetailModal
      open={cardViewMode === 'modal' && showCardDetailModal}
      card={selectedCard}
      projectId={board.project.id}
      {tags}
      onClose={closeCardDetailModal}
      onUpdated={handleCardUpdated}
      onAutoSaved={refreshBoard}
      onTagsChanged={reloadTags}
      viewMode={cardViewMode}
      onViewModeChange={setCardViewMode}
    />
  {/if}

  <!-- Card Detail Panel (when in panel mode) -->
  {#if cardViewMode === 'panel'}
    <CardDetailPanel
      card={selectedCard}
      projectId={board?.project?.id ?? ''}
      {tags}
      isOpen={showCardDetailModal}
      onClose={closeCardDetailModal}
      onUpdated={handleCardUpdated}
      onAutoSaved={refreshBoard}
      onTagsChanged={reloadTags}
      viewMode={cardViewMode}
      onViewModeChange={setCardViewMode}
    />
  {/if}
{/if}
