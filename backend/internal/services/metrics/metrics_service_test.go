package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/audit"
	auditMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/audit/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column"
	columnMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/card"
	cardMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/card/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/metrics_history"
	metricsHistMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/metrics_history/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/sprint"
	sprintMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/sprint/mocks"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func setupMocks(t *testing.T) (*gomock.Controller, *sprintMocks.MockRepository, *cardMocks.MockRepository, *columnMocks.MockRepository, *metricsHistMocks.MockRepository, *auditMocks.MockRepository) {
	ctrl := gomock.NewController(t)
	return ctrl,
		sprintMocks.NewMockRepository(ctrl),
		cardMocks.NewMockRepository(ctrl),
		columnMocks.NewMockRepository(ctrl),
		metricsHistMocks.NewMockRepository(ctrl),
		auditMocks.NewMockRepository(ctrl)
}

func TestGetSprintStats(t *testing.T) {
	ctrl, mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo := setupMocks(t)
	defer ctrl.Finish()

	svc := NewService(mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo)
	ctx := context.Background()

	sprintID := uuid.New()
	boardID := uuid.New()
	todoColumnID := uuid.New()
	doneColumnID := uuid.New()

	now := time.Now()
	startDate := now.Add(-7 * 24 * time.Hour)
	endDate := now.Add(7 * 24 * time.Hour)

	t.Run("success - returns correct stats", func(t *testing.T) {
		storyPoints1 := 5
		storyPoints2 := 3
		storyPoints3 := 8

		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(&sprint.Sprint{
				ID:        sprintID,
				BoardID:   boardID,
				StartDate: &startDate,
				EndDate:   &endDate,
			}, nil)

		mockCardRepo.EXPECT().
			GetBySprintID(gomock.Any(), sprintID).
			Return([]*card.Card{
				{ID: uuid.New(), ColumnID: todoColumnID, StoryPoints: &storyPoints1},
				{ID: uuid.New(), ColumnID: todoColumnID, StoryPoints: &storyPoints2},
				{ID: uuid.New(), ColumnID: doneColumnID, StoryPoints: &storyPoints3},
			}, nil)

		mockColumnRepo.EXPECT().
			GetByBoardID(gomock.Any(), boardID).
			Return([]*board_column.BoardColumn{
				{ID: todoColumnID, Name: "Todo", IsDone: false},
				{ID: doneColumnID, Name: "Done", IsDone: true},
			}, nil)

		stats, err := svc.GetSprintStats(ctx, sprintID)
		require.NoError(t, err)
		assert.Equal(t, 3, stats.TotalCards)
		assert.Equal(t, 1, stats.CompletedCards)
		assert.Equal(t, 16, stats.TotalStoryPoints) // 5 + 3 + 8
		assert.Equal(t, 8, stats.CompletedStoryPoints)
		// Days elapsed/remaining can vary by 1 due to time calculation, so use range check
		assert.True(t, stats.DaysElapsed >= 6 && stats.DaysElapsed <= 8, "DaysElapsed should be ~7")
		assert.True(t, stats.DaysRemaining >= 6 && stats.DaysRemaining <= 8, "DaysRemaining should be ~7")
	})

	t.Run("sprint not found", func(t *testing.T) {
		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(nil, gorm.ErrRecordNotFound)

		stats, err := svc.GetSprintStats(ctx, sprintID)
		assert.Nil(t, stats)
		assert.ErrorIs(t, err, ErrSprintNotFound)
	})

	t.Run("handles cards without story points", func(t *testing.T) {
		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(&sprint.Sprint{
				ID:        sprintID,
				BoardID:   boardID,
				StartDate: &startDate,
				EndDate:   &endDate,
			}, nil)

		mockCardRepo.EXPECT().
			GetBySprintID(gomock.Any(), sprintID).
			Return([]*card.Card{
				{ID: uuid.New(), ColumnID: todoColumnID, StoryPoints: nil},
				{ID: uuid.New(), ColumnID: doneColumnID, StoryPoints: nil},
			}, nil)

		mockColumnRepo.EXPECT().
			GetByBoardID(gomock.Any(), boardID).
			Return([]*board_column.BoardColumn{
				{ID: todoColumnID, Name: "Todo", IsDone: false},
				{ID: doneColumnID, Name: "Done", IsDone: true},
			}, nil)

		stats, err := svc.GetSprintStats(ctx, sprintID)
		require.NoError(t, err)
		assert.Equal(t, 2, stats.TotalCards)
		assert.Equal(t, 1, stats.CompletedCards)
		assert.Equal(t, 0, stats.TotalStoryPoints)
		assert.Equal(t, 0, stats.CompletedStoryPoints)
	})
}

