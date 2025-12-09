package resolvers

import (
	"context"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/graph/model"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/board"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/board_column"
	boardService "github.com/thatcatdev/pulse-backend/internal/services/board"
	cardService "github.com/thatcatdev/pulse-backend/internal/services/card"
	orgService "github.com/thatcatdev/pulse-backend/internal/services/organization"
	projectService "github.com/thatcatdev/pulse-backend/internal/services/project"
)

// Board returns a board by ID
func Board(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, projSvc projectService.Service, id string) (*model.Board, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	boardID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	b, err := boardSvc.GetBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// Get project to check org membership
	proj, err := boardSvc.GetProject(ctx, boardID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	return boardToModel(b), nil
}

// Boards returns all boards for a project
func Boards(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, projSvc projectService.Service, projectID string) ([]*model.Board, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	projID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, err
	}

	// Check membership
	proj, err := projSvc.GetProject(ctx, projID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	boards, err := boardSvc.GetBoardsByProjectID(ctx, projID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Board, len(boards))
	for i, b := range boards {
		result[i] = boardToModel(b)
	}
	return result, nil
}

// CreateBoard creates a new board
func CreateBoard(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, projSvc projectService.Service, input model.CreateBoardInput) (*model.Board, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	projID, err := uuid.Parse(input.ProjectID)
	if err != nil {
		return nil, err
	}

	// Check membership
	proj, err := projSvc.GetProject(ctx, projID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	description := ""
	if input.Description != nil {
		description = *input.Description
	}

	b, err := boardSvc.CreateBoard(ctx, projID, input.Name, description, userID)
	if err != nil {
		return nil, err
	}

	return boardToModel(b), nil
}

// UpdateBoard updates a board
func UpdateBoard(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, input model.UpdateBoardInput) (*model.Board, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	boardID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	b, err := boardSvc.GetBoard(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// Check membership
	proj, err := boardSvc.GetProject(ctx, boardID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	if input.Name != nil {
		b.Name = *input.Name
	}
	if input.Description != nil {
		b.Description = *input.Description
	}

	updated, err := boardSvc.UpdateBoard(ctx, b)
	if err != nil {
		return nil, err
	}

	return boardToModel(updated), nil
}

// DeleteBoard deletes a board
func DeleteBoard(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	boardID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Check membership
	proj, err := boardSvc.GetProject(ctx, boardID)
	if err != nil {
		return false, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return false, err
	}
	if !isMember {
		return false, ErrUnauthorized
	}

	if err := boardSvc.DeleteBoard(ctx, boardID); err != nil {
		return false, err
	}

	return true, nil
}

// CreateColumn creates a new board column
func CreateColumn(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, input model.CreateColumnInput) (*model.BoardColumn, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	boardID, err := uuid.Parse(input.BoardID)
	if err != nil {
		return nil, err
	}

	// Check membership
	proj, err := boardSvc.GetProject(ctx, boardID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	isBacklog := false
	if input.IsBacklog != nil {
		isBacklog = *input.IsBacklog
	}

	col, err := boardSvc.CreateColumn(ctx, boardID, input.Name, isBacklog)
	if err != nil {
		return nil, err
	}

	return columnToModel(col), nil
}

// UpdateColumn updates a board column
func UpdateColumn(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, input model.UpdateColumnInput) (*model.BoardColumn, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	colID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, err
	}

	col, err := boardSvc.GetColumn(ctx, colID)
	if err != nil {
		return nil, err
	}

	// Check membership
	b, err := boardSvc.GetBoardByColumnID(ctx, colID)
	if err != nil {
		return nil, err
	}

	proj, err := boardSvc.GetProject(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	if input.Name != nil {
		col.Name = *input.Name
	}
	if input.Color != nil {
		col.Color = *input.Color
	}
	if input.WipLimit != nil {
		col.WipLimit = input.WipLimit
	}

	updated, err := boardSvc.UpdateColumn(ctx, col)
	if err != nil {
		return nil, err
	}

	return columnToModel(updated), nil
}

// ReorderColumns reorders columns in a board
func ReorderColumns(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, input model.ReorderColumnsInput) ([]*model.BoardColumn, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	boardID, err := uuid.Parse(input.BoardID)
	if err != nil {
		return nil, err
	}

	// Check membership
	proj, err := boardSvc.GetProject(ctx, boardID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	columnIDs := make([]uuid.UUID, len(input.ColumnIds))
	for i, id := range input.ColumnIds {
		colID, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		columnIDs[i] = colID
	}

	cols, err := boardSvc.ReorderColumns(ctx, boardID, columnIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*model.BoardColumn, len(cols))
	for i, col := range cols {
		result[i] = columnToModel(col)
	}
	return result, nil
}

// ToggleColumnVisibility toggles column visibility
func ToggleColumnVisibility(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, id string) (*model.BoardColumn, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return nil, ErrUnauthorized
	}

	colID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	// Check membership
	b, err := boardSvc.GetBoardByColumnID(ctx, colID)
	if err != nil {
		return nil, err
	}

	proj, err := boardSvc.GetProject(ctx, b.ID)
	if err != nil {
		return nil, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrUnauthorized
	}

	col, err := boardSvc.ToggleColumnVisibility(ctx, colID)
	if err != nil {
		return nil, err
	}

	return columnToModel(col), nil
}

// DeleteColumn deletes a column
func DeleteColumn(ctx context.Context, orgSvc orgService.Service, boardSvc boardService.Service, id string) (bool, error) {
	userID := middleware.GetUserIDFromContext(ctx)
	if userID == nil {
		return false, ErrUnauthorized
	}

	colID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	// Check membership
	b, err := boardSvc.GetBoardByColumnID(ctx, colID)
	if err != nil {
		return false, err
	}

	proj, err := boardSvc.GetProject(ctx, b.ID)
	if err != nil {
		return false, err
	}

	isMember, err := orgSvc.IsMember(ctx, proj.OrganizationID, *userID)
	if err != nil {
		return false, err
	}
	if !isMember {
		return false, ErrUnauthorized
	}

	if err := boardSvc.DeleteColumn(ctx, colID); err != nil {
		return false, err
	}

	return true, nil
}

// BoardProject resolves the project field of a Board
func BoardProject(ctx context.Context, boardSvc boardService.Service, orgSvc orgService.Service, b *model.Board) (*model.Project, error) {
	boardID, err := uuid.Parse(b.ID)
	if err != nil {
		return nil, err
	}

	proj, err := boardSvc.GetProject(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// Get the organization for the project
	org, err := orgSvc.GetOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return nil, err
	}

	return projectToModelWithOrg(proj, organizationToModel(org)), nil
}

// BoardColumns resolves the columns field of a Board
func BoardColumns(ctx context.Context, boardSvc boardService.Service, b *model.Board) ([]*model.BoardColumn, error) {
	boardID, err := uuid.Parse(b.ID)
	if err != nil {
		return nil, err
	}

	cols, err := boardSvc.GetColumnsByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.BoardColumn, len(cols))
	for i, col := range cols {
		result[i] = columnToModel(col)
	}
	return result, nil
}

// ColumnBoard resolves the board field of a BoardColumn
func ColumnBoard(ctx context.Context, boardSvc boardService.Service, col *model.BoardColumn) (*model.Board, error) {
	colID, err := uuid.Parse(col.ID)
	if err != nil {
		return nil, err
	}

	b, err := boardSvc.GetBoardByColumnID(ctx, colID)
	if err != nil {
		return nil, err
	}

	return boardToModel(b), nil
}

// ColumnCards resolves the cards field of a BoardColumn
func ColumnCards(ctx context.Context, cardSvc cardService.Service, col *model.BoardColumn) ([]*model.Card, error) {
	colID, err := uuid.Parse(col.ID)
	if err != nil {
		return nil, err
	}

	cards, err := cardSvc.GetCardsByColumnID(ctx, colID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Card, len(cards))
	for i, c := range cards {
		result[i] = cardToModel(c)
	}
	return result, nil
}

// ProjectBoards resolves the boards field of a Project
func ProjectBoards(ctx context.Context, boardSvc boardService.Service, proj *model.Project) ([]*model.Board, error) {
	projID, err := uuid.Parse(proj.ID)
	if err != nil {
		return nil, err
	}

	boards, err := boardSvc.GetBoardsByProjectID(ctx, projID)
	if err != nil {
		return nil, err
	}

	result := make([]*model.Board, len(boards))
	for i, b := range boards {
		result[i] = boardToModel(b)
	}
	return result, nil
}

// ProjectDefaultBoard resolves the defaultBoard field of a Project
func ProjectDefaultBoard(ctx context.Context, boardSvc boardService.Service, proj *model.Project) (*model.Board, error) {
	projID, err := uuid.Parse(proj.ID)
	if err != nil {
		return nil, err
	}

	b, err := boardSvc.GetDefaultBoard(ctx, projID)
	if err != nil {
		// If no default board exists, return nil instead of error
		if err == boardService.ErrBoardNotFound {
			return nil, nil
		}
		return nil, err
	}

	return boardToModel(b), nil
}

func boardToModel(b *board.Board) *model.Board {
	var description *string
	if b.Description != "" {
		description = &b.Description
	}
	return &model.Board{
		ID:          b.ID.String(),
		Name:        b.Name,
		Description: description,
		IsDefault:   b.IsDefault,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func columnToModel(col *board_column.BoardColumn) *model.BoardColumn {
	var color *string
	if col.Color != "" {
		color = &col.Color
	}
	return &model.BoardColumn{
		ID:        col.ID.String(),
		Name:      col.Name,
		Position:  col.Position,
		IsBacklog: col.IsBacklog,
		IsHidden:  col.IsHidden,
		Color:     color,
		WipLimit:  col.WipLimit,
		CreatedAt: col.CreatedAt,
		UpdatedAt: col.UpdatedAt,
	}
}
