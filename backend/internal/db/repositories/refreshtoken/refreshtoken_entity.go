package refreshtoken

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null"`
	TokenHash  string     `gorm:"type:varchar(255);not null"`
	ExpiresAt  time.Time  `gorm:"not null"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	RevokedAt  *time.Time `gorm:"type:timestamp with time zone"`
	ReplacedBy *uuid.UUID `gorm:"type:uuid"`
	UserAgent  *string    `gorm:"type:text"`
	IPAddress  *string    `gorm:"type:varchar(45)"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsValid checks if the token is not expired and not revoked
func (rt *RefreshToken) IsValid() bool {
	return rt.RevokedAt == nil && time.Now().Before(rt.ExpiresAt)
}
