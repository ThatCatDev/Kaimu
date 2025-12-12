<script lang="ts">
  import { onMount } from 'svelte';
  import { dndzone } from 'svelte-dnd-action';
  import { toast } from 'svelte-sonner';
  import KanbanColumn from './KanbanColumn.svelte';
  import CreateCardModal from './CreateCardModal.svelte';
  import CreateColumnModal from './CreateColumnModal.svelte';
  import EditColumnModal from './EditColumnModal.svelte';
  import CardDetailModal from './CardDetailModal.svelte';
  import CardDetailPanel from './CardDetailPanel.svelte';
  import { ConfirmModal } from '../ui';
  import type { BoardWithColumns, BoardColumn, BoardCard, Tag } from '../../lib/api/boards';
  import { getBoard, moveCard, getTags, deleteCard, reorderColumns, toggleColumnVisibility, deleteColumn, updateBoard } from '../../lib/api/boards';
  import { moveCardToBacklog, getActiveSprint, addCardToSprint, type SprintData } from '../../lib/api/sprints';
  import EditableTitle from '../EditableTitle.svelte';
  import EditableDescription from '../EditableDescription.svelte';
  import { Permissions } from '../../lib/stores/permissions.svelte';
  import { getMyPermissions } from '../../lib/api/rbac';

  interface Props {
    boardId: string;
    initialCardId?: string | null;
  }

  let { boardId, initialCardId }: Props = $props();

  let board = $state<BoardWithColumns | null>(null);
  let tags = $state<Tag[]>([]);
  let activeSprint = $state<SprintData | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let showHiddenColumns = $state(false);

  // Modal states
  let showCreateCardModal = $state(false);
  let createCardColumnId = $state<string | null>(null);
  let showCardDetailModal = $state(false);
  // Use a cloned card object for editing to prevent board updates from affecting it
  let editingCard = $state<BoardCard | null>(null);
  let selectedProjectId = $state<string>(''); // Stable project ID for card detail view

  // Column modal states
  let showCreateColumnModal = $state(false);
  let showEditColumnModal = $state(false);
  let editColumnMode = $state<'rename' | 'color' | 'wipLimit' | 'isDone'>('rename');
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

  // Permission state - loaded client-side after board loads
  let permissions = $state<string[]>([]);

  // Permission checks
  let canManageBoard = $derived(permissions.includes(Permissions.BOARD_MANAGE));
  let canCreateCard = $derived(permissions.includes(Permissions.CARD_CREATE));
  let canEditCard = $derived(permissions.includes(Permissions.CARD_EDIT));
  let canMoveCard = $derived(permissions.includes(Permissions.CARD_MOVE));
  let canDeleteCard = $derived(permissions.includes(Permissions.CARD_DELETE));


  // Filter cards based on sprint status:
  // - If active sprint: only show cards in the active sprint
  // - If no active sprint: show cards NOT in any closed sprint
  function filterCardsBySprint(cards: BoardCard[]): BoardCard[] {
    const currentSprint = activeSprint;
    if (currentSprint) {
      // Only show cards that belong to the active sprint
      return cards.filter(card =>
        card.sprints?.some(s => s.id === currentSprint.id) ?? false
      );
    } else {
      // No active sprint: show cards that are NOT in any CLOSED sprint
      // This includes: cards with no sprints, cards in FUTURE sprints only
      return cards.filter(card => {
        if (!card.sprints || card.sprints.length === 0) return true;
        // If any sprint is CLOSED, hide the card
        return !card.sprints.some(s => s.status === 'CLOSED');
      });
    }
  }

  let visibleColumns = $derived(
    board?.columns.filter(col => showHiddenColumns || !col.isHidden).sort((a, b) => a.position - b.position) ?? []
  );

  // Columns with cards filtered by sprint
  let filteredColumns = $derived(
    visibleColumns.map(col => ({
      ...col,
      cards: filterCardsBySprint(col.cards)
    }))
  );

  // Sync columnItems with filteredColumns (cards filtered by sprint)
  $effect(() => {
    if (!isDraggingColumn) {
      columnItems = filteredColumns.map(col => ({ ...col }));
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
        const [projectTags, perms, active] = await Promise.all([
          getTags(board.project.id),
          getMyPermissions('project', board.project.id),
          getActiveSprint(boardId)
        ]);
        tags = projectTags;
        permissions = perms;
        activeSprint = active;

        // If initialCardId is provided, find and open the card
        if (initialCardId) {
          const card = findCardById(initialCardId);
          if (card) {
            handleCardClick(card);
          }
        }
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load board';
    } finally {
      loading = false;
    }
  }

  // Helper to find a card by ID across all columns
  function findCardById(cardId: string): BoardCard | null {
    if (!board) return null;
    for (const column of board.columns) {
      const card = column.cards.find(c => c.id === cardId);
      if (card) return card;
    }
    return null;
  }

  async function handleUpdateBoardName(newName: string) {
    if (!board) return;
    const updated = await updateBoard(boardId, newName);
    board = { ...board, name: updated.name };
  }

  async function handleUpdateBoardDescription(newDescription: string) {
    if (!board) return;
    const updated = await updateBoard(boardId, undefined, newDescription);
    board = { ...board, description: updated.description };
  }

  // Export refreshBoard so parent components can trigger refresh
  export async function refreshBoard() {
    try {
      const [newBoard, active] = await Promise.all([
        getBoard(boardId),
        getActiveSprint(boardId)
      ]);
      board = newBoard;
      activeSprint = active;
      if (board) {
        tags = await getTags(board.project.id);
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to refresh board';
    }
  }

  // Update a single card in the board display without affecting editingCard
  // This allows the board to show updated card data while editing without resetting the form
  function updateCardInBoard(cardId: string, updates: Partial<BoardCard>) {
    if (!board) return;

    board = {
      ...board,
      columns: board.columns.map(col => ({
        ...col,
        cards: col.cards.map(c =>
          c.id === cardId ? { ...c, ...updates } : c
        )
      }))
    };
  }

  async function reloadTags() {
    if (board) {
      tags = await getTags(board.project.id);
    }
  }

  async function handleCardMove(cardId: string, columnId: string, afterCardId: string | null) {
    if (!board) return;

    // Find source and destination columns
    const card = findCardById(cardId);
    const sourceColumn = board.columns.find(col => col.cards.some(c => c.id === cardId));
    const destColumn = board.columns.find(col => col.id === columnId);

    // Optimistically update board state so column reordering doesn't lose card positions
    if (card) {
      board = {
        ...board,
        columns: board.columns.map(col => {
          // Remove card from its current column
          const filteredCards = col.cards.filter(c => c.id !== cardId);

          if (col.id === columnId) {
            // Add card to destination column at the right position
            const insertIndex = afterCardId
              ? filteredCards.findIndex(c => c.id === afterCardId) + 1
              : 0;
            const newCards = [...filteredCards];
            newCards.splice(insertIndex, 0, card);
            return { ...col, cards: newCards };
          }

          return { ...col, cards: filteredCards };
        })
      };
    }

    try {
      await moveCard(cardId, columnId, afterCardId ?? undefined);

      // Handle sprint changes based on column transitions
      if (sourceColumn && destColumn) {
        // Moving TO backlog column -> remove from all sprints
        if (destColumn.isBacklog && !sourceColumn.isBacklog) {
          await moveCardToBacklog(cardId);
        }
        // Moving FROM backlog column to non-backlog -> add to active sprint
        else if (!destColumn.isBacklog && sourceColumn.isBacklog) {
          const activeSprint = await getActiveSprint(board.id);
          if (activeSprint) {
            await addCardToSprint(cardId, activeSprint.id);
          }
        }
      }

      // Don't refresh - the UI is already updated optimistically
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to move card';
      if (message.toLowerCase().includes('permission') || message.toLowerCase().includes('unauthorized')) {
        toast.error('Permission denied: You cannot move cards');
      } else {
        toast.error(message);
      }
      // Revert on error by refetching
      await refreshBoard();
    }
  }

  function handleCardClick(card: BoardCard) {
    // Only update editingCard if it's a different card
    // This prevents unnecessary resets when clicking the same card or during re-renders
    if (!editingCard || editingCard.id !== card.id) {
      editingCard = { ...card, tags: card.tags ? [...card.tags] : [] };
    }
    selectedProjectId = board?.project?.id ?? '';
    showCardDetailModal = true;

    // Update URL with card parameter for sharing
    updateUrlWithCard(card.id);
  }

  // Update URL to include/exclude card parameter
  function updateUrlWithCard(cardId: string | null) {
    const url = new URL(window.location.href);
    if (cardId) {
      url.searchParams.set('card', cardId);
    } else {
      url.searchParams.delete('card');
    }
    window.history.replaceState({}, '', url.toString());
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

  function handleColumnEditIsDone(column: BoardColumn) {
    selectedColumn = column;
    editColumnMode = 'isDone';
    showEditColumnModal = true;
  }

  async function handleColumnToggleVisibility(column: BoardColumn) {
    try {
      await toggleColumnVisibility(column.id);
      await refreshBoard();
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to toggle column visibility';
      toast.error(message);
    }
  }

  function handleColumnDelete(column: BoardColumn) {
    // Prevent deleting backlog column
    if (column.isBacklog) {
      toast.error('Cannot delete the backlog column. You can hide it instead.');
      return;
    }
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
      toast.success('Column deleted');
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to delete column';
      toast.error(message);
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
        const message = e instanceof Error ? e.message : 'Failed to reorder columns';
        toast.error(message);
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
    editingCard = null;
    updateUrlWithCard(null);
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
      toast.success('Card deleted');
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to delete card';
      toast.error(message);
    }
  }

  function closeCreateCardModal() {
    showCreateCardModal = false;
    createCardColumnId = null;
  }

  function closeCardDetailModal() {
    showCardDetailModal = false;
    editingCard = null;
    updateUrlWithCard(null);
  }
</script>

{#if loading}
  <div class="h-full flex flex-col">
    <div class="flex items-center justify-end py-3">
      <div class="flex items-center gap-4">
        <div class="h-8 w-32 bg-gray-200 rounded animate-pulse"></div>
        <div class="h-5 w-40 bg-gray-200 rounded animate-pulse"></div>
      </div>
    </div>
    <div class="flex-1 overflow-x-auto">
      <div class="flex gap-4 h-full pb-4">
        {#each [1, 2, 3, 4] as _}
          <div class="w-72 flex-shrink-0 bg-gray-100 rounded-lg p-3">
            <div class="flex items-center justify-between mb-3">
              <div class="h-5 w-24 bg-gray-200 rounded animate-pulse"></div>
              <div class="h-5 w-8 bg-gray-200 rounded animate-pulse"></div>
            </div>
            <div class="space-y-2">
              {#each [1, 2] as __}
                <div class="bg-white rounded-lg shadow-sm p-3">
                  <div class="h-4 w-full bg-gray-200 rounded animate-pulse mb-2"></div>
                  <div class="h-3 w-2/3 bg-gray-200 rounded animate-pulse"></div>
                </div>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    </div>
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
    <!-- Board controls -->
    <div class="flex items-center justify-end py-3">
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
          dragDisabled: !canManageBoard,
        }}
        onconsider={handleColumnConsider}
        onfinalize={handleColumnFinalize}
      >
        {#each columnItems as column (column.id)}
          <KanbanColumn
            {column}
            cards={column.cards}
            onCardMove={canMoveCard ? handleCardMove : () => {}}
            onCardClick={handleCardClick}
            onAddCard={canCreateCard ? handleAddCard : undefined}
            onRename={canManageBoard ? () => handleColumnRename(column) : undefined}
            onEditColor={canManageBoard ? () => handleColumnEditColor(column) : undefined}
            onEditWipLimit={canManageBoard ? () => handleColumnEditWipLimit(column) : undefined}
            onEditIsDone={canManageBoard ? () => handleColumnEditIsDone(column) : undefined}
            onToggleVisibility={canManageBoard ? () => handleColumnToggleVisibility(column) : undefined}
            onDelete={canManageBoard ? () => handleColumnDelete(column) : undefined}
            onQuickDelete={canDeleteCard ? handleQuickDelete : undefined}
            {priorityStyle}
            {canManageBoard}
            {canEditCard}
            {canMoveCard}
            {canDeleteCard}
          />
        {/each}

        <!-- Add column button - only show if user can manage board -->
        {#if canManageBoard}
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
        {/if}
      </div>
    </div>
  </div>

  <!-- Create Card Modal -->
  {#if board}
    <CreateCardModal
      open={showCreateCardModal && createCardColumnId !== null}
      columnId={createCardColumnId ?? ''}
      projectId={board.project.id}
      boardId={board.id}
      isBacklogColumn={board.columns.find(c => c.id === createCardColumnId)?.isBacklog ?? false}
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
      card={editingCard}
      projectId={selectedProjectId}
      organizationId={board?.project.organization.id ?? ''}
      {boardId}
      {tags}
      onClose={closeCardDetailModal}
      onUpdated={handleCardUpdated}
      onCardDataChanged={updateCardInBoard}
      onTagsChanged={reloadTags}
      viewMode={cardViewMode}
      onViewModeChange={setCardViewMode}
      canEditCard={canEditCard}
      canDeleteCard={canDeleteCard}
    />
  {/if}

  <!-- Card Detail Panel (when in panel mode) -->
  {#if cardViewMode === 'panel'}
    <CardDetailPanel
      card={editingCard}
      projectId={selectedProjectId}
      organizationId={board?.project.organization.id ?? ''}
      {boardId}
      {tags}
      isOpen={showCardDetailModal}
      onClose={closeCardDetailModal}
      onUpdated={handleCardUpdated}
      onCardDataChanged={updateCardInBoard}
      onTagsChanged={reloadTags}
      viewMode={cardViewMode}
      onViewModeChange={setCardViewMode}
      canEditCard={canEditCard}
      canDeleteCard={canDeleteCard}
    />
  {/if}
{/if}
