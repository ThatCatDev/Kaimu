package organization_member

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationMember struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null"`
	Role           string     `gorm:"type:varchar(50);not null;default:'member'"` // Deprecated: use RoleID
	RoleID         *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}
