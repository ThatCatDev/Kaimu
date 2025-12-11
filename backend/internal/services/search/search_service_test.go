package search

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member"
	memberMocks "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member/mocks"
	"github.com/thatcatdev/kaimu/backend/internal/services/search/mocks"
	"github.com/typesense/typesense-go/v2/typesense/api"
	"go.uber.org/mock/gomock"
)

// Helper function to create a pointer to a value
func ptr[T any](v T) *T {
	return &v
}

func TestInitializeCollections(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("creates collections when they don't exist", func(t *testing.T) {
		schemas := GetAllSchemas()

		// Expect RetrieveCollection to fail for all collections (indicating they don't exist)
		for _, schema := range schemas {
			mockClient.EXPECT().
				RetrieveCollection(gomock.Any(), schema.Name).
				Return(nil, errors.New("collection not found"))
		}

		// Expect CreateCollection to be called for all collections
		for _, schema := range schemas {
			mockClient.EXPECT().
				CreateCollection(gomock.Any(), schema).
				Return(&api.CollectionResponse{Name: schema.Name}, nil)
		}

		err := svc.InitializeCollections(ctx)
		require.NoError(t, err)
	})

	t.Run("skips existing collections", func(t *testing.T) {
		schemas := GetAllSchemas()

		// Expect RetrieveCollection to succeed for all collections (indicating they exist)
		for _, schema := range schemas {
			mockClient.EXPECT().
				RetrieveCollection(gomock.Any(), schema.Name).
				Return(&api.CollectionResponse{Name: schema.Name}, nil)
		}

		// CreateCollection should not be called

		err := svc.InitializeCollections(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error if collection creation fails", func(t *testing.T) {
		// First collection doesn't exist
		mockClient.EXPECT().
			RetrieveCollection(gomock.Any(), CollectionOrganizations).
			Return(nil, errors.New("collection not found"))

		// CreateCollection fails
		mockClient.EXPECT().
			CreateCollection(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("creation failed"))

		err := svc.InitializeCollections(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create collection")
	})
}

func TestSearch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	userID := uuid.New()
	orgID := uuid.New()

	t.Run("returns empty results when user has no organizations", func(t *testing.T) {
		mockMemberRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return([]*organization_member.OrganizationMember{}, nil)

		results, err := svc.Search(ctx, userID, "test query", nil, 10)
		require.NoError(t, err)
		assert.Empty(t, results.Results)
		assert.Equal(t, 0, results.TotalCount)
		assert.Equal(t, "test query", results.Query)
	})

	t.Run("returns empty results when user doesn't have access to scoped org", func(t *testing.T) {
		// User belongs to orgID
		mockMemberRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return([]*organization_member.OrganizationMember{
				{OrganizationID: orgID, UserID: userID},
			}, nil)

		// But scope is for a different org
		differentOrgID := uuid.New()
		scope := &SearchScope{OrganizationID: differentOrgID.String()}

		results, err := svc.Search(ctx, userID, "test query", scope, 10)
		require.NoError(t, err)
		assert.Empty(t, results.Results)
		assert.Equal(t, 0, results.TotalCount)
	})

	t.Run("returns error when member repo fails", func(t *testing.T) {
		mockMemberRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return(nil, errors.New("database error"))

		results, err := svc.Search(ctx, userID, "test query", nil, 10)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "failed to get user organizations")
	})

	t.Run("successfully performs search with results", func(t *testing.T) {
		// User belongs to org
		mockMemberRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return([]*organization_member.OrganizationMember{
				{OrganizationID: orgID, UserID: userID},
			}, nil)

		// Mock multi-search response with card results
		cardDoc := map[string]interface{}{
			"id":                "card-123",
			"title":             "Test Card",
			"description":       "Card description",
			"organization_id":   orgID.String(),
			"organization_name": "Test Org",
			"project_id":        "proj-123",
			"project_name":      "Test Project",
			"board_id":          "board-123",
			"board_name":        "Test Board",
		}

		foundCount := 1
		textMatch := int64(100)
		mockClient.EXPECT().
			MultiSearch(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&api.MultiSearchResult{
				Results: []api.SearchResult{
					{
						Found: &foundCount,
						Hits: &[]api.SearchResultHit{
							{
								Document:  &cardDoc,
								TextMatch: &textMatch,
							},
						},
					},
					{Found: ptr(0), Hits: &[]api.SearchResultHit{}},
					{Found: ptr(0), Hits: &[]api.SearchResultHit{}},
					{Found: ptr(0), Hits: &[]api.SearchResultHit{}},
					{Found: ptr(0), Hits: &[]api.SearchResultHit{}},
				},
			}, nil)

		results, err := svc.Search(ctx, userID, "test", nil, 10)
		require.NoError(t, err)
		assert.Equal(t, 1, len(results.Results))
		assert.Equal(t, 1, results.TotalCount)
		assert.Equal(t, "Test Card", results.Results[0].Title)
		assert.Equal(t, EntityTypeCard, results.Results[0].Type)
		assert.Equal(t, "/projects/proj-123/board/board-123?card=card-123", results.Results[0].URL)
	})

	t.Run("enforces limit bounds", func(t *testing.T) {
		// User belongs to org
		mockMemberRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return([]*organization_member.OrganizationMember{
				{OrganizationID: orgID, UserID: userID},
			}, nil)

		mockClient.EXPECT().
			MultiSearch(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, params *api.MultiSearchParams, searches api.MultiSearchSearchesParameter) (*api.MultiSearchResult, error) {
				// Check that limit is capped at 50
				for _, search := range searches.Searches {
					assert.LessOrEqual(t, *search.PerPage, 50)
				}
				return &api.MultiSearchResult{
					Results: []api.SearchResult{
						{Found: ptr(0)},
						{Found: ptr(0)},
						{Found: ptr(0)},
						{Found: ptr(0)},
						{Found: ptr(0)},
					},
				}, nil
			})

		// Request with limit > 50
		_, err := svc.Search(ctx, userID, "test", nil, 100)
		require.NoError(t, err)
	})

	t.Run("applies default limit when limit is 0", func(t *testing.T) {
		mockMemberRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return([]*organization_member.OrganizationMember{
				{OrganizationID: orgID, UserID: userID},
			}, nil)

		mockClient.EXPECT().
			MultiSearch(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, params *api.MultiSearchParams, searches api.MultiSearchSearchesParameter) (*api.MultiSearchResult, error) {
				// Check that default limit of 20 is applied
				for _, search := range searches.Searches {
					assert.Equal(t, 20, *search.PerPage)
				}
				return &api.MultiSearchResult{
					Results: []api.SearchResult{
						{Found: ptr(0)},
						{Found: ptr(0)},
						{Found: ptr(0)},
						{Found: ptr(0)},
						{Found: ptr(0)},
					},
				}, nil
			})

		_, err := svc.Search(ctx, userID, "test", nil, 0)
		require.NoError(t, err)
	})

	t.Run("returns error when search fails", func(t *testing.T) {
		mockMemberRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return([]*organization_member.OrganizationMember{
				{OrganizationID: orgID, UserID: userID},
			}, nil)

		mockClient.EXPECT().
			MultiSearch(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("search failed"))

		results, err := svc.Search(ctx, userID, "test", nil, 10)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "search failed")
	})

	t.Run("applies organization scope filter", func(t *testing.T) {
		mockMemberRepo.EXPECT().
			GetByUserID(gomock.Any(), userID).
			Return([]*organization_member.OrganizationMember{
				{OrganizationID: orgID, UserID: userID},
			}, nil)

		mockClient.EXPECT().
			MultiSearch(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, params *api.MultiSearchParams, searches api.MultiSearchSearchesParameter) (*api.MultiSearchResult, error) {
				// Verify that org filter is applied with specific org ID
				assert.Contains(t, *searches.Searches[0].FilterBy, "organization_id:="+orgID.String())
				return &api.MultiSearchResult{
					Results: []api.SearchResult{
						{Found: ptr(0)},
						{Found: ptr(0)},
						{Found: ptr(0)},
						{Found: ptr(0)},
						{Found: ptr(0)},
					},
				}, nil
			})

		scope := &SearchScope{OrganizationID: orgID.String()}
		_, err := svc.Search(ctx, userID, "test", scope, 10)
		require.NoError(t, err)
	})
}

