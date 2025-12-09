<script lang="ts">
  import { onMount } from 'svelte';
  import { getOrganization } from '../lib/api/organizations';
  import { getMe } from '../lib/api/auth';
  import ProjectCard from './ProjectCard.svelte';
  import type { OrganizationQuery, User } from '../lib/graphql/generated';

  interface Props {
    organizationId: string;
  }

  let { organizationId }: Props = $props();

  type OrgWithProjects = NonNullable<OrganizationQuery['organization']>;

  let organization = $state<OrgWithProjects | null>(null);
  let user = $state<User | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    try {
      const [me, org] = await Promise.all([getMe(), getOrganization(organizationId)]);
      if (!me) {
        window.location.href = '/login';
        return;
      }
      user = me;
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
</script>

{#if loading}
  <div class="flex items-center justify-center min-h-64">
    <div class="text-gray-500">Loading...</div>
  </div>
{:else if error}
  <div class="rounded-md bg-red-50 p-4">
    <p class="text-sm text-red-700">{error}</p>
  </div>
{:else if organization}
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <nav class="text-sm text-gray-500 mb-2">
          <a href="/dashboard" class="hover:text-gray-700">Dashboard</a>
          <span class="mx-2">/</span>
          <span class="text-gray-900">{organization.name}</span>
        </nav>
        <h1 class="text-2xl font-bold text-gray-900">{organization.name}</h1>
        <p class="text-sm text-gray-500">/{organization.slug}</p>
        {#if organization.description}
          <p class="mt-2 text-gray-600">{organization.description}</p>
        {/if}
      </div>
      <a
        href={`/organizations/${organizationId}/projects/new`}
        class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
      >
        <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
        New Project
      </a>
    </div>

    <div>
      <h2 class="text-lg font-medium text-gray-900 mb-4">Projects</h2>
      {#if organization.projects.length === 0}
        <div class="text-center py-12 bg-white rounded-lg shadow">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No projects</h3>
          <p class="mt-1 text-sm text-gray-500">Get started by creating a new project.</p>
          <div class="mt-6">
            <a
              href={`/organizations/${organizationId}/projects/new`}
              class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
            >
              <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              New Project
            </a>
          </div>
        </div>
      {:else}
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {#each organization.projects as project (project.id)}
            <ProjectCard
              id={project.id}
              name={project.name}
              projectKey={project.key}
              description={project.description}
            />
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/if}
