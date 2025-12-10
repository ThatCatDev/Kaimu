package resolvers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member"
	boardService "github.com/thatcatdev/kaimu/backend/internal/services/board"
	orgService "github.com/thatcatdev/kaimu/backend/internal/services/organization"
	projectService "github.com/thatcatdev/kaimu/backend/internal/services/project"
)

var ErrUnauthorized = errors.New("unauthorized")

// CreateOrganization creates a new organization
func CreateOrganization(ctx context.Context, svc orgService.Service, input model.CreateOrganizationInput) (*model.Organization, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	description := ""
	if input.Description != nil {
		description = *input.Description
	}

	org, err := svc.CreateOrganization(ctx, *userID, input.Name, description)
	if err != nil {
		return nil, err
	}

	// Get owner for the response
	owner, err := svc.GetOwner(ctx, org.ID)
	if err != nil {
		return nil, err
	}

	return organizationToModelWithRelations(org, UserToModel(owner), nil, nil), nil
}

// Organizations returns all organizations for the current user
func Organizations(ctx context.Context, svc orgService.Service, projectSvc projectService.Service, boardSvc boardService.Service) ([]*model.Organization, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgs, err := svc.GetUserOrganizations(ctx, *userID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Organization, len(orgs))
	for i, org := range orgs {
		// Fetch owner
		owner, err := svc.GetOwner(ctx, org.ID)
		if err != nil {
			return nil, err
		}

		// Fetch projects for each organization
		projects, err := projectSvc.GetOrgProjects(ctx, org.ID)
		if err != nil {
			return nil, err
		}

		projectModels := make([]*model.Project, len(projects))
		for j, proj := range projects {
			// Fetch boards for each project
			boards, err := boardSvc.GetBoardsByProjectID(ctx, proj.ID)
			if err != nil {
				return nil, err
			}
			projectModels[j] = projectToModelWithBoards(proj, boards)
		}

		result[i] = organizationToModelWithRelations(org, UserToModel(owner), nil, projectModels)
	}
	return result, nil
}

// Organization returns a specific organization by ID
func Organization(ctx context.Context, svc orgService.Service, projectSvc projectService.Service, id string) (*model.Organization, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	// Check if user is a member
	isMember, err := svc.IsMember(ctx, orgID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	org, err := svc.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Fetch owner
	owner, err := svc.GetOwner(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Fetch projects
	projects, err := projectSvc.GetOrgProjects(ctx, orgID)
	if err != nil {
		return nil, err
	}

	projectModels := make([]*model.Project, len(projects))
	for i, proj := range projects {
		projectModels[i] = projectToModel(proj)
	}

	return organizationToModelWithRelations(org, UserToModel(owner), nil, projectModels), nil
}

// OrganizationOwner resolves the owner field of an Organization
func OrganizationOwner(ctx context.Context, svc orgService.Service, org *model.Organization) (*model.User, error) {
	orgID, err := uuid.Parse(org.ID)
	if err != nil {
		return nil, err
	}

	owner, err := svc.GetOwner(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return UserToModel(owner), nil
}

// OrganizationMembers resolves the members field of an Organization
func OrganizationMembers(ctx context.Context, svc orgService.Service, org *model.Organization) ([]*model.OrganizationMember, error) {
	orgID, err := uuid.Parse(org.ID)
	if err != nil {
		return nil, err
	}

	members, err := svc.GetMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.OrganizationMember, len(members))
	for i, member := range members {
		result[i] = organizationMemberToModel(member)
	}
	return result, nil
}

// OrganizationProjects resolves the projects field of an Organization
func OrganizationProjects(ctx context.Context, projectSvc projectService.Service, org *model.Organization) ([]*model.Project, error) {
	orgID, err := uuid.Parse(org.ID)
	if err != nil {
		return nil, err
	}

	projects, err := projectSvc.GetOrgProjects(ctx, orgID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Project, len(projects))
	for i, proj := range projects {
		result[i] = projectToModel(proj)
	}
	return result, nil
}

// UpdateOrganization updates an organization
func UpdateOrganization(ctx context.Context, svc orgService.Service, input model.UpdateOrganizationInput) (*model.Organization, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	// Check if user is a member
	isMember, err := svc.IsMember(ctx, orgID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	// Get current org
	org, err := svc.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if input.Name != nil {
		org.Name = *input.Name
	}
	if input.Description != nil {
		org.Description = *input.Description
	}

	updated, err := svc.UpdateOrganization(ctx, org)
	if err != nil {
		return nil, err
	}

	// Get owner for the response
	owner, err := svc.GetOwner(ctx, updated.ID)
	if err != nil {
		return nil, err
	}

	return organizationToModelWithRelations(updated, UserToModel(owner), nil, nil), nil
}

// DeleteOrganization deletes an organization by ID
func DeleteOrganization(ctx context.Context, svc orgService.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	orgID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Check if user is a member (and ideally owner, but for now just member)
	isMember, err := svc.IsMember(ctx, orgID, *userID)
	if err != nil {
		return false, err
	}
	if !isMember {
		return false, ErrUnauthorized
	}

	err = svc.DeleteOrganization(ctx, orgID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// OrganizationMemberUser resolves the user field of an OrganizationMember
// Note: The member model needs a UserID field to make this work properly.
// For now, we'll need to store the user ID in the model temporarily.

func organizationToModel(org *organization.Organization) *model.Organization {
	var description *string
	if org.Description != "" {
		description = &org.Description
	}
	return &model.Organization{
		ID:          org.ID.String(),
		Name:        org.Name,
		Slug:        org.Slug,
		Description: description,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
		// Note: Owner, Members, Projects are nil - they need to be populated separately
		Owner:    nil,
		Members:  []*model.OrganizationMember{},
		Projects: []*model.Project{},
	}
}

func organizationToModelWithRelations(org *organization.Organization, owner *model.User, members []*model.OrganizationMember, projects []*model.Project) *model.Organization {
	var description *string
	if org.Description != "" {
		description = &org.Description
	}
	if members == nil {
		members = []*model.OrganizationMember{}
	}
	if projects == nil {
		projects = []*model.Project{}
	}
	return &model.Organization{
		ID:          org.ID.String(),
		Name:        org.Name,
		Slug:        org.Slug,
		Description: description,
		Owner:       owner,
		Members:     members,
		Projects:    projects,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}
}

func organizationMemberToModel(member *organization_member.OrganizationMember) *model.OrganizationMember {
	return &model.OrganizationMember{
		ID:         member.ID.String(),
		LegacyRole: member.Role,
		CreatedAt:  member.CreatedAt,
		User:       nil, // Needs to be populated separately via field resolver
		Role:       nil, // Needs to be populated separately via field resolver
	}
}

func organizationMemberToModelWithUser(member *organization_member.OrganizationMember, user *model.User) *model.OrganizationMember {
	return &model.OrganizationMember{
		ID:         member.ID.String(),
		User:       user,
		LegacyRole: member.Role,
		CreatedAt:  member.CreatedAt,
		Role:       nil, // Needs to be populated separately via field resolver
	}
}
