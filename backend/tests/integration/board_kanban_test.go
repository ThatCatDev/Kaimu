package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
	tagService "github.com/thatcatdev/kaimu/backend/internal/services/tag"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type BoardTestServer struct {
	handler http.Handler
	db      *gorm.DB
}

func setupBoardTestServer(t *testing.T) *BoardTestServer {
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

	// Clean up tables before test (order matters due to foreign keys)
	testDB.Exec("DELETE FROM card_tags")
	testDB.Exec("DELETE FROM cards")
	testDB.Exec("DELETE FROM tags")
	testDB.Exec("DELETE FROM board_columns")
	testDB.Exec("DELETE FROM boards")
	testDB.Exec("DELETE FROM projects")
	testDB.Exec("DELETE FROM organization_members")
	testDB.Exec("DELETE FROM organizations")
	testDB.Exec("DELETE FROM refresh_tokens")
	testDB.Exec("DELETE FROM users")

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
	permissionRepository := permissionRepo.NewRepository(testDB)
	roleRepository := roleRepo.NewRepository(testDB)
	rolePermissionRepository := rolePermissionRepo.NewRepository(testDB)

	// Create services
	refreshRepository := refreshTokenRepo.NewRepository(testDB)
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

	return &BoardTestServer{
		handler: wrappedHandler,
		db:      testDB,
	}
}

func (s *BoardTestServer) cleanup() {
	s.db.Exec("DELETE FROM card_tags")
	s.db.Exec("DELETE FROM cards")
	s.db.Exec("DELETE FROM tags")
	s.db.Exec("DELETE FROM board_columns")
	s.db.Exec("DELETE FROM boards")
	s.db.Exec("DELETE FROM projects")
	s.db.Exec("DELETE FROM organization_members")
	s.db.Exec("DELETE FROM organizations")
	s.db.Exec("DELETE FROM users")
}

type graphQLResponse struct {
	Data   json.RawMessage          `json:"data"`
	Errors []map[string]interface{} `json:"errors"`
}

func (s *BoardTestServer) executeQuery(query string, cookie string) *graphQLResponse {
	body, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: middleware.AccessTokenCookie, Value: cookie})
	}

	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, req)

	var resp graphQLResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return &resp
}

