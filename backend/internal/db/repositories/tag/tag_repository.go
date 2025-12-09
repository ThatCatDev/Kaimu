package tag

//go:generate mockgen -source=tag_repository.go -destination=mocks/tag_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, tag *Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tag, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*Tag, error)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Tag, error)
	GetByName(ctx context.Context, projectID uuid.UUID, name string) (*Tag, error)
	Update(ctx context.Context, tag *Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, tag *Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Tag, error) {
	var tag Tag
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*Tag, error) {
	var tags []*Tag
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("name ASC").
		Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *repository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*Tag, error) {
	var tags []*Tag
	if len(ids) == 0 {
		return tags, nil
	}
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *repository) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*Tag, error) {
	var tag Tag
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID, name).
		First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *repository) Update(ctx context.Context, tag *Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Tag{}, "id = ?", id).Error
}
