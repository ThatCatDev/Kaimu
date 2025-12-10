package role

import (
	"time"

	"github.com/google/uuid"
)

// System role UUIDs - these are fixed for reference in code
var (
	OwnerRoleID  = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	AdminRoleID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	MemberRoleID = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	ViewerRoleID = uuid.MustParse("00000000-0000-0000-0000-000000000004")
)

type Role struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrganizationID *uuid.UUID `gorm:"type:uuid"` // NULL for system roles
	Name           string     `gorm:"type:varchar(100);not null"`
	Description    *string    `gorm:"type:text"`
	IsSystem       bool       `gorm:"type:boolean;not null;default:false"`
	Scope          string     `gorm:"type:varchar(50);not null;default:'organization'"`
	CreatedAt      time.Time  `gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime"`
}

func (Role) TableName() string {
	return "roles"
}

// IsOwnerRole returns true if this is the system Owner role
func (r *Role) IsOwnerRole() bool {
	return r.ID == OwnerRoleID
}

// IsAdminRole returns true if this is the system Admin role
func (r *Role) IsAdminRole() bool {
	return r.ID == AdminRoleID
}
