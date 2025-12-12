<script lang="ts">
  import { onMount } from 'svelte';
  import { sidebarStore, type SidebarOrganization } from '../lib/stores/sidebar.svelte';

  type ProjectItem = SidebarOrganization['projects'][number];

  interface Props {
    collapsed?: boolean;
    onToggle?: () => void;
    currentPath?: string;
  }

  let { collapsed = false, onToggle, currentPath = '' }: Props = $props();

  // Use store data
  let organizations = $derived(sidebarStore.organizations);
  let loading = $derived(sidebarStore.loading && !sidebarStore.initialized);

  let expandedOrgs = $state<Set<string>>(new Set());
  let expandedProjects = $state<Set<string>>(new Set());

  // Persist expanded state to sessionStorage
  function saveExpandedState() {
    sessionStorage.setItem('expandedOrgs', JSON.stringify([...expandedOrgs]));
    sessionStorage.setItem('expandedProjects', JSON.stringify([...expandedProjects]));
  }

  function loadExpandedState() {
    try {
      const savedOrgs = sessionStorage.getItem('expandedOrgs');
      const savedProjects = sessionStorage.getItem('expandedProjects');
      if (savedOrgs) expandedOrgs = new Set(JSON.parse(savedOrgs));
      if (savedProjects) expandedProjects = new Set(JSON.parse(savedProjects));
    } catch {
      // Ignore parse errors
    }
  }

  // Auto-expand based on current path
  function autoExpandForPath() {
    // Auto-expand org if we're viewing one of its pages
    const orgMatch = currentPath.match(/\/organizations\/([^/]+)/);
    if (orgMatch && orgMatch[1] !== 'new') {
      const orgId = orgMatch[1];
      if (!expandedOrgs.has(orgId)) {
        expandedOrgs.add(orgId);
        expandedOrgs = new Set(expandedOrgs);
        saveExpandedState();
      }
    }

    // Auto-expand org and project if we're viewing a project or board
    const projectMatch = currentPath.match(/\/projects\/([^/]+)/);
    if (projectMatch) {
      const projectId = projectMatch[1];
      for (const org of organizations) {
        const project = org.projects?.find(p => p.id === projectId);
        if (project) {
          if (!expandedOrgs.has(org.id)) {
            expandedOrgs.add(org.id);
            expandedOrgs = new Set(expandedOrgs);
          }
          // Always expand the project to show boards
          if (project.boards && project.boards.length > 0 && !expandedProjects.has(project.id)) {
            expandedProjects.add(project.id);
            expandedProjects = new Set(expandedProjects);
          }
          saveExpandedState();
          break;
        }
      }
    }
  }

  // React to organizations changes for auto-expand
  $effect(() => {
    if (organizations.length > 0) {
      autoExpandForPath();
    }
  });

  onMount(async () => {
    // Load persisted expanded state first
    loadExpandedState();

    // Initialize from cache for instant render
    sidebarStore.initializeFromCache();

    // Then fetch fresh data
    await sidebarStore.loadOrganizations();
  });

  function toggleOrg(orgId: string) {
    if (expandedOrgs.has(orgId)) {
      expandedOrgs.delete(orgId);
    } else {
      expandedOrgs.add(orgId);
    }
    expandedOrgs = new Set(expandedOrgs);
    saveExpandedState();
  }

  function toggleProject(projectId: string) {
    if (expandedProjects.has(projectId)) {
      expandedProjects.delete(projectId);
    } else {
      expandedProjects.add(projectId);
    }
    expandedProjects = new Set(expandedProjects);
    saveExpandedState();
  }

  function isActive(path: string): boolean {
    return currentPath === path || currentPath.startsWith(path + '/');
  }

</script>

<aside
  class="h-full bg-gray-900 text-gray-100 flex flex-col transition-all duration-300 {collapsed ? 'w-16' : 'w-64'}"
