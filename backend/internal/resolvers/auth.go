package resolvers

import (
	"context"
	"errors"

	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/services/auth"
)

func Register(ctx context.Context, authService auth.Service, input model.RegisterInput, isSecure bool) (*model.AuthPayload, error) {
	u, token, err := authService.Register(ctx, input.Username, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, errors.New("username already taken")
		}
		return nil, err
	}

	// Set auth cookie
	w := middleware.GetResponseWriter(ctx)
	if w != nil {
		middleware.SetAuthCookie(w, token, isSecure)
	}

	return &model.AuthPayload{
		User: UserToModel(u),
	}, nil
}

func Login(ctx context.Context, authService auth.Service, input model.LoginInput, isSecure bool) (*model.AuthPayload, error) {
	u, token, err := authService.Login(ctx, input.Username, input.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// Set auth cookie
	w := middleware.GetResponseWriter(ctx)
	if w != nil {
		middleware.SetAuthCookie(w, token, isSecure)
	}

	return &model.AuthPayload{
		User: UserToModel(u),
	}, nil
}

func Logout(ctx context.Context) (bool, error) {
	w := middleware.GetResponseWriter(ctx)
	if w != nil {
		middleware.ClearAuthCookie(w)
	}
	return true, nil
}

func Me(ctx context.Context, authService auth.Service) (*model.User, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, nil
	}

	u, err := authService.GetUserByID(ctx, *userID)
	if err != nil {
		return nil, nil
	}

	return UserToModel(u), nil
}

func UserToModel(u *user.User) *model.User {
	return &model.User{
		ID:            u.ID.String(),
		Username:      u.Username,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		DisplayName:   u.DisplayName,
		AvatarURL:     u.AvatarURL,
		CreatedAt:     u.CreatedAt,
	}
}
