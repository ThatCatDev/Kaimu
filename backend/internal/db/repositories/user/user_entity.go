package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Username     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}
