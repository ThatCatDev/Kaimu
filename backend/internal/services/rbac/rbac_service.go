package rbac

//go:generate mockgen -source=rbac_service.go -destination=mocks/rbac_service_mock.go -package=mocks

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/organization_member"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/permission"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/project_member"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/role"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/role_permission"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"github.com/thatcatdev/pulse-backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrCannotModifySystem = errors.New("cannot modify system role")
	ErrCannotDeleteOwner  = errors.New("cannot delete owner role assignment")
	ErrLastOwner          = errors.New("cannot remove the last owner")
	ErrInvalidPermission  = errors.New("invalid permission code")
)

type Service interface {
	// Permission checks
	HasOrgPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) (bool, error)
	HasProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permission string) (bool, error)
	HasBoardPermission(ctx context.Context, userID, boardID uuid.UUID, permission string) (bool, error)
	GetUserOrgPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error)
	GetUserProjectPermissions(ctx context.Context, userID, projectID uuid.UUID) ([]string, error)

	// Role queries
	GetAllPermissions(ctx context.Context) ([]*permission.Permission, error)
	GetRolesForOrg(ctx context.Context, orgID uuid.UUID) ([]*role.Role, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*role.Role, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*permission.Permission, error)

	// Role management
	CreateRole(ctx context.Context, orgID uuid.UUID, name, description string, permissionCodes []string) (*role.Role, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, name, description *string, permissionCodes []string) (*role.Role, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error

	// Role assignments
	AssignOrgRole(ctx context.Context, orgID, userID, roleID uuid.UUID) (*organization_member.OrganizationMember, error)
	AssignProjectRole(ctx context.Context, projectID, userID uuid.UUID, roleID *uuid.UUID) (*project_member.ProjectMember, error)
	GetUserOrgRole(ctx context.Context, orgID, userID uuid.UUID) (*role.Role, error)
	GetUserProjectRole(ctx context.Context, projectID, userID uuid.UUID) (*role.Role, error)

	// Member queries
	GetOrgMembers(ctx context.Context, orgID uuid.UUID) ([]*organization_member.OrganizationMember, error)
	GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*project_member.ProjectMember, error)
	RemoveOrgMember(ctx context.Context, orgID, userID, actorID uuid.UUID) error
	RemoveProjectMember(ctx context.Context, projectID, userID uuid.UUID) error

	// Field resolver helpers for OrganizationMember
	GetOrgMemberUser(ctx context.Context, memberID uuid.UUID) (*user.User, error)
	GetOrgMemberRole(ctx context.Context, memberID uuid.UUID) (*role.Role, error)

	// Field resolver helpers for ProjectMember
	GetProjectMemberUser(ctx context.Context, memberID uuid.UUID) (*user.User, error)
	GetProjectMemberRole(ctx context.Context, memberID uuid.UUID) (*role.Role, error)
	GetProjectMemberProject(ctx context.Context, memberID uuid.UUID) (*project.Project, error)
}

type service struct {
	permissionRepo     permission.Repository
	roleRepo           role.Repository
	rolePermissionRepo role_permission.Repository
	orgMemberRepo      organization_member.Repository
	projectMemberRepo  project_member.Repository
	projectRepo        project.Repository
	userRepo           user.Repository
}

func NewService(
	permissionRepo permission.Repository,
	roleRepo role.Repository,
	rolePermissionRepo role_permission.Repository,
	orgMemberRepo organization_member.Repository,
	projectMemberRepo project_member.Repository,
	projectRepo project.Repository,
	userRepo user.Repository,
) Service {
	return &service{
		permissionRepo:     permissionRepo,
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
		orgMemberRepo:      orgMemberRepo,
		projectMemberRepo:  projectMemberRepo,
		projectRepo:        projectRepo,
		userRepo:           userRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "rbac.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "rbac"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

// HasOrgPermission checks if a user has a specific permission in an organization
func (s *service) HasOrgPermission(ctx context.Context, userID, orgID uuid.UUID, permissionCode string) (bool, error) {
	ctx, span := s.startServiceSpan(ctx, "HasOrgPermission")
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("org.id", orgID.String()),
		attribute.String("permission", permissionCode),
	)
	defer span.End()

	permissions, err := s.GetUserOrgPermissions(ctx, userID, orgID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == permissionCode {
			return true, nil
		}
	}
	return false, nil
}

// HasProjectPermission checks if a user has a specific permission in a project
func (s *service) HasProjectPermission(ctx context.Context, userID, projectID uuid.UUID, permissionCode string) (bool, error) {
	ctx, span := s.startServiceSpan(ctx, "HasProjectPermission")
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("project.id", projectID.String()),
		attribute.String("permission", permissionCode),
	)
	defer span.End()

	permissions, err := s.GetUserProjectPermissions(ctx, userID, projectID)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == permissionCode {
			return true, nil
		}
	}
	return false, nil
}

