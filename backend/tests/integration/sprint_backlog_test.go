package integration

import (
	"bytes"
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
	auditRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/audit"
	metricsHistoryRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/metrics_history"
	orgRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization"
	memberRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member"
	projectRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/project"
	refreshTokenRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/refreshtoken"
	sprintRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/sprint"
	tagRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/tag"
	userRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/directives"
	"github.com/thatcatdev/kaimu/backend/internal/services/auth"
	boardService "github.com/thatcatdev/kaimu/backend/internal/services/board"
	cardService "github.com/thatcatdev/kaimu/backend/internal/services/card"
	metricsService "github.com/thatcatdev/kaimu/backend/internal/services/metrics"
	orgService "github.com/thatcatdev/kaimu/backend/internal/services/organization"
	projectService "github.com/thatcatdev/kaimu/backend/internal/services/project"
	sprintService "github.com/thatcatdev/kaimu/backend/internal/services/sprint"
	tagService "github.com/thatcatdev/kaimu/backend/internal/services/tag"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SprintTestServer struct {
	handler http.Handler
	db      *gorm.DB
}

func setupSprintTestServer(t *testing.T) *SprintTestServer {
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

	// Run migrations for all tables including sprint-related
	err = testDB.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS organizations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS organization_members (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL DEFAULT 'member',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(organization_id, user_id)
		);

		CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			key VARCHAR(10) NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(organization_id, key)
		);

		CREATE TABLE IF NOT EXISTS boards (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			is_default BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_by UUID REFERENCES users(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS board_columns (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			is_backlog BOOLEAN NOT NULL DEFAULT FALSE,
			is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
			is_done BOOLEAN NOT NULL DEFAULT FALSE,
			color VARCHAR(7) DEFAULT '#6B7280',
			wip_limit INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS tags (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			color VARCHAR(7) NOT NULL DEFAULT '#6B7280',
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE (project_id, name)
		);

		DO $$ BEGIN
			CREATE TYPE card_priority AS ENUM ('none', 'low', 'medium', 'high', 'urgent');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		CREATE TABLE IF NOT EXISTS cards (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			column_id UUID NOT NULL REFERENCES board_columns(id) ON DELETE CASCADE,
			board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
			title VARCHAR(500) NOT NULL,
			description TEXT,
			position FLOAT NOT NULL DEFAULT 0,
			priority card_priority NOT NULL DEFAULT 'none',
			assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
			due_date TIMESTAMP WITH TIME ZONE,
			story_points INTEGER,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_by UUID REFERENCES users(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS card_tags (
			card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			PRIMARY KEY (card_id, tag_id)
		);

		DO $$ BEGIN
			CREATE TYPE sprint_status AS ENUM ('future', 'active', 'closed');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		CREATE TABLE IF NOT EXISTS sprints (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			goal TEXT,
			start_date TIMESTAMP WITH TIME ZONE,
			end_date TIMESTAMP WITH TIME ZONE,
			status sprint_status NOT NULL DEFAULT 'future',
			position INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			created_by UUID REFERENCES users(id) ON DELETE SET NULL
		);

		CREATE TABLE IF NOT EXISTS card_sprints (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
			added_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(card_id, sprint_id)
		);

		CREATE TABLE IF NOT EXISTS metrics_history (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
			recorded_date DATE NOT NULL,
			total_cards INTEGER NOT NULL DEFAULT 0,
			completed_cards INTEGER NOT NULL DEFAULT 0,
			total_story_points INTEGER NOT NULL DEFAULT 0,
			completed_story_points INTEGER NOT NULL DEFAULT 0,
			column_snapshot JSONB NOT NULL DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE(sprint_id, recorded_date)
		);

		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) NOT NULL UNIQUE,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			revoked_at TIMESTAMP WITH TIME ZONE,
			replaced_by UUID,
			user_agent TEXT,
			ip_address VARCHAR(45),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS audit_log (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			entity_type VARCHAR(50) NOT NULL,
			entity_id UUID NOT NULL,
			action VARCHAR(50) NOT NULL,
			user_id UUID REFERENCES users(id) ON DELETE SET NULL,
			old_values JSONB,
			new_values JSONB,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
	`).Error
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clean up tables before test (order matters due to foreign keys)
	testDB.Exec("DELETE FROM audit_log")
	testDB.Exec("DELETE FROM metrics_history")
	testDB.Exec("DELETE FROM card_sprints")
	testDB.Exec("DELETE FROM sprints")
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
	boardRepository := boardRepo.NewRepository(testDB)
	columnRepository := columnRepo.NewRepository(testDB)
	cardRepository := cardRepo.NewRepository(testDB)
	tagRepository := tagRepo.NewRepository(testDB)
	cardTagRepository := cardTagRepo.NewRepository(testDB)
	sprintRepository := sprintRepo.NewRepository(testDB)
	metricsHistoryRepository := metricsHistoryRepo.NewRepository(testDB)
	refreshRepository := refreshTokenRepo.NewRepository(testDB)
	auditRepository := auditRepo.NewRepository(testDB)

	// Create services
	authSvc := auth.NewService(userRepository, refreshRepository, "test-jwt-secret", 15, 7)
	orgSvc := orgService.NewService(orgRepository, memberRepository, userRepository)
	projSvc := projectService.NewService(projectRepository, orgRepository)
	boardSvc := boardService.NewService(boardRepository, columnRepository, projectRepository)
	cardSvc := cardService.NewService(cardRepository, columnRepository, boardRepository, tagRepository, cardTagRepository)
	tagSvc := tagService.NewService(tagRepository, projectRepository)
	sprintSvc := sprintService.NewService(sprintRepository, cardRepository, boardRepository, columnRepository)
	metricsSvc := metricsService.NewService(sprintRepository, cardRepository, columnRepository, metricsHistoryRepository, auditRepository)

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
		SprintService:       sprintSvc,
		MetricsService:      metricsSvc,
	}

	// Create GraphQL handler
	gqlConfig := generated.Config{
		Resolvers:  resolver,
		Directives: directives.GetDirectives(),
	}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(gqlConfig))

	// Wrap with auth middleware
	wrappedHandler := middleware.AuthMiddleware(authSvc)(srv)

	return &SprintTestServer{
		handler: wrappedHandler,
		db:      testDB,
	}
}

func (s *SprintTestServer) cleanup() {
	s.db.Exec("DELETE FROM metrics_history")
	s.db.Exec("DELETE FROM card_sprints")
	s.db.Exec("DELETE FROM sprints")
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

func (s *SprintTestServer) executeQuery(query string, cookie string) *graphQLResponse {
	body, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "pulse_token", Value: cookie})
	}

	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, req)

	var resp graphQLResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	return &resp
}

func (s *SprintTestServer) registerUser(username, password string) (string, error) {
	query := fmt.Sprintf(`mutation {
		register(input: { username: "%s", email: "%s@test.com", password: "%s" }) {
			user { id username }
		}
	}`, username, username, password)

	body, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.handler.ServeHTTP(w, req)

	// Extract cookie from response
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "pulse_token" {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("no auth cookie returned")
}

// Helper to setup a project with board
func (s *SprintTestServer) setupProject(t *testing.T, token, name, key string) (projectID, boardID string, columns map[string]string) {
	createOrgQuery := fmt.Sprintf(`mutation { createOrganization(input: { name: "%s Org" }) { id } }`, name)
	orgResp := s.executeQuery(createOrgQuery, token)
	require.Empty(t, orgResp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(orgResp.Data, &orgData)

	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: { organizationId: "%s", name: "%s", key: "%s" }) {
			id
			defaultBoard {
				id
				columns { id name isBacklog }
			}
		}
	}`, orgData.CreateOrganization.ID, name, key)

	projResp := s.executeQuery(createProjectQuery, token)
	require.Empty(t, projResp.Errors)

	var projData struct {
		CreateProject struct {
			ID           string `json:"id"`
			DefaultBoard struct {
				ID      string `json:"id"`
				Columns []struct {
					ID        string `json:"id"`
					Name      string `json:"name"`
					IsBacklog bool   `json:"isBacklog"`
				} `json:"columns"`
			} `json:"defaultBoard"`
		} `json:"createProject"`
	}
	json.Unmarshal(projResp.Data, &projData)

	columns = make(map[string]string)
	for _, col := range projData.CreateProject.DefaultBoard.Columns {
		columns[col.Name] = col.ID
	}

	return projData.CreateProject.ID, projData.CreateProject.DefaultBoard.ID, columns
}

func TestSprintCRUD(t *testing.T) {
	server := setupSprintTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("sprintuser", "password123")
	require.NoError(t, err)

	_, boardID, _ := server.setupProject(t, token, "Sprint Test", "SPT")

	// Create a future sprint
	startDate := time.Now().AddDate(0, 0, 7).Format(time.RFC3339)
	endDate := time.Now().AddDate(0, 0, 21).Format(time.RFC3339)

	createSprintQuery := fmt.Sprintf(`mutation {
		createSprint(input: {
			boardId: "%s"
			name: "Sprint 1"
			goal: "Complete first milestone"
			startDate: "%s"
			endDate: "%s"
		}) {
			id
			name
			goal
			status
		}
	}`, boardID, startDate, endDate)

	sprintResp := server.executeQuery(createSprintQuery, token)
	require.Empty(t, sprintResp.Errors, "Create sprint errors: %v", sprintResp.Errors)

	var sprintData struct {
		CreateSprint struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Goal   string `json:"goal"`
			Status string `json:"status"`
		} `json:"createSprint"`
	}
	json.Unmarshal(sprintResp.Data, &sprintData)

	assert.Equal(t, "Sprint 1", sprintData.CreateSprint.Name)
	assert.Equal(t, "Complete first milestone", sprintData.CreateSprint.Goal)
	assert.Equal(t, "FUTURE", sprintData.CreateSprint.Status)

	sprintID := sprintData.CreateSprint.ID

	// Start the sprint
	startSprintQuery := fmt.Sprintf(`mutation {
		startSprint(id: "%s") {
			id
			status
		}
	}`, sprintID)

	startResp := server.executeQuery(startSprintQuery, token)
	require.Empty(t, startResp.Errors, "Start sprint errors: %v", startResp.Errors)

	var startData struct {
		StartSprint struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"startSprint"`
	}
	json.Unmarshal(startResp.Data, &startData)
	assert.Equal(t, "ACTIVE", startData.StartSprint.Status)
}

func TestAddCardToSprint(t *testing.T) {
	server := setupSprintTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("addcarduser", "password123")
	require.NoError(t, err)

	_, boardID, columns := server.setupProject(t, token, "Add Card Test", "ACT")

	// Create a sprint and start it
	startDate := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	endDate := time.Now().AddDate(0, 0, 13).Format(time.RFC3339)

	createSprintQuery := fmt.Sprintf(`mutation {
		createSprint(input: {
			boardId: "%s"
			name: "Active Sprint"
			startDate: "%s"
			endDate: "%s"
		}) { id }
	}`, boardID, startDate, endDate)

	sprintResp := server.executeQuery(createSprintQuery, token)
	require.Empty(t, sprintResp.Errors)

	var sprintData struct {
		CreateSprint struct {
			ID string `json:"id"`
		} `json:"createSprint"`
	}
	json.Unmarshal(sprintResp.Data, &sprintData)
	sprintID := sprintData.CreateSprint.ID

	// Start the sprint
	server.executeQuery(fmt.Sprintf(`mutation { startSprint(id: "%s") { id } }`, sprintID), token)

	// Create a card
	todoColumnID := columns["Todo"]
	createCardQuery := fmt.Sprintf(`mutation {
		createCard(input: {
			columnId: "%s"
			title: "Test Card"
			storyPoints: 5
		}) {
			id
			title
			storyPoints
			sprints { id name }
		}
	}`, todoColumnID)

	cardResp := server.executeQuery(createCardQuery, token)
	require.Empty(t, cardResp.Errors, "Create card errors: %v", cardResp.Errors)

	var cardData struct {
		CreateCard struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			StoryPoints int    `json:"storyPoints"`
			Sprints     []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"sprints"`
		} `json:"createCard"`
	}
	json.Unmarshal(cardResp.Data, &cardData)
	cardID := cardData.CreateCard.ID
	assert.Equal(t, 5, cardData.CreateCard.StoryPoints)

	// Add card to sprint
	addToSprintQuery := fmt.Sprintf(`mutation {
		addCardToSprint(input: { cardId: "%s", sprintId: "%s" }) {
			id
			sprints { id name }
		}
	}`, cardID, sprintID)

	addResp := server.executeQuery(addToSprintQuery, token)
	require.Empty(t, addResp.Errors, "Add to sprint errors: %v", addResp.Errors)

	var addData struct {
		AddCardToSprint struct {
			ID      string `json:"id"`
			Sprints []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"sprints"`
		} `json:"addCardToSprint"`
	}
	json.Unmarshal(addResp.Data, &addData)

	assert.Equal(t, 1, len(addData.AddCardToSprint.Sprints))
	assert.Equal(t, "Active Sprint", addData.AddCardToSprint.Sprints[0].Name)
}

func TestMoveCardToBacklogRemovesFromSprint(t *testing.T) {
	server := setupSprintTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("backloguser", "password123")
	require.NoError(t, err)

	_, boardID, columns := server.setupProject(t, token, "Backlog Test", "BLT")

	// Create and start a sprint
	startDate := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	endDate := time.Now().AddDate(0, 0, 13).Format(time.RFC3339)

	createSprintQuery := fmt.Sprintf(`mutation {
		createSprint(input: {
			boardId: "%s"
			name: "Backlog Test Sprint"
			startDate: "%s"
			endDate: "%s"
		}) { id }
	}`, boardID, startDate, endDate)

	sprintResp := server.executeQuery(createSprintQuery, token)
	var sprintData struct {
		CreateSprint struct {
			ID string `json:"id"`
		} `json:"createSprint"`
	}
	json.Unmarshal(sprintResp.Data, &sprintData)
	sprintID := sprintData.CreateSprint.ID

	server.executeQuery(fmt.Sprintf(`mutation { startSprint(id: "%s") { id } }`, sprintID), token)

	// Create a card in Todo column
	todoColumnID := columns["Todo"]
	createCardQuery := fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "Card to move to backlog" }) { id }
	}`, todoColumnID)

	cardResp := server.executeQuery(createCardQuery, token)
	var cardData struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(cardResp.Data, &cardData)
	cardID := cardData.CreateCard.ID

	// Add card to sprint
	server.executeQuery(fmt.Sprintf(`mutation {
		addCardToSprint(input: { cardId: "%s", sprintId: "%s" }) { id }
	}`, cardID, sprintID), token)

	// Verify card is in sprint
	getCardQuery := fmt.Sprintf(`query { card(id: "%s") { id sprints { id } } }`, cardID)
	getResp := server.executeQuery(getCardQuery, token)
	var getCardData struct {
		Card struct {
			ID      string `json:"id"`
			Sprints []struct {
				ID string `json:"id"`
			} `json:"sprints"`
		} `json:"card"`
	}
	json.Unmarshal(getResp.Data, &getCardData)
	assert.Equal(t, 1, len(getCardData.Card.Sprints))

	// Move card to backlog
	moveToBacklogQuery := fmt.Sprintf(`mutation {
		moveCardToBacklog(cardId: "%s") {
			id
			sprints { id }
		}
	}`, cardID)

	moveResp := server.executeQuery(moveToBacklogQuery, token)
	require.Empty(t, moveResp.Errors, "Move to backlog errors: %v", moveResp.Errors)

	var moveData struct {
		MoveCardToBacklog struct {
			ID      string `json:"id"`
			Sprints []struct {
				ID string `json:"id"`
			} `json:"sprints"`
		} `json:"moveCardToBacklog"`
	}
	json.Unmarshal(moveResp.Data, &moveData)

	// Verify card is no longer in any sprint
	assert.Equal(t, 0, len(moveData.MoveCardToBacklog.Sprints))
}

func TestGetBacklogCards(t *testing.T) {
	server := setupSprintTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("getbackloguser", "password123")
	require.NoError(t, err)

	_, boardID, columns := server.setupProject(t, token, "Get Backlog Test", "GBT")

	// Create a sprint
	startDate := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)
	endDate := time.Now().AddDate(0, 0, 13).Format(time.RFC3339)

	createSprintQuery := fmt.Sprintf(`mutation {
		createSprint(input: {
			boardId: "%s"
			name: "Get Backlog Sprint"
			startDate: "%s"
			endDate: "%s"
		}) { id }
	}`, boardID, startDate, endDate)

	sprintResp := server.executeQuery(createSprintQuery, token)
	var sprintData struct {
		CreateSprint struct {
			ID string `json:"id"`
		} `json:"createSprint"`
	}
	json.Unmarshal(sprintResp.Data, &sprintData)
	sprintID := sprintData.CreateSprint.ID

	server.executeQuery(fmt.Sprintf(`mutation { startSprint(id: "%s") { id } }`, sprintID), token)

	// Create two cards in Todo
	todoColumnID := columns["Todo"]

	// Card 1 - will be in sprint
	card1Resp := server.executeQuery(fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "Sprint Card", storyPoints: 3 }) { id }
	}`, todoColumnID), token)
	var card1Data struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(card1Resp.Data, &card1Data)
	card1ID := card1Data.CreateCard.ID

	// Card 2 - will stay in backlog
	card2Resp := server.executeQuery(fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "Backlog Card", storyPoints: 5 }) { id }
	}`, todoColumnID), token)
	var card2Data struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(card2Resp.Data, &card2Data)

	// Add only card1 to sprint
	server.executeQuery(fmt.Sprintf(`mutation {
		addCardToSprint(input: { cardId: "%s", sprintId: "%s" }) { id }
	}`, card1ID, sprintID), token)

	// Query backlog cards
	backlogQuery := fmt.Sprintf(`query {
		backlogCards(boardId: "%s") {
			id
			title
			storyPoints
		}
	}`, boardID)

	backlogResp := server.executeQuery(backlogQuery, token)
	require.Empty(t, backlogResp.Errors, "Backlog query errors: %v", backlogResp.Errors)

	var backlogData struct {
		BacklogCards []struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			StoryPoints int    `json:"storyPoints"`
		} `json:"backlogCards"`
	}
	json.Unmarshal(backlogResp.Data, &backlogData)

	// Only card2 should be in backlog (not in any sprint)
	assert.Equal(t, 1, len(backlogData.BacklogCards))
	assert.Equal(t, "Backlog Card", backlogData.BacklogCards[0].Title)
	assert.Equal(t, 5, backlogData.BacklogCards[0].StoryPoints)
}

