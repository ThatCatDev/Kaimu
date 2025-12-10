import { graphql } from './client';
import type {
  PermissionsQuery,
  RolesQuery,
  RolesQueryVariables,
  RoleQuery,
  RoleQueryVariables,
  OrganizationMembersQuery,
  OrganizationMembersQueryVariables,
  ProjectMembersQuery,
  ProjectMembersQueryVariables,
  InvitationsQuery,
  InvitationsQueryVariables,
  HasPermissionQuery,
  HasPermissionQueryVariables,
  MyPermissionsQuery,
  MyPermissionsQueryVariables,
  CreateRoleMutation,
  CreateRoleMutationVariables,
  UpdateRoleMutation,
  UpdateRoleMutationVariables,
  DeleteRoleMutation,
  DeleteRoleMutationVariables,
  InviteMemberMutation,
  InviteMemberMutationVariables,
  CancelInvitationMutation,
  CancelInvitationMutationVariables,
  ResendInvitationMutation,
  ResendInvitationMutationVariables,
  AcceptInvitationMutation,
  AcceptInvitationMutationVariables,
  ChangeMemberRoleMutation,
  ChangeMemberRoleMutationVariables,
  RemoveMemberMutation,
  RemoveMemberMutationVariables,
  AssignProjectRoleMutation,
  AssignProjectRoleMutationVariables,
  RemoveProjectMemberMutation,
  RemoveProjectMemberMutationVariables,
} from '../graphql/generated';

// Type exports
export type Permission = PermissionsQuery['permissions'][number];
export type Role = RolesQuery['roles'][number];
export type RoleWithDetails = NonNullable<RoleQuery['role']>;
export type OrganizationMember = OrganizationMembersQuery['organizationMembers'][number];
export type ProjectMember = ProjectMembersQuery['projectMembers'][number];
export type Invitation = InvitationsQuery['invitations'][number];

// Mutation result types (may have fewer fields than query types)
type CreatedRole = CreateRoleMutation['createRole'];
type UpdatedRole = UpdateRoleMutation['updateRole'];
type CreatedInvitation = InviteMemberMutation['inviteMember'];
type ResendedInvitation = ResendInvitationMutation['resendInvitation'];
type ChangedMemberRole = ChangeMemberRoleMutation['changeMemberRole'];
type AssignedProjectRole = AssignProjectRoleMutation['assignProjectRole'];

// Queries

const PERMISSIONS_QUERY = `
  query Permissions {
    permissions {
      id
      code
      name
      description
      resourceType
    }
  }
`;

const ROLES_QUERY = `
  query Roles($organizationId: ID!) {
    roles(organizationId: $organizationId) {
      id
      name
      description
      isSystem
      scope
      createdAt
      updatedAt
      permissions {
        id
        code
        name
      }
    }
  }
`;

const ROLE_QUERY = `
  query Role($id: ID!) {
    role(id: $id) {
      id
      name
      description
      isSystem
      scope
      createdAt
      updatedAt
      permissions {
        id
        code
        name
        description
        resourceType
      }
    }
  }
`;

const ORGANIZATION_MEMBERS_QUERY = `
  query OrganizationMembers($organizationId: ID!) {
    organizationMembers(organizationId: $organizationId) {
      id
      legacyRole
      createdAt
      user {
        id
        email
        displayName
      }
      role {
        id
        name
        description
        isSystem
      }
    }
  }
`;

const PROJECT_MEMBERS_QUERY = `
  query ProjectMembers($projectId: ID!) {
    projectMembers(projectId: $projectId) {
      id
      createdAt
      user {
        id
        email
        displayName
      }
      role {
        id
        name
        description
        isSystem
      }
      project {
        id
        name
      }
    }
  }
`;

const INVITATIONS_QUERY = `
  query Invitations($organizationId: ID!) {
    invitations(organizationId: $organizationId) {
      id
      email
      expiresAt
      createdAt
      role {
        id
        name
      }
      invitedBy {
        id
        email
        displayName
      }
    }
  }
`;

const HAS_PERMISSION_QUERY = `
  query HasPermission($permission: String!, $resourceType: String!, $resourceId: ID!) {
    hasPermission(permission: $permission, resourceType: $resourceType, resourceId: $resourceId)
  }
`;

