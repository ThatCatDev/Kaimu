package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/thatcatdev/pulse-backend/internal/services/auth"
	"github.com/thatcatdev/pulse-backend/internal/services/auth/mocks"
	"go.uber.org/mock/gomock"
)

func TestAuthMiddleware_WithValidCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockService(ctrl)
	userID := uuid.New()
	claims := &auth.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	mockAuth.EXPECT().ValidateToken("valid-token").Return(claims, nil)

	middleware := AuthMiddleware(mockAuth)

	var capturedUserID *uuid.UUID
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "pulse_token", Value: "valid-token"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, capturedUserID)
	assert.Equal(t, userID, *capturedUserID)
}

func TestAuthMiddleware_WithInvalidCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockService(ctrl)

	mockAuth.EXPECT().ValidateToken("invalid-token").Return(nil, auth.ErrInvalidToken)

	middleware := AuthMiddleware(mockAuth)

	var capturedUserID *uuid.UUID
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "pulse_token", Value: "invalid-token"})
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Request should still succeed, just without user context
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, capturedUserID)
}

func TestAuthMiddleware_WithNoCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockService(ctrl)

	middleware := AuthMiddleware(mockAuth)

	var capturedUserID *uuid.UUID
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Nil(t, capturedUserID)
}

func TestGetUserIDFromContext_NoUser(t *testing.T) {
	ctx := context.Background()
	userID := GetUserIDFromContext(ctx)
	assert.Nil(t, userID)
}

func TestGetUserIDFromContext_WithUser(t *testing.T) {
	expectedID := uuid.New()
	ctx := context.WithValue(context.Background(), UserIDKey, expectedID)

	userID := GetUserIDFromContext(ctx)

	assert.NotNil(t, userID)
	assert.Equal(t, expectedID, *userID)
}

func TestSetAuthCookie(t *testing.T) {
	rr := httptest.NewRecorder()

	SetAuthCookie(rr, "test-token", false)

	cookies := rr.Result().Cookies()
	assert.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "pulse_token", cookie.Name)
	assert.Equal(t, "test-token", cookie.Value)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, "/", cookie.Path)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
}

func TestSetAuthCookie_Secure(t *testing.T) {
	rr := httptest.NewRecorder()

	SetAuthCookie(rr, "test-token", true)

	cookies := rr.Result().Cookies()
	cookie := cookies[0]
	assert.True(t, cookie.Secure)
}

func TestClearAuthCookie(t *testing.T) {
	rr := httptest.NewRecorder()

	ClearAuthCookie(rr)

	cookies := rr.Result().Cookies()
	assert.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "pulse_token", cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge)
}

func TestGetResponseWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := context.WithValue(context.Background(), ResponseKey, http.ResponseWriter(rr))

	w := GetResponseWriter(ctx)

	assert.NotNil(t, w)
	assert.Equal(t, rr, w)
}

func TestGetResponseWriter_NoWriter(t *testing.T) {
	ctx := context.Background()

	w := GetResponseWriter(ctx)

	assert.Nil(t, w)
}