func TestRecordDailySnapshot(t *testing.T) {
	ctrl, mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo := setupMocks(t)
	defer ctrl.Finish()

	svc := NewService(mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo)
	ctx := context.Background()

	sprintID := uuid.New()
	boardID := uuid.New()
	todoColumnID := uuid.New()
	doneColumnID := uuid.New()

	t.Run("success - creates snapshot", func(t *testing.T) {
		storyPoints1 := 5
		storyPoints2 := 8

		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(&sprint.Sprint{ID: sprintID, BoardID: boardID}, nil)

		mockCardRepo.EXPECT().
			GetBySprintID(gomock.Any(), sprintID).
			Return([]*card.Card{
				{ID: uuid.New(), ColumnID: todoColumnID, StoryPoints: &storyPoints1},
				{ID: uuid.New(), ColumnID: doneColumnID, StoryPoints: &storyPoints2},
			}, nil)

		mockColumnRepo.EXPECT().
			GetByBoardID(gomock.Any(), boardID).
			Return([]*board_column.BoardColumn{
				{ID: todoColumnID, Name: "Todo", IsDone: false},
				{ID: doneColumnID, Name: "Done", IsDone: true},
			}, nil)

		mockMetricsHistRepo.EXPECT().
			Upsert(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, h *metrics_history.MetricsHistory) error {
				assert.Equal(t, sprintID, h.SprintID)
				assert.Equal(t, 2, h.TotalCards)
				assert.Equal(t, 1, h.CompletedCards)
				assert.Equal(t, 13, h.TotalStoryPoints)
				assert.Equal(t, 8, h.CompletedStoryPoints)
				return nil
			})

		history, err := svc.RecordDailySnapshot(ctx, sprintID)
		require.NoError(t, err)
		assert.NotNil(t, history)
	})

	t.Run("sprint not found", func(t *testing.T) {
		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(nil, gorm.ErrRecordNotFound)

		history, err := svc.RecordDailySnapshot(ctx, sprintID)
		assert.Nil(t, history)
		assert.ErrorIs(t, err, ErrSprintNotFound)
	})
}

func TestGetBurnDownData(t *testing.T) {
	ctrl, mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo := setupMocks(t)
	defer ctrl.Finish()

	svc := NewService(mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo)
	ctx := context.Background()

	sprintID := uuid.New()
	boardID := uuid.New()

	now := time.Now().Truncate(24 * time.Hour)
	startDate := now.Add(-7 * 24 * time.Hour)
	endDate := now.Add(7 * 24 * time.Hour)

	t.Run("success - card count mode with no audit events", func(t *testing.T) {
		theSprint := &sprint.Sprint{
			ID:        sprintID,
			Name:      "Sprint 1",
			BoardID:   boardID,
			StartDate: &startDate,
			EndDate:   &endDate,
		}

		colID := uuid.New()
		doneColID := uuid.New()
		sp := 5

		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(theSprint, nil)

		mockColumnRepo.EXPECT().
			GetByBoardID(gomock.Any(), boardID).
			Return([]*board_column.BoardColumn{
				{ID: colID, Name: "Todo", IsDone: false},
				{ID: doneColID, Name: "Done", IsDone: true},
			}, nil)

		mockCardRepo.EXPECT().
			GetBySprintID(gomock.Any(), sprintID).
			Return([]*card.Card{
				{ID: uuid.New(), ColumnID: colID, StoryPoints: &sp},
				{ID: uuid.New(), ColumnID: doneColID, StoryPoints: &sp},
			}, nil)

		mockAuditRepo.EXPECT().
			GetCardMovementsByBoardAndDateRange(gomock.Any(), boardID, startDate, gomock.Any()).
			Return([]*audit.AuditEvent{}, nil)

		data, err := svc.GetBurnDownData(ctx, sprintID, MetricModeCardCount)
		require.NoError(t, err)
		assert.Equal(t, sprintID, data.SprintID)
		assert.Equal(t, "Sprint 1", data.SprintName)
		assert.NotEmpty(t, data.IdealLine)
		assert.NotEmpty(t, data.ActualLine)
	})

	t.Run("sprint not found", func(t *testing.T) {
		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(nil, gorm.ErrRecordNotFound)

		data, err := svc.GetBurnDownData(ctx, sprintID, MetricModeCardCount)
		assert.Nil(t, data)
		assert.ErrorIs(t, err, ErrSprintNotFound)
	})
}

