package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/thatcatdev/kaimu/backend/graph"
	"github.com/thatcatdev/kaimu/backend/graph/generated"
	"github.com/thatcatdev/kaimu/backend/http/middleware"
	boardRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board"
	columnRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column"
	cardRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/card"
	cardTagRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/card_tag"
	orgRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization"
	memberRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member"
	permissionRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/permission"
	projectRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/project"
	projectMemberRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/project_member"
	refreshTokenRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/refreshtoken"
	roleRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/role"
	rolePermissionRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/role_permission"
	tagRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/tag"
	userRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/directives"
	"github.com/thatcatdev/kaimu/backend/internal/services/auth"
	boardService "github.com/thatcatdev/kaimu/backend/internal/services/board"
	cardService "github.com/thatcatdev/kaimu/backend/internal/services/card"
	orgService "github.com/thatcatdev/kaimu/backend/internal/services/organization"
	projectService "github.com/thatcatdev/kaimu/backend/internal/services/project"
	rbacService "github.com/thatcatdev/kaimu/backend/internal/services/rbac"
	"github.com/thatcatdev/kaimu/backend/internal/services/search"
	tagService "github.com/thatcatdev/kaimu/backend/internal/services/tag"
	"github.com/typesense/typesense-go/v2/typesense"
	"github.com/typesense/typesense-go/v2/typesense/api"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SearchTestServer struct {
	handler       http.Handler
	db            *gorm.DB
	searchService search.Service
	tsClient      *typesense.Client
}

func setupSearchTestServer(t *testing.T) *SearchTestServer {
	// Check if Typesense is available
	tsHost := os.Getenv("TEST_TYPESENSE_HOST")
	if tsHost == "" {
		tsHost = "localhost"
	}
	tsPort := os.Getenv("TEST_TYPESENSE_PORT")
	if tsPort == "" {
		tsPort = "8108"
	}
	tsAPIKey := os.Getenv("TEST_TYPESENSE_API_KEY")
	if tsAPIKey == "" {
		tsAPIKey = "dev_api_key"
	}

	// Create Typesense client
	tsClient := typesense.NewClient(
		typesense.WithServer(fmt.Sprintf("http://%s:%s", tsHost, tsPort)),
		typesense.WithAPIKey(tsAPIKey),
	)

	// Test connection to Typesense
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := tsClient.Health(ctx, 5*time.Second)
	if err != nil {
		t.Skipf("Skipping integration test: could not connect to Typesense at %s:%s: %v", tsHost, tsPort, err)
	}

	// Database setup (same as other integration tests)
	dbHost := os.Getenv("TEST_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("TEST_DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("TEST_DB_USER")
	if dbUser == "" {
		dbUser = "pulse"
	}
	dbPassword := os.Getenv("TEST_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "mysecretpassword"
	}
	dbName := os.Getenv("TEST_DB_NAME")
	if dbName == "" {
		dbName = "pulse_test"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	testDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping integration test: could not connect to test database: %v", err)
	}

	// Clean up tables before test
	cleanupTestData(testDB)

	// Create repositories
	userRepository := userRepo.NewRepository(testDB)
	orgRepository := orgRepo.NewRepository(testDB)
	memberRepository := memberRepo.NewRepository(testDB)
	projectRepository := projectRepo.NewRepository(testDB)
	projectMemberRepository := projectMemberRepo.NewRepository(testDB)
	boardRepository := boardRepo.NewRepository(testDB)
	columnRepository := columnRepo.NewRepository(testDB)
	cardRepository := cardRepo.NewRepository(testDB)
	tagRepository := tagRepo.NewRepository(testDB)
	cardTagRepository := cardTagRepo.NewRepository(testDB)
	refreshRepository := refreshTokenRepo.NewRepository(testDB)
	permissionRepository := permissionRepo.NewRepository(testDB)
	roleRepository := roleRepo.NewRepository(testDB)
	rolePermissionRepository := rolePermissionRepo.NewRepository(testDB)

	// Create Typesense client interface
	tsClientInterface := search.NewTypesenseClientFromRaw(tsClient)

	// Create search service
	searchSvc := search.NewService(tsClientInterface, memberRepository)

	// Initialize search collections
	err = searchSvc.InitializeCollections(context.Background())
	if err != nil {
		t.Fatalf("Failed to initialize search collections: %v", err)
	}

	// Create services
	authSvc := auth.NewService(userRepository, refreshRepository, "test-jwt-secret", 15, 7)
	orgSvc := orgService.NewService(orgRepository, memberRepository, userRepository)
	projSvc := projectService.NewService(projectRepository, orgRepository)
	boardSvc := boardService.NewService(boardRepository, columnRepository, projectRepository)
	cardSvc := cardService.NewService(cardRepository, columnRepository, boardRepository, tagRepository, cardTagRepository)
	tagSvc := tagService.NewService(tagRepository, projectRepository)
	rbacSvc := rbacService.NewService(
		permissionRepository,
		roleRepository,
		rolePermissionRepository,
		memberRepository,
		projectMemberRepository,
		projectRepository,
		boardRepository,
		userRepository,
	)

	// Create resolver
	cfg := config.Config{
		AppConfig: config.AppConfig{
			Env: "test",
		},
	}
	resolver := &graph.Resolver{
		Config:              cfg,
		AuthService:         authSvc,
		OrganizationService: orgSvc,
		ProjectService:      projSvc,
		BoardService:        boardSvc,
		CardService:         cardSvc,
		TagService:          tagSvc,
		SearchService:       searchSvc,
		RBACService:         rbacSvc,
	}

	// Create GraphQL handler
	gqlConfig := generated.Config{
		Resolvers:  resolver,
		Directives: directives.GetDirectives(),
	}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(gqlConfig))

	// Wrap with auth middleware
	wrappedHandler := middleware.AuthMiddleware(authSvc)(srv)

	return &SearchTestServer{
		handler:       wrappedHandler,
		db:            testDB,
		searchService: searchSvc,
		tsClient:      tsClient,
	}
}

