package sprint

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/board"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/card"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/sprint"
	"github.com/thatcatdev/kaimu/backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrSprintNotFound            = errors.New("sprint not found")
	ErrBoardNotFound             = errors.New("board not found")
	ErrActiveSprintExists        = errors.New("an active sprint already exists for this board")
	ErrSprintAlreadyActive       = errors.New("sprint is already active")
	ErrSprintAlreadyClosed       = errors.New("sprint is already closed")
	ErrCannotStartClosedSprint   = errors.New("cannot start a closed sprint")
	ErrCannotCloseInactiveSprint = errors.New("can only close an active sprint")
)

type UpdateSprintInput struct {
	Name      *string
	Goal      *string
	StartDate *time.Time
	EndDate   *time.Time
}

type Service interface {
	// Sprint CRUD operations
	CreateSprint(ctx context.Context, boardID uuid.UUID, name, goal string, startDate, endDate *time.Time, createdBy *uuid.UUID) (*sprint.Sprint, error)
	GetSprint(ctx context.Context, id uuid.UUID) (*sprint.Sprint, error)
	GetBoardSprints(ctx context.Context, boardID uuid.UUID) ([]*sprint.Sprint, error)
	GetActiveSprint(ctx context.Context, boardID uuid.UUID) (*sprint.Sprint, error)
	GetFutureSprints(ctx context.Context, boardID uuid.UUID) ([]*sprint.Sprint, error)
	GetClosedSprints(ctx context.Context, boardID uuid.UUID) ([]*sprint.Sprint, error)
	GetClosedSprintsPaginated(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]*sprint.Sprint, int, error)
	UpdateSprint(ctx context.Context, id uuid.UUID, input UpdateSprintInput) (*sprint.Sprint, error)
	DeleteSprint(ctx context.Context, id uuid.UUID) error

	// Sprint lifecycle operations
	StartSprint(ctx context.Context, id uuid.UUID) (*sprint.Sprint, error)
	CompleteSprint(ctx context.Context, id uuid.UUID, moveIncompleteToBacklog bool) (*sprint.Sprint, error)

	// Card-Sprint operations (many-to-many)
	GetSprintCards(ctx context.Context, sprintID uuid.UUID) ([]*card.Card, error)
	GetBacklogCards(ctx context.Context, boardID uuid.UUID) ([]*card.Card, error)
	GetCardByID(ctx context.Context, cardID uuid.UUID) (*card.Card, error)
	GetCardSprintIDs(ctx context.Context, cardID uuid.UUID) ([]uuid.UUID, error)
	AddCardToSprint(ctx context.Context, cardID, sprintID uuid.UUID) (*card.Card, error)
	RemoveCardFromSprint(ctx context.Context, cardID, sprintID uuid.UUID) (*card.Card, error)
	SetCardSprints(ctx context.Context, cardID uuid.UUID, sprintIDs []uuid.UUID) (*card.Card, error)
	MoveCardToBacklog(ctx context.Context, cardID uuid.UUID) (*card.Card, error)

	// Get board for sprint
	GetBoard(ctx context.Context, sprintID uuid.UUID) (*board.Board, error)
}

type service struct {
	sprintRepo sprint.Repository
	cardRepo   card.Repository
	boardRepo  board.Repository
}

func NewService(sprintRepo sprint.Repository, cardRepo card.Repository, boardRepo board.Repository) Service {
	return &service{
		sprintRepo: sprintRepo,
		cardRepo:   cardRepo,
		boardRepo:  boardRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "sprint.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "sprint"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

// Sprint CRUD operations

func (s *service) CreateSprint(ctx context.Context, boardID uuid.UUID, name, goal string, startDate, endDate *time.Time, createdBy *uuid.UUID) (*sprint.Sprint, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateSprint")
	span.SetAttributes(
		attribute.String("sprint.board_id", boardID.String()),
		attribute.String("sprint.name", name),
	)
	defer span.End()

	// Verify board exists
	_, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, err
	}

	// Get next position
	position, err := s.sprintRepo.GetNextPosition(ctx, boardID)
	if err != nil {
		return nil, err
	}

	sp := &sprint.Sprint{
		BoardID:   boardID,
		Name:      name,
		Goal:      goal,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    sprint.SprintStatusFuture,
		Position:  position,
		CreatedBy: createdBy,
	}

	if err := s.sprintRepo.Create(ctx, sp); err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *service) GetSprint(ctx context.Context, id uuid.UUID) (*sprint.Sprint, error) {
	ctx, span := s.startServiceSpan(ctx, "GetSprint")
	span.SetAttributes(attribute.String("sprint.id", id.String()))
	defer span.End()

	sp, err := s.sprintRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}
	return sp, nil
}