func TestGetBurnUpData(t *testing.T) {
	ctrl, mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo := setupMocks(t)
	defer ctrl.Finish()

	svc := NewService(mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo)
	ctx := context.Background()

	sprintID := uuid.New()
	boardID := uuid.New()

	now := time.Now().Truncate(24 * time.Hour)
	startDate := now.Add(-7 * 24 * time.Hour)
	endDate := now.Add(7 * 24 * time.Hour)

	t.Run("success - shows scope and done lines", func(t *testing.T) {
		theSprint := &sprint.Sprint{
			ID:        sprintID,
			Name:      "Sprint 1",
			BoardID:   boardID,
			StartDate: &startDate,
			EndDate:   &endDate,
		}

		colID := uuid.New()
		doneColID := uuid.New()

		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(theSprint, nil)

		mockColumnRepo.EXPECT().
			GetByBoardID(gomock.Any(), boardID).
			Return([]*board_column.BoardColumn{
				{ID: colID, Name: "Todo", IsDone: false},
				{ID: doneColID, Name: "Done", IsDone: true},
			}, nil)

		mockCardRepo.EXPECT().
			GetBySprintID(gomock.Any(), sprintID).
			Return([]*card.Card{
				{ID: uuid.New(), ColumnID: colID},
				{ID: uuid.New(), ColumnID: doneColID},
			}, nil)

		mockAuditRepo.EXPECT().
			GetCardMovementsByBoardAndDateRange(gomock.Any(), boardID, startDate, gomock.Any()).
			Return([]*audit.AuditEvent{}, nil)

		data, err := svc.GetBurnUpData(ctx, sprintID, MetricModeCardCount)
		require.NoError(t, err)
		assert.NotEmpty(t, data.ScopeLine)
		assert.NotEmpty(t, data.DoneLine)
		assert.Equal(t, sprintID, data.SprintID)
		assert.Equal(t, "Sprint 1", data.SprintName)
	})
}

func TestGetVelocityData(t *testing.T) {
	ctrl, mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo := setupMocks(t)
	defer ctrl.Finish()

	svc := NewService(mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo)
	ctx := context.Background()

	boardID := uuid.New()
	sprint1ID := uuid.New()
	sprint2ID := uuid.New()

	t.Run("success - returns velocity for closed sprints", func(t *testing.T) {
		mockSprintRepo.EXPECT().
			GetClosedByBoardIDPaginated(gomock.Any(), boardID, 10, 0).
			Return([]*sprint.Sprint{
				{ID: sprint1ID, BoardID: boardID, Name: "Sprint 1"},
				{ID: sprint2ID, BoardID: boardID, Name: "Sprint 2"},
			}, 2, nil)

		mockMetricsHistRepo.EXPECT().
			GetLatestBySprintID(gomock.Any(), sprint1ID).
			Return(&metrics_history.MetricsHistory{
				SprintID:             sprint1ID,
				CompletedCards:       8,
				CompletedStoryPoints: 24,
			}, nil)

		mockMetricsHistRepo.EXPECT().
			GetLatestBySprintID(gomock.Any(), sprint2ID).
			Return(&metrics_history.MetricsHistory{
				SprintID:             sprint2ID,
				CompletedCards:       10,
				CompletedStoryPoints: 30,
			}, nil)

		data, err := svc.GetVelocityData(ctx, boardID, 10, MetricModeCardCount)
		require.NoError(t, err)
		// Sprints are reversed to show oldest first
		assert.Equal(t, 2, len(data.Sprints))
		assert.Equal(t, "Sprint 2", data.Sprints[0].SprintName)
		assert.Equal(t, "Sprint 1", data.Sprints[1].SprintName)
		assert.Equal(t, 10, data.Sprints[0].CompletedCards)
		assert.Equal(t, 8, data.Sprints[1].CompletedCards)
	})

	t.Run("success - handles missing history", func(t *testing.T) {
		doneColumnID := uuid.New()
		storyPoints := 5

		mockSprintRepo.EXPECT().
			GetClosedByBoardIDPaginated(gomock.Any(), boardID, 10, 0).
			Return([]*sprint.Sprint{
				{ID: sprint1ID, BoardID: boardID, Name: "Sprint 1"},
			}, 1, nil)

		// No history exists - returns error
		mockMetricsHistRepo.EXPECT().
			GetLatestBySprintID(gomock.Any(), sprint1ID).
			Return(nil, gorm.ErrRecordNotFound)

		// Should fallback to calculating from cards
		mockCardRepo.EXPECT().
			GetBySprintID(gomock.Any(), sprint1ID).
			Return([]*card.Card{
				{ID: uuid.New(), ColumnID: doneColumnID, StoryPoints: &storyPoints},
			}, nil)

		mockColumnRepo.EXPECT().
			GetByBoardID(gomock.Any(), boardID).
			Return([]*board_column.BoardColumn{
				{ID: doneColumnID, Name: "Done", IsDone: true},
			}, nil)

		data, err := svc.GetVelocityData(ctx, boardID, 10, MetricModeCardCount)
		require.NoError(t, err)
		assert.Equal(t, 1, len(data.Sprints))
		assert.Equal(t, 1, data.Sprints[0].CompletedCards)
		assert.Equal(t, 5, data.Sprints[0].CompletedPoints)
	})
}

