package resolvers

import (
	"context"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/graph/model"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/tag"
	orgService "github.com/thatcatdev/pulse-backend/internal/services/organization"
	projectService "github.com/thatcatdev/pulse-backend/internal/services/project"
	tagService "github.com/thatcatdev/pulse-backend/internal/services/tag"
)

// Tags returns all tags for a project
func Tags(ctx context.Context, orgSvc orgService.Service, tagSvc tagService.Service, projSvc projectService.Service, projectID string) ([]*model.Tag, error) {
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

	tags, err := tagSvc.GetTagsByProjectID(ctx, projID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Tag, len(tags))
	for i, t := range tags {
		result[i] = tagToModel(t)
	}
	return result, nil
}

// CreateTag creates a new tag
func CreateTag(ctx context.Context, orgSvc orgService.Service, tagSvc tagService.Service, projSvc projectService.Service, input model.CreateTagInput) (*model.Tag, error) {
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

	t, err := tagSvc.CreateTag(ctx, projID, input.Name, input.Color, description)
	if err != nil {
		return nil, err
	}

	return tagToModel(t), nil
}

// UpdateTag updates a tag
func UpdateTag(ctx context.Context, orgSvc orgService.Service, tagSvc tagService.Service, input model.UpdateTagInput) (*model.Tag, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	tagID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	t, err := tagSvc.GetTag(ctx, tagID)
	if err != nil {
		return nil, err
	}

	// Check membership
	proj, err := tagSvc.GetProject(ctx, tagID)
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
		t.Name = *input.Name
	}
	if input.Color != nil {
		t.Color = *input.Color
	}
	if input.Description != nil {
		t.Description = *input.Description
	}

	updated, err := tagSvc.UpdateTag(ctx, t)
	if err != nil {
		return nil, err
	}

	return tagToModel(updated), nil
}

// DeleteTag deletes a tag
func DeleteTag(ctx context.Context, orgSvc orgService.Service, tagSvc tagService.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	tagID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Check membership
	proj, err := tagSvc.GetProject(ctx, tagID)
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

	if err := tagSvc.DeleteTag(ctx, tagID); err != nil {
		return false, err
	}

	return true, nil
}

// TagProject resolves the project field of a Tag
func TagProject(ctx context.Context, tagSvc tagService.Service, orgSvc orgService.Service, t *model.Tag) (*model.Project, error) {
	tagID, err := uuid.Parse(t.ID)
	if err != nil {
		return nil, err
	}

	proj, err := tagSvc.GetProject(ctx, tagID)
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

func tagToModel(t *tag.Tag) *model.Tag {
	var description *string
	if t.Description != "" {
		description = &t.Description
	}
	return &model.Tag{
		ID:          t.ID.String(),
		Name:        t.Name,
		Color:       t.Color,
		Description: description,
		CreatedAt:   t.CreatedAt,
	}
}