func (s *service) GetBoardSprints(ctx context.Context, boardID uuid.UUID) ([]*sprint.Sprint, error) {
	ctx, span := s.startServiceSpan(ctx, "GetBoardSprints")
	span.SetAttributes(attribute.String("sprint.board_id", boardID.String()))
	defer span.End()

	return s.sprintRepo.GetByBoardID(ctx, boardID)
}

func (s *service) GetActiveSprint(ctx context.Context, boardID uuid.UUID) (*sprint.Sprint, error) {
	ctx, span := s.startServiceSpan(ctx, "GetActiveSprint")
	span.SetAttributes(attribute.String("sprint.board_id", boardID.String()))
	defer span.End()

	sp, err := s.sprintRepo.GetActiveByBoardID(ctx, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No active sprint is valid
		}
		return nil, err
	}
	return sp, nil
}

func (s *service) GetFutureSprints(ctx context.Context, boardID uuid.UUID) ([]*sprint.Sprint, error) {
	ctx, span := s.startServiceSpan(ctx, "GetFutureSprints")
	span.SetAttributes(attribute.String("sprint.board_id", boardID.String()))
	defer span.End()

	return s.sprintRepo.GetFutureByBoardID(ctx, boardID)
}

func (s *service) GetClosedSprints(ctx context.Context, boardID uuid.UUID) ([]*sprint.Sprint, error) {
	ctx, span := s.startServiceSpan(ctx, "GetClosedSprints")
	span.SetAttributes(attribute.String("sprint.board_id", boardID.String()))
	defer span.End()

	return s.sprintRepo.GetClosedByBoardID(ctx, boardID)
}

func (s *service) GetClosedSprintsPaginated(ctx context.Context, boardID uuid.UUID, limit, offset int) ([]*sprint.Sprint, int, error) {
	ctx, span := s.startServiceSpan(ctx, "GetClosedSprintsPaginated")
	span.SetAttributes(
		attribute.String("sprint.board_id", boardID.String()),
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)
	defer span.End()

	return s.sprintRepo.GetClosedByBoardIDPaginated(ctx, boardID, limit, offset)
}

func (s *service) UpdateSprint(ctx context.Context, id uuid.UUID, input UpdateSprintInput) (*sprint.Sprint, error) {
	ctx, span := s.startServiceSpan(ctx, "UpdateSprint")
	span.SetAttributes(attribute.String("sprint.id", id.String()))
	defer span.End()

	sp, err := s.sprintRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	if input.Name != nil {
		sp.Name = *input.Name
	}
	if input.Goal != nil {
		sp.Goal = *input.Goal
	}
	if input.StartDate != nil {
		sp.StartDate = input.StartDate
	}
	if input.EndDate != nil {
		sp.EndDate = input.EndDate
	}

	if err := s.sprintRepo.Update(ctx, sp); err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *service) DeleteSprint(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteSprint")
	span.SetAttributes(attribute.String("sprint.id", id.String()))
	defer span.End()

	// Verify sprint exists
	sp, err := s.sprintRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSprintNotFound
		}
		return err
	}

	// Remove all card-sprint associations for this sprint
	// (cards will be removed from this sprint but may remain in other sprints)
	cards, err := s.cardRepo.GetBySprintID(ctx, id)
	if err != nil {
		return err
	}

	for _, c := range cards {
		if err := s.cardRepo.RemoveCardFromSprint(ctx, c.ID, id); err != nil {
			return err
		}
	}

	// Delete sprint
	return s.sprintRepo.Delete(ctx, sp.ID)
}

// Sprint lifecycle operations

func (s *service) StartSprint(ctx context.Context, id uuid.UUID) (*sprint.Sprint, error) {
	ctx, span := s.startServiceSpan(ctx, "StartSprint")
	span.SetAttributes(attribute.String("sprint.id", id.String()))
	defer span.End()

	sp, err := s.sprintRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	// Check if already active
	if sp.Status == sprint.SprintStatusActive {
		return nil, ErrSprintAlreadyActive
	}

	// Check if closed
	if sp.Status == sprint.SprintStatusClosed {
		return nil, ErrCannotStartClosedSprint
	}

	// Check if another sprint is already active in this board
	activeSprint, err := s.sprintRepo.GetActiveByBoardID(ctx, sp.BoardID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if activeSprint != nil {
		return nil, ErrActiveSprintExists
	}

	// Start the sprint
	sp.Status = sprint.SprintStatusActive
	if sp.StartDate == nil {
		now := time.Now()
		sp.StartDate = &now
	}

	if err := s.sprintRepo.Update(ctx, sp); err != nil {
		return nil, err
	}

	return sp, nil
}

func (s *service) CompleteSprint(ctx context.Context, id uuid.UUID, moveIncompleteToBacklog bool) (*sprint.Sprint, error) {
	ctx, span := s.startServiceSpan(ctx, "CompleteSprint")
	span.SetAttributes(attribute.String("sprint.id", id.String()))
	defer span.End()

	sp, err := s.sprintRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	// Check if already closed
	if sp.Status == sprint.SprintStatusClosed {
		return nil, ErrSprintAlreadyClosed
	}

	// Check if active (only active sprints can be closed)
	if sp.Status != sprint.SprintStatusActive {
		return nil, ErrCannotCloseInactiveSprint
	}

	// If moveIncompleteToBacklog is true, remove cards from this sprint
	// (they may still be in other sprints if they carried over)
	if moveIncompleteToBacklog {
		cards, err := s.cardRepo.GetBySprintID(ctx, id)
		if err != nil {
			return nil, err
		}

		for _, c := range cards {
			// Remove card from this sprint
			if err := s.cardRepo.RemoveCardFromSprint(ctx, c.ID, id); err != nil {
				return nil, err
			}
		}
	}

	// Close the sprint
	sp.Status = sprint.SprintStatusClosed
	if sp.EndDate == nil {
		now := time.Now()
		sp.EndDate = &now
	}

	if err := s.sprintRepo.Update(ctx, sp); err != nil {
		return nil, err
	}

	return sp, nil
}

// Card-Sprint operations

func (s *service) GetSprintCards(ctx context.Context, sprintID uuid.UUID) ([]*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "GetSprintCards")
	span.SetAttributes(attribute.String("sprint.id", sprintID.String()))
	defer span.End()

	return s.cardRepo.GetBySprintID(ctx, sprintID)
}