func cleanupTestData(db *gorm.DB) {
	db.Exec("DELETE FROM card_tags")
	db.Exec("DELETE FROM cards")
	db.Exec("DELETE FROM tags")
	db.Exec("DELETE FROM board_columns")
	db.Exec("DELETE FROM boards")
	db.Exec("DELETE FROM projects")
	db.Exec("DELETE FROM organization_members")
	db.Exec("DELETE FROM organizations")
	db.Exec("DELETE FROM users")
}

func (s *SearchTestServer) cleanup() {
	cleanupTestData(s.db)

	// Clean up Typesense collections
	ctx := context.Background()
	collections := []string{"organizations", "users", "projects", "boards", "cards"}
	for _, col := range collections {
		// Delete all documents from collection
		s.tsClient.Collection(col).Documents().Delete(ctx, &api.DeleteDocumentsParams{
			FilterBy: strPtr("id:!=''"),
		})
	}
}

func (s *SearchTestServer) executeQuery(query string, cookie string) *searchGraphQLResponse {
	body, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: middleware.AccessTokenCookie, Value: cookie})
	}

	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, req)

	var resp searchGraphQLResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return &resp
}

type searchGraphQLResponse struct {
	Data   json.RawMessage          `json:"data"`
	Errors []map[string]interface{} `json:"errors"`
}

func (s *SearchTestServer) registerUser(username, password string) (string, error) {
	query := fmt.Sprintf(`mutation {
		register(input: { username: "%s", password: "%s", email: "%s@test.com" }) {
			user { id username }
		}
	}`, username, password, username)

	body, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, req)

	// Extract cookie from response
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == middleware.AccessTokenCookie {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("no auth cookie returned")
}

