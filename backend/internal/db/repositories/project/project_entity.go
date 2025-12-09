package project

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	Name           string    `gorm:"type:varchar(255);not null"`
	Key            string    `gorm:"type:varchar(10);not null"`
	Description    string    `gorm:"type:text"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func (Project) TableName() string {
	return "projects"
}
