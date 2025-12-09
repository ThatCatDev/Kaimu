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
	"github.com/thatcatdev/pulse-backend/config"
	"github.com/thatcatdev/pulse-backend/graph"
	"github.com/thatcatdev/pulse-backend/graph/generated"
	"github.com/thatcatdev/pulse-backend/http/middleware"
	boardRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/board"
	columnRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/board_column"
	cardRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/card"
	cardTagRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/card_tag"
	orgRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/organization"
	memberRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/organization_member"
	projectRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	tagRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/tag"
	userRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"github.com/thatcatdev/pulse-backend/internal/directives"
	"github.com/thatcatdev/pulse-backend/internal/services/auth"
	boardService "github.com/thatcatdev/pulse-backend/internal/services/board"
	cardService "github.com/thatcatdev/pulse-backend/internal/services/card"
	orgService "github.com/thatcatdev/pulse-backend/internal/services/organization"
	projectService "github.com/thatcatdev/pulse-backend/internal/services/project"
	tagService "github.com/thatcatdev/pulse-backend/internal/services/tag"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type OrgProjectTestServer struct {
	handler http.Handler
	db      *gorm.DB
}

func setupOrgProjectTestServer(t *testing.T) *OrgProjectTestServer {
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

	// Run migrations
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
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS board_columns (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			color VARCHAR(7),
			wip_limit INTEGER,
			is_backlog BOOLEAN NOT NULL DEFAULT FALSE,
			is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS tags (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			color VARCHAR(7) NOT NULL DEFAULT '#6366f1',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(project_id, name)
		);
		CREATE TABLE IF NOT EXISTS cards (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			column_id UUID NOT NULL REFERENCES board_columns(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			position FLOAT NOT NULL DEFAULT 0,
			priority VARCHAR(20) NOT NULL DEFAULT 'NONE',
			due_date TIMESTAMP WITH TIME ZONE,
			assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
			created_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS card_tags (
			card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (card_id, tag_id)
		);
	`).Error
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clean up tables before test
	testDB.Exec("DELETE FROM card_tags")
	testDB.Exec("DELETE FROM cards")
	testDB.Exec("DELETE FROM tags")
	testDB.Exec("DELETE FROM board_columns")
	testDB.Exec("DELETE FROM boards")
	testDB.Exec("DELETE FROM projects")
	testDB.Exec("DELETE FROM organization_members")
	testDB.Exec("DELETE FROM organizations")
	testDB.Exec("DELETE FROM users")

	// Create repositories
	userRepository := userRepo.NewRepository(testDB)
	orgRepository := orgRepo.NewRepository(testDB)
	memberRepository := memberRepo.NewRepository(testDB)
	projectRepository := projectRepo.NewRepository(testDB)
	boardRepository := boardRepo.NewRepository(testDB)
	columnRepository := columnRepo.NewRepository(testDB)
	cardRepository := cardRepo.NewRepository(testDB)
	cardTagRepository := cardTagRepo.NewRepository(testDB)
	tagRepository := tagRepo.NewRepository(testDB)

	// Create services
	authSvc := auth.NewService(userRepository, "test-jwt-secret", 24)
	orgSvc := orgService.NewService(orgRepository, memberRepository, userRepository)
	projSvc := projectService.NewService(projectRepository, orgRepository)
	boardSvc := boardService.NewService(boardRepository, columnRepository, projectRepository)
	cardSvc := cardService.NewService(cardRepository, columnRepository, boardRepository, tagRepository, cardTagRepository)
	tagSvc := tagService.NewService(tagRepository, projectRepository)

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
	}

	// Create GraphQL handler
	gqlConfig := generated.Config{
		Resolvers:  resolver,
		Directives: directives.GetDirectives(),
	}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(gqlConfig))

	// Wrap with auth middleware
	wrappedHandler := middleware.AuthMiddleware(authSvc)(srv)

	return &OrgProjectTestServer{
		handler: wrappedHandler,
		db:      testDB,
	}
}

func (ts *OrgProjectTestServer) cleanup(t *testing.T) {
	ts.db.Exec("DELETE FROM card_tags")
	ts.db.Exec("DELETE FROM cards")
	ts.db.Exec("DELETE FROM tags")
	ts.db.Exec("DELETE FROM board_columns")
	ts.db.Exec("DELETE FROM boards")
	ts.db.Exec("DELETE FROM projects")
	ts.db.Exec("DELETE FROM organization_members")
	ts.db.Exec("DELETE FROM organizations")
	ts.db.Exec("DELETE FROM users")
}

func (ts *OrgProjectTestServer) executeGraphQL(t *testing.T, query string, cookies []*http.Cookie) (*GraphQLResponse, []*http.Cookie) {
	reqBody := GraphQLRequest{Query: query}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	rr := httptest.NewRecorder()
	ts.handler.ServeHTTP(rr, req)

	var resp GraphQLResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err, "Failed to parse response: %s", rr.Body.String())

	return &resp, rr.Result().Cookies()
}

func (ts *OrgProjectTestServer) registerUser(t *testing.T, username, password string) []*http.Cookie {
	query := fmt.Sprintf(`mutation {
		register(input: {username: "%s", password: "%s"}) {
			user { id }
		}
	}`, username, password)

	resp, cookies := ts.executeGraphQL(t, query, nil)
	require.Empty(t, resp.Errors, "Registration failed: %v", resp.Errors)
	return cookies
}

// Organization Tests

func TestIntegration_CreateOrganization_Success(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "orgowner", "password123")

	query := `mutation {
		createOrganization(input: {name: "Test Organization", description: "A test org"}) {
			id
			name
			slug
			description
			owner {
				username
			}
		}
	}`

	resp, _ := ts.executeGraphQL(t, query, cookies)

	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var data struct {
		CreateOrganization struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
			Owner       struct {
				Username string `json:"username"`
			} `json:"owner"`
		} `json:"createOrganization"`
	}
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)

	assert.NotEmpty(t, data.CreateOrganization.ID)
	assert.Equal(t, "Test Organization", data.CreateOrganization.Name)
	assert.Equal(t, "test-organization", data.CreateOrganization.Slug)
	assert.Equal(t, "A test org", data.CreateOrganization.Description)
	assert.Equal(t, "orgowner", data.CreateOrganization.Owner.Username)
}

func TestIntegration_CreateOrganization_Unauthenticated(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	query := `mutation {
		createOrganization(input: {name: "Unauthorized Org"}) {
			id
		}
	}`

	resp, _ := ts.executeGraphQL(t, query, nil)

	assert.NotEmpty(t, resp.Errors, "Expected error for unauthenticated user")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestIntegration_GetOrganizations_Success(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "multiorgowner", "password123")

	// Create multiple organizations
	createOrg := func(name string) {
		query := fmt.Sprintf(`mutation {
			createOrganization(input: {name: "%s"}) {
				id
			}
		}`, name)
		resp, _ := ts.executeGraphQL(t, query, cookies)
		require.Empty(t, resp.Errors)
	}

	createOrg("Org One")
	createOrg("Org Two")
	createOrg("Org Three")

	// Get all organizations
	query := `query {
		organizations {
			id
			name
		}
	}`

	resp, _ := ts.executeGraphQL(t, query, cookies)

	assert.Empty(t, resp.Errors)

	var data struct {
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organizations"`
	}
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)

	assert.Len(t, data.Organizations, 3)
}

func TestIntegration_GetOrganizations_WithProjects(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "orglistowner", "password123")

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: {name: "Org With Projects"}) {
			id
		}
	}`
	resp, _ := ts.executeGraphQL(t, createOrgQuery, cookies)
	require.Empty(t, resp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(resp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// Create projects in the organization
	createProject := func(name, key string) {
		query := fmt.Sprintf(`mutation {
			createProject(input: {organizationId: "%s", name: "%s", key: "%s"}) {
				id
			}
		}`, orgID, name, key)
		resp, _ := ts.executeGraphQL(t, query, cookies)
		require.Empty(t, resp.Errors, "Failed to create project: %v", resp.Errors)
	}

	createProject("Project One", "PROJA")
	createProject("Project Two", "PROJB")

	// Get all organizations - should include projects
	query := `query {
		organizations {
			id
			name
			projects {
				id
				name
				key
			}
		}
	}`

	resp, _ = ts.executeGraphQL(t, query, cookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var data struct {
		Organizations []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Projects []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Key  string `json:"key"`
			} `json:"projects"`
		} `json:"organizations"`
	}
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)

	require.Len(t, data.Organizations, 1)
	assert.Equal(t, "Org With Projects", data.Organizations[0].Name)
	assert.Len(t, data.Organizations[0].Projects, 2)

	// Verify project details
	projectNames := []string{data.Organizations[0].Projects[0].Name, data.Organizations[0].Projects[1].Name}
	assert.Contains(t, projectNames, "Project One")
	assert.Contains(t, projectNames, "Project Two")
}

func TestIntegration_GetOrganization_WithProjects(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "projowner", "password123")

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: {name: "Project Org"}) {
			id
		}
	}`
	resp, _ := ts.executeGraphQL(t, createOrgQuery, cookies)
	require.Empty(t, resp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(resp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// Create projects
	createProject := func(name, key string) {
		query := fmt.Sprintf(`mutation {
			createProject(input: {organizationId: "%s", name: "%s", key: "%s"}) {
				id
			}
		}`, orgID, name, key)
		resp, _ := ts.executeGraphQL(t, query, cookies)
		require.Empty(t, resp.Errors, "Failed to create project: %v", resp.Errors)
	}

	createProject("Project Alpha", "ALPHA")
	createProject("Project Beta", "BETA")

	// Get organization with projects
	query := fmt.Sprintf(`query {
		organization(id: "%s") {
			id
			name
			projects {
				id
				name
				key
			}
		}
	}`, orgID)

	resp, _ = ts.executeGraphQL(t, query, cookies)
	assert.Empty(t, resp.Errors)

	var data struct {
		Organization struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Projects []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Key  string `json:"key"`
			} `json:"projects"`
		} `json:"organization"`
	}
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)

	assert.Equal(t, "Project Org", data.Organization.Name)
	assert.Len(t, data.Organization.Projects, 2)
}

