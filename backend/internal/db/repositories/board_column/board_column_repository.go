package board_column

//go:generate mockgen -source=board_column_repository.go -destination=mocks/board_column_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, column *BoardColumn) error
	GetByID(ctx context.Context, id uuid.UUID) (*BoardColumn, error)
	GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*BoardColumn, error)
	GetVisibleByBoardID(ctx context.Context, boardID uuid.UUID) ([]*BoardColumn, error)
	GetMaxPosition(ctx context.Context, boardID uuid.UUID) (int, error)
	Update(ctx context.Context, column *BoardColumn) error
	UpdatePositions(ctx context.Context, columns []*BoardColumn) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, column *BoardColumn) error {
	return r.db.WithContext(ctx).Create(column).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*BoardColumn, error) {
	var column BoardColumn
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&column).Error
	if err != nil {
		return nil, err
	}
	return &column, nil
}

func (r *repository) GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*BoardColumn, error) {
	var columns []*BoardColumn
	err := r.db.WithContext(ctx).
		Where("board_id = ?", boardID).
		Order("position ASC").
		Find(&columns).Error
	if err != nil {
		return nil, err
	}
	return columns, nil
}

func (r *repository) GetVisibleByBoardID(ctx context.Context, boardID uuid.UUID) ([]*BoardColumn, error) {
	var columns []*BoardColumn
	err := r.db.WithContext(ctx).
		Where("board_id = ? AND is_hidden = FALSE", boardID).
		Order("position ASC").
		Find(&columns).Error
	if err != nil {
		return nil, err
	}
	return columns, nil
}

func (r *repository) GetMaxPosition(ctx context.Context, boardID uuid.UUID) (int, error) {
	var maxPos *int
	err := r.db.WithContext(ctx).
		Model(&BoardColumn{}).
		Where("board_id = ?", boardID).
		Select("COALESCE(MAX(position), -1)").
		Scan(&maxPos).Error
	if err != nil {
		return 0, err
	}
	if maxPos == nil {
		return -1, nil
	}
	return *maxPos, nil
}

func (r *repository) Update(ctx context.Context, column *BoardColumn) error {
	return r.db.WithContext(ctx).Save(column).Error
}

func (r *repository) UpdatePositions(ctx context.Context, columns []*BoardColumn) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, col := range columns {
			if err := tx.Model(&BoardColumn{}).
				Where("id = ?", col.ID).
				Update("position", col.Position).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&BoardColumn{}, "id = ?", id).Error
}