func TestIndexOrganization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		doc := &OrganizationDocument{
			ID:   "org-123",
			Name: "Test Org",
			Slug: "test-org",
		}

		mockClient.EXPECT().
			UpsertDocument(gomock.Any(), CollectionOrganizations, doc).
			Return(map[string]interface{}{"id": "org-123"}, nil)

		err := svc.IndexOrganization(ctx, doc)
		require.NoError(t, err)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		doc := &OrganizationDocument{ID: "org-123"}

		mockClient.EXPECT().
			UpsertDocument(gomock.Any(), CollectionOrganizations, doc).
			Return(nil, errors.New("upsert failed"))

		err := svc.IndexOrganization(ctx, doc)
		assert.Error(t, err)
	})
}

func TestIndexUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		doc := &UserDocument{
			ID:       "user-123",
			Username: "testuser",
			Email:    "test@example.com",
		}

		mockClient.EXPECT().
			UpsertDocument(gomock.Any(), CollectionUsers, doc).
			Return(map[string]interface{}{"id": "user-123"}, nil)

		err := svc.IndexUser(ctx, doc)
		require.NoError(t, err)
	})
}

func TestIndexProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		doc := &ProjectDocument{
			ID:             "proj-123",
			Name:           "Test Project",
			Key:            "TEST",
			OrganizationID: "org-123",
		}

		mockClient.EXPECT().
			UpsertDocument(gomock.Any(), CollectionProjects, doc).
			Return(map[string]interface{}{"id": "proj-123"}, nil)

		err := svc.IndexProject(ctx, doc)
		require.NoError(t, err)
	})
}

