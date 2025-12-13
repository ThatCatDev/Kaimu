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
	userAgent := middleware.GetUserAgentFromContext(ctx)
	ipAddress := middleware.GetIPAddressFromContext(ctx)

	u, tokenPair, err := authService.Register(ctx, input.Username, input.Email, input.Password, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, errors.New("username already taken")
		}
		return nil, err
	}

	// Set auth cookies
	w := middleware.GetResponseWriter(ctx)
	if w != nil {
		middleware.SetAuthCookies(w, tokenPair.AccessToken, tokenPair.RefreshToken, isSecure)
	}

	return &model.AuthPayload{
		User: UserToModel(u),
	}, nil
}

func Login(ctx context.Context, authService auth.Service, input model.LoginInput, isSecure bool) (*model.AuthPayload, error) {
	userAgent := middleware.GetUserAgentFromContext(ctx)
	ipAddress := middleware.GetIPAddressFromContext(ctx)

	u, tokenPair, err := authService.Login(ctx, input.Username, input.Password, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// Set auth cookies
	w := middleware.GetResponseWriter(ctx)
	if w != nil {
		middleware.SetAuthCookies(w, tokenPair.AccessToken, tokenPair.RefreshToken, isSecure)
	}

	return &model.AuthPayload{
		User: UserToModel(u),
	}, nil
}

func Logout(ctx context.Context, authService auth.Service) (bool, error) {
	// Revoke the refresh token if present
	refreshToken := middleware.GetRefreshTokenFromContext(ctx)
	if refreshToken != "" {
		_ = authService.RevokeRefreshToken(ctx, refreshToken)
	}

	// Clear cookies
	w := middleware.GetResponseWriter(ctx)
	if w != nil {
		middleware.ClearAuthCookies(w)
	}
	return true, nil
}

func RefreshToken(ctx context.Context, authService auth.Service, isSecure bool) (*model.RefreshTokenPayload, error) {
	refreshToken := middleware.GetRefreshTokenFromContext(ctx)
	if refreshToken == "" {
		return nil, errors.New("no refresh token provided")
	}

	userAgent := middleware.GetUserAgentFromContext(ctx)
	ipAddress := middleware.GetIPAddressFromContext(ctx)

	tokenPair, err := authService.RefreshTokens(ctx, refreshToken, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) || errors.Is(err, auth.ErrRefreshTokenRevoked) {
			// Clear cookies on invalid/revoked refresh token
			w := middleware.GetResponseWriter(ctx)
			if w != nil {
				middleware.ClearAuthCookies(w)
			}
			return nil, errors.New("session expired, please login again")
		}
		return nil, err
	}

	// Set new auth cookies
	w := middleware.GetResponseWriter(ctx)
	if w != nil {
		middleware.SetAuthCookies(w, tokenPair.AccessToken, tokenPair.RefreshToken, isSecure)
	}

	return &model.RefreshTokenPayload{
		Success:   true,
		ExpiresIn: int(tokenPair.ExpiresIn),
	}, nil
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