func (s *BoardTestServer) registerUser(username, password string) (string, error) {
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

func TestBoardCreationWithProject(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	// Register user and get token
	token, err := server.registerUser("boarduser", "password123")
	require.NoError(t, err)

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: { name: "Board Test Org" }) {
			id name slug
		}
	}`
	orgResp := server.executeQuery(createOrgQuery, token)
	require.Empty(t, orgResp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID
	require.NotEmpty(t, orgID)

	// Create project (should auto-create default board)
	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Board Test Project", key: "BTP" }) {
			id
			name
			boards { id name isDefault }
			defaultBoard {
				id
				name
				isDefault
				columns { id name position isBacklog isHidden }
			}
		}
	}`, orgID)

	projResp := server.executeQuery(createProjectQuery, token)
	require.Empty(t, projResp.Errors, "Expected no errors but got: %v", projResp.Errors)

	var projData struct {
		CreateProject struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Boards []struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				IsDefault bool   `json:"isDefault"`
			} `json:"boards"`
			DefaultBoard struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				IsDefault bool   `json:"isDefault"`
				Columns   []struct {
					ID        string `json:"id"`
					Name      string `json:"name"`
					Position  int    `json:"position"`
					IsBacklog bool   `json:"isBacklog"`
					IsHidden  bool   `json:"isHidden"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(projResp.Data, &projData)

	// Verify default board was created
	assert.Equal(t, 1, len(projData.CreateProject.Boards))
	assert.True(t, projData.CreateProject.Boards[0].IsDefault)
	assert.Equal(t, "Default Board", projData.CreateProject.DefaultBoard.Name)

	// Verify default columns were created
	columns := projData.CreateProject.DefaultBoard.Columns
	assert.Equal(t, 4, len(columns))

	// Check column names and properties
	columnNames := make(map[string]bool)
	for _, col := range columns {
		columnNames[col.Name] = true
		if col.Name == "Backlog" {
			assert.True(t, col.IsBacklog)
			assert.True(t, col.IsHidden)
		} else {
			assert.False(t, col.IsBacklog)
			assert.False(t, col.IsHidden)
		}
	}
	assert.True(t, columnNames["Backlog"])
	assert.True(t, columnNames["Todo"])
	assert.True(t, columnNames["In Progress"])
	assert.True(t, columnNames["Done"])
}

func TestCardCRUD(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	// Setup: Register user, create org, project
	token, err := server.registerUser("carduser", "password123")
	require.NoError(t, err)

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: { name: "Card Test Org" }) {
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

	// Create project
	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Card Test Project", key: "CTP" }) {
			id
			defaultBoard {
				id
				columns { id name }
			}
		}
	}`, orgData.CreateOrganization.ID)

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

	// Find the Todo column
	var todoColumnID string
	for _, col := range projData.CreateProject.DefaultBoard.Columns {
		if col.Name == "Todo" {
			todoColumnID = col.ID
			break
		}
	}
	require.NotEmpty(t, todoColumnID)

	// Test: Create a card
	createCardQuery := fmt.Sprintf(`mutation {
		createCard(input: {
			columnId: "%s"
			title: "Test Card"
			description: "This is a test card"
			priority: MEDIUM
		}) {
			id
			title
			description
			priority
			position
			column { id name }
		}
	}`, todoColumnID)

	cardResp := server.executeQuery(createCardQuery, token)
	require.Empty(t, cardResp.Errors, "Create card errors: %v", cardResp.Errors)

	var cardData struct {
		CreateCard struct {
			ID          string  `json:"id"`
			Title       string  `json:"title"`
			Description string  `json:"description"`
			Priority    string  `json:"priority"`
			Position    float64 `json:"position"`
			Column      struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"column"`
		} `json:"createCard"`
	}
	json.Unmarshal(cardResp.Data, &cardData)

	assert.Equal(t, "Test Card", cardData.CreateCard.Title)
	assert.Equal(t, "This is a test card", cardData.CreateCard.Description)
	assert.Equal(t, "MEDIUM", cardData.CreateCard.Priority)
	assert.Equal(t, float64(1000), cardData.CreateCard.Position)
	assert.Equal(t, "Todo", cardData.CreateCard.Column.Name)

	cardID := cardData.CreateCard.ID

	// Test: Update card
	updateCardQuery := fmt.Sprintf(`mutation {
		updateCard(input: {
			id: "%s"
			title: "Updated Card Title"
			priority: HIGH
		}) {
			id
			title
			priority
		}
	}`, cardID)

	updateResp := server.executeQuery(updateCardQuery, token)
	require.Empty(t, updateResp.Errors)

	var updateData struct {
		UpdateCard struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Priority string `json:"priority"`
		} `json:"updateCard"`
	}
	json.Unmarshal(updateResp.Data, &updateData)

	assert.Equal(t, "Updated Card Title", updateData.UpdateCard.Title)
	assert.Equal(t, "HIGH", updateData.UpdateCard.Priority)

	// Test: Query card
	queryCardQuery := fmt.Sprintf(`query {
		card(id: "%s") {
			id
			title
			priority
		}
	}`, cardID)

	getResp := server.executeQuery(queryCardQuery, token)
	require.Empty(t, getResp.Errors)

	// Test: Delete card
	deleteCardQuery := fmt.Sprintf(`mutation {
		deleteCard(id: "%s")
	}`, cardID)

	deleteResp := server.executeQuery(deleteCardQuery, token)
	require.Empty(t, deleteResp.Errors)
}

func TestMoveCard(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	// Setup
	token, err := server.registerUser("moveuser", "password123")
	require.NoError(t, err)

	createOrgQuery := `mutation { createOrganization(input: { name: "Move Test Org" }) { id } }`
	orgResp := server.executeQuery(createOrgQuery, token)
	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)

	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Move Test", key: "MVT" }) {
			defaultBoard { columns { id name } }
		}
	}`, orgData.CreateOrganization.ID)
	projResp := server.executeQuery(createProjectQuery, token)

	var projData struct {
		CreateProject struct {
			DefaultBoard struct {
				Columns []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(projResp.Data, &projData)

	var todoColID, inProgressColID string
	for _, col := range projData.CreateProject.DefaultBoard.Columns {
		if col.Name == "Todo" {
			todoColID = col.ID
		}
		if col.Name == "In Progress" {
			inProgressColID = col.ID
		}
	}

	// Create card in Todo
	createCardQuery := fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "Card to Move" }) {
			id
			column { name }
			position
		}
	}`, todoColID)
	cardResp := server.executeQuery(createCardQuery, token)

	var cardData struct {
		CreateCard struct {
			ID       string  `json:"id"`
			Position float64 `json:"position"`
			Column   struct {
				Name string `json:"name"`
			} `json:"column"`
		} `json:"createCard"`
	}
	json.Unmarshal(cardResp.Data, &cardData)
	cardID := cardData.CreateCard.ID
	assert.Equal(t, "Todo", cardData.CreateCard.Column.Name)

	// Move card to In Progress
	moveCardQuery := fmt.Sprintf(`mutation {
		moveCard(input: {
			cardId: "%s"
			targetColumnId: "%s"
		}) {
			id
			column { id name }
			position
		}
	}`, cardID, inProgressColID)

	moveResp := server.executeQuery(moveCardQuery, token)
	require.Empty(t, moveResp.Errors, "Move card errors: %v", moveResp.Errors)

	var moveData struct {
		MoveCard struct {
			ID       string  `json:"id"`
			Position float64 `json:"position"`
			Column   struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"column"`
		} `json:"moveCard"`
	}
	json.Unmarshal(moveResp.Data, &moveData)

	assert.Equal(t, "In Progress", moveData.MoveCard.Column.Name)
}

