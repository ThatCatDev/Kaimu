package card_label

import (
	"time"

	"github.com/google/uuid"
)

type CardLabel struct {
	CardID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	LabelID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (CardLabel) TableName() string {
	return "card_labels"
}
