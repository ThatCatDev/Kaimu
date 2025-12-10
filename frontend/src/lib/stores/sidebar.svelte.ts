import { getOrganizations } from '../api/organizations';
import type { OrganizationsQuery } from '../graphql/generated';

export type SidebarOrganization = OrganizationsQuery['organizations'][number];

// Global sidebar state using Svelte 5 runes
let organizations = $state<SidebarOrganization[]>([]);
let loading = $state(false);
let initialized = $state(false);

// Refresh version - incremented to trigger re-fetch
let refreshVersion = $state(0);

/**
 * Load organizations for the sidebar
 */
async function loadOrganizations(): Promise<void> {
  if (loading) return;

  loading = true;
  try {
    const data = await getOrganizations();
    organizations = data;
    initialized = true;
    // Update cache
    try {
      sessionStorage.setItem('sidebarOrganizations', JSON.stringify(data));
    } catch {
      // Ignore storage errors
    }
  } catch (error) {
    console.error('Failed to load sidebar organizations:', error);
  } finally {
    loading = false;
  }
}

/**
 * Initialize sidebar with cached data, then fetch fresh data
 */
function initializeFromCache(): void {
  try {
    const cached = sessionStorage.getItem('sidebarOrganizations');
    if (cached) {
      organizations = JSON.parse(cached);
      initialized = true;
    }
  } catch {
    // Ignore parse errors
  }
}

/**
 * Trigger a refresh of the sidebar data
 */
function refresh(): void {
  refreshVersion++;
  loadOrganizations();
}

/**
 * Clear sidebar cache (useful for logout)
 */
function clearCache(): void {
  try {
    sessionStorage.removeItem('sidebarOrganizations');
  } catch {
    // Ignore storage errors
  }
  organizations = [];
  initialized = false;
}

// Export the store
export const sidebarStore = {
  get organizations() { return organizations; },
  get loading() { return loading; },
  get initialized() { return initialized; },
  get refreshVersion() { return refreshVersion; },
  loadOrganizations,
  initializeFromCache,
  refresh,
  clearCache,
};
