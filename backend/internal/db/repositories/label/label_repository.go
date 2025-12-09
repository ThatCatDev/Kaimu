package label

//go:generate mockgen -source=label_repository.go -destination=mocks/label_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, label *Label) error
	GetByID(ctx context.Context, id uuid.UUID) (*Label, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*Label, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Label, error)
	GetByName(ctx context.Context, projectID uuid.UUID, name string) (*Label, error)
	Update(ctx context.Context, label *Label) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, label *Label) error {
	return r.db.WithContext(ctx).Create(label).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Label, error) {
	var label Label
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&label).Error
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (r *repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*Label, error) {
	var labels []*Label
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("name ASC").
		Find(&labels).Error
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func (r *repository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Label, error) {
	var labels []*Label
	if len(ids) == 0 {
		return labels, nil
	}
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&labels).Error
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func (r *repository) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*Label, error) {
	var label Label
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID, name).
		First(&label).Error
	if err != nil {
		return nil, err
	}
	return &label, nil
}

func (r *repository) Update(ctx context.Context, label *Label) error {
	return r.db.WithContext(ctx).Save(label).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Label{}, "id = ?", id).Error
}
