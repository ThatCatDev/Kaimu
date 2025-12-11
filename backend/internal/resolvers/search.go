package resolvers

import (
	"context"
	"errors"

	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	"github.com/thatcatdev/kaimu/backend/internal/services/search"
)

// Search performs a full-text search across multiple entity types
func Search(ctx context.Context, searchService search.Service, query string, scope *model.SearchScope, limit *int) (*model.SearchResults, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, errors.New("not authenticated")
	}

	// Convert GraphQL scope to service scope
	var serviceScope *search.SearchScope
	if scope != nil {
		serviceScope = &search.SearchScope{}
		if scope.OrganizationID != nil {
			serviceScope.OrganizationID = *scope.OrganizationID
		}
		if scope.ProjectID != nil {
			serviceScope.ProjectID = *scope.ProjectID
		}
	}

	// Get limit with default
	searchLimit := 20
	if limit != nil {
		searchLimit = *limit
	}

	// Perform search
	results, err := searchService.Search(ctx, *userID, query, serviceScope, searchLimit)
	if err != nil {
		return nil, err
	}

	// Convert service results to GraphQL model
	modelResults := make([]*model.SearchResult, len(results.Results))
	for i, r := range results.Results {
		modelResults[i] = &model.SearchResult{
			Type:             convertEntityType(r.Type),
			ID:               r.ID,
			Title:            r.Title,
			Description:      stringPtr(r.Description),
			Highlight:        r.Highlight,
			OrganizationID:   r.OrganizationID,
			OrganizationName: r.OrganizationName,
			ProjectID:        stringPtr(r.ProjectID),
			ProjectName:      stringPtr(r.ProjectName),
			BoardID:          stringPtr(r.BoardID),
			BoardName:        stringPtr(r.BoardName),
			URL:              r.URL,
			Score:            r.Score,
		}
	}

	return &model.SearchResults{
		Results:    modelResults,
		TotalCount: results.TotalCount,
		Query:      results.Query,
	}, nil
}

func convertEntityType(t search.EntityType) model.SearchEntityType {
	switch t {
	case search.EntityTypeCard:
		return model.SearchEntityTypeCard
	case search.EntityTypeProject:
		return model.SearchEntityTypeProject
	case search.EntityTypeBoard:
		return model.SearchEntityTypeBoard
	case search.EntityTypeOrganization:
		return model.SearchEntityTypeOrganization
	case search.EntityTypeUser:
		return model.SearchEntityTypeUser
	default:
		return model.SearchEntityTypeCard
	}
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
