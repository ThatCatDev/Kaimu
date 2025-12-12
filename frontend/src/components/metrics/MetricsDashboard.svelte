<script lang="ts">
  import { onMount } from 'svelte';
  import { toast } from 'svelte-sonner';
  import MetricModeToggle from './MetricModeToggle.svelte';
  import SprintSelector from './SprintSelector.svelte';
  import BurnDownChart from './BurnDownChart.svelte';
  import BurnUpChart from './BurnUpChart.svelte';
  import VelocityChart from './VelocityChart.svelte';
  import CumulativeFlowChart from './CumulativeFlowChart.svelte';
  import BoardPageLayout from '../kanban/BoardPageLayout.svelte';
  import {
    getBurnDownData,
    getBurnUpData,
    getVelocityData,
    getCumulativeFlowData,
    getSprintStats,
    type BurnDownData,
    type BurnUpData,
    type VelocityData,
    type CumulativeFlowData,
    type SprintStats,
    type MetricMode,
  } from '../../lib/api/metrics';
  import { getBacklogCards, type BacklogCard } from '../../lib/api/sprints';
  import { getBoard, type BoardWithColumns } from '../../lib/api/boards';

  interface Props {
    boardId: string;
    projectId: string;
  }

  let { boardId, projectId }: Props = $props();

  let board = $state<BoardWithColumns | null>(null);
  let mode = $state<MetricMode>('CARD_COUNT');
  let selectedSprintId = $state<string | null>(null);

  let burnDownData = $state<BurnDownData | null>(null);
  let burnUpData = $state<BurnUpData | null>(null);
  let velocityData = $state<VelocityData | null>(null);
  let cumulativeFlowData = $state<CumulativeFlowData | null>(null);
  let sprintStats = $state<SprintStats | null>(null);

  // Backlog stats
  let backlogCards = $state<BacklogCard[]>([]);
  let backlogLoading = $state(false);

  let loading = $state(false);

  // Computed backlog stats
  const backlogStats = $derived({
    totalCards: backlogCards.length,
    totalStoryPoints: backlogCards.reduce((sum, card) => sum + (card.storyPoints ?? 0), 0),
  });

  async function loadSprintMetrics() {
    if (!selectedSprintId) {
      burnDownData = null;
      burnUpData = null;
      cumulativeFlowData = null;
      sprintStats = null;
      return;
    }

    loading = true;
    try {
      const [burnDown, burnUp, cumFlow, stats] = await Promise.all([
        getBurnDownData(selectedSprintId, mode),
        getBurnUpData(selectedSprintId, mode),
        getCumulativeFlowData(selectedSprintId, mode),
        getSprintStats(selectedSprintId),
      ]);

      burnDownData = burnDown;
      burnUpData = burnUp;
      cumulativeFlowData = cumFlow;
      sprintStats = stats;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load metrics';
      toast.error(message);
    } finally {
      loading = false;
    }
  }

  async function loadVelocityData() {
    try {
      velocityData = await getVelocityData(boardId, mode, 10);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load velocity data';
      toast.error(message);
    }
  }

  function handleSprintChange(sprintId: string | null) {
    selectedSprintId = sprintId;
    loadSprintMetrics();
  }

  function handleModeChange(newMode: MetricMode) {
    mode = newMode;
    loadSprintMetrics();
    loadVelocityData();
  }

  async function loadBacklogCards() {
    backlogLoading = true;
    try {
      backlogCards = await getBacklogCards(boardId);
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to load backlog';
      toast.error(message);
    } finally {
      backlogLoading = false;
    }
  }

  onMount(async () => {
    board = await getBoard(boardId);
    loadVelocityData();
    loadBacklogCards();
  });

  function formatPercent(value: number, total: number): string {
    if (total === 0) return '0%';
    return `${Math.round((value / total) * 100)}%`;
  }
</script>

