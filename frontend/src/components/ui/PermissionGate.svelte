<script lang="ts">
  import { useHasPermission, type PermissionCode } from '../../lib/stores/permissions.svelte';
  import type { Snippet } from 'svelte';

  interface Props {
    permission: PermissionCode | string;
    resourceType: 'organization' | 'project' | 'board';
    resourceId: string;
    children: Snippet;
    fallback?: Snippet;
  }

  let { permission, resourceType, resourceId, children, fallback }: Props = $props();

  const permissionCheck = useHasPermission(permission, resourceType, resourceId);
  let allowed = $derived($permissionCheck.data ?? false);
  let loading = $derived($permissionCheck.isLoading);
</script>

{#if loading}
  <!-- Optionally show nothing or a loading state while checking -->
{:else if allowed}
  {@render children()}
{:else if fallback}
  {@render fallback()}
{/if}
