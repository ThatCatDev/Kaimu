<script lang="ts">
  import { onMount } from 'svelte';
  import { logout } from '../lib/stores/auth.svelte';
  import { getMe } from '../lib/api/auth';
  import type { User } from '../lib/graphql/generated';

  let user = $state<User | null>(null);
  let isLoading = $state(true);
  let loggingOut = $state(false);

  onMount(async () => {
    try {
      user = await getMe();
    } catch {
      user = null;
    } finally {
      isLoading = false;
    }
  });

  async function handleLogout() {
    loggingOut = true;
    try {
      await logout();
      window.location.href = '/';
    } catch (e) {
      console.error('Logout failed:', e);
    } finally {
      loggingOut = false;
    }
  }
</script>

<nav class="bg-white shadow">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="flex justify-between h-16">
      <div class="flex items-center">
        <a href="/" class="text-xl font-bold text-indigo-600">Kaimu</a>
      </div>

      <div class="flex items-center gap-4">
        {#if isLoading}
          <div class="flex items-center gap-3">
            <div class="h-4 w-24 bg-gray-200 rounded animate-pulse"></div>
            <div class="h-8 w-20 bg-gray-200 rounded animate-pulse"></div>
          </div>
        {:else if user}
          <a
            href="/dashboard"
            class="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-900"
          >
            Dashboard
          </a>
          <span class="text-gray-600">Hello, {user.username}</span>
          <button
            onclick={handleLogout}
            disabled={loggingOut}
            class="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-900 disabled:opacity-50"
          >
            {loggingOut ? 'Logging out...' : 'Logout'}
          </button>
        {:else}
          <a
            href="/login"
            class="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-900"
          >
            Login
          </a>
          <a
            href="/register"
            class="px-4 py-2 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700"
          >
            Register
          </a>
        {/if}
      </div>
    </div>
  </div>
</nav>
