package email_verification_token

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailVerificationTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, email string, expiresIn time.Duration) (*EmailVerificationToken, error)
	FindByToken(ctx context.Context, token string) (*EmailVerificationToken, error)
	MarkAsUsed(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

type emailVerificationTokenRepository struct {
	db *gorm.DB
}

func NewEmailVerificationTokenRepository(db *gorm.DB) EmailVerificationTokenRepository {
	return &emailVerificationTokenRepository{db: db}
}

func (r *emailVerificationTokenRepository) Create(ctx context.Context, userID uuid.UUID, email string, expiresIn time.Duration) (*EmailVerificationToken, error) {
	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	verificationToken := &EmailVerificationToken{
		UserID:    userID,
		Token:     token,
		Email:     email,
		ExpiresAt: time.Now().Add(expiresIn),
	}

	if err := r.db.WithContext(ctx).Create(verificationToken).Error; err != nil {
		return nil, err
	}

	return verificationToken, nil
}

func (r *emailVerificationTokenRepository) FindByToken(ctx context.Context, token string) (*EmailVerificationToken, error) {
	var verificationToken EmailVerificationToken
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&verificationToken).Error; err != nil {
		return nil, err
	}
	return &verificationToken, nil
}

func (r *emailVerificationTokenRepository) MarkAsUsed(ctx context.Context, token string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&EmailVerificationToken{}).
		Where("token = ?", token).
		Update("used_at", &now).Error
}

func (r *emailVerificationTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ? OR used_at IS NOT NULL", time.Now()).
		Delete(&EmailVerificationToken{}).Error
}

func (r *emailVerificationTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&EmailVerificationToken{}).Error
}
