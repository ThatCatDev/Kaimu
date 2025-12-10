package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"github.com/thatcatdev/pulse-backend/internal/services/oidc"
	"github.com/thatcatdev/pulse-backend/internal/services/oidc/mocks"
	"go.uber.org/mock/gomock"
)

func TestListProviders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	providers := []oidc.ProviderInfo{
		{Slug: "dex", Name: "Dex"},
		{Slug: "google", Name: "Google"},
	}

	mockService.EXPECT().GetProviders(gomock.Any()).Return(providers, nil)

	req := httptest.NewRequest("GET", "/auth/oidc/providers", nil)
	w := httptest.NewRecorder()

	handler.ListProviders(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var result []oidc.ProviderInfo
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "dex", result[0].Slug)
	assert.Equal(t, "google", result[1].Slug)
}

func TestListProviders_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().GetProviders(gomock.Any()).Return([]oidc.ProviderInfo{}, nil)

	req := httptest.NewRequest("GET", "/auth/oidc/providers", nil)
	w := httptest.NewRecorder()

	handler.ListProviders(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []oidc.ProviderInfo
	err := json.NewDecoder(w.Body).Decode(&result)
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListProviders_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().GetProviders(gomock.Any()).Return(nil, assert.AnError)

	req := httptest.NewRequest("GET", "/auth/oidc/providers", nil)
	w := httptest.NewRecorder()

	handler.ListProviders(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthorize_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	authResponse := &oidc.AuthorizationResponse{
		AuthURL:      "https://idp.example.com/auth?client_id=test&state=abc123",
		State:        "abc123",
		CodeVerifier: "code-verifier",
	}

	mockService.EXPECT().GetAuthorizationURL(gomock.Any(), "dex", "http://localhost:4321/dashboard").Return(authResponse, nil)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/authorize", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Authorize(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://idp.example.com/auth?client_id=test&state=abc123", w.Header().Get("Location"))
}

func TestAuthorize_WithCustomRedirectURI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	authResponse := &oidc.AuthorizationResponse{
		AuthURL:      "https://idp.example.com/auth?client_id=test&state=abc123",
		State:        "abc123",
		CodeVerifier: "code-verifier",
	}

	mockService.EXPECT().GetAuthorizationURL(gomock.Any(), "dex", "http://localhost:4321/custom").Return(authResponse, nil)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/authorize?redirect_uri=http://localhost:4321/custom", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Authorize(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
}

func TestAuthorize_NoProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	req := httptest.NewRequest("GET", "/auth/oidc//authorize", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": ""})
	w := httptest.NewRecorder()

	handler.Authorize(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthorize_ProviderNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().GetAuthorizationURL(gomock.Any(), "nonexistent", gomock.Any()).Return(nil, oidc.ErrProviderNotFound)

	req := httptest.NewRequest("GET", "/auth/oidc/nonexistent/authorize", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "nonexistent"})
	w := httptest.NewRecorder()

	handler.Authorize(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAuthorize_ProviderDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().GetAuthorizationURL(gomock.Any(), "disabled", gomock.Any()).Return(nil, oidc.ErrProviderDisabled)

	req := httptest.NewRequest("GET", "/auth/oidc/disabled/authorize", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "disabled"})
	w := httptest.NewRecorder()

	handler.Authorize(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCallback_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	userID := uuid.New()
	testUser := &user.User{
		ID:       userID,
		Username: "testuser",
	}

	callbackResult := &oidc.CallbackResult{
		User:             testUser,
		IsNewUser:        false,
		LinkedToExisting: false,
	}

	mockService.EXPECT().HandleCallback(gomock.Any(), "dex", "auth-code", "state-123").Return(callbackResult, "jwt-token", nil)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=auth-code&state=state-123", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "http://localhost:4321/dashboard", w.Header().Get("Location"))

	// Check cookie is set
	cookies := w.Result().Cookies()
	var foundCookie bool
	for _, c := range cookies {
		if c.Name == "pulse_token" {
			foundCookie = true
			assert.Equal(t, "jwt-token", c.Value)
		}
	}
	assert.True(t, foundCookie, "Expected pulse_token cookie to be set")
}

func TestCallback_NewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	userID := uuid.New()
	testUser := &user.User{
		ID:       userID,
		Username: "newuser",
	}

	callbackResult := &oidc.CallbackResult{
		User:             testUser,
		IsNewUser:        true,
		LinkedToExisting: false,
	}

	mockService.EXPECT().HandleCallback(gomock.Any(), "dex", "auth-code", "state-123").Return(callbackResult, "jwt-token", nil)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=auth-code&state=state-123", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "http://localhost:4321/dashboard?welcome=true", w.Header().Get("Location"))
}