>
  <!-- Header -->
  <div class="h-16 flex items-center justify-between px-4 border-b border-gray-800">
    {#if !collapsed}
      <a href="/dashboard" class="text-xl font-bold text-indigo-400">Kaimu</a>
    {/if}
    <button
      type="button"
      onclick={onToggle}
      class="p-2 rounded-md hover:bg-gray-800 text-gray-400 hover:text-gray-200 transition-colors {collapsed ? 'mx-auto' : ''}"
      title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        {#if collapsed}
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
        {:else}
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
        {/if}
      </svg>
    </button>
  </div>

  <!-- Navigation -->
  <nav class="flex-1 overflow-y-auto py-4">
    <!-- Dashboard link -->
    <a
      href="/dashboard"
      class="flex items-center gap-3 px-4 py-2.5 text-sm font-medium transition-colors {isActive('/dashboard') ? 'bg-gray-800 text-white' : 'text-gray-300 hover:bg-gray-800 hover:text-white'}"
      title={collapsed ? 'Dashboard' : undefined}
    >
      <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
      </svg>
      {#if !collapsed}
        <span>Dashboard</span>
      {/if}
    </a>

    <!-- Organizations section -->
    {#if !collapsed}
      <div class="mt-6 px-4">
        <div class="flex items-center justify-between">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Organizations</h3>
          <a
            href="/organizations/new"
            class="p-1 rounded hover:bg-gray-800 text-gray-500 hover:text-gray-300 transition-colors"
            title="New organization"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
          </a>
        </div>
      </div>
    {:else}
      <a
        href="/organizations/new"
        class="flex items-center justify-center py-2.5 mt-6 text-gray-500 hover:text-gray-300 hover:bg-gray-800 transition-colors"
        title="New organization"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
      </a>
    {/if}

    {#if loading}
      <div class="px-4 py-2 space-y-2">
        {#if !collapsed}
          {#each [1, 2, 3] as _}
            <div class="flex items-center gap-2">
              <div class="w-6 h-6 bg-gray-700 rounded animate-pulse"></div>
              <div class="flex-1 h-4 bg-gray-700 rounded animate-pulse"></div>
            </div>
          {/each}
        {:else}
          {#each [1, 2, 3] as _}
            <div class="flex justify-center py-1">
              <div class="w-6 h-6 bg-gray-700 rounded animate-pulse"></div>
            </div>
          {/each}
        {/if}
      </div>
    {:else}
      <div class="mt-2 space-y-1">
        {#each organizations as org (org.id)}
          {#if collapsed}
            <!-- Collapsed: just show icon linking to org -->
            <a
              href={`/organizations/${org.id}`}
              class="flex items-center justify-center py-2.5 transition-colors {isActive(`/organizations/${org.id}`) ? 'bg-gray-800 text-white' : 'text-gray-400 hover:bg-gray-800 hover:text-white'}"
              title={org.name}
            >
              <span class="w-6 h-6 rounded bg-indigo-600 flex items-center justify-center text-xs font-medium text-white">
                {org.name.charAt(0).toUpperCase()}
              </span>
            </a>
          {:else}
            <!-- Expanded: show org with expandable projects -->
            <div>
              <div class="flex items-center gap-1 px-2 py-1 {isActive(`/organizations/${org.id}`) ? 'bg-gray-800' : ''}">
                <!-- Expand/collapse button -->
                <button
                  type="button"
                  onclick={() => toggleOrg(org.id)}
                  class="p-1.5 rounded hover:bg-gray-700 text-gray-400 hover:text-white transition-colors"
                  title={expandedOrgs.has(org.id) ? 'Collapse' : 'Expand'}
                >
                  <svg
                    class="w-3 h-3 transition-transform {expandedOrgs.has(org.id) ? 'rotate-90' : ''}"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                  </svg>
                </button>
                <!-- Org link -->
                <a
                  href={`/organizations/${org.id}`}
                  onclick={() => { if (!expandedOrgs.has(org.id)) toggleOrg(org.id); }}
                  class="flex-1 flex items-center gap-2 px-2 py-1 rounded text-sm transition-colors {isActive(`/organizations/${org.id}`) ? 'text-white' : 'text-gray-300 hover:bg-gray-700 hover:text-white'}"
                >
                  <span class="w-6 h-6 rounded bg-indigo-600 flex items-center justify-center text-xs font-medium text-white flex-shrink-0">
                    {org.name.charAt(0).toUpperCase()}
                  </span>
                  <span class="truncate">{org.name}</span>
                </a>
                {#if org.projects && org.projects.length > 0}
                  <span class="text-xs text-gray-500 bg-gray-800 px-1.5 py-0.5 rounded">
                    {org.projects.length}
                  </span>
                {/if}
              </div>

              {#if expandedOrgs.has(org.id)}
                <div class="ml-6 border-l border-gray-700 pl-2 py-1 space-y-0.5">
                  {#if org.projects && org.projects.length > 0}
                    {#each org.projects as project (project.id)}
                      <div>
                        <div class="flex items-center gap-1 {isActive(`/projects/${project.id}`) ? 'bg-gray-800 rounded' : ''}">
                          <!-- Expand/collapse button for project -->
                          {#if project.boards && project.boards.length > 0}
                            <button
                              type="button"
                              onclick={() => toggleProject(project.id)}
                              class="p-1 rounded hover:bg-gray-700 text-gray-500 hover:text-white transition-colors"
                              title={expandedProjects.has(project.id) ? 'Collapse' : 'Expand'}
                            >
                              <svg
                                class="w-3 h-3 transition-transform {expandedProjects.has(project.id) ? 'rotate-90' : ''}"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                              >
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                              </svg>
                            </button>
                          {:else}
                            <span class="w-5"></span>
                          {/if}
                          <!-- Project link -->
                          <a
                            href={`/projects/${project.id}`}
                            onclick={() => { if (project.boards && project.boards.length > 0 && !expandedProjects.has(project.id)) toggleProject(project.id); }}
                            class="flex-1 flex items-center gap-2 px-2 py-1 rounded text-sm transition-colors {isActive(`/projects/${project.id}`) ? 'text-white' : 'text-gray-400 hover:bg-gray-700 hover:text-white'}"
                          >
                            <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                            </svg>
                            <span class="truncate">{project.name}</span>
                            <span class="text-xs text-gray-600">{project.key}</span>
                          </a>
                        </div>

                        {#if expandedProjects.has(project.id) && project.boards && project.boards.length > 0}
                          <div class="ml-5 border-l border-gray-700 pl-3 py-1 space-y-0.5">
                            {#each project.boards as board (board.id)}
                              <a
                                href={`/projects/${project.id}/board/${board.id}`}
                                class="flex items-center gap-2 px-2 py-1 text-xs rounded transition-colors {currentPath === `/projects/${project.id}/board/${board.id}` ? 'bg-gray-800 text-white' : 'text-gray-500 hover:bg-gray-700 hover:text-gray-300'}"
                              >
                                <svg class="w-3.5 h-3.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17V7m0 10a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h2a2 2 0 012 2m0 10a2 2 0 002 2h2a2 2 0 002-2M9 7a2 2 0 012-2h2a2 2 0 012 2m0 10V7m0 10a2 2 0 002 2h2a2 2 0 002-2V7a2 2 0 00-2-2h-2a2 2 0 00-2 2" />
                                </svg>
                                <span class="truncate">{board.name}</span>
                              </a>
                            {/each}
                          </div>
                        {/if}
                      </div>
                    {/each}
                  {:else}
                    <span class="block px-2 py-1.5 text-sm text-gray-500">No projects</span>
                  {/if}
                  <a
                    href={`/organizations/${org.id}/projects/new`}
                    class="flex items-center gap-2 px-2 py-1.5 text-sm rounded text-gray-500 hover:bg-gray-700 hover:text-gray-300 transition-colors"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                    <span>New Project</span>
                  </a>
                </div>
              {/if}
            </div>
          {/if}
        {/each}
      </div>
    {/if}
  </nav>

  <!-- Footer with user info (optional) -->
  <div class="border-t border-gray-800 p-4">
    {#if collapsed}
      <button
        type="button"
        class="w-full flex items-center justify-center p-2 rounded-md text-gray-400 hover:bg-gray-800 hover:text-white transition-colors"
        title="Settings"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
      </button>
    {:else}
      <a
        href="/settings"
        class="flex items-center gap-3 px-2 py-2 rounded-md text-gray-400 hover:bg-gray-800 hover:text-white transition-colors"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        <span>Settings</span>
      </a>
    {/if}
  </div>
</aside>