func TestIndexBoard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		doc := &BoardDocument{
			ID:        "board-123",
			Name:      "Test Board",
			ProjectID: "proj-123",
		}

		mockClient.EXPECT().
			UpsertDocument(gomock.Any(), CollectionBoards, doc).
			Return(map[string]interface{}{"id": "board-123"}, nil)

		err := svc.IndexBoard(ctx, doc)
		require.NoError(t, err)
	})
}

func TestIndexCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		doc := &CardDocument{
			ID:        "card-123",
			Title:     "Test Card",
			BoardID:   "board-123",
			ProjectID: "proj-123",
		}

		mockClient.EXPECT().
			UpsertDocument(gomock.Any(), CollectionCards, doc).
			Return(map[string]interface{}{"id": "card-123"}, nil)

		err := svc.IndexCard(ctx, doc)
		require.NoError(t, err)
	})
}

func TestDeleteOrganization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().
			DeleteDocument(gomock.Any(), CollectionOrganizations, "org-123").
			Return(map[string]interface{}{"id": "org-123"}, nil)

		err := svc.DeleteOrganization(ctx, "org-123")
		require.NoError(t, err)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mockClient.EXPECT().
			DeleteDocument(gomock.Any(), CollectionOrganizations, "org-123").
			Return(nil, errors.New("delete failed"))

		err := svc.DeleteOrganization(ctx, "org-123")
		assert.Error(t, err)
	})
}

func TestDeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().
			DeleteDocument(gomock.Any(), CollectionUsers, "user-123").
			Return(map[string]interface{}{"id": "user-123"}, nil)

		err := svc.DeleteUser(ctx, "user-123")
		require.NoError(t, err)
	})
}

func TestDeleteProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().
			DeleteDocument(gomock.Any(), CollectionProjects, "proj-123").
			Return(map[string]interface{}{"id": "proj-123"}, nil)

		err := svc.DeleteProject(ctx, "proj-123")
		require.NoError(t, err)
	})
}

