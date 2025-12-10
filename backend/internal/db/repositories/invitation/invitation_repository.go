package invitation

//go:generate mockgen -source=invitation_repository.go -destination=mocks/invitation_repository_mock.go -package=mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, inv *Invitation) error
	GetByID(ctx context.Context, id uuid.UUID) (*Invitation, error)
	GetByToken(ctx context.Context, token string) (*Invitation, error)
	GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Invitation, error)
	GetPendingByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Invitation, error)
	GetByOrgAndEmail(ctx context.Context, orgID uuid.UUID, email string) (*Invitation, error)
	Update(ctx context.Context, inv *Invitation) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, inv *Invitation) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Invitation, error) {
	var inv Invitation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *repository) GetByToken(ctx context.Context, token string) (*Invitation, error) {
	var inv Invitation
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *repository) GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Invitation, error) {
	var invs []*Invitation
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&invs).Error
	if err != nil {
		return nil, err
	}
	return invs, nil
}

func (r *repository) GetPendingByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Invitation, error) {
	var invs []*Invitation
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND accepted_at IS NULL AND expires_at > ?", orgID, time.Now()).
		Order("created_at DESC").
		Find(&invs).Error
	if err != nil {
		return nil, err
	}
	return invs, nil
}

func (r *repository) GetByOrgAndEmail(ctx context.Context, orgID uuid.UUID, email string) (*Invitation, error) {
	var inv Invitation
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND email = ?", orgID, email).
		First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *repository) Update(ctx context.Context, inv *Invitation) error {
	return r.db.WithContext(ctx).Save(inv).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Invitation{}, "id = ?", id).Error
}

func (r *repository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Delete(&Invitation{}, "expires_at < ? AND accepted_at IS NULL", time.Now()).Error
}
