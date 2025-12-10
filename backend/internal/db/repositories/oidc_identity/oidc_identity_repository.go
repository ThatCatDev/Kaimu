package oidc_identity

//go:generate mockgen -source=oidc_identity_repository.go -destination=mocks/oidc_identity_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, identity *OIDCIdentity) error
	GetByIssuerAndSubject(ctx context.Context, issuer, subject string) (*OIDCIdentity, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*OIDCIdentity, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserIDAndIssuer(ctx context.Context, userID uuid.UUID, issuer string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, identity *OIDCIdentity) error {
	return r.db.WithContext(ctx).Create(identity).Error
}

func (r *repository) GetByIssuerAndSubject(ctx context.Context, issuer, subject string) (*OIDCIdentity, error) {
	var identity OIDCIdentity
	err := r.db.WithContext(ctx).
		Where("issuer = ? AND subject = ?", issuer, subject).
		First(&identity).Error
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*OIDCIdentity, error) {
	var identities []*OIDCIdentity
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&identities).Error
	if err != nil {
		return nil, err
	}
	return identities, nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&OIDCIdentity{}, "id = ?", id).Error
}

func (r *repository) DeleteByUserIDAndIssuer(ctx context.Context, userID uuid.UUID, issuer string) error {
	return r.db.WithContext(ctx).
		Delete(&OIDCIdentity{}, "user_id = ? AND issuer = ?", userID, issuer).Error
}