const MY_PERMISSIONS_QUERY = `
  query MyPermissions($resourceType: String!, $resourceId: ID!) {
    myPermissions(resourceType: $resourceType, resourceId: $resourceId)
  }
`;

// Mutations

const CREATE_ROLE_MUTATION = `
  mutation CreateRole($input: CreateRoleInput!) {
    createRole(input: $input) {
      id
      name
      description
      isSystem
      scope
      createdAt
      updatedAt
      permissions {
        id
        code
        name
      }
    }
  }
`;

const UPDATE_ROLE_MUTATION = `
  mutation UpdateRole($input: UpdateRoleInput!) {
    updateRole(input: $input) {
      id
      name
      description
      isSystem
      scope
      updatedAt
      permissions {
        id
        code
        name
      }
    }
  }
`;

const DELETE_ROLE_MUTATION = `
  mutation DeleteRole($id: ID!) {
    deleteRole(id: $id)
  }
`;

const INVITE_MEMBER_MUTATION = `
  mutation InviteMember($input: InviteMemberInput!) {
    inviteMember(input: $input) {
      id
      email
      token
      expiresAt
      createdAt
      role {
        id
        name
      }
    }
  }
`;

const CANCEL_INVITATION_MUTATION = `
  mutation CancelInvitation($id: ID!) {
    cancelInvitation(id: $id)
  }
`;

const RESEND_INVITATION_MUTATION = `
  mutation ResendInvitation($id: ID!) {
    resendInvitation(id: $id) {
      id
      email
      expiresAt
      createdAt
    }
  }
`;

const ACCEPT_INVITATION_MUTATION = `
  mutation AcceptInvitation($token: String!) {
    acceptInvitation(token: $token) {
      id
      name
      slug
    }
  }
`;

const CHANGE_MEMBER_ROLE_MUTATION = `
  mutation ChangeMemberRole($organizationId: ID!, $input: ChangeMemberRoleInput!) {
    changeMemberRole(organizationId: $organizationId, input: $input) {
      id
      legacyRole
      user {
        id
        email
        displayName
      }
      role {
        id
        name
      }
    }
  }
`;

const REMOVE_MEMBER_MUTATION = `
  mutation RemoveMember($organizationId: ID!, $userId: ID!) {
    removeMember(organizationId: $organizationId, userId: $userId)
  }
`;

const ASSIGN_PROJECT_ROLE_MUTATION = `
  mutation AssignProjectRole($input: AssignProjectRoleInput!) {
    assignProjectRole(input: $input) {
      id
      user {
        id
        email
        displayName
      }
      role {
        id
        name
      }
      project {
        id
        name
      }
    }
  }
`;

const REMOVE_PROJECT_MEMBER_MUTATION = `
  mutation RemoveProjectMember($projectId: ID!, $userId: ID!) {
    removeProjectMember(projectId: $projectId, userId: $userId)
  }
`;

// API Functions

export async function getPermissions(): Promise<Permission[]> {
  const data = await graphql<PermissionsQuery>(PERMISSIONS_QUERY);
  return data.permissions;
}

export async function getRoles(organizationId: string): Promise<Role[]> {
  const data = await graphql<RolesQuery>(ROLES_QUERY, {
    organizationId,
  } as RolesQueryVariables);
  return data.roles;
}

export async function getRole(id: string): Promise<RoleWithDetails | null> {
  const data = await graphql<RoleQuery>(ROLE_QUERY, { id } as RoleQueryVariables);
  return data.role ?? null;
}

export async function getOrganizationMembers(organizationId: string): Promise<OrganizationMember[]> {
  const data = await graphql<OrganizationMembersQuery>(ORGANIZATION_MEMBERS_QUERY, {
    organizationId,
  } as OrganizationMembersQueryVariables);
  return data.organizationMembers;
}

export async function getProjectMembers(projectId: string): Promise<ProjectMember[]> {
  const data = await graphql<ProjectMembersQuery>(PROJECT_MEMBERS_QUERY, {
    projectId,
  } as ProjectMembersQueryVariables);
  return data.projectMembers;
}

export async function getInvitations(organizationId: string): Promise<Invitation[]> {
  const data = await graphql<InvitationsQuery>(INVITATIONS_QUERY, {
    organizationId,
  } as InvitationsQueryVariables);
  return data.invitations;
}

