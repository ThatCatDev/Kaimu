<script lang="ts">
  import { onMount } from 'svelte';
  import { getOrganizationActivity, type AuditEvent, type AuditEventConnection } from '../../lib/api/activity';
  import { AuditAction, AuditEntityType, type AuditFilters } from '../../lib/graphql/generated';
  import ActivityItem from './ActivityItem.svelte';

  interface Props {
    organizationId: string;
  }

  let { organizationId }: Props = $props();

  let events = $state<AuditEvent[]>([]);
  let loading = $state(true);
  let loadingMore = $state(false);
  let error = $state<string | null>(null);
  let pageInfo = $state<AuditEventConnection['pageInfo'] | null>(null);
  let totalCount = $state(0);
  let endCursor = $state<string | null>(null);

  // Filters
  let selectedActions = $state<AuditAction[]>([]);
  let selectedEntityTypes = $state<AuditEntityType[]>([]);
  let showFilters = $state(false);

  const allActions: AuditAction[] = [
    AuditAction.Created, AuditAction.Updated, AuditAction.Deleted, AuditAction.CardMoved,
    AuditAction.CardAssigned, AuditAction.CardUnassigned,
    AuditAction.SprintStarted, AuditAction.SprintCompleted,
    AuditAction.CardAddedToSprint, AuditAction.CardRemovedFromSprint,
    AuditAction.MemberInvited, AuditAction.MemberJoined, AuditAction.MemberRemoved, AuditAction.MemberRoleChanged,
    AuditAction.ColumnReordered, AuditAction.ColumnVisibilityToggled
  ];

  const allEntityTypes: AuditEntityType[] = [
    AuditEntityType.Card, AuditEntityType.Board, AuditEntityType.BoardColumn,
    AuditEntityType.Project, AuditEntityType.Sprint, AuditEntityType.Tag,
    AuditEntityType.Role, AuditEntityType.Invitation, AuditEntityType.User
  ];

  async function loadActivity(reset = false) {
    if (reset) {
      events = [];
      endCursor = null;
      loading = true;
    } else {
      loadingMore = true;
    }
    error = null;

    try {
      const filters: AuditFilters | undefined = (selectedActions.length > 0 || selectedEntityTypes.length > 0)
        ? {
            actions: selectedActions.length > 0 ? selectedActions : undefined,
            entityTypes: selectedEntityTypes.length > 0 ? selectedEntityTypes : undefined,
          }
        : undefined;

      const result = await getOrganizationActivity(
        organizationId,
        20,
        reset ? undefined : (endCursor || undefined),
        filters
      );

      const newEvents = result.edges.map(edge => edge.node);
      if (reset) {
        events = newEvents;
      } else {
        events = [...events, ...newEvents];
      }

      pageInfo = result.pageInfo;
      totalCount = result.totalCount;
      endCursor = result.pageInfo.endCursor || null;
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load activity';
    } finally {
      loading = false;
      loadingMore = false;
    }
  }

  function toggleAction(action: AuditAction) {
    if (selectedActions.includes(action)) {
      selectedActions = selectedActions.filter(a => a !== action);
    } else {
      selectedActions = [...selectedActions, action];
    }
  }

  function toggleEntityType(type: AuditEntityType) {
    if (selectedEntityTypes.includes(type)) {
      selectedEntityTypes = selectedEntityTypes.filter(t => t !== type);
    } else {
      selectedEntityTypes = [...selectedEntityTypes, type];
    }
  }

  function clearFilters() {
    selectedActions = [];
    selectedEntityTypes = [];
  }

  function applyFilters() {
    loadActivity(true);
  }

  const hasActiveFilters = $derived(selectedActions.length > 0 || selectedEntityTypes.length > 0);

  onMount(() => {
    loadActivity(true);
  });
</script>

