package project_member

import (
	"time"

	"github.com/google/uuid"
)

type ProjectMember struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ProjectID uuid.UUID  `gorm:"type:uuid;not null"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	RoleID    *uuid.UUID `gorm:"type:uuid"` // NULL means inherit from org
	CreatedAt time.Time  `gorm:"autoCreateTime"`
}

func (ProjectMember) TableName() string {
	return "project_members"
}
