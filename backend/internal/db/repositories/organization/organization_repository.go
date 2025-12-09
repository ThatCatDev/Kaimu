package organization

//go:generate mockgen -source=organization_repository.go -destination=mocks/organization_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, org *Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*Organization, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Organization, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, org *Organization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	var org Organization
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *repository) GetBySlug(ctx context.Context, slug string) (*Organization, error) {
	var org Organization
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *repository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*Organization, error) {
	var orgs []*Organization
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&orgs).Error
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

// GetByUserID returns all organizations the user is a member of (including owned)
func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Organization, error) {
	var orgs []*Organization
	err := r.db.WithContext(ctx).
		Joins("LEFT JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organizations.owner_id = ? OR organization_members.user_id = ?", userID, userID).
		Distinct().
		Find(&orgs).Error
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (r *repository) Update(ctx context.Context, org *Organization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Organization{}, "id = ?", id).Error
}
