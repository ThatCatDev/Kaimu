package board

//go:generate mockgen -source=board_repository.go -destination=mocks/board_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, board *Board) error
	GetByID(ctx context.Context, id uuid.UUID) (*Board, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*Board, error)
	GetDefaultByProjectID(ctx context.Context, projectID uuid.UUID) (*Board, error)
	GetAll(ctx context.Context) ([]*Board, error)
	Update(ctx context.Context, board *Board) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, board *Board) error {
	return r.db.WithContext(ctx).Create(board).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Board, error) {
	var board Board
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&board).Error
	if err != nil {
		return nil, err
	}
	return &board, nil
}

func (r *repository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*Board, error) {
	var boards []*Board
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at ASC").
		Find(&boards).Error
	if err != nil {
		return nil, err
	}
	return boards, nil
}

func (r *repository) GetDefaultByProjectID(ctx context.Context, projectID uuid.UUID) (*Board, error) {
	var board Board
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND is_default = TRUE", projectID).
		First(&board).Error
	if err != nil {
		return nil, err
	}
	return &board, nil
}

func (r *repository) GetAll(ctx context.Context) ([]*Board, error) {
	var boards []*Board
	err := r.db.WithContext(ctx).Find(&boards).Error
	if err != nil {
		return nil, err
	}
	return boards, nil
}

func (r *repository) Update(ctx context.Context, board *Board) error {
	return r.db.WithContext(ctx).Save(board).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Board{}, "id = ?", id).Error
}