func TestTagCRUD(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	// Setup
	token, err := server.registerUser("taguser", "password123")
	require.NoError(t, err)

	createOrgQuery := `mutation { createOrganization(input: { name: "Tag Test Org" }) { id } }`
	orgResp := server.executeQuery(createOrgQuery, token)
	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)

	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Tag Test", key: "TAG" }) {
			id
		}
	}`, orgData.CreateOrganization.ID)
	projResp := server.executeQuery(createProjectQuery, token)
	var projData struct {
		CreateProject struct {
			ID string `json:"id"`
		} `json:"createProject"`
	}
	json.Unmarshal(projResp.Data, &projData)
	projectID := projData.CreateProject.ID

	// Create tag
	createTagQuery := fmt.Sprintf(`mutation {
		createTag(input: {
			projectId: "%s"
			name: "Bug"
			color: "#EF4444"
			description: "Bug fixes"
		}) {
			id
			name
			color
			description
		}
	}`, projectID)

	tagResp := server.executeQuery(createTagQuery, token)
	require.Empty(t, tagResp.Errors, "Create tag errors: %v", tagResp.Errors)

	var tagData struct {
		CreateTag struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Color       string `json:"color"`
			Description string `json:"description"`
		} `json:"createTag"`
	}
	json.Unmarshal(tagResp.Data, &tagData)

	assert.Equal(t, "Bug", tagData.CreateTag.Name)
	assert.Equal(t, "#EF4444", tagData.CreateTag.Color)
	tagID := tagData.CreateTag.ID

	// Query tags
	queryTagsQuery := fmt.Sprintf(`query {
		tags(projectId: "%s") {
			id
			name
			color
		}
	}`, projectID)

	queryResp := server.executeQuery(queryTagsQuery, token)
	require.Empty(t, queryResp.Errors)

	var queryData struct {
		Tags []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"tags"`
	}
	json.Unmarshal(queryResp.Data, &queryData)
	assert.Equal(t, 1, len(queryData.Tags))

	// Update tag
	updateTagQuery := fmt.Sprintf(`mutation {
		updateTag(input: {
			id: "%s"
			name: "Critical Bug"
			color: "#DC2626"
		}) {
			id
			name
			color
		}
	}`, tagID)

	updateResp := server.executeQuery(updateTagQuery, token)
	require.Empty(t, updateResp.Errors)

	var updateData struct {
		UpdateTag struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		} `json:"updateTag"`
	}
	json.Unmarshal(updateResp.Data, &updateData)
	assert.Equal(t, "Critical Bug", updateData.UpdateTag.Name)

	// Delete tag
	deleteTagQuery := fmt.Sprintf(`mutation { deleteTag(id: "%s") }`, tagID)
	deleteResp := server.executeQuery(deleteTagQuery, token)
	require.Empty(t, deleteResp.Errors)
}

