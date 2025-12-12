package sprint

//go:generate mockgen -source=sprint_repository.go -destination=mocks/sprint_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, sprint *Sprint) error
	GetByID(ctx context.Context, id uuid.UUID) (*Sprint, error)
	GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Sprint, error)
	GetActiveByBoardID(ctx context.Context, boardID uuid.UUID) (*Sprint, error)
	GetFutureByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Sprint, error)
	GetClosedByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Sprint, error)
	GetClosedByBoardIDPaginated(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]*Sprint, int, error)
	Update(ctx context.Context, sprint *Sprint) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetNextPosition(ctx context.Context, boardID uuid.UUID) (int, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, sprint *Sprint) error {
	return r.db.WithContext(ctx).Create(sprint).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Sprint, error) {
	var sprint Sprint
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&sprint).Error
	if err != nil {
		return nil, err
	}
	return &sprint, nil
}

func (r *repository) GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Sprint, error) {
	var sprints []*Sprint
	err := r.db.WithContext(ctx).
		Where("board_id = ?", boardID).
		Order("position ASC, created_at ASC").
		Find(&sprints).Error
	if err != nil {
		return nil, err
	}
	return sprints, nil
}

func (r *repository) GetActiveByBoardID(ctx context.Context, boardID uuid.UUID) (*Sprint, error) {
	var sprint Sprint
	err := r.db.WithContext(ctx).
		Where("board_id = ? AND status = ?", boardID, SprintStatusActive).
		First(&sprint).Error
	if err != nil {
		return nil, err
	}
	return &sprint, nil
}

func (r *repository) GetFutureByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Sprint, error) {
	var sprints []*Sprint
	err := r.db.WithContext(ctx).
		Where("board_id = ? AND status = ?", boardID, SprintStatusFuture).
		Order("position ASC, created_at ASC").
		Find(&sprints).Error
	if err != nil {
		return nil, err
	}
	return sprints, nil
}

func (r *repository) GetClosedByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Sprint, error) {
	var sprints []*Sprint
	err := r.db.WithContext(ctx).
		Where("board_id = ? AND status = ?", boardID, SprintStatusClosed).
		Order("end_date DESC, created_at DESC").
		Find(&sprints).Error
	if err != nil {
		return nil, err
	}
	return sprints, nil
}

func (r *repository) GetClosedByBoardIDPaginated(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]*Sprint, int, error) {
	var sprints []*Sprint
	var totalCount int64

	// Get total count
	err := r.db.WithContext(ctx).
		Model(&Sprint{}).
		Where("board_id = ? AND status = ?", boardID, SprintStatusClosed).
		Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = r.db.WithContext(ctx).
		Where("board_id = ? AND status = ?", boardID, SprintStatusClosed).
		Order("end_date DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sprints).Error
	if err != nil {
		return nil, 0, err
	}

	return sprints, int(totalCount), nil
}

func (r *repository) Update(ctx context.Context, sprint *Sprint) error {
	return r.db.WithContext(ctx).Save(sprint).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Sprint{}, "id = ?", id).Error
}

func (r *repository) GetNextPosition(ctx context.Context, boardID uuid.UUID) (int, error) {
	var maxPosition int
	err := r.db.WithContext(ctx).
		Model(&Sprint{}).
		Where("board_id = ?", boardID).
		Select("COALESCE(MAX(position), -1)").
		Scan(&maxPosition).Error
	if err != nil {
		return 0, err
	}
	return maxPosition + 1, nil
}
