<script lang="ts">
  import { onMount } from 'svelte';
  import { toast } from 'svelte-sonner';
  import PlanningSection from './PlanningSection.svelte';
  import PlanningCardRow from './PlanningCardRow.svelte';
  import CardDetailPanel from '../kanban/CardDetailPanel.svelte';
  import BoardPageLayout from '../kanban/BoardPageLayout.svelte';
  import { Button } from '../ui';
  import {
    getActiveSprint,
    getFutureSprints,
    getClosedSprints,
    getBacklogCards,
    getSprintCards,
    addCardToSprint,
    moveCardToBacklog,
    type SprintData,
    type SprintCard,
    type BacklogCard,
  } from '../../lib/api/sprints';
  import { getBoard, getTags, type BoardWithColumns, type BoardCard, type Tag } from '../../lib/api/boards';

  interface Props {
    boardId: string;
    projectId: string;
    initialCardId?: string | null;
  }

  let { boardId, projectId, initialCardId }: Props = $props();

  // Board data
  let board = $state<BoardWithColumns | null>(null);
  let tags = $state<Tag[]>([]);

  // Sprint data
  let activeSprint = $state<SprintData | null>(null);
  let futureSprints = $state<SprintData[]>([]);
  let closedSprints = $state<SprintData[]>([]);
  let closedSprintsPageInfo = $state<{ hasNextPage: boolean; endCursor: string | null; totalCount: number } | null>(null);
  let backlogCards = $state<BacklogCard[]>([]);

  // Sprint cards - lazy loaded
  let sprintCards = $state<Map<string, SprintCard[]>>(new Map());
  let loadingSprintCards = $state<Set<string>>(new Set());

  // UI state
  let loading = $state(true);
  let expandedSections = $state<Set<string>>(new Set(['active', 'backlog']));

  // Card detail panel state
  let selectedCard = $state<SprintCard | BacklogCard | null>(null);
  let showCardPanel = $state(false);
  let cardViewMode = $state<'modal' | 'panel'>('panel');

  // Computed: all available sprints for moving cards
  let availableSprints = $derived([
    ...(activeSprint ? [activeSprint] : []),
    ...futureSprints,
  ]);

  onMount(async () => {
    // Load card view mode preference
    const savedMode = localStorage.getItem('cardViewMode');
    if (savedMode === 'panel' || savedMode === 'modal') {
      cardViewMode = savedMode;
    }

    await loadData();

    // Open card from URL if provided
    if (initialCardId) {
      await openCardById(initialCardId);
    }
  });

  async function loadData() {
    try {
      loading = true;

      // Load board, tags, sprints, and backlog in parallel
      const [boardData, tagsData, active, future, closedResult, backlog] = await Promise.all([
        getBoard(boardId),
        getTags(projectId),
        getActiveSprint(boardId),
        getFutureSprints(boardId),
        getClosedSprints(boardId, 10),
        getBacklogCards(boardId),
      ]);

      board = boardData;
      tags = tagsData;
      activeSprint = active;
      futureSprints = future;
      closedSprints = closedResult.sprints;
      closedSprintsPageInfo = closedResult.pageInfo;
      backlogCards = backlog;

      // Load active sprint cards immediately if expanded
      if (activeSprint && expandedSections.has('active')) {
        await loadSprintCards(activeSprint.id);
      }
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load data';
      toast.error(message);
    } finally {
      loading = false;
    }
  }

  async function loadSprintCards(sprintId: string) {
    if (sprintCards.has(sprintId) || loadingSprintCards.has(sprintId)) return;

    loadingSprintCards.add(sprintId);
    loadingSprintCards = new Set(loadingSprintCards);

    try {
      const cards = await getSprintCards(sprintId);
      sprintCards.set(sprintId, cards);
      sprintCards = new Map(sprintCards);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load sprint cards';
      toast.error(message);
    } finally {
      loadingSprintCards.delete(sprintId);
      loadingSprintCards = new Set(loadingSprintCards);
    }
  }

  async function loadMoreClosedSprints() {
    if (!closedSprintsPageInfo?.hasNextPage) return;

    try {
      const result = await getClosedSprints(boardId, 10, closedSprintsPageInfo.endCursor ?? undefined);
      closedSprints = [...closedSprints, ...result.sprints];
      closedSprintsPageInfo = result.pageInfo;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load more sprints';
      toast.error(message);
    }
  }

  function handleSectionToggle(sectionId: string, expanded: boolean) {
    if (expanded) {
      expandedSections.add(sectionId);
    } else {
      expandedSections.delete(sectionId);
    }
    expandedSections = new Set(expandedSections);

    // Lazy load sprint cards when expanding
    if (expanded && sectionId !== 'backlog' && sectionId !== 'closed') {
      loadSprintCards(sectionId);
    }
  }

  function handleCardClick(card: SprintCard | BacklogCard) {
    selectedCard = card;
    showCardPanel = true;
    updateUrlWithCard(card.id);
  }

  async function handleMoveToSprint(cardId: string, sprintId: string) {
    try {
      await addCardToSprint(cardId, sprintId);

      // Refresh data
      await loadData();
      // Clear cached sprint cards to force reload
      sprintCards.clear();
      sprintCards = new Map(sprintCards);

      // Reload expanded sprints
      for (const sectionId of expandedSections) {
        if (sectionId !== 'backlog' && sectionId !== 'closed') {
          await loadSprintCards(sectionId);
        }
      }

      toast.success('Card moved to sprint');
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to move card';
      toast.error(message);
    }
  }

  async function handleMoveToBacklog(cardId: string) {
    try {
      await moveCardToBacklog(cardId);

      // Refresh data
      await loadData();
      // Clear cached sprint cards
      sprintCards.clear();
      sprintCards = new Map(sprintCards);

      // Reload expanded sprints
      for (const sectionId of expandedSections) {
        if (sectionId !== 'backlog' && sectionId !== 'closed') {
          await loadSprintCards(sectionId);
        }
      }

      toast.success('Card moved to backlog');
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to move card';
      toast.error(message);
    }
  }

  function closeCardPanel() {
    showCardPanel = false;
    selectedCard = null;
    updateUrlWithCard(null);
  }

  async function handleCardUpdated() {
    // Refresh data after card update
    await loadData();
    sprintCards.clear();
    sprintCards = new Map(sprintCards);

    for (const sectionId of expandedSections) {
      if (sectionId !== 'backlog' && sectionId !== 'closed') {
        await loadSprintCards(sectionId);
      }
    }
  }

  function calculateStoryPoints(cards: SprintCard[] | BacklogCard[]): number {
    return cards.reduce((sum, card) => sum + (card.storyPoints ?? 0), 0);
  }

  function setViewMode(mode: 'modal' | 'panel') {
    cardViewMode = mode;
    localStorage.setItem('cardViewMode', mode);
  }

  function updateUrlWithCard(cardId: string | null) {
    const url = new URL(window.location.href);
    if (cardId) {
      url.searchParams.set('card', cardId);
    } else {
      url.searchParams.delete('card');
    }
    window.history.replaceState({}, '', url.toString());
  }

  // Find a card by ID across all data sources
  function findCardById(cardId: string): SprintCard | BacklogCard | null {
    // Check backlog cards
    const backlogCard = backlogCards.find(c => c.id === cardId);
    if (backlogCard) return backlogCard;

    // Check all loaded sprint cards
    for (const [, cards] of sprintCards) {
      const sprintCard = cards.find(c => c.id === cardId);
      if (sprintCard) return sprintCard;
    }

    return null;
  }

  // Open a card by ID (loads sprint cards if needed)
  async function openCardById(cardId: string) {
    // First check if already loaded
    let card = findCardById(cardId);

    // If not found, load all sprint cards and check again
    if (!card) {
      const sprintsToLoad = [
        ...(activeSprint ? [activeSprint.id] : []),
        ...futureSprints.map(s => s.id),
      ];

      await Promise.all(sprintsToLoad.map(id => loadSprintCards(id)));
      card = findCardById(cardId);
    }

    if (card) {
      selectedCard = card;
      showCardPanel = true;
    }
  }

  // Convert SprintCard/BacklogCard to BoardCard type for CardDetailPanel
  function toBoardCard(card: SprintCard | BacklogCard): BoardCard {
    return {
      ...card,
      sprints: [], // Will be loaded by CardDetailPanel
    } as BoardCard;
  }
