<script lang="ts">
  import KanbanBoard from './KanbanBoard.svelte';
  import { SprintSelector } from '../sprint';

  interface Props {
    boardId: string;
    initialCardId?: string | null;
  }

  let { boardId, initialCardId }: Props = $props();

  let kanbanBoardRef = $state<KanbanBoard | null>(null);
  let sprintSelectorRef = $state<SprintSelector | null>(null);

  function handleBoardRefresh() {
    kanbanBoardRef?.refreshBoard?.();
    sprintSelectorRef?.refresh?.();
  }
</script>

<div class="h-full flex flex-col">
  <!-- Sprint selector in header -->
  <div class="flex-shrink-0 px-4 sm:px-6 lg:px-8 pt-4 pb-2">
    <div class="flex items-center gap-4">
      <SprintSelector
        bind:this={sprintSelectorRef}
        {boardId}
        onSprintChange={handleBoardRefresh}
      />
    </div>
  </div>

  <!-- Main board area -->
  <div class="flex-1 overflow-hidden">
    <div class="max-w-full mx-auto px-4 sm:px-6 lg:px-8 h-full">
      <KanbanBoard
        bind:this={kanbanBoardRef}
        {boardId}
        {initialCardId}
      />
    </div>
  </div>
</div>
