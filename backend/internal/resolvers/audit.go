package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	auditrepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/audit"
	"github.com/thatcatdev/kaimu/backend/internal/services/audit"
	boardService "github.com/thatcatdev/kaimu/backend/internal/services/board"
	orgService "github.com/thatcatdev/kaimu/backend/internal/services/organization"
	projectService "github.com/thatcatdev/kaimu/backend/internal/services/project"
	rbacService "github.com/thatcatdev/kaimu/backend/internal/services/rbac"
	userService "github.com/thatcatdev/kaimu/backend/internal/services/user"
)

const defaultLimit = 20
const maxLimit = 50

// AuditServices holds all services needed for audit resolvers
type AuditServices struct {
	UserSvc    userService.Service
	OrgSvc     orgService.Service
	ProjectSvc projectService.Service
	BoardSvc   boardService.Service
}

// OrganizationActivity returns audit events for an organization
func OrganizationActivity(
	ctx context.Context,
	rbacSvc rbacService.Service,
	auditSvc audit.Service,
	services *AuditServices,
	organizationID string,
	first *int,
	after *string,
	filters *model.AuditFilters,
) (*model.AuditEventConnection, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasPermission, err := rbacSvc.HasOrgPermission(ctx, *userID, orgID, "org:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	limit := defaultLimit
	if first != nil && *first > 0 {
		limit = *first
		if limit > maxLimit {
			limit = maxLimit
		}
	}

	offset := 0
	if after != nil {
		offset, err = auditDecodeCursor(*after)
		if err != nil {
			return nil, err
		}
	}

	var events []*auditrepo.AuditEvent
	var total int64

	if filters != nil && hasFilters(filters) {
		queryFilters := convertFilters(filters)
		events, total, err = auditSvc.GetOrganizationActivityWithFilters(ctx, orgID, queryFilters, limit, offset)
	} else {
		events, total, err = auditSvc.GetOrganizationActivity(ctx, orgID, limit, offset)
	}
	if err != nil {
		return nil, err
	}

	return buildAuditEventConnection(ctx, events, total, limit, offset, services), nil
}

// ProjectActivity returns audit events for a project
func ProjectActivity(
	ctx context.Context,
	rbacSvc rbacService.Service,
	auditSvc audit.Service,
	services *AuditServices,
	projectID string,
	first *int,
	after *string,
) (*model.AuditEventConnection, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	pID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasPermission, err := rbacSvc.HasProjectPermission(ctx, *userID, pID, "project:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	limit := defaultLimit
	if first != nil && *first > 0 {
		limit = *first
		if limit > maxLimit {
			limit = maxLimit
		}
	}

	offset := 0
	if after != nil {
		offset, err = auditDecodeCursor(*after)
		if err != nil {
			return nil, err
		}
	}

	events, total, err := auditSvc.GetProjectActivity(ctx, pID, limit, offset)
	if err != nil {
		return nil, err
	}

	return buildAuditEventConnection(ctx, events, total, limit, offset, services), nil
}

// BoardActivity returns audit events for a board
func BoardActivity(
	ctx context.Context,
	rbacSvc rbacService.Service,
	auditSvc audit.Service,
	services *AuditServices,
	boardID string,
	first *int,
	after *string,
) (*model.AuditEventConnection, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	bID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, bID, "board:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	limit := defaultLimit
	if first != nil && *first > 0 {
		limit = *first
		if limit > maxLimit {
			limit = maxLimit
		}
	}

	offset := 0
	if after != nil {
		offset, err = auditDecodeCursor(*after)
		if err != nil {
			return nil, err
		}
	}

	events, total, err := auditSvc.GetBoardActivity(ctx, bID, limit, offset)
	if err != nil {
		return nil, err
	}

	return buildAuditEventConnection(ctx, events, total, limit, offset, services), nil
}

// EntityHistory returns audit events for a specific entity
func EntityHistory(
	ctx context.Context,
	rbacSvc rbacService.Service,
	auditSvc audit.Service,
	services *AuditServices,
	entityType model.AuditEntityType,
	entityID string,
	first *int,
	after *string,
) (*model.AuditEventConnection, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	eID, err := uuid.Parse(entityID)
	if err != nil {
		return nil, err
	}

	// TODO: Add permission check based on entity type
	// For now, allow authenticated users to view entity history

	limit := defaultLimit
	if first != nil && *first > 0 {
		limit = *first
		if limit > maxLimit {
			limit = maxLimit
		}
	}

	offset := 0
	if after != nil {
		offset, err = auditDecodeCursor(*after)
		if err != nil {
			return nil, err
		}
	}

	repoEntityType := modelEntityTypeToRepo(entityType)
	events, total, err := auditSvc.GetEntityHistory(ctx, repoEntityType, eID, limit, offset)
	if err != nil {
		return nil, err
	}

	return buildAuditEventConnection(ctx, events, total, limit, offset, services), nil
}

// UserActivity returns audit events by a specific user
func UserActivity(
	ctx context.Context,
	rbacSvc rbacService.Service,
	auditSvc audit.Service,
	services *AuditServices,
	targetUserID string,
	first *int,
	after *string,
) (*model.AuditEventConnection, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	targetID, err := uuid.Parse(targetUserID)
	if err != nil {
		return nil, err
	}

	// Users can view their own activity, or admins can view any user's activity
	// For now, only allow viewing own activity
	if *userID != targetID {
		return nil, ErrUnauthorized
	}

	limit := defaultLimit
	if first != nil && *first > 0 {
		limit = *first
		if limit > maxLimit {
			limit = maxLimit
		}
	}

	offset := 0
	if after != nil {
		offset, err = auditDecodeCursor(*after)
		if err != nil {
			return nil, err
		}
	}

	events, total, err := auditSvc.GetUserActivity(ctx, targetID, limit, offset)
	if err != nil {
		return nil, err
	}

	return buildAuditEventConnection(ctx, events, total, limit, offset, services), nil
}

// Helper functions

func hasFilters(filters *model.AuditFilters) bool {
	return filters != nil && (len(filters.Actions) > 0 || len(filters.EntityTypes) > 0 || filters.ActorID != nil || filters.StartDate != nil || filters.EndDate != nil)
}

func convertFilters(filters *model.AuditFilters) auditrepo.QueryFilters {
	qf := auditrepo.QueryFilters{}

	if len(filters.Actions) > 0 {
		qf.Actions = make([]auditrepo.AuditAction, len(filters.Actions))
		for i, a := range filters.Actions {
			qf.Actions[i] = modelActionToRepo(a)
		}
	}

	if len(filters.EntityTypes) > 0 {
		qf.EntityTypes = make([]auditrepo.EntityType, len(filters.EntityTypes))
		for i, e := range filters.EntityTypes {
			qf.EntityTypes[i] = modelEntityTypeToRepo(e)
		}
	}

	if filters.ActorID != nil {
		actorID, err := uuid.Parse(*filters.ActorID)
		if err == nil {
			qf.ActorID = &actorID
		}
	}

	if filters.StartDate != nil {
		qf.StartDate = filters.StartDate
	}

	if filters.EndDate != nil {
		qf.EndDate = filters.EndDate
	}

	return qf
}

func buildAuditEventConnection(ctx context.Context, events []*auditrepo.AuditEvent, total int64, limit, offset int, services *AuditServices) *model.AuditEventConnection {
	edges := make([]*model.AuditEventEdge, len(events))
	for i, e := range events {
		edges[i] = &model.AuditEventEdge{
			Node:   auditEventToModel(ctx, e, services),
			Cursor: auditEncodeCursor(offset + i),
		}
	}

	hasNext := offset+len(events) < int(total)
	hasPrev := offset > 0

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.AuditEventConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNext,
			HasPreviousPage: hasPrev,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: int(total),
	}
}

func auditEventToModel(ctx context.Context, e *auditrepo.AuditEvent, services *AuditServices) *model.AuditEvent {
	event := &model.AuditEvent{
		ID:         e.ID.String(),
		OccurredAt: e.OccurredAt,
		Action:     repoActionToModel(e.Action),
		EntityType: repoEntityTypeToModel(e.EntityType),
		EntityID:   e.EntityID.String(),
	}

	// Fetch related entities if services provided
	if services != nil {
		// Fetch actor
		if e.ActorID != nil && services.UserSvc != nil {
			if user, err := services.UserSvc.GetByID(ctx, *e.ActorID); err == nil && user != nil {
				event.Actor = UserToModel(user)
			}
		}

		// Fetch organization
		if e.OrganizationID != nil && services.OrgSvc != nil {
			if org, err := services.OrgSvc.GetOrganization(ctx, *e.OrganizationID); err == nil && org != nil {
				event.Organization = OrganizationToModel(org)
			}
		}

		// Fetch project
		if e.ProjectID != nil && services.ProjectSvc != nil {
			if proj, err := services.ProjectSvc.GetProject(ctx, *e.ProjectID); err == nil && proj != nil {
				event.Project = ProjectToModel(proj)
			}
		}

		// Fetch board
		if e.BoardID != nil && services.BoardSvc != nil {
			if board, err := services.BoardSvc.GetBoard(ctx, *e.BoardID); err == nil && board != nil {
				event.Board = BoardToModel(board)
			}
		}
	}

	// Convert JSONB fields to strings for the GraphQL response
	if e.StateBefore != nil {
		s := string(e.StateBefore)
		event.StateBefore = &s
	}
	if e.StateAfter != nil {
		s := string(e.StateAfter)
		event.StateAfter = &s
	}
	if e.Metadata != nil {
		s := string(e.Metadata)
		event.Metadata = &s
	}

	event.IPAddress = e.IPAddress
	event.UserAgent = e.UserAgent
	event.TraceID = e.TraceID

	return event
}

func auditEncodeCursor(offset int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("offset:%d", offset)))
}

