package audit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditAction represents the type of action performed
type AuditAction string

const (
	ActionCreated               AuditAction = "created"
	ActionUpdated               AuditAction = "updated"
	ActionDeleted               AuditAction = "deleted"
	ActionCardMoved             AuditAction = "card_moved"
	ActionCardAssigned          AuditAction = "card_assigned"
	ActionCardUnassigned        AuditAction = "card_unassigned"
	ActionSprintStarted         AuditAction = "sprint_started"
	ActionSprintCompleted       AuditAction = "sprint_completed"
	ActionCardAddedToSprint     AuditAction = "card_added_to_sprint"
	ActionCardRemovedFromSprint AuditAction = "card_removed_from_sprint"
	ActionMemberInvited         AuditAction = "member_invited"
	ActionMemberJoined          AuditAction = "member_joined"
	ActionMemberRemoved         AuditAction = "member_removed"
	ActionMemberRoleChanged     AuditAction = "member_role_changed"
	ActionColumnReordered       AuditAction = "column_reordered"
	ActionColumnVisibilityToggled AuditAction = "column_visibility_toggled"
	ActionUserLoggedIn          AuditAction = "user_logged_in"
	ActionUserLoggedOut         AuditAction = "user_logged_out"
)

// EntityType represents the type of entity being audited
type EntityType string

const (
	EntityUser         EntityType = "user"
	EntityOrganization EntityType = "organization"
	EntityProject      EntityType = "project"
	EntityBoard        EntityType = "board"
	EntityBoardColumn  EntityType = "board_column"
	EntityCard         EntityType = "card"
	EntitySprint       EntityType = "sprint"
	EntityTag          EntityType = "tag"
	EntityRole         EntityType = "role"
	EntityInvitation   EntityType = "invitation"
)

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	ID             uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OccurredAt     time.Time       `gorm:"type:timestamptz;not null;default:now()"`
	ActorID        *uuid.UUID      `gorm:"type:uuid"`
	Action         AuditAction     `gorm:"type:audit_action;not null"`
	EntityType     EntityType      `gorm:"type:audit_entity_type;not null"`
	EntityID       uuid.UUID       `gorm:"type:uuid;not null"`
	OrganizationID *uuid.UUID      `gorm:"type:uuid"`
	ProjectID      *uuid.UUID      `gorm:"type:uuid"`
	BoardID        *uuid.UUID      `gorm:"type:uuid"`
	StateBefore    json.RawMessage `gorm:"type:jsonb"`
	StateAfter     json.RawMessage `gorm:"type:jsonb"`
	Metadata       json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"`
	IPAddress      *string         `gorm:"type:inet"`
	UserAgent      *string         `gorm:"type:text"`
	TraceID        *string         `gorm:"type:text"`
	CreatedAt      time.Time       `gorm:"autoCreateTime"`
}

func (AuditEvent) TableName() string {
	return "audit_events"
}

// SetStateBefore serializes the before state into JSONB
func (e *AuditEvent) SetStateBefore(state interface{}) error {
	if state == nil {
		e.StateBefore = nil
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	e.StateBefore = data
	return nil
}

// SetStateAfter serializes the after state into JSONB
func (e *AuditEvent) SetStateAfter(state interface{}) error {
	if state == nil {
		e.StateAfter = nil
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	e.StateAfter = data
	return nil
}

// SetMetadata serializes metadata into JSONB
func (e *AuditEvent) SetMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		e.Metadata = json.RawMessage("{}")
		return nil
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	e.Metadata = data
	return nil
}

// GetStateBefore parses the before state from JSONB
func (e *AuditEvent) GetStateBefore() (map[string]interface{}, error) {
	if e.StateBefore == nil {
		return nil, nil
	}
	var state map[string]interface{}
	if err := json.Unmarshal(e.StateBefore, &state); err != nil {
		return nil, err
	}
	return state, nil
}

// GetStateAfter parses the after state from JSONB
func (e *AuditEvent) GetStateAfter() (map[string]interface{}, error) {
	if e.StateAfter == nil {
		return nil, nil
	}
	var state map[string]interface{}
	if err := json.Unmarshal(e.StateAfter, &state); err != nil {
		return nil, err
	}
	return state, nil
}

// GetMetadata parses metadata from JSONB
func (e *AuditEvent) GetMetadata() (map[string]interface{}, error) {
	if e.Metadata == nil {
		return nil, nil
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal(e.Metadata, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}
