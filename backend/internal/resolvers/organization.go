package resolvers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/graph/model"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/organization"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/organization_member"
	orgService "github.com/thatcatdev/pulse-backend/internal/services/organization"
	projectService "github.com/thatcatdev/pulse-backend/internal/services/project"
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

	return organizationToModelWithRelations(org, userToModel(owner), nil, nil), nil
}

// Organizations returns all organizations for the current user
func Organizations(ctx context.Context, svc orgService.Service) ([]*model.Organization, error) {
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
		result[i] = organizationToModel(org)
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

	return organizationToModelWithRelations(org, userToModel(owner), nil, projectModels), nil
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

	return userToModel(owner), nil
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
		ID:        member.ID.String(),
		Role:      member.Role,
		CreatedAt: member.CreatedAt,
		User:      nil, // Needs to be populated separately
	}
}

func organizationMemberToModelWithUser(member *organization_member.OrganizationMember, user *model.User) *model.OrganizationMember {
	return &model.OrganizationMember{
		ID:        member.ID.String(),
		User:      user,
		Role:      member.Role,
		CreatedAt: member.CreatedAt,
	}
}