</script>

<BoardPageLayout {board} {boardId} {projectId} currentPage="planning">
  {#snippet children()}
  <div class="h-full flex flex-col bg-gray-50">
    <div class="flex-1 overflow-auto p-6">
    {#if loading}
      <div class="flex items-center justify-center py-12">
        <span class="text-gray-500">Loading...</span>
      </div>
    {:else}
      <div class="max-w-5xl mx-auto space-y-4">
        <!-- Active Sprint -->
        {#if activeSprint}
          {@const cards = sprintCards.get(activeSprint.id) ?? []}
          {@const isLoading = loadingSprintCards.has(activeSprint.id)}
          <PlanningSection
            title={activeSprint.name}
            cardCount={cards.length}
            storyPoints={calculateStoryPoints(cards)}
            expanded={expandedSections.has('active')}
            variant="active"
            badge="Active"
            onToggle={(expanded) => handleSectionToggle('active', expanded)}
          >
            {#if isLoading}
              <div class="px-4 py-8 text-center text-gray-500">Loading cards...</div>
            {:else if cards.length === 0}
              <div class="px-4 py-8 text-center text-gray-500">No cards in this sprint</div>
            {:else}
              <div class="divide-y divide-gray-100">
                {#each cards as card (card.id)}
                  <PlanningCardRow
                    {card}
                    availableSprints={futureSprints}
                    onCardClick={handleCardClick}
                    onMoveToSprint={handleMoveToSprint}
                    onMoveToBacklog={handleMoveToBacklog}
                  />
                {/each}
              </div>
            {/if}
          </PlanningSection>
        {/if}

        <!-- Future Sprints -->
        {#each futureSprints as sprint (sprint.id)}
          {@const cards = sprintCards.get(sprint.id) ?? []}
          {@const isLoading = loadingSprintCards.has(sprint.id)}
          {@const isExpanded = expandedSections.has(sprint.id)}
          <PlanningSection
            title={sprint.name}
            cardCount={isExpanded ? cards.length : 0}
            storyPoints={isExpanded ? calculateStoryPoints(cards) : 0}
            expanded={isExpanded}
            variant="future"
            badge="Future"
            onToggle={(expanded) => handleSectionToggle(sprint.id, expanded)}
          >
            {#if isLoading}
              <div class="px-4 py-8 text-center text-gray-500">Loading cards...</div>
            {:else if cards.length === 0}
              <div class="px-4 py-8 text-center text-gray-500">No cards in this sprint</div>
            {:else}
              <div class="divide-y divide-gray-100">
                {#each cards as card (card.id)}
                  <PlanningCardRow
                    {card}
                    availableSprints={[...(activeSprint ? [activeSprint] : []), ...futureSprints.filter(s => s.id !== sprint.id)]}
                    onCardClick={handleCardClick}
                    onMoveToSprint={handleMoveToSprint}
                    onMoveToBacklog={handleMoveToBacklog}
                  />
                {/each}
              </div>
            {/if}
          </PlanningSection>
        {/each}

        <!-- Backlog -->
        <PlanningSection
          title="Backlog"
          cardCount={backlogCards.length}
          storyPoints={calculateStoryPoints(backlogCards)}
          expanded={expandedSections.has('backlog')}
          variant="backlog"
          onToggle={(expanded) => handleSectionToggle('backlog', expanded)}
        >
          {#if backlogCards.length === 0}
            <div class="px-4 py-8 text-center text-gray-500">No cards in backlog</div>
          {:else}
            <div class="divide-y divide-gray-100">
              {#each backlogCards as card (card.id)}
                <PlanningCardRow
                  {card}
                  {availableSprints}
                  onCardClick={handleCardClick}
                  onMoveToSprint={handleMoveToSprint}
                />
              {/each}
            </div>
          {/if}
        </PlanningSection>

        <!-- Closed Sprints -->
        {#if closedSprints.length > 0}
          <PlanningSection
            title="Closed Sprints"
            cardCount={closedSprintsPageInfo?.totalCount ?? closedSprints.length}
            expanded={expandedSections.has('closed')}
            variant="closed"
            onToggle={(expanded) => handleSectionToggle('closed', expanded)}
          >
            <div class="divide-y divide-gray-100">
              {#each closedSprints as sprint (sprint.id)}
                {@const cards = sprintCards.get(sprint.id) ?? []}
                {@const isLoading = loadingSprintCards.has(sprint.id)}
                {@const isExpanded = expandedSections.has(`closed-${sprint.id}`)}
                <div class="py-2">
                  <button
                    type="button"
                    class="w-full px-4 py-2 flex items-center justify-between hover:bg-gray-50"
                    onclick={() => {
                      handleSectionToggle(`closed-${sprint.id}`, !isExpanded);
                      if (!isExpanded) loadSprintCards(sprint.id);
                    }}
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
                      <span class="text-sm font-medium text-gray-700">{sprint.name}</span>
                    </div>
                    {#if isExpanded}
                      <span class="text-xs text-gray-500">{cards.length} cards | {calculateStoryPoints(cards)} pts</span>
                    {/if}
                  </button>
                  {#if isExpanded}
                    <div class="ml-6 border-l border-gray-200">
                      {#if isLoading}
                        <div class="px-4 py-4 text-center text-gray-500 text-sm">Loading...</div>
                      {:else if cards.length === 0}
                        <div class="px-4 py-4 text-center text-gray-500 text-sm">No cards</div>
                      {:else}
                        <div class="divide-y divide-gray-100">
                          {#each cards as card (card.id)}
                            <PlanningCardRow
                              {card}
                              {availableSprints}
                              onCardClick={handleCardClick}
                              onMoveToSprint={handleMoveToSprint}
                            />
                          {/each}
                        </div>
                      {/if}
                    </div>
                  {/if}
                </div>
              {/each}

              {#if closedSprintsPageInfo?.hasNextPage}
                <button
                  type="button"
                  class="w-full px-4 py-3 text-sm text-gray-500 hover:text-gray-700 hover:bg-gray-50"
                  onclick={loadMoreClosedSprints}
                >
                  Load more closed sprints ({closedSprintsPageInfo.totalCount - closedSprints.length} remaining)
                </button>
              {/if}
            </div>
          </PlanningSection>
        {/if}
      </div>
    {/if}
    </div>

    <!-- Card Detail Panel - Always render so ESC key handler works -->
    {#if board}
      <CardDetailPanel
        card={selectedCard ? toBoardCard(selectedCard) : null}
        {projectId}
        organizationId={board.project.organization.id}
        {boardId}
        {tags}
        isOpen={showCardPanel}
        onClose={closeCardPanel}
        onUpdated={handleCardUpdated}
        viewMode={cardViewMode}
        onViewModeChange={setViewMode}
        canEditCard={true}
        canDeleteCard={true}
      />
    {/if}
  </div>
  {/snippet}
</BoardPageLayout>
