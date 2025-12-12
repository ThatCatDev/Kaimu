<script lang="ts">
  import { onMount } from 'svelte';
  import { toast } from 'svelte-sonner';
  import { getOrganization, deleteOrganization, updateOrganization } from '../lib/api/organizations';
  import { deleteProject } from '../lib/api/projects';
  import { getMe } from '../lib/api/auth';
  import ProjectCard from './ProjectCard.svelte';
  import EditableTitle from './EditableTitle.svelte';
  import { ConfirmModal } from './ui';
  import type { OrganizationQuery, User } from '../lib/graphql/generated';
  import { Permissions } from '../lib/stores/permissions.svelte';
  import { getMyPermissions } from '../lib/api/rbac';
  import { sidebarStore } from '../lib/stores/sidebar.svelte';

  interface Props {
    organizationId: string;
  }

  let { organizationId }: Props = $props();

  // Permissions - loaded client-side
  let permissions = $state<string[]>([]);

  let canManageOrg = $derived(permissions.includes(Permissions.ORG_MANAGE));
  let canDeleteOrg = $derived(permissions.includes(Permissions.ORG_DELETE));
  let canCreateProject = $derived(permissions.includes(Permissions.PROJECT_CREATE));
  let canDeleteProject = $derived(permissions.includes(Permissions.PROJECT_DELETE));

  type OrgWithProjects = NonNullable<OrganizationQuery['organization']>;

  let organization = $state<OrgWithProjects | null>(null);
  let user = $state<User | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let showDeleteModal = $state(false);
  let deleting = $state(false);

  // Project deletion state
  type Project = OrgWithProjects['projects'][0];
  let showDeleteProjectModal = $state(false);
  let projectToDelete = $state<Project | null>(null);
  let deletingProject = $state(false);

  onMount(async () => {
    try {
      const [me, org, perms] = await Promise.all([
        getMe(),
        getOrganization(organizationId),
        getMyPermissions('organization', organizationId)
      ]);
      if (!me) {
        window.location.href = '/login';
        return;
      }
      user = me;
      permissions = perms;
      if (!org) {
        error = 'Organization not found';
        return;
      }
      organization = org as OrgWithProjects;
    } catch (e) {
      if (e instanceof Error && e.message.includes('unauthorized')) {
        window.location.href = '/login';
        return;
      }
      error = e instanceof Error ? e.message : 'Failed to load organization';
    } finally {
      loading = false;
    }
  });

  async function handleDelete() {
    try {
      deleting = true;
      await deleteOrganization(organizationId);
      sidebarStore.refresh();
      toast.success('Organization deleted');
      window.location.href = '/dashboard';
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to delete organization';
      toast.error(message);
      showDeleteModal = false;
    } finally {
      deleting = false;
    }
  }

  async function handleRename(newName: string) {
    if (!organization) return;
    const updated = await updateOrganization(organizationId, { name: newName });
    organization = { ...organization, name: updated.name, slug: updated.slug };
    sidebarStore.refresh();
  }

  function openDeleteProjectModal(project: Project) {
    projectToDelete = project;
    showDeleteProjectModal = true;
  }

  async function handleDeleteProject() {
    if (!projectToDelete || !organization) return;
    try {
      deletingProject = true;
      await deleteProject(projectToDelete.id);
      // Remove the project from the local list
      organization = {
        ...organization,
        projects: organization.projects.filter(p => p.id !== projectToDelete!.id)
      };
      showDeleteProjectModal = false;
      projectToDelete = null;
      sidebarStore.refresh();
      toast.success('Project deleted');
    } catch (e) {
      const message = e instanceof Error ? e.message : 'Failed to delete project';
      toast.error(message);
    } finally {
      deletingProject = false;
    }
  }
</script>