func TestGetCumulativeFlowData(t *testing.T) {
	ctrl, mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo := setupMocks(t)
	defer ctrl.Finish()

	svc := NewService(mockSprintRepo, mockCardRepo, mockColumnRepo, mockMetricsHistRepo, mockAuditRepo)
	ctx := context.Background()

	sprintID := uuid.New()
	boardID := uuid.New()
	todoColumnID := uuid.New()
	inProgressColumnID := uuid.New()
	doneColumnID := uuid.New()

	now := time.Now().Truncate(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)

	t.Run("success - returns column flow data", func(t *testing.T) {
		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(&sprint.Sprint{ID: sprintID, Name: "Sprint 1", BoardID: boardID}, nil)

		mockColumnRepo.EXPECT().
			GetByBoardID(gomock.Any(), boardID).
			Return([]*board_column.BoardColumn{
				{ID: todoColumnID, Name: "Todo", Color: "#EF4444", IsHidden: false},
				{ID: inProgressColumnID, Name: "In Progress", Color: "#3B82F6", IsHidden: false},
				{ID: doneColumnID, Name: "Done", Color: "#10B981", IsHidden: false},
			}, nil)

		// Create history with column snapshots
		history1 := &metrics_history.MetricsHistory{
			SprintID:     sprintID,
			RecordedDate: yesterday,
		}
		_ = history1.SetColumnSnapshot(map[string]metrics_history.ColumnSnapshotData{
			todoColumnID.String():       {Name: "Todo", CardCount: 5, StoryPoints: 20},
			inProgressColumnID.String(): {Name: "In Progress", CardCount: 2, StoryPoints: 8},
			doneColumnID.String():       {Name: "Done", CardCount: 0, StoryPoints: 0},
		})

		history2 := &metrics_history.MetricsHistory{
			SprintID:     sprintID,
			RecordedDate: now,
		}
		_ = history2.SetColumnSnapshot(map[string]metrics_history.ColumnSnapshotData{
			todoColumnID.String():       {Name: "Todo", CardCount: 3, StoryPoints: 12},
			inProgressColumnID.String(): {Name: "In Progress", CardCount: 2, StoryPoints: 8},
			doneColumnID.String():       {Name: "Done", CardCount: 2, StoryPoints: 8},
		})

		mockMetricsHistRepo.EXPECT().
			GetBySprintID(gomock.Any(), sprintID).
			Return([]*metrics_history.MetricsHistory{history1, history2}, nil)

		data, err := svc.GetCumulativeFlowData(ctx, sprintID, MetricModeCardCount)
		require.NoError(t, err)
		assert.Equal(t, sprintID, data.SprintID)
		assert.Equal(t, 3, len(data.Columns))
		assert.Equal(t, 2, len(data.Dates))

		// Check Todo column values
		todoCol := data.Columns[0]
		assert.Equal(t, "Todo", todoCol.ColumnName)
		assert.Equal(t, 5, todoCol.Values[0]) // Yesterday
		assert.Equal(t, 3, todoCol.Values[1]) // Today
	})

	t.Run("excludes hidden columns", func(t *testing.T) {
		mockSprintRepo.EXPECT().
			GetByID(gomock.Any(), sprintID).
			Return(&sprint.Sprint{ID: sprintID, Name: "Sprint 1", BoardID: boardID}, nil)

		mockColumnRepo.EXPECT().
			GetByBoardID(gomock.Any(), boardID).
			Return([]*board_column.BoardColumn{
				{ID: todoColumnID, Name: "Todo", IsHidden: false},
				{ID: inProgressColumnID, Name: "Hidden", IsHidden: true},
				{ID: doneColumnID, Name: "Done", IsHidden: false},
			}, nil)

		history := &metrics_history.MetricsHistory{SprintID: sprintID, RecordedDate: now}
		_ = history.SetColumnSnapshot(map[string]metrics_history.ColumnSnapshotData{})

		mockMetricsHistRepo.EXPECT().
			GetBySprintID(gomock.Any(), sprintID).
			Return([]*metrics_history.MetricsHistory{history}, nil)

		data, err := svc.GetCumulativeFlowData(ctx, sprintID, MetricModeCardCount)
		require.NoError(t, err)
		// Should only have 2 columns (Todo and Done, not Hidden)
		assert.Equal(t, 2, len(data.Columns))
	})
}

func TestGenerateDateRange(t *testing.T) {
	t.Run("generates correct date range", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

		dates := generateDateRange(start, end)
		assert.Equal(t, 5, len(dates))
		assert.Equal(t, start, dates[0])
		assert.Equal(t, end, dates[4])
	})

	t.Run("handles same day", func(t *testing.T) {
		date := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		dates := generateDateRange(date, date)
		assert.Equal(t, 1, len(dates))
	})
}
