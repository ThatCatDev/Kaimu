package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/refreshtoken"
	refreshtokenMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/refreshtoken/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	userMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/user/mocks"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	// User doesn't exist - use gomock.Any() for context since tracing modifies it
	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "newuser").Return(nil, gorm.ErrRecordNotFound)

	// Create user will be called - use DoAndReturn to set the ID
	mockUserRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, u *user.User) error {
		u.ID = uuid.New()
		u.CreatedAt = time.Now()
		return nil
	})

	// Create refresh token will be called
	mockRefreshRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	u, tokenPair, err := svc.Register(context.Background(), "newuser", "email@test.com", "password123", "Test-Agent", "127.0.0.1")

	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "newuser", u.Username)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
}

func TestRegister_UserExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	existingUser := &user.User{
		ID:       uuid.New(),
		Username: "existinguser",
	}

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "existinguser").Return(existingUser, nil)

	u, tokenPair, err := svc.Register(context.Background(), "existinguser", "email@test.com", "password123", "Test-Agent", "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, u)
	assert.Nil(t, tokenPair)
}

func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	// Hash password for test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	passwordStr := string(hashedPassword)
	existingUser := &user.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: &passwordStr,
		CreatedAt:    time.Now(),
	}

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(existingUser, nil)
	mockRefreshRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	u, tokenPair, err := svc.Login(context.Background(), "testuser", "correctpassword", "Test-Agent", "127.0.0.1")

	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "testuser", u.Username)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
}

func TestLogin_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	passwordStr := string(hashedPassword)
	existingUser := &user.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: &passwordStr,
	}

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(existingUser, nil)

	u, tokenPair, err := svc.Login(context.Background(), "testuser", "wrongpassword", "Test-Agent", "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, u)
	assert.Nil(t, tokenPair)
}

func TestLogin_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	u, tokenPair, err := svc.Login(context.Background(), "nonexistent", "password", "Test-Agent", "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, u)
	assert.Nil(t, tokenPair)
}

func TestLogin_PasswordLoginDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	// User without password (OIDC-only user)
	existingUser := &user.User{
		ID:           uuid.New(),
		Username:     "oidcuser",
		PasswordHash: nil,
	}

	mockUserRepo.EXPECT().GetByUsername(gomock.Any(), "oidcuser").Return(existingUser, nil)

	u, tokenPair, err := svc.Login(context.Background(), "oidcuser", "password", "Test-Agent", "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, ErrPasswordLoginDisabled, err)
	assert.Nil(t, u)
	assert.Nil(t, tokenPair)
}

func TestValidateToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	// Generate a valid token
	userID := uuid.New()
	s := svc.(*service)
	token, err := s.generateAccessToken(userID)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)

	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	claims, err := svc.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Nil(t, claims)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc1 := NewService(mockUserRepo, mockRefreshRepo, "secret1", 5, 7)
	svc2 := NewService(mockUserRepo, mockRefreshRepo, "secret2", 5, 7)

	// Generate token with first service
	userID := uuid.New()
	s := svc1.(*service)
	token, _ := s.generateAccessToken(userID)

	// Validate with second service (different secret)
	claims, err := svc2.ValidateToken(token)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Nil(t, claims)
}

func TestGetUserByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	userID := uuid.New()
	expectedUser := &user.User{
		ID:       userID,
		Username: "testuser",
	}

	mockUserRepo.EXPECT().GetByID(gomock.Any(), userID).Return(expectedUser, nil)

	u, err := svc.GetUserByID(context.Background(), userID)

	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, userID, u.ID)
	assert.Equal(t, "testuser", u.Username)
}

func TestGetUserByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	userID := uuid.New()

	mockUserRepo.EXPECT().GetByID(gomock.Any(), userID).Return(nil, gorm.ErrRecordNotFound)

	u, err := svc.GetUserByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, u)
}

