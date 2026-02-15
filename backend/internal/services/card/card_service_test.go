package card

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/board"
	boardMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column"
	columnMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/card"
	cardMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/card/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/card_tag"
	cardTagMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/card_tag/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/tag"
	tagMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/tag/mocks"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreateCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
	ctx := context.Background()

	columnID := uuid.New()
	boardID := uuid.New()
	userID := uuid.New()

	t.Run("success without tags", func(t *testing.T) {
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

	t.Run("success with tags", func(t *testing.T) {
		tagID1 := uuid.New()
		tagID2 := uuid.New()

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

		mockCardTagRepo.EXPECT().
			SetTagsForCard(gomock.Any(), gomock.Any(), []uuid.UUID{tagID1, tagID2}).
			Return(nil)

		input := CreateCardInput{
			ColumnID:  columnID,
			Title:     "Card with Tags",
			TagIDs:    []uuid.UUID{tagID1, tagID2},
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
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
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
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
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
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
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

	t.Run("success - update tags", func(t *testing.T) {
		tagID := uuid.New()
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

		mockCardTagRepo.EXPECT().
			SetTagsForCard(gomock.Any(), cardID, []uuid.UUID{tagID}).
			Return(nil)

		input := UpdateCardInput{
			ID:     cardID,
			TagIDs: []uuid.UUID{tagID},
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
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
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
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
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

func TestGetTagsForCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
	ctx := context.Background()

	cardID := uuid.New()
	tagID1 := uuid.New()
	tagID2 := uuid.New()

	t.Run("success with multiple tags", func(t *testing.T) {
		cardTags := []*card_tag.CardTag{
			{CardID: cardID, TagID: tagID1},
			{CardID: cardID, TagID: tagID2},
		}
		mockCardTagRepo.EXPECT().
			GetByCardID(gomock.Any(), cardID).
			Return(cardTags, nil)

		mockTagRepo.EXPECT().
			GetByIDs(gomock.Any(), []uuid.UUID{tagID1, tagID2}).
			Return([]*tag.Tag{
				{ID: tagID1, Name: "Bug", Color: "#EF4444"},
				{ID: tagID2, Name: "Feature", Color: "#10B981"},
			}, nil)

		result, err := svc.GetTagsForCard(ctx, cardID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("success empty tags", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			GetByCardID(gomock.Any(), cardID).
			Return([]*card_tag.CardTag{}, nil)

		result, err := svc.GetTagsForCard(ctx, cardID)
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
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
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
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
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
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
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

func TestGetCardsByIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
	ctx := context.Background()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	boardID := uuid.New()
	columnID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1"},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2"},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), []uuid.UUID{cardID1, cardID2}).
			Return(expected, nil)

		result, err := svc.GetCardsByIDs(ctx, []uuid.UUID{cardID1, cardID2})
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, cardID1, result[0].ID)
		assert.Equal(t, cardID2, result[1].ID)
	})

	t.Run("repo error", func(t *testing.T) {
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), []uuid.UUID{cardID1}).
			Return(nil, gorm.ErrInvalidDB)

		result, err := svc.GetCardsByIDs(ctx, []uuid.UUID{cardID1})
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestBulkMoveToColumn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
	ctx := context.Background()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	boardID := uuid.New()
	sourceColumnID := uuid.New()
	targetColumnID := uuid.New()
	cardIDs := []uuid.UUID{cardID1, cardID2}

	t.Run("success", func(t *testing.T) {
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), targetColumnID).
			Return(&board_column.BoardColumn{ID: targetColumnID, BoardID: boardID}, nil)

		mockCardRepo.EXPECT().
			MoveToColumn(gomock.Any(), cardIDs, targetColumnID).
			Return(nil)

		updatedCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: targetColumnID, Title: "Card 1"},
			{ID: cardID2, BoardID: boardID, ColumnID: targetColumnID, Title: "Card 2"},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(updatedCards, nil)

		result, err := svc.BulkMoveToColumn(ctx, cardIDs, targetColumnID)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, targetColumnID, result[0].ColumnID)
		assert.Equal(t, targetColumnID, result[1].ColumnID)
	})

	t.Run("column not found", func(t *testing.T) {
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), targetColumnID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.BulkMoveToColumn(ctx, cardIDs, targetColumnID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrColumnNotFound)
	})

	t.Run("column lookup error", func(t *testing.T) {
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), targetColumnID).
			Return(nil, gorm.ErrInvalidDB)

		result, err := svc.BulkMoveToColumn(ctx, cardIDs, targetColumnID)
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrColumnNotFound)
	})

	t.Run("move repo error", func(t *testing.T) {
		mockColumnRepo.EXPECT().
			GetByID(gomock.Any(), targetColumnID).
			Return(&board_column.BoardColumn{ID: targetColumnID, BoardID: boardID}, nil)

		mockCardRepo.EXPECT().
			MoveToColumn(gomock.Any(), cardIDs, targetColumnID).
			Return(gorm.ErrInvalidDB)

		result, err := svc.BulkMoveToColumn(ctx, cardIDs, targetColumnID)
		assert.Nil(t, result)
		assert.Error(t, err)
	})

	_ = sourceColumnID // used for clarity in test naming
}

