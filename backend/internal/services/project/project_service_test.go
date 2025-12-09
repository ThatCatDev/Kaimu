package project

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/organization"
	orgMocks "github.com/thatcatdev/pulse-backend/internal/db/repositories/organization/mocks"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	projectMocks "github.com/thatcatdev/pulse-backend/internal/db/repositories/project/mocks"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreateProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	orgID := uuid.New()
	org := &organization.Organization{
		ID:   orgID,
		Name: "Test Org",
	}

	// Organization exists
	mockOrgRepo.EXPECT().GetByID(gomock.Any(), orgID).Return(org, nil)

	// Key is not taken
	mockProjectRepo.EXPECT().GetByKey(gomock.Any(), orgID, "TEST").Return(nil, gorm.ErrRecordNotFound)

	// Create project
	mockProjectRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, p *project.Project) error {
		p.ID = uuid.New()
		p.CreatedAt = time.Now()
		return nil
	})

	proj, err := svc.CreateProject(context.Background(), orgID, "Test Project", "test", "A test project")

	require.NoError(t, err)
	assert.NotNil(t, proj)
	assert.Equal(t, "Test Project", proj.Name)
	assert.Equal(t, "TEST", proj.Key) // Should be uppercase
	assert.Equal(t, orgID, proj.OrganizationID)
}

func TestCreateProject_KeyTaken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	orgID := uuid.New()
	org := &organization.Organization{
		ID:   orgID,
		Name: "Test Org",
	}
	existingProject := &project.Project{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "Existing",
		Key:            "TEST",
	}

	// Organization exists
	mockOrgRepo.EXPECT().GetByID(gomock.Any(), orgID).Return(org, nil)

	// Key is already taken
	mockProjectRepo.EXPECT().GetByKey(gomock.Any(), orgID, "TEST").Return(existingProject, nil)

	proj, err := svc.CreateProject(context.Background(), orgID, "Test Project", "TEST", "A test project")

	assert.Error(t, err)
	assert.Equal(t, ErrKeyTaken, err)
	assert.Nil(t, proj)
}

func TestCreateProject_OrgNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	orgID := uuid.New()

	// Organization doesn't exist
	mockOrgRepo.EXPECT().GetByID(gomock.Any(), orgID).Return(nil, gorm.ErrRecordNotFound)

	proj, err := svc.CreateProject(context.Background(), orgID, "Test Project", "TEST", "A test project")

	assert.Error(t, err)
	assert.Equal(t, ErrOrgNotFound, err)
	assert.Nil(t, proj)
}

func TestCreateProject_InvalidKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	orgID := uuid.New()

	tests := []struct {
		name string
		key  string
	}{
		{"too short", "A"},
		{"too long", "ABCDEFGHIJK"},
		{"contains numbers", "TEST123"},
		{"contains special chars", "TEST!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj, err := svc.CreateProject(context.Background(), orgID, "Test Project", tt.key, "A test project")

			assert.Error(t, err)
			assert.Equal(t, ErrInvalidKey, err)
			assert.Nil(t, proj)
		})
	}
}

func TestGetProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	projectID := uuid.New()
	expectedProject := &project.Project{
		ID:   projectID,
		Name: "Test Project",
		Key:  "TEST",
	}

	mockProjectRepo.EXPECT().GetByID(gomock.Any(), projectID).Return(expectedProject, nil)

	proj, err := svc.GetProject(context.Background(), projectID)

	require.NoError(t, err)
	assert.NotNil(t, proj)
	assert.Equal(t, projectID, proj.ID)
	assert.Equal(t, "Test Project", proj.Name)
}

func TestGetProject_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	projectID := uuid.New()

	mockProjectRepo.EXPECT().GetByID(gomock.Any(), projectID).Return(nil, gorm.ErrRecordNotFound)

	proj, err := svc.GetProject(context.Background(), projectID)

	assert.Error(t, err)
	assert.Equal(t, ErrProjectNotFound, err)
	assert.Nil(t, proj)
}

func TestGetProjectByKey_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	orgID := uuid.New()
	expectedProject := &project.Project{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           "Test Project",
		Key:            "TEST",
	}

	mockProjectRepo.EXPECT().GetByKey(gomock.Any(), orgID, "TEST").Return(expectedProject, nil)

	proj, err := svc.GetProjectByKey(context.Background(), orgID, "test") // lowercase should work

	require.NoError(t, err)
	assert.NotNil(t, proj)
	assert.Equal(t, "TEST", proj.Key)
}