func (s *service) GetBacklogCards(ctx context.Context, boardID uuid.UUID) ([]*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "GetBacklogCards")
	span.SetAttributes(attribute.String("board.id", boardID.String()))
	defer span.End()

	return s.cardRepo.GetBacklogByBoardID(ctx, boardID)
}

func (s *service) GetCardSprintIDs(ctx context.Context, cardID uuid.UUID) ([]uuid.UUID, error) {
	ctx, span := s.startServiceSpan(ctx, "GetCardSprintIDs")
	span.SetAttributes(attribute.String("card.id", cardID.String()))
	defer span.End()

	return s.cardRepo.GetSprintIDsForCard(ctx, cardID)
}

func (s *service) AddCardToSprint(ctx context.Context, cardID, sprintID uuid.UUID) (*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "AddCardToSprint")
	span.SetAttributes(
		attribute.String("card.id", cardID.String()),
		attribute.String("sprint.id", sprintID.String()),
	)
	defer span.End()

	// Verify card exists
	c, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// Verify sprint exists
	_, err = s.sprintRepo.GetByID(ctx, sprintID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	// Add card to sprint (many-to-many)
	if err := s.cardRepo.AddCardToSprint(ctx, cardID, sprintID); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *service) RemoveCardFromSprint(ctx context.Context, cardID, sprintID uuid.UUID) (*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "RemoveCardFromSprint")
	span.SetAttributes(
		attribute.String("card.id", cardID.String()),
		attribute.String("sprint.id", sprintID.String()),
	)
	defer span.End()

	// Verify card exists
	c, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// Remove card from sprint
	if err := s.cardRepo.RemoveCardFromSprint(ctx, cardID, sprintID); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *service) SetCardSprints(ctx context.Context, cardID uuid.UUID, sprintIDs []uuid.UUID) (*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "SetCardSprints")
	span.SetAttributes(attribute.String("card.id", cardID.String()))
	defer span.End()

	// Verify card exists
	c, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// Verify all sprints exist
	for _, sprintID := range sprintIDs {
		_, err = s.sprintRepo.GetByID(ctx, sprintID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrSprintNotFound
			}
			return nil, err
		}
	}

	// Set card sprints (replaces all existing assignments)
	if err := s.cardRepo.SetCardSprints(ctx, cardID, sprintIDs); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *service) MoveCardToBacklog(ctx context.Context, cardID uuid.UUID) (*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "MoveCardToBacklog")
	span.SetAttributes(attribute.String("card.id", cardID.String()))
	defer span.End()

	// Verify card exists
	c, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// Remove card from all sprints
	if err := s.cardRepo.RemoveCardFromAllSprints(ctx, cardID); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *service) GetBoard(ctx context.Context, sprintID uuid.UUID) (*board.Board, error) {
	ctx, span := s.startServiceSpan(ctx, "GetBoard")
	span.SetAttributes(attribute.String("sprint.id", sprintID.String()))
	defer span.End()

	sp, err := s.sprintRepo.GetByID(ctx, sprintID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	b, err := s.boardRepo.GetByID(ctx, sp.BoardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, err
	}

	return b, nil
}

func (s *service) GetCardByID(ctx context.Context, cardID uuid.UUID) (*card.Card, error) {
	ctx, span := s.startServiceSpan(ctx, "GetCardByID")
	span.SetAttributes(attribute.String("card.id", cardID.String()))
	defer span.End()

	return s.cardRepo.GetByID(ctx, cardID)
}
