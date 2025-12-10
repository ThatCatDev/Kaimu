package organization

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization"
	orgMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member"
	memberMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	userMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/user/mocks"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreateOrganization_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	userID := uuid.New()

	// Slug doesn't exist
	mockOrgRepo.EXPECT().GetBySlug(gomock.Any(), "test-org").Return(nil, gorm.ErrRecordNotFound)

	// Create org
	mockOrgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, org *organization.Organization) error {
		org.ID = uuid.New()
		org.CreatedAt = time.Now()
		return nil
	})

	// Add owner as member
	mockMemberRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	org, err := svc.CreateOrganization(context.Background(), userID, "Test Org", "A test organization")

	require.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, "Test Org", org.Name)
	assert.Equal(t, "test-org", org.Slug)
	assert.Equal(t, userID, org.OwnerID)
}

func TestCreateOrganization_SlugTaken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	userID := uuid.New()
	existingOrg := &organization.Organization{
		ID:   uuid.New(),
		Slug: "test-org",
	}

	// Slug exists - will generate unique slug
	mockOrgRepo.EXPECT().GetBySlug(gomock.Any(), "test-org").Return(existingOrg, nil)

	// Create org with unique slug
	mockOrgRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, org *organization.Organization) error {
		org.ID = uuid.New()
		org.CreatedAt = time.Now()
		// Slug should have been modified with suffix
		assert.Contains(t, org.Slug, "test-org-")
		return nil
	})

	// Add owner as member
	mockMemberRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	org, err := svc.CreateOrganization(context.Background(), userID, "Test Org", "A test organization")

	require.NoError(t, err)
	assert.NotNil(t, org)
	assert.Contains(t, org.Slug, "test-org-")
}

func TestGetOrganization_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()
	expectedOrg := &organization.Organization{
		ID:   orgID,
		Name: "Test Org",
		Slug: "test-org",
	}

	mockOrgRepo.EXPECT().GetByID(gomock.Any(), orgID).Return(expectedOrg, nil)

	org, err := svc.GetOrganization(context.Background(), orgID)

	require.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, orgID, org.ID)
	assert.Equal(t, "Test Org", org.Name)
}

func TestGetOrganization_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()

	mockOrgRepo.EXPECT().GetByID(gomock.Any(), orgID).Return(nil, gorm.ErrRecordNotFound)

	org, err := svc.GetOrganization(context.Background(), orgID)

	assert.Error(t, err)
	assert.Equal(t, ErrOrgNotFound, err)
	assert.Nil(t, org)
}

func TestGetOrganizationBySlug_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	expectedOrg := &organization.Organization{
		ID:   uuid.New(),
		Name: "Test Org",
		Slug: "test-org",
	}

	mockOrgRepo.EXPECT().GetBySlug(gomock.Any(), "test-org").Return(expectedOrg, nil)

	org, err := svc.GetOrganizationBySlug(context.Background(), "test-org")

	require.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, "test-org", org.Slug)
}

func TestGetOrganizationBySlug_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	mockOrgRepo.EXPECT().GetBySlug(gomock.Any(), "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	org, err := svc.GetOrganizationBySlug(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Equal(t, ErrOrgNotFound, err)
	assert.Nil(t, org)
}

func TestGetUserOrganizations_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	userID := uuid.New()
	expectedOrgs := []*organization.Organization{
		{ID: uuid.New(), Name: "Org 1", Slug: "org-1"},
		{ID: uuid.New(), Name: "Org 2", Slug: "org-2"},
	}

	mockOrgRepo.EXPECT().GetByUserID(gomock.Any(), userID).Return(expectedOrgs, nil)

	orgs, err := svc.GetUserOrganizations(context.Background(), userID)

	require.NoError(t, err)
	assert.Len(t, orgs, 2)
}

func TestAddMember_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()
	userID := uuid.New()

	// User is not already a member
	mockMemberRepo.EXPECT().GetByOrgAndUser(gomock.Any(), orgID, userID).Return(nil, gorm.ErrRecordNotFound)

	// Create membership
	mockMemberRepo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, m *organization_member.OrganizationMember) error {
		m.ID = uuid.New()
		m.CreatedAt = time.Now()
		return nil
	})

	member, err := svc.AddMember(context.Background(), orgID, userID, "member")

	require.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, orgID, member.OrganizationID)
	assert.Equal(t, userID, member.UserID)
	assert.Equal(t, "member", member.Role)
}

