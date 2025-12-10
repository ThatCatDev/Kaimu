package oidc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/oidc_identity"
	oidc_identity_mocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/oidc_identity/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	user_mocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/user/mocks"
	"go.uber.org/mock/gomock"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// mockStateManager is a test implementation of StateManager
type mockStateManager struct {
	states map[string]*StateData
}

func newMockStateManager() *mockStateManager {
	return &mockStateManager{
		states: make(map[string]*StateData),
	}
}

func (m *mockStateManager) CreateState(providerSlug, redirectURI string) (string, *StateData, error) {
	state := "test-state-" + uuid.New().String()[:8]
	data := &StateData{
		ProviderSlug: providerSlug,
		CodeVerifier: "test-code-verifier",
		Nonce:        "test-nonce",
		RedirectURI:  redirectURI,
		CreatedAt:    time.Now(),
	}
	m.states[state] = data
	return state, data, nil
}

func (m *mockStateManager) GetState(state string) (*StateData, error) {
	data, ok := m.states[state]
	if !ok {
		return nil, ErrInvalidState
	}
	return data, nil
}

func (m *mockStateManager) DeleteState(state string) {
	delete(m.states, state)
}

func (m *mockStateManager) Cleanup() {}

// SetState allows tests to manually set state data
func (m *mockStateManager) SetState(state string, data *StateData) {
	m.states[state] = data
}

func createTestProviders() []config.OIDCProvider {
	return []config.OIDCProvider{
		{
			Name:         "Test Provider",
			Slug:         "test",
			IssuerURL:    "https://issuer.example.com",
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Scopes:       "openid email profile",
		},
		{
			Name:         "Another Provider",
			Slug:         "another",
			IssuerURL:    "https://another.example.com",
			ClientID:     "another-client-id",
			ClientSecret: "another-client-secret",
			Scopes:       "openid email",
		},
	}
}

func TestGetProviders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	providers := createTestProviders()
	svc := NewService(
		providers,
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	)

	result, err := svc.GetProviders(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "test", result[0].Slug)
	assert.Equal(t, "Test Provider", result[0].Name)
	assert.Equal(t, "another", result[1].Slug)
	assert.Equal(t, "Another Provider", result[1].Name)
}

func TestGetProviders_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	svc := NewService(
		[]config.OIDCProvider{},
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	)

	result, err := svc.GetProviders(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetUserIdentities_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	providers := createTestProviders()
	svc := NewService(
		providers,
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	)

	userID := uuid.New()
	email := "test@example.com"
	now := time.Now()

	mockIdentityRepo.EXPECT().GetByUserID(gomock.Any(), userID).Return([]*oidc_identity.OIDCIdentity{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Issuer:    "https://issuer.example.com",
			Subject:   "sub-123",
			Email:     &email,
			CreatedAt: now,
		},
	}, nil)

	result, err := svc.GetUserIdentities(context.Background(), userID)

	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "test", result[0].ProviderSlug)
	assert.Equal(t, "Test Provider", result[0].ProviderName)
	assert.Equal(t, &email, result[0].Email)
}

func TestGetUserIdentities_NoIdentities(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	providers := createTestProviders()
	svc := NewService(
		providers,
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	)

	userID := uuid.New()

	mockIdentityRepo.EXPECT().GetByUserID(gomock.Any(), userID).Return([]*oidc_identity.OIDCIdentity{}, nil)

	result, err := svc.GetUserIdentities(context.Background(), userID)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetUserIdentities_UnknownProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	providers := createTestProviders()
	svc := NewService(
		providers,
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	)

	userID := uuid.New()
	email := "test@example.com"
	now := time.Now()

	// Return identity from unknown provider - should be filtered out
	mockIdentityRepo.EXPECT().GetByUserID(gomock.Any(), userID).Return([]*oidc_identity.OIDCIdentity{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Issuer:    "https://unknown.example.com",
			Subject:   "sub-123",
			Email:     &email,
			CreatedAt: now,
		},
	}, nil)

	result, err := svc.GetUserIdentities(context.Background(), userID)

	require.NoError(t, err)
	assert.Empty(t, result) // Unknown provider should be filtered out
}

