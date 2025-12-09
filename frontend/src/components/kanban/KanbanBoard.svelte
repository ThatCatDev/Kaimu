<script lang="ts">
  import { onMount } from 'svelte';
  import KanbanColumn from './KanbanColumn.svelte';
  import CreateCardModal from './CreateCardModal.svelte';
  import CardDetailModal from './CardDetailModal.svelte';
  import CardDetailPanel from './CardDetailPanel.svelte';
  import type { BoardWithColumns, BoardColumn, BoardCard, Tag } from '../../lib/api/boards';
  import { getBoard, moveCard, getTags, deleteCard } from '../../lib/api/boards';

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

  // Card view mode: modal or panel
  let cardViewMode = $state<'modal' | 'panel'>('modal');

  // Priority display style: 'border' (left border) or 'badge' (text badge)
  let priorityStyle = $state<'border' | 'badge'>('badge');

  let visibleColumns = $derived(
    board?.columns.filter(col => showHiddenColumns || !col.isHidden).sort((a, b) => a.position - b.position) ?? []
  );

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
      await refreshBoard();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to move card';
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

  function handleColumnSettings(column: BoardColumn) {
    // TODO: Implement column settings modal
    console.log('Column settings', column);
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
        <h1 class="text-2xl font-bold text-gray-900">{board.name}</h1>
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

    <!-- Columns container -->
    <div class="flex-1 overflow-x-auto">
      <div class="flex gap-4 h-full pb-4">
        {#each visibleColumns as column (column.id)}
          <KanbanColumn
            {column}
            cards={column.cards}
            onCardMove={handleCardMove}
            onCardClick={handleCardClick}
            onAddCard={handleAddCard}
            onColumnSettings={handleColumnSettings}
            onQuickDelete={handleQuickDelete}
            {priorityStyle}
          />
        {/each}

        <!-- Add column button -->
        <button
          type="button"
          class="flex-shrink-0 w-72 bg-gray-50 border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center text-gray-500 hover:text-gray-700 hover:border-gray-400 transition-colors"
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
  {#if showCreateCardModal && createCardColumnId && board}
    <CreateCardModal
      columnId={createCardColumnId}
      projectId={board.project.id}
      {tags}
      onClose={closeCreateCardModal}
      onCreated={handleCardCreated}
      onTagsChanged={reloadTags}
    />
  {/if}

  <!-- Card Detail Modal (when in modal mode) -->
  {#if cardViewMode === 'modal' && showCardDetailModal && selectedCard && board}
    <CardDetailModal
      card={selectedCard}
      projectId={board.project.id}
      {tags}
      onClose={closeCardDetailModal}
      onUpdated={handleCardUpdated}
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
      onTagsChanged={reloadTags}
      viewMode={cardViewMode}
      onViewModeChange={setCardViewMode}
    />
  {/if}
{/if}