func auditDecodeCursor(cursor string) (int, error) {
	data, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, err
	}
	var offset int
	_, err = fmt.Sscanf(string(data), "offset:%d", &offset)
	if err != nil {
		return 0, err
	}
	return offset + 1, nil // +1 to get the next page
}

// Conversion functions between model and repository types

func modelActionToRepo(a model.AuditAction) auditrepo.AuditAction {
	switch a {
	case model.AuditActionCreated:
		return auditrepo.ActionCreated
	case model.AuditActionUpdated:
		return auditrepo.ActionUpdated
	case model.AuditActionDeleted:
		return auditrepo.ActionDeleted
	case model.AuditActionCardMoved:
		return auditrepo.ActionCardMoved
	case model.AuditActionCardAssigned:
		return auditrepo.ActionCardAssigned
	case model.AuditActionCardUnassigned:
		return auditrepo.ActionCardUnassigned
	case model.AuditActionSprintStarted:
		return auditrepo.ActionSprintStarted
	case model.AuditActionSprintCompleted:
		return auditrepo.ActionSprintCompleted
	case model.AuditActionCardAddedToSprint:
		return auditrepo.ActionCardAddedToSprint
	case model.AuditActionCardRemovedFromSprint:
		return auditrepo.ActionCardRemovedFromSprint
	case model.AuditActionMemberInvited:
		return auditrepo.ActionMemberInvited
	case model.AuditActionMemberJoined:
		return auditrepo.ActionMemberJoined
	case model.AuditActionMemberRemoved:
		return auditrepo.ActionMemberRemoved
	case model.AuditActionMemberRoleChanged:
		return auditrepo.ActionMemberRoleChanged
	case model.AuditActionColumnReordered:
		return auditrepo.ActionColumnReordered
	case model.AuditActionColumnVisibilityToggled:
		return auditrepo.ActionColumnVisibilityToggled
	case model.AuditActionUserLoggedIn:
		return auditrepo.ActionUserLoggedIn
	case model.AuditActionUserLoggedOut:
		return auditrepo.ActionUserLoggedOut
	default:
		return auditrepo.ActionCreated
	}
}

