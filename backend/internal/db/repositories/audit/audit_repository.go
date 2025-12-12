package audit

//go:generate mockgen -source=audit_repository.go -destination=mocks/audit_repository_mock.go -package=mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QueryFilters contains optional filters for audit event queries
type QueryFilters struct {
	Actions     []AuditAction
	EntityTypes []EntityType
	ActorID     *uuid.UUID
	StartDate   *time.Time
	EndDate     *time.Time
}

type Repository interface {
	// Write operations
	Create(ctx context.Context, event *AuditEvent) error
	CreateBatch(ctx context.Context, events []*AuditEvent) error

	// Query by organization (activity feed)
	GetByOrganizationID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error)
	GetByOrganizationIDWithFilters(ctx context.Context, orgID uuid.UUID, filters QueryFilters, limit, offset int) ([]*AuditEvent, int64, error)

	// Query by project
	GetByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error)

	// Query by board
	GetByBoardID(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error)

	// Query by entity (entity history)
	GetByEntity(ctx context.Context, entityType EntityType, entityID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error)

	// Query by actor (user activity)
	GetByActorID(ctx context.Context, actorID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error)

	// Metrics queries for burn charts
	GetCardMovementsByBoardAndDateRange(ctx context.Context, boardID uuid.UUID, startDate, endDate time.Time) ([]*AuditEvent, error)
	GetSprintCardEvents(ctx context.Context, sprintID uuid.UUID, startDate, endDate time.Time) ([]*AuditEvent, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, event *AuditEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *repository) CreateBatch(ctx context.Context, events []*AuditEvent) error {
	if len(events) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(events).Error
}

func (r *repository) GetByOrganizationID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error) {
	var events []*AuditEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditEvent{}).Where("organization_id = ?", orgID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("occurred_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *repository) GetByOrganizationIDWithFilters(ctx context.Context, orgID uuid.UUID, filters QueryFilters, limit, offset int) ([]*AuditEvent, int64, error) {
	var events []*AuditEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditEvent{}).Where("organization_id = ?", orgID)

	// Apply filters
	if len(filters.Actions) > 0 {
		query = query.Where("action IN ?", filters.Actions)
	}
	if len(filters.EntityTypes) > 0 {
		query = query.Where("entity_type IN ?", filters.EntityTypes)
	}
	if filters.ActorID != nil {
		query = query.Where("actor_id = ?", *filters.ActorID)
	}
	if filters.StartDate != nil {
		query = query.Where("occurred_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("occurred_at <= ?", *filters.EndDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("occurred_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *repository) GetByProjectID(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error) {
	var events []*AuditEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditEvent{}).Where("project_id = ?", projectID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("occurred_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *repository) GetByBoardID(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error) {
	var events []*AuditEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditEvent{}).Where("board_id = ?", boardID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("occurred_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *repository) GetByEntity(ctx context.Context, entityType EntityType, entityID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error) {
	var events []*AuditEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditEvent{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("occurred_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

func (r *repository) GetByActorID(ctx context.Context, actorID uuid.UUID, limit, offset int) ([]*AuditEvent, int64, error) {
	var events []*AuditEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&AuditEvent{}).Where("actor_id = ?", actorID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("occurred_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// GetCardMovementsByBoardAndDateRange returns card movement events for metrics calculation
func (r *repository) GetCardMovementsByBoardAndDateRange(ctx context.Context, boardID uuid.UUID, startDate, endDate time.Time) ([]*AuditEvent, error) {
	var events []*AuditEvent

	err := r.db.WithContext(ctx).
		Where("board_id = ?", boardID).
		Where("entity_type = ?", EntityCard).
		Where("action IN ?", []AuditAction{
			ActionCreated,
			ActionDeleted,
			ActionCardMoved,
			ActionCardAddedToSprint,
			ActionCardRemovedFromSprint,
		}).
		Where("occurred_at >= ? AND occurred_at <= ?", startDate, endDate).
		Order("occurred_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetSprintCardEvents returns card events related to a specific sprint
func (r *repository) GetSprintCardEvents(ctx context.Context, sprintID uuid.UUID, startDate, endDate time.Time) ([]*AuditEvent, error) {
	var events []*AuditEvent

	// Query events where the sprint_id is in metadata
	err := r.db.WithContext(ctx).
		Where("entity_type = ?", EntityCard).
		Where("action IN ?", []AuditAction{
			ActionCreated,
			ActionDeleted,
			ActionCardMoved,
			ActionCardAddedToSprint,
			ActionCardRemovedFromSprint,
		}).
		Where("occurred_at >= ? AND occurred_at <= ?", startDate, endDate).
		Where("metadata->>'sprint_id' = ? OR metadata->'sprint_ids' ? ?", sprintID.String(), sprintID.String()).
		Order("occurred_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}

	return events, nil
}
