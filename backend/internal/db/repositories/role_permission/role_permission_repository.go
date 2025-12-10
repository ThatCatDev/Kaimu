package role_permission

//go:generate mockgen -source=role_permission_repository.go -destination=mocks/role_permission_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/permission"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, rp *RolePermission) error
	CreateBatch(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
	GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]*RolePermission, error)
	GetPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]*permission.Permission, error)
	GetPermissionCodesByRoleID(ctx context.Context, roleID uuid.UUID) ([]string, error)
	DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error
	Delete(ctx context.Context, roleID, permissionID uuid.UUID) error
	ReplaceForRole(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, rp *RolePermission) error {
	return r.db.WithContext(ctx).Create(rp).Error
}

func (r *repository) CreateBatch(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	rolePermissions := make([]*RolePermission, len(permissionIDs))
	for i, permID := range permissionIDs {
		rolePermissions[i] = &RolePermission{
			RoleID:       roleID,
			PermissionID: permID,
		}
	}
	return r.db.WithContext(ctx).Create(&rolePermissions).Error
}

func (r *repository) GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]*RolePermission, error) {
	var rps []*RolePermission
	err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Find(&rps).Error
	if err != nil {
		return nil, err
	}
	return rps, nil
}

func (r *repository) GetPermissionsByRoleID(ctx context.Context, roleID uuid.UUID) ([]*permission.Permission, error) {
	var permissions []*permission.Permission
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *repository) GetPermissionCodesByRoleID(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	var codes []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("permissions.code").
		Joins("INNER JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Pluck("code", &codes).Error
	if err != nil {
		return nil, err
	}
	return codes, nil
}

func (r *repository) DeleteByRoleID(ctx context.Context, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&RolePermission{}, "role_id = ?", roleID).Error
}

func (r *repository) Delete(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&RolePermission{}, "role_id = ? AND permission_id = ?", roleID, permissionID).Error
}

// ReplaceForRole deletes all existing permissions for a role and creates new ones
func (r *repository) ReplaceForRole(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing
		if err := tx.Delete(&RolePermission{}, "role_id = ?", roleID).Error; err != nil {
			return err
		}

		// Create new
		if len(permissionIDs) == 0 {
			return nil
		}

		rolePermissions := make([]*RolePermission, len(permissionIDs))
		for i, permID := range permissionIDs {
			rolePermissions[i] = &RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
		}
		return tx.Create(&rolePermissions).Error
	})
}