<div class="space-y-4">
  <!-- Header with filter toggle -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-2">
      <h3 class="text-lg font-medium text-gray-900">Activity</h3>
      {#if totalCount > 0}
        <span class="text-sm text-gray-500">({totalCount} events)</span>
      {/if}
    </div>
    <button
      type="button"
      onclick={() => showFilters = !showFilters}
      class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md {hasActiveFilters ? 'bg-indigo-100 text-indigo-700' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'}"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
      </svg>
      Filters
      {#if hasActiveFilters}
        <span class="bg-indigo-600 text-white text-xs rounded-full px-1.5 py-0.5">
          {selectedActions.length + selectedEntityTypes.length}
        </span>
      {/if}
    </button>
  </div>

  <!-- Filters panel -->
  {#if showFilters}
    <div class="bg-gray-50 rounded-lg p-4 space-y-4">
      <!-- Entity types -->
      <div>
        <h4 class="text-sm font-medium text-gray-700 mb-2">Entity Type</h4>
        <div class="flex flex-wrap gap-2">
          {#each allEntityTypes as type}
            <button
              type="button"
              onclick={() => toggleEntityType(type)}
              class="px-2.5 py-1 text-xs font-medium rounded-full {selectedEntityTypes.includes(type) ? 'bg-indigo-600 text-white' : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'}"
            >
              {type.replace(/_/g, ' ').toLowerCase()}
            </button>
          {/each}
        </div>
      </div>

      <!-- Actions -->
      <div>
        <h4 class="text-sm font-medium text-gray-700 mb-2">Action</h4>
        <div class="flex flex-wrap gap-2">
          {#each allActions as action}
            <button
              type="button"
              onclick={() => toggleAction(action)}
              class="px-2.5 py-1 text-xs font-medium rounded-full {selectedActions.includes(action) ? 'bg-indigo-600 text-white' : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'}"
            >
              {action.replace(/_/g, ' ').toLowerCase()}
            </button>
          {/each}
        </div>
      </div>

      <!-- Filter actions -->
      <div class="flex items-center gap-2 pt-2 border-t border-gray-200">
        <button
          type="button"
          onclick={applyFilters}
          class="px-3 py-1.5 text-sm font-medium text-white bg-indigo-600 rounded-md hover:bg-indigo-700"
        >
          Apply Filters
        </button>
        {#if hasActiveFilters}
          <button
            type="button"
            onclick={() => { clearFilters(); loadActivity(true); }}
            class="px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Clear All
          </button>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Activity list -->
  {#if loading}
    <div class="space-y-4 py-4">
      {#each [1, 2, 3, 4, 5] as _}
        <div class="flex gap-3">
          <div class="h-8 w-8 bg-gray-200 rounded-full animate-pulse flex-shrink-0"></div>
          <div class="flex-1">
            <div class="h-4 w-48 bg-gray-200 rounded animate-pulse mb-2"></div>
            <div class="h-3 w-64 bg-gray-200 rounded animate-pulse mb-1"></div>
            <div class="h-3 w-24 bg-gray-200 rounded animate-pulse"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else if error}
    <div class="rounded-md bg-red-50 p-4">
      <p class="text-sm text-red-700">{error}</p>
      <button
        type="button"
        onclick={() => loadActivity(true)}
        class="mt-2 text-sm text-red-600 hover:text-red-800 underline"
      >
        Try again
      </button>
    </div>
  {:else if events.length === 0}
    <div class="text-center py-12">
      <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
      </svg>
      <h3 class="mt-2 text-sm font-medium text-gray-900">No activity yet</h3>
      <p class="mt-1 text-sm text-gray-500">
        {hasActiveFilters ? 'No events match your filters.' : 'Activity will appear here as changes are made.'}
      </p>
    </div>
  {:else}
    <div class="bg-white rounded-lg border border-gray-200 divide-y divide-gray-100">
      {#each events as event (event.id)}
        <ActivityItem {event} />
      {/each}
    </div>

    <!-- Load more -->
    {#if pageInfo?.hasNextPage}
      <div class="flex justify-center pt-4">
        <button
          type="button"
          onclick={() => loadActivity(false)}
          disabled={loadingMore}
          class="px-4 py-2 text-sm font-medium text-indigo-600 bg-white border border-indigo-300 rounded-md hover:bg-indigo-50 disabled:opacity-50"
        >
          {loadingMore ? 'Loading...' : 'Load More'}
        </button>
      </div>
    {/if}
  {/if}
</div>