func repoActionToModel(a auditrepo.AuditAction) model.AuditAction {
	switch a {
	case auditrepo.ActionCreated:
		return model.AuditActionCreated
	case auditrepo.ActionUpdated:
		return model.AuditActionUpdated
	case auditrepo.ActionDeleted:
		return model.AuditActionDeleted
	case auditrepo.ActionCardMoved:
		return model.AuditActionCardMoved
	case auditrepo.ActionCardAssigned:
		return model.AuditActionCardAssigned
	case auditrepo.ActionCardUnassigned:
		return model.AuditActionCardUnassigned
	case auditrepo.ActionSprintStarted:
		return model.AuditActionSprintStarted
	case auditrepo.ActionSprintCompleted:
		return model.AuditActionSprintCompleted
	case auditrepo.ActionCardAddedToSprint:
		return model.AuditActionCardAddedToSprint
	case auditrepo.ActionCardRemovedFromSprint:
		return model.AuditActionCardRemovedFromSprint
	case auditrepo.ActionMemberInvited:
		return model.AuditActionMemberInvited
	case auditrepo.ActionMemberJoined:
		return model.AuditActionMemberJoined
	case auditrepo.ActionMemberRemoved:
		return model.AuditActionMemberRemoved
	case auditrepo.ActionMemberRoleChanged:
		return model.AuditActionMemberRoleChanged
	case auditrepo.ActionColumnReordered:
		return model.AuditActionColumnReordered
	case auditrepo.ActionColumnVisibilityToggled:
		return model.AuditActionColumnVisibilityToggled
	case auditrepo.ActionUserLoggedIn:
		return model.AuditActionUserLoggedIn
	case auditrepo.ActionUserLoggedOut:
		return model.AuditActionUserLoggedOut
	default:
		return model.AuditActionCreated
	}
}