func TestCardWithTags(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	// Setup
	token, err := server.registerUser("cardtaguser", "password123")
	require.NoError(t, err)

	createOrgQuery := `mutation { createOrganization(input: { name: "Card Tag Org" }) { id } }`
	orgResp := server.executeQuery(createOrgQuery, token)
	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)

	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Card Tag Test", key: "CTT" }) {
			id
			defaultBoard { columns { id name } }
		}
	}`, orgData.CreateOrganization.ID)
	projResp := server.executeQuery(createProjectQuery, token)
	var projData struct {
		CreateProject struct {
			ID           string `json:"id"`
			DefaultBoard struct {
				Columns []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(projResp.Data, &projData)
	projectID := projData.CreateProject.ID

	var todoColID string
	for _, col := range projData.CreateProject.DefaultBoard.Columns {
		if col.Name == "Todo" {
			todoColID = col.ID
			break
		}
	}

	// Create tags
	tag1Query := fmt.Sprintf(`mutation {
		createTag(input: { projectId: "%s", name: "Bug", color: "#EF4444" }) { id }
	}`, projectID)
	tag1Resp := server.executeQuery(tag1Query, token)
	var tag1Data struct {
		CreateTag struct {
			ID string `json:"id"`
		} `json:"createTag"`
	}
	json.Unmarshal(tag1Resp.Data, &tag1Data)
	tag1ID := tag1Data.CreateTag.ID

	tag2Query := fmt.Sprintf(`mutation {
		createTag(input: { projectId: "%s", name: "Feature", color: "#10B981" }) { id }
	}`, projectID)
	tag2Resp := server.executeQuery(tag2Query, token)
	var tag2Data struct {
		CreateTag struct {
			ID string `json:"id"`
		} `json:"createTag"`
	}
	json.Unmarshal(tag2Resp.Data, &tag2Data)
	tag2ID := tag2Data.CreateTag.ID

	// Create card with tags
	createCardQuery := fmt.Sprintf(`mutation {
		createCard(input: {
			columnId: "%s"
			title: "Card with Tags"
			tagIds: ["%s", "%s"]
		}) {
			id
			title
			tags { id name color }
		}
	}`, todoColID, tag1ID, tag2ID)

	cardResp := server.executeQuery(createCardQuery, token)
	require.Empty(t, cardResp.Errors, "Create card with tags errors: %v", cardResp.Errors)

	var cardData struct {
		CreateCard struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Tags  []struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"tags"`
		} `json:"createCard"`
	}
	json.Unmarshal(cardResp.Data, &cardData)

	assert.Equal(t, 2, len(cardData.CreateCard.Tags))

	// Update card to remove one tag
	updateCardQuery := fmt.Sprintf(`mutation {
		updateCard(input: {
			id: "%s"
			tagIds: ["%s"]
		}) {
			tags { id name }
		}
	}`, cardData.CreateCard.ID, tag1ID)

	updateResp := server.executeQuery(updateCardQuery, token)
	require.Empty(t, updateResp.Errors)

	var updateData struct {
		UpdateCard struct {
			Tags []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"tags"`
		} `json:"updateCard"`
	}
	json.Unmarshal(updateResp.Data, &updateData)
	assert.Equal(t, 1, len(updateData.UpdateCard.Tags))
	assert.Equal(t, "Bug", updateData.UpdateCard.Tags[0].Name)
}

func TestColumnOperations(t *testing.T) {
	server := setupBoardTestServer(t)
	defer server.cleanup()

	// Setup
	token, err := server.registerUser("columnuser", "password123")
	require.NoError(t, err)

	createOrgQuery := `mutation { createOrganization(input: { name: "Column Test Org" }) { id } }`
	orgResp := server.executeQuery(createOrgQuery, token)
	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)

	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "Column Test", key: "COL" }) {
			defaultBoard { id columns { id name position } }
		}
	}`, orgData.CreateOrganization.ID)
	projResp := server.executeQuery(createProjectQuery, token)
	var projData struct {
		CreateProject struct {
			DefaultBoard struct {
				ID      string `json:"id"`
				Columns []struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					Position int    `json:"position"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(projResp.Data, &projData)
	boardID := projData.CreateProject.DefaultBoard.ID

	// Create new column
	createColumnQuery := fmt.Sprintf(`mutation {
		createColumn(input: { boardId: "%s", name: "Review" }) {
			id
			name
			position
		}
	}`, boardID)

	colResp := server.executeQuery(createColumnQuery, token)
	require.Empty(t, colResp.Errors, "Create column errors: %v", colResp.Errors)

	var colData struct {
		CreateColumn struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Position int    `json:"position"`
		} `json:"createColumn"`
	}
	json.Unmarshal(colResp.Data, &colData)

	assert.Equal(t, "Review", colData.CreateColumn.Name)
	assert.Equal(t, 4, colData.CreateColumn.Position) // After the 4 default columns (0-3)

	// Update column
	updateColumnQuery := fmt.Sprintf(`mutation {
		updateColumn(input: { id: "%s", name: "Code Review", color: "#8B5CF6" }) {
			id
			name
			color
		}
	}`, colData.CreateColumn.ID)

	updateResp := server.executeQuery(updateColumnQuery, token)
	require.Empty(t, updateResp.Errors)

	// Toggle visibility
	toggleQuery := fmt.Sprintf(`mutation {
		toggleColumnVisibility(id: "%s") {
			id
			isHidden
		}
	}`, colData.CreateColumn.ID)

	toggleResp := server.executeQuery(toggleQuery, token)
	require.Empty(t, toggleResp.Errors)

	var toggleData struct {
		ToggleColumnVisibility struct {
			IsHidden bool `json:"isHidden"`
		} `json:"toggleColumnVisibility"`
	}
	json.Unmarshal(toggleResp.Data, &toggleData)
	assert.True(t, toggleData.ToggleColumnVisibility.IsHidden)
}
