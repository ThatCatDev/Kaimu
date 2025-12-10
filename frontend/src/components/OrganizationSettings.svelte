<script lang="ts">
  import { onMount } from 'svelte';
  import { getOrganization } from '../lib/api/organizations';
  import { getMe } from '../lib/api/auth';
  import { MembersList, RolesList } from './settings';
  import type { OrganizationQuery, User } from '../lib/graphql/generated';

  interface Props {
    organizationId: string;
  }

  let { organizationId }: Props = $props();

  type OrgData = NonNullable<OrganizationQuery['organization']>;

  let organization = $state<OrgData | null>(null);
  let user = $state<User | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let activeTab = $state<'members' | 'roles'>('members');

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
      organization = org as OrgData;
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
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <nav class="flex items-center gap-2 text-sm text-gray-500 mb-2">
          <a href={`/organizations/${organizationId}`} class="hover:text-gray-700">{organization.name}</a>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
          <span class="text-gray-700">Settings</span>
        </nav>
        <h1 class="text-2xl font-bold text-gray-900">Organization Settings</h1>
      </div>
      <a
        href={`/organizations/${organizationId}`}
        class="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
      >
        <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
        </svg>
        Back to Organization
      </a>
    </div>

    <!-- Tabs -->
    <div class="border-b border-gray-200">
      <nav class="-mb-px flex space-x-8">
        <button
          type="button"
          onclick={() => activeTab = 'members'}
          class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm {activeTab === 'members' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
        >
          Members
        </button>
        <button
          type="button"
          onclick={() => activeTab = 'roles'}
          class="whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm {activeTab === 'roles' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
        >
          Roles
        </button>
      </nav>
    </div>

    <!-- Tab Content -->
    <div>
      {#if activeTab === 'members'}
        <MembersList {organizationId} />
      {:else if activeTab === 'roles'}
        <RolesList {organizationId} />
      {/if}
    </div>
  </div>
{/if}
