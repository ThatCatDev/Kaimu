package resolvers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/sprint"
	boardService "github.com/thatcatdev/kaimu/backend/internal/services/board"
	rbacService "github.com/thatcatdev/kaimu/backend/internal/services/rbac"
	sprintService "github.com/thatcatdev/kaimu/backend/internal/services/sprint"
	userService "github.com/thatcatdev/kaimu/backend/internal/services/user"
)

// Sprint returns a sprint by ID
func Sprint(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, id string) (*model.Sprint, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	sprintID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	sp, err := sprintSvc.GetSprint(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// Get board to check permission
	board, err := sprintSvc.GetBoard(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, board.ID, "sprint:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	return sprintToModel(sp), nil
}

// Sprints returns all sprints for a board
func Sprints(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, boardID string) ([]*model.Sprint, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	bID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, bID, "sprint:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	sprints, err := sprintSvc.GetBoardSprints(ctx, bID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Sprint, len(sprints))
	for i, sp := range sprints {
		result[i] = sprintToModel(sp)
	}
	return result, nil
}

// ActiveSprint returns the active sprint for a board
func ActiveSprint(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, boardID string) (*model.Sprint, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	bID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, bID, "sprint:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	sp, err := sprintSvc.GetActiveSprint(ctx, bID)
	if err != nil {
		return nil, err
	}
	if sp == nil {
		return nil, nil
	}

	return sprintToModel(sp), nil
}

// FutureSprints returns future sprints for a board
func FutureSprints(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, boardID string) ([]*model.Sprint, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	bID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, bID, "sprint:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	sprints, err := sprintSvc.GetFutureSprints(ctx, bID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Sprint, len(sprints))
	for i, sp := range sprints {
		result[i] = sprintToModel(sp)
	}
	return result, nil
}

// ClosedSprints returns closed sprints for a board with pagination
func ClosedSprints(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, boardID string, first *int, after *string) (*model.SprintConnection, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	bID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, bID, "sprint:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	// Default pagination values
	limit := 20
	if first != nil && *first > 0 {
		limit = *first
	}

	// Parse cursor (offset)
	offset := 0
	if after != nil && *after != "" {
		var err error
		offset, err = parseCursor(*after)
		if err != nil {
			return nil, err
		}
	}

	// Get paginated sprints and total count
	sprints, totalCount, err := sprintSvc.GetClosedSprintsPaginated(ctx, bID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Build edges
	edges := make([]*model.SprintEdge, len(sprints))
	for i, sp := range sprints {
		edges[i] = &model.SprintEdge{
			Node:   sprintToModel(sp),
			Cursor: encodeCursor(offset + i),
		}
	}

	// Build page info
	hasNextPage := offset+len(sprints) < totalCount
	hasPreviousPage := offset > 0

	var startCursor, endCursor *string
	if len(edges) > 0 {
		startCursor = &edges[0].Cursor
		endCursor = &edges[len(edges)-1].Cursor
	}

	return &model.SprintConnection{
		Edges: edges,
		PageInfo: &model.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: hasPreviousPage,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
			TotalCount:      totalCount,
		},
	}, nil
}

// encodeCursor encodes an offset as a cursor string
func encodeCursor(offset int) string {
	return fmt.Sprintf("cursor:%d", offset+1)
}

// parseCursor parses a cursor string to an offset
func parseCursor(cursor string) (int, error) {
	var offset int
	_, err := fmt.Sscanf(cursor, "cursor:%d", &offset)
	if err != nil {
		return 0, err
	}
	return offset - 1 + 1, nil // Convert 1-based to 0-based, then add 1 to skip the item at cursor
}

// SprintCards returns cards in a sprint
func SprintCards(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, sprintID string) ([]*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	spID, err := uuid.Parse(sprintID)
	if err != nil {
		return nil, err
	}

	// Get board to check permission
	board, err := sprintSvc.GetBoard(ctx, spID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, board.ID, "sprint:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	cards, err := sprintSvc.GetSprintCards(ctx, spID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Card, len(cards))
	for i, c := range cards {
		result[i] = cardToModel(c)
	}
	return result, nil
}

// BacklogCards returns backlog cards for a board
func BacklogCards(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, boardSvc boardService.Service, boardID string) ([]*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	bID, err := uuid.Parse(boardID)
	if err != nil {
		return nil, err
	}

	// Check board-level permission
	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, bID, "sprint:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	cards, err := sprintSvc.GetBacklogCards(ctx, bID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Card, len(cards))
	for i, c := range cards {
		result[i] = cardToModel(c)
	}
	return result, nil
}

// CreateSprint creates a new sprint
func CreateSprint(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, input model.CreateSprintInput) (*model.Sprint, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	boardID, err := uuid.Parse(input.BoardID)
	if err != nil {
		return nil, err
	}

	// Check permission
	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, boardID, "sprint:manage")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	goal := ""
	if input.Goal != nil {
		goal = *input.Goal
	}

	sp, err := sprintSvc.CreateSprint(ctx, boardID, input.Name, goal, input.StartDate, input.EndDate, userID)
	if err != nil {
		return nil, err
	}

	return sprintToModel(sp), nil
}

// UpdateSprint updates a sprint
func UpdateSprint(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, id string, input model.UpdateSprintInput) (*model.Sprint, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	sprintID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	// Get board to check permission
	board, err := sprintSvc.GetBoard(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, board.ID, "sprint:manage")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	updateInput := sprintService.UpdateSprintInput{
		Name:      input.Name,
		Goal:      input.Goal,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
	}

	sp, err := sprintSvc.UpdateSprint(ctx, sprintID, updateInput)
	if err != nil {
		return nil, err
	}

	return sprintToModel(sp), nil
}

// DeleteSprint deletes a sprint
func DeleteSprint(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	sprintID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Get board to check permission
	board, err := sprintSvc.GetBoard(ctx, sprintID)
	if err != nil {
		return false, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, board.ID, "sprint:manage")
	if err != nil {
		return false, err
	}
	if !hasPermission {
		return false, ErrUnauthorized
	}

	if err := sprintSvc.DeleteSprint(ctx, sprintID); err != nil {
		return false, err
	}

	return true, nil
}

// StartSprint starts a sprint
func StartSprint(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, id string) (*model.Sprint, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	sprintID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	// Get board to check permission
	board, err := sprintSvc.GetBoard(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, board.ID, "sprint:manage")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	sp, err := sprintSvc.StartSprint(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	return sprintToModel(sp), nil
}

// CompleteSprint completes a sprint
func CompleteSprint(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, id string, moveIncompleteToBacklog bool) (*model.Sprint, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	sprintID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	// Get board to check permission
	board, err := sprintSvc.GetBoard(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, board.ID, "sprint:manage")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	sp, err := sprintSvc.CompleteSprint(ctx, sprintID, moveIncompleteToBacklog)
	if err != nil {
		return nil, err
	}

	return sprintToModel(sp), nil
}

// AddCardToSprint adds a card to a sprint (cards can be in multiple sprints)
func AddCardToSprint(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, input model.MoveCardToSprintInput) (*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	cardID, err := uuid.Parse(input.CardID)
	if err != nil {
		return nil, err
	}

	sprintID, err := uuid.Parse(input.SprintID)
	if err != nil {
		return nil, err
	}

	// Get board to check permission
	board, err := sprintSvc.GetBoard(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, board.ID, "card:move")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	c, err := sprintSvc.AddCardToSprint(ctx, cardID, sprintID)
	if err != nil {
		return nil, err
	}

	return cardToModel(c), nil
}

// RemoveCardFromSprint removes a card from a sprint
func RemoveCardFromSprint(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, input model.MoveCardToSprintInput) (*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	cardID, err := uuid.Parse(input.CardID)
	if err != nil {
		return nil, err
	}

	sprintID, err := uuid.Parse(input.SprintID)
	if err != nil {
		return nil, err
	}

	// Get board to check permission
	board, err := sprintSvc.GetBoard(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, board.ID, "card:move")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	c, err := sprintSvc.RemoveCardFromSprint(ctx, cardID, sprintID)
	if err != nil {
		return nil, err
	}

	return cardToModel(c), nil
}

// SetCardSprints sets all sprints for a card (replaces existing assignments)
func SetCardSprints(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, cardID string, sprintIDs []string) (*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	cID, err := uuid.Parse(cardID)
	if err != nil {
		return nil, err
	}

	// Get card to find its board
	card, err := sprintSvc.GetCardByID(ctx, cID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, card.BoardID, "card:move")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	// Parse sprint IDs
	sIDs := make([]uuid.UUID, len(sprintIDs))
	for i, id := range sprintIDs {
		sID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		sIDs[i] = sID
	}

	c, err := sprintSvc.SetCardSprints(ctx, cID, sIDs)
	if err != nil {
		return nil, err
	}

	return cardToModel(c), nil
}

// MoveCardToBacklog moves a card to backlog
func MoveCardToBacklog(ctx context.Context, rbacSvc rbacService.Service, sprintSvc sprintService.Service, boardSvc boardService.Service, cardID string) (*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	cID, err := uuid.Parse(cardID)
	if err != nil {
		return nil, err
	}

	// Get card to find its board
	card, err := sprintSvc.GetCardByID(ctx, cID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasBoardPermission(ctx, *userID, card.BoardID, "card:move")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	c, err := sprintSvc.MoveCardToBacklog(ctx, cID)
	if err != nil {
		return nil, err
	}

	return cardToModel(c), nil
}

// SprintBoard resolves the board field of a Sprint
func SprintBoard(ctx context.Context, sprintSvc sprintService.Service, sp *model.Sprint) (*model.Board, error) {
	sprintID, err := uuid.Parse(sp.ID)
	if err != nil {
		return nil, err
	}

	board, err := sprintSvc.GetBoard(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	return boardToModel(board), nil
}

// SprintCardsResolver resolves the cards field of a Sprint
func SprintCardsResolver(ctx context.Context, sprintSvc sprintService.Service, sp *model.Sprint) ([]*model.Card, error) {
	sprintID, err := uuid.Parse(sp.ID)
	if err != nil {
		return nil, err
	}

	cards, err := sprintSvc.GetSprintCards(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Card, len(cards))
	for i, c := range cards {
		result[i] = cardToModel(c)
	}
	return result, nil
}

// SprintCreatedBy resolves the createdBy field of a Sprint
func SprintCreatedBy(ctx context.Context, userSvc userService.Service, sprintSvc sprintService.Service, sp *model.Sprint) (*model.User, error) {
	sprintID, err := uuid.Parse(sp.ID)
	if err != nil {
		return nil, err
	}

	sprintEntity, err := sprintSvc.GetSprint(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	if sprintEntity.CreatedBy == nil {
		return nil, nil
	}

	user, err := userSvc.GetByID(ctx, *sprintEntity.CreatedBy)
	if err != nil {
		return nil, err
	}

	return UserToModel(user), nil
}

// BoardSprints resolves the sprints field of a Board
func BoardSprints(ctx context.Context, sprintSvc sprintService.Service, board *model.Board) ([]*model.Sprint, error) {
	boardID, err := uuid.Parse(board.ID)
	if err != nil {
		return nil, err
	}

	sprints, err := sprintSvc.GetBoardSprints(ctx, boardID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Sprint, len(sprints))
	for i, sp := range sprints {
		result[i] = sprintToModel(sp)
	}
	return result, nil
}

// BoardActiveSprint resolves the activeSprint field of a Board
func BoardActiveSprint(ctx context.Context, sprintSvc sprintService.Service, board *model.Board) (*model.Sprint, error) {
	boardID, err := uuid.Parse(board.ID)
	if err != nil {
		return nil, err
	}

	sp, err := sprintSvc.GetActiveSprint(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if sp == nil {
		return nil, nil
	}

	return sprintToModel(sp), nil
}

// CardSprints resolves the sprints field of a Card (many-to-many)
func CardSprints(ctx context.Context, sprintSvc sprintService.Service, c *model.Card) ([]*model.Sprint, error) {
	cardID, err := uuid.Parse(c.ID)
	if err != nil {
		return nil, err
	}

	// Get all sprint IDs for this card
	sprintIDs, err := sprintSvc.GetCardSprintIDs(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if len(sprintIDs) == 0 {
		return []*model.Sprint{}, nil
	}

	// Fetch each sprint
	result := make([]*model.Sprint, 0, len(sprintIDs))
	for _, sprintID := range sprintIDs {
		sp, err := sprintSvc.GetSprint(ctx, sprintID)
		if err != nil {
			// Skip sprints that can't be fetched (shouldn't happen but be safe)
			continue
		}
		result = append(result, sprintToModel(sp))
	}

	return result, nil
}

func sprintToModel(sp *sprint.Sprint) *model.Sprint {
	var goal *string
	if sp.Goal != "" {
		goal = &sp.Goal
	}

	return &model.Sprint{
		ID:        sp.ID.String(),
		Name:      sp.Name,
		Goal:      goal,
		StartDate: sp.StartDate,
		EndDate:   sp.EndDate,
		Status:    sprintStatusToModel(sp.Status),
		Position:  sp.Position,
		CreatedAt: sp.CreatedAt,
		UpdatedAt: sp.UpdatedAt,
		// Board and CreatedBy are resolved by field resolvers
	}
}

func sprintStatusToModel(status sprint.SprintStatus) model.SprintStatus {
	switch status {
	case sprint.SprintStatusActive:
		return model.SprintStatusActive
	case sprint.SprintStatusClosed:
		return model.SprintStatusClosed
	default:
		return model.SprintStatusFuture
	}
}
