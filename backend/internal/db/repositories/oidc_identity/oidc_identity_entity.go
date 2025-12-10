package oidc_identity

import (
	"time"

	"github.com/google/uuid"
)

type OIDCIdentity struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"`
	Issuer        string    `gorm:"type:varchar(512);not null"`
	Subject       string    `gorm:"type:varchar(512);not null"`
	Email         *string   `gorm:"type:varchar(255)"`
	EmailVerified bool      `gorm:"default:false"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func (OIDCIdentity) TableName() string {
	return "oidc_identities"
}