func TestDeleteBoard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().
			DeleteDocument(gomock.Any(), CollectionBoards, "board-123").
			Return(map[string]interface{}{"id": "board-123"}, nil)

		err := svc.DeleteBoard(ctx, "board-123")
		require.NoError(t, err)
	})
}

func TestDeleteCard(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockTypesenseClient(ctrl)
	mockMemberRepo := memberMocks.NewMockRepository(ctrl)

	svc := NewService(mockClient, mockMemberRepo)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockClient.EXPECT().
			DeleteDocument(gomock.Any(), CollectionCards, "card-123").
			Return(map[string]interface{}{"id": "card-123"}, nil)

		err := svc.DeleteCard(ctx, "card-123")
		require.NoError(t, err)
	})
}

func TestGetStringField(t *testing.T) {
	t.Run("returns string value when present", func(t *testing.T) {
		doc := map[string]interface{}{"name": "test"}
		result := getStringField(doc, "name")
		assert.Equal(t, "test", result)
	})

	t.Run("returns empty string when key not found", func(t *testing.T) {
		doc := map[string]interface{}{}
		result := getStringField(doc, "name")
		assert.Equal(t, "", result)
	})

	t.Run("returns empty string when value is not string", func(t *testing.T) {
		doc := map[string]interface{}{"count": 123}
		result := getStringField(doc, "count")
		assert.Equal(t, "", result)
	})
}

