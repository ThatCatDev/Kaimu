package card

import (
	"time"

	"github.com/google/uuid"
)

type CardPriority string

const (
	PriorityNone   CardPriority = "none"
	PriorityLow    CardPriority = "low"
	PriorityMedium CardPriority = "medium"
	PriorityHigh   CardPriority = "high"
	PriorityUrgent CardPriority = "urgent"
)

type Card struct {
	ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ColumnID    uuid.UUID    `gorm:"type:uuid;not null"`
	BoardID     uuid.UUID    `gorm:"type:uuid;not null"`
	Title       string       `gorm:"type:varchar(500);not null"`
	Description string       `gorm:"type:text"`
	Position    float64      `gorm:"type:float;not null;default:0"`
	Priority    CardPriority `gorm:"type:card_priority;not null;default:'none'"`
	AssigneeID  *uuid.UUID   `gorm:"type:uuid"`
	DueDate     *time.Time   `gorm:"type:timestamptz"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
	CreatedBy   *uuid.UUID   `gorm:"type:uuid"`
}

func (Card) TableName() string {
	return "cards"
}
