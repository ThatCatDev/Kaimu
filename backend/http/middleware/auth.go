package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/internal/services/auth"
)

type contextKey string

const (
	UserIDKey     contextKey = "userID"
	ResponseKey   contextKey = "httpResponseWriter"
)

func AuthMiddleware(authService auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add response writer to context for cookie setting
			ctx := context.WithValue(r.Context(), ResponseKey, w)

			// Try to get token from cookie
			cookie, err := r.Cookie("pulse_token")
			if err == nil && cookie.Value != "" {
				claims, err := authService.ValidateToken(cookie.Value)
				if err == nil {
					ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) *uuid.UUID {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	if !ok {
		return nil
	}
	return &userID
}

func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	w, _ := ctx.Value(ResponseKey).(http.ResponseWriter)
	return w
}

func SetAuthCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "pulse_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 days
	})
}

func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "pulse_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}