func TestUnlinkIdentity_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	providers := createTestProviders()
	svc := NewService(
		providers,
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	)

	userID := uuid.New()

	mockIdentityRepo.EXPECT().DeleteByUserIDAndIssuer(
		gomock.Any(),
		userID,
		"https://issuer.example.com",
	).Return(nil)

	err := svc.UnlinkIdentity(context.Background(), userID, "test")

	require.NoError(t, err)
}

func TestUnlinkIdentity_ProviderNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	providers := createTestProviders()
	svc := NewService(
		providers,
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	)

	userID := uuid.New()

	err := svc.UnlinkIdentity(context.Background(), userID, "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, ErrProviderNotFound, err)
}

func TestStateManager_CreateAndGetState(t *testing.T) {
	sm := NewStateManager(10)

	state, data, err := sm.CreateState("test-provider", "http://redirect.com")

	require.NoError(t, err)
	assert.NotEmpty(t, state)
	assert.Equal(t, "test-provider", data.ProviderSlug)
	assert.Equal(t, "http://redirect.com", data.RedirectURI)
	assert.NotEmpty(t, data.CodeVerifier)
	assert.NotEmpty(t, data.Nonce)

	// Get the state
	retrieved, err := sm.GetState(state)

	require.NoError(t, err)
	assert.Equal(t, data.ProviderSlug, retrieved.ProviderSlug)
	assert.Equal(t, data.CodeVerifier, retrieved.CodeVerifier)
	assert.Equal(t, data.Nonce, retrieved.Nonce)
}

func TestStateManager_InvalidState(t *testing.T) {
	sm := NewStateManager(10)

	_, err := sm.GetState("nonexistent-state")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidState, err)
}

func TestStateManager_DeleteState(t *testing.T) {
	sm := NewStateManager(10)

	state, _, err := sm.CreateState("test-provider", "http://redirect.com")
	require.NoError(t, err)

	// Verify state exists
	_, err = sm.GetState(state)
	require.NoError(t, err)

	// Delete state
	sm.DeleteState(state)

	// Verify state no longer exists
	_, err = sm.GetState(state)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidState, err)
}

func TestGenerateCodeChallenge(t *testing.T) {
	verifier := "test-code-verifier"
	challenge := GenerateCodeChallenge(verifier)

	assert.NotEmpty(t, challenge)
	// Same verifier should always produce same challenge
	assert.Equal(t, challenge, GenerateCodeChallenge(verifier))
	// Different verifiers should produce different challenges
	assert.NotEqual(t, challenge, GenerateCodeChallenge("different-verifier"))
}

func TestGenerateUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	svc := NewService(
		[]config.OIDCProvider{},
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	).(*service)

	// Test email prefix
	username := svc.generateUsername("john.doe@example.com", "", "")
	assert.Contains(t, username, "john.doe_")

	// Test name fallback
	username = svc.generateUsername("", "John Doe", "")
	assert.Contains(t, username, "john_doe_")

	// Test subject fallback
	username = svc.generateUsername("", "", "subject-123")
	assert.Contains(t, username, "user_")
}

func TestNilIfEmpty(t *testing.T) {
	assert.Nil(t, nilIfEmpty(""))

	value := nilIfEmpty("test")
	assert.NotNil(t, value)
	assert.Equal(t, "test", *value)
}

func TestFindOrCreateUser_ExistingIdentity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	providers := createTestProviders()
	svc := NewService(
		providers,
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	).(*service)

	provider := &providers[0]
	userID := uuid.New()
	email := "existing@example.com"

	existingIdentity := &oidc_identity.OIDCIdentity{
		ID:      uuid.New(),
		UserID:  userID,
		Issuer:  provider.IssuerURL,
		Subject: "sub-123",
	}

	existingUser := &user.User{
		ID:       userID,
		Username: "existinguser",
		Email:    &email,
	}

	mockIdentityRepo.EXPECT().GetByIssuerAndSubject(
		gomock.Any(),
		provider.IssuerURL,
		"sub-123",
	).Return(existingIdentity, nil)

	mockUserRepo.EXPECT().GetByID(gomock.Any(), userID).Return(existingUser, nil)

	// User info might be updated if claims have new data
	mockUserRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	claims := &struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Nonce         string `json:"nonce"`
	}{
		Subject:       "sub-123",
		Email:         "existing@example.com",
		EmailVerified: true,
		Name:          "Existing User",
	}

	result, err := svc.findOrCreateUser(context.Background(), provider, claims)

	require.NoError(t, err)
	assert.False(t, result.IsNewUser)
	assert.False(t, result.LinkedToExisting)
	assert.Equal(t, userID, result.User.ID)
}

