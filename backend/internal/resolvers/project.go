package resolvers

import (
	"context"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/graph/model"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	boardService "github.com/thatcatdev/pulse-backend/internal/services/board"
	orgService "github.com/thatcatdev/pulse-backend/internal/services/organization"
	projectService "github.com/thatcatdev/pulse-backend/internal/services/project"
)

// CreateProject creates a new project
func CreateProject(ctx context.Context, orgSvc orgService.Service, projSvc projectService.Service, boardSvc boardService.Service, input model.CreateProjectInput) (*model.Project, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		return nil, err
	}

	// Check if user is a member of the organization
	isMember, err := orgSvc.IsMember(ctx, orgID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	description := ""
	if input.Description != nil {
		description = *input.Description
	}

	proj, err := projSvc.CreateProject(ctx, orgID, input.Name, input.Key, description)
	if err != nil {
		return nil, err
	}

	// Create default board for the project
	_, err = boardSvc.CreateDefaultBoard(ctx, proj.ID, userID)
	if err != nil {
		// Log error but don't fail project creation
		// The board can be created later
	}

	// Fetch the organization for the project
	org, err := orgSvc.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return projectToModelWithOrg(proj, organizationToModel(org)), nil
}

// Project returns a specific project by ID
func Project(ctx context.Context, orgSvc orgService.Service, projSvc projectService.Service, id string) (*model.Project, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	projID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	proj, err := projSvc.GetProject(ctx, projID)
	if err != nil {
		return nil, err
	}

	// Check if user is a member of the organization
	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	// Fetch the organization for the project
	org, err := projSvc.GetOrganization(ctx, projID)
	if err != nil {
		return nil, err
	}

	return projectToModelWithOrg(proj, organizationToModel(org)), nil
}

// ProjectOrganization resolves the organization field of a Project
func ProjectOrganization(ctx context.Context, projSvc projectService.Service, proj *model.Project) (*model.Organization, error) {
	projID, err := uuid.Parse(proj.ID)
	if err != nil {
		return nil, err
	}

	org, err := projSvc.GetOrganization(ctx, projID)
	if err != nil {
		return nil, err
	}

	return organizationToModel(org), nil
}

func projectToModel(proj *project.Project) *model.Project {
	var description *string
	if proj.Description != "" {
		description = &proj.Description
	}
	return &model.Project{
		ID:           proj.ID.String(),
		Name:         proj.Name,
		Key:          proj.Key,
		Description:  description,
		Organization: nil, // Needs to be populated separately
		CreatedAt:    proj.CreatedAt,
		UpdatedAt:    proj.UpdatedAt,
	}
}

func projectToModelWithOrg(proj *project.Project, org *model.Organization) *model.Project {
	var description *string
	if proj.Description != "" {
		description = &proj.Description
	}
	return &model.Project{
		ID:           proj.ID.String(),
		Organization: org,
		Name:         proj.Name,
		Key:          proj.Key,
		Description:  description,
		CreatedAt:    proj.CreatedAt,
		UpdatedAt:    proj.UpdatedAt,
	}
}
