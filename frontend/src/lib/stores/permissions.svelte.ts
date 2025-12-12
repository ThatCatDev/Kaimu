import { useSWR, mutate } from 'sswr';
import * as rbacApi from '../api/rbac';

// Keys for caching
function permissionsKey(resourceType: string, resourceId: string): string {
  return `permissions:${resourceType}:${resourceId}`;
}

function hasPermissionKey(permission: string, resourceType: string, resourceId: string): string {
  return `hasPermission:${permission}:${resourceType}:${resourceId}`;
}

// Fetcher that only runs on client side
async function fetchPermissions(resourceType: string, resourceId: string): Promise<string[]> {
  if (typeof window === 'undefined') {
    return [];
  }
  return rbacApi.getMyPermissions(resourceType, resourceId);
}

async function fetchHasPermission(
  permission: string,
  resourceType: string,
  resourceId: string
): Promise<boolean> {
  if (typeof window === 'undefined') {
    return false;
  }
  return rbacApi.hasPermission(permission, resourceType, resourceId);
}

/**
 * Hook to get all permissions for a specific resource
 * @param resourceType - 'organization', 'project', or 'board'
 * @param resourceId - The ID of the resource
 */
export function usePermissions(resourceType: string, resourceId: string) {
  return useSWR<string[]>(
    permissionsKey(resourceType, resourceId),
    () => fetchPermissions(resourceType, resourceId)
  );
}

/**
 * Hook to check if the user has a specific permission
 * @param permission - The permission code (e.g., 'org:manage')
 * @param resourceType - 'organization', 'project', or 'board'
 * @param resourceId - The ID of the resource
 */
export function useHasPermission(permission: string, resourceType: string, resourceId: string) {
  return useSWR<boolean>(
    hasPermissionKey(permission, resourceType, resourceId),
    () => fetchHasPermission(permission, resourceType, resourceId)
  );
}

/**
 * Invalidate cached permissions for a resource
 * Call this after role changes or membership updates
 */
export function invalidatePermissions(resourceType: string, resourceId: string): void {
  mutate(permissionsKey(resourceType, resourceId), undefined);
}

/**
 * Invalidate a specific permission check
 */
export function invalidateHasPermission(
  permission: string,
  resourceType: string,
  resourceId: string
): void {
  mutate(hasPermissionKey(permission, resourceType, resourceId), undefined);
}

/**
 * Invalidate all cached permissions for a resource type
 * This is useful when a user's role changes organization-wide
 */
export function invalidateAllPermissions(): void {
  // Clear all permission-related cache entries
  // sswr will refetch on next access
  mutate((key: unknown) => typeof key === 'string' && key.startsWith('permissions:'), undefined);
  mutate((key: unknown) => typeof key === 'string' && key.startsWith('hasPermission:'), undefined);
}

// Common permission codes as constants
export const Permissions = {
  // Organization permissions
  ORG_VIEW: 'org:view',
  ORG_MANAGE: 'org:manage',
  ORG_DELETE: 'org:delete',
  ORG_INVITE: 'org:invite',
  ORG_REMOVE_MEMBERS: 'org:remove_members',
  ORG_MANAGE_ROLES: 'org:manage_roles',

  // Project permissions
  PROJECT_VIEW: 'project:view',
  PROJECT_CREATE: 'project:create',
  PROJECT_MANAGE: 'project:manage',
  PROJECT_DELETE: 'project:delete',
  PROJECT_MANAGE_MEMBERS: 'project:manage_members',

  // Board permissions
  BOARD_VIEW: 'board:view',
  BOARD_CREATE: 'board:create',
  BOARD_MANAGE: 'board:manage',
  BOARD_DELETE: 'board:delete',

  // Card permissions
  CARD_VIEW: 'card:view',
  CARD_CREATE: 'card:create',
  CARD_EDIT: 'card:edit',
  CARD_MOVE: 'card:move',
  CARD_DELETE: 'card:delete',
  CARD_ASSIGN: 'card:assign',

  // Sprint permissions
  SPRINT_VIEW: 'sprint:view',
  SPRINT_MANAGE: 'sprint:manage',
} as const;

export type PermissionCode = (typeof Permissions)[keyof typeof Permissions];
