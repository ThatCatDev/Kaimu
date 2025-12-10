<script lang="ts">
  import { onMount } from 'svelte';
  import { acceptInvitation } from '../lib/api/rbac';
  import { getMe } from '../lib/api/auth';
  import type { User } from '../lib/graphql/generated';
  import Button from './ui/Button.svelte';

  interface Props {
    token: string;
  }

  let { token }: Props = $props();

  let user = $state<User | null>(null);
  let loading = $state(true);
  let accepting = $state(false);
  let error = $state<string | null>(null);
  let success = $state<{ orgName: string; orgId: string } | null>(null);

  onMount(async () => {
    try {
      user = await getMe();
    } catch {
      user = null;
    } finally {
      loading = false;
    }
  });

  async function handleAccept() {
    if (!token) {
      error = 'Invalid invitation link';
      return;
    }

    accepting = true;
    error = null;

    try {
      const org = await acceptInvitation(token);
      success = { orgName: org.name, orgId: org.id };
    } catch (e) {
      if (e instanceof Error) {
        // Map common errors to user-friendly messages
        if (e.message.includes('not found')) {
          error = 'This invitation link is invalid or has already been used.';
        } else if (e.message.includes('expired')) {
          error = 'This invitation has expired. Please ask for a new invitation.';
        } else if (e.message.includes('already been accepted')) {
          error = 'This invitation has already been accepted.';
        } else if (e.message.includes('already a member')) {
          error = 'You are already a member of this organization.';
        } else if (e.message.includes('email does not match')) {
          error = 'Your account email does not match the invitation email. Please use the correct account.';
        } else {
          error = e.message;
        }
      } else {
        error = 'Failed to accept invitation. Please try again.';
      }
    } finally {
      accepting = false;
    }
  }

  function goToLogin() {
    // Store the current URL to redirect back after login
    const currentUrl = window.location.href;
    sessionStorage.setItem('redirectAfterLogin', currentUrl);
    window.location.href = '/login';
  }

  function goToOrganization() {
    if (success) {
      window.location.href = `/organizations/${success.orgId}`;
    }
  }

  function goToDashboard() {
    window.location.href = '/dashboard';
  }
</script>

<div class="bg-white py-8 px-4 shadow sm:rounded-lg sm:px-10">
  {#if loading}
    <div class="flex justify-center items-center py-8">
      <svg class="animate-spin h-8 w-8 text-indigo-600" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </div>
  {:else if success}
    <!-- Success state -->
    <div class="text-center">
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100">
        <svg class="h-6 w-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
      </div>
      <h3 class="mt-4 text-lg font-medium text-gray-900">Welcome to {success.orgName}!</h3>
      <p class="mt-2 text-sm text-gray-500">
        You have successfully joined the organization.
      </p>
      <div class="mt-6">
        <Button onclick={goToOrganization} variant="primary">
          Go to Organization
        </Button>
      </div>
    </div>
  {:else if !user}
    <!-- Not logged in -->
    <div class="text-center">
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-yellow-100">
        <svg class="h-6 w-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h3 class="mt-4 text-lg font-medium text-gray-900">Login Required</h3>
      <p class="mt-2 text-sm text-gray-500">
        Please log in or create an account to accept this invitation.
      </p>
      <div class="mt-6 space-y-3">
        <Button onclick={goToLogin} variant="primary" class="w-full">
          Log In
        </Button>
        <a
          href="/register"
          class="block w-full text-center px-4 py-2 text-sm font-medium text-indigo-600 bg-white border border-indigo-600 rounded-md hover:bg-indigo-50"
        >
          Create Account
        </a>
      </div>
    </div>
  {:else if error}
    <!-- Error state -->
    <div class="text-center">
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100">
        <svg class="h-6 w-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </div>
      <h3 class="mt-4 text-lg font-medium text-gray-900">Unable to Accept Invitation</h3>
      <p class="mt-2 text-sm text-red-600">{error}</p>
      <div class="mt-6">
        <Button onclick={goToDashboard} variant="secondary">
          Go to Dashboard
        </Button>
      </div>
    </div>
  {:else}
    <!-- Ready to accept -->
    <div class="text-center">
      <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-indigo-100">
        <svg class="h-6 w-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
        </svg>
      </div>
      <h3 class="mt-4 text-lg font-medium text-gray-900">You've Been Invited!</h3>
      <p class="mt-2 text-sm text-gray-500">
        You have been invited to join an organization. Click below to accept the invitation.
      </p>
      <div class="mt-2 text-sm text-gray-700">
        Logged in as <span class="font-medium">{user.displayName || user.username}</span>
      </div>
      <div class="mt-6 space-y-3">
        <Button onclick={handleAccept} variant="primary" loading={accepting} class="w-full">
          {accepting ? 'Accepting...' : 'Accept Invitation'}
        </Button>
        <Button onclick={goToDashboard} variant="secondary" class="w-full">
          Cancel
        </Button>
      </div>
    </div>
  {/if}
</div>
