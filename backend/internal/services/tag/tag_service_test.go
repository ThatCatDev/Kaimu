package tag

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/project"
	projectMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/project/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/tag"
	tagMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/tag/mocks"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreateTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockTagRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID}, nil)

		mockTagRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Bug").
			Return(nil, gorm.ErrRecordNotFound)

		mockTagRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, tg *tag.Tag) error {
				tg.ID = uuid.New()
				assert.Equal(t, projectID, tg.ProjectID)
				assert.Equal(t, "Bug", tg.Name)
				assert.Equal(t, "#EF4444", tg.Color)
				assert.Equal(t, "Bug fixes", tg.Description)
				return nil
			})

		result, err := svc.CreateTag(ctx, projectID, "Bug", "#EF4444", "Bug fixes")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Bug", result.Name)
	})

	t.Run("success with default color", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID}, nil)

		mockTagRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Feature").
			Return(nil, gorm.ErrRecordNotFound)

		mockTagRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, tg *tag.Tag) error {
				tg.ID = uuid.New()
				assert.Equal(t, "#6B7280", tg.Color) // Default color
				return nil
			})

		result, err := svc.CreateTag(ctx, projectID, "Feature", "", "")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("project not found", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.CreateTag(ctx, projectID, "Bug", "#EF4444", "")
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})

	t.Run("tag name already taken", func(t *testing.T) {
		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID}, nil)

		mockTagRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Bug").
			Return(&tag.Tag{ID: uuid.New(), Name: "Bug"}, nil)

		result, err := svc.CreateTag(ctx, projectID, "Bug", "#EF4444", "")
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTagNameTaken)
	})
}

func TestGetTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockTagRepo, mockProjectRepo)
	ctx := context.Background()

	tagID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := &tag.Tag{
			ID:    tagID,
			Name:  "Bug",
			Color: "#EF4444",
		}
		mockTagRepo.EXPECT().
			GetByID(gomock.Any(), tagID).
			Return(expected, nil)

		result, err := svc.GetTag(ctx, tagID)
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockTagRepo.EXPECT().
			GetByID(gomock.Any(), tagID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetTag(ctx, tagID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTagNotFound)
	})
}

func TestGetTagsByProjectID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockTagRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expected := []*tag.Tag{
			{ID: uuid.New(), Name: "Bug", Color: "#EF4444"},
			{ID: uuid.New(), Name: "Feature", Color: "#10B981"},
			{ID: uuid.New(), Name: "Enhancement", Color: "#3B82F6"},
		}
		mockTagRepo.EXPECT().
			GetByProjectID(gomock.Any(), projectID).
			Return(expected, nil)

		result, err := svc.GetTagsByProjectID(ctx, projectID)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("empty result", func(t *testing.T) {
		mockTagRepo.EXPECT().
			GetByProjectID(gomock.Any(), projectID).
			Return([]*tag.Tag{}, nil)

		result, err := svc.GetTagsByProjectID(ctx, projectID)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestGetTagsByIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockTagRepo, mockProjectRepo)
	ctx := context.Background()

	tagID1 := uuid.New()
	tagID2 := uuid.New()

	t.Run("success", func(t *testing.T) {
		ids := []uuid.UUID{tagID1, tagID2}
		expected := []*tag.Tag{
			{ID: tagID1, Name: "Bug"},
			{ID: tagID2, Name: "Feature"},
		}
		mockTagRepo.EXPECT().
			GetByIDs(gomock.Any(), ids).
			Return(expected, nil)

		result, err := svc.GetTagsByIDs(ctx, ids)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestUpdateTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockTagRepo, mockProjectRepo)
	ctx := context.Background()

	projectID := uuid.New()
	tagID := uuid.New()

	t.Run("success - update name and color", func(t *testing.T) {
		tg := &tag.Tag{
			ID:        tagID,
			ProjectID: projectID,
			Name:      "New Bug Name",
			Color:     "#FF0000",
		}

		mockTagRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "New Bug Name").
			Return(nil, gorm.ErrRecordNotFound)

		mockTagRepo.EXPECT().
			Update(gomock.Any(), tg).
			Return(nil)

		result, err := svc.UpdateTag(ctx, tg)
		require.NoError(t, err)
		assert.Equal(t, "New Bug Name", result.Name)
	})

	t.Run("success - update same name (no conflict)", func(t *testing.T) {
		tg := &tag.Tag{
			ID:        tagID,
			ProjectID: projectID,
			Name:      "Bug",
			Color:     "#FF0000",
		}

		// Same ID returned means it's the same tag
		mockTagRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Bug").
			Return(&tag.Tag{ID: tagID, Name: "Bug"}, nil)

		mockTagRepo.EXPECT().
			Update(gomock.Any(), tg).
			Return(nil)

		result, err := svc.UpdateTag(ctx, tg)
		require.NoError(t, err)
		assert.Equal(t, "Bug", result.Name)
	})

	t.Run("fail - name conflict with different tag", func(t *testing.T) {
		otherTagID := uuid.New()
		tg := &tag.Tag{
			ID:        tagID,
			ProjectID: projectID,
			Name:      "Feature",
			Color:     "#FF0000",
		}

		mockTagRepo.EXPECT().
			GetByName(gomock.Any(), projectID, "Feature").
			Return(&tag.Tag{ID: otherTagID, Name: "Feature"}, nil)

		result, err := svc.UpdateTag(ctx, tg)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTagNameTaken)
	})
}

func TestDeleteTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockTagRepo, mockProjectRepo)
	ctx := context.Background()

	tagID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockTagRepo.EXPECT().
			Delete(gomock.Any(), tagID).
			Return(nil)

		err := svc.DeleteTag(ctx, tagID)
		require.NoError(t, err)
	})
}

func TestGetProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTagRepo := tagMocks.NewMockRepository(ctrl)
	mockProjectRepo := projectMocks.NewMockRepository(ctrl)

	svc := NewService(mockTagRepo, mockProjectRepo)
	ctx := context.Background()

	tagID := uuid.New()
	projectID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockTagRepo.EXPECT().
			GetByID(gomock.Any(), tagID).
			Return(&tag.Tag{ID: tagID, ProjectID: projectID}, nil)

		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(&project.Project{ID: projectID, Name: "Test Project"}, nil)

		result, err := svc.GetProject(ctx, tagID)
		require.NoError(t, err)
		assert.Equal(t, projectID, result.ID)
	})

	t.Run("tag not found", func(t *testing.T) {
		mockTagRepo.EXPECT().
			GetByID(gomock.Any(), tagID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetProject(ctx, tagID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTagNotFound)
	})

	t.Run("project not found", func(t *testing.T) {
		mockTagRepo.EXPECT().
			GetByID(gomock.Any(), tagID).
			Return(&tag.Tag{ID: tagID, ProjectID: projectID}, nil)

		mockProjectRepo.EXPECT().
			GetByID(gomock.Any(), projectID).
			Return(nil, gorm.ErrRecordNotFound)

		result, err := svc.GetProject(ctx, tagID)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrProjectNotFound)
	})
}
