package audit

//go:generate mockgen -source=audit_service.go -destination=mocks/audit_service_mock.go -package=mocks

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	auditrepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/audit"
)

// EventInput contains the data needed to create an audit event
type EventInput struct {
	ActorID        *uuid.UUID
	Action         auditrepo.AuditAction
	EntityType     auditrepo.EntityType
	EntityID       uuid.UUID
	OrganizationID *uuid.UUID
	ProjectID      *uuid.UUID
	BoardID        *uuid.UUID
	StateBefore    interface{}
	StateAfter     interface{}
	Metadata       map[string]interface{}
}

// Service defines the audit logging service interface
type Service interface {
	// LogEvent creates an audit event synchronously
	LogEvent(ctx context.Context, input EventInput) error

	// LogEventAsync creates an audit event asynchronously (fire-and-forget)
	LogEventAsync(ctx context.Context, input EventInput)

	// Query methods for activity feeds
	GetOrganizationActivity(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error)
	GetOrganizationActivityWithFilters(ctx context.Context, orgID uuid.UUID, filters auditrepo.QueryFilters, limit, offset int) ([]*auditrepo.AuditEvent, int64, error)
	GetProjectActivity(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error)
	GetBoardActivity(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error)

	// Query methods for history views
	GetEntityHistory(ctx context.Context, entityType auditrepo.EntityType, entityID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error)
	GetUserActivity(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error)

	// Query methods for metrics
	GetCardMovementsByBoardAndDateRange(ctx context.Context, boardID uuid.UUID, startDate, endDate time.Time) ([]*auditrepo.AuditEvent, error)
	GetSprintCardEvents(ctx context.Context, sprintID uuid.UUID, startDate, endDate time.Time) ([]*auditrepo.AuditEvent, error)
}

type service struct {
	repo auditrepo.Repository
}

// NewService creates a new audit service
func NewService(repo auditrepo.Repository) Service {
	return &service{repo: repo}
}

// LogEvent creates an audit event synchronously
func (s *service) LogEvent(ctx context.Context, input EventInput) error {
	event, err := s.buildEvent(ctx, input)
	if err != nil {
		return err
	}
	return s.repo.Create(ctx, event)
}

// LogEventAsync creates an audit event asynchronously
func (s *service) LogEventAsync(ctx context.Context, input EventInput) {
	// Extract request context before spawning goroutine
	reqCtx := GetRequestContext(ctx)

	go func() {
		// Create a new background context for the async operation
		asyncCtx := context.Background()
		if reqCtx != nil {
			asyncCtx = WithRequestContext(asyncCtx, reqCtx)
		}

		event, err := s.buildEvent(asyncCtx, input)
		if err != nil {
			log.Printf("Failed to build audit event: %v", err)
			return
		}

		if err := s.repo.Create(asyncCtx, event); err != nil {
			log.Printf("Failed to create audit event: %v", err)
		}
	}()
}

// buildEvent constructs an AuditEvent from EventInput and context
func (s *service) buildEvent(ctx context.Context, input EventInput) (*auditrepo.AuditEvent, error) {
	event := &auditrepo.AuditEvent{
		OccurredAt:     time.Now(),
		ActorID:        input.ActorID,
		Action:         input.Action,
		EntityType:     input.EntityType,
		EntityID:       input.EntityID,
		OrganizationID: input.OrganizationID,
		ProjectID:      input.ProjectID,
		BoardID:        input.BoardID,
	}

	// Set state before
	if err := event.SetStateBefore(input.StateBefore); err != nil {
		return nil, err
	}

	// Set state after
	if err := event.SetStateAfter(input.StateAfter); err != nil {
		return nil, err
	}

	// Set metadata
	if err := event.SetMetadata(input.Metadata); err != nil {
		return nil, err
	}

	// Extract request context if available
	if reqCtx := GetRequestContext(ctx); reqCtx != nil {
		if reqCtx.IPAddress != "" {
			event.IPAddress = &reqCtx.IPAddress
		}
		if reqCtx.UserAgent != "" {
			event.UserAgent = &reqCtx.UserAgent
		}
		if reqCtx.TraceID != "" {
			event.TraceID = &reqCtx.TraceID
		}
	}

	return event, nil
}

// GetOrganizationActivity returns audit events for an organization
func (s *service) GetOrganizationActivity(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error) {
	return s.repo.GetByOrganizationID(ctx, orgID, limit, offset)
}

// GetOrganizationActivityWithFilters returns filtered audit events for an organization
func (s *service) GetOrganizationActivityWithFilters(ctx context.Context, orgID uuid.UUID, filters auditrepo.QueryFilters, limit, offset int) ([]*auditrepo.AuditEvent, int64, error) {
	return s.repo.GetByOrganizationIDWithFilters(ctx, orgID, filters, limit, offset)
}

// GetProjectActivity returns audit events for a project
func (s *service) GetProjectActivity(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error) {
	return s.repo.GetByProjectID(ctx, projectID, limit, offset)
}

// GetBoardActivity returns audit events for a board
func (s *service) GetBoardActivity(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error) {
	return s.repo.GetByBoardID(ctx, boardID, limit, offset)
}

// GetEntityHistory returns audit events for a specific entity
func (s *service) GetEntityHistory(ctx context.Context, entityType auditrepo.EntityType, entityID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error) {
	return s.repo.GetByEntity(ctx, entityType, entityID, limit, offset)
}

// GetUserActivity returns audit events by a specific user
func (s *service) GetUserActivity(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*auditrepo.AuditEvent, int64, error) {
	return s.repo.GetByActorID(ctx, userID, limit, offset)
}

// GetCardMovementsByBoardAndDateRange returns card movement events for metrics
func (s *service) GetCardMovementsByBoardAndDateRange(ctx context.Context, boardID uuid.UUID, startDate, endDate time.Time) ([]*auditrepo.AuditEvent, error) {
	return s.repo.GetCardMovementsByBoardAndDateRange(ctx, boardID, startDate, endDate)
}

// GetSprintCardEvents returns card events for a sprint
func (s *service) GetSprintCardEvents(ctx context.Context, sprintID uuid.UUID, startDate, endDate time.Time) ([]*auditrepo.AuditEvent, error) {
	return s.repo.GetSprintCardEvents(ctx, sprintID, startDate, endDate)
}