<BoardPageLayout {board} {boardId} {projectId} currentPage="metrics">
  {#snippet headerActions()}
    <SprintSelector
      {boardId}
      {selectedSprintId}
      onSprintChange={handleSprintChange}
    />
    <MetricModeToggle {mode} onModeChange={handleModeChange} />
  {/snippet}
  {#snippet children()}
  <div class="h-full flex flex-col bg-gray-50">
    <div class="flex-1 overflow-auto p-6">
    {#if loading}
      <div class="space-y-6">
        <div>
          <div class="h-4 w-16 bg-gray-200 rounded animate-pulse mb-3"></div>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            {#each [1, 2, 3, 4] as _}
              <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
                <div class="h-4 w-24 bg-gray-200 rounded animate-pulse mb-2"></div>
                <div class="h-8 w-16 bg-gray-200 rounded animate-pulse"></div>
              </div>
            {/each}
          </div>
        </div>
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {#each [1, 2, 3, 4] as _}
            <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
              <div class="h-4 w-24 bg-gray-200 rounded animate-pulse mb-3"></div>
              <div class="h-64 bg-gray-100 rounded animate-pulse"></div>
            </div>
          {/each}
        </div>
      </div>
    {:else}
      <!-- Backlog Stats (always visible) -->
      <div class="mb-6">
        <h2 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">Backlog</h2>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
            <div class="text-sm text-gray-500">Backlog Cards</div>
            <div class="text-2xl font-semibold text-gray-900">
              {backlogLoading ? '...' : backlogStats.totalCards}
            </div>
          </div>
          <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
            <div class="text-sm text-gray-500">Backlog Points</div>
            <div class="text-2xl font-semibold text-gray-900">
              {backlogLoading ? '...' : backlogStats.totalStoryPoints}
            </div>
          </div>
          {#if sprintStats && selectedSprintId}
            <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
              <div class="text-sm text-gray-500">Total Planned</div>
              <div class="text-2xl font-semibold text-indigo-600">
                {sprintStats.totalCards + backlogStats.totalCards} cards
              </div>
            </div>
            <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
              <div class="text-sm text-gray-500">Planning Ratio</div>
              <div class="text-2xl font-semibold text-gray-900">
                {formatPercent(sprintStats.totalCards, sprintStats.totalCards + backlogStats.totalCards)} in sprint
              </div>
            </div>
          {/if}
        </div>
      </div>

      <!-- Sprint Stats Summary -->
      {#if sprintStats && selectedSprintId}
        <div class="mb-6">
          <h2 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">Sprint Progress</h2>
          <div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
            <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
              <div class="text-sm text-gray-500">Sprint Cards</div>
              <div class="text-2xl font-semibold text-gray-900">{sprintStats.totalCards}</div>
            </div>
            <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
              <div class="text-sm text-gray-500">Completed Cards</div>
              <div class="text-2xl font-semibold text-green-600">
                {sprintStats.completedCards}
                <span class="text-sm text-gray-400">
                  ({formatPercent(sprintStats.completedCards, sprintStats.totalCards)})
                </span>
              </div>
            </div>
            <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
              <div class="text-sm text-gray-500">Sprint Points</div>
              <div class="text-2xl font-semibold text-gray-900">{sprintStats.totalStoryPoints}</div>
            </div>
            <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
              <div class="text-sm text-gray-500">Completed Points</div>
              <div class="text-2xl font-semibold text-green-600">
                {sprintStats.completedStoryPoints}
                <span class="text-sm text-gray-400">
                  ({formatPercent(sprintStats.completedStoryPoints, sprintStats.totalStoryPoints)})
                </span>
              </div>
            </div>
            <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
              <div class="text-sm text-gray-500">Days Elapsed</div>
              <div class="text-2xl font-semibold text-gray-900">{sprintStats.daysElapsed}</div>
            </div>
            <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
              <div class="text-sm text-gray-500">Days Remaining</div>
              <div class="text-2xl font-semibold {sprintStats.daysRemaining <= 2 ? 'text-orange-600' : 'text-gray-900'}">
                {sprintStats.daysRemaining}
              </div>
            </div>
          </div>
        </div>
      {/if}

      <!-- Charts Grid -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Burn Down Chart -->
        {#if burnDownData}
          <BurnDownChart data={burnDownData} {mode} />
        {:else if selectedSprintId}
          <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
            <h3 class="text-sm font-medium text-gray-700 mb-3">Burn Down</h3>
            <div class="h-64 flex items-center justify-center text-gray-400">
              No data available
            </div>
          </div>
        {/if}

        <!-- Burn Up Chart -->
        {#if burnUpData}
          <BurnUpChart data={burnUpData} {mode} />
        {:else if selectedSprintId}
          <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
            <h3 class="text-sm font-medium text-gray-700 mb-3">Burn Up</h3>
            <div class="h-64 flex items-center justify-center text-gray-400">
              No data available
            </div>
          </div>
        {/if}

        <!-- Cumulative Flow Chart -->
        {#if cumulativeFlowData}
          <CumulativeFlowChart data={cumulativeFlowData} {mode} />
        {:else if selectedSprintId}
          <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
            <h3 class="text-sm font-medium text-gray-700 mb-3">Cumulative Flow</h3>
            <div class="h-64 flex items-center justify-center text-gray-400">
              No data available
            </div>
          </div>
        {/if}

        <!-- Velocity Chart (always visible) -->
        {#if velocityData && velocityData.sprints.length > 0}
          <VelocityChart data={velocityData} {mode} />
        {:else}
          <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
            <h3 class="text-sm font-medium text-gray-700 mb-3">Velocity</h3>
            <div class="h-64 flex items-center justify-center text-gray-400">
              No completed sprints yet
            </div>
          </div>
        {/if}
      </div>

      {#if !selectedSprintId}
        <div class="text-center py-12 text-gray-500">
          Select a sprint to view detailed metrics
        </div>
      {/if}
    {/if}
    </div>
  </div>
  {/snippet}
</BoardPageLayout>
