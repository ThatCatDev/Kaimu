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
	StoryPoints *int         `gorm:"type:integer"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
	CreatedBy   *uuid.UUID   `gorm:"type:uuid"`
}

// CardSprint represents the many-to-many relationship between cards and sprints
type CardSprint struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CardID   uuid.UUID `gorm:"type:uuid;not null"`
	SprintID uuid.UUID `gorm:"type:uuid;not null"`
	AddedAt  time.Time `gorm:"autoCreateTime"`
}

func (CardSprint) TableName() string {
	return "card_sprints"
}

func (Card) TableName() string {
	return "cards"
}
