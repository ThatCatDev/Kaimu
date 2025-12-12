package resolvers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/card"
	boardService "github.com/thatcatdev/kaimu/backend/internal/services/board"
	cardService "github.com/thatcatdev/kaimu/backend/internal/services/card"
	rbacService "github.com/thatcatdev/kaimu/backend/internal/services/rbac"
	tagService "github.com/thatcatdev/kaimu/backend/internal/services/tag"
	userService "github.com/thatcatdev/kaimu/backend/internal/services/user"
)

// Card returns a card by ID
func Card(ctx context.Context, rbacSvc rbacService.Service, cardSvc cardService.Service, boardSvc boardService.Service, id string) (*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	cardID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	c, err := cardSvc.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// Check permission via board -> project
	b, err := cardSvc.GetBoardByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	proj, err := boardSvc.GetProject(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasProjectPermission(ctx, *userID, proj.ID, "card:view")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	return cardToModel(c), nil
}

// MyCards returns all cards assigned to the current user
func MyCards(ctx context.Context, cardSvc cardService.Service) ([]*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	cards, err := cardSvc.GetCardsByAssigneeID(ctx, *userID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Card, len(cards))
	for i, c := range cards {
		result[i] = cardToModel(c)
	}
	return result, nil
}

// CreateCard creates a new card
func CreateCard(ctx context.Context, rbacSvc rbacService.Service, cardSvc cardService.Service, boardSvc boardService.Service, input model.CreateCardInput) (*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	colID, err := uuid.Parse(input.ColumnID)
	if err != nil {
		return nil, err
	}

	// Check permission via column -> board -> project
	b, err := boardSvc.GetBoardByColumnID(ctx, colID)
	if err != nil {
		return nil, err
	}

	proj, err := boardSvc.GetProject(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasProjectPermission(ctx, *userID, proj.ID, "card:create")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	createInput := cardService.CreateCardInput{
		ColumnID:  colID,
		Title:     input.Title,
		Priority:  card.PriorityNone,
		CreatedBy: userID,
	}

	if input.Description != nil {
		createInput.Description = *input.Description
	}
	if input.Priority != nil {
		createInput.Priority = modelPriorityToCard(*input.Priority)
	}
	if input.AssigneeID != nil {
		assigneeID, err := uuid.Parse(*input.AssigneeID)
		if err != nil {
			return nil, err
		}
		createInput.AssigneeID = &assigneeID
	}
	if input.TagIds != nil {
		tagIDs := make([]uuid.UUID, len(input.TagIds))
		for i, id := range input.TagIds {
			tagID, err := uuid.Parse(id)
			if err != nil {
				return nil, err
			}
			tagIDs[i] = tagID
		}
		createInput.TagIDs = tagIDs
	}
	if input.DueDate != nil {
		createInput.DueDate = input.DueDate
	}
	if input.StoryPoints != nil {
		createInput.StoryPoints = input.StoryPoints
	}

	c, err := cardSvc.CreateCard(ctx, createInput)
	if err != nil {
		return nil, err
	}

	return cardToModel(c), nil
}

// UpdateCard updates a card
func UpdateCard(ctx context.Context, rbacSvc rbacService.Service, cardSvc cardService.Service, boardSvc boardService.Service, input model.UpdateCardInput) (*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	cardID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	// Check permission via card -> board -> project
	b, err := cardSvc.GetBoardByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	proj, err := boardSvc.GetProject(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasProjectPermission(ctx, *userID, proj.ID, "card:edit")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	updateInput := cardService.UpdateCardInput{
		ID: cardID,
	}

	if input.Title != nil {
		updateInput.Title = input.Title
	}
	if input.Description != nil {
		updateInput.Description = input.Description
	}
	if input.Priority != nil {
		p := modelPriorityToCard(*input.Priority)
		updateInput.Priority = &p
	}
	if input.ClearAssignee != nil && *input.ClearAssignee {
		updateInput.ClearAssignee = true
	} else if input.AssigneeID != nil {
		assigneeID, err := uuid.Parse(*input.AssigneeID)
		if err != nil {
			return nil, err
		}
		updateInput.AssigneeID = &assigneeID
	}
	if input.TagIds != nil {
		tagIDs := make([]uuid.UUID, len(input.TagIds))
		for i, id := range input.TagIds {
			tagID, err := uuid.Parse(id)
			if err != nil {
				return nil, err
			}
			tagIDs[i] = tagID
		}
		updateInput.TagIDs = tagIDs
	}
	if input.ClearDueDate != nil && *input.ClearDueDate {
		updateInput.ClearDueDate = true
	} else if input.DueDate != nil {
		updateInput.DueDate = input.DueDate
	}
	if input.ClearStoryPoints != nil && *input.ClearStoryPoints {
		updateInput.ClearStoryPoints = true
	} else if input.StoryPoints != nil {
		updateInput.StoryPoints = input.StoryPoints
	}

	c, err := cardSvc.UpdateCard(ctx, updateInput)
	if err != nil {
		return nil, err
	}

	return cardToModel(c), nil
}

// MoveCard moves a card to a different column
func MoveCard(ctx context.Context, rbacSvc rbacService.Service, cardSvc cardService.Service, boardSvc boardService.Service, input model.MoveCardInput) (*model.Card, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	cardID, err := uuid.Parse(input.CardID)
	if err != nil {
		return nil, err
	}

	targetColID, err := uuid.Parse(input.TargetColumnID)
	if err != nil {
		return nil, err
	}

	// Check permission via card -> board -> project
	b, err := cardSvc.GetBoardByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	proj, err := boardSvc.GetProject(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	hasPermission, err := rbacSvc.HasProjectPermission(ctx, *userID, proj.ID, "card:move")
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, ErrUnauthorized
	}

	var afterCardID *uuid.UUID
	if input.AfterCardID != nil {
		id, err := uuid.Parse(*input.AfterCardID)
		if err != nil {
			return nil, err
		}
		afterCardID = &id
	}

	c, err := cardSvc.MoveCard(ctx, cardID, targetColID, afterCardID)
	if err != nil {
		return nil, err
	}

	return cardToModel(c), nil
}

// DeleteCard deletes a card
func DeleteCard(ctx context.Context, rbacSvc rbacService.Service, cardSvc cardService.Service, boardSvc boardService.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	cardID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Check permission via card -> board -> project
	b, err := cardSvc.GetBoardByCardID(ctx, cardID)
	if err != nil {
		return false, err
	}

	proj, err := boardSvc.GetProject(ctx, b.ID)
	if err != nil {
		return false, err
	}

	hasPermission, err := rbacSvc.HasProjectPermission(ctx, *userID, proj.ID, "card:delete")
	if err != nil {
		return false, err
	}
	if !hasPermission {
		return false, ErrUnauthorized
	}

	if err := cardSvc.DeleteCard(ctx, cardID); err != nil {
		return false, err
	}

	return true, nil
}

// CardColumn resolves the column field of a Card
func CardColumn(ctx context.Context, cardSvc cardService.Service, c *model.Card) (*model.BoardColumn, error) {
	cardID, err := uuid.Parse(c.ID)
	if err != nil {
		return nil, err
	}

	col, err := cardSvc.GetColumnByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	return columnToModel(col), nil
}

// CardBoard resolves the board field of a Card
func CardBoard(ctx context.Context, cardSvc cardService.Service, c *model.Card) (*model.Board, error) {
	cardID, err := uuid.Parse(c.ID)
	if err != nil {
		return nil, err
	}

	b, err := cardSvc.GetBoardByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	return boardToModel(b), nil
}

// CardTags resolves the tags field of a Card
func CardTags(ctx context.Context, cardSvc cardService.Service, c *model.Card) ([]*model.Tag, error) {
	cardID, err := uuid.Parse(c.ID)
	if err != nil {
		return nil, err
	}

	tags, err := cardSvc.GetTagsForCard(ctx, cardID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Tag, len(tags))
	for i, t := range tags {
		result[i] = tagToModel(t)
	}
	return result, nil
}

// CardAssignee resolves the assignee field of a Card
func CardAssignee(ctx context.Context, cardSvc cardService.Service, userSvc userService.Service, c *model.Card) (*model.User, error) {
	cardID, err := uuid.Parse(c.ID)
	if err != nil {
		return nil, err
	}

	cardEntity, err := cardSvc.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if cardEntity.AssigneeID == nil {
		return nil, nil
	}

	user, err := userSvc.GetByID(ctx, *cardEntity.AssigneeID)
	if err != nil {
		return nil, err
	}

	return UserToModel(user), nil
}

// CardCreatedBy resolves the createdBy field of a Card
func CardCreatedBy(ctx context.Context, cardSvc cardService.Service, userSvc userService.Service, c *model.Card) (*model.User, error) {
	cardID, err := uuid.Parse(c.ID)
	if err != nil {
		return nil, err
	}

	cardEntity, err := cardSvc.GetCard(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if cardEntity.CreatedBy == nil {
		return nil, nil
	}

	user, err := userSvc.GetByID(ctx, *cardEntity.CreatedBy)
	if err != nil {
		return nil, err
	}

	return UserToModel(user), nil
}

func cardToModel(c *card.Card) *model.Card {
	var description *string
	if c.Description != "" {
		description = &c.Description
	}
	var dueDate *time.Time
	if c.DueDate != nil {
		dueDate = c.DueDate
	}
	return &model.Card{
		ID:          c.ID.String(),
		Title:       c.Title,
		Description: description,
		Position:    c.Position,
		Priority:    cardPriorityToModel(c.Priority),
		DueDate:     dueDate,
		StoryPoints: c.StoryPoints,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// CardToModel converts a card entity to a GraphQL model (exported for audit logging)
func CardToModel(c *card.Card) *model.Card {
	return cardToModel(c)
}

func cardPriorityToModel(p card.CardPriority) model.CardPriority {
	switch p {
	case card.PriorityLow:
		return model.CardPriorityLow
	case card.PriorityMedium:
		return model.CardPriorityMedium
	case card.PriorityHigh:
		return model.CardPriorityHigh
	case card.PriorityUrgent:
		return model.CardPriorityUrgent
	default:
		return model.CardPriorityNone
	}
}

func modelPriorityToCard(p model.CardPriority) card.CardPriority {
	switch p {
	case model.CardPriorityLow:
		return card.PriorityLow
	case model.CardPriorityMedium:
		return card.PriorityMedium
	case model.CardPriorityHigh:
		return card.PriorityHigh
	case model.CardPriorityUrgent:
		return card.PriorityUrgent
	default:
		return card.PriorityNone
	}
}

// ProjectTags resolves the tags field of a Project
func ProjectTags(ctx context.Context, tagSvc tagService.Service, proj *model.Project) ([]*model.Tag, error) {
	projID, err := uuid.Parse(proj.ID)
	if err != nil {
		return nil, err
	}

	tags, err := tagSvc.GetTagsByProjectID(ctx, projID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Tag, len(tags))
	for i, t := range tags {
		result[i] = tagToModel(t)
	}
	return result, nil
}
