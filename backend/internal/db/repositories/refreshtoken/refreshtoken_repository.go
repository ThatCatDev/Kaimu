package refreshtoken

//go:generate mockgen -source=refreshtoken_repository.go -destination=mocks/refreshtoken_repository_mock.go -package=mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	GetByID(ctx context.Context, id uuid.UUID) (*RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID, replacedBy *uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error)
	GetActiveTokensForUser(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, token *RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *repository) GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var token RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*RefreshToken, error) {
	var token RefreshToken
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *repository) Revoke(ctx context.Context, id uuid.UUID, replacedBy *uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&RefreshToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"revoked_at":  now,
			"replaced_by": replacedBy,
		}).Error
}

func (r *repository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func (r *repository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&RefreshToken{})
	return result.RowsAffected, result.Error
}

func (r *repository) GetActiveTokensForUser(ctx context.Context, userID uuid.UUID) ([]*RefreshToken, error) {
	var tokens []*RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
