package board

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/board"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/project"
	"github.com/thatcatdev/kaimu/backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrBoardNotFound       = errors.New("board not found")
	ErrColumnNotFound      = errors.New("column not found")
	ErrProjectNotFound     = errors.New("project not found")
	ErrCannotDeleteDefault = errors.New("cannot delete default board")
)

type Service interface {
	// Board operations
	CreateBoard(ctx context.Context, projectID uuid.UUID, name, description string, createdBy *uuid.UUID) (*board.Board, error)
	CreateDefaultBoard(ctx context.Context, projectID uuid.UUID, createdBy *uuid.UUID) (*board.Board, error)
	GetBoard(ctx context.Context, id uuid.UUID) (*board.Board, error)
	GetBoardsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*board.Board, error)
	GetDefaultBoard(ctx context.Context, projectID uuid.UUID) (*board.Board, error)
	UpdateBoard(ctx context.Context, b *board.Board) (*board.Board, error)
	DeleteBoard(ctx context.Context, id uuid.UUID) error
	GetProject(ctx context.Context, boardID uuid.UUID) (*project.Project, error)

	// Column operations
	CreateColumn(ctx context.Context, boardID uuid.UUID, name string, isBacklog bool) (*board_column.BoardColumn, error)
	GetColumn(ctx context.Context, id uuid.UUID) (*board_column.BoardColumn, error)
	GetColumnsByBoardID(ctx context.Context, boardID uuid.UUID) ([]*board_column.BoardColumn, error)
	GetVisibleColumns(ctx context.Context, boardID uuid.UUID) ([]*board_column.BoardColumn, error)
	UpdateColumn(ctx context.Context, col *board_column.BoardColumn) (*board_column.BoardColumn, error)
	ReorderColumns(ctx context.Context, boardID uuid.UUID, columnIDs []uuid.UUID) ([]*board_column.BoardColumn, error)
	ToggleColumnVisibility(ctx context.Context, id uuid.UUID) (*board_column.BoardColumn, error)
	DeleteColumn(ctx context.Context, id uuid.UUID) error
	GetBoardByColumnID(ctx context.Context, columnID uuid.UUID) (*board.Board, error)
}

type service struct {
	boardRepo   board.Repository
	columnRepo  board_column.Repository
	projectRepo project.Repository
}

func NewService(boardRepo board.Repository, columnRepo board_column.Repository, projectRepo project.Repository) Service {
	return &service{
		boardRepo:   boardRepo,
		columnRepo:  columnRepo,
		projectRepo: projectRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "board.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "board"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

// Board operations

func (s *service) CreateBoard(ctx context.Context, projectID uuid.UUID, name, description string, createdBy *uuid.UUID) (*board.Board, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateBoard")
	span.SetAttributes(
		attribute.String("board.project_id", projectID.String()),
		attribute.String("board.name", name),
	)
	defer span.End()

	// Verify project exists
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	b := &board.Board{
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		IsDefault:   false,
		CreatedBy:   createdBy,
	}

	if err := s.boardRepo.Create(ctx, b); err != nil {
		return nil, err
	}

	// Create default columns for this board
	if err := s.createDefaultColumns(ctx, b.ID); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *service) CreateDefaultBoard(ctx context.Context, projectID uuid.UUID, createdBy *uuid.UUID) (*board.Board, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateDefaultBoard")
	span.SetAttributes(attribute.String("board.project_id", projectID.String()))
	defer span.End()

	b := &board.Board{
		ProjectID:   projectID,
		Name:        "Main Board",
		Description: "Kanban board for tracking tasks",
		IsDefault:   false,
		CreatedBy:   createdBy,
	}

	if err := s.boardRepo.Create(ctx, b); err != nil {
		return nil, err
	}

	// Create default columns
	if err := s.createDefaultColumns(ctx, b.ID); err != nil {
		return nil, err
	}

	return b, nil
}

func (s *service) createDefaultColumns(ctx context.Context, boardID uuid.UUID) error {
	columns := []struct {
		Name      string
		Position  int
		IsBacklog bool
		IsHidden  bool
		Color     string
	}{
		{Name: "Backlog", Position: 0, IsBacklog: true, IsHidden: true, Color: "#6B7280"},
		{Name: "Todo", Position: 1, IsBacklog: false, IsHidden: false, Color: "#3B82F6"},
		{Name: "In Progress", Position: 2, IsBacklog: false, IsHidden: false, Color: "#F59E0B"},
		{Name: "Done", Position: 3, IsBacklog: false, IsHidden: false, Color: "#10B981"},
	}

	for _, col := range columns {
		c := &board_column.BoardColumn{
			BoardID:   boardID,
			Name:      col.Name,
			Position:  col.Position,
			IsBacklog: col.IsBacklog,
			IsHidden:  col.IsHidden,
			Color:     col.Color,
		}
		if err := s.columnRepo.Create(ctx, c); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) GetBoard(ctx context.Context, id uuid.UUID) (*board.Board, error) {
	ctx, span := s.startServiceSpan(ctx, "GetBoard")
	span.SetAttributes(attribute.String("board.id", id.String()))
	defer span.End()

	b, err := s.boardRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *service) GetBoardsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*board.Board, error) {
	ctx, span := s.startServiceSpan(ctx, "GetBoardsByProjectID")
	span.SetAttributes(attribute.String("board.project_id", projectID.String()))
	defer span.End()

	return s.boardRepo.GetByProjectID(ctx, projectID)
}

func (s *service) GetDefaultBoard(ctx context.Context, projectID uuid.UUID) (*board.Board, error) {
	ctx, span := s.startServiceSpan(ctx, "GetDefaultBoard")
	span.SetAttributes(attribute.String("board.project_id", projectID.String()))
	defer span.End()

	b, err := s.boardRepo.GetDefaultByProjectID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, err
	}
	return b, nil
}

func (s *service) UpdateBoard(ctx context.Context, b *board.Board) (*board.Board, error) {
	ctx, span := s.startServiceSpan(ctx, "UpdateBoard")
	span.SetAttributes(attribute.String("board.id", b.ID.String()))
	defer span.End()

	if err := s.boardRepo.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *service) DeleteBoard(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteBoard")
	span.SetAttributes(attribute.String("board.id", id.String()))
	defer span.End()

	// Verify board exists
	_, err := s.boardRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBoardNotFound
		}
		return err
	}

	return s.boardRepo.Delete(ctx, id)
}

