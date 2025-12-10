package role

//go:generate mockgen -source=role_repository.go -destination=mocks/role_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Role, error)
	GetSystemRoles(ctx context.Context) ([]*Role, error)
	GetAllForOrg(ctx context.Context, orgID uuid.UUID) ([]*Role, error) // System roles + org custom roles
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, role *Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Role, error) {
	var role Role
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Role, error) {
	var roles []*Role
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *repository) GetSystemRoles(ctx context.Context) ([]*Role, error) {
	var roles []*Role
	err := r.db.WithContext(ctx).Where("is_system = ?", true).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetAllForOrg returns system roles + organization custom roles
func (r *repository) GetAllForOrg(ctx context.Context, orgID uuid.UUID) ([]*Role, error) {
	var roles []*Role
	err := r.db.WithContext(ctx).
		Where("is_system = ? OR organization_id = ?", true, orgID).
		Order("is_system DESC, name ASC").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *repository) Update(ctx context.Context, role *Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Role{}, "id = ?", id).Error
}