func modelEntityTypeToRepo(e model.AuditEntityType) auditrepo.EntityType {
	switch e {
	case model.AuditEntityTypeUser:
		return auditrepo.EntityUser
	case model.AuditEntityTypeOrganization:
		return auditrepo.EntityOrganization
	case model.AuditEntityTypeProject:
		return auditrepo.EntityProject
	case model.AuditEntityTypeBoard:
		return auditrepo.EntityBoard
	case model.AuditEntityTypeBoardColumn:
		return auditrepo.EntityBoardColumn
	case model.AuditEntityTypeCard:
		return auditrepo.EntityCard
	case model.AuditEntityTypeSprint:
		return auditrepo.EntitySprint
	case model.AuditEntityTypeTag:
		return auditrepo.EntityTag
	case model.AuditEntityTypeRole:
		return auditrepo.EntityRole
	case model.AuditEntityTypeInvitation:
		return auditrepo.EntityInvitation
	default:
		return auditrepo.EntityUser
	}
}

func repoEntityTypeToModel(e auditrepo.EntityType) model.AuditEntityType {
	switch e {
	case auditrepo.EntityUser:
		return model.AuditEntityTypeUser
	case auditrepo.EntityOrganization:
		return model.AuditEntityTypeOrganization
	case auditrepo.EntityProject:
		return model.AuditEntityTypeProject
	case auditrepo.EntityBoard:
		return model.AuditEntityTypeBoard
	case auditrepo.EntityBoardColumn:
		return model.AuditEntityTypeBoardColumn
	case auditrepo.EntityCard:
		return model.AuditEntityTypeCard
	case auditrepo.EntitySprint:
		return model.AuditEntityTypeSprint
	case auditrepo.EntityTag:
		return model.AuditEntityTypeTag
	case auditrepo.EntityRole:
		return model.AuditEntityTypeRole
	case auditrepo.EntityInvitation:
		return model.AuditEntityTypeInvitation
	default:
		return model.AuditEntityTypeUser
	}
}

// Unused but might be needed for date parsing
var _ = time.Now
var _ = strconv.Itoa