{#if loading}
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <div class="h-8 w-48 bg-gray-200 rounded animate-pulse"></div>
        <div class="mt-1 h-4 w-24 bg-gray-200 rounded animate-pulse"></div>
        <div class="mt-2 h-4 w-64 bg-gray-200 rounded animate-pulse"></div>
      </div>
      <div class="flex items-center gap-4">
        <div class="h-10 w-24 bg-gray-200 rounded-md animate-pulse"></div>
        <div class="h-10 w-32 bg-gray-200 rounded-md animate-pulse"></div>
      </div>
    </div>
    <div>
      <div class="h-6 w-24 bg-gray-200 rounded animate-pulse mb-4"></div>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {#each [1, 2, 3] as _}
          <div class="bg-white rounded-lg shadow p-6">
            <div class="h-5 w-3/4 bg-gray-200 rounded animate-pulse mb-2"></div>
            <div class="h-4 w-16 bg-gray-200 rounded animate-pulse mb-4"></div>
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
{:else if organization}
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">
          {#if canManageOrg}
            <EditableTitle value={organization.name} onSave={handleRename} />
          {:else}
            {organization.name}
          {/if}
        </h1>
        <p class="text-sm text-gray-500">/{organization.slug}</p>
        {#if organization.description}
          <p class="mt-2 text-gray-600">{organization.description}</p>
        {/if}
      </div>
      <div class="flex items-center gap-4">
        {#if canManageOrg}
          <a
            href={`/organizations/${organizationId}/settings`}
            class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Settings
          </a>
        {/if}
        {#if canDeleteOrg}
          <button
            type="button"
            onclick={() => showDeleteModal = true}
            class="inline-flex items-center px-4 py-2 border border-red-300 text-sm font-medium rounded-md text-red-700 bg-white hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
          >
            <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Delete
          </button>
        {/if}
        {#if canCreateProject && organization.projects.length > 0}
          <a
            href={`/organizations/${organizationId}/projects/new`}
            class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            New Project
          </a>
        {/if}
      </div>
    </div>

    <div>
      <h2 class="text-lg font-medium text-gray-900 mb-4">Projects</h2>
      {#if organization.projects.length === 0}
        <div class="text-center py-12 bg-white rounded-lg shadow">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No projects</h3>
          {#if canCreateProject}
            <p class="mt-1 text-sm text-gray-500">Get started by creating a new project.</p>
            <div class="mt-6">
              <a
                href={`/organizations/${organizationId}/projects/new`}
                class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
              >
                <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                </svg>
                New Project
              </a>
            </div>
          {:else}
            <p class="mt-1 text-sm text-gray-500">No projects have been created yet.</p>
          {/if}
        </div>
      {:else}
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {#each organization.projects as project (project.id)}
            <ProjectCard
              id={project.id}
              name={project.name}
              projectKey={project.key}
              description={project.description}
              boardCount={project.boards?.length ?? 0}
              canDelete={canDeleteProject}
              onDelete={() => openDeleteProjectModal(project)}
            />
          {/each}
        </div>
      {/if}
    </div>
  </div>

  <ConfirmModal
    isOpen={showDeleteModal}
    title="Delete Organization"
    message="Are you sure you want to delete this organization? This will permanently delete all projects, boards, and cards within it. This action cannot be undone."
    confirmText={deleting ? 'Deleting...' : 'Delete Organization'}
    cancelText="Cancel"
    variant="danger"
    onConfirm={handleDelete}
    onCancel={() => showDeleteModal = false}
  />

  <ConfirmModal
    isOpen={showDeleteProjectModal}
    title="Delete Project"
    message={`Are you sure you want to delete "${projectToDelete?.name}"? This will permanently delete all boards and cards in this project. This action cannot be undone.`}
    confirmText={deletingProject ? 'Deleting...' : 'Delete Project'}
    cancelText="Cancel"
    variant="danger"
    onConfirm={handleDeleteProject}
    onCancel={() => { showDeleteProjectModal = false; projectToDelete = null; }}
  />
{/if}
