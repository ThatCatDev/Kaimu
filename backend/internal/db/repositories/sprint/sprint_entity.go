package sprint

import (
	"time"

	"github.com/google/uuid"
)

type SprintStatus string

const (
	SprintStatusFuture SprintStatus = "future"
	SprintStatusActive SprintStatus = "active"
	SprintStatusClosed SprintStatus = "closed"
)

type Sprint struct {
	ID      uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BoardID uuid.UUID    `gorm:"type:uuid;not null"`
	Name    string       `gorm:"type:varchar(255);not null"`
	Goal      string       `gorm:"type:text"`
	StartDate *time.Time   `gorm:"type:timestamp with time zone"`
	EndDate   *time.Time   `gorm:"type:timestamp with time zone"`
	Status    SprintStatus `gorm:"type:sprint_status;not null;default:'future'"`
	Position  int          `gorm:"type:integer;not null;default:0"`
	CreatedAt time.Time    `gorm:"autoCreateTime"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime"`
	CreatedBy *uuid.UUID   `gorm:"type:uuid"`
}

func (Sprint) TableName() string {
	return "sprints"
}
