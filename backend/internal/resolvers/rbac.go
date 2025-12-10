package resolvers

import (
	"context"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/graph/model"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/invitation"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/organization_member"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/permission"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/project_member"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/role"
	invitationSvc "github.com/thatcatdev/pulse-backend/internal/services/invitation"
	"github.com/thatcatdev/pulse-backend/internal/services/rbac"
)

// Permissions returns all available permissions
func Permissions(ctx context.Context, svc rbac.Service) ([]*model.Permission, error) {
	perms, err := svc.GetAllPermissions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Permission, len(perms))
	for i, p := range perms {
		result[i] = permissionToModel(p)
	}
	return result, nil
}

// Roles returns all roles for an organization
func Roles(ctx context.Context, svc rbac.Service, organizationID string) ([]*model.Role, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, err
	}

	// Check if user has permission to view roles
	hasAccess, err := svc.HasOrgPermission(ctx, *userID, orgID, "org:view")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	roles, err := svc.GetRolesForOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Role, len(roles))
	for i, r := range roles {
		result[i] = roleToModel(r)
	}
	return result, nil
}

// Role returns a specific role by ID
func Role(ctx context.Context, svc rbac.Service, id string) (*model.Role, error) {
	roleID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	r, err := svc.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	return roleToModel(r), nil
}

// RolePermissions resolves the permissions field for a Role
func RolePermissions(ctx context.Context, svc rbac.Service, r *model.Role) ([]*model.Permission, error) {
	roleID, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, err
	}

	perms, err := svc.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Permission, len(perms))
	for i, p := range perms {
		result[i] = permissionToModel(p)
	}
	return result, nil
}

// GetOrganizationMembersRBAC returns all members of an organization using RBAC service
func GetOrganizationMembersRBAC(ctx context.Context, svc rbac.Service, organizationID string) ([]*model.OrganizationMember, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasAccess, err := svc.HasOrgPermission(ctx, *userID, orgID, "org:view")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	members, err := svc.GetOrgMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.OrganizationMember, len(members))
	for i, m := range members {
		result[i] = orgMemberToModel(m)
	}
	return result, nil
}

// ProjectMembers returns all members of a project
func ProjectMembers(ctx context.Context, svc rbac.Service, projectID string) ([]*model.ProjectMember, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	projID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasAccess, err := svc.HasProjectPermission(ctx, *userID, projID, "project:view")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	members, err := svc.GetProjectMembers(ctx, projID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.ProjectMember, len(members))
	for i, m := range members {
		result[i] = projectMemberToModel(m)
	}
	return result, nil
}

// HasPermission checks if the current user has a specific permission
func HasPermission(ctx context.Context, svc rbac.Service, permissionCode, resourceType, resourceID string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, nil
	}

	resID, err := uuid.Parse(resourceID)
	if err != nil {
		return false, err
	}

	switch resourceType {
	case "organization":
		return svc.HasOrgPermission(ctx, *userID, resID, permissionCode)
	case "project":
		return svc.HasProjectPermission(ctx, *userID, resID, permissionCode)
	default:
		return false, nil
	}
}

// MyPermissions returns all permissions the current user has for a resource
func MyPermissions(ctx context.Context, svc rbac.Service, resourceType, resourceID string) ([]string, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return []string{}, nil
	}

	resID, err := uuid.Parse(resourceID)
	if err != nil {
		return nil, err
	}

	switch resourceType {
	case "organization":
		return svc.GetUserOrgPermissions(ctx, *userID, resID)
	case "project":
		return svc.GetUserProjectPermissions(ctx, *userID, resID)
	default:
		return []string{}, nil
	}
}

// CreateRole creates a new custom role
func CreateRole(ctx context.Context, svc rbac.Service, input model.CreateRoleInput) (*model.Role, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasAccess, err := svc.HasOrgPermission(ctx, *userID, orgID, "org:manage_roles")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	description := ""
	if input.Description != nil {
		description = *input.Description
	}

	r, err := svc.CreateRole(ctx, orgID, input.Name, description, input.PermissionCodes)
	if err != nil {
		return nil, err
	}

	return roleToModel(r), nil
}

