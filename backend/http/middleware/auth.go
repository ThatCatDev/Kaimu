package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/internal/services/auth"
)

type contextKey string

const (
	UserIDKey       contextKey = "userID"
	ResponseKey     contextKey = "httpResponseWriter"
	RefreshTokenKey contextKey = "refreshToken"
	UserAgentKey    contextKey = "userAgent"
	IPAddressKey    contextKey = "ipAddress"

	// Cookie names
	AccessTokenCookie  = "kaimu_access_token"
	RefreshTokenCookie = "kaimu_refresh_token"

	// Cookie durations
	AccessTokenMaxAge  = 300       // 5 minutes (matches JWT expiry)
	RefreshTokenMaxAge = 604800    // 7 days
)

func AuthMiddleware(authService auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add response writer to context for cookie setting
			ctx := context.WithValue(r.Context(), ResponseKey, w)

			// Add request metadata to context
			ctx = context.WithValue(ctx, UserAgentKey, r.Header.Get("User-Agent"))
			ctx = context.WithValue(ctx, IPAddressKey, GetClientIP(r))

			// Try to get access token from cookie
			cookie, err := r.Cookie(AccessTokenCookie)
			if err == nil && cookie.Value != "" {
				claims, err := authService.ValidateToken(cookie.Value)
				if err == nil {
					ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
				}
			}

			// Also store refresh token in context if present (for refresh endpoint)
			refreshCookie, err := r.Cookie(RefreshTokenCookie)
			if err == nil && refreshCookie.Value != "" {
				ctx = context.WithValue(ctx, RefreshTokenKey, refreshCookie.Value)
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

func GetRefreshTokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(RefreshTokenKey).(string)
	return token
}

func GetUserAgentFromContext(ctx context.Context) string {
	ua, _ := ctx.Value(UserAgentKey).(string)
	return ua
}

func GetIPAddressFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(IPAddressKey).(string)
	return ip
}

func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	w, _ := ctx.Value(ResponseKey).(http.ResponseWriter)
	return w
}

// SetAuthCookies sets both access and refresh token cookies
func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, secure bool) {
	// Access token cookie (short-lived, matches JWT expiry)
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   AccessTokenMaxAge,
	})

	// Refresh token cookie (longer-lived)
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   RefreshTokenMaxAge,
	})
}

// SetAuthCookie sets the access token cookie (legacy support, use SetAuthCookies instead)
func SetAuthCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   AccessTokenMaxAge,
	})
}

// ClearAuthCookies clears both access and refresh token cookies
func ClearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// ClearAuthCookie clears the access token cookie (legacy support)
func ClearAuthCookie(w http.ResponseWriter) {
	ClearAuthCookies(w)
}

// GetClientIP extracts the client IP address from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxied requests)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, the first one is the client
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	// RemoteAddr is in the format "IP:port" or "[IPv6]:port"
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
