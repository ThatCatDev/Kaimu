package card_tag

//go:generate mockgen -source=card_tag_repository.go -destination=mocks/card_tag_repository_mock.go -package=mocks

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, cardTag *CardTag) error
	GetByCardID(ctx context.Context, cardID uuid.UUID) ([]*CardTag, error)
	GetByTagID(ctx context.Context, tagID uuid.UUID) ([]*CardTag, error)
	DeleteByCardID(ctx context.Context, cardID uuid.UUID) error
	DeleteByCardAndTag(ctx context.Context, cardID, tagID uuid.UUID) error
	SetTagsForCard(ctx context.Context, cardID uuid.UUID, tagIDs []uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, cardTag *CardTag) error {
	return r.db.WithContext(ctx).Create(cardTag).Error
}

func (r *repository) GetByCardID(ctx context.Context, cardID uuid.UUID) ([]*CardTag, error) {
	var cardTags []*CardTag
	err := r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Find(&cardTags).Error
	if err != nil {
		return nil, err
	}
	return cardTags, nil
}

func (r *repository) GetByTagID(ctx context.Context, tagID uuid.UUID) ([]*CardTag, error) {
	var cardTags []*CardTag
	err := r.db.WithContext(ctx).
		Where("tag_id = ?", tagID).
		Find(&cardTags).Error
	if err != nil {
		return nil, err
	}
	return cardTags, nil
}

func (r *repository) DeleteByCardID(ctx context.Context, cardID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("card_id = ?", cardID).
		Delete(&CardTag{}).Error
}

func (r *repository) DeleteByCardAndTag(ctx context.Context, cardID, tagID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("card_id = ? AND tag_id = ?", cardID, tagID).
		Delete(&CardTag{}).Error
}

func (r *repository) SetTagsForCard(ctx context.Context, cardID uuid.UUID, tagIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing tags for this card
		if err := tx.Where("card_id = ?", cardID).Delete(&CardTag{}).Error; err != nil {
			return err
		}

		// Insert new tags
		for _, tagID := range tagIDs {
			cardTag := CardTag{
				CardID: cardID,
				TagID:  tagID,
			}
			if err := tx.Create(&cardTag).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