func TestAddMember_AlreadyMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()
	userID := uuid.New()
	existingMember := &organization_member.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           "member",
	}

	// User is already a member
	mockMemberRepo.EXPECT().GetByOrgAndUser(gomock.Any(), orgID, userID).Return(existingMember, nil)

	member, err := svc.AddMember(context.Background(), orgID, userID, "member")

	assert.Error(t, err)
	assert.Equal(t, ErrAlreadyMember, err)
	assert.Nil(t, member)
}

func TestRemoveMember_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()
	userID := uuid.New()

	mockMemberRepo.EXPECT().Delete(gomock.Any(), orgID, userID).Return(nil)

	err := svc.RemoveMember(context.Background(), orgID, userID)

	require.NoError(t, err)
}

func TestIsMember_True(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()
	userID := uuid.New()
	member := &organization_member.OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         userID,
		Role:           "member",
	}

	mockMemberRepo.EXPECT().GetByOrgAndUser(gomock.Any(), orgID, userID).Return(member, nil)

	isMember, err := svc.IsMember(context.Background(), orgID, userID)

	require.NoError(t, err)
	assert.True(t, isMember)
}

func TestIsMember_False(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()
	userID := uuid.New()

	mockMemberRepo.EXPECT().GetByOrgAndUser(gomock.Any(), orgID, userID).Return(nil, gorm.ErrRecordNotFound)

	isMember, err := svc.IsMember(context.Background(), orgID, userID)

	require.NoError(t, err)
	assert.False(t, isMember)
}

func TestGetMembers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()
	expectedMembers := []*organization_member.OrganizationMember{
		{ID: uuid.New(), OrganizationID: orgID, UserID: uuid.New(), Role: "owner"},
		{ID: uuid.New(), OrganizationID: orgID, UserID: uuid.New(), Role: "member"},
	}

	mockMemberRepo.EXPECT().GetByOrgID(gomock.Any(), orgID).Return(expectedMembers, nil)

	members, err := svc.GetMembers(context.Background(), orgID)

	require.NoError(t, err)
	assert.Len(t, members, 2)
}

func TestGetOwner_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()
	ownerID := uuid.New()
	org := &organization.Organization{
		ID:      orgID,
		Name:    "Test Org",
		OwnerID: ownerID,
	}
	owner := &user.User{
		ID:       ownerID,
		Username: "owner",
	}

	mockOrgRepo.EXPECT().GetByID(gomock.Any(), orgID).Return(org, nil)
	mockUserRepo.EXPECT().GetByID(gomock.Any(), ownerID).Return(owner, nil)

	u, err := svc.GetOwner(context.Background(), orgID)

	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, ownerID, u.ID)
	assert.Equal(t, "owner", u.Username)
}

func TestGetOwner_OrgNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	orgID := uuid.New()

	mockOrgRepo.EXPECT().GetByID(gomock.Any(), orgID).Return(nil, gorm.ErrRecordNotFound)

	u, err := svc.GetOwner(context.Background(), orgID)

	assert.Error(t, err)
	assert.Equal(t, ErrOrgNotFound, err)
	assert.Nil(t, u)
}

func TestGetUserByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	userID := uuid.New()
	expectedUser := &user.User{
		ID:       userID,
		Username: "testuser",
	}

	mockUserRepo.EXPECT().GetByID(gomock.Any(), userID).Return(expectedUser, nil)

	u, err := svc.GetUserByID(context.Background(), userID)

	require.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, userID, u.ID)
}

func TestGetUserByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrgRepo := orgMocks.NewMockRepository(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)
	mockUserRepo := userMocks.NewMockRepository(ctrl)

	svc := NewService(mockOrgRepo, mockMemberRepo, mockUserRepo)

	userID := uuid.New()

	mockUserRepo.EXPECT().GetByID(gomock.Any(), userID).Return(nil, gorm.ErrRecordNotFound)

	u, err := svc.GetUserByID(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, u)
}

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "Test Org",
			expected: "test-org",
		},
		{
			name:     "multiple spaces",
			input:    "My  Test   Org",
			expected: "my-test-org",
		},
		{
			name:     "special characters",
			input:    "Test@Org#123!",
			expected: "testorg123",
		},
		{
			name:     "leading/trailing spaces",
			input:    "  Test Org  ",
			expected: "test-org",
		},
		{
			name:     "already lowercase",
			input:    "test-org",
			expected: "test-org",
		},
		{
			name:     "unicode characters",
			input:    "Café Org",
			expected: "caf-org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
