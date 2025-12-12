<script lang="ts">
  import type { AuditEvent } from '../../lib/api/activity';
  import { AuditAction, AuditEntityType } from '../../lib/graphql/generated';

  interface Props {
    event: AuditEvent;
    organizationId?: string;
  }

  let { event, organizationId }: Props = $props();

  const actionConfig: Record<AuditAction, { verb: string; color: string; icon: string }> = {
    [AuditAction.Created]: { verb: 'created', color: 'text-green-600 bg-green-100', icon: 'M12 4v16m8-8H4' },
    [AuditAction.Updated]: { verb: 'updated', color: 'text-blue-600 bg-blue-100', icon: 'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z' },
    [AuditAction.Deleted]: { verb: 'deleted', color: 'text-red-600 bg-red-100', icon: 'M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16' },
    [AuditAction.CardMoved]: { verb: 'moved', color: 'text-purple-600 bg-purple-100', icon: 'M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4' },
    [AuditAction.CardAssigned]: { verb: 'assigned', color: 'text-indigo-600 bg-indigo-100', icon: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' },
    [AuditAction.CardUnassigned]: { verb: 'unassigned', color: 'text-gray-600 bg-gray-100', icon: 'M13 7a4 4 0 11-8 0 4 4 0 018 0zM9 14a6 6 0 00-6 6v1h12v-1a6 6 0 00-6-6zM21 12h-6' },
    [AuditAction.SprintStarted]: { verb: 'started sprint', color: 'text-green-600 bg-green-100', icon: 'M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
    [AuditAction.SprintCompleted]: { verb: 'completed sprint', color: 'text-green-600 bg-green-100', icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z' },
    [AuditAction.CardAddedToSprint]: { verb: 'added to sprint', color: 'text-blue-600 bg-blue-100', icon: 'M12 4v16m8-8H4' },
    [AuditAction.CardRemovedFromSprint]: { verb: 'removed from sprint', color: 'text-orange-600 bg-orange-100', icon: 'M20 12H4' },
    [AuditAction.MemberInvited]: { verb: 'invited', color: 'text-indigo-600 bg-indigo-100', icon: 'M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z' },
    [AuditAction.MemberJoined]: { verb: 'joined', color: 'text-green-600 bg-green-100', icon: 'M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z' },
    [AuditAction.MemberRemoved]: { verb: 'removed', color: 'text-red-600 bg-red-100', icon: 'M13 7a4 4 0 11-8 0 4 4 0 018 0zM9 14a6 6 0 00-6 6v1h12v-1a6 6 0 00-6-6zM21 12h-6' },
    [AuditAction.MemberRoleChanged]: { verb: 'changed role', color: 'text-yellow-600 bg-yellow-100', icon: 'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z' },
    [AuditAction.ColumnReordered]: { verb: 'reordered columns', color: 'text-gray-600 bg-gray-100', icon: 'M4 6h16M4 12h16M4 18h16' },
    [AuditAction.ColumnVisibilityToggled]: { verb: 'toggled column visibility', color: 'text-gray-600 bg-gray-100', icon: 'M15 12a3 3 0 11-6 0 3 3 0 016 0z M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z' },
    [AuditAction.UserLoggedIn]: { verb: 'logged in', color: 'text-green-600 bg-green-100', icon: 'M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1' },
    [AuditAction.UserLoggedOut]: { verb: 'logged out', color: 'text-gray-600 bg-gray-100', icon: 'M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1' },
  };

  const entityLabels: Record<AuditEntityType, string> = {
    [AuditEntityType.User]: 'user',
    [AuditEntityType.Organization]: 'organization',
    [AuditEntityType.Project]: 'project',
    [AuditEntityType.Board]: 'board',
    [AuditEntityType.BoardColumn]: 'column',
    [AuditEntityType.Card]: 'card',
    [AuditEntityType.Sprint]: 'sprint',
    [AuditEntityType.Tag]: 'tag',
    [AuditEntityType.Role]: 'role',
    [AuditEntityType.Invitation]: 'invitation',
  };

  const config = $derived(actionConfig[event.action] || actionConfig[AuditAction.Updated]);
  const entityLabel = $derived(entityLabels[event.entityType] || event.entityType.toLowerCase());
  const actorName = $derived(event.actor?.displayName || event.actor?.username || 'System');
  const actorInitials = $derived(actorName.charAt(0).toUpperCase());

  function formatTimeAgo(date: string): string {
    const now = new Date();
    const then = new Date(date);
    const diffMs = now.getTime() - then.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);

    if (diffSec < 60) return 'just now';
    if (diffMin < 60) return `${diffMin}m ago`;
    if (diffHour < 24) return `${diffHour}h ago`;
    if (diffDay < 7) return `${diffDay}d ago`;
    return then.toLocaleDateString();
  }

  function getEntityName(): string {
    try {
      if (event.stateAfter) {
        const state = JSON.parse(event.stateAfter);
        return state.title || state.name || '';
      }
      if (event.stateBefore) {
        const state = JSON.parse(event.stateBefore);
        return state.title || state.name || '';
      }
    } catch {
      // Ignore parsing errors
    }
    return '';
  }

  function getMoveDetails(): { from: string; to: string } | null {
    if (event.action !== AuditAction.CardMoved || !event.metadata) return null;
    try {
      const metadata = JSON.parse(event.metadata);
      return {
        from: metadata.from_column_name || 'unknown',
        to: metadata.to_column_name || 'unknown',
      };
    } catch {
      return null;
    }
  }

  function getEntityLink(): string | null {
    // Don't link deleted entities
    if (event.action === AuditAction.Deleted) return null;

    const projectId = event.project?.id;
    const boardId = event.board?.id;

    switch (event.entityType) {
      case AuditEntityType.Card:
        if (projectId && boardId) {
          return `/projects/${projectId}/board/${boardId}?card=${event.entityId}`;
        }
        break;
      case AuditEntityType.Board:
        if (projectId) {
          return `/projects/${projectId}/board/${event.entityId}`;
        }
        break;
      case AuditEntityType.Project:
        return `/projects/${event.entityId}`;
      case AuditEntityType.Organization:
        return `/organizations/${event.entityId}`;
      case AuditEntityType.Sprint:
        if (projectId && boardId) {
          return `/projects/${projectId}/board/${boardId}`;
        }
        break;
    }
    return null;
  }

  function getProjectLink(): string | null {
    if (event.project?.id) {
      return `/projects/${event.project.id}`;
    }
    return null;
  }

  function getBoardLink(): string | null {
    if (event.project?.id && event.board?.id) {
      return `/projects/${event.project.id}/board/${event.board.id}`;
    }
    return null;
  }

  const entityName = $derived(getEntityName());
  const moveDetails = $derived(getMoveDetails());
  const entityLink = $derived(getEntityLink());
  const projectLink = $derived(getProjectLink());
  const boardLink = $derived(getBoardLink());
</script>

<div class="flex gap-3 px-4 py-3">
  <!-- Avatar/Icon -->
  <div class="flex-shrink-0">
    {#if event.actor?.avatarUrl}
      <img
        src={event.actor.avatarUrl}
        alt={actorName}
        class="h-8 w-8 rounded-full"
      />
    {:else}
      <div class="h-8 w-8 rounded-full bg-gray-200 flex items-center justify-center">
        <span class="text-gray-600 font-medium text-xs">{actorInitials}</span>
      </div>
    {/if}
  </div>

  <!-- Content -->
  <div class="flex-1 min-w-0">
    <div class="flex items-start gap-2">
      <div class="flex-1">
        <p class="text-sm text-gray-900">
          <span class="font-medium">{actorName}</span>
          {' '}
          <span class="text-gray-600">{config.verb}</span>
          {' '}
          {#if entityName}
            {#if entityLink}
              <a href={entityLink} class="font-semibold underline decoration-gray-300 hover:decoration-gray-500">{entityName}</a>
            {:else}
              <span class="font-medium">{entityName}</span>
            {/if}
          {:else}
            <span class="text-gray-600">{entityLabel}</span>
          {/if}
          {#if moveDetails}
            <span class="text-gray-600">
              from <span class="font-medium">{moveDetails.from}</span>
              to <span class="font-medium">{moveDetails.to}</span>
            </span>
          {/if}
        </p>
        <div class="flex items-center gap-2 mt-1">
          <span class="text-xs text-gray-500">{formatTimeAgo(event.occurredAt)}</span>
          {#if event.project}
            <span class="text-xs text-gray-400">in</span>
            {#if projectLink}
              <a href={projectLink} class="text-xs text-gray-600 underline decoration-gray-300 hover:decoration-gray-500">{event.project.name}</a>
            {:else}
              <span class="text-xs text-gray-600">{event.project.name}</span>
            {/if}
          {/if}
          {#if event.board}
            <span class="text-xs text-gray-400">/</span>
            {#if boardLink}
              <a href={boardLink} class="text-xs text-gray-600 underline decoration-gray-300 hover:decoration-gray-500">{event.board.name}</a>
            {:else}
              <span class="text-xs text-gray-600">{event.board.name}</span>
            {/if}
          {/if}
        </div>
      </div>
      <!-- Action badge -->
      <span class={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${config.color}`}>
        {event.action.replace(/_/g, ' ').toLowerCase()}
      </span>
    </div>
  </div>
</div>