func TestSearchIntegration_EmptyResults(t *testing.T) {
	server := setupSearchTestServer(t)
	defer server.cleanup()

	// Register user and get token
	token, err := server.registerUser("searchuser1", "password123")
	require.NoError(t, err)

	// Search without any data - should return empty results
	searchQuery := `query {
		search(query: "test") {
			results { id title type }
			totalCount
			query
		}
	}`

	resp := server.executeQuery(searchQuery, token)
	require.Empty(t, resp.Errors, "Expected no errors but got: %v", resp.Errors)

	var searchData struct {
		Search struct {
			Results    []interface{} `json:"results"`
			TotalCount int           `json:"totalCount"`
			Query      string        `json:"query"`
		} `json:"search"`
	}
	err = json.Unmarshal(resp.Data, &searchData)
	require.NoError(t, err)

	assert.Equal(t, "test", searchData.Search.Query)
	assert.Equal(t, 0, searchData.Search.TotalCount)
	assert.Empty(t, searchData.Search.Results)
}

func TestSearchIntegration_OrganizationSearch(t *testing.T) {
	server := setupSearchTestServer(t)
	defer server.cleanup()

	ctx := context.Background()

	// Register user
	token, err := server.registerUser("searchuser2", "password123")
	require.NoError(t, err)

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: { name: "Searchable Org", description: "A test organization for search" }) {
			id name slug
		}
	}`
	orgResp := server.executeQuery(createOrgQuery, token)
	require.Empty(t, orgResp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)
	require.NotEmpty(t, orgData.CreateOrganization.ID)

	// Index the organization for search
	err = server.searchService.IndexOrganization(ctx, &search.OrganizationDocument{
		ID:          orgData.CreateOrganization.ID,
		Name:        orgData.CreateOrganization.Name,
		Slug:        orgData.CreateOrganization.Slug,
		Description: "A test organization for search",
		MemberIDs:   []string{}, // Will be populated from the membership
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	})
	require.NoError(t, err)

	// Wait for indexing to propagate
	time.Sleep(500 * time.Millisecond)

	// Search for the organization
	searchQuery := `query {
		search(query: "Searchable") {
			results {
				id
				title
				type
				description
				url
			}
			totalCount
			query
		}
	}`

	resp := server.executeQuery(searchQuery, token)
	require.Empty(t, resp.Errors, "Expected no errors but got: %v", resp.Errors)

	var searchData struct {
		Search struct {
			Results []struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Type        string `json:"type"`
				Description string `json:"description"`
				URL         string `json:"url"`
			} `json:"results"`
			TotalCount int    `json:"totalCount"`
			Query      string `json:"query"`
		} `json:"search"`
	}
	err = json.Unmarshal(resp.Data, &searchData)
	require.NoError(t, err)

	assert.Equal(t, "Searchable", searchData.Search.Query)
	assert.GreaterOrEqual(t, searchData.Search.TotalCount, 1)

	// Find our organization in results
	found := false
	for _, result := range searchData.Search.Results {
		if result.ID == orgData.CreateOrganization.ID {
			found = true
			assert.Equal(t, "Searchable Org", result.Title)
			assert.Equal(t, "ORGANIZATION", result.Type)
			assert.Contains(t, result.URL, "/organizations/")
			break
		}
	}
	assert.True(t, found, "Organization should be in search results")
}

func TestSearchIntegration_CardSearch(t *testing.T) {
	server := setupSearchTestServer(t)
	defer server.cleanup()

	ctx := context.Background()

	// Register user
	token, err := server.registerUser("searchuser3", "password123")
	require.NoError(t, err)

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: { name: "Card Search Org" }) {
			id
		}
	}`
	orgResp := server.executeQuery(createOrgQuery, token)
	require.Empty(t, orgResp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// Create project
	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Card Search Project", key: "CSP" }) {
			id
			name
			defaultBoard {
				id
				name
				columns { id name }
			}
		}
	}`, orgID)
	projResp := server.executeQuery(createProjectQuery, token)
	require.Empty(t, projResp.Errors)

	var projData struct {
		CreateProject struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			DefaultBoard struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Columns []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(projResp.Data, &projData)
	projectID := projData.CreateProject.ID
	boardID := projData.CreateProject.DefaultBoard.ID

	// Find Todo column
	var todoColID string
	for _, col := range projData.CreateProject.DefaultBoard.Columns {
		if col.Name == "Todo" {
			todoColID = col.ID
			break
		}
	}
	require.NotEmpty(t, todoColID)

	// Create card
	createCardQuery := fmt.Sprintf(`mutation {
		createCard(input: {
			columnId: "%s"
			title: "Searchable Bug Fix"
			description: "This card is about fixing a critical bug in the system"
		}) {
			id
			title
			description
		}
	}`, todoColID)
	cardResp := server.executeQuery(createCardQuery, token)
	require.Empty(t, cardResp.Errors, "Create card errors: %v", cardResp.Errors)

	var cardData struct {
		CreateCard struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"createCard"`
	}
	json.Unmarshal(cardResp.Data, &cardData)
	cardID := cardData.CreateCard.ID

	// Index the card for search
	err = server.searchService.IndexCard(ctx, &search.CardDocument{
		ID:               cardID,
		Title:            "Searchable Bug Fix",
		Description:      "This card is about fixing a critical bug in the system",
		Priority:         "none",
		BoardID:          boardID,
		BoardName:        "Default Board",
		ProjectID:        projectID,
		ProjectName:      "Card Search Project",
		ProjectKey:       "CSP",
		OrganizationID:   orgID,
		OrganizationName: "Card Search Org",
		Tags:             []string{},
		CreatedAt:        time.Now().Unix(),
		UpdatedAt:        time.Now().Unix(),
	})
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Search for the card by title
	searchQuery := `query {
		search(query: "Searchable Bug") {
			results {
				id
				title
				type
				description
				url
				projectId
				projectName
				boardId
				boardName
			}
			totalCount
		}
	}`

	resp := server.executeQuery(searchQuery, token)
	require.Empty(t, resp.Errors)

	var searchData struct {
		Search struct {
			Results []struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Type        string `json:"type"`
				Description string `json:"description"`
				URL         string `json:"url"`
				ProjectID   string `json:"projectId"`
				ProjectName string `json:"projectName"`
				BoardID     string `json:"boardId"`
				BoardName   string `json:"boardName"`
			} `json:"results"`
			TotalCount int `json:"totalCount"`
		} `json:"search"`
	}
	err = json.Unmarshal(resp.Data, &searchData)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, searchData.Search.TotalCount, 1)

	// Find our card in results
	found := false
	for _, result := range searchData.Search.Results {
		if result.ID == cardID {
			found = true
			assert.Equal(t, "Searchable Bug Fix", result.Title)
			assert.Equal(t, "CARD", result.Type)
			assert.Contains(t, result.URL, "?card="+cardID)
			assert.Equal(t, projectID, result.ProjectID)
			assert.Equal(t, boardID, result.BoardID)
			break
		}
	}
	assert.True(t, found, "Card should be in search results")
}

