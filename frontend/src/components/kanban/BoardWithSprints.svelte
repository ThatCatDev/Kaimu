<script lang="ts">
  import { onMount } from 'svelte';
  import KanbanBoard from './KanbanBoard.svelte';
  import BoardPageLayout from './BoardPageLayout.svelte';
  import { SprintSelector } from '../sprint';
  import { getBoard, type BoardWithColumns } from '../../lib/api/boards';

  interface Props {
    boardId: string;
    projectId: string;
    initialCardId?: string | null;
  }

  let { boardId, projectId, initialCardId }: Props = $props();

  let board = $state<BoardWithColumns | null>(null);
  let kanbanBoardRef = $state<KanbanBoard | null>(null);
  let sprintSelectorRef = $state<SprintSelector | null>(null);

  onMount(async () => {
    board = await getBoard(boardId);
  });

  function handleBoardRefresh() {
    kanbanBoardRef?.refreshBoard?.();
    sprintSelectorRef?.refresh?.();
  }
</script>

<BoardPageLayout {board} {boardId} {projectId} currentPage="board">
  {#snippet headerActions()}
    <SprintSelector
      bind:this={sprintSelectorRef}
      {boardId}
      onSprintChange={handleBoardRefresh}
    />
  {/snippet}
  {#snippet children()}
    <div class="max-w-full mx-auto px-6 h-full">
      <KanbanBoard
        bind:this={kanbanBoardRef}
        {boardId}
        {initialCardId}
      />
    </div>
  {/snippet}
</BoardPageLayout>
