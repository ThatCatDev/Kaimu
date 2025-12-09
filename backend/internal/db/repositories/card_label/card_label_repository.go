package card_label

//go:generate mockgen -source=card_label_repository.go -destination=mocks/card_label_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, cardLabel *CardLabel) error
	GetByCardID(ctx context.Context, cardID uuid.UUID) ([]*CardLabel, error)
	GetByLabelID(ctx context.Context, labelID uuid.UUID) ([]*CardLabel, error)
	DeleteByCardID(ctx context.Context, cardID uuid.UUID) error
	DeleteByCardAndLabel(ctx context.Context, cardID, labelID uuid.UUID) error
	SetLabelsForCard(ctx context.Context, cardID uuid.UUID, labelIDs []uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, cardLabel *CardLabel) error {
	return r.db.WithContext(ctx).Create(cardLabel).Error
}

func (r *repository) GetByCardID(ctx context.Context, cardID uuid.UUID) ([]*CardLabel, error) {
	var cardLabels []*CardLabel
	err := r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Find(&cardLabels).Error
	if err != nil {
		return nil, err
	}
	return cardLabels, nil
}

func (r *repository) GetByLabelID(ctx context.Context, labelID uuid.UUID) ([]*CardLabel, error) {
	var cardLabels []*CardLabel
	err := r.db.WithContext(ctx).
		Where("label_id = ?", labelID).
		Find(&cardLabels).Error
	if err != nil {
		return nil, err
	}
	return cardLabels, nil
}

func (r *repository) DeleteByCardID(ctx context.Context, cardID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Delete(&CardLabel{}).Error
}

func (r *repository) DeleteByCardAndLabel(ctx context.Context, cardID, labelID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("card_id = ? AND label_id = ?", cardID, labelID).
		Delete(&CardLabel{}).Error
}

func (r *repository) SetLabelsForCard(ctx context.Context, cardID uuid.UUID, labelIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing labels for this card
		if err := tx.Where("card_id = ?", cardID).Delete(&CardLabel{}).Error; err != nil {
			return err
		}

		// Insert new labels
		for _, labelID := range labelIDs {
			cardLabel := CardLabel{
				CardID:  cardID,
				LabelID: labelID,
			}
			if err := tx.Create(&cardLabel).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
