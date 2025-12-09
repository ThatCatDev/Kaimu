<script lang="ts">
  import { onMount } from 'svelte';
  import { getProject, deleteProject, updateProject } from '../lib/api/projects';
  import { getMe } from '../lib/api/auth';
  import { getBoards, createBoard, deleteBoard } from '../lib/api/boards';
  import EditableTitle from './EditableTitle.svelte';
  import { ConfirmModal, Button, Input, Textarea } from './ui';
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

  async function handleDeleteProject() {
    if (!project) return;
    try {
      deletingProject = true;
      await deleteProject(projectId);
      window.location.href = `/organizations/${project.organization.id}`;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete project';
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
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to create board';
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
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete board';
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
  }
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
            <EditableTitle value={project.name} onSave={handleRename} />
          </h1>
          <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800 mt-2">
            {project.key}
          </span>
          {#if project.description}
            <p class="mt-4 text-gray-600">{project.description}</p>
          {/if}
        </div>
        <div class="flex items-center gap-4">
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
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {#if defaultBoard}
        <a
          href={`/projects/${projectId}/board/${defaultBoard.id}`}
          class="bg-white shadow rounded-lg p-6 hover:shadow-md transition-shadow group"
        >
          <div class="flex items-center gap-4 mb-2">
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
          <h3 class="text-lg font-medium text-gray-900 mb-4">Other Boards</h3>
          <ul class="space-y-2">
            {#each boards.filter(b => !b.isDefault) as board}
              <li class="flex items-center justify-between group">
                <a
                  href={`/projects/${projectId}/board/${board.id}`}
                  class="text-indigo-600 hover:text-indigo-800 text-sm font-medium"
                >
                  {board.name}
                </a>
                <button
                  type="button"
                  onclick={() => openDeleteBoardModal(board)}
                  class="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-red-600 transition-opacity"
                  title="Delete board"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </li>
            {/each}
          </ul>
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