// HasBoardPermission checks if a user has a specific permission for a board
// Boards inherit permissions from their parent project
func (s *service) HasBoardPermission(ctx context.Context, userID, boardID uuid.UUID, permissionCode string) (bool, error) {
	ctx, span := s.startServiceSpan(ctx, "HasBoardPermission")
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("board.id", boardID.String()),
		attribute.String("permission", permissionCode),
	)
	defer span.End()

	// For now, board permissions inherit from project
	// We need to get the project ID from the board
	// This requires access to the board repository, which we don't have here
	// Instead, we'll need the caller to provide the project ID or add board repo
	// For now, return permission denied - this should be called with project ID instead
	return false, errors.New("use HasProjectPermission for board operations")
}

// GetUserOrgPermissions returns all permission codes a user has in an organization
func (s *service) GetUserOrgPermissions(ctx context.Context, userID, orgID uuid.UUID) ([]string, error) {
	ctx, span := s.startServiceSpan(ctx, "GetUserOrgPermissions")
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("org.id", orgID.String()),
	)
	defer span.End()

	// Get user's organization membership
	member, err := s.orgMemberRepo.GetByOrgAndUser(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []string{}, nil // Not a member, no permissions
		}
		return nil, err
	}

	// Get the role ID (prefer RoleID, fall back to legacy Role field)
	var roleID uuid.UUID
	if member.RoleID != nil {
		roleID = *member.RoleID
	} else {
		// Legacy fallback
		switch member.Role {
		case "owner":
			roleID = role.OwnerRoleID
		case "admin":
			roleID = role.AdminRoleID
		case "member":
			roleID = role.MemberRoleID
		default:
			roleID = role.ViewerRoleID
		}
	}

	// Get permissions for this role
	return s.rolePermissionRepo.GetPermissionCodesByRoleID(ctx, roleID)
}

// GetUserProjectPermissions returns all permission codes a user has in a project
func (s *service) GetUserProjectPermissions(ctx context.Context, userID, projectID uuid.UUID) ([]string, error) {
	ctx, span := s.startServiceSpan(ctx, "GetUserProjectPermissions")
	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.String("project.id", projectID.String()),
	)
	defer span.End()

	// Get the project to find its organization
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Check for project-specific role first
	projectMember, err := s.projectMemberRepo.GetByProjectAndUser(ctx, projectID, userID)
	if err == nil && projectMember != nil && projectMember.RoleID != nil {
		// User has project-specific role
		return s.rolePermissionRepo.GetPermissionCodesByRoleID(ctx, *projectMember.RoleID)
	}

	// Fall back to organization role
	return s.GetUserOrgPermissions(ctx, userID, proj.OrganizationID)
}

// GetAllPermissions returns all defined permissions
func (s *service) GetAllPermissions(ctx context.Context) ([]*permission.Permission, error) {
	ctx, span := s.startServiceSpan(ctx, "GetAllPermissions")
	defer span.End()

	return s.permissionRepo.GetAll(ctx)
}

// GetRolesForOrg returns all roles available for an organization (system + custom)
func (s *service) GetRolesForOrg(ctx context.Context, orgID uuid.UUID) ([]*role.Role, error) {
	ctx, span := s.startServiceSpan(ctx, "GetRolesForOrg")
	span.SetAttributes(attribute.String("org.id", orgID.String()))
	defer span.End()

	return s.roleRepo.GetAllForOrg(ctx, orgID)
}

