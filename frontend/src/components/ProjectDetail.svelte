<script lang="ts">
  import { onMount } from 'svelte';
  import { getProject } from '../lib/api/projects';
  import { getMe } from '../lib/api/auth';
  import { getBoards } from '../lib/api/boards';
  import type { ProjectQuery, User, BoardsQuery } from '../lib/graphql/generated';

  interface Props {
    projectId: string;
  }

  let { projectId }: Props = $props();

  type ProjectWithOrg = NonNullable<ProjectQuery['project']>;
  type Board = BoardsQuery['boards'][0];

  let project = $state<ProjectWithOrg | null>(null);
  let user = $state<User | null>(null);
  let boards = $state<Board[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  let defaultBoard = $derived(boards.find(b => b.isDefault) ?? boards[0]);

  onMount(async () => {
    try {
      const [me, proj] = await Promise.all([getMe(), getProject(projectId)]);
      if (!me) {
        window.location.href = '/login';
        return;
      }
      user = me;
      if (!proj) {
        error = 'Project not found';
        return;
      }
      project = proj as ProjectWithOrg;
      boards = await getBoards(projectId);
    } catch (e) {
      if (e instanceof Error && e.message.includes('unauthorized')) {
        window.location.href = '/login';
        return;
      }
      error = e instanceof Error ? e.message : 'Failed to load project';
    } finally {
      loading = false;
    }
  });
</script>

{#if loading}
  <div class="flex items-center justify-center min-h-64">
    <div class="text-gray-500">Loading...</div>
  </div>
{:else if error}
  <div class="rounded-md bg-red-50 p-4">
    <p class="text-sm text-red-700">{error}</p>
  </div>
{:else if project}
  <div class="space-y-8">
    <div>
      <nav class="text-sm text-gray-500 mb-2">
        <a href="/dashboard" class="hover:text-gray-700">Dashboard</a>
        <span class="mx-2">/</span>
        <a href={`/organizations/${project.organization.id}`} class="hover:text-gray-700">{project.organization.name}</a>
        <span class="mx-2">/</span>
        <span class="text-gray-900">{project.name}</span>
      </nav>
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">{project.name}</h1>
          <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800 mt-2">
            {project.key}
          </span>
          {#if project.description}
            <p class="mt-3 text-gray-600">{project.description}</p>
          {/if}
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#if defaultBoard}
        <a
          href={`/projects/${projectId}/board/${defaultBoard.id}`}
          class="bg-white shadow rounded-lg p-6 hover:shadow-md transition-shadow group"
        >
          <div class="flex items-center gap-3 mb-2">
            <div class="p-2 bg-indigo-100 rounded-lg group-hover:bg-indigo-200 transition-colors">
              <svg class="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
              </svg>
            </div>
            <div>
              <h3 class="text-lg font-medium text-gray-900">Kanban Board</h3>
              <p class="text-sm text-gray-500">{defaultBoard.name}</p>
            </div>
          </div>
          <p class="text-sm text-gray-600">
            {defaultBoard.description ?? 'View and manage tasks with drag-and-drop'}
          </p>
        </a>
      {/if}

      {#if boards.length > 1}
        <div class="bg-white shadow rounded-lg p-6">
          <h3 class="text-lg font-medium text-gray-900 mb-3">Other Boards</h3>
          <ul class="space-y-2">
            {#each boards.filter(b => !b.isDefault) as board}
              <li>
                <a
                  href={`/projects/${projectId}/board/${board.id}`}
                  class="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
                >
                  {board.name}
                </a>
              </li>
            {/each}
          </ul>
        </div>
      {/if}
    </div>
  </div>
{/if}
