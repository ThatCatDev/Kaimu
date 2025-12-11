package card

//go:generate mockgen -source=card_repository.go -destination=mocks/card_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, card *Card) error
	GetByID(ctx context.Context, id uuid.UUID) (*Card, error)
	GetByColumnID(ctx context.Context, columnID uuid.UUID) ([]*Card, error)
	GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Card, error)
	GetByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*Card, error)
	GetAll(ctx context.Context) ([]*Card, error)
	GetMaxPosition(ctx context.Context, columnID uuid.UUID) (float64, error)
	GetPositionBetween(ctx context.Context, columnID uuid.UUID, afterCardID *uuid.UUID) (float64, error)
	Update(ctx context.Context, card *Card) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, card *Card) error {
	return r.db.WithContext(ctx).Create(card).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Card, error) {
	var card Card
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *repository) GetByColumnID(ctx context.Context, columnID uuid.UUID) ([]*Card, error) {
	var cards []*Card
	err := r.db.WithContext(ctx).
		Where("column_id = ?", columnID).
		Order("position ASC").
		Find(&cards).Error
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func (r *repository) GetByBoardID(ctx context.Context, boardID uuid.UUID) ([]*Card, error) {
	var cards []*Card
	err := r.db.WithContext(ctx).
		Where("board_id = ?", boardID).
		Order("position ASC").
		Find(&cards).Error
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func (r *repository) GetByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*Card, error) {
	var cards []*Card
	err := r.db.WithContext(ctx).
		Where("assignee_id = ?", assigneeID).
		Order("due_date ASC NULLS LAST, created_at DESC").
		Find(&cards).Error
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func (r *repository) GetAll(ctx context.Context) ([]*Card, error) {
	var cards []*Card
	err := r.db.WithContext(ctx).Find(&cards).Error
	if err != nil {
		return nil, err
	}
	return cards, nil
}

func (r *repository) GetMaxPosition(ctx context.Context, columnID uuid.UUID) (float64, error) {
	var maxPos *float64
	err := r.db.WithContext(ctx).
		Model(&Card{}).
		Where("column_id = ?", columnID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPos).Error
	if err != nil {
		return 0, err
	}
	if maxPos == nil {
		return 0, nil
	}
	return *maxPos, nil
}

func (r *repository) GetPositionBetween(ctx context.Context, columnID uuid.UUID, afterCardID *uuid.UUID) (float64, error) {
	// If afterCardID is nil, insert at the beginning
	if afterCardID == nil {
		var minPos *float64
		err := r.db.WithContext(ctx).
			Model(&Card{}).
			Where("column_id = ?", columnID).
			Select("MIN(position)").
			Scan(&minPos).Error
		if err != nil {
			return 0, err
		}
		if minPos == nil || *minPos >= 1000 {
			return 500, nil
		}
		return *minPos / 2, nil
	}

	// Get the card we're inserting after
	var afterCard Card
	err := r.db.WithContext(ctx).Where("id = ?", *afterCardID).First(&afterCard).Error
	if err != nil {
		return 0, err
	}

	// Get the next card
	var nextCard Card
	err = r.db.WithContext(ctx).
		Where("column_id = ? AND position > ?", columnID, afterCard.Position).
		Order("position ASC").
		First(&nextCard).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No card after, use afterCard.Position + 1000
			return afterCard.Position + 1000, nil
		}
		return 0, err
	}

	// Return position between the two cards
	return (afterCard.Position + nextCard.Position) / 2, nil
}

func (r *repository) Update(ctx context.Context, card *Card) error {
	return r.db.WithContext(ctx).Save(card).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Card{}, "id = ?", id).Error
}
