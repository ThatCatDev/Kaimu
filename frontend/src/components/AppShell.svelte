<script lang="ts">
  import { onMount } from 'svelte';
  import Sidebar from './Sidebar.svelte';
  import { logout } from '../lib/stores/auth.svelte';
  import { getMe } from '../lib/api/auth';
  import type { User } from '../lib/graphql/generated';
  import type { Snippet } from 'svelte';

  interface Props {
    currentPath?: string;
    children: Snippet;
  }

  let { currentPath = '', children }: Props = $props();

  let user = $state<User | null>(null);
  let isLoading = $state(true);
  let loggingOut = $state(false);
  let sidebarCollapsed = $state(false);

  onMount(() => {
    // Load sidebar state from localStorage
    const savedState = localStorage.getItem('sidebarCollapsed');
    if (savedState !== null) {
      sidebarCollapsed = savedState === 'true';
    }

    // Load user
    loadUser();
  });

  async function loadUser() {
    try {
      user = await getMe();
    } catch {
      user = null;
    } finally {
      isLoading = false;
    }
  }

  function toggleSidebar() {
    sidebarCollapsed = !sidebarCollapsed;
    localStorage.setItem('sidebarCollapsed', String(sidebarCollapsed));
  }

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

<div class="h-screen flex overflow-hidden bg-gray-50">
  <!-- Sidebar -->
  <Sidebar collapsed={sidebarCollapsed} onToggle={toggleSidebar} {currentPath} />

  <!-- Main content area -->
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Top header bar -->
    <header class="h-16 bg-white shadow-sm flex items-center justify-between px-6 flex-shrink-0">
      <div class="flex items-center gap-4">
        <!-- Mobile menu button (shows on small screens) -->
        <button
          type="button"
          onclick={toggleSidebar}
          class="lg:hidden p-2 rounded-md text-gray-400 hover:text-gray-600 hover:bg-gray-100"
        >
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>

      <!-- Right side: user menu -->
      <div class="flex items-center gap-4">
        {#if isLoading}
          <span class="text-sm text-gray-400">Loading...</span>
        {:else if user}
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-2">
              <div class="w-8 h-8 rounded-full bg-indigo-100 flex items-center justify-center">
                <span class="text-sm font-medium text-indigo-600">
                  {user.username.charAt(0).toUpperCase()}
                </span>
              </div>
              <span class="text-sm font-medium text-gray-700 hidden sm:block">{user.username}</span>
            </div>
            <button
              type="button"
              onclick={handleLogout}
              disabled={loggingOut}
              class="px-3 py-1.5 text-sm font-medium text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-md transition-colors disabled:opacity-50"
            >
              {loggingOut ? 'Logging out...' : 'Logout'}
            </button>
          </div>
        {:else}
          <a
            href="/login"
            class="px-3 py-1.5 text-sm font-medium text-gray-600 hover:text-gray-900"
          >
            Login
          </a>
          <a
            href="/register"
            class="px-3 py-1.5 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700"
          >
            Register
          </a>
        {/if}
      </div>
    </header>

    <!-- Main content -->
    <main class="flex-1 overflow-auto">
      {@render children()}
    </main>
  </div>
</div>
