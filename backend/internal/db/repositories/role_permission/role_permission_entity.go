package role_permission

import (
	"time"

	"github.com/google/uuid"
)

type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	RoleID       uuid.UUID `gorm:"type:uuid;not null"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