// Project Tests

func TestIntegration_CreateProject_Success(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "projectcreator", "password123")

	// Create organization first
	createOrgQuery := `mutation {
		createOrganization(input: {name: "Dev Org"}) {
			id
		}
	}`
	resp, _ := ts.executeGraphQL(t, createOrgQuery, cookies)
	require.Empty(t, resp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(resp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// Create project
	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: {organizationId: "%s", name: "My Project", key: "MYPRJ", description: "A cool project"}) {
			id
			name
			key
			description
			organization {
				id
				name
			}
		}
	}`, orgID)

	resp, _ = ts.executeGraphQL(t, createProjectQuery, cookies)

	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var projectData struct {
		CreateProject struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Key          string `json:"key"`
			Description  string `json:"description"`
			Organization struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"organization"`
		} `json:"createProject"`
	}
	err := json.Unmarshal(resp.Data, &projectData)
	require.NoError(t, err)

	assert.NotEmpty(t, projectData.CreateProject.ID)
	assert.Equal(t, "My Project", projectData.CreateProject.Name)
	assert.Equal(t, "MYPRJ", projectData.CreateProject.Key)
	assert.Equal(t, "A cool project", projectData.CreateProject.Description)
	assert.Equal(t, orgID, projectData.CreateProject.Organization.ID)
}

