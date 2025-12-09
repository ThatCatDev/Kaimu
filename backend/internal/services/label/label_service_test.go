package label

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/label"
	labelMocks "github.com/thatcatdev/pulse-backend/internal/db/repositories/label/mocks"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	projectMocks "github.com/thatcatdev/pulse-backend/internal/db/repositories/project/mocks"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreateLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockLabelRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID}, nil)

		mockLabelRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Bug").
			Return(nil, gorm.ErrRecordNotFound)

		mockLabelRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, l *label.Label) error {
				l.ID = uuid.New()
				assert.Equal(t, projectID, l.ProjectID)
				assert.Equal(t, "Bug", l.Name)
				assert.Equal(t, "#EF4444", l.Color)
				assert.Equal(t, "Bug fixes", l.Description)
				return nil
			})

		result, err := svc.CreateLabel(ctx, projectID, "Bug", "#EF4444", "Bug fixes")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Bug", result.Name)
	})

	t.Run("success with default color", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID}, nil)

		mockLabelRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Feature").
			Return(nil, gorm.ErrRecordNotFound)

		mockLabelRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, l *label.Label) error {
				l.ID = uuid.New()
				assert.Equal(t, "#6B7280", l.Color) // Default color
				return nil
			})

		result, err := svc.CreateLabel(ctx, projectID, "Feature", "", "")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("project not found", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.CreateLabel(ctx, projectID, "Bug", "#EF4444", "")
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})

	t.Run("label name already taken", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID}, nil)

		mockLabelRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Bug").
			Return(&label.Label{ID: uuid.New(), Name: "Bug"}, nil)

		result, err := svc.CreateLabel(ctx, projectID, "Bug", "#EF4444", "")
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrLabelNameTaken)
	})
}

func TestGetLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockLabelRepo, mockProjectRepo)
	ctx := context.Background()

	labelID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := &label.Label{
			ID:    labelID,
			Name:  "Bug",
			Color: "#EF4444",
		}
		mockLabelRepo.EXPECT().
			GetByID(gomock.Any(), labelID).
			Return(expected, nil)

		result, err := svc.GetLabel(ctx, labelID)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockLabelRepo.EXPECT().
			GetByID(gomock.Any(), labelID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetLabel(ctx, labelID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrLabelNotFound)
	})
}

func TestGetLabelsByProjectID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockLabelRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := []*label.Label{
			{ID: uuid.New(), Name: "Bug", Color: "#EF4444"},
			{ID: uuid.New(), Name: "Feature", Color: "#10B981"},
			{ID: uuid.New(), Name: "Enhancement", Color: "#3B82F6"},
		}
		mockLabelRepo.EXPECT().
			GetByProjectID(gomock.Any(), projectID).
			Return(expected, nil)

		result, err := svc.GetLabelsByProjectID(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("empty result", func(t *testing.T) {
		mockLabelRepo.EXPECT().
			GetByProjectID(gomock.Any(), projectID).
			Return([]*label.Label{}, nil)

		result, err := svc.GetLabelsByProjectID(ctx, projectID)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestGetLabelsByIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockLabelRepo, mockProjectRepo)
	ctx := context.Background()

	labelID1 := uuid.New()
	labelID2 := uuid.New()

	t.Run("success", func(t *testing.T) {
		ids := []uuid.UUID{labelID1, labelID2}
		expected := []*label.Label{
			{ID: labelID1, Name: "Bug"},
			{ID: labelID2, Name: "Feature"},
		}
		mockLabelRepo.EXPECT().
			GetByIDs(gomock.Any(), ids).
			Return(expected, nil)

		result, err := svc.GetLabelsByIDs(ctx, ids)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestUpdateLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockLabelRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()
	labelID := uuid.New()

	t.Run("success - update name and color", func(t *testing.T) {
		l := &label.Label{
			ID:        labelID,
			ProjectID: projectID,
			Name:      "New Bug Name",
			Color:     "#FF0000",
		}

		mockLabelRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "New Bug Name").
			Return(nil, gorm.ErrRecordNotFound)

		mockLabelRepo.EXPECT().
			Update(gomock.Any(), l).
			Return(nil)

		result, err := svc.UpdateLabel(ctx, l)
		require.NoError(t, err)
		assert.Equal(t, "New Bug Name", result.Name)
	})

	t.Run("success - update same name (no conflict)", func(t *testing.T) {
		l := &label.Label{
			ID:        labelID,
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF0000",
		}

		// Same ID returned means it's the same label
		mockLabelRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Bug").
			Return(&label.Label{ID: labelID, Name: "Bug"}, nil)

		mockLabelRepo.EXPECT().
			Update(gomock.Any(), l).
			Return(nil)

		result, err := svc.UpdateLabel(ctx, l)
		require.NoError(t, err)
		assert.Equal(t, "Bug", result.Name)
	})

	t.Run("fail - name conflict with different label", func(t *testing.T) {
		otherLabelID := uuid.New()
		l := &label.Label{
			ID:        labelID,
			ProjectID: projectID,
			Name:      "Feature",
			Color:     "#FF0000",
		}

		mockLabelRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Feature").
			Return(&label.Label{ID: otherLabelID, Name: "Feature"}, nil)

		result, err := svc.UpdateLabel(ctx, l)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrLabelNameTaken)
	})
}

func TestDeleteLabel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockLabelRepo, mockProjectRepo)
	ctx := context.Background()

	labelID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockLabelRepo.EXPECT().
			Delete(gomock.Any(), labelID).
			Return(nil)

		err := svc.DeleteLabel(ctx, labelID)
		require.NoError(t, err)
	})
}

func TestGetProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLabelRepo := labelMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockLabelRepo, mockProjectRepo)
	ctx := context.Background()

	labelID := uuid.New()
	projectID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockLabelRepo.EXPECT().
			GetByID(gomock.Any(), labelID).
			Return(&label.Label{ID: labelID, ProjectID: projectID}, nil)

		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID, Name: "Test Project"}, nil)

		result, err := svc.GetProject(ctx, labelID)
		require.NoError(t, err)
		assert.Equal(t, projectID, result.ID)
	})

	t.Run("label not found", func(t *testing.T) {
		mockLabelRepo.EXPECT().
			GetByID(gomock.Any(), labelID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetProject(ctx, labelID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrLabelNotFound)
	})

	t.Run("project not found", func(t *testing.T) {
		mockLabelRepo.EXPECT().
			GetByID(gomock.Any(), labelID).
			Return(&label.Label{ID: labelID, ProjectID: projectID}, nil)

		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetProject(ctx, labelID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})
}
