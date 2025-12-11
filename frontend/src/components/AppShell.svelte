<script lang="ts">
  import { onMount } from 'svelte';
  import { Toaster } from 'svelte-sonner';
  import Sidebar from './Sidebar.svelte';
  import UserSettingsModal from './settings/UserSettingsModal.svelte';
  import CommandPalette from './search/CommandPalette.svelte';
  import { logout } from '../lib/stores/auth.svelte';
  import { getMe } from '../lib/api/auth';
  import type { User } from '../lib/graphql/generated';
  import type { Snippet } from 'svelte';
  import type { SearchScope } from '../lib/api/search';

  interface Props {
    currentPath?: string;
    children: Snippet;
  }

  let { currentPath = '', children }: Props = $props();

  let user = $state<User | null>(null);
  let isLoading = $state(true);
  let loggingOut = $state(false);
  let sidebarCollapsed = $state(false);
  let userMenuOpen = $state(false);
  let settingsModalOpen = $state(false);
  let commandPaletteOpen = $state(false);

  // Derive search scope from current path
  const searchScope = $derived<SearchScope | undefined>(() => {
    // Parse currentPath to extract org/project context
    // Format: /org-slug/project-key/...
    const parts = currentPath.split('/').filter(Boolean);
    if (parts.length >= 2) {
      // We have at least org/project - scope search to this context
      // Note: We'd need the actual IDs from the page, for now we don't scope
      return undefined;
    }
    return undefined;
  });

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

  function handleOpenSettings() {
    userMenuOpen = false;
    settingsModalOpen = true;
  }

  function handleUserUpdate(updatedUser: User) {
    user = updatedUser;
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('[data-user-menu]')) {
      userMenuOpen = false;
    }
  }

  function handleGlobalKeydown(event: KeyboardEvent) {
    // ⌘K or Ctrl+K to open command palette
    if ((event.metaKey || event.ctrlKey) && event.key === 'k') {
      event.preventDefault();
      commandPaletteOpen = true;
    }
  }
</script>

<svelte:window onclick={handleClickOutside} onkeydown={handleGlobalKeydown} />

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

      <!-- Search button -->
      <button
        type="button"
        onclick={() => commandPaletteOpen = true}
        class="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-500 bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <span class="hidden sm:inline">Search</span>
        <kbd class="hidden md:inline-flex items-center px-1.5 py-0.5 text-xs text-gray-400 bg-white rounded border border-gray-200">
          ⌘K
        </kbd>
      </button>

      <!-- Right side: user menu -->
      <div class="flex items-center gap-4">
        {#if isLoading}
          <span class="text-sm text-gray-400">Loading...</span>
        {:else if user}
          <div class="relative" data-user-menu>
            <button
              type="button"
              onclick={() => userMenuOpen = !userMenuOpen}
              class="flex items-center gap-2 p-1.5 rounded-md hover:bg-gray-100 transition-colors"
            >
              <div class="w-8 h-8 rounded-full bg-indigo-100 flex items-center justify-center">
                <span class="text-sm font-medium text-indigo-600">
                  {(user.displayName || user.username).charAt(0).toUpperCase()}
                </span>
              </div>
              <span class="text-sm font-medium text-gray-700 hidden sm:block">
                {user.displayName || user.username}
              </span>
              <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>

            {#if userMenuOpen}
              <div class="absolute right-0 mt-2 w-56 bg-white rounded-md shadow-lg ring-1 ring-black ring-opacity-5 z-50">
                <div class="px-4 py-3 border-b border-gray-100">
                  <p class="text-sm font-medium text-gray-900">{user.displayName || user.username}</p>
                  <p class="text-xs text-gray-500 truncate">{user.email || user.username}</p>
                </div>
                <div class="py-1">
                  <button
                    type="button"
                    onclick={handleOpenSettings}
                    class="w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2 text-left"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    Account Settings
                  </button>
                  <button
                    type="button"
                    onclick={handleLogout}
                    disabled={loggingOut}
                    class="w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2 text-left disabled:opacity-50"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                    </svg>
                    {loggingOut ? 'Logging out...' : 'Logout'}
                  </button>
                </div>
              </div>
            {/if}
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

{#if user}
  <UserSettingsModal
    bind:open={settingsModalOpen}
    {user}
    onUpdate={handleUserUpdate}
  />
{/if}

<Toaster position="bottom-right" richColors />

<CommandPalette bind:open={commandPaletteOpen} scope={searchScope()} />
