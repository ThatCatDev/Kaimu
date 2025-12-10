<script lang="ts">
  interface Props {
    id: string;
    name: string;
    projectKey: string;
    description?: string | null;
    boardCount?: number;
    canDelete?: boolean;
    onDelete?: () => void;
  }

  let { id, name, projectKey, description, boardCount = 0, canDelete = false, onDelete }: Props = $props();
</script>

<div class="relative group bg-white rounded-lg shadow hover:shadow-md transition-shadow duration-200">
  <a href={`/projects/${id}`} class="block p-6">
    <div class="flex items-start justify-between">
      <div class="flex-1 min-w-0">
        <h3 class="text-lg font-semibold text-gray-900 truncate">{name}</h3>
        <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800 mt-1">
          {projectKey}
        </span>
      </div>
    </div>
    {#if description}
      <p class="mt-2 text-sm text-gray-600 line-clamp-2">{description}</p>
    {/if}
    <div class="mt-4 flex items-center text-sm text-gray-500">
      <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
      </svg>
      {boardCount} {boardCount === 1 ? 'board' : 'boards'}
    </div>
  </a>
  {#if canDelete && onDelete}
    <button
      type="button"
      onclick={(e) => { e.preventDefault(); e.stopPropagation(); onDelete(); }}
      class="absolute top-4 right-4 opacity-0 group-hover:opacity-100 p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded transition-all"
      title="Delete project"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
      </svg>
    </button>
  {/if}
</div>
