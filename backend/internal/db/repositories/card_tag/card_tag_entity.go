package card_tag

import (
	"time"

	"github.com/google/uuid"
)

type CardTag struct {
	CardID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	TagID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (CardTag) TableName() string {
	return "card_tags"
}
