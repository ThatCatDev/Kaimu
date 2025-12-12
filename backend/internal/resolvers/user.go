package resolvers

import (
	"context"
	"errors"

	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	organizationService "github.com/thatcatdev/kaimu/backend/internal/services/organization"
	userService "github.com/thatcatdev/kaimu/backend/internal/services/user"
)

var (
	ErrNotAuthenticated = errors.New("not authenticated")
)

func UpdateMe(ctx context.Context, userSvc userService.Service, orgSvc organizationService.Service, searchIndexer *SearchIndexer, input model.UpdateMeInput) (*model.User, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrNotAuthenticated
	}

	u, err := userSvc.Update(ctx, *userID, input.DisplayName, input.Email)
	if err != nil {
		if errors.Is(err, userService.ErrUserNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Re-index user in search after update
	if searchIndexer != nil {
		orgs, err := orgSvc.GetUserOrganizations(ctx, *userID)
		if err == nil {
			orgIDs := make([]string, len(orgs))
			for i, org := range orgs {
				orgIDs[i] = org.ID.String()
			}
			searchIndexer.IndexUserAsync(ctx, *userID, orgIDs)
		}
	}

	return UserToModel(u), nil
}
