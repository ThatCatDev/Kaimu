package card

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/board"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/board_column"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/card"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/card_tag"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/tag"
	"github.com/thatcatdev/pulse-backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrCardNotFound   = errors.New("card not found")
	ErrColumnNotFound = errors.New("column not found")
	ErrBoardNotFound  = errors.New("board not found")
)

type CreateCardInput struct {
	ColumnID    uuid.UUID
	Title       string
	Description string
	Priority    card.CardPriority
	AssigneeID  *uuid.UUID
	TagIDs      []uuid.UUID
	DueDate     *time.Time
	CreatedBy   *uuid.UUID
}

type UpdateCardInput struct {
	ID           uuid.UUID
	Title        *string
	Description  *string
	Priority     *card.CardPriority
	AssigneeID   *uuid.UUID
	TagIDs       []uuid.UUID
	DueDate      *time.Time
	ClearDueDate bool
}

type Service interface {
	CreateCard(ctx context.Context, input CreateCardInput) (*card.Card, error)
	GetCard(ctx context.Context, id uuid.UUID) (*card.Card, error)
	GetCardsByColumnID(ctx context.Context, columnID uuid.UUID) ([]*card.Card, error)
	GetCardsByBoardID(ctx context.Context, boardID uuid.UUID) ([]*card.Card, error)
	GetCardsByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*card.Card, error)
	UpdateCard(ctx context.Context, input UpdateCardInput) (*card.Card, error)
	MoveCard(ctx context.Context, cardID, targetColumnID uuid.UUID, afterCardID *uuid.UUID) (*card.Card, error)
	DeleteCard(ctx context.Context, id uuid.UUID) error
	GetTagsForCard(ctx context.Context, cardID uuid.UUID) ([]*tag.Tag, error)
	GetBoardByCardID(ctx context.Context, cardID uuid.UUID) (*board.Board, error)
	GetColumnByCardID(ctx context.Context, cardID uuid.UUID) (*board_column.BoardColumn, error)
}

type service struct {
	cardRepo    card.Repository
	columnRepo  board_column.Repository
	boardRepo   board.Repository
	tagRepo     tag.Repository
	cardTagRepo card_tag.Repository
}

func NewService(
	cardRepo card.Repository,
	columnRepo board_column.Repository,
	boardRepo board.Repository,
	tagRepo tag.Repository,
	cardTagRepo card_tag.Repository,
) Service {
	return &service{
		cardRepo:    cardRepo,
		columnRepo:  columnRepo,
		boardRepo:   boardRepo,
		tagRepo:     tagRepo,
		cardTagRepo: cardTagRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "card.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "card"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

func (s *service) CreateCard(ctx context.Context, input CreateCardInput) (*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateCard")
	span.SetAttributes(
		attribute.String("card.column_id", input.ColumnID.String()),
		attribute.String("card.title", input.Title),
	)
	defer span.End()

	// Get the column to find the board ID
	col, err := s.columnRepo.GetByID(ctx, input.ColumnID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrColumnNotFound
		}
		return nil, err
	}

	// Get max position in column
	maxPos, err := s.cardRepo.GetMaxPosition(ctx, input.ColumnID)
	if err != nil {
		return nil, err
	}

	c := &card.Card{
		ColumnID:    input.ColumnID,
		BoardID:     col.BoardID,
		Title:       input.Title,
		Description: input.Description,
		Position:    maxPos + 1000, // Start at 1000 intervals
		Priority:    input.Priority,
		AssigneeID:  input.AssigneeID,
		DueDate:     input.DueDate,
		CreatedBy:   input.CreatedBy,
	}

	if c.Priority == "" {
		c.Priority = card.PriorityNone
	}

	if err := s.cardRepo.Create(ctx, c); err != nil {
		return nil, err
	}

	// Add tags if provided
	if len(input.TagIDs) > 0 {
		if err := s.cardTagRepo.SetTagsForCard(ctx, c.ID, input.TagIDs); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (s *service) GetCard(ctx context.Context, id uuid.UUID) (*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "GetCard")
	span.SetAttributes(attribute.String("card.id", id.String()))
	defer span.End()

	c, err := s.cardRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCardNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *service) GetCardsByColumnID(ctx context.Context, columnID uuid.UUID) ([]*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "GetCardsByColumnID")
	span.SetAttributes(attribute.String("card.column_id", columnID.String()))
	defer span.End()

	return s.cardRepo.GetByColumnID(ctx, columnID)
}

func (s *service) GetCardsByBoardID(ctx context.Context, boardID uuid.UUID) ([]*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "GetCardsByBoardID")
	span.SetAttributes(attribute.String("card.board_id", boardID.String()))
	defer span.End()

	return s.cardRepo.GetByBoardID(ctx, boardID)
}

