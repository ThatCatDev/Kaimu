package organization

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Slug        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Description string    `gorm:"type:text"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (Organization) TableName() string {
	return "organizations"
}
