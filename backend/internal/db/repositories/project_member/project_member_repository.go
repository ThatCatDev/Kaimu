package project_member

//go:generate mockgen -source=project_member_repository.go -destination=mocks/project_member_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, pm *ProjectMember) error
	GetByID(ctx context.Context, id uuid.UUID) (*ProjectMember, error)
	GetByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*ProjectMember, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*ProjectMember, error)
	Update(ctx context.Context, pm *ProjectMember) error
	Delete(ctx context.Context, projectID, userID uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, pm *ProjectMember) error {
	return r.db.WithContext(ctx).Create(pm).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*ProjectMember, error) {
	var pm ProjectMember
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&pm).Error
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *repository) GetByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*ProjectMember, error) {
	var pm ProjectMember
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&pm).Error
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*ProjectMember, error) {
	var pms []*ProjectMember
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&pms).Error
	if err != nil {
		return nil, err
	}
	return pms, nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*ProjectMember, error) {
	var pms []*ProjectMember
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pms).Error
	if err != nil {
		return nil, err
	}
	return pms, nil
}

func (r *repository) Update(ctx context.Context, pm *ProjectMember) error {
	return r.db.WithContext(ctx).Save(pm).Error
}

func (r *repository) Delete(ctx context.Context, projectID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&ProjectMember{}, "project_id = ? AND user_id = ?", projectID, userID).Error
}