func TestSearchIntegration_ScopedSearch(t *testing.T) {
	server := setupSearchTestServer(t)
	defer server.cleanup()

	ctx := context.Background()

	// Register user
	token, err := server.registerUser("searchuser4", "password123")
	require.NoError(t, err)

	// Create two organizations
	createOrg1Query := `mutation {
		createOrganization(input: { name: "Org One" }) { id }
	}`
	org1Resp := server.executeQuery(createOrg1Query, token)
	require.Empty(t, org1Resp.Errors)
	var org1Data struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(org1Resp.Data, &org1Data)
	org1ID := org1Data.CreateOrganization.ID

	createOrg2Query := `mutation {
		createOrganization(input: { name: "Org Two" }) { id }
	}`
	org2Resp := server.executeQuery(createOrg2Query, token)
	require.Empty(t, org2Resp.Errors)
	var org2Data struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(org2Resp.Data, &org2Data)
	org2ID := org2Data.CreateOrganization.ID

	// Create project in each org
	createProj1Query := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Alpha Project", key: "ALP" }) {
			id
			defaultBoard { id columns { id name } }
		}
	}`, org1ID)
	proj1Resp := server.executeQuery(createProj1Query, token)
	require.Empty(t, proj1Resp.Errors)
	var proj1Data struct {
		CreateProject struct {
			ID           string `json:"id"`
			DefaultBoard struct {
				ID      string `json:"id"`
				Columns []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(proj1Resp.Data, &proj1Data)

	createProj2Query := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Beta Project", key: "BET" }) {
			id
			defaultBoard { id columns { id name } }
		}
	}`, org2ID)
	proj2Resp := server.executeQuery(createProj2Query, token)
	require.Empty(t, proj2Resp.Errors)
	var proj2Data struct {
		CreateProject struct {
			ID           string `json:"id"`
			DefaultBoard struct {
				ID      string `json:"id"`
				Columns []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(proj2Resp.Data, &proj2Data)

	// Find Todo columns
	var todoCol1, todoCol2 string
	for _, col := range proj1Data.CreateProject.DefaultBoard.Columns {
		if col.Name == "Todo" {
			todoCol1 = col.ID
			break
		}
	}
	for _, col := range proj2Data.CreateProject.DefaultBoard.Columns {
		if col.Name == "Todo" {
			todoCol2 = col.ID
			break
		}
	}

	// Create cards in each project with "UniqueSearchTerm"
	createCard1Query := fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "UniqueSearchTerm Card in Org One" }) { id }
	}`, todoCol1)
	card1Resp := server.executeQuery(createCard1Query, token)
	require.Empty(t, card1Resp.Errors)
	var card1Data struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(card1Resp.Data, &card1Data)

	createCard2Query := fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "UniqueSearchTerm Card in Org Two" }) { id }
	}`, todoCol2)
	card2Resp := server.executeQuery(createCard2Query, token)
	require.Empty(t, card2Resp.Errors)
	var card2Data struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(card2Resp.Data, &card2Data)

	// Index both cards
	err = server.searchService.IndexCard(ctx, &search.CardDocument{
		ID:               card1Data.CreateCard.ID,
		Title:            "UniqueSearchTerm Card in Org One",
		OrganizationID:   org1ID,
		OrganizationName: "Org One",
		ProjectID:        proj1Data.CreateProject.ID,
		ProjectName:      "Alpha Project",
		BoardID:          proj1Data.CreateProject.DefaultBoard.ID,
		BoardName:        "Default Board",
		CreatedAt:        time.Now().Unix(),
		UpdatedAt:        time.Now().Unix(),
	})
	require.NoError(t, err)

	err = server.searchService.IndexCard(ctx, &search.CardDocument{
		ID:               card2Data.CreateCard.ID,
		Title:            "UniqueSearchTerm Card in Org Two",
		OrganizationID:   org2ID,
		OrganizationName: "Org Two",
		ProjectID:        proj2Data.CreateProject.ID,
		ProjectName:      "Beta Project",
		BoardID:          proj2Data.CreateProject.DefaultBoard.ID,
		BoardName:        "Default Board",
		CreatedAt:        time.Now().Unix(),
		UpdatedAt:        time.Now().Unix(),
	})
	require.NoError(t, err)

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Search globally - should find both cards
	globalSearchQuery := `query {
		search(query: "UniqueSearchTerm") {
			results { id title }
			totalCount
		}
	}`
	globalResp := server.executeQuery(globalSearchQuery, token)
	require.Empty(t, globalResp.Errors)

	var globalData struct {
		Search struct {
			Results    []interface{} `json:"results"`
			TotalCount int           `json:"totalCount"`
		} `json:"search"`
	}
	json.Unmarshal(globalResp.Data, &globalData)
	assert.Equal(t, 2, globalData.Search.TotalCount, "Global search should find both cards")

	// Search scoped to Org One - should find only one card
	scopedSearchQuery := fmt.Sprintf(`query {
		search(query: "UniqueSearchTerm", scope: { organizationId: "%s" }) {
			results { id title }
			totalCount
		}
	}`, org1ID)
	scopedResp := server.executeQuery(scopedSearchQuery, token)
	require.Empty(t, scopedResp.Errors)

	var scopedData struct {
		Search struct {
			Results []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"results"`
			TotalCount int `json:"totalCount"`
		} `json:"search"`
	}
	json.Unmarshal(scopedResp.Data, &scopedData)
	assert.Equal(t, 1, scopedData.Search.TotalCount, "Scoped search should find only one card")
	assert.Contains(t, scopedData.Search.Results[0].Title, "Org One", "Scoped search should find the card in Org One")
}

func TestSearchIntegration_SearchWithLimit(t *testing.T) {
	server := setupSearchTestServer(t)
	defer server.cleanup()

	ctx := context.Background()

	// Register user
	token, err := server.registerUser("searchuser5", "password123")
	require.NoError(t, err)

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: { name: "Limit Test Org" }) { id }
	}`
	orgResp := server.executeQuery(createOrgQuery, token)
	require.Empty(t, orgResp.Errors)
	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// Create project
	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Limit Test Project", key: "LTP" }) {
			id
			defaultBoard { id columns { id name } }
		}
	}`, orgID)
	projResp := server.executeQuery(createProjectQuery, token)
	require.Empty(t, projResp.Errors)
	var projData struct {
		CreateProject struct {
			ID           string `json:"id"`
			DefaultBoard struct {
				ID      string `json:"id"`
				Columns []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(projResp.Data, &projData)

	// Find Todo column
	var todoColID string
	for _, col := range projData.CreateProject.DefaultBoard.Columns {
		if col.Name == "Todo" {
			todoColID = col.ID
			break
		}
	}

	// Create multiple cards
	for i := 1; i <= 5; i++ {
		createCardQuery := fmt.Sprintf(`mutation {
			createCard(input: { columnId: "%s", title: "LimitTestCard %d" }) { id }
		}`, todoColID, i)
		cardResp := server.executeQuery(createCardQuery, token)
		require.Empty(t, cardResp.Errors)

		var cardData struct {
			CreateCard struct {
				ID string `json:"id"`
			} `json:"createCard"`
		}
		json.Unmarshal(cardResp.Data, &cardData)

		// Index each card
		err = server.searchService.IndexCard(ctx, &search.CardDocument{
			ID:               cardData.CreateCard.ID,
			Title:            fmt.Sprintf("LimitTestCard %d", i),
			OrganizationID:   orgID,
			OrganizationName: "Limit Test Org",
			ProjectID:        projData.CreateProject.ID,
			ProjectName:      "Limit Test Project",
			BoardID:          projData.CreateProject.DefaultBoard.ID,
			BoardName:        "Default Board",
			CreatedAt:        time.Now().Unix(),
			UpdatedAt:        time.Now().Unix(),
		})
		require.NoError(t, err)
	}

	// Wait for indexing
	time.Sleep(500 * time.Millisecond)

	// Search with limit of 2
	searchQuery := `query {
		search(query: "LimitTestCard", limit: 2) {
			results { id title }
			totalCount
		}
	}`
	resp := server.executeQuery(searchQuery, token)
	require.Empty(t, resp.Errors)

	var searchData struct {
		Search struct {
			Results []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"results"`
			TotalCount int `json:"totalCount"`
		} `json:"search"`
	}
	json.Unmarshal(resp.Data, &searchData)

	// TotalCount should be 5 (all matching), but results should be limited to 2
	assert.Equal(t, 5, searchData.Search.TotalCount, "Total count should include all matching cards")
	assert.LessOrEqual(t, len(searchData.Search.Results), 2, "Results should be limited to 2")
}

// Helper function to create a string pointer
func strPtr(s string) *string {
	return &s
}
