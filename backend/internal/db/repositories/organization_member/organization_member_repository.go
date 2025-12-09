package organization_member

//go:generate mockgen -source=organization_member_repository.go -destination=mocks/organization_member_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, member *OrganizationMember) error
	GetByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*OrganizationMember, error)
	GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*OrganizationMember, error)
	Delete(ctx context.Context, orgID, userID uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, member *OrganizationMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *repository) GetByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*OrganizationMember, error) {
	var member OrganizationMember
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND user_id = ?", orgID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *repository) GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error) {
	var members []*OrganizationMember
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*OrganizationMember, error) {
	var members []*OrganizationMember
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *repository) Delete(ctx context.Context, orgID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&OrganizationMember{}, "organization_id = ? AND user_id = ?", orgID, userID).Error
}
