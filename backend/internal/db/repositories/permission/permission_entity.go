package permission

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Code         string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Name         string    `gorm:"type:varchar(255);not null"`
	Description  *string   `gorm:"type:text"`
	ResourceType string    `gorm:"type:varchar(50);not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (Permission) TableName() string {
	return "permissions"
}
