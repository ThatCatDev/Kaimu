<script lang="ts">
  import { onMount } from 'svelte';
  import { getOrganizations } from '../lib/api/organizations';
  import { getMe } from '../lib/api/auth';
  import OrganizationCard from './OrganizationCard.svelte';
  import type { OrganizationsQuery, User } from '../lib/graphql/generated';

  type OrganizationListItem = OrganizationsQuery['organizations'][number];

  let organizations = $state<OrganizationListItem[]>([]);
  let user = $state<User | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    try {
      const [me, orgs] = await Promise.all([getMe(), getOrganizations()]);
      if (!me) {
        window.location.href = '/login';
        return;
      }
      user = me;
      organizations = orgs;
    } catch (e) {
      if (e instanceof Error && e.message.includes('unauthorized')) {
        window.location.href = '/login';
        return;
      }
      error = e instanceof Error ? e.message : 'Failed to load dashboard';
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
{:else}
  <div class="space-y-8">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p class="mt-1 text-sm text-gray-500">Welcome back, {user?.username}</p>
      </div>
      <a
        href="/organizations/new"
        class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
      >
        <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
        New Organization
      </a>
    </div>

    <div>
      <h2 class="text-lg font-medium text-gray-900 mb-4">Your Organizations</h2>
      {#if organizations.length === 0}
        <div class="text-center py-12 bg-white rounded-lg shadow">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No organizations</h3>
          <p class="mt-1 text-sm text-gray-500">Get started by creating a new organization.</p>
          <div class="mt-6">
            <a
              href="/organizations/new"
              class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
              </svg>
              New Organization
            </a>
          </div>
        </div>
      {:else}
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {#each organizations as org (org.id)}
            <OrganizationCard
              id={org.id}
              name={org.name}
              slug={org.slug}
              description={org.description}
              projectCount={org.projects?.length ?? 0}
            />
          {/each}
        </div>
      {/if}
    </div>
  </div>
{/if}