// GetRole returns a role by ID
func (s *service) GetRole(ctx context.Context, roleID uuid.UUID) (*role.Role, error) {
	ctx, span := s.startServiceSpan(ctx, "GetRole")
	span.SetAttributes(attribute.String("role.id", roleID.String()))
	defer span.End()

	r, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return r, nil
}

// GetRolePermissions returns all permissions for a role
func (s *service) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*permission.Permission, error) {
	ctx, span := s.startServiceSpan(ctx, "GetRolePermissions")
	span.SetAttributes(attribute.String("role.id", roleID.String()))
	defer span.End()

	return s.rolePermissionRepo.GetPermissionsByRoleID(ctx, roleID)
}

// CreateRole creates a new custom role for an organization
func (s *service) CreateRole(ctx context.Context, orgID uuid.UUID, name, description string, permissionCodes []string) (*role.Role, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateRole")
	span.SetAttributes(
		attribute.String("org.id", orgID.String()),
		attribute.String("role.name", name),
	)
	defer span.End()

	// Get permission IDs from codes
	permissions, err := s.permissionRepo.GetByCodes(ctx, permissionCodes)
	if err != nil {
		return nil, err
	}
	if len(permissions) != len(permissionCodes) {
		return nil, ErrInvalidPermission
	}

	// Create the role
	desc := description
	newRole := &role.Role{
		OrganizationID: &orgID,
		Name:           name,
		Description:    &desc,
		IsSystem:       false,
		Scope:          "organization",
	}

	if err := s.roleRepo.Create(ctx, newRole); err != nil {
		return nil, err
	}

	// Assign permissions
	permissionIDs := make([]uuid.UUID, len(permissions))
	for i, p := range permissions {
		permissionIDs[i] = p.ID
	}

	if err := s.rolePermissionRepo.CreateBatch(ctx, newRole.ID, permissionIDs); err != nil {
		return nil, err
	}

	return newRole, nil
}

// UpdateRole updates a custom role
func (s *service) UpdateRole(ctx context.Context, roleID uuid.UUID, name, description *string, permissionCodes []string) (*role.Role, error) {
	ctx, span := s.startServiceSpan(ctx, "UpdateRole")
	span.SetAttributes(attribute.String("role.id", roleID.String()))
	defer span.End()

	// Get existing role
	existingRole, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	// Cannot modify system roles
	if existingRole.IsSystem {
		return nil, ErrCannotModifySystem
	}

	// Update fields
	if name != nil {
		existingRole.Name = *name
	}
	if description != nil {
		existingRole.Description = description
	}

	if err := s.roleRepo.Update(ctx, existingRole); err != nil {
		return nil, err
	}

	// Update permissions if provided
	if permissionCodes != nil {
		permissions, err := s.permissionRepo.GetByCodes(ctx, permissionCodes)
		if err != nil {
			return nil, err
		}
		if len(permissions) != len(permissionCodes) {
			return nil, ErrInvalidPermission
		}

		permissionIDs := make([]uuid.UUID, len(permissions))
		for i, p := range permissions {
			permissionIDs[i] = p.ID
		}

		if err := s.rolePermissionRepo.ReplaceForRole(ctx, roleID, permissionIDs); err != nil {
			return nil, err
		}
	}

	return existingRole, nil
}

// DeleteRole deletes a custom role
func (s *service) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteRole")
	span.SetAttributes(attribute.String("role.id", roleID.String()))
	defer span.End()

	// Get existing role
	existingRole, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	// Cannot delete system roles
	if existingRole.IsSystem {
		return ErrCannotModifySystem
	}

	return s.roleRepo.Delete(ctx, roleID)
}