// UpdateRole updates an existing custom role
func UpdateRole(ctx context.Context, svc rbac.Service, input model.UpdateRoleInput) (*model.Role, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	roleID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	// Get the role to find its organization
	existingRole, err := svc.GetRole(ctx, roleID)
	if err != nil {
		return nil, err
	}

	// System roles can't be modified
	if existingRole.IsSystem {
		return nil, rbac.ErrCannotModifySystem
	}

	// Check permission
	if existingRole.OrganizationID != nil {
		hasAccess, err := svc.HasOrgPermission(ctx, *userID, *existingRole.OrganizationID, "org:manage_roles")
		if err != nil {
			return nil, err
		}
		if !hasAccess {
			return nil, ErrUnauthorized
		}
	}

	r, err := svc.UpdateRole(ctx, roleID, input.Name, input.Description, input.PermissionCodes)
	if err != nil {
		return nil, err
	}

	return roleToModel(r), nil
}

// DeleteRole deletes a custom role
func DeleteRole(ctx context.Context, svc rbac.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	roleID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Get the role to find its organization
	existingRole, err := svc.GetRole(ctx, roleID)
	if err != nil {
		return false, err
	}

	// System roles can't be deleted
	if existingRole.IsSystem {
		return false, rbac.ErrCannotModifySystem
	}

	// Check permission
	if existingRole.OrganizationID != nil {
		hasAccess, err := svc.HasOrgPermission(ctx, *userID, *existingRole.OrganizationID, "org:manage_roles")
		if err != nil {
			return false, err
		}
		if !hasAccess {
			return false, ErrUnauthorized
		}
	}

	err = svc.DeleteRole(ctx, roleID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ChangeMemberRole changes a member's role in an organization
func ChangeMemberRole(ctx context.Context, svc rbac.Service, organizationID string, input model.ChangeMemberRoleInput) (*model.OrganizationMember, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, err
	}

	targetUserID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, err
	}

	roleID, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}

	// Check permission - need org:manage_roles or org:invite (admin-level)
	hasAccess, err := svc.HasOrgPermission(ctx, *userID, orgID, "org:manage_roles")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	member, err := svc.AssignOrgRole(ctx, orgID, targetUserID, roleID)
	if err != nil {
		return nil, err
	}

	return orgMemberToModel(member), nil
}

