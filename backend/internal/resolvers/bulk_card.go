package resolvers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/card"
	boardService "github.com/thatcatdev/kaimu/backend/internal/services/board"
	cardService "github.com/thatcatdev/kaimu/backend/internal/services/card"
	rbacService "github.com/thatcatdev/kaimu/backend/internal/services/rbac"
	sprintService "github.com/thatcatdev/kaimu/backend/internal/services/sprint"
)

const maxBulkBatchSize = 100

var ErrBatchTooLarge = fmt.Errorf("batch size exceeds maximum of %d", maxBulkBatchSize)

func parseCardIDs(ids []string) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid card ID at index %d: %w", i, err)
		}
		uuids[i] = parsed
	}
	return uuids, nil
}

func parseTagIDs(ids []string) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("invalid tag ID at index %d: %w", i, err)
		}
		uuids[i] = parsed
	}
	return uuids, nil
}

func checkBoardPermission(ctx context.Context, rbacSvc rbacService.Service, boardSvc boardService.Service, userID uuid.UUID, boardID string, permission string) error {
	bID, err := uuid.Parse(boardID)
	if err != nil {
		return err
	}

	proj, err := boardSvc.GetProject(ctx, bID)
	if err != nil {
		return err
	}

	hasPermission, err := rbacSvc.HasProjectPermission(ctx, userID, proj.ID, permission)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrUnauthorized
	}
	return nil
}

func cardsToModelSlice(cards []*card.Card) []*model.Card {
	result := make([]*model.Card, len(cards))
	for i, c := range cards {
		result[i] = cardToModel(c)
	}
	return result
}

func toBulkCardResult(cards []*card.Card) *model.BulkCardResult {
	modelCards := cardsToModelSlice(cards)
	return &model.BulkCardResult{
		SuccessCount: len(modelCards),
		TotalCount:   len(modelCards),
		Cards:        modelCards,
	}
}

// BulkMoveCardsToColumn moves multiple cards to a target column
func BulkMoveCardsToColumn(
	ctx context.Context,
	rbacSvc rbacService.Service,
	cardSvc cardService.Service,
	boardSvc boardService.Service,
	input model.BulkMoveCardsToColumnInput,
) (*model.BulkCardResult, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	if len(input.CardIds) > maxBulkBatchSize {
		return nil, ErrBatchTooLarge
	}

	if err := checkBoardPermission(ctx, rbacSvc, boardSvc, *userID, input.BoardID, "card:move"); err != nil {
		return nil, err
	}

	cardIDs, err := parseCardIDs(input.CardIds)
	if err != nil {
		return nil, err
	}

	targetColID, err := uuid.Parse(input.TargetColumnID)
	if err != nil {
		return nil, err
	}

	cards, err := cardSvc.BulkMoveToColumn(ctx, cardIDs, targetColID)
	if err != nil {
		return nil, err
	}

	return toBulkCardResult(cards), nil
}

// BulkUpdateCardSprints adds or removes multiple cards from a sprint
func BulkUpdateCardSprints(
	ctx context.Context,
	rbacSvc rbacService.Service,
	sprintSvc sprintService.Service,
	boardSvc boardService.Service,
	input model.BulkUpdateCardSprintsInput,
) (*model.BulkCardResult, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	if len(input.CardIds) > maxBulkBatchSize {
		return nil, ErrBatchTooLarge
	}

	if err := checkBoardPermission(ctx, rbacSvc, boardSvc, *userID, input.BoardID, "card:move"); err != nil {
		return nil, err
	}

	cardIDs, err := parseCardIDs(input.CardIds)
	if err != nil {
		return nil, err
	}

	sprintID, err := uuid.Parse(input.SprintID)
	if err != nil {
		return nil, err
	}

	var cards []*card.Card
	if input.Add {
		cards, err = sprintSvc.BulkAddCardsToSprint(ctx, cardIDs, sprintID)
	} else {
		cards, err = sprintSvc.BulkRemoveCardsFromSprint(ctx, cardIDs, sprintID)
	}
	if err != nil {
		return nil, err
	}

	return toBulkCardResult(cards), nil
}