// AssignOrgRole assigns a role to a user in an organization
func (s *service) AssignOrgRole(ctx context.Context, orgID, userID, roleID uuid.UUID) (*organization_member.OrganizationMember, error) {
	ctx, span := s.startServiceSpan(ctx, "AssignOrgRole")
	span.SetAttributes(
		attribute.String("org.id", orgID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("role.id", roleID.String()),
	)
	defer span.End()

	// Get existing membership
	member, err := s.orgMemberRepo.GetByOrgAndUser(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	// If demoting from owner, check there are other owners
	if member.RoleID != nil && *member.RoleID == role.OwnerRoleID && roleID != role.OwnerRoleID {
		ownerCount, err := s.countOrgOwners(ctx, orgID)
		if err != nil {
			return nil, err
		}
		if ownerCount <= 1 {
			return nil, ErrLastOwner
		}
	}

	// Update role
	member.RoleID = &roleID
	member.Role = "" // Clear legacy field

	if err := s.orgMemberRepo.Update(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// countOrgOwners counts the number of owners in an organization
func (s *service) countOrgOwners(ctx context.Context, orgID uuid.UUID) (int, error) {
	members, err := s.orgMemberRepo.GetByOrgID(ctx, orgID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, m := range members {
		if m.RoleID != nil && *m.RoleID == role.OwnerRoleID {
			count++
		} else if m.Role == "owner" {
			count++
		}
	}
	return count, nil
}

// AssignProjectRole assigns a project-specific role to a user
func (s *service) AssignProjectRole(ctx context.Context, projectID, userID uuid.UUID, roleID *uuid.UUID) (*project_member.ProjectMember, error) {
	ctx, span := s.startServiceSpan(ctx, "AssignProjectRole")
	span.SetAttributes(
		attribute.String("project.id", projectID.String()),
		attribute.String("user.id", userID.String()),
	)
	defer span.End()

	// Check if member exists
	member, err := s.projectMemberRepo.GetByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new project member
			member = &project_member.ProjectMember{
				ProjectID: projectID,
				UserID:    userID,
				RoleID:    roleID,
			}
			if err := s.projectMemberRepo.Create(ctx, member); err != nil {
				return nil, err
			}
			return member, nil
		}
		return nil, err
	}

	// Update existing
	member.RoleID = roleID
	if err := s.projectMemberRepo.Update(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// GetUserOrgRole returns a user's role in an organization
func (s *service) GetUserOrgRole(ctx context.Context, orgID, userID uuid.UUID) (*role.Role, error) {
	ctx, span := s.startServiceSpan(ctx, "GetUserOrgRole")
	span.SetAttributes(
		attribute.String("org.id", orgID.String()),
		attribute.String("user.id", userID.String()),
	)
	defer span.End()

	member, err := s.orgMemberRepo.GetByOrgAndUser(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}

	var roleID uuid.UUID
	if member.RoleID != nil {
		roleID = *member.RoleID
	} else {
		// Legacy fallback
		switch member.Role {
		case "owner":
			roleID = role.OwnerRoleID
		case "admin":
			roleID = role.AdminRoleID
		case "member":
			roleID = role.MemberRoleID
		default:
			roleID = role.ViewerRoleID
		}
	}

	return s.roleRepo.GetByID(ctx, roleID)
}

// GetUserProjectRole returns a user's project-specific role (if any)
func (s *service) GetUserProjectRole(ctx context.Context, projectID, userID uuid.UUID) (*role.Role, error) {
	ctx, span := s.startServiceSpan(ctx, "GetUserProjectRole")
	span.SetAttributes(
		attribute.String("project.id", projectID.String()),
		attribute.String("user.id", userID.String()),
	)
	defer span.End()

	member, err := s.projectMemberRepo.GetByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	if member.RoleID == nil {
		return nil, nil // No project-specific role
	}

	return s.roleRepo.GetByID(ctx, *member.RoleID)
}

// GetOrgMembers returns all members of an organization
func (s *service) GetOrgMembers(ctx context.Context, orgID uuid.UUID) ([]*organization_member.OrganizationMember, error) {
	ctx, span := s.startServiceSpan(ctx, "GetOrgMembers")
	span.SetAttributes(attribute.String("org.id", orgID.String()))
	defer span.End()

	return s.orgMemberRepo.GetByOrgID(ctx, orgID)
}

// GetProjectMembers returns all members of a project
func (s *service) GetProjectMembers(ctx context.Context, projectID uuid.UUID) ([]*project_member.ProjectMember, error) {
	ctx, span := s.startServiceSpan(ctx, "GetProjectMembers")
	span.SetAttributes(attribute.String("project.id", projectID.String()))
	defer span.End()

	return s.projectMemberRepo.GetByProjectID(ctx, projectID)
}

// RemoveOrgMember removes a member from an organization
func (s *service) RemoveOrgMember(ctx context.Context, orgID, userID, actorID uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "RemoveOrgMember")
	span.SetAttributes(
		attribute.String("org.id", orgID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("actor.id", actorID.String()),
	)
	defer span.End()

	// Get member to check role
	member, err := s.orgMemberRepo.GetByOrgAndUser(ctx, orgID, userID)
	if err != nil {
		return err
	}

	// Check if trying to remove an owner
	isOwner := (member.RoleID != nil && *member.RoleID == role.OwnerRoleID) || member.Role == "owner"
	if isOwner {
		// Count owners
		ownerCount, err := s.countOrgOwners(ctx, orgID)
		if err != nil {
			return err
		}
		if ownerCount <= 1 {
			return ErrLastOwner
		}
	}

	return s.orgMemberRepo.Delete(ctx, orgID, userID)
}

// RemoveProjectMember removes a member from a project
func (s *service) RemoveProjectMember(ctx context.Context, projectID, userID uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "RemoveProjectMember")
	span.SetAttributes(
		attribute.String("project.id", projectID.String()),
		attribute.String("user.id", userID.String()),
	)
	defer span.End()

	return s.projectMemberRepo.Delete(ctx, projectID, userID)
}

// GetOrgMemberUser returns the user for an organization member
func (s *service) GetOrgMemberUser(ctx context.Context, memberID uuid.UUID) (*user.User, error) {
	ctx, span := s.startServiceSpan(ctx, "GetOrgMemberUser")
	span.SetAttributes(attribute.String("member.id", memberID.String()))
	defer span.End()

	member, err := s.orgMemberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, member.UserID)
}

// GetOrgMemberRole returns the role for an organization member
func (s *service) GetOrgMemberRole(ctx context.Context, memberID uuid.UUID) (*role.Role, error) {
	ctx, span := s.startServiceSpan(ctx, "GetOrgMemberRole")
	span.SetAttributes(attribute.String("member.id", memberID.String()))
	defer span.End()

	member, err := s.orgMemberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	// Get role ID from member
	var roleID uuid.UUID
	if member.RoleID != nil {
		roleID = *member.RoleID
	} else {
		// Legacy fallback
		switch member.Role {
		case "owner":
			roleID = role.OwnerRoleID
		case "admin":
			roleID = role.AdminRoleID
		case "member":
			roleID = role.MemberRoleID
		default:
			roleID = role.ViewerRoleID
		}
	}

	return s.roleRepo.GetByID(ctx, roleID)
}

// GetProjectMemberUser returns the user for a project member
func (s *service) GetProjectMemberUser(ctx context.Context, memberID uuid.UUID) (*user.User, error) {
	ctx, span := s.startServiceSpan(ctx, "GetProjectMemberUser")
	span.SetAttributes(attribute.String("member.id", memberID.String()))
	defer span.End()

	member, err := s.projectMemberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, member.UserID)
}

// GetProjectMemberRole returns the role for a project member
func (s *service) GetProjectMemberRole(ctx context.Context, memberID uuid.UUID) (*role.Role, error) {
	ctx, span := s.startServiceSpan(ctx, "GetProjectMemberRole")
	span.SetAttributes(attribute.String("member.id", memberID.String()))
	defer span.End()

	member, err := s.projectMemberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	if member.RoleID == nil {
		return nil, nil
	}

	return s.roleRepo.GetByID(ctx, *member.RoleID)
}

// GetProjectMemberProject returns the project for a project member
func (s *service) GetProjectMemberProject(ctx context.Context, memberID uuid.UUID) (*project.Project, error) {
	ctx, span := s.startServiceSpan(ctx, "GetProjectMemberProject")
	span.SetAttributes(attribute.String("member.id", memberID.String()))
	defer span.End()

	member, err := s.projectMemberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	return s.projectRepo.GetByID(ctx, member.ProjectID)
}
