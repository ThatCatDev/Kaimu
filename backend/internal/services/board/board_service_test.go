package board

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/board"
	boardMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column"
	columnMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/project"
	projectMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/project/mocks"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreateBoard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID}, nil)

		mockBoardRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, b *board.Board) error {
				b.ID = uuid.New()
				assert.Equal(t, projectID, b.ProjectID)
				assert.Equal(t, "Test Board", b.Name)
				assert.Equal(t, "Test Description", b.Description)
				assert.False(t, b.IsDefault)
				return nil
			})

		// Expect 4 default columns to be created
		mockColumnRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Times(4).
			Return(nil)

		result, err := svc.CreateBoard(ctx, projectID, "Test Board", "Test Description", &userID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Board", result.Name)
	})

	t.Run("project not found", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.CreateBoard(ctx, projectID, "Test Board", "Test Description", &userID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})
}

func TestCreateDefaultBoard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockBoardRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, b *board.Board) error {
				b.ID = uuid.New()
				assert.Equal(t, projectID, b.ProjectID)
				assert.Equal(t, "Default Board", b.Name)
				assert.True(t, b.IsDefault)
				return nil
			})

		// Expect 4 default columns to be created
		mockColumnRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Times(4).
			Return(nil)

		result, err := svc.CreateDefaultBoard(ctx, projectID, &userID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Default Board", result.Name)
		assert.True(t, result.IsDefault)
	})
}

func TestGetBoard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	boardID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := &board.Board{
			ID:   boardID,
			Name: "Test Board",
		}
		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(expected, nil)

		result, err := svc.GetBoard(ctx, boardID)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetBoard(ctx, boardID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrBoardNotFound)
	})
}

func TestGetBoardsByProjectID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := []*board.Board{
			{ID: uuid.New(), Name: "Board 1"},
			{ID: uuid.New(), Name: "Board 2"},
		}
		mockBoardRepo.EXPECT().
			GetByProjectID(gomock.Any(), projectID).
			Return(expected, nil)

		result, err := svc.GetBoardsByProjectID(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestGetDefaultBoard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := &board.Board{
			ID:        uuid.New(),
			Name:      "Default Board",
			IsDefault: true,
		}
		mockBoardRepo.EXPECT().
			GetDefaultByProjectID(gomock.Any(), projectID).
			Return(expected, nil)

		result, err := svc.GetDefaultBoard(ctx, projectID)
		require.NoError(t, err)
		assert.True(t, result.IsDefault)
	})

	t.Run("not found", func(t *testing.T) {
		mockBoardRepo.EXPECT().
			GetDefaultByProjectID(gomock.Any(), projectID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetDefaultBoard(ctx, projectID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrBoardNotFound)
	})
}

func TestDeleteBoard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	t.Run("success - non-default board", func(t *testing.T) {
		boardID := uuid.New()
		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(&board.Board{ID: boardID, IsDefault: false}, nil)

		mockBoardRepo.EXPECT().
			Delete(gomock.Any(), boardID).
			Return(nil)

		err := svc.DeleteBoard(ctx, boardID)
		require.NoError(t, err)
	})

	t.Run("fail - cannot delete default board", func(t *testing.T) {
		boardID := uuid.New()
		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(&board.Board{ID: boardID, IsDefault: true}, nil)

		err := svc.DeleteBoard(ctx, boardID)
		assert.ErrorIs(t, err, ErrCannotDeleteDefault)
	})

	t.Run("fail - board not found", func(t *testing.T) {
		boardID := uuid.New()
		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(nil, gorm.ErrRecordNotFound)

		err := svc.DeleteBoard(ctx, boardID)
		assert.ErrorIs(t, err, ErrBoardNotFound)
	})
}