// BulkUpdateCardProperties updates properties on multiple cards
func BulkUpdateCardProperties(
	ctx context.Context,
	rbacSvc rbacService.Service,
	cardSvc cardService.Service,
	boardSvc boardService.Service,
	input model.BulkUpdateCardPropertiesInput,
) (*model.BulkCardResult, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	if len(input.CardIds) > maxBulkBatchSize {
		return nil, ErrBatchTooLarge
	}

	if err := checkBoardPermission(ctx, rbacSvc, boardSvc, *userID, input.BoardID, "card:edit"); err != nil {
		return nil, err
	}

	cardIDs, err := parseCardIDs(input.CardIds)
	if err != nil {
		return nil, err
	}

	svcInput := cardService.BulkUpdatePropertiesInput{
		CardIDs: cardIDs,
	}

	if input.Priority != nil {
		p := modelPriorityToCard(*input.Priority)
		svcInput.Priority = &p
	}
	if input.ClearAssignee != nil && *input.ClearAssignee {
		svcInput.ClearAssignee = true
	} else if input.AssigneeID != nil {
		assigneeID, err := uuid.Parse(*input.AssigneeID)
		if err != nil {
			return nil, err
		}
		svcInput.AssigneeID = &assigneeID
	}
	if input.ClearStoryPoints != nil && *input.ClearStoryPoints {
		svcInput.ClearStoryPoints = true
	} else if input.StoryPoints != nil {
		svcInput.StoryPoints = input.StoryPoints
	}

	cards, err := cardSvc.BulkUpdateProperties(ctx, svcInput)
	if err != nil {
		return nil, err
	}

	return toBulkCardResult(cards), nil
}

// BulkTagCards adds, removes, or sets tags on multiple cards
func BulkTagCards(
	ctx context.Context,
	rbacSvc rbacService.Service,
	cardSvc cardService.Service,
	boardSvc boardService.Service,
	input model.BulkTagCardsInput,
) (*model.BulkCardResult, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	if len(input.CardIds) > maxBulkBatchSize {
		return nil, ErrBatchTooLarge
	}

	if err := checkBoardPermission(ctx, rbacSvc, boardSvc, *userID, input.BoardID, "card:edit"); err != nil {
		return nil, err
	}

	cardIDs, err := parseCardIDs(input.CardIds)
	if err != nil {
		return nil, err
	}

	tagIDs, err := parseTagIDs(input.TagIds)
	if err != nil {
		return nil, err
	}

	var cards []*card.Card
	switch input.Operation {
	case model.TagOperationAdd:
		cards, err = cardSvc.BulkAddTags(ctx, cardIDs, tagIDs)
	case model.TagOperationRemove:
		cards, err = cardSvc.BulkRemoveTags(ctx, cardIDs, tagIDs)
	case model.TagOperationSet:
		cards, err = cardSvc.BulkSetTags(ctx, cardIDs, tagIDs)
	default:
		return nil, fmt.Errorf("invalid tag operation: %s", input.Operation)
	}
	if err != nil {
		return nil, err
	}

	return toBulkCardResult(cards), nil
}

// BulkDeleteCards deletes multiple cards
func BulkDeleteCards(
	ctx context.Context,
	rbacSvc rbacService.Service,
	cardSvc cardService.Service,
	boardSvc boardService.Service,
	input model.BulkDeleteCardsInput,
) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	if len(input.CardIds) > maxBulkBatchSize {
		return false, ErrBatchTooLarge
	}

	if err := checkBoardPermission(ctx, rbacSvc, boardSvc, *userID, input.BoardID, "card:delete"); err != nil {
		return false, err
	}

	cardIDs, err := parseCardIDs(input.CardIds)
	if err != nil {
		return false, err
	}

	if err := cardSvc.BulkDelete(ctx, cardIDs); err != nil {
		return false, err
	}

	return true, nil
}

// BulkMoveCardsToBacklog removes multiple cards from all sprints
func BulkMoveCardsToBacklog(
	ctx context.Context,
	rbacSvc rbacService.Service,
	sprintSvc sprintService.Service,
	boardSvc boardService.Service,
	input model.BulkMoveCardsToBacklogInput,
) (*model.BulkCardResult, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	if len(input.CardIds) > maxBulkBatchSize {
		return nil, ErrBatchTooLarge
	}

	if err := checkBoardPermission(ctx, rbacSvc, boardSvc, *userID, input.BoardID, "card:move"); err != nil {
		return nil, err
	}

	cardIDs, err := parseCardIDs(input.CardIds)
	if err != nil {
		return nil, err
	}

	cards, err := sprintSvc.BulkMoveCardsToBacklog(ctx, cardIDs)
	if err != nil {
		return nil, err
	}

	return toBulkCardResult(cards), nil
}