func (s *service) GetCardsByAssigneeID(ctx context.Context, assigneeID uuid.UUID) ([]*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "GetCardsByAssigneeID")
	span.SetAttributes(attribute.String("card.assignee_id", assigneeID.String()))
	defer span.End()

	return s.cardRepo.GetByAssigneeID(ctx, assigneeID)
}

func (s *service) UpdateCard(ctx context.Context, input UpdateCardInput) (*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "UpdateCard")
	span.SetAttributes(attribute.String("card.id", input.ID.String()))
	defer span.End()

	c, err := s.cardRepo.GetByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCardNotFound
		}
		return nil, err
	}

	if input.Title != nil {
		c.Title = *input.Title
	}
	if input.Description != nil {
		c.Description = *input.Description
	}
	if input.Priority != nil {
		c.Priority = *input.Priority
	}
	if input.AssigneeID != nil {
		c.AssigneeID = input.AssigneeID
	}
	if input.ClearDueDate {
		c.DueDate = nil
	} else if input.DueDate != nil {
		c.DueDate = input.DueDate
	}

	if err := s.cardRepo.Update(ctx, c); err != nil {
		return nil, err
	}

	// Update tags if provided
	if input.TagIDs != nil {
		if err := s.cardTagRepo.SetTagsForCard(ctx, c.ID, input.TagIDs); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (s *service) MoveCard(ctx context.Context, cardID, targetColumnID uuid.UUID, afterCardID *uuid.UUID) (*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "MoveCard")
	span.SetAttributes(
		attribute.String("card.id", cardID.String()),
		attribute.String("card.target_column_id", targetColumnID.String()),
	)
	defer span.End()

	c, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCardNotFound
		}
		return nil, err
	}

	// Verify target column exists and get its board ID
	col, err := s.columnRepo.GetByID(ctx, targetColumnID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrColumnNotFound
		}
		return nil, err
	}

	// Calculate new position
	newPos, err := s.cardRepo.GetPositionBetween(ctx, targetColumnID, afterCardID)
	if err != nil {
		return nil, err
	}

	c.ColumnID = targetColumnID
	c.BoardID = col.BoardID
	c.Position = newPos

	if err := s.cardRepo.Update(ctx, c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *service) DeleteCard(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteCard")
	span.SetAttributes(attribute.String("card.id", id.String()))
	defer span.End()

	return s.cardRepo.Delete(ctx, id)
}

func (s *service) GetTagsForCard(ctx context.Context, cardID uuid.UUID) ([]*tag.Tag, error) {
	ctx, span := s.startServiceSpan(ctx, "GetTagsForCard")
	span.SetAttributes(attribute.String("card.id", cardID.String()))
	defer span.End()

	cardTags, err := s.cardTagRepo.GetByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if len(cardTags) == 0 {
		return []*tag.Tag{}, nil
	}

	tagIDs := make([]uuid.UUID, len(cardTags))
	for i, ct := range cardTags {
		tagIDs[i] = ct.TagID
	}

	return s.tagRepo.GetByIDs(ctx, tagIDs)
}

func (s *service) GetBoardByCardID(ctx context.Context, cardID uuid.UUID) (*board.Board, error) {
	ctx, span := s.startServiceSpan(ctx, "GetBoardByCardID")
	span.SetAttributes(attribute.String("card.id", cardID.String()))
	defer span.End()

	c, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCardNotFound
		}
		return nil, err
	}

	b, err := s.boardRepo.GetByID(ctx, c.BoardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, err
	}

	return b, nil
}

func (s *service) GetColumnByCardID(ctx context.Context, cardID uuid.UUID) (*board_column.BoardColumn, error) {
	ctx, span := s.startServiceSpan(ctx, "GetColumnByCardID")
	span.SetAttributes(attribute.String("card.id", cardID.String()))
	defer span.End()

	c, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCardNotFound
		}
		return nil, err
	}

	col, err := s.columnRepo.GetByID(ctx, c.ColumnID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrColumnNotFound
		}
		return nil, err
	}

	return col, nil
}
