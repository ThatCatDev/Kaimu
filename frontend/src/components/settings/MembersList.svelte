<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getOrganizationMembers,
    getRoles,
    changeMemberRole,
    removeMember,
    getInvitations,
    cancelInvitation,
    resendInvitation,
    type OrganizationMember,
    type Role,
    type Invitation
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
  let invitations = $state<Invitation[]>([]);
  let roles = $state<Role[]>([]);
  let currentUser = $state<User | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  // Modal state
  let showInviteModal = $state(false);
  let showChangeRoleModal = $state(false);
  let showRemoveModal = $state(false);
  let showCancelInviteModal = $state(false);
  let selectedMember = $state<OrganizationMember | null>(null);
  let selectedInvitation = $state<Invitation | null>(null);
  let newRoleId = $state('');
  let processing = $state(false);

  // Resend button state tracking per invitation
  let resendingId = $state<string | null>(null);
  let resentId = $state<string | null>(null);

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    try {
      loading = true;
      error = null;
      const [me, membersData, rolesData, invitationsData] = await Promise.all([
        getMe(),
        getOrganizationMembers(organizationId),
        getRoles(organizationId),
        getInvitations(organizationId)
      ]);
      currentUser = me;
      members = membersData;
      roles = rolesData;
      invitations = invitationsData;
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

  function handleCancelInvite(invite: Invitation) {
    selectedInvitation = invite;
    showCancelInviteModal = true;
  }

  async function confirmCancelInvite() {
    if (!selectedInvitation) return;
    try {
      processing = true;
      await cancelInvitation(selectedInvitation.id);
      await loadData();
      showCancelInviteModal = false;
      selectedInvitation = null;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to cancel invitation';
    } finally {
      processing = false;
    }
  }

  async function handleResendInvite(invite: Invitation) {
    try {
      resendingId = invite.id;
      error = null;
      await resendInvitation(invite.id);
      // Show success state
      resendingId = null;
      resentId = invite.id;
      // Reset to original after 2 seconds
      setTimeout(() => {
        resentId = null;
      }, 2000);
      // Don't reload data - resend only updates token/expiration which aren't displayed
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to resend invitation';
      resendingId = null;
    }
  }

  function formatDate(dateString: string): string {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }

  function isExpired(expiresAt: string): boolean {
    return new Date(expiresAt) < new Date();
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
    <div class="bg-white shadow-sm rounded-lg border border-gray-200 divide-y divide-gray-200">
      {#each [1, 2, 3] as _}
        <div class="px-4 py-4 flex items-center gap-4">
          <div class="h-10 w-10 bg-gray-200 rounded-full animate-pulse"></div>
          <div class="flex-1">
            <div class="h-4 w-32 bg-gray-200 rounded animate-pulse mb-2"></div>
            <div class="h-3 w-48 bg-gray-200 rounded animate-pulse"></div>
          </div>
          <div class="h-6 w-20 bg-gray-200 rounded-full animate-pulse"></div>
          <div class="h-8 w-8 bg-gray-200 rounded animate-pulse"></div>
        </div>
      {/each}
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

  <!-- Pending Invitations -->
  {#if !loading && invitations.length > 0}
    <div class="mt-8">
      <h3 class="text-sm font-medium text-gray-700 mb-3">Pending Invitations ({invitations.length})</h3>
      <div class="bg-white shadow-sm rounded-lg border border-gray-200 divide-y divide-gray-200">
        {#each invitations as invite (invite.id)}
          <div class="p-4 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center">
                <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-gray-900">{invite.email}</span>
                  {#if isExpired(invite.expiresAt)}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800">
                      Expired
                    </span>
                  {:else}
                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                      Pending
                    </span>
                  {/if}
                </div>
                <div class="text-xs text-gray-500 mt-0.5">
                  {invite.role?.name ?? 'Member'} &middot; Invited {formatDate(invite.createdAt)} &middot; {isExpired(invite.expiresAt) ? 'Expired' : `Expires ${formatDate(invite.expiresAt)}`}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                onclick={() => handleResendInvite(invite)}
                disabled={resendingId === invite.id || resentId === invite.id}
                class="inline-flex items-center px-2.5 py-1.5 text-xs font-medium rounded-md transition-all duration-200 {resentId === invite.id ? 'text-green-700 bg-green-50 border border-green-300' : 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50'} disabled:cursor-not-allowed"
              >
                {#if resendingId === invite.id}
                  <svg class="w-3.5 h-3.5 mr-1 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Sending...
                {:else if resentId === invite.id}
                  <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                  Sent!
                {:else}
                  <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  Resend
                {/if}
              </button>
              <button
                type="button"
                onclick={() => handleCancelInvite(invite)}
                disabled={processing}
                class="inline-flex items-center px-2.5 py-1.5 text-xs font-medium text-red-700 bg-white border border-red-300 rounded-md hover:bg-red-50 disabled:opacity-50"
              >
                <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
                Cancel
              </button>
            </div>
          </div>
        {/each}
      </div>
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

<ConfirmModal
  isOpen={showCancelInviteModal}
  title="Cancel Invitation"
  message={`Are you sure you want to cancel the invitation for ${selectedInvitation?.email || 'this email'}? They will no longer be able to join using the invitation link.`}
  confirmText={processing ? 'Canceling...' : 'Cancel Invitation'}
  cancelText="Keep Invitation"
  variant="danger"
  onConfirm={confirmCancelInvite}
  onCancel={() => showCancelInviteModal = false}
/>