func TestIntegration_CreateProject_DuplicateKey(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "dupkeyowner", "password123")

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: {name: "Dup Key Org"}) {
			id
		}
	}`
	resp, _ := ts.executeGraphQL(t, createOrgQuery, cookies)
	require.Empty(t, resp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(resp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// Create first project
	query := fmt.Sprintf(`mutation {
		createProject(input: {organizationId: "%s", name: "First Project", key: "DUP"}) {
			id
		}
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, query, cookies)
	require.Empty(t, resp.Errors)

	// Try to create second project with same key
	query = fmt.Sprintf(`mutation {
		createProject(input: {organizationId: "%s", name: "Second Project", key: "DUP"}) {
			id
		}
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, query, cookies)

	assert.NotEmpty(t, resp.Errors, "Expected error for duplicate key")
	assert.Contains(t, resp.Errors[0].Message, "already taken")
}

func TestIntegration_CreateProject_InvalidKey(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "invalidkeyowner", "password123")

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: {name: "Invalid Key Org"}) {
			id
		}
	}`
	resp, _ := ts.executeGraphQL(t, createOrgQuery, cookies)
	require.Empty(t, resp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(resp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// Try to create project with invalid key (too short)
	query := fmt.Sprintf(`mutation {
		createProject(input: {organizationId: "%s", name: "Bad Key Project", key: "A"}) {
			id
		}
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, query, cookies)

	assert.NotEmpty(t, resp.Errors, "Expected error for invalid key")
	assert.Contains(t, resp.Errors[0].Message, "2-10 uppercase")
}

func TestIntegration_CreateProject_Unauthenticated(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	// First create a user and org
	cookies := ts.registerUser(t, "tempowner", "password123")

	createOrgQuery := `mutation {
		createOrganization(input: {name: "Temp Org"}) {
			id
		}
	}`
	resp, _ := ts.executeGraphQL(t, createOrgQuery, cookies)
	require.Empty(t, resp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(resp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// Try to create project without authentication
	query := fmt.Sprintf(`mutation {
		createProject(input: {organizationId: "%s", name: "Unauthorized Project", key: "NOAUTH"}) {
			id
		}
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, query, nil)

	assert.NotEmpty(t, resp.Errors, "Expected error for unauthenticated user")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestIntegration_GetProject_Success(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "getprojowner", "password123")

	// Create organization
	createOrgQuery := `mutation {
		createOrganization(input: {name: "Get Proj Org"}) {
			id
		}
	}`
	resp, _ := ts.executeGraphQL(t, createOrgQuery, cookies)
	require.Empty(t, resp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(resp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// Create project
	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: {organizationId: "%s", name: "Fetch Me", key: "FETCH"}) {
			id
		}
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, createProjectQuery, cookies)
	require.Empty(t, resp.Errors)

	var projectData struct {
		CreateProject struct {
			ID string `json:"id"`
		} `json:"createProject"`
	}
	json.Unmarshal(resp.Data, &projectData)
	projectID := projectData.CreateProject.ID

	// Get project
	getProjectQuery := fmt.Sprintf(`query {
		project(id: "%s") {
			id
			name
			key
			organization {
				id
				name
			}
		}
	}`, projectID)

	resp, _ = ts.executeGraphQL(t, getProjectQuery, cookies)

	assert.Empty(t, resp.Errors)

	var getData struct {
		Project struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Key          string `json:"key"`
			Organization struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"organization"`
		} `json:"project"`
	}
	err := json.Unmarshal(resp.Data, &getData)
	require.NoError(t, err)

	assert.Equal(t, projectID, getData.Project.ID)
	assert.Equal(t, "Fetch Me", getData.Project.Name)
	assert.Equal(t, "FETCH", getData.Project.Key)
	assert.Equal(t, orgID, getData.Project.Organization.ID)
}

func TestIntegration_FullOrgProjectFlow(t *testing.T) {
	ts := setupOrgProjectTestServer(t)
	defer ts.cleanup(t)

	// 1. Register user
	cookies := ts.registerUser(t, "fullflowuser", "password123")

	// 2. Create organization
	createOrgQuery := `mutation {
		createOrganization(input: {name: "Full Flow Org", description: "Testing full flow"}) {
			id
			name
			slug
		}
	}`
	resp, _ := ts.executeGraphQL(t, createOrgQuery, cookies)
	assert.Empty(t, resp.Errors)

	var orgData struct {
		CreateOrganization struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(resp.Data, &orgData)
	orgID := orgData.CreateOrganization.ID

	// 3. Create multiple projects
	projectIDs := make([]string, 0)
	for _, p := range []struct{ name, key string }{
		{"Backend API", "API"},
		{"Frontend App", "WEB"},
		{"Mobile App", "MOB"},
	} {
		query := fmt.Sprintf(`mutation {
			createProject(input: {organizationId: "%s", name: "%s", key: "%s"}) {
				id
				name
				key
			}
		}`, orgID, p.name, p.key)

		resp, _ = ts.executeGraphQL(t, query, cookies)
		assert.Empty(t, resp.Errors)

		var projData struct {
			CreateProject struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Key  string `json:"key"`
			} `json:"createProject"`
		}
		json.Unmarshal(resp.Data, &projData)
		projectIDs = append(projectIDs, projData.CreateProject.ID)
	}

	// 4. Verify organizations list shows our org
	orgsQuery := `query { organizations { id name } }`
	resp, _ = ts.executeGraphQL(t, orgsQuery, cookies)
	assert.Empty(t, resp.Errors)

	var orgsData struct {
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organizations"`
	}
	json.Unmarshal(resp.Data, &orgsData)
	assert.Len(t, orgsData.Organizations, 1)
	assert.Equal(t, "Full Flow Org", orgsData.Organizations[0].Name)

	// 5. Verify organization detail shows all projects
	orgDetailQuery := fmt.Sprintf(`query {
		organization(id: "%s") {
			id
			name
			projects {
				id
				name
				key
			}
		}
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, orgDetailQuery, cookies)
	assert.Empty(t, resp.Errors)

	var orgDetailData struct {
		Organization struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Projects []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Key  string `json:"key"`
			} `json:"projects"`
		} `json:"organization"`
	}
	json.Unmarshal(resp.Data, &orgDetailData)
	assert.Len(t, orgDetailData.Organization.Projects, 3)

	// 6. Verify each project can be fetched individually
	for i, projID := range projectIDs {
		query := fmt.Sprintf(`query { project(id: "%s") { id name organization { id } } }`, projID)
		resp, _ = ts.executeGraphQL(t, query, cookies)
		assert.Empty(t, resp.Errors, "Failed to fetch project %d", i)
	}
}