func TestRefreshTokens_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	userID := uuid.New()

	// Generate a refresh token (using package functions)
	refreshTokenStr, _ := generateRandomToken(32)
	tokenHash := hashToken(refreshTokenStr)

	storedToken := &refreshtoken.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// Mock expectations in order of execution:
	// 1. Find the old refresh token
	mockRefreshRepo.EXPECT().GetByTokenHash(gomock.Any(), tokenHash).Return(storedToken, nil)
	// 2. Create new refresh token (from generateTokenPairInternal)
	mockRefreshRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	// 3. Get new token to find its ID for replacedByID (returns nil is OK)
	mockRefreshRepo.EXPECT().GetByTokenHash(gomock.Any(), gomock.Any()).Return(nil, nil)
	// 4. Revoke old refresh token
	mockRefreshRepo.EXPECT().Revoke(gomock.Any(), storedToken.ID, gomock.Any()).Return(nil)

	tokenPair, err := svc.RefreshTokens(context.Background(), refreshTokenStr, "Test-Agent", "127.0.0.1")

	require.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
}

func TestRefreshTokens_TokenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	mockRefreshRepo.EXPECT().GetByTokenHash(gomock.Any(), gomock.Any()).Return(nil, gorm.ErrRecordNotFound)

	tokenPair, err := svc.RefreshTokens(context.Background(), "nonexistent-token", "Test-Agent", "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRefreshToken, err)
	assert.Nil(t, tokenPair)
}

func TestRefreshTokens_TokenExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	userID := uuid.New()
	refreshTokenStr, _ := generateRandomToken(32)
	tokenHash := hashToken(refreshTokenStr)

	// Expired token
	storedToken := &refreshtoken.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}

	mockRefreshRepo.EXPECT().GetByTokenHash(gomock.Any(), tokenHash).Return(storedToken, nil)

	tokenPair, err := svc.RefreshTokens(context.Background(), refreshTokenStr, "Test-Agent", "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, ErrRefreshTokenRevoked, err) // Expired tokens are treated as revoked for security
	assert.Nil(t, tokenPair)
}

func TestRefreshTokens_TokenRevoked(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	userID := uuid.New()
	refreshTokenStr, _ := generateRandomToken(32)
	tokenHash := hashToken(refreshTokenStr)

	// Revoked token
	revokedAt := time.Now().Add(-1 * time.Hour)
	storedToken := &refreshtoken.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		RevokedAt: &revokedAt,
	}

	mockRefreshRepo.EXPECT().GetByTokenHash(gomock.Any(), tokenHash).Return(storedToken, nil)
	// When a revoked token is reused, all user tokens should be revoked
	mockRefreshRepo.EXPECT().RevokeAllForUser(gomock.Any(), userID).Return(nil)

	tokenPair, err := svc.RefreshTokens(context.Background(), refreshTokenStr, "Test-Agent", "127.0.0.1")

	assert.Error(t, err)
	assert.Equal(t, ErrRefreshTokenRevoked, err)
	assert.Nil(t, tokenPair)
}

func TestRevokeRefreshToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	refreshTokenStr, _ := generateRandomToken(32)
	tokenHash := hashToken(refreshTokenStr)

	storedToken := &refreshtoken.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	mockRefreshRepo.EXPECT().GetByTokenHash(gomock.Any(), tokenHash).Return(storedToken, nil)
	mockRefreshRepo.EXPECT().Revoke(gomock.Any(), storedToken.ID, nil).Return(nil)

	err := svc.RevokeRefreshToken(context.Background(), refreshTokenStr)

	assert.NoError(t, err)
}

func TestRevokeAllUserTokens_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	userID := uuid.New()

	mockRefreshRepo.EXPECT().RevokeAllForUser(gomock.Any(), userID).Return(nil)

	err := svc.RevokeAllUserTokens(context.Background(), userID)

	assert.NoError(t, err)
}

func TestGenerateTokenPair_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := userMocks.NewMockRepository(ctrl)
	mockRefreshRepo := refreshtokenMocks.NewMockRepository(ctrl)
	svc := NewService(mockUserRepo, mockRefreshRepo, "test-secret", 5, 7)

	userID := uuid.New()

	mockRefreshRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	tokenPair, err := svc.GenerateTokenPair(context.Background(), userID, "Test-Agent", "127.0.0.1")

	require.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, int64(5*60), tokenPair.ExpiresIn)
}
