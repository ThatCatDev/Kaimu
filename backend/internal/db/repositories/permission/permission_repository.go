package permission

//go:generate mockgen -source=permission_repository.go -destination=mocks/permission_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	GetAll(ctx context.Context) ([]*Permission, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	GetByCode(ctx context.Context, code string) (*Permission, error)
	GetByCodes(ctx context.Context, codes []string) ([]*Permission, error)
	GetByResourceType(ctx context.Context, resourceType string) ([]*Permission, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll(ctx context.Context) ([]*Permission, error) {
	var permissions []*Permission
	err := r.db.WithContext(ctx).Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Permission, error) {
	var permission Permission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *repository) GetByCode(ctx context.Context, code string) (*Permission, error) {
	var permission Permission
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *repository) GetByCodes(ctx context.Context, codes []string) ([]*Permission, error) {
	var permissions []*Permission
	err := r.db.WithContext(ctx).Where("code IN ?", codes).Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *repository) GetByResourceType(ctx context.Context, resourceType string) ([]*Permission, error) {
	var permissions []*Permission
	err := r.db.WithContext(ctx).Where("resource_type = ?", resourceType).Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
