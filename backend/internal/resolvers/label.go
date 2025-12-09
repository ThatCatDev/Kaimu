package resolvers

import (
	"context"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/graph/model"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/label"
	labelService "github.com/thatcatdev/pulse-backend/internal/services/label"
	orgService "github.com/thatcatdev/pulse-backend/internal/services/organization"
	projectService "github.com/thatcatdev/pulse-backend/internal/services/project"
)

// Labels returns all labels for a project
func Labels(ctx context.Context, orgSvc orgService.Service, labelSvc labelService.Service, projSvc projectService.Service, projectID string) ([]*model.Label, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	projID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, err
	}

	// Check membership
	proj, err := projSvc.GetProject(ctx, projID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	labels, err := labelSvc.GetLabelsByProjectID(ctx, projID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Label, len(labels))
	for i, l := range labels {
		result[i] = labelToModel(l)
	}
	return result, nil
}

// CreateLabel creates a new label
func CreateLabel(ctx context.Context, orgSvc orgService.Service, labelSvc labelService.Service, projSvc projectService.Service, input model.CreateLabelInput) (*model.Label, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	projID, err := uuid.Parse(input.ProjectID)
	if err != nil {
		return nil, err
	}

	// Check membership
	proj, err := projSvc.GetProject(ctx, projID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
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

	l, err := labelSvc.CreateLabel(ctx, projID, input.Name, input.Color, description)
	if err != nil {
		return nil, err
	}

	return labelToModel(l), nil
}

// UpdateLabel updates a label
func UpdateLabel(ctx context.Context, orgSvc orgService.Service, labelSvc labelService.Service, input model.UpdateLabelInput) (*model.Label, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	labelID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	l, err := labelSvc.GetLabel(ctx, labelID)
	if err != nil {
		return nil, err
	}

	// Check membership
	proj, err := labelSvc.GetProject(ctx, labelID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	if input.Name != nil {
		l.Name = *input.Name
	}
	if input.Color != nil {
		l.Color = *input.Color
	}
	if input.Description != nil {
		l.Description = *input.Description
	}

	updated, err := labelSvc.UpdateLabel(ctx, l)
	if err != nil {
		return nil, err
	}

	return labelToModel(updated), nil
}

// DeleteLabel deletes a label
func DeleteLabel(ctx context.Context, orgSvc orgService.Service, labelSvc labelService.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	labelID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Check membership
	proj, err := labelSvc.GetProject(ctx, labelID)
	if err != nil {
		return false, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return false, err
	}
	if !isMember {
		return false, ErrUnauthorized
	}

	if err := labelSvc.DeleteLabel(ctx, labelID); err != nil {
		return false, err
	}

	return true, nil
}

// LabelProject resolves the project field of a Label
func LabelProject(ctx context.Context, labelSvc labelService.Service, orgSvc orgService.Service, l *model.Label) (*model.Project, error) {
	labelID, err := uuid.Parse(l.ID)
	if err != nil {
		return nil, err
	}

	proj, err := labelSvc.GetProject(ctx, labelID)
	if err != nil {
		return nil, err
	}

	// Get the organization for the project
	org, err := orgSvc.GetOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	return projectToModelWithOrg(proj, organizationToModel(org)), nil
}

func labelToModel(l *label.Label) *model.Label {
	var description *string
	if l.Description != "" {
		description = &l.Description
	}
	return &model.Label{
		ID:          l.ID.String(),
		Name:        l.Name,
		Color:       l.Color,
		Description: description,
		CreatedAt:   l.CreatedAt,
	}
}