func TestCannotDeleteBacklogColumn(t *testing.T) {
	server := setupSprintTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("deletebackloguser", "password123")
	require.NoError(t, err)

	_, _, columns := server.setupProject(t, token, "Delete Backlog Test", "DBT")

	backlogColumnID := columns["Backlog"]

	// Try to delete backlog column
	deleteQuery := fmt.Sprintf(`mutation { deleteColumn(id: "%s") }`, backlogColumnID)
	deleteResp := server.executeQuery(deleteQuery, token)

	// Should get an error
	assert.NotEmpty(t, deleteResp.Errors, "Expected error when deleting backlog column")
}

func TestSprintStats(t *testing.T) {
	server := setupSprintTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("statsuser", "password123")
	require.NoError(t, err)

	_, boardID, columns := server.setupProject(t, token, "Sprint Stats Test", "SST")

	// First, mark Done column as isDone
	doneColumnID := columns["Done"]
	server.executeQuery(fmt.Sprintf(`mutation {
		updateColumn(input: { id: "%s", isDone: true }) { id isDone }
	}`, doneColumnID), token)

	// Create and start a sprint
	startDate := time.Now().AddDate(0, 0, -7).Format(time.RFC3339)
	endDate := time.Now().AddDate(0, 0, 7).Format(time.RFC3339)

	createSprintQuery := fmt.Sprintf(`mutation {
		createSprint(input: {
			boardId: "%s"
			name: "Stats Sprint"
			startDate: "%s"
			endDate: "%s"
		}) { id }
	}`, boardID, startDate, endDate)

	sprintResp := server.executeQuery(createSprintQuery, token)
	var sprintData struct {
		CreateSprint struct {
			ID string `json:"id"`
		} `json:"createSprint"`
	}
	json.Unmarshal(sprintResp.Data, &sprintData)
	sprintID := sprintData.CreateSprint.ID

	server.executeQuery(fmt.Sprintf(`mutation { startSprint(id: "%s") { id } }`, sprintID), token)

	todoColumnID := columns["Todo"]

	// Create cards and add to sprint
	createCard1 := fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "Card 1", storyPoints: 3 }) { id }
	}`, todoColumnID)
	card1Resp := server.executeQuery(createCard1, token)
	var card1Data struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(card1Resp.Data, &card1Data)

	createCard2 := fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "Card 2", storyPoints: 5 }) { id }
	}`, todoColumnID)
	card2Resp := server.executeQuery(createCard2, token)
	var card2Data struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(card2Resp.Data, &card2Data)

	// Add cards to sprint
	server.executeQuery(fmt.Sprintf(`mutation {
		addCardToSprint(input: { cardId: "%s", sprintId: "%s" }) { id }
	}`, card1Data.CreateCard.ID, sprintID), token)

	server.executeQuery(fmt.Sprintf(`mutation {
		addCardToSprint(input: { cardId: "%s", sprintId: "%s" }) { id }
	}`, card2Data.CreateCard.ID, sprintID), token)

	// Move one card to Done
	server.executeQuery(fmt.Sprintf(`mutation {
		moveCard(input: { cardId: "%s", targetColumnId: "%s" }) { id }
	}`, card1Data.CreateCard.ID, doneColumnID), token)

	// Query sprint stats
	statsQuery := fmt.Sprintf(`query {
		sprintStats(sprintId: "%s") {
			totalCards
			completedCards
			totalStoryPoints
			completedStoryPoints
		}
	}`, sprintID)

	statsResp := server.executeQuery(statsQuery, token)
	require.Empty(t, statsResp.Errors, "Sprint stats errors: %v", statsResp.Errors)

	var statsData struct {
		SprintStats struct {
			TotalCards           int `json:"totalCards"`
			CompletedCards       int `json:"completedCards"`
			TotalStoryPoints     int `json:"totalStoryPoints"`
			CompletedStoryPoints int `json:"completedStoryPoints"`
		} `json:"sprintStats"`
	}
	json.Unmarshal(statsResp.Data, &statsData)

	assert.Equal(t, 2, statsData.SprintStats.TotalCards)
	assert.Equal(t, 1, statsData.SprintStats.CompletedCards)
	assert.Equal(t, 8, statsData.SprintStats.TotalStoryPoints)
	assert.Equal(t, 3, statsData.SprintStats.CompletedStoryPoints)
}