func (s *service) GetProject(ctx context.Context, boardID uuid.UUID) (*project.Project, error) {
	ctx, span := s.startServiceSpan(ctx, "GetProject")
	span.SetAttributes(attribute.String("board.id", boardID.String()))
	defer span.End()

	b, err := s.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, err
	}

	proj, err := s.projectRepo.GetByID(ctx, b.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return proj, nil
}

// Column operations

func (s *service) CreateColumn(ctx context.Context, boardID uuid.UUID, name string, isBacklog bool) (*board_column.BoardColumn, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateColumn")
	span.SetAttributes(
		attribute.String("column.board_id", boardID.String()),
		attribute.String("column.name", name),
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

	// Get max position
	maxPos, err := s.columnRepo.GetMaxPosition(ctx, boardID)
	if err != nil {
		return nil, err
	}

	col := &board_column.BoardColumn{
		BoardID:   boardID,
		Name:      name,
		Position:  maxPos + 1,
		IsBacklog: isBacklog,
		IsHidden:  false,
		Color:     "#6B7280",
	}

	if err := s.columnRepo.Create(ctx, col); err != nil {
		return nil, err
	}

	return col, nil
}

func (s *service) GetColumn(ctx context.Context, id uuid.UUID) (*board_column.BoardColumn, error) {
	ctx, span := s.startServiceSpan(ctx, "GetColumn")
	span.SetAttributes(attribute.String("column.id", id.String()))
	defer span.End()

	col, err := s.columnRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrColumnNotFound
		}
		return nil, err
	}
	return col, nil
}

func (s *service) GetColumnsByBoardID(ctx context.Context, boardID uuid.UUID) ([]*board_column.BoardColumn, error) {
	ctx, span := s.startServiceSpan(ctx, "GetColumnsByBoardID")
	span.SetAttributes(attribute.String("column.board_id", boardID.String()))
	defer span.End()

	return s.columnRepo.GetByBoardID(ctx, boardID)
}

func (s *service) GetVisibleColumns(ctx context.Context, boardID uuid.UUID) ([]*board_column.BoardColumn, error) {
	ctx, span := s.startServiceSpan(ctx, "GetVisibleColumns")
	span.SetAttributes(attribute.String("column.board_id", boardID.String()))
	defer span.End()

	return s.columnRepo.GetVisibleByBoardID(ctx, boardID)
}

func (s *service) UpdateColumn(ctx context.Context, col *board_column.BoardColumn) (*board_column.BoardColumn, error) {
	ctx, span := s.startServiceSpan(ctx, "UpdateColumn")
	span.SetAttributes(attribute.String("column.id", col.ID.String()))
	defer span.End()

	if err := s.columnRepo.Update(ctx, col); err != nil {
		return nil, err
	}
	return col, nil
}

func (s *service) ReorderColumns(ctx context.Context, boardID uuid.UUID, columnIDs []uuid.UUID) ([]*board_column.BoardColumn, error) {
	ctx, span := s.startServiceSpan(ctx, "ReorderColumns")
	span.SetAttributes(attribute.String("column.board_id", boardID.String()))
	defer span.End()

	// Build update list
	columns := make([]*board_column.BoardColumn, len(columnIDs))
	for i, id := range columnIDs {
		columns[i] = &board_column.BoardColumn{
			ID:       id,
			Position: i,
		}
	}

	if err := s.columnRepo.UpdatePositions(ctx, columns); err != nil {
		return nil, err
	}

	// Return updated columns
	return s.columnRepo.GetByBoardID(ctx, boardID)
}

func (s *service) ToggleColumnVisibility(ctx context.Context, id uuid.UUID) (*board_column.BoardColumn, error) {
	ctx, span := s.startServiceSpan(ctx, "ToggleColumnVisibility")
	span.SetAttributes(attribute.String("column.id", id.String()))
	defer span.End()

	col, err := s.columnRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrColumnNotFound
		}
		return nil, err
	}

	col.IsHidden = !col.IsHidden

	if err := s.columnRepo.Update(ctx, col); err != nil {
		return nil, err
	}

	return col, nil
}

func (s *service) DeleteColumn(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteColumn")
	span.SetAttributes(attribute.String("column.id", id.String()))
	defer span.End()

	return s.columnRepo.Delete(ctx, id)
}

func (s *service) GetBoardByColumnID(ctx context.Context, columnID uuid.UUID) (*board.Board, error) {
	ctx, span := s.startServiceSpan(ctx, "GetBoardByColumnID")
	span.SetAttributes(attribute.String("column.id", columnID.String()))
	defer span.End()

	col, err := s.columnRepo.GetByID(ctx, columnID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrColumnNotFound
		}
		return nil, err
	}

	return s.boardRepo.GetByID(ctx, col.BoardID)
}
