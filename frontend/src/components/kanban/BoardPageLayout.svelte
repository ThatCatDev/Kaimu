<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { BoardWithColumns } from '../../lib/api/boards';
  import { getCachedBoard, setCachedBoard } from '../../lib/stores/boardCache';

  interface Props {
    board: BoardWithColumns | null;
    boardId: string;
    projectId: string;
    currentPage: 'board' | 'planning' | 'metrics';
    headerActions?: Snippet;
    children: Snippet;
  }

  let { board, boardId, projectId, currentPage, headerActions, children }: Props = $props();

  // Use cached board for instant display, update cache when board prop changes
  let displayBoard = $derived.by(() => {
    if (board) {
      setCachedBoard(boardId, board);
      return board;
    }
    return getCachedBoard(boardId);
  });
</script>

<div class="h-full flex flex-col">
  <!-- Header -->
  <div class="flex-shrink-0 bg-white border-b border-gray-200 px-6 py-4">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <div class="min-w-[120px]">
          {#if displayBoard}
            <h1 class="text-xl font-semibold text-gray-900">{displayBoard.name}</h1>
            {#if displayBoard.description}
              <p class="text-sm text-gray-500">{displayBoard.description}</p>
            {/if}
          {:else}
            <div class="h-7 w-32 bg-gray-200 rounded animate-pulse"></div>
          {/if}
        </div>
        <div class="flex items-center gap-1 bg-gray-100 rounded-lg p-1">
          {#if currentPage === 'board'}
            <span class="px-3 py-1.5 text-sm font-medium rounded-md bg-white text-gray-900 shadow-sm">
              Board
            </span>
          {:else}
            <a
              href={`/projects/${projectId}/board/${boardId}`}
              class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors text-gray-600 hover:text-gray-900 hover:bg-white"
            >
              Board
            </a>
          {/if}
          {#if currentPage === 'planning'}
            <span class="px-3 py-1.5 text-sm font-medium rounded-md bg-white text-gray-900 shadow-sm">
              Planning
            </span>
          {:else}
            <a
              href={`/projects/${projectId}/board/${boardId}/planning`}
              class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors text-gray-600 hover:text-gray-900 hover:bg-white"
            >
              Planning
            </a>
          {/if}
          {#if currentPage === 'metrics'}
            <span class="px-3 py-1.5 text-sm font-medium rounded-md bg-white text-gray-900 shadow-sm">
              Metrics
            </span>
          {:else}
            <a
              href={`/projects/${projectId}/board/${boardId}/metrics`}
              class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors text-gray-600 hover:text-gray-900 hover:bg-white"
            >
              Metrics
            </a>
          {/if}
        </div>
      </div>
      {#if headerActions}
        <div class="flex items-center gap-4">
          {@render headerActions()}
        </div>
      {/if}
    </div>
  </div>

  <!-- Content -->
  <div class="flex-1 overflow-hidden">
    {@render children()}
  </div>
</div>