func TestCompleteSprint(t *testing.T) {
	server := setupSprintTestServer(t)
	defer server.cleanup()

	token, err := server.registerUser("completeuser", "password123")
	require.NoError(t, err)

	_, boardID, columns := server.setupProject(t, token, "Complete Sprint Test", "CST")

	// Create and start a sprint
	startDate := time.Now().AddDate(0, 0, -14).Format(time.RFC3339)
	endDate := time.Now().AddDate(0, 0, -1).Format(time.RFC3339)

	createSprintQuery := fmt.Sprintf(`mutation {
		createSprint(input: {
			boardId: "%s"
			name: "Sprint to Complete"
			startDate: "%s"
			endDate: "%s"
		}) { id }
	}`, boardID, startDate, endDate)

	sprintResp := server.executeQuery(createSprintQuery, token)
	var sprintData struct {
		CreateSprint struct {
			ID string `json:"id"`
		} `json:"createSprint"`
	}
	json.Unmarshal(sprintResp.Data, &sprintData)
	sprintID := sprintData.CreateSprint.ID

	server.executeQuery(fmt.Sprintf(`mutation { startSprint(id: "%s") { id } }`, sprintID), token)

	// Create a card and add to sprint
	todoColumnID := columns["Todo"]
	createCardQuery := fmt.Sprintf(`mutation {
		createCard(input: { columnId: "%s", title: "Incomplete Card" }) { id }
	}`, todoColumnID)
	cardResp := server.executeQuery(createCardQuery, token)
	var cardData struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(cardResp.Data, &cardData)
	cardID := cardData.CreateCard.ID

	server.executeQuery(fmt.Sprintf(`mutation {
		addCardToSprint(input: { cardId: "%s", sprintId: "%s" }) { id }
	}`, cardID, sprintID), token)

	// Complete sprint with moveIncompleteToBacklog = true
	completeQuery := fmt.Sprintf(`mutation {
		completeSprint(id: "%s", moveIncompleteToBacklog: true) {
			id
			status
		}
	}`, sprintID)

	completeResp := server.executeQuery(completeQuery, token)
	require.Empty(t, completeResp.Errors, "Complete sprint errors: %v", completeResp.Errors)

	var completeData struct {
		CompleteSprint struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"completeSprint"`
	}
	json.Unmarshal(completeResp.Data, &completeData)
	assert.Equal(t, "CLOSED", completeData.CompleteSprint.Status)

	// Verify card is no longer in sprint (moved to backlog)
	getCardQuery := fmt.Sprintf(`query { card(id: "%s") { id sprints { id } } }`, cardID)
	getResp := server.executeQuery(getCardQuery, token)
	var getCardData struct {
		Card struct {
			ID      string `json:"id"`
			Sprints []struct {
				ID string `json:"id"`
			} `json:"sprints"`
		} `json:"card"`
	}
	json.Unmarshal(getResp.Data, &getCardData)

	// Card should not be in the completed sprint anymore
	inSprint := false
	for _, s := range getCardData.Card.Sprints {
		if s.ID == sprintID {
			inSprint = true
			break
		}
	}
	assert.False(t, inSprint, "Card should not be in completed sprint after moveIncompleteToBacklog=true")
}
