package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/user/mocks"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewService(mockRepo, "test-secret", 24)

	// User doesn't exist - use gomock.Any() for context since tracing modifies it
	mockRepo.EXPECT().GetByUsername(gomock.Any(), "newuser").Return(nil, gorm.ErrRecordNotFound)

	// Create will be called - use DoAndReturn to set the ID
	mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, u *user.User) error {
		u.ID = uuid.New()
		u.CreatedAt = time.Now()
		return nil
	})

	u, token, err := svc.Register(context.Background(), "newuser", "password123")

	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "newuser", u.Username)
	assert.NotEmpty(t, token)
}

func TestRegister_UserExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewService(mockRepo, "test-secret", 24)

	existingUser := &user.User{
		ID:       uuid.New(),
		Username: "existinguser",
	}

	mockRepo.EXPECT().GetByUsername(gomock.Any(), "existinguser").Return(existingUser, nil)

	u, token, err := svc.Register(context.Background(), "existinguser", "password123")

	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)
	assert.Nil(t, u)
	assert.Empty(t, token)
}

func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewService(mockRepo, "test-secret", 24)

	// Hash password for test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	existingUser := &user.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
	}

	mockRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(existingUser, nil)

	u, token, err := svc.Login(context.Background(), "testuser", "correctpassword")

	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, "testuser", u.Username)
	assert.NotEmpty(t, token)
}

func TestLogin_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewService(mockRepo, "test-secret", 24)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	existingUser := &user.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
	}

	mockRepo.EXPECT().GetByUsername(gomock.Any(), "testuser").Return(existingUser, nil)

	u, token, err := svc.Login(context.Background(), "testuser", "wrongpassword")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, u)
	assert.Empty(t, token)
}

func TestLogin_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewService(mockRepo, "test-secret", 24)

	mockRepo.EXPECT().GetByUsername(gomock.Any(), "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	u, token, err := svc.Login(context.Background(), "nonexistent", "password")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	assert.Nil(t, u)
	assert.Empty(t, token)
}

func TestValidateToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewService(mockRepo, "test-secret", 24)

	// Generate a valid token
	userID := uuid.New()
	s := svc.(*service)
	token, err := s.generateToken(userID)
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)

	require.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewService(mockRepo, "test-secret", 24)

	claims, err := svc.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Nil(t, claims)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc1 := NewService(mockRepo, "secret1", 24)
	svc2 := NewService(mockRepo, "secret2", 24)

	// Generate token with first service
	userID := uuid.New()
	s := svc1.(*service)
	token, _ := s.generateToken(userID)

	// Validate with second service (different secret)
	claims, err := svc2.ValidateToken(token)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
	assert.Nil(t, claims)
}

func TestGetUserByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewService(mockRepo, "test-secret", 24)

	userID := uuid.New()
	expectedUser := &user.User{
		ID:       userID,
		Username: "testuser",
	}

	mockRepo.EXPECT().GetByID(gomock.Any(), userID).Return(expectedUser, nil)

	u, err := svc.GetUserByID(context.Background(), userID)

	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, userID, u.ID)
	assert.Equal(t, "testuser", u.Username)
}

func TestGetUserByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	svc := NewService(mockRepo, "test-secret", 24)

	userID := uuid.New()

	mockRepo.EXPECT().GetByID(gomock.Any(), userID).Return(nil, gorm.ErrRecordNotFound)

	u, err := svc.GetUserByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, u)
}