func TestBulkUpdateProperties(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
	ctx := context.Background()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	boardID := uuid.New()
	columnID := uuid.New()
	cardIDs := []uuid.UUID{cardID1, cardID2}

	t.Run("set priority", func(t *testing.T) {
		existingCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1", Priority: card.PriorityLow},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2", Priority: card.PriorityNone},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(existingCards, nil)

		newPriority := card.PriorityHigh
		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				assert.Equal(t, newPriority, c.Priority)
				return nil
			}).Times(2)

		updatedCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1", Priority: card.PriorityHigh},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2", Priority: card.PriorityHigh},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(updatedCards, nil)

		input := BulkUpdatePropertiesInput{
			CardIDs:  cardIDs,
			Priority: &newPriority,
		}

		result, err := svc.BulkUpdateProperties(ctx, input)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, card.PriorityHigh, result[0].Priority)
		assert.Equal(t, card.PriorityHigh, result[1].Priority)
	})

	t.Run("set assignee", func(t *testing.T) {
		assigneeID := uuid.New()
		existingCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1"},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2"},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(existingCards, nil)

		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				assert.NotNil(t, c.AssigneeID)
				assert.Equal(t, assigneeID, *c.AssigneeID)
				return nil
			}).Times(2)

		updatedCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1", AssigneeID: &assigneeID},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2", AssigneeID: &assigneeID},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(updatedCards, nil)

		input := BulkUpdatePropertiesInput{
			CardIDs:    cardIDs,
			AssigneeID: &assigneeID,
		}

		result, err := svc.BulkUpdateProperties(ctx, input)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, &assigneeID, result[0].AssigneeID)
		assert.Equal(t, &assigneeID, result[1].AssigneeID)
	})

	t.Run("clear assignee", func(t *testing.T) {
		assigneeID := uuid.New()
		existingCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1", AssigneeID: &assigneeID},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2", AssigneeID: &assigneeID},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(existingCards, nil)

		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				assert.Nil(t, c.AssigneeID)
				return nil
			}).Times(2)

		updatedCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1", AssigneeID: nil},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2", AssigneeID: nil},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(updatedCards, nil)

		input := BulkUpdatePropertiesInput{
			CardIDs:       cardIDs,
			ClearAssignee: true,
		}

		result, err := svc.BulkUpdateProperties(ctx, input)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Nil(t, result[0].AssigneeID)
		assert.Nil(t, result[1].AssigneeID)
	})

	t.Run("set story points", func(t *testing.T) {
		storyPoints := 5
		existingCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1"},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2"},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(existingCards, nil)

		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				assert.NotNil(t, c.StoryPoints)
				assert.Equal(t, 5, *c.StoryPoints)
				return nil
			}).Times(2)

		updatedCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1", StoryPoints: &storyPoints},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2", StoryPoints: &storyPoints},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(updatedCards, nil)

		input := BulkUpdatePropertiesInput{
			CardIDs:     cardIDs,
			StoryPoints: &storyPoints,
		}

		result, err := svc.BulkUpdateProperties(ctx, input)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, &storyPoints, result[0].StoryPoints)
		assert.Equal(t, &storyPoints, result[1].StoryPoints)
	})

	t.Run("clear story points", func(t *testing.T) {
		storyPoints := 3
		existingCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1", StoryPoints: &storyPoints},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2", StoryPoints: &storyPoints},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(existingCards, nil)

		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, c *card.Card) error {
				assert.Nil(t, c.StoryPoints)
				return nil
			}).Times(2)

		updatedCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1", StoryPoints: nil},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2", StoryPoints: nil},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(updatedCards, nil)

		input := BulkUpdatePropertiesInput{
			CardIDs:          cardIDs,
			ClearStoryPoints: true,
		}

		result, err := svc.BulkUpdateProperties(ctx, input)
		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Nil(t, result[0].StoryPoints)
		assert.Nil(t, result[1].StoryPoints)
	})

	t.Run("get cards error", func(t *testing.T) {
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(nil, gorm.ErrInvalidDB)

		input := BulkUpdatePropertiesInput{
			CardIDs: cardIDs,
		}

		result, err := svc.BulkUpdateProperties(ctx, input)
		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("update error", func(t *testing.T) {
		existingCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1"},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), []uuid.UUID{cardID1}).
			Return(existingCards, nil)

		newPriority := card.PriorityUrgent
		mockCardRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(gorm.ErrInvalidDB)

		input := BulkUpdatePropertiesInput{
			CardIDs:  []uuid.UUID{cardID1},
			Priority: &newPriority,
		}

		result, err := svc.BulkUpdateProperties(ctx, input)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestBulkAddTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
	ctx := context.Background()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	tagID1 := uuid.New()
	tagID2 := uuid.New()
	boardID := uuid.New()
	columnID := uuid.New()
	cardIDs := []uuid.UUID{cardID1, cardID2}
	tagIDs := []uuid.UUID{tagID1, tagID2}

	t.Run("success", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			AddTagsToCards(gomock.Any(), cardIDs, tagIDs).
			Return(nil)

		updatedCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1"},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2"},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(updatedCards, nil)

		result, err := svc.BulkAddTags(ctx, cardIDs, tagIDs)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("add tags error", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			AddTagsToCards(gomock.Any(), cardIDs, tagIDs).
			Return(gorm.ErrInvalidDB)

		result, err := svc.BulkAddTags(ctx, cardIDs, tagIDs)
		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("get cards after add error", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			AddTagsToCards(gomock.Any(), cardIDs, tagIDs).
			Return(nil)

		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(nil, gorm.ErrInvalidDB)

		result, err := svc.BulkAddTags(ctx, cardIDs, tagIDs)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestBulkRemoveTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
	ctx := context.Background()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	tagID1 := uuid.New()
	tagID2 := uuid.New()
	boardID := uuid.New()
	columnID := uuid.New()
	cardIDs := []uuid.UUID{cardID1, cardID2}
	tagIDs := []uuid.UUID{tagID1, tagID2}

	t.Run("success", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			RemoveTagsFromCards(gomock.Any(), cardIDs, tagIDs).
			Return(nil)

		updatedCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1"},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2"},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(updatedCards, nil)

		result, err := svc.BulkRemoveTags(ctx, cardIDs, tagIDs)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("remove tags error", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			RemoveTagsFromCards(gomock.Any(), cardIDs, tagIDs).
			Return(gorm.ErrInvalidDB)

		result, err := svc.BulkRemoveTags(ctx, cardIDs, tagIDs)
		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("get cards after remove error", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			RemoveTagsFromCards(gomock.Any(), cardIDs, tagIDs).
			Return(nil)

		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(nil, gorm.ErrInvalidDB)

		result, err := svc.BulkRemoveTags(ctx, cardIDs, tagIDs)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestBulkSetTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
	ctx := context.Background()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	tagID1 := uuid.New()
	tagID2 := uuid.New()
	boardID := uuid.New()
	columnID := uuid.New()
	cardIDs := []uuid.UUID{cardID1, cardID2}
	tagIDs := []uuid.UUID{tagID1, tagID2}

	t.Run("success", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			SetTagsForCards(gomock.Any(), cardIDs, tagIDs).
			Return(nil)

		updatedCards := []*card.Card{
			{ID: cardID1, BoardID: boardID, ColumnID: columnID, Title: "Card 1"},
			{ID: cardID2, BoardID: boardID, ColumnID: columnID, Title: "Card 2"},
		}
		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(updatedCards, nil)

		result, err := svc.BulkSetTags(ctx, cardIDs, tagIDs)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("set tags error", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			SetTagsForCards(gomock.Any(), cardIDs, tagIDs).
			Return(gorm.ErrInvalidDB)

		result, err := svc.BulkSetTags(ctx, cardIDs, tagIDs)
		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("get cards after set error", func(t *testing.T) {
		mockCardTagRepo.EXPECT().
			SetTagsForCards(gomock.Any(), cardIDs, tagIDs).
			Return(nil)

		mockCardRepo.EXPECT().
			GetByIDs(gomock.Any(), cardIDs).
			Return(nil, gorm.ErrInvalidDB)

		result, err := svc.BulkSetTags(ctx, cardIDs, tagIDs)
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestBulkDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCardRepo := cardMocks.NewMockRepository(ctrl)
	mockColumnRepo := columnMocks.NewMockRepository(ctrl)
	mockBoardRepo := boardMocks.NewMockRepository(ctrl)
	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockCardTagRepo := cardTagMocks.NewMockRepository(ctrl)

	svc := NewService(mockCardRepo, mockColumnRepo, mockBoardRepo, mockTagRepo, mockCardTagRepo)
	ctx := context.Background()

	cardID1 := uuid.New()
	cardID2 := uuid.New()
	cardIDs := []uuid.UUID{cardID1, cardID2}

	t.Run("success", func(t *testing.T) {
		mockCardRepo.EXPECT().
			DeleteMany(gomock.Any(), cardIDs).
			Return(nil)

		err := svc.BulkDelete(ctx, cardIDs)
		require.NoError(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		mockCardRepo.EXPECT().
			DeleteMany(gomock.Any(), cardIDs).
			Return(gorm.ErrInvalidDB)

		err := svc.BulkDelete(ctx, cardIDs)
		assert.Error(t, err)
	})
}
