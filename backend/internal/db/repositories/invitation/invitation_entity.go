package invitation

import (
	"time"

	"github.com/google/uuid"
)

type Invitation struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null"`
	Email          string     `gorm:"type:varchar(255);not null"`
	RoleID         *uuid.UUID `gorm:"type:uuid"`
	InvitedBy      uuid.UUID  `gorm:"type:uuid;not null"`
	Token          string     `gorm:"type:varchar(255);uniqueIndex;not null"`
	ExpiresAt      time.Time  `gorm:"not null"`
	AcceptedAt     *time.Time
	CreatedAt      time.Time `gorm:"autoCreateTime"`
}

func (Invitation) TableName() string {
	return "invitations"
}

// IsExpired returns true if the invitation has expired
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsAccepted returns true if the invitation has been accepted
func (i *Invitation) IsAccepted() bool {
	return i.AcceptedAt != nil
}

// IsPending returns true if the invitation is still pending (not expired and not accepted)
func (i *Invitation) IsPending() bool {
	return !i.IsExpired() && !i.IsAccepted()
}
