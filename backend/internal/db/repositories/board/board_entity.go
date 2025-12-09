package board

import (
	"time"

	"github.com/google/uuid"
)

type Board struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID  `gorm:"type:uuid;not null"`
	Name        string     `gorm:"type:varchar(255);not null"`
	Description string     `gorm:"type:text"`
	IsDefault   bool       `gorm:"type:boolean;not null;default:false"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid"`
}

func (Board) TableName() string {
	return "boards"
}