// RemoveMember removes a member from an organization
func RemoveMember(ctx context.Context, svc rbac.Service, organizationID, targetUserID string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return false, err
	}

	targetUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return false, err
	}

	// Check permission
	hasAccess, err := svc.HasOrgPermission(ctx, *userID, orgID, "org:remove_members")
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrUnauthorized
	}

	err = svc.RemoveOrgMember(ctx, orgID, targetUID, *userID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// AssignProjectRole assigns a project-specific role to a user
func AssignProjectRole(ctx context.Context, svc rbac.Service, input model.AssignProjectRoleInput) (*model.ProjectMember, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	projectID, err := uuid.Parse(input.ProjectID)
	if err != nil {
		return nil, err
	}

	targetUserID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasAccess, err := svc.HasProjectPermission(ctx, *userID, projectID, "project:manage_members")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	var roleID *uuid.UUID
	if input.RoleID != nil {
		parsed, err := uuid.Parse(*input.RoleID)
		if err != nil {
			return nil, err
		}
		roleID = &parsed
	}

	member, err := svc.AssignProjectRole(ctx, projectID, targetUserID, roleID)
	if err != nil {
		return nil, err
	}

	return projectMemberToModel(member), nil
}

// RemoveProjectMember removes a member from a project
func RemoveProjectMember(ctx context.Context, svc rbac.Service, projectID, targetUserID string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	projID, err := uuid.Parse(projectID)
	if err != nil {
		return false, err
	}

	targetUID, err := uuid.Parse(targetUserID)
	if err != nil {
		return false, err
	}

	// Check permission
	hasAccess, err := svc.HasProjectPermission(ctx, *userID, projID, "project:manage_members")
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrUnauthorized
	}

	err = svc.RemoveProjectMember(ctx, projID, targetUID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Model conversion helpers

func permissionToModel(p *permission.Permission) *model.Permission {
	var desc *string
	if p.Description != nil {
		desc = p.Description
	}
	return &model.Permission{
		ID:           p.ID.String(),
		Code:         p.Code,
		Name:         p.Name,
		Description:  desc,
		ResourceType: p.ResourceType,
	}
}

func roleToModel(r *role.Role) *model.Role {
	var desc *string
	if r.Description != nil {
		desc = r.Description
	}
	return &model.Role{
		ID:          r.ID.String(),
		Name:        r.Name,
		Description: desc,
		IsSystem:    r.IsSystem,
		Scope:       r.Scope,
		Permissions: nil, // Resolved by field resolver
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func orgMemberToModel(m *organization_member.OrganizationMember) *model.OrganizationMember {
	return &model.OrganizationMember{
		ID:         m.ID.String(),
		User:       nil, // Resolved by field resolver
		Role:       nil, // Resolved by field resolver
		LegacyRole: m.Role,
		CreatedAt:  m.CreatedAt,
	}
}

func projectMemberToModel(m *project_member.ProjectMember) *model.ProjectMember {
	return &model.ProjectMember{
		ID:        m.ID.String(),
		User:      nil, // Resolved by field resolver
		Role:      nil, // Resolved by field resolver
		Project:   nil, // Resolved by field resolver
		CreatedAt: m.CreatedAt,
	}
}

func invitationToModel(inv *invitation.Invitation) *model.Invitation {
	return &model.Invitation{
		ID:           inv.ID.String(),
		Email:        inv.Email,
		Token:        inv.Token,
		Role:         nil, // Resolved by field resolver
		Organization: nil, // Resolved by field resolver
		InvitedBy:    nil, // Resolved by field resolver
		ExpiresAt:    inv.ExpiresAt,
		CreatedAt:    inv.CreatedAt,
	}
}

// Invitation resolvers

// Invitations returns all pending invitations for an organization
func Invitations(ctx context.Context, svc invitationSvc.Service, rbacSvc rbac.Service, organizationID string) ([]*model.Invitation, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasAccess, err := rbacSvc.HasOrgPermission(ctx, *userID, orgID, "org:invite")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	invitations, err := svc.GetPendingInvitations(ctx, orgID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Invitation, len(invitations))
	for i, inv := range invitations {
		result[i] = invitationToModel(inv)
	}
	return result, nil
}

// InviteMember creates a new invitation
func InviteMember(ctx context.Context, svc invitationSvc.Service, rbacSvc rbac.Service, input model.InviteMemberInput) (*model.Invitation, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		return nil, err
	}

	roleID, err := uuid.Parse(input.RoleID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasAccess, err := rbacSvc.HasOrgPermission(ctx, *userID, orgID, "org:invite")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	inv, err := svc.CreateInvitation(ctx, orgID, input.Email, roleID, *userID)
	if err != nil {
		return nil, err
	}

	return invitationToModel(inv), nil
}

// CancelInvitation cancels a pending invitation
func CancelInvitation(ctx context.Context, svc invitationSvc.Service, rbacSvc rbac.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	invID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Get invitation to check permission
	inv, err := svc.GetInvitation(ctx, invID)
	if err != nil {
		return false, err
	}

	// Check permission
	hasAccess, err := rbacSvc.HasOrgPermission(ctx, *userID, inv.OrganizationID, "org:invite")
	if err != nil {
		return false, err
	}
	if !hasAccess {
		return false, ErrUnauthorized
	}

	err = svc.CancelInvitation(ctx, invID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ResendInvitation resends an invitation
func ResendInvitation(ctx context.Context, svc invitationSvc.Service, rbacSvc rbac.Service, id string) (*model.Invitation, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	invID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	// Get invitation to check permission
	existingInv, err := svc.GetInvitation(ctx, invID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasAccess, err := rbacSvc.HasOrgPermission(ctx, *userID, existingInv.OrganizationID, "org:invite")
	if err != nil {
		return nil, err
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	inv, err := svc.ResendInvitation(ctx, invID)
	if err != nil {
		return nil, err
	}

	return invitationToModel(inv), nil
}

// AcceptInvitation accepts an invitation and joins the organization
func AcceptInvitation(ctx context.Context, svc invitationSvc.Service, token string) (*model.Organization, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	org, err := svc.AcceptInvitation(ctx, token, *userID)
	if err != nil {
		return nil, err
	}

	return organizationToModel(org), nil
}

// Field resolvers for OrganizationMember

// OrgMemberUser resolves the user field of OrganizationMember
func OrgMemberUser(ctx context.Context, svc rbac.Service, member *model.OrganizationMember) (*model.User, error) {
	memberID, err := uuid.Parse(member.ID)
	if err != nil {
		return nil, err
	}

	user, err := svc.GetOrgMemberUser(ctx, memberID)
	if err != nil {
		return nil, err
	}

	return UserToModel(user), nil
}

// OrgMemberRole resolves the role field of OrganizationMember
func OrgMemberRole(ctx context.Context, svc rbac.Service, member *model.OrganizationMember) (*model.Role, error) {
	memberID, err := uuid.Parse(member.ID)
	if err != nil {
		return nil, err
	}

	r, err := svc.GetOrgMemberRole(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}

	return roleToModel(r), nil
}

// Field resolvers for ProjectMember

// ProjectMemberUser resolves the user field of ProjectMember
func ProjectMemberUser(ctx context.Context, svc rbac.Service, member *model.ProjectMember) (*model.User, error) {
	memberID, err := uuid.Parse(member.ID)
	if err != nil {
		return nil, err
	}

	user, err := svc.GetProjectMemberUser(ctx, memberID)
	if err != nil {
		return nil, err
	}

	return UserToModel(user), nil
}

// ProjectMemberRole resolves the role field of ProjectMember
func ProjectMemberRole(ctx context.Context, svc rbac.Service, member *model.ProjectMember) (*model.Role, error) {
	memberID, err := uuid.Parse(member.ID)
	if err != nil {
		return nil, err
	}

	r, err := svc.GetProjectMemberRole(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}

	return roleToModel(r), nil
}

// ProjectMemberProject resolves the project field of ProjectMember
func ProjectMemberProject(ctx context.Context, svc rbac.Service, member *model.ProjectMember) (*model.Project, error) {
	memberID, err := uuid.Parse(member.ID)
	if err != nil {
		return nil, err
	}

	proj, err := svc.GetProjectMemberProject(ctx, memberID)
	if err != nil {
		return nil, err
	}

	return projectToModel(proj), nil
}

// Field resolvers for Invitation

// InvitationRole resolves the role field of Invitation
func InvitationRole(ctx context.Context, svc invitationSvc.Service, inv *model.Invitation) (*model.Role, error) {
	invID, err := uuid.Parse(inv.ID)
	if err != nil {
		return nil, err
	}

	r, err := svc.GetInvitationRole(ctx, invID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}

	return roleToModel(r), nil
}

// InvitationOrganization resolves the organization field of Invitation
func InvitationOrganization(ctx context.Context, svc invitationSvc.Service, inv *model.Invitation) (*model.Organization, error) {
	invID, err := uuid.Parse(inv.ID)
	if err != nil {
		return nil, err
	}

	org, err := svc.GetInvitationOrganization(ctx, invID)
	if err != nil {
		return nil, err
	}

	return organizationToModel(org), nil
}

// InvitationInvitedBy resolves the invitedBy field of Invitation
func InvitationInvitedBy(ctx context.Context, svc invitationSvc.Service, inv *model.Invitation) (*model.User, error) {
	invID, err := uuid.Parse(inv.ID)
	if err != nil {
		return nil, err
	}

	user, err := svc.GetInviter(ctx, invID)
	if err != nil {
		return nil, err
	}

	return UserToModel(user), nil
}