func TestGetProjectByKey_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	orgID := uuid.New()

	mockProjectRepo.EXPECT().GetByKey(gomock.Any(), orgID, "NONEXIST").Return(nil, gorm.ErrRecordNotFound)

	proj, err := svc.GetProjectByKey(context.Background(), orgID, "NONEXIST")

	assert.Error(t, err)
	assert.Equal(t, ErrProjectNotFound, err)
	assert.Nil(t, proj)
}

func TestGetOrgProjects_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	orgID := uuid.New()
	expectedProjects := []*project.Project{
		{ID: uuid.New(), OrganizationID: orgID, Name: "Project 1", Key: "PRJ1"},
		{ID: uuid.New(), OrganizationID: orgID, Name: "Project 2", Key: "PRJ2"},
	}

	mockProjectRepo.EXPECT().GetByOrgID(gomock.Any(), orgID).Return(expectedProjects, nil)

	projects, err := svc.GetOrgProjects(context.Background(), orgID)

	require.NoError(t, err)
	assert.Len(t, projects, 2)
}

func TestGetOrgProjects_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	orgID := uuid.New()

	mockProjectRepo.EXPECT().GetByOrgID(gomock.Any(), orgID).Return([]*project.Project{}, nil)

	projects, err := svc.GetOrgProjects(context.Background(), orgID)

	require.NoError(t, err)
	assert.Empty(t, projects)
}

func TestUpdateProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	proj := &project.Project{
		ID:          uuid.New(),
		Name:        "Updated Project",
		Key:         "UPD",
		Description: "Updated description",
	}

	mockProjectRepo.EXPECT().Update(gomock.Any(), proj).Return(nil)

	updated, err := svc.UpdateProject(context.Background(), proj)

	require.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "Updated Project", updated.Name)
}

func TestDeleteProject_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	projectID := uuid.New()

	mockProjectRepo.EXPECT().Delete(gomock.Any(), projectID).Return(nil)

	err := svc.DeleteProject(context.Background(), projectID)

	require.NoError(t, err)
}

func TestGetOrganization_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	projectID := uuid.New()
	orgID := uuid.New()
	proj := &project.Project{
		ID:             projectID,
		OrganizationID: orgID,
		Name:           "Test Project",
	}
	org := &organization.Organization{
		ID:   orgID,
		Name: "Test Org",
	}

	mockProjectRepo.EXPECT().GetByID(gomock.Any(), projectID).Return(proj, nil)
	mockOrgRepo.EXPECT().GetByID(gomock.Any(), orgID).Return(org, nil)

	fetchedOrg, err := svc.GetOrganization(context.Background(), projectID)

	require.NoError(t, err)
	assert.NotNil(t, fetchedOrg)
	assert.Equal(t, orgID, fetchedOrg.ID)
	assert.Equal(t, "Test Org", fetchedOrg.Name)
}

func TestGetOrganization_ProjectNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	projectID := uuid.New()

	mockProjectRepo.EXPECT().GetByID(gomock.Any(), projectID).Return(nil, gorm.ErrRecordNotFound)

	org, err := svc.GetOrganization(context.Background(), projectID)

	assert.Error(t, err)
	assert.Equal(t, ErrProjectNotFound, err)
	assert.Nil(t, org)
}

func TestGetOrganization_OrgNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProjectRepo := projectMocks.NewMockRepository(ctrl)
	mockOrgRepo := orgMocks.NewMockRepository(ctrl)

	svc := NewService(mockProjectRepo, mockOrgRepo)

	projectID := uuid.New()
	orgID := uuid.New()
	proj := &project.Project{
		ID:             projectID,
		OrganizationID: orgID,
		Name:           "Test Project",
	}

	mockProjectRepo.EXPECT().GetByID(gomock.Any(), projectID).Return(proj, nil)
	mockOrgRepo.EXPECT().GetByID(gomock.Any(), orgID).Return(nil, gorm.ErrRecordNotFound)

	org, err := svc.GetOrganization(context.Background(), projectID)

	assert.Error(t, err)
	assert.Equal(t, ErrOrgNotFound, err)
	assert.Nil(t, org)
}

func TestValidateKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"valid 2 chars", "AB", false},
		{"valid 10 chars", "ABCDEFGHIJ", false},
		{"valid middle", "TEST", false},
		{"too short", "A", true},
		{"too long", "ABCDEFGHIJK", true},
		{"contains number", "TEST1", true},
		{"contains lowercase", "Test", true},
		{"contains space", "TE ST", true},
		{"contains hyphen", "TE-ST", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKey(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidKey, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
