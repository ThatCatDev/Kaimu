<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getRoles,
    getPermissions,
    deleteRole,
    type Role,
    type Permission
  } from '../../lib/api/rbac';
  import { ConfirmModal } from '../ui';
  import RoleEditor from './RoleEditor.svelte';

  interface Props {
    organizationId: string;
  }

  let { organizationId }: Props = $props();

  let roles = $state<Role[]>([]);
  let permissions = $state<Permission[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  // Modal state
  let showEditor = $state(false);
  let showDeleteModal = $state(false);
  let selectedRole = $state<Role | null>(null);
  let templateRole = $state<Role | null>(null);  // Role to base new role on
  let deleting = $state(false);

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      error = null;
      const [rolesData, permissionsData] = await Promise.all([
        getRoles(organizationId),
        getPermissions()
      ]);
      roles = rolesData;
      permissions = permissionsData;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load roles';
    } finally {
      loading = false;
    }
  }

  function handleCreate() {
    selectedRole = null;
    templateRole = null;
    showEditor = true;
  }

  function handleDuplicate(role: Role) {
    selectedRole = null;  // Creating new, not editing
    templateRole = role;  // Use this role as template
    showEditor = true;
  }

  function handleEdit(role: Role) {
    selectedRole = role;
    templateRole = null;
    showEditor = true;
  }

  function handleDelete(role: Role) {
    selectedRole = role;
    showDeleteModal = true;
  }

  async function confirmDelete() {
    if (!selectedRole) return;
    try {
      deleting = true;
      await deleteRole(selectedRole.id);
      await loadData();
      showDeleteModal = false;
      selectedRole = null;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to delete role';
    } finally {
      deleting = false;
    }
  }

  function handleEditorClose() {
    showEditor = false;
    selectedRole = null;
    templateRole = null;
    loadData();
  }

  const systemRoles = $derived(roles.filter(r => r.isSystem));
  const customRoles = $derived(roles.filter(r => !r.isSystem));
</script>

<div class="space-y-6">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium text-gray-900">Roles</h2>
    <button
      type="button"
      onclick={handleCreate}
      class="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
    >
      <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
      </svg>
      Create Role
    </button>
  </div>

  {#if error}
    <div class="rounded-md bg-red-50 p-4">
      <p class="text-sm text-red-700">{error}</p>
    </div>
  {/if}

  {#if loading}
    <div class="space-y-6">
      <div class="space-y-3">
        <div class="h-4 w-24 bg-gray-200 rounded animate-pulse"></div>
        <div class="bg-white shadow-sm rounded-lg border border-gray-200 divide-y divide-gray-200">
          {#each [1, 2] as _}
            <div class="px-4 py-4 flex items-center justify-between">
              <div class="flex-1">
                <div class="flex items-center gap-2 mb-1">
                  <div class="h-4 w-24 bg-gray-200 rounded animate-pulse"></div>
                  <div class="h-5 w-16 bg-gray-200 rounded animate-pulse"></div>
                </div>
                <div class="h-3 w-48 bg-gray-200 rounded animate-pulse mt-2"></div>
              </div>
            </div>
          {/each}
        </div>
      </div>
      <div class="space-y-3">
        <div class="h-4 w-28 bg-gray-200 rounded animate-pulse"></div>
        <div class="bg-white shadow-sm rounded-lg border border-gray-200 divide-y divide-gray-200">
          {#each [1, 2, 3] as _}
            <div class="px-4 py-4 flex items-center justify-between">
              <div class="flex-1">
                <div class="h-4 w-32 bg-gray-200 rounded animate-pulse mb-2"></div>
                <div class="h-3 w-56 bg-gray-200 rounded animate-pulse"></div>
              </div>
              <div class="h-8 w-16 bg-gray-200 rounded animate-pulse"></div>
            </div>
          {/each}
        </div>
      </div>
    </div>
  {:else}
    <!-- System Roles -->
    <div class="space-y-3">
      <h3 class="text-sm font-medium text-gray-700">System Roles</h3>
      <div class="bg-white shadow-sm rounded-lg border border-gray-200 divide-y divide-gray-200">
        {#each systemRoles as role (role.id)}
          <div class="px-4 py-4 flex items-center justify-between">
            <div>
              <div class="flex items-center gap-2">
                <p class="text-sm font-medium text-gray-900">{role.name}</p>
                <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                  System
                </span>
              </div>
              {#if role.description}
                <p class="text-sm text-gray-500 mt-1">{role.description}</p>
              {/if}
              <p class="text-xs text-gray-400 mt-1">
                {role.permissions?.length || 0} permissions
              </p>
            </div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                onclick={() => handleDuplicate(role)}
                class="text-gray-400 hover:text-indigo-600 p-1 rounded hover:bg-indigo-50"
                title="Duplicate role"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                </svg>
              </button>
              <button
                type="button"
                onclick={() => handleEdit(role)}
                class="text-gray-400 hover:text-gray-600 p-1 rounded hover:bg-gray-100"
                title="View permissions"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>

    <!-- Custom Roles -->
    <div class="space-y-3">
      <h3 class="text-sm font-medium text-gray-700">Custom Roles</h3>
      {#if customRoles.length === 0}
        <div class="text-center py-8 bg-white rounded-lg border border-gray-200">
          <svg class="mx-auto h-10 w-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <h3 class="mt-2 text-sm font-medium text-gray-900">No custom roles</h3>
          <p class="mt-1 text-sm text-gray-500">Create custom roles for fine-grained access control.</p>
        </div>
      {:else}
        <div class="bg-white shadow-sm rounded-lg border border-gray-200 divide-y divide-gray-200">
          {#each customRoles as role (role.id)}
            <div class="px-4 py-4 flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-gray-900">{role.name}</p>
                {#if role.description}
                  <p class="text-sm text-gray-500 mt-1">{role.description}</p>
                {/if}
                <p class="text-xs text-gray-400 mt-1">
                  {role.permissions?.length || 0} permissions
                </p>
              </div>
              <div class="flex items-center gap-2">
                <button
                  type="button"
                  onclick={() => handleDuplicate(role)}
                  class="text-gray-400 hover:text-indigo-600 p-1 rounded hover:bg-indigo-50"
                  title="Duplicate role"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
                <button
                  type="button"
                  onclick={() => handleEdit(role)}
                  class="text-gray-400 hover:text-gray-600 p-1 rounded hover:bg-gray-100"
                  title="Edit role"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </button>
                <button
                  type="button"
                  onclick={() => handleDelete(role)}
                  class="text-gray-400 hover:text-red-600 p-1 rounded hover:bg-red-50"
                  title="Delete role"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<RoleEditor
  {organizationId}
  role={selectedRole}
  template={templateRole}
  allRoles={roles}
  {permissions}
  open={showEditor}
  onClose={handleEditorClose}
/>

<ConfirmModal
  isOpen={showDeleteModal}
  title="Delete Role"
  message={`Are you sure you want to delete the "${selectedRole?.name}" role? Members with this role will need to be reassigned.`}
  confirmText={deleting ? 'Deleting...' : 'Delete Role'}
  cancelText="Cancel"
  variant="danger"
  onConfirm={confirmDelete}
  onCancel={() => showDeleteModal = false}
/>