func TestFindOrCreateUser_LinkToExistingByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	providers := createTestProviders()
	svc := NewService(
		providers,
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	).(*service)

	provider := &providers[0]
	userID := uuid.New()
	email := "existing@example.com"

	existingUser := &user.User{
		ID:       userID,
		Username: "existinguser",
		Email:    &email,
	}

	// Identity doesn't exist
	mockIdentityRepo.EXPECT().GetByIssuerAndSubject(
		gomock.Any(),
		provider.IssuerURL,
		"new-sub-456",
	).Return(nil, gorm.ErrRecordNotFound)

	// But user with email exists
	mockUserRepo.EXPECT().GetByEmail(gomock.Any(), "existing@example.com").Return(existingUser, nil)

	// Link identity to existing user
	mockIdentityRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	// Update user info
	mockUserRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	claims := &struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Nonce         string `json:"nonce"`
	}{
		Subject:       "new-sub-456",
		Email:         "existing@example.com",
		EmailVerified: true,
		Name:          "Existing User",
	}

	result, err := svc.findOrCreateUser(context.Background(), provider, claims)

	require.NoError(t, err)
	assert.False(t, result.IsNewUser)
	assert.True(t, result.LinkedToExisting)
	assert.Equal(t, userID, result.User.ID)
}

func TestFindOrCreateUser_CreateNewUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	providers := createTestProviders()
	svc := NewService(
		providers,
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	).(*service)

	provider := &providers[0]

	// Identity doesn't exist
	mockIdentityRepo.EXPECT().GetByIssuerAndSubject(
		gomock.Any(),
		provider.IssuerURL,
		"new-sub-789",
	).Return(nil, gorm.ErrRecordNotFound)

	// User with email doesn't exist
	mockUserRepo.EXPECT().GetByEmail(gomock.Any(), "newuser@example.com").Return(nil, gorm.ErrRecordNotFound)

	// Create new user
	mockUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, u *user.User) error {
		u.ID = uuid.New()
		return nil
	})

	// Create identity
	mockIdentityRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	claims := &struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Nonce         string `json:"nonce"`
	}{
		Subject:       "new-sub-789",
		Email:         "newuser@example.com",
		EmailVerified: true,
		Name:          "New User",
	}

	result, err := svc.findOrCreateUser(context.Background(), provider, claims)

	require.NoError(t, err)
	assert.True(t, result.IsNewUser)
	assert.False(t, result.LinkedToExisting)
	assert.NotEqual(t, uuid.Nil, result.User.ID)
}

func TestGenerateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	svc := NewService(
		[]config.OIDCProvider{},
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	).(*service)

	userID := uuid.New()
	token, err := svc.generateToken(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestRewriteEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIdentityRepo := oidc_identity_mocks.NewMockRepository(ctrl)
	mockUserRepo := user_mocks.NewMockRepository(ctrl)
	stateManager := newMockStateManager()

	svc := NewService(
		[]config.OIDCProvider{},
		mockIdentityRepo,
		mockUserRepo,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
		"test-jwt-secret",
		24,
	).(*service)

	endpoint := oauth2.Endpoint{
		AuthURL:  "http://localhost:5556/dex/auth",
		TokenURL: "http://localhost:5556/dex/token",
	}

	rewritten := svc.rewriteEndpoint(endpoint, "http://localhost:5556/dex", "http://dex:5556/dex")

	// AuthURL should stay the same (browser needs localhost)
	assert.Equal(t, "http://localhost:5556/dex/auth", rewritten.AuthURL)
	// TokenURL should be rewritten (backend needs container hostname)
	assert.Equal(t, "http://dex:5556/dex/token", rewritten.TokenURL)
}