func TestCreateColumn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	boardID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(&board.Board{ID: boardID}, nil)

		mockColumnRepo.EXPECT().
			GetMaxPosition(gomock.Any(), boardID).
			Return(3, nil)

		mockColumnRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, col *board_column.BoardColumn) error {
				col.ID = uuid.New()
				assert.Equal(t, boardID, col.BoardID)
				assert.Equal(t, "New Column", col.Name)
				assert.Equal(t, 4, col.Position) // maxPos + 1
				return nil
			})

		result, err := svc.CreateColumn(ctx, boardID, "New Column", false)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("board not found", func(t *testing.T) {
		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.CreateColumn(ctx, boardID, "New Column", false)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrBoardNotFound)
	})
}

func TestGetColumn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	columnID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := &board_column.BoardColumn{
			ID:   columnID,
			Name: "Test Column",
		}
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(expected, nil)

		result, err := svc.GetColumn(ctx, columnID)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetColumn(ctx, columnID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrColumnNotFound)
	})
}

func TestToggleColumnVisibility(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	columnID := uuid.New()

	t.Run("toggle hidden to visible", func(t *testing.T) {
		col := &board_column.BoardColumn{
			ID:       columnID,
			Name:     "Test Column",
			IsHidden: true,
		}
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(col, nil)

		mockColumnRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *board_column.BoardColumn) error {
				assert.False(t, c.IsHidden)
				return nil
			})

		result, err := svc.ToggleColumnVisibility(ctx, columnID)
		require.NoError(t, err)
		assert.False(t, result.IsHidden)
	})

	t.Run("toggle visible to hidden", func(t *testing.T) {
		col := &board_column.BoardColumn{
			ID:       columnID,
			Name:     "Test Column",
			IsHidden: false,
		}
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(col, nil)

		mockColumnRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *board_column.BoardColumn) error {
				assert.True(t, c.IsHidden)
				return nil
			})

		result, err := svc.ToggleColumnVisibility(ctx, columnID)
		require.NoError(t, err)
		assert.True(t, result.IsHidden)
	})
}

func TestReorderColumns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	boardID := uuid.New()
	col1ID := uuid.New()
	col2ID := uuid.New()
	col3ID := uuid.New()

	t.Run("success", func(t *testing.T) {
		columnIDs := []uuid.UUID{col3ID, col1ID, col2ID}

		mockColumnRepo.EXPECT().
			UpdatePositions(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, cols []*board_column.BoardColumn) error {
				assert.Len(t, cols, 3)
				assert.Equal(t, col3ID, cols[0].ID)
				assert.Equal(t, 0, cols[0].Position)
				assert.Equal(t, col1ID, cols[1].ID)
				assert.Equal(t, 1, cols[1].Position)
				assert.Equal(t, col2ID, cols[2].ID)
				assert.Equal(t, 2, cols[2].Position)
				return nil
			})

		mockColumnRepo.EXPECT().
			GetByBoardID(gomock.Any(), boardID).
			Return([]*board_column.BoardColumn{
				{ID: col3ID, Position: 0},
				{ID: col1ID, Position: 1},
				{ID: col2ID, Position: 2},
			}, nil)

		result, err := svc.ReorderColumns(ctx, boardID, columnIDs)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})
}

func TestGetProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	boardID := uuid.New()
	projectID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(&board.Board{ID: boardID, ProjectID: projectID}, nil)

		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID, Name: "Test Project"}, nil)

		result, err := svc.GetProject(ctx, boardID)
		require.NoError(t, err)
		assert.Equal(t, projectID, result.ID)
	})

	t.Run("board not found", func(t *testing.T) {
		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetProject(ctx, boardID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrBoardNotFound)
	})
}

func TestGetBoardByColumnID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockBoardRepo, mockColumnRepo, mockProjectRepo)
	ctx := context.Background()

	columnID := uuid.New()
	boardID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(&board_column.BoardColumn{ID: columnID, BoardID: boardID}, nil)

		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(&board.Board{ID: boardID, Name: "Test Board"}, nil)

		result, err := svc.GetBoardByColumnID(ctx, columnID)
		require.NoError(t, err)
		assert.Equal(t, boardID, result.ID)
	})

	t.Run("column not found", func(t *testing.T) {
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetBoardByColumnID(ctx, columnID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrColumnNotFound)
	})
}
