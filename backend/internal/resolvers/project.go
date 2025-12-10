package resolvers

import (
	"context"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/graph/model"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/board"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	boardService "github.com/thatcatdev/pulse-backend/internal/services/board"
	orgService "github.com/thatcatdev/pulse-backend/internal/services/organization"
	projectService "github.com/thatcatdev/pulse-backend/internal/services/project"
	rbacService "github.com/thatcatdev/pulse-backend/internal/services/rbac"
)

// CreateProject creates a new project
func CreateProject(ctx context.Context, rbacSvc rbacService.Service, orgSvc orgService.Service, projSvc projectService.Service, boardSvc boardService.Service, input model.CreateProjectInput) (*model.Project, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(input.OrganizationID)
	if err != nil {
		return nil, err
	}

	// Check if user has permission to create projects in this organization
	hasPermission, err := rbacSvc.HasOrgPermission(ctx, *userID, orgID, "project:create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
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
func Project(ctx context.Context, rbacSvc rbacService.Service, projSvc projectService.Service, id string) (*model.Project, error) {
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

	// Check if user has permission to view the project
	hasPermission, err := rbacSvc.HasProjectPermission(ctx, *userID, projID, "project:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
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

// UpdateProject updates a project
func UpdateProject(ctx context.Context, rbacSvc rbacService.Service, projSvc projectService.Service, input model.UpdateProjectInput) (*model.Project, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	projID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	// Get current project
	proj, err := projSvc.GetProject(ctx, projID)
	if err != nil {
		return nil, err
	}

	// Check if user has permission to manage the project
	hasPermission, err := rbacSvc.HasProjectPermission(ctx, *userID, projID, "project:manage")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	// Apply updates
	if input.Name != nil {
		proj.Name = *input.Name
	}
	if input.Key != nil {
		proj.Key = *input.Key
	}
	if input.Description != nil {
		proj.Description = *input.Description
	}

	updated, err := projSvc.UpdateProject(ctx, proj)
	if err != nil {
		return nil, err
	}

	// Fetch the organization for the project
	org, err := projSvc.GetOrganization(ctx, updated.ID)
	if err != nil {
		return nil, err
	}

	return projectToModelWithOrg(updated, organizationToModel(org)), nil
}

// DeleteProject deletes a project by ID
func DeleteProject(ctx context.Context, rbacSvc rbacService.Service, projSvc projectService.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	projID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Check if user has permission to delete the project
	hasPermission, err := rbacSvc.HasProjectPermission(ctx, *userID, projID, "project:delete")
	if err != nil {
		return false, err
	}
	if !hasPermission {
		return false, ErrUnauthorized
	}

	err = projSvc.DeleteProject(ctx, projID)
	if err != nil {
		return false, err
	}

	return true, nil
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

func projectToModelWithBoards(proj *project.Project, boards []*board.Board) *model.Project {
	var description *string
	if proj.Description != "" {
		description = &proj.Description
	}

	boardModels := make([]*model.Board, len(boards))
	for i, b := range boards {
		var boardDesc *string
		if b.Description != "" {
			boardDesc = &b.Description
		}
		boardModels[i] = &model.Board{
			ID:          b.ID.String(),
			Name:        b.Name,
			Description: boardDesc,
			IsDefault:   b.IsDefault,
			CreatedAt:   b.CreatedAt,
			UpdatedAt:   b.UpdatedAt,
		}
	}

	return &model.Project{
		ID:          proj.ID.String(),
		Name:        proj.Name,
		Key:         proj.Key,
		Description: description,
		Boards:      boardModels,
		CreatedAt:   proj.CreatedAt,
		UpdatedAt:   proj.UpdatedAt,
	}
}
