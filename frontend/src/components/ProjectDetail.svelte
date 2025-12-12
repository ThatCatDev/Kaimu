<script lang="ts">
  import { onMount } from 'svelte';
  import { toast } from 'svelte-sonner';
  import { getProject, deleteProject, updateProject } from '../lib/api/projects';
  import { getMe } from '../lib/api/auth';
  import { getBoards, createBoard, deleteBoard } from '../lib/api/boards';
  import EditableTitle from './EditableTitle.svelte';
  import { ConfirmModal, Button, Input, Textarea } from './ui';
  import type { ProjectQuery, User, BoardsQuery } from '../lib/graphql/generated';
  import { Permissions } from '../lib/stores/permissions.svelte';
  import { getMyPermissions } from '../lib/api/rbac';
  import { sidebarStore } from '../lib/stores/sidebar.svelte';

  interface Props {
    projectId: string;
  }

  let { projectId }: Props = $props();

  // Permissions - loaded client-side
  let permissions = $state<string[]>([]);

  let canManageProject = $derived(permissions.includes(Permissions.PROJECT_MANAGE));
  let canDeleteProject = $derived(permissions.includes(Permissions.PROJECT_DELETE));
  let canCreateBoard = $derived(permissions.includes(Permissions.BOARD_CREATE));
  let canDeleteBoard = $derived(permissions.includes(Permissions.BOARD_DELETE));

  type ProjectWithOrg = NonNullable<ProjectQuery['project']>;
  type Board = BoardsQuery['boards'][0];

  let project = $state<ProjectWithOrg | null>(null);
  let user = $state<User | null>(null);
  let boards = $state<Board[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  // Delete project modal state
  let showDeleteProjectModal = $state(false);
  let deletingProject = $state(false);

  // Create board modal state
  let showCreateBoardModal = $state(false);
  let newBoardName = $state('');
  let newBoardDescription = $state('');
  let creatingBoard = $state(false);

  // Delete board modal state
  let showDeleteBoardModal = $state(false);
  let boardToDelete = $state<Board | null>(null);
  let deletingBoard = $state(false);

  onMount(async () => {
    try {
      const [me, proj, perms] = await Promise.all([
        getMe(),
        getProject(projectId),
        getMyPermissions('project', projectId)
      ]);
      if (!me) {
        window.location.href = '/login';
        return;
      }
      user = me;
      permissions = perms;
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

  async function handleDeleteProject() {
    if (!project) return;
    try {
      deletingProject = true;
      await deleteProject(projectId);
      sidebarStore.refresh();
      toast.success('Project deleted');
      window.location.href = `/organizations/${project.organization.id}`;
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to delete project';
      toast.error(message);
      showDeleteProjectModal = false;
    } finally {
      deletingProject = false;
    }
  }

  async function handleCreateBoard() {
    if (!newBoardName.trim()) return;
    try {
      creatingBoard = true;
      await createBoard(projectId, newBoardName.trim(), newBoardDescription.trim() || undefined);
      boards = await getBoards(projectId);
      showCreateBoardModal = false;
      newBoardName = '';
      newBoardDescription = '';
      sidebarStore.refresh();
      toast.success('Board created');
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to create board';
      toast.error(message);
    } finally {
      creatingBoard = false;
    }
  }

  async function handleDeleteBoard() {
    if (!boardToDelete) return;
    try {
      deletingBoard = true;
      await deleteBoard(boardToDelete.id);
      boards = await getBoards(projectId);
      showDeleteBoardModal = false;
      boardToDelete = null;
      sidebarStore.refresh();
      toast.success('Board deleted');
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to delete board';
      toast.error(message);
    } finally {
      deletingBoard = false;
    }
  }

  function openDeleteBoardModal(board: Board) {
    boardToDelete = board;
    showDeleteBoardModal = true;
  }

  async function handleRename(newName: string) {
    if (!project) return;
    const updated = await updateProject(projectId, { name: newName });
    project = { ...project, name: updated.name };
    sidebarStore.refresh();
  }
</script>

{#if loading}
  <div class="space-y-8">
    <div>
      <div class="flex items-center gap-2 text-sm mb-2">
        <div class="h-4 w-24 bg-gray-200 rounded animate-pulse"></div>
        <span class="text-gray-400">/</span>
        <div class="h-4 w-20 bg-gray-200 rounded animate-pulse"></div>
      </div>
      <div class="flex items-start justify-between">
        <div>
          <div class="h-8 w-48 bg-gray-200 rounded animate-pulse"></div>
          <div class="h-6 w-16 bg-gray-200 rounded-full animate-pulse mt-2"></div>
          <div class="h-4 w-64 bg-gray-200 rounded animate-pulse mt-4"></div>
        </div>
        <div class="flex items-center gap-4">
          <div class="h-10 w-24 bg-gray-200 rounded-md animate-pulse"></div>
          <div class="h-10 w-28 bg-gray-200 rounded-md animate-pulse"></div>
        </div>
      </div>
    </div>
    <div>
      <div class="h-6 w-32 bg-gray-200 rounded animate-pulse mb-4"></div>
      <div class="grid gap-4 sm:grid-cols-2">
        {#each [1, 2] as _}
          <div class="bg-white rounded-lg shadow p-6">
            <div class="flex items-center gap-3 mb-3">
              <div class="h-10 w-10 bg-gray-200 rounded animate-pulse"></div>
              <div class="h-5 w-32 bg-gray-200 rounded animate-pulse"></div>
            </div>
            <div class="h-4 w-full bg-gray-200 rounded animate-pulse"></div>
          </div>
        {/each}
      </div>
    </div>
  </div>
{:else if error}
  <div class="rounded-md bg-red-50 p-4">
    <p class="text-sm text-red-700">{error}</p>
  </div>
{:else if project}
  <div class="space-y-8">
    <div>
      <div class="text-sm text-gray-500 mb-2">
        <a href={`/organizations/${project.organization.id}`} class="hover:text-indigo-600 hover:underline">
          {project.organization.name}
        </a>
        <span class="mx-2">/</span>
        <span class="text-gray-900">{project.name}</span>
      </div>
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">
            {#if canManageProject}
              <EditableTitle value={project.name} onSave={handleRename} />
            {:else}
              {project.name}
            {/if}
          </h1>
          <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800 mt-2">
            {project.key}
          </span>
          {#if project.description}
            <p class="mt-4 text-gray-600">{project.description}</p>
          {/if}
        </div>
        <div class="flex items-center gap-4">
          {#if canDeleteProject}
            <button
              type="button"
              onclick={() => showDeleteProjectModal = true}
              class="inline-flex items-center px-4 py-2 border border-red-300 text-sm font-medium rounded-md text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
            >
              <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              Delete
            </button>
          {/if}
          {#if canCreateBoard}
            <button
              type="button"
              onclick={() => showCreateBoardModal = true}
              class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              New Board
            </button>
          {/if}
        </div>
      </div>
    </div>

    <div>
      <h2 class="text-lg font-medium text-gray-900 mb-4">Boards</h2>
      {#if boards.length === 0}
        <div class="text-center py-12 bg-white rounded-lg shadow">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No boards</h3>
          {#if canCreateBoard}
            <p class="mt-1 text-sm text-gray-500">Get started by creating a new board.</p>
            <div class="mt-6">
              <button
                type="button"
                onclick={() => showCreateBoardModal = true}
                class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
              >
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
                New Board
              </button>
            </div>
          {:else}
            <p class="mt-1 text-sm text-gray-500">No boards have been created yet.</p>
          {/if}
        </div>
      {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {#each boards as board (board.id)}
            <div class="bg-white shadow rounded-lg hover:shadow-md transition-shadow group relative">
              <a
                href={`/projects/${projectId}/board/${board.id}`}
                class="block p-6"
              >
                <div class="flex items-start justify-between">
                  <div class="flex items-center gap-3">
                    <div class="p-2 bg-indigo-100 rounded-lg group-hover:bg-indigo-200 transition-colors flex-shrink-0">
                      <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
                      </svg>
                    </div>
                    <div class="min-w-0">
                      <h3 class="text-base font-semibold text-gray-900 truncate">{board.name}</h3>
                    </div>
                  </div>
                </div>
                {#if board.description}
                  <p class="mt-3 text-sm text-gray-600 line-clamp-2">{board.description}</p>
                {/if}
              </a>
              {#if canDeleteBoard}
                <button
                  type="button"
                  onclick={(e) => { e.preventDefault(); e.stopPropagation(); openDeleteBoardModal(board); }}
                  class="absolute top-4 right-4 opacity-0 group-hover:opacity-100 p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded transition-all"
                  title="Delete board"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>

  <!-- Delete Project Modal -->
  <ConfirmModal
    isOpen={showDeleteProjectModal}
    title="Delete Project"
    message="Are you sure you want to delete this project? This will permanently delete all boards and cards within it. This action cannot be undone."
    confirmText={deletingProject ? 'Deleting...' : 'Delete Project'}
    cancelText="Cancel"
    variant="danger"
    onConfirm={handleDeleteProject}
    onCancel={() => showDeleteProjectModal = false}
  />

  <!-- Create Board Modal -->
  {#if showCreateBoardModal}
    <div class="fixed inset-0 bg-gray-900/60 backdrop-blur-sm z-50">
      <div class="fixed inset-0 flex items-center justify-center p-4">
        <div class="bg-white rounded-xl shadow-2xl max-w-md w-full">
          <form onsubmit={(e) => { e.preventDefault(); handleCreateBoard(); }}>
            <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h2 class="text-lg font-semibold text-gray-900">Create Board</h2>
              <button
                type="button"
                class="text-gray-400 hover:text-gray-600 transition-colors"
                onclick={() => showCreateBoardModal = false}
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div class="px-6 py-4 space-y-4">
              <Input
                id="boardName"
                label="Board Name"
                bind:value={newBoardName}
                placeholder="Enter board name"
                required
              />

              <Textarea
                id="boardDescription"
                label="Description"
                bind:value={newBoardDescription}
                rows={3}
                placeholder="Add a description (optional)"
              />
            </div>

            <div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-4">
              <Button variant="secondary" onclick={() => showCreateBoardModal = false} disabled={creatingBoard}>
                Cancel
              </Button>
              <Button type="submit" loading={creatingBoard}>
                {creatingBoard ? 'Creating...' : 'Create Board'}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  {/if}

  <!-- Delete Board Modal -->
  <ConfirmModal
    isOpen={showDeleteBoardModal}
    title="Delete Board"
    message={`Are you sure you want to delete "${boardToDelete?.name}"? This will permanently delete all columns and cards in this board. This action cannot be undone.`}
    confirmText={deletingBoard ? 'Deleting...' : 'Delete Board'}
    cancelText="Cancel"
    variant="danger"
    onConfirm={handleDeleteBoard}
    onCancel={() => { showDeleteBoardModal = false; boardToDelete = null; }}
  />
{/if}
