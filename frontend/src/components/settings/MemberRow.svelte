<script lang="ts">
  import type { OrganizationMember } from '../../lib/api/rbac';

  interface Props {
    member: OrganizationMember;
    isCurrentUser: boolean;
    onChangeRole: () => void;
    onRemove: () => void;
  }

  let { member, isCurrentUser, onChangeRole, onRemove }: Props = $props();

  const isOwner = $derived(member.role?.name === 'Owner');
  const displayName = $derived(member.user.displayName || member.user.email || 'Unknown');
  const initials = $derived(displayName.charAt(0).toUpperCase());
</script>

<div class="px-4 py-4 flex items-center justify-between">
  <div class="flex items-center min-w-0">
    <div class="flex-shrink-0">
      <div class="h-10 w-10 rounded-full bg-indigo-100 flex items-center justify-center">
        <span class="text-indigo-700 font-medium text-sm">
          {initials}
        </span>
      </div>
    </div>
    <div class="ml-4 min-w-0">
      <p class="text-sm font-medium text-gray-900 truncate">
        {displayName}
        {#if isCurrentUser}
          <span class="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
            You
          </span>
        {/if}
      </p>
      <p class="text-sm text-gray-500 truncate">{member.user.email || 'No email'}</p>
    </div>
  </div>

  <div class="flex items-center gap-4">
    <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {isOwner ? 'bg-purple-100 text-purple-800' : 'bg-gray-100 text-gray-800'}">
      {member.role?.name || member.legacyRole || 'Member'}
    </span>

    {#if !isOwner}
      <div class="flex items-center gap-2">
        <button
          type="button"
          onclick={onChangeRole}
          class="text-gray-400 hover:text-gray-600 p-1 rounded hover:bg-gray-100"
          title="Change role"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
        </button>
        {#if !isCurrentUser}
          <button
            type="button"
            onclick={onRemove}
            class="text-gray-400 hover:text-red-600 p-1 rounded hover:bg-red-50"
            title="Remove member"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        {/if}
      </div>
    {/if}
  </div>
</div>