func TestCallback_NoProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	req := httptest.NewRequest("GET", "/auth/oidc//callback?code=auth-code&state=state-123", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": ""})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
}

func TestCallback_MissingCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?state=state-123", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
}

func TestCallback_MissingState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=auth-code", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
}

func TestCallback_ProviderError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	// Simulate OIDC provider returning an error
	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?error=access_denied&error_description=User%20denied%20access", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
}

func TestCallback_InvalidState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().HandleCallback(gomock.Any(), "dex", "auth-code", "invalid-state").Return(nil, "", oidc.ErrInvalidState)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=auth-code&state=invalid-state", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
	assert.Contains(t, w.Header().Get("Location"), "expired")
}

func TestCallback_StateExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().HandleCallback(gomock.Any(), "dex", "auth-code", "expired-state").Return(nil, "", oidc.ErrStateExpired)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=auth-code&state=expired-state", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
	assert.Contains(t, w.Header().Get("Location"), "expired")
}

func TestCallback_TokenExchangeFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().HandleCallback(gomock.Any(), "dex", "bad-code", "state-123").Return(nil, "", oidc.ErrTokenExchangeFailed)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=bad-code&state=state-123", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
	assert.Contains(t, w.Header().Get("Location"), "authentication")
}

func TestCallback_InvalidIDToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().HandleCallback(gomock.Any(), "dex", "auth-code", "state-123").Return(nil, "", oidc.ErrInvalidIDToken)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=auth-code&state=state-123", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
	assert.Contains(t, w.Header().Get("Location"), "Invalid")
}

func TestCallback_NonceMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().HandleCallback(gomock.Any(), "dex", "auth-code", "state-123").Return(nil, "", oidc.ErrNonceMismatch)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=auth-code&state=state-123", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
	assert.Contains(t, w.Header().Get("Location"), "Invalid")
}

func TestCallback_GenericError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	handler := NewOIDCHandler(mockService, "http://localhost:4321", false)

	mockService.EXPECT().HandleCallback(gomock.Any(), "dex", "auth-code", "state-123").Return(nil, "", assert.AnError)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=auth-code&state=state-123", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
	assert.Contains(t, w.Header().Get("Location"), "failed")
}

func TestCallback_SecureCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	// isSecure = true for HTTPS
	handler := NewOIDCHandler(mockService, "http://localhost:4321", true)

	userID := uuid.New()
	testUser := &user.User{
		ID:       userID,
		Username: "testuser",
	}

	callbackResult := &oidc.CallbackResult{
		User:             testUser,
		IsNewUser:        false,
		LinkedToExisting: false,
	}

	mockService.EXPECT().HandleCallback(gomock.Any(), "dex", "auth-code", "state-123").Return(callbackResult, "jwt-token", nil)

	req := httptest.NewRequest("GET", "/auth/oidc/dex/callback?code=auth-code&state=state-123", nil)
	req = mux.SetURLVars(req, map[string]string{"provider": "dex"})
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	// Check cookie is set with Secure flag
	cookies := w.Result().Cookies()
	var foundCookie bool
	for _, c := range cookies {
		if c.Name == "pulse_token" {
			foundCookie = true
			assert.True(t, c.Secure, "Expected Secure flag to be set")
		}
	}
	assert.True(t, foundCookie, "Expected pulse_token cookie to be set")
}
