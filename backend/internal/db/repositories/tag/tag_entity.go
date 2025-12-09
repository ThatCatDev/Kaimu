package tag

import (
	"time"

	"github.com/google/uuid"
)

type Tag struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProjectID   uuid.UUID `gorm:"type:uuid;not null"`
	Name        string    `gorm:"type:varchar(100);not null"`
	Color       string    `gorm:"type:varchar(7);not null;default:'#6B7280'"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

func (Tag) TableName() string {
	return "tags"
}
