<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getOrganizationMembers,
    getRoles,
    changeMemberRole,
    removeMember,
    type OrganizationMember,
    type Role
  } from '../../lib/api/rbac';
  import { getMe } from '../../lib/api/auth';
  import { ConfirmModal, Modal, BitsSelect } from '../ui';
  import type { User } from '../../lib/graphql/generated';
  import MemberRow from './MemberRow.svelte';
  import InviteMemberModal from './InviteMemberModal.svelte';

  interface Props {
    organizationId: string;
  }

  let { organizationId }: Props = $props();

  let members = $state<OrganizationMember[]>([]);
  let roles = $state<Role[]>([]);
  let currentUser = $state<User | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  // Modal state
  let showInviteModal = $state(false);
  let showChangeRoleModal = $state(false);
  let showRemoveModal = $state(false);
  let selectedMember = $state<OrganizationMember | null>(null);
  let newRoleId = $state('');
  let processing = $state(false);

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      error = null;
      const [me, membersData, rolesData] = await Promise.all([
        getMe(),
        getOrganizationMembers(organizationId),
        getRoles(organizationId)
      ]);
      currentUser = me;
      members = membersData;
      roles = rolesData;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load members';
    } finally {
      loading = false;
    }
  }

  function handleChangeRole(member: OrganizationMember) {
    selectedMember = member;
    newRoleId = member.role?.id ?? '';
    showChangeRoleModal = true;
  }

  function handleRemove(member: OrganizationMember) {
    selectedMember = member;
    showRemoveModal = true;
  }

  async function confirmChangeRole() {
    if (!selectedMember || !newRoleId) return;
    try {
      processing = true;
      await changeMemberRole(organizationId, selectedMember.user.id, newRoleId);
      await loadData();
      showChangeRoleModal = false;
      selectedMember = null;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to change role';
    } finally {
      processing = false;
    }
  }

  async function confirmRemove() {
    if (!selectedMember) return;
    try {
      processing = true;
      await removeMember(organizationId, selectedMember.user.id);
      await loadData();
      showRemoveModal = false;
      selectedMember = null;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to remove member';
    } finally {
      processing = false;
    }
  }

  function handleInviteComplete() {
    showInviteModal = false;
    loadData();
  }

  const roleOptions = $derived(
    roles.map(r => ({ value: r.id, label: r.name }))
  );
</script>

<div class="space-y-4">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium text-gray-900">Members</h2>
    <button
      type="button"
      onclick={() => showInviteModal = true}
      class="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
    >
      <svg class="w-4 h-4 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
      </svg>
      Invite Member
    </button>
  </div>

  {#if error}
    <div class="rounded-md bg-red-50 p-4">
      <p class="text-sm text-red-700">{error}</p>
    </div>
  {/if}

  {#if loading}
    <div class="flex items-center justify-center py-12">
      <div class="text-gray-500">Loading members...</div>
    </div>
  {:else if members.length === 0}
    <div class="text-center py-12 bg-white rounded-lg border border-gray-200">
      <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>
      <h3 class="mt-2 text-sm font-medium text-gray-900">No members</h3>
      <p class="mt-1 text-sm text-gray-500">Get started by inviting team members.</p>
    </div>
  {:else}
    <div class="bg-white shadow-sm rounded-lg border border-gray-200 divide-y divide-gray-200">
      {#each members as member (member.id)}
        <MemberRow
          {member}
          isCurrentUser={currentUser?.id === member.user.id}
          onChangeRole={() => handleChangeRole(member)}
          onRemove={() => handleRemove(member)}
        />
      {/each}
    </div>
  {/if}
</div>

<InviteMemberModal
  {organizationId}
  {roles}
  open={showInviteModal}
  onOpenChange={(open) => showInviteModal = open}
  onInvited={handleInviteComplete}
/>

<Modal
  open={showChangeRoleModal}
  onOpenChange={(open) => showChangeRoleModal = open}
  title="Change Role"
  size="sm"
>
  <div class="p-6 space-y-4">
    {#if selectedMember}
      <p class="text-sm text-gray-600">
        Change the role for <span class="font-medium">{selectedMember.user.displayName || selectedMember.user.email || 'this member'}</span>
      </p>
      <BitsSelect
        options={roleOptions}
        bind:value={newRoleId}
        label="Role"
      />
    {/if}
  </div>
  {#snippet footer()}
    <div class="flex justify-end gap-3">
      <button
        type="button"
        onclick={() => showChangeRoleModal = false}
        class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
      >
        Cancel
      </button>
      <button
        type="button"
        onclick={confirmChangeRole}
        disabled={processing || !newRoleId}
        class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {processing ? 'Saving...' : 'Change Role'}
      </button>
    </div>
  {/snippet}
</Modal>

<ConfirmModal
  isOpen={showRemoveModal}
  title="Remove Member"
  message={`Are you sure you want to remove ${selectedMember?.user.displayName || selectedMember?.user.email || 'this member'} from this organization? They will lose access to all projects and boards.`}
  confirmText={processing ? 'Removing...' : 'Remove Member'}
  cancelText="Cancel"
  variant="danger"
  onConfirm={confirmRemove}
  onCancel={() => showRemoveModal = false}
/>
