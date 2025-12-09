package project

//go:generate mockgen -source=project_repository.go -destination=mocks/project_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Project, error)
	GetByKey(ctx context.Context, orgID uuid.UUID, key string) (*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, project *Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	var project Project
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *repository) GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]*Project, error) {
	var projects []*Project
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *repository) GetByKey(ctx context.Context, orgID uuid.UUID, key string) (*Project, error) {
	var project Project
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND key = ?", orgID, key).
		First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *repository) Update(ctx context.Context, project *Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Project{}, "id = ?", id).Error
}
