package email_verification_token

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerificationToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	Token     string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	Email     string     `gorm:"type:varchar(255);not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time `gorm:""`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}