export async function hasPermission(
  permission: string,
  resourceType: string,
  resourceId: string
): Promise<boolean> {
  const data = await graphql<HasPermissionQuery>(HAS_PERMISSION_QUERY, {
    permission,
    resourceType,
    resourceId,
  } as HasPermissionQueryVariables);
  return data.hasPermission;
}

export async function getMyPermissions(resourceType: string, resourceId: string): Promise<string[]> {
  const data = await graphql<MyPermissionsQuery>(MY_PERMISSIONS_QUERY, {
    resourceType,
    resourceId,
  } as MyPermissionsQueryVariables);
  return data.myPermissions;
}

export async function createRole(
  organizationId: string,
  name: string,
  description?: string,
  permissionCodes?: string[]
): Promise<CreatedRole> {
  const data = await graphql<CreateRoleMutation>(CREATE_ROLE_MUTATION, {
    input: { organizationId, name, description, permissionCodes },
  } as CreateRoleMutationVariables);
  return data.createRole;
}

export async function updateRole(
  id: string,
  updates: { name?: string; description?: string; permissionCodes?: string[] }
): Promise<UpdatedRole> {
  const data = await graphql<UpdateRoleMutation>(UPDATE_ROLE_MUTATION, {
    input: { id, ...updates },
  } as UpdateRoleMutationVariables);
  return data.updateRole;
}

export async function deleteRole(id: string): Promise<boolean> {
  const data = await graphql<DeleteRoleMutation>(DELETE_ROLE_MUTATION, {
    id,
  } as DeleteRoleMutationVariables);
  return data.deleteRole;
}

export async function inviteMember(
  organizationId: string,
  email: string,
  roleId: string
): Promise<CreatedInvitation> {
  const data = await graphql<InviteMemberMutation>(INVITE_MEMBER_MUTATION, {
    input: { organizationId, email, roleId },
  } as InviteMemberMutationVariables);
  return data.inviteMember;
}

export async function cancelInvitation(id: string): Promise<boolean> {
  const data = await graphql<CancelInvitationMutation>(CANCEL_INVITATION_MUTATION, {
    id,
  } as CancelInvitationMutationVariables);
  return data.cancelInvitation;
}

export async function resendInvitation(id: string): Promise<ResendedInvitation> {
  const data = await graphql<ResendInvitationMutation>(RESEND_INVITATION_MUTATION, {
    id,
  } as ResendInvitationMutationVariables);
  return data.resendInvitation;
}

export async function acceptInvitation(token: string): Promise<{ id: string; name: string; slug: string }> {
  const data = await graphql<AcceptInvitationMutation>(ACCEPT_INVITATION_MUTATION, {
    token,
  } as AcceptInvitationMutationVariables);
  return data.acceptInvitation;
}

export async function changeMemberRole(
  organizationId: string,
  userId: string,
  roleId: string
): Promise<ChangedMemberRole> {
  const data = await graphql<ChangeMemberRoleMutation>(CHANGE_MEMBER_ROLE_MUTATION, {
    organizationId,
    input: { userId, roleId },
  } as ChangeMemberRoleMutationVariables);
  return data.changeMemberRole;
}

export async function removeMember(organizationId: string, userId: string): Promise<boolean> {
  const data = await graphql<RemoveMemberMutation>(REMOVE_MEMBER_MUTATION, {
    organizationId,
    userId,
  } as RemoveMemberMutationVariables);
  return data.removeMember;
}

export async function assignProjectRole(
  projectId: string,
  userId: string,
  roleId?: string
): Promise<AssignedProjectRole> {
  const data = await graphql<AssignProjectRoleMutation>(ASSIGN_PROJECT_ROLE_MUTATION, {
    input: { projectId, userId, roleId },
  } as AssignProjectRoleMutationVariables);
  return data.assignProjectRole;
}

export async function removeProjectMember(projectId: string, userId: string): Promise<boolean> {
  const data = await graphql<RemoveProjectMemberMutation>(REMOVE_PROJECT_MEMBER_MUTATION, {
    projectId,
    userId,
  } as RemoveProjectMemberMutationVariables);
  return data.removeProjectMember;
}
