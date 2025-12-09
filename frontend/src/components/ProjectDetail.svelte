<script lang="ts">
  import { onMount } from 'svelte';
  import { getProject } from '../lib/api/projects';
  import { getMe } from '../lib/api/auth';
  import type { ProjectQuery, User } from '../lib/graphql/generated';

  interface Props {
    projectId: string;
  }

  let { projectId }: Props = $props();

  type ProjectWithOrg = NonNullable<ProjectQuery['project']>;

  let project = $state<ProjectWithOrg | null>(null);
  let user = $state<User | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

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

    <div class="bg-white shadow rounded-lg p-6">
      <h2 class="text-lg font-medium text-gray-900 mb-4">Project Overview</h2>
      <p class="text-gray-500">
        This is where project details, boards, and issues will be displayed.
      </p>
    </div>
  </div>
{/if}
