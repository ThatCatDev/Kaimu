package card

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/board"
	boardMocks "github.com/thatcatdev/pulse-backend/internal/db/repositories/board/mocks"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/board_column"
	columnMocks "github.com/thatcatdev/pulse-backend/internal/db/repositories/board_column/mocks"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/card"
	cardMocks "github.com/thatcatdev/pulse-backend/internal/db/repositories/card/mocks"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/card_label"
	cardLabelMocks "github.com/thatcatdev/pulse-backend/internal/db/repositories/card_label/mocks"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/label"
	labelMocks "github.com/thatcatdev/pulse-backend/internal/db/repositories/label/mocks"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreateCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	columnID := uuid.New()
	boardID := uuid.New()
	userID := uuid.New()

	t.Run("success without labels", func(t *testing.T) {
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(&board_column.BoardColumn{ID: columnID, BoardID: boardID}, nil)

		mockCardRepo.EXPECT().
			GetMaxPosition(gomock.Any(), columnID).
			Return(float64(2000), nil)

		mockCardRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				c.ID = uuid.New()
				assert.Equal(t, columnID, c.ColumnID)
				assert.Equal(t, boardID, c.BoardID)
				assert.Equal(t, "Test Card", c.Title)
				assert.Equal(t, float64(3000), c.Position) // 2000 + 1000
				return nil
			})

		input := CreateCardInput{
			ColumnID:  columnID,
			Title:     "Test Card",
			Priority:  card.PriorityMedium,
			CreatedBy: &userID,
		}

		result, err := svc.CreateCard(ctx, input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Card", result.Title)
	})

	t.Run("success with labels", func(t *testing.T) {
		labelID1 := uuid.New()
		labelID2 := uuid.New()

		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(&board_column.BoardColumn{ID: columnID, BoardID: boardID}, nil)

		mockCardRepo.EXPECT().
			GetMaxPosition(gomock.Any(), columnID).
			Return(float64(0), nil)

		mockCardRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				c.ID = uuid.New()
				return nil
			})

		mockCardLabelRepo.EXPECT().
			SetLabelsForCard(gomock.Any(), gomock.Any(), []uuid.UUID{labelID1, labelID2}).
			Return(nil)

		input := CreateCardInput{
			ColumnID:  columnID,
			Title:     "Card with Labels",
			LabelIDs:  []uuid.UUID{labelID1, labelID2},
			CreatedBy: &userID,
		}

		result, err := svc.CreateCard(ctx, input)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("column not found", func(t *testing.T) {
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(nil, gorm.ErrRecordNotFound)

		input := CreateCardInput{
			ColumnID: columnID,
			Title:    "Test Card",
		}

		result, err := svc.CreateCard(ctx, input)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrColumnNotFound)
	})
}

func TestGetCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	cardID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := &card.Card{
			ID:    cardID,
			Title: "Test Card",
		}
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(expected, nil)

		result, err := svc.GetCard(ctx, cardID)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetCard(ctx, cardID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrCardNotFound)
	})
}

func TestGetCardsByColumnID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	columnID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := []*card.Card{
			{ID: uuid.New(), Title: "Card 1"},
			{ID: uuid.New(), Title: "Card 2"},
		}
		mockCardRepo.EXPECT().
			GetByColumnID(gomock.Any(), columnID).
			Return(expected, nil)

		result, err := svc.GetCardsByColumnID(ctx, columnID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestUpdateCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	cardID := uuid.New()

	t.Run("success - update title and priority", func(t *testing.T) {
		existingCard := &card.Card{
			ID:       cardID,
			Title:    "Old Title",
			Priority: card.PriorityLow,
		}
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(existingCard, nil)

		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				assert.Equal(t, "New Title", c.Title)
				assert.Equal(t, card.PriorityHigh, c.Priority)
				return nil
			})

		newTitle := "New Title"
		newPriority := card.PriorityHigh
		input := UpdateCardInput{
			ID:       cardID,
			Title:    &newTitle,
			Priority: &newPriority,
		}

		result, err := svc.UpdateCard(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, "New Title", result.Title)
	})

	t.Run("success - update labels", func(t *testing.T) {
		labelID := uuid.New()
		existingCard := &card.Card{
			ID:    cardID,
			Title: "Test Card",
		}
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(existingCard, nil)

		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		mockCardLabelRepo.EXPECT().
			SetLabelsForCard(gomock.Any(), cardID, []uuid.UUID{labelID}).
			Return(nil)

		input := UpdateCardInput{
			ID:       cardID,
			LabelIDs: []uuid.UUID{labelID},
		}

		result, err := svc.UpdateCard(ctx, input)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("card not found", func(t *testing.T) {
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(nil, gorm.ErrRecordNotFound)

		input := UpdateCardInput{ID: cardID}
		result, err := svc.UpdateCard(ctx, input)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrCardNotFound)
	})
}

func TestMoveCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	cardID := uuid.New()
	sourceColumnID := uuid.New()
	targetColumnID := uuid.New()
	boardID := uuid.New()

	t.Run("success - move to empty column", func(t *testing.T) {
		existingCard := &card.Card{
			ID:       cardID,
			ColumnID: sourceColumnID,
			BoardID:  boardID,
			Position: 1000,
		}
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(existingCard, nil)

		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), targetColumnID).
			Return(&board_column.BoardColumn{ID: targetColumnID, BoardID: boardID}, nil)

		mockCardRepo.EXPECT().
			GetPositionBetween(gomock.Any(), targetColumnID, (*uuid.UUID)(nil)).
			Return(float64(500), nil)

		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				assert.Equal(t, targetColumnID, c.ColumnID)
				assert.Equal(t, float64(500), c.Position)
				return nil
			})

		result, err := svc.MoveCard(ctx, cardID, targetColumnID, nil)
		require.NoError(t, err)
		assert.Equal(t, targetColumnID, result.ColumnID)
	})

	t.Run("success - move after another card", func(t *testing.T) {
		afterCardID := uuid.New()
		existingCard := &card.Card{
			ID:       cardID,
			ColumnID: sourceColumnID,
			BoardID:  boardID,
		}
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(existingCard, nil)

		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), targetColumnID).
			Return(&board_column.BoardColumn{ID: targetColumnID, BoardID: boardID}, nil)

		mockCardRepo.EXPECT().
			GetPositionBetween(gomock.Any(), targetColumnID, &afterCardID).
			Return(float64(1500), nil) // Between 1000 and 2000

		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				assert.Equal(t, float64(1500), c.Position)
				return nil
			})

		result, err := svc.MoveCard(ctx, cardID, targetColumnID, &afterCardID)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("card not found", func(t *testing.T) {
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.MoveCard(ctx, cardID, targetColumnID, nil)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrCardNotFound)
	})

	t.Run("column not found", func(t *testing.T) {
		existingCard := &card.Card{ID: cardID}
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(existingCard, nil)

		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), targetColumnID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.MoveCard(ctx, cardID, targetColumnID, nil)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrColumnNotFound)
	})
}

func TestDeleteCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	cardID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockCardRepo.EXPECT().
			Delete(gomock.Any(), cardID).
			Return(nil)

		err := svc.DeleteCard(ctx, cardID)
		require.NoError(t, err)
	})
}

func TestGetLabelsForCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	cardID := uuid.New()
	labelID1 := uuid.New()
	labelID2 := uuid.New()

	t.Run("success with multiple labels", func(t *testing.T) {
		cardLabels := []*card_label.CardLabel{
			{CardID: cardID, LabelID: labelID1},
			{CardID: cardID, LabelID: labelID2},
		}
		mockCardLabelRepo.EXPECT().
			GetByCardID(gomock.Any(), cardID).
			Return(cardLabels, nil)

		mockLabelRepo.EXPECT().
			GetByIDs(gomock.Any(), []uuid.UUID{labelID1, labelID2}).
			Return([]*label.Label{
				{ID: labelID1, Name: "Bug", Color: "#EF4444"},
				{ID: labelID2, Name: "Feature", Color: "#10B981"},
			}, nil)

		result, err := svc.GetLabelsForCard(ctx, cardID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("success empty labels", func(t *testing.T) {
		mockCardLabelRepo.EXPECT().
			GetByCardID(gomock.Any(), cardID).
			Return([]*card_label.CardLabel{}, nil)

		result, err := svc.GetLabelsForCard(ctx, cardID)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestGetBoardByCardID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	cardID := uuid.New()
	boardID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(&card.Card{ID: cardID, BoardID: boardID}, nil)

		mockBoardRepo.EXPECT().
			GetByID(gomock.Any(), boardID).
			Return(&board.Board{ID: boardID, Name: "Test Board"}, nil)

		result, err := svc.GetBoardByCardID(ctx, cardID)
		require.NoError(t, err)
		assert.Equal(t, boardID, result.ID)
	})

	t.Run("card not found", func(t *testing.T) {
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetBoardByCardID(ctx, cardID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrCardNotFound)
	})
}

func TestGetColumnByCardID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	cardID := uuid.New()
	columnID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(&card.Card{ID: cardID, ColumnID: columnID}, nil)

		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), columnID).
			Return(&board_column.BoardColumn{ID: columnID, Name: "Todo"}, nil)

		result, err := svc.GetColumnByCardID(ctx, cardID)
		require.NoError(t, err)
		assert.Equal(t, columnID, result.ID)
	})

	t.Run("card not found", func(t *testing.T) {
		mockCardRepo.EXPECT().
			GetByID(gomock.Any(), cardID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetColumnByCardID(ctx, cardID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrCardNotFound)
	})
}

func TestGetCardsByAssigneeID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockCardLabelRepo := cardLabelMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockLabelRepo, mockCardLabelRepo)
	ctx := context.Background()

	assigneeID := uuid.New()

	t.Run("success", func(t *testing.T) {
		dueDate := time.Now().Add(24 * time.Hour)
		expected := []*card.Card{
			{ID: uuid.New(), Title: "My Card 1", AssigneeID: &assigneeID, DueDate: &dueDate},
			{ID: uuid.New(), Title: "My Card 2", AssigneeID: &assigneeID},
		}
		mockCardRepo.EXPECT().
			GetByAssigneeID(gomock.Any(), assigneeID).
			Return(expected, nil)

		result, err := svc.GetCardsByAssigneeID(ctx, assigneeID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}