func TestHitToSearchResult(t *testing.T) {
	svc := &service{}

	t.Run("returns nil for nil document", func(t *testing.T) {
		hit := api.SearchResultHit{Document: nil}
		result := svc.hitToSearchResult(hit, 0)
		assert.Nil(t, result)
	})

	t.Run("returns nil for invalid collection index", func(t *testing.T) {
		doc := map[string]interface{}{"id": "test"}
		hit := api.SearchResultHit{Document: &doc}
		result := svc.hitToSearchResult(hit, 99)
		assert.Nil(t, result)
	})

	t.Run("correctly parses card result (index 0)", func(t *testing.T) {
		doc := map[string]interface{}{
			"id":                "card-123",
			"title":             "Test Card",
			"description":       "Card description",
			"organization_id":   "org-123",
			"organization_name": "Test Org",
			"project_id":        "proj-123",
			"project_name":      "Test Project",
			"board_id":          "board-123",
			"board_name":        "Test Board",
		}
		textMatch := int64(100)
		hit := api.SearchResultHit{Document: &doc, TextMatch: &textMatch}

		result := svc.hitToSearchResult(hit, 0)
		require.NotNil(t, result)
		assert.Equal(t, EntityTypeCard, result.Type)
		assert.Equal(t, "card-123", result.ID)
		assert.Equal(t, "Test Card", result.Title)
		assert.Equal(t, "Card description", result.Description)
		assert.Equal(t, "/projects/proj-123/board/board-123?card=card-123", result.URL)
		assert.Equal(t, float64(100), result.Score)
	})

	t.Run("correctly parses project result (index 1)", func(t *testing.T) {
		doc := map[string]interface{}{
			"id":                "proj-123",
			"name":              "Test Project",
			"description":       "Project description",
			"organization_id":   "org-123",
			"organization_name": "Test Org",
		}
		hit := api.SearchResultHit{Document: &doc}

		result := svc.hitToSearchResult(hit, 1)
		require.NotNil(t, result)
		assert.Equal(t, EntityTypeProject, result.Type)
		assert.Equal(t, "proj-123", result.ID)
		assert.Equal(t, "Test Project", result.Title)
		assert.Equal(t, "/projects/proj-123", result.URL)
	})

	t.Run("correctly parses board result (index 2)", func(t *testing.T) {
		doc := map[string]interface{}{
			"id":                "board-123",
			"name":              "Test Board",
			"description":       "Board description",
			"organization_id":   "org-123",
			"organization_name": "Test Org",
			"project_id":        "proj-123",
			"project_name":      "Test Project",
		}
		hit := api.SearchResultHit{Document: &doc}

		result := svc.hitToSearchResult(hit, 2)
		require.NotNil(t, result)
		assert.Equal(t, EntityTypeBoard, result.Type)
		assert.Equal(t, "board-123", result.ID)
		assert.Equal(t, "Test Board", result.Title)
		assert.Equal(t, "/projects/proj-123/board/board-123", result.URL)
	})

	t.Run("correctly parses organization result (index 3)", func(t *testing.T) {
		doc := map[string]interface{}{
			"id":          "org-123",
			"name":        "Test Org",
			"description": "Org description",
		}
		hit := api.SearchResultHit{Document: &doc}

		result := svc.hitToSearchResult(hit, 3)
		require.NotNil(t, result)
		assert.Equal(t, EntityTypeOrganization, result.Type)
		assert.Equal(t, "org-123", result.ID)
		assert.Equal(t, "Test Org", result.Title)
		assert.Equal(t, "/organizations/org-123", result.URL)
	})

	t.Run("correctly parses user result (index 4)", func(t *testing.T) {
		doc := map[string]interface{}{
			"id":           "user-123",
			"username":     "testuser",
			"display_name": "Test User",
			"email":        "test@example.com",
		}
		hit := api.SearchResultHit{Document: &doc}

		result := svc.hitToSearchResult(hit, 4)
		require.NotNil(t, result)
		assert.Equal(t, EntityTypeUser, result.Type)
		assert.Equal(t, "user-123", result.ID)
		assert.Equal(t, "Test User", result.Title) // Uses display_name
		assert.Equal(t, "test@example.com", result.Description)
		assert.Equal(t, "/users/user-123", result.URL)
	})

	t.Run("user falls back to username when display_name is empty", func(t *testing.T) {
		doc := map[string]interface{}{
			"id":           "user-123",
			"username":     "testuser",
			"display_name": "",
			"email":        "test@example.com",
		}
		hit := api.SearchResultHit{Document: &doc}

		result := svc.hitToSearchResult(hit, 4)
		require.NotNil(t, result)
		assert.Equal(t, "testuser", result.Title)
	})

	t.Run("correctly extracts highlights", func(t *testing.T) {
		doc := map[string]interface{}{"id": "card-123", "title": "Test"}
		snippet1 := "matched <b>text</b>"
		snippet2 := "another <b>match</b>"
		highlights := []api.SearchHighlight{
			{Snippet: &snippet1},
			{Snippet: &snippet2},
		}
		hit := api.SearchResultHit{Document: &doc, Highlights: &highlights}

		result := svc.hitToSearchResult(hit, 0)
		require.NotNil(t, result)
		assert.Equal(t, "matched <b>text</b> ... another <b>match</b>", result.Highlight)
	})
}

func TestToUnixTimestamp(t *testing.T) {
	t.Run("returns 0 for zero time", func(t *testing.T) {
		var zeroTime = time.Time{}
		result := ToUnixTimestamp(zeroTime)
		assert.Equal(t, int64(0), result)
	})

	t.Run("returns correct timestamp", func(t *testing.T) {
		testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := ToUnixTimestamp(testTime)
		assert.Equal(t, testTime.Unix(), result)
	})
}

func TestToUnixTimestampPtr(t *testing.T) {
	t.Run("returns 0 for nil pointer", func(t *testing.T) {
		result := ToUnixTimestampPtr(nil)
		assert.Equal(t, int64(0), result)
	})

	t.Run("returns 0 for zero time pointer", func(t *testing.T) {
		zeroTime := time.Time{}
		result := ToUnixTimestampPtr(&zeroTime)
		assert.Equal(t, int64(0), result)
	})

	t.Run("returns correct timestamp", func(t *testing.T) {
		testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := ToUnixTimestampPtr(&testTime)
		assert.Equal(t, testTime.Unix(), result)
	})
}
