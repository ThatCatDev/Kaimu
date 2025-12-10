package resolvers

import (
	"context"
	"errors"

	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	userService "github.com/thatcatdev/kaimu/backend/internal/services/user"
)

var (
	ErrNotAuthenticated = errors.New("not authenticated")
)

func UpdateMe(ctx context.Context, userSvc userService.Service, input model.UpdateMeInput) (*model.User, error) {
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

	return UserToModel(u), nil
}
