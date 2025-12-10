<script lang="ts">
  import { inviteMember, type Role } from '../../lib/api/rbac';
  import { Modal, Input, BitsSelect, Button } from '../ui';

  interface Props {
    organizationId: string;
    roles: Role[];
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onInvited: () => void;
  }

  let { organizationId, roles, open, onOpenChange, onInvited }: Props = $props();

  let email = $state('');
  let roleId = $state('');
  let error = $state<string | null>(null);
  let sending = $state(false);
  let invitationLink = $state<string | null>(null);
  let copied = $state(false);

  // Reset form when modal opens
  $effect(() => {
    if (open) {
      email = '';
      error = null;
      invitationLink = null;
      copied = false;
      // Default to first non-owner role, or fall back to first role
      const defaultRole = roles.find(r => r.name !== 'Owner') || roles[0];
      roleId = defaultRole?.id || '';
    }
  });

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!email || !roleId) return;

    try {
      sending = true;
      error = null;
      const invitation = await inviteMember(organizationId, email, roleId);
      // Build the invitation link
      const baseUrl = window.location.origin;
      invitationLink = `${baseUrl}/invite/${invitation.token}`;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to send invitation';
    } finally {
      sending = false;
    }
  }

  async function copyLink() {
    if (!invitationLink) return;
    try {
      await navigator.clipboard.writeText(invitationLink);
      copied = true;
      setTimeout(() => copied = false, 2000);
    } catch {
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = invitationLink;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      copied = true;
      setTimeout(() => copied = false, 2000);
    }
  }

  function handleDone() {
    onInvited();
    onOpenChange(false);
  }

  const roleOptions = $derived(
    roles
      .filter(r => r.name !== 'Owner') // Can't invite as owner
      .map(r => ({ value: r.id, label: r.name }))
  );
</script>

<Modal
  {open}
  {onOpenChange}
  title={invitationLink ? "Invitation Created" : "Invite Member"}
  description={invitationLink ? "Share this link with the invitee" : "Send an invitation to join your organization"}
  size="sm"
>
  {#if invitationLink}
    <!-- Success state with link -->
    <div class="p-6 space-y-4">
      <div class="flex items-center justify-center">
        <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100">
          <svg class="h-6 w-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </div>
      </div>

      <p class="text-sm text-gray-600 text-center">
        Invitation sent to <span class="font-medium">{email}</span>
      </p>

      <div class="mt-4">
        <label class="block text-sm font-medium text-gray-700 mb-1">Invitation Link</label>
        <div class="flex gap-2">
          <input
            type="text"
            value={invitationLink}
            readonly
            class="flex-1 px-3 py-2 text-sm border border-gray-300 rounded-md bg-gray-50 text-gray-600"
          />
          <button
            type="button"
            onclick={copyLink}
            class="px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 flex items-center gap-1"
          >
            {#if copied}
              <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              Copied!
            {:else}
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
              </svg>
              Copy
            {/if}
          </button>
        </div>
        <p class="mt-2 text-xs text-gray-500">
          Share this link with the invitee. The link expires in 7 days.
        </p>
      </div>
    </div>
  {:else}
    <!-- Form state -->
    <form onsubmit={handleSubmit} class="p-6 space-y-4">
      {#if error}
        <div class="rounded-md bg-red-50 p-3">
          <p class="text-sm text-red-700">{error}</p>
        </div>
      {/if}

      <Input
        label="Email Address"
        type="email"
        bind:value={email}
        placeholder="colleague@example.com"
        required
      />

      <BitsSelect
        options={roleOptions}
        bind:value={roleId}
        label="Role"
        placeholder="Select a role"
        required
      />

      <p class="text-xs text-gray-500">
        An invitation link will be generated that you can share with the invitee.
      </p>
    </form>
  {/if}

  {#snippet footer()}
    <div class="flex justify-end gap-3">
      {#if invitationLink}
        <Button variant="primary" onclick={handleDone}>
          Done
        </Button>
      {:else}
        <button
          type="button"
          onclick={() => onOpenChange(false)}
          class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
        >
          Cancel
        </button>
        <button
          type="button"
          onclick={handleSubmit}
          disabled={sending || !email || !roleId}
          class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {sending ? 'Creating...' : 'Create Invitation'}
        </button>
      {/if}
    </div>
  {/snippet}
</Modal>
