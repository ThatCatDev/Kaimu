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
	invRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/invitation"
	memberRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/organization_member"
	orgRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/organization"
	permRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/permission"
	projectRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	projectMemberRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/project_member"
	roleRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/role"
	rolePermRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/role_permission"
	tagRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/tag"
	userRepo "github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"github.com/thatcatdev/pulse-backend/internal/directives"
	"github.com/thatcatdev/pulse-backend/internal/services/auth"
	boardService "github.com/thatcatdev/pulse-backend/internal/services/board"
	cardService "github.com/thatcatdev/pulse-backend/internal/services/card"
	invitationSvc "github.com/thatcatdev/pulse-backend/internal/services/invitation"
	orgService "github.com/thatcatdev/pulse-backend/internal/services/organization"
	projectService "github.com/thatcatdev/pulse-backend/internal/services/project"
	rbacSvc "github.com/thatcatdev/pulse-backend/internal/services/rbac"
	tagService "github.com/thatcatdev/pulse-backend/internal/services/tag"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RBACTestServer struct {
	handler http.Handler
	db      *gorm.DB
}

func setupRBACTestServer(t *testing.T) *RBACTestServer {
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

	// Clean up any existing tables first to ensure clean schema
	cleanupRBACTables(testDB)

	// Run migrations for RBAC tables
	err = testDB.Exec(`
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		-- Users table
		CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255),
			display_name VARCHAR(255),
			avatar_url TEXT,
			password_hash VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Organizations table
		CREATE TABLE organizations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL UNIQUE,
			description TEXT,
			owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Permissions table
		CREATE TABLE permissions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			code VARCHAR(100) NOT NULL UNIQUE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			resource_type VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Roles table
		CREATE TABLE roles (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			is_system BOOLEAN NOT NULL DEFAULT FALSE,
			scope VARCHAR(50) NOT NULL DEFAULT 'organization',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT unique_role_name_per_org UNIQUE (organization_id, name)
		);

		-- Role permissions junction table
		CREATE TABLE role_permissions (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
			permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT unique_role_permission UNIQUE (role_id, permission_id)
		);

		-- Organization members table (must be created after roles)
		CREATE TABLE organization_members (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL DEFAULT 'member',
			role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(organization_id, user_id)
		);

		-- Projects table
		CREATE TABLE projects (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			key VARCHAR(10) NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(organization_id, key)
		);

		-- Project members table
		CREATE TABLE project_members (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT unique_project_member UNIQUE (project_id, user_id)
		);

		-- Invitations table
		CREATE TABLE invitations (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			email VARCHAR(255) NOT NULL,
			role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
			invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) NOT NULL UNIQUE,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			accepted_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Boards table (needed for project creation)
		CREATE TABLE boards (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			is_default BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_by UUID REFERENCES users(id) ON DELETE SET NULL
		);

		-- Board columns table (needed for board creation)
		CREATE TABLE board_columns (
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

		-- Tags table (for card tagging)
		CREATE TABLE tags (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			name VARCHAR(100) NOT NULL,
			color VARCHAR(7),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(project_id, name)
		);

		-- Card priority enum
		DO $$ BEGIN
			CREATE TYPE card_priority AS ENUM ('none', 'low', 'medium', 'high', 'urgent');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;

		-- Cards table
		CREATE TABLE cards (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			column_id UUID NOT NULL REFERENCES board_columns(id) ON DELETE CASCADE,
			board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
			title VARCHAR(500) NOT NULL,
			description TEXT,
			position FLOAT NOT NULL DEFAULT 0,
			priority card_priority NOT NULL DEFAULT 'none',
			assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
			due_date TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			created_by UUID REFERENCES users(id) ON DELETE SET NULL
		);

		-- Card tags junction table
		CREATE TABLE card_tags (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(card_id, tag_id)
		);
	`).Error
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed permissions and system roles
	seedRBACData(t, testDB)

	// Create repositories
	userRepository := userRepo.NewRepository(testDB)
	orgRepository := orgRepo.NewRepository(testDB)
	memberRepository := memberRepo.NewRepository(testDB)
	projectRepository := projectRepo.NewRepository(testDB)
	permRepository := permRepo.NewRepository(testDB)
	roleRepository := roleRepo.NewRepository(testDB)
	rolePermRepository := rolePermRepo.NewRepository(testDB)
	projectMemberRepository := projectMemberRepo.NewRepository(testDB)
	invitationRepository := invRepo.NewRepository(testDB)
	boardRepository := boardRepo.NewRepository(testDB)
	columnRepository := columnRepo.NewRepository(testDB)
	cardRepository := cardRepo.NewRepository(testDB)
	tagRepository := tagRepo.NewRepository(testDB)
	cardTagRepository := cardTagRepo.NewRepository(testDB)

	// Create services
	authSvc := auth.NewService(userRepository, "test-jwt-secret", 24)
	orgSvc := orgService.NewService(orgRepository, memberRepository, userRepository)
	projSvc := projectService.NewService(projectRepository, orgRepository)
	boardSvc := boardService.NewService(boardRepository, columnRepository, projectRepository)
	cardSvc := cardService.NewService(cardRepository, columnRepository, boardRepository, tagRepository, cardTagRepository)
	tagSvc := tagService.NewService(tagRepository, projectRepository)
	rbacService := rbacSvc.NewService(
		permRepository,
		roleRepository,
		rolePermRepository,
		memberRepository,
		projectMemberRepository,
		projectRepository,
		userRepository,
	)
	invSvc := invitationSvc.NewService(
		invitationRepository,
		orgRepository,
		memberRepository,
		userRepository,
		roleRepository,
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
		RBACService:         rbacService,
		InvitationService:   invSvc,
	}

	// Create GraphQL handler
	gqlConfig := generated.Config{
		Resolvers:  resolver,
		Directives: directives.GetDirectives(),
	}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(gqlConfig))

	// Wrap with auth middleware
	wrappedHandler := middleware.AuthMiddleware(authSvc)(srv)

	return &RBACTestServer{
		handler: wrappedHandler,
		db:      testDB,
	}
}

func cleanupRBACTables(db *gorm.DB) {
	// Drop and recreate tables to ensure schema is up to date
	db.Exec("DROP TABLE IF EXISTS card_tags CASCADE")
	db.Exec("DROP TABLE IF EXISTS cards CASCADE")
	db.Exec("DROP TABLE IF EXISTS tags CASCADE")
	db.Exec("DROP TABLE IF EXISTS board_columns CASCADE")
	db.Exec("DROP TABLE IF EXISTS boards CASCADE")
	db.Exec("DROP TABLE IF EXISTS invitations CASCADE")
	db.Exec("DROP TABLE IF EXISTS project_members CASCADE")
	db.Exec("DROP TABLE IF EXISTS role_permissions CASCADE")
	db.Exec("DROP TABLE IF EXISTS organization_members CASCADE")
	db.Exec("DROP TABLE IF EXISTS projects CASCADE")
	db.Exec("DROP TABLE IF EXISTS roles CASCADE")
	db.Exec("DROP TABLE IF EXISTS permissions CASCADE")
	db.Exec("DROP TABLE IF EXISTS organizations CASCADE")
	db.Exec("DROP TABLE IF EXISTS users CASCADE")
	db.Exec("DROP TYPE IF EXISTS card_priority")
}

func seedRBACData(t *testing.T, db *gorm.DB) {
	// Insert permissions if not exists
	err := db.Exec(`
		INSERT INTO permissions (code, name, description, resource_type) VALUES
		('org:view', 'View Organization', 'Can view organization details', 'organization'),
		('org:manage', 'Manage Organization', 'Can edit organization settings', 'organization'),
		('org:delete', 'Delete Organization', 'Can delete the organization', 'organization'),
		('org:invite', 'Invite Members', 'Can invite new members to organization', 'organization'),
		('org:remove_members', 'Remove Members', 'Can remove members from organization', 'organization'),
		('org:manage_roles', 'Manage Roles', 'Can create and edit custom roles', 'organization'),
		('project:view', 'View Project', 'Can view project details', 'project'),
		('project:create', 'Create Project', 'Can create new projects', 'project'),
		('project:manage', 'Manage Project', 'Can edit project settings', 'project'),
		('project:delete', 'Delete Project', 'Can delete projects', 'project'),
		('project:manage_members', 'Manage Project Members', 'Can add/remove project members', 'project'),
		('board:view', 'View Board', 'Can view board and columns', 'board'),
		('board:create', 'Create Board', 'Can create new boards', 'board'),
		('board:manage', 'Manage Board', 'Can edit board settings and columns', 'board'),
		('board:delete', 'Delete Board', 'Can delete boards', 'board'),
		('card:view', 'View Cards', 'Can view cards on boards', 'card'),
		('card:create', 'Create Cards', 'Can create new cards', 'card'),
		('card:edit', 'Edit Cards', 'Can edit card details', 'card'),
		('card:move', 'Move Cards', 'Can move cards between columns', 'card'),
		('card:delete', 'Delete Cards', 'Can delete cards', 'card'),
		('card:assign', 'Assign Cards', 'Can assign cards to users', 'card')
		ON CONFLICT (code) DO NOTHING
	`).Error
	require.NoError(t, err, "Failed to seed permissions")

	// Insert system roles if not exists
	err = db.Exec(`
		INSERT INTO roles (id, organization_id, name, description, is_system, scope) VALUES
		('00000000-0000-0000-0000-000000000001', NULL, 'Owner', 'Full access to everything', TRUE, 'organization'),
		('00000000-0000-0000-0000-000000000002', NULL, 'Admin', 'Administrative access', TRUE, 'organization'),
		('00000000-0000-0000-0000-000000000003', NULL, 'Member', 'Standard member', TRUE, 'organization'),
		('00000000-0000-0000-0000-000000000004', NULL, 'Viewer', 'Read-only access', TRUE, 'organization')
		ON CONFLICT DO NOTHING
	`).Error
	require.NoError(t, err, "Failed to seed roles")

	// Owner gets all permissions
	err = db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT '00000000-0000-0000-0000-000000000001', id FROM permissions
		ON CONFLICT DO NOTHING
	`).Error
	require.NoError(t, err, "Failed to seed owner permissions")

	// Admin gets all except org:delete and org:manage_roles
	err = db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT '00000000-0000-0000-0000-000000000002', id FROM permissions
		WHERE code NOT IN ('org:delete', 'org:manage_roles')
		ON CONFLICT DO NOTHING
	`).Error
	require.NoError(t, err, "Failed to seed admin permissions")

	// Member gets view + create + edit
	err = db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT '00000000-0000-0000-0000-000000000003', id FROM permissions
		WHERE code IN (
			'org:view',
			'project:view', 'project:create',
			'board:view', 'board:create',
			'card:view', 'card:create', 'card:edit', 'card:move', 'card:assign'
		)
		ON CONFLICT DO NOTHING
	`).Error
	require.NoError(t, err, "Failed to seed member permissions")

	// Viewer gets read-only
	err = db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT '00000000-0000-0000-0000-000000000004', id FROM permissions
		WHERE code IN ('org:view', 'project:view', 'board:view', 'card:view')
		ON CONFLICT DO NOTHING
	`).Error
	require.NoError(t, err, "Failed to seed viewer permissions")
}

func (ts *RBACTestServer) cleanup(t *testing.T) {
	cleanupRBACTables(ts.db)
}

func (ts *RBACTestServer) executeGraphQL(t *testing.T, query string, cookies []*http.Cookie) (*GraphQLResponse, []*http.Cookie) {
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

func (ts *RBACTestServer) registerUser(t *testing.T, username, password string) []*http.Cookie {
	query := fmt.Sprintf(`mutation {
		register(input: {username: "%s", password: "%s"}) {
			user { id }
		}
	}`, username, password)

	resp, cookies := ts.executeGraphQL(t, query, nil)
	require.Empty(t, resp.Errors, "Registration failed: %v", resp.Errors)
	return cookies
}

func (ts *RBACTestServer) createOrganization(t *testing.T, cookies []*http.Cookie, name string) string {
	query := fmt.Sprintf(`mutation {
		createOrganization(input: {name: "%s"}) {
			id
		}
	}`, name)

	resp, _ := ts.executeGraphQL(t, query, cookies)
	require.Empty(t, resp.Errors, "Create org failed: %v", resp.Errors)

	var data struct {
		CreateOrganization struct {
			ID string `json:"id"`
		} `json:"createOrganization"`
	}
	json.Unmarshal(resp.Data, &data)
	return data.CreateOrganization.ID
}

func (ts *RBACTestServer) createProject(t *testing.T, cookies []*http.Cookie, orgID, name, key string) string {
	query := fmt.Sprintf(`mutation {
		createProject(input: {organizationId: "%s", name: "%s", key: "%s"}) {
			id
		}
	}`, orgID, name, key)

	resp, _ := ts.executeGraphQL(t, query, cookies)
	require.Empty(t, resp.Errors, "Create project failed: %v", resp.Errors)

	var data struct {
		CreateProject struct {
			ID string `json:"id"`
		} `json:"createProject"`
	}
	json.Unmarshal(resp.Data, &data)
	return data.CreateProject.ID
}

func (ts *RBACTestServer) getBoard(t *testing.T, cookies []*http.Cookie, projectID string) (boardID string, columnID string) {
	query := fmt.Sprintf(`query {
		boards(projectId: "%s") {
			id
			columns {
				id
				name
			}
		}
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, query, cookies)
	require.Empty(t, resp.Errors, "Get boards failed: %v", resp.Errors)

	var data struct {
		Boards []struct {
			ID      string `json:"id"`
			Columns []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"columns"`
		} `json:"boards"`
	}
	json.Unmarshal(resp.Data, &data)
	require.NotEmpty(t, data.Boards, "Expected at least one board")
	require.NotEmpty(t, data.Boards[0].Columns, "Expected at least one column")
	return data.Boards[0].ID, data.Boards[0].Columns[0].ID
}

func (ts *RBACTestServer) createCard(t *testing.T, cookies []*http.Cookie, columnID, title string) string {
	query := fmt.Sprintf(`mutation {
		createCard(input: {columnId: "%s", title: "%s"}) {
			id
		}
	}`, columnID, title)

	resp, _ := ts.executeGraphQL(t, query, cookies)
	require.Empty(t, resp.Errors, "Create card failed: %v", resp.Errors)

	var data struct {
		CreateCard struct {
			ID string `json:"id"`
		} `json:"createCard"`
	}
	json.Unmarshal(resp.Data, &data)
	return data.CreateCard.ID
}

func (ts *RBACTestServer) inviteAndAccept(t *testing.T, ownerCookies []*http.Cookie, memberCookies []*http.Cookie, orgID, email, roleID string) {
	// Invite
	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "%s"
			roleId: "%s"
		}) { token }
	}`, orgID, email, roleID)
	resp, _ := ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors, "Invite failed: %v", resp.Errors)

	var inviteData struct {
		InviteMember struct {
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)

	// Accept
	acceptQuery := fmt.Sprintf(`mutation { acceptInvitation(token: "%s") { id } }`, inviteData.InviteMember.Token)
	resp, _ = ts.executeGraphQL(t, acceptQuery, memberCookies)
	require.Empty(t, resp.Errors, "Accept failed: %v", resp.Errors)
}

// =============================================================================
// Permission Query Tests
// =============================================================================

func TestRBAC_Permissions_Query(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "permuser", "password123")

	query := `query {
		permissions {
			id
			code
			name
			description
			resourceType
		}
	}`

	resp, _ := ts.executeGraphQL(t, query, cookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var data struct {
		Permissions []struct {
			ID           string  `json:"id"`
			Code         string  `json:"code"`
			Name         string  `json:"name"`
			Description  *string `json:"description"`
			ResourceType string  `json:"resourceType"`
		} `json:"permissions"`
	}
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)

	// Should have all seeded permissions
	assert.GreaterOrEqual(t, len(data.Permissions), 21, "Expected at least 21 permissions")

	// Verify some known permissions exist
	codes := make([]string, len(data.Permissions))
	for i, p := range data.Permissions {
		codes[i] = p.Code
	}
	assert.Contains(t, codes, "org:view")
	assert.Contains(t, codes, "project:create")
	assert.Contains(t, codes, "card:edit")
}

func TestRBAC_HasPermission_Query(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Create owner with org
	ownerCookies := ts.registerUser(t, "hasowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "HasPerm Org")

	// Owner should have org:manage_roles
	query := fmt.Sprintf(`query {
		hasPermission(permission: "org:manage_roles", resourceType: "organization", resourceId: "%s")
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.Empty(t, resp.Errors)

	var data struct {
		HasPermission bool `json:"hasPermission"`
	}
	json.Unmarshal(resp.Data, &data)
	assert.True(t, data.HasPermission, "Owner should have org:manage_roles")
}

func TestRBAC_HasPermission_Denied(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Create owner with org
	ownerCookies := ts.registerUser(t, "permowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Perm Org")

	// Create a viewer - add as member with viewer role
	viewerCookies := ts.registerUser(t, "viewer", "password123")

	// Viewer has no permissions for this org (not a member yet)
	query := fmt.Sprintf(`query {
		hasPermission(permission: "org:manage_roles", resourceType: "organization", resourceId: "%s")
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, viewerCookies)
	assert.Empty(t, resp.Errors)

	var data struct {
		HasPermission bool `json:"hasPermission"`
	}
	json.Unmarshal(resp.Data, &data)
	assert.False(t, data.HasPermission, "Non-member should not have org:manage_roles")
}

func TestRBAC_MyPermissions_Query(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "mypermowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "MyPerm Org")

	query := fmt.Sprintf(`query {
		myPermissions(resourceType: "organization", resourceId: "%s")
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.Empty(t, resp.Errors)

	var data struct {
		MyPermissions []string `json:"myPermissions"`
	}
	json.Unmarshal(resp.Data, &data)

	// Owner should have all org permissions
	assert.Contains(t, data.MyPermissions, "org:view")
	assert.Contains(t, data.MyPermissions, "org:manage")
	assert.Contains(t, data.MyPermissions, "org:delete")
	assert.Contains(t, data.MyPermissions, "org:invite")
	assert.Contains(t, data.MyPermissions, "org:remove_members")
	assert.Contains(t, data.MyPermissions, "org:manage_roles")
}

// =============================================================================
// Role Query Tests
// =============================================================================

func TestRBAC_Roles_Query(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "roleowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Role Org")

	query := fmt.Sprintf(`query {
		roles(organizationId: "%s") {
			id
			name
			description
			isSystem
			scope
			permissions {
				code
			}
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var data struct {
		Roles []struct {
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
			IsSystem    bool    `json:"isSystem"`
			Scope       string  `json:"scope"`
			Permissions []struct {
				Code string `json:"code"`
			} `json:"permissions"`
		} `json:"roles"`
	}
	json.Unmarshal(resp.Data, &data)

	// Should have at least 4 system roles
	assert.GreaterOrEqual(t, len(data.Roles), 4)

	// Find Owner role and verify it has all permissions
	var ownerRole *struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		IsSystem    bool    `json:"isSystem"`
		Scope       string  `json:"scope"`
		Permissions []struct {
			Code string `json:"code"`
		} `json:"permissions"`
	}
	for i := range data.Roles {
		if data.Roles[i].Name == "Owner" {
			ownerRole = &data.Roles[i]
			break
		}
	}
	require.NotNil(t, ownerRole, "Owner role should exist")
	assert.True(t, ownerRole.IsSystem, "Owner should be a system role")
	assert.GreaterOrEqual(t, len(ownerRole.Permissions), 21, "Owner should have all permissions")
}

func TestRBAC_Roles_Unauthorized(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "roleowner2", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Role Org 2")

	// Different user tries to query roles
	otherCookies := ts.registerUser(t, "otheruser", "password123")

	query := fmt.Sprintf(`query {
		roles(organizationId: "%s") {
			id
			name
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, otherCookies)
	assert.NotEmpty(t, resp.Errors, "Expected error for unauthorized user")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_Role_Query(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	cookies := ts.registerUser(t, "singlerole", "password123")

	// Query for the Owner system role
	query := `query {
		role(id: "00000000-0000-0000-0000-000000000001") {
			id
			name
			description
			isSystem
			permissions {
				code
				name
			}
		}
	}`

	resp, _ := ts.executeGraphQL(t, query, cookies)
	assert.Empty(t, resp.Errors)

	var data struct {
		Role struct {
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
			IsSystem    bool    `json:"isSystem"`
			Permissions []struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"permissions"`
		} `json:"role"`
	}
	json.Unmarshal(resp.Data, &data)

	assert.Equal(t, "Owner", data.Role.Name)
	assert.True(t, data.Role.IsSystem)
	assert.GreaterOrEqual(t, len(data.Role.Permissions), 21)
}

// =============================================================================
// Role Mutation Tests
// =============================================================================

func TestRBAC_CreateRole_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "createroleowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CreateRole Org")

	query := fmt.Sprintf(`mutation {
		createRole(input: {
			organizationId: "%s"
			name: "Custom Developer"
			description: "Custom role for developers"
			permissionCodes: ["org:view", "project:view", "project:create", "card:create", "card:edit"]
		}) {
			id
			name
			description
			isSystem
			permissions {
				code
			}
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var data struct {
		CreateRole struct {
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
			IsSystem    bool    `json:"isSystem"`
			Permissions []struct {
				Code string `json:"code"`
			} `json:"permissions"`
		} `json:"createRole"`
	}
	json.Unmarshal(resp.Data, &data)

	assert.NotEmpty(t, data.CreateRole.ID)
	assert.Equal(t, "Custom Developer", data.CreateRole.Name)
	assert.False(t, data.CreateRole.IsSystem)
	assert.Len(t, data.CreateRole.Permissions, 5)
}

func TestRBAC_CreateRole_Unauthorized(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "createrolewowner2", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CreateRole Org 2")

	// Different user tries to create role
	otherCookies := ts.registerUser(t, "otheruser2", "password123")

	query := fmt.Sprintf(`mutation {
		createRole(input: {
			organizationId: "%s"
			name: "Hacker Role"
			permissionCodes: ["org:view"]
		}) {
			id
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, otherCookies)
	assert.NotEmpty(t, resp.Errors, "Expected error for unauthorized user")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_UpdateRole_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "updateroleowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "UpdateRole Org")

	// Create a custom role first
	createQuery := fmt.Sprintf(`mutation {
		createRole(input: {
			organizationId: "%s"
			name: "Temp Role"
			permissionCodes: ["org:view"]
		}) {
			id
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, createQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var createData struct {
		CreateRole struct {
			ID string `json:"id"`
		} `json:"createRole"`
	}
	json.Unmarshal(resp.Data, &createData)
	roleID := createData.CreateRole.ID

	// Update the role
	updateQuery := fmt.Sprintf(`mutation {
		updateRole(input: {
			id: "%s"
			name: "Updated Role"
			description: "Updated description"
			permissionCodes: ["org:view", "project:view"]
		}) {
			id
			name
			description
			permissions {
				code
			}
		}
	}`, roleID)

	resp, _ = ts.executeGraphQL(t, updateQuery, ownerCookies)
	assert.Empty(t, resp.Errors)

	var updateData struct {
		UpdateRole struct {
			ID          string  `json:"id"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
			Permissions []struct {
				Code string `json:"code"`
			} `json:"permissions"`
		} `json:"updateRole"`
	}
	json.Unmarshal(resp.Data, &updateData)

	assert.Equal(t, "Updated Role", updateData.UpdateRole.Name)
	assert.NotNil(t, updateData.UpdateRole.Description)
	assert.Equal(t, "Updated description", *updateData.UpdateRole.Description)
	assert.Len(t, updateData.UpdateRole.Permissions, 2)
}

func TestRBAC_UpdateRole_SystemRoleFails(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "updatesysowner", "password123")
	_ = ts.createOrganization(t, ownerCookies, "UpdateSys Org")

	// Try to update the Owner system role
	query := `mutation {
		updateRole(input: {
			id: "00000000-0000-0000-0000-000000000001"
			name: "Hacked Owner"
		}) {
			id
		}
	}`

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.NotEmpty(t, resp.Errors, "Expected error when modifying system role")
	assert.Contains(t, resp.Errors[0].Message, "system")
}

func TestRBAC_DeleteRole_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "deleteroleowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "DeleteRole Org")

	// Create a custom role
	createQuery := fmt.Sprintf(`mutation {
		createRole(input: {
			organizationId: "%s"
			name: "ToDelete Role"
			permissionCodes: ["org:view"]
		}) {
			id
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, createQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var createData struct {
		CreateRole struct {
			ID string `json:"id"`
		} `json:"createRole"`
	}
	json.Unmarshal(resp.Data, &createData)
	roleID := createData.CreateRole.ID

	// Delete the role
	deleteQuery := fmt.Sprintf(`mutation {
		deleteRole(id: "%s")
	}`, roleID)

	resp, _ = ts.executeGraphQL(t, deleteQuery, ownerCookies)
	assert.Empty(t, resp.Errors)

	var deleteData struct {
		DeleteRole bool `json:"deleteRole"`
	}
	json.Unmarshal(resp.Data, &deleteData)
	assert.True(t, deleteData.DeleteRole)
}

func TestRBAC_DeleteRole_SystemRoleFails(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "delsysowner", "password123")
	_ = ts.createOrganization(t, ownerCookies, "DelSys Org")

	// Try to delete the Owner system role
	query := `mutation {
		deleteRole(id: "00000000-0000-0000-0000-000000000001")
	}`

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.NotEmpty(t, resp.Errors, "Expected error when deleting system role")
	assert.Contains(t, resp.Errors[0].Message, "system")
}

// =============================================================================
// Organization Member Tests
// =============================================================================

func TestRBAC_OrganizationMembers_Query(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "membersowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Members Org")

	query := fmt.Sprintf(`query {
		organizationMembers(organizationId: "%s") {
			id
			legacyRole
			user {
				username
			}
			role {
				name
			}
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var data struct {
		OrganizationMembers []struct {
			ID         string `json:"id"`
			LegacyRole string `json:"legacyRole"`
			User       struct {
				Username string `json:"username"`
			} `json:"user"`
			Role struct {
				Name string `json:"name"`
			} `json:"role"`
		} `json:"organizationMembers"`
	}
	json.Unmarshal(resp.Data, &data)

	assert.Len(t, data.OrganizationMembers, 1)
	assert.Equal(t, "membersowner", data.OrganizationMembers[0].User.Username)
	assert.Equal(t, "Owner", data.OrganizationMembers[0].Role.Name)
}

func TestRBAC_ChangeMemberRole_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Create owner
	ownerCookies := ts.registerUser(t, "changeowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ChangeRole Org")

	// Create another user and invite them
	memberCookies := ts.registerUser(t, "changemember", "password123")

	// Get member user ID
	meQuery := `query { me { id } }`
	resp, _ := ts.executeGraphQL(t, meQuery, memberCookies)
	var meData struct {
		Me struct {
			ID string `json:"id"`
		} `json:"me"`
	}
	json.Unmarshal(resp.Data, &meData)
	memberUserID := meData.Me.ID

	// Owner invites member
	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "changemember@test.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) {
			token
		}
	}`, orgID)

	resp, _ = ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors, "Invite failed: %v", resp.Errors)

	var inviteData struct {
		InviteMember struct {
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)

	// Member accepts invitation
	acceptQuery := fmt.Sprintf(`mutation {
		acceptInvitation(token: "%s") {
			id
		}
	}`, inviteData.InviteMember.Token)

	resp, _ = ts.executeGraphQL(t, acceptQuery, memberCookies)
	require.Empty(t, resp.Errors, "Accept failed: %v", resp.Errors)

	// Owner changes member role to Admin
	changeQuery := fmt.Sprintf(`mutation {
		changeMemberRole(organizationId: "%s", input: {
			userId: "%s"
			roleId: "00000000-0000-0000-0000-000000000002"
		}) {
			id
			role {
				name
			}
		}
	}`, orgID, memberUserID)

	resp, _ = ts.executeGraphQL(t, changeQuery, ownerCookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var changeData struct {
		ChangeMemberRole struct {
			ID   string `json:"id"`
			Role struct {
				Name string `json:"name"`
			} `json:"role"`
		} `json:"changeMemberRole"`
	}
	json.Unmarshal(resp.Data, &changeData)

	assert.Equal(t, "Admin", changeData.ChangeMemberRole.Role.Name)
}

func TestRBAC_RemoveMember_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "removeowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Remove Org")

	// Create and add member
	memberCookies := ts.registerUser(t, "removemember", "password123")

	meQuery := `query { me { id } }`
	resp, _ := ts.executeGraphQL(t, meQuery, memberCookies)
	var meData struct {
		Me struct {
			ID string `json:"id"`
		} `json:"me"`
	}
	json.Unmarshal(resp.Data, &meData)
	memberUserID := meData.Me.ID

	// Invite and accept
	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "removemember@test.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) {
			token
		}
	}`, orgID)

	resp, _ = ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var inviteData struct {
		InviteMember struct {
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)

	acceptQuery := fmt.Sprintf(`mutation {
		acceptInvitation(token: "%s") { id }
	}`, inviteData.InviteMember.Token)
	resp, _ = ts.executeGraphQL(t, acceptQuery, memberCookies)
	require.Empty(t, resp.Errors)

	// Verify member is in org
	membersQuery := fmt.Sprintf(`query {
		organizationMembers(organizationId: "%s") { id }
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, membersQuery, ownerCookies)
	var beforeData struct {
		OrganizationMembers []struct {
			ID string `json:"id"`
		} `json:"organizationMembers"`
	}
	json.Unmarshal(resp.Data, &beforeData)
	assert.Len(t, beforeData.OrganizationMembers, 2)

	// Remove member
	removeQuery := fmt.Sprintf(`mutation {
		removeMember(organizationId: "%s", userId: "%s")
	}`, orgID, memberUserID)

	resp, _ = ts.executeGraphQL(t, removeQuery, ownerCookies)
	assert.Empty(t, resp.Errors)

	var removeData struct {
		RemoveMember bool `json:"removeMember"`
	}
	json.Unmarshal(resp.Data, &removeData)
	assert.True(t, removeData.RemoveMember)

	// Verify member is removed
	resp, _ = ts.executeGraphQL(t, membersQuery, ownerCookies)
	var afterData struct {
		OrganizationMembers []struct {
			ID string `json:"id"`
		} `json:"organizationMembers"`
	}
	json.Unmarshal(resp.Data, &afterData)
	assert.Len(t, afterData.OrganizationMembers, 1)
}

func TestRBAC_RemoveMember_LastOwnerFails(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "lastowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "LastOwner Org")

	// Get owner user ID
	meQuery := `query { me { id } }`
	resp, _ := ts.executeGraphQL(t, meQuery, ownerCookies)
	var meData struct {
		Me struct {
			ID string `json:"id"`
		} `json:"me"`
	}
	json.Unmarshal(resp.Data, &meData)
	ownerUserID := meData.Me.ID

	// Try to remove self (last owner)
	removeQuery := fmt.Sprintf(`mutation {
		removeMember(organizationId: "%s", userId: "%s")
	}`, orgID, ownerUserID)

	resp, _ = ts.executeGraphQL(t, removeQuery, ownerCookies)
	assert.NotEmpty(t, resp.Errors, "Expected error when removing last owner")
	assert.Contains(t, resp.Errors[0].Message, "last owner")
}

// =============================================================================
// Invitation Tests
// =============================================================================

func TestRBAC_InviteMember_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "inviteowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Invite Org")

	query := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "newinvitee@example.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) {
			id
			email
			token
			role {
				name
			}
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var data struct {
		InviteMember struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Token string `json:"token"`
			Role  struct {
				Name string `json:"name"`
			} `json:"role"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &data)

	assert.NotEmpty(t, data.InviteMember.ID)
	assert.Equal(t, "newinvitee@example.com", data.InviteMember.Email)
	assert.NotEmpty(t, data.InviteMember.Token)
	assert.Equal(t, "Member", data.InviteMember.Role.Name)
}

func TestRBAC_Invitations_Query(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "listinviteowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ListInvite Org")

	// Create some invitations
	for i := 1; i <= 3; i++ {
		inviteQuery := fmt.Sprintf(`mutation {
			inviteMember(input: {
				organizationId: "%s"
				email: "invite%d@example.com"
				roleId: "00000000-0000-0000-0000-000000000003"
			}) { id }
		}`, orgID, i)
		ts.executeGraphQL(t, inviteQuery, ownerCookies)
	}

	// Query invitations
	query := fmt.Sprintf(`query {
		invitations(organizationId: "%s") {
			id
			email
			expiresAt
			role {
				name
			}
			invitedBy {
				username
			}
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.Empty(t, resp.Errors)

	var data struct {
		Invitations []struct {
			ID        string `json:"id"`
			Email     string `json:"email"`
			ExpiresAt string `json:"expiresAt"`
			Role      struct {
				Name string `json:"name"`
			} `json:"role"`
			InvitedBy struct {
				Username string `json:"username"`
			} `json:"invitedBy"`
		} `json:"invitations"`
	}
	json.Unmarshal(resp.Data, &data)

	assert.Len(t, data.Invitations, 3)
	assert.Equal(t, "listinviteowner", data.Invitations[0].InvitedBy.Username)
}

func TestRBAC_CancelInvitation_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "cancelowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Cancel Org")

	// Create invitation
	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "tocancel@example.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) {
			id
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var inviteData struct {
		InviteMember struct {
			ID string `json:"id"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)
	inviteID := inviteData.InviteMember.ID

	// Cancel invitation
	cancelQuery := fmt.Sprintf(`mutation {
		cancelInvitation(id: "%s")
	}`, inviteID)

	resp, _ = ts.executeGraphQL(t, cancelQuery, ownerCookies)
	assert.Empty(t, resp.Errors)

	var cancelData struct {
		CancelInvitation bool `json:"cancelInvitation"`
	}
	json.Unmarshal(resp.Data, &cancelData)
	assert.True(t, cancelData.CancelInvitation)

	// Verify invitation is gone
	listQuery := fmt.Sprintf(`query {
		invitations(organizationId: "%s") { id }
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, listQuery, ownerCookies)
	var listData struct {
		Invitations []struct {
			ID string `json:"id"`
		} `json:"invitations"`
	}
	json.Unmarshal(resp.Data, &listData)
	assert.Len(t, listData.Invitations, 0)
}

func TestRBAC_ResendInvitation_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "resendowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Resend Org")

	// Create invitation
	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "toresend@example.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) {
			id
			token
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var inviteData struct {
		InviteMember struct {
			ID    string `json:"id"`
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)
	inviteID := inviteData.InviteMember.ID
	originalToken := inviteData.InviteMember.Token

	// Resend invitation
	resendQuery := fmt.Sprintf(`mutation {
		resendInvitation(id: "%s") {
			id
			email
		}
	}`, inviteID)

	resp, _ = ts.executeGraphQL(t, resendQuery, ownerCookies)
	assert.Empty(t, resp.Errors)

	var resendData struct {
		ResendInvitation struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"resendInvitation"`
	}
	json.Unmarshal(resp.Data, &resendData)
	assert.Equal(t, inviteID, resendData.ResendInvitation.ID)

	// Verify the old token no longer works
	query := fmt.Sprintf(`mutation {
		acceptInvitation(token: "%s") { id }
	}`, originalToken)

	newUserCookies := ts.registerUser(t, "newtokenuser", "password123")
	resp, _ = ts.executeGraphQL(t, query, newUserCookies)
	assert.NotEmpty(t, resp.Errors, "Old token should be invalid")
}

func TestRBAC_AcceptInvitation_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "acceptowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Accept Org")

	// Create invitation
	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "acceptme@example.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) {
			token
		}
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var inviteData struct {
		InviteMember struct {
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)
	token := inviteData.InviteMember.Token

	// New user accepts invitation
	newUserCookies := ts.registerUser(t, "acceptuser", "password123")

	acceptQuery := fmt.Sprintf(`mutation {
		acceptInvitation(token: "%s") {
			id
			name
			slug
		}
	}`, token)

	resp, _ = ts.executeGraphQL(t, acceptQuery, newUserCookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var acceptData struct {
		AcceptInvitation struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"acceptInvitation"`
	}
	json.Unmarshal(resp.Data, &acceptData)

	assert.Equal(t, orgID, acceptData.AcceptInvitation.ID)
	assert.Equal(t, "Accept Org", acceptData.AcceptInvitation.Name)

	// Verify user is now a member
	membersQuery := fmt.Sprintf(`query {
		organizationMembers(organizationId: "%s") {
			user { username }
		}
	}`, orgID)

	resp, _ = ts.executeGraphQL(t, membersQuery, ownerCookies)
	var membersData struct {
		OrganizationMembers []struct {
			User struct {
				Username string `json:"username"`
			} `json:"user"`
		} `json:"organizationMembers"`
	}
	json.Unmarshal(resp.Data, &membersData)

	usernames := make([]string, len(membersData.OrganizationMembers))
	for i, m := range membersData.OrganizationMembers {
		usernames[i] = m.User.Username
	}
	assert.Contains(t, usernames, "acceptuser")
}

func TestRBAC_AcceptInvitation_InvalidToken(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	userCookies := ts.registerUser(t, "invalidtokenuser", "password123")

	query := `mutation {
		acceptInvitation(token: "invalid-token-12345") {
			id
		}
	}`

	resp, _ := ts.executeGraphQL(t, query, userCookies)
	assert.NotEmpty(t, resp.Errors, "Expected error for invalid token")
	assert.Contains(t, resp.Errors[0].Message, "not found")
}

// =============================================================================
// Project Member Tests
// =============================================================================

func TestRBAC_ProjectMembers_Query(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "projmemberowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjMember Org")
	projectID := ts.createProject(t, ownerCookies, orgID, "Test Project", "TEST")

	query := fmt.Sprintf(`query {
		projectMembers(projectId: "%s") {
			id
			user {
				username
			}
			role {
				name
			}
			project {
				name
			}
		}
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, query, ownerCookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var data struct {
		ProjectMembers []struct {
			ID   string `json:"id"`
			User struct {
				Username string `json:"username"`
			} `json:"user"`
			Role *struct {
				Name string `json:"name"`
			} `json:"role"`
			Project struct {
				Name string `json:"name"`
			} `json:"project"`
		} `json:"projectMembers"`
	}
	json.Unmarshal(resp.Data, &data)

	// Initially no project-specific members (permissions come from org)
	assert.Len(t, data.ProjectMembers, 0)
}

func TestRBAC_AssignProjectRole_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "assignprojowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "AssignProj Org")
	projectID := ts.createProject(t, ownerCookies, orgID, "Assign Project", "ASGN")

	// Add a member to org first
	memberCookies := ts.registerUser(t, "projmember", "password123")

	meQuery := `query { me { id } }`
	resp, _ := ts.executeGraphQL(t, meQuery, memberCookies)
	var meData struct {
		Me struct {
			ID string `json:"id"`
		} `json:"me"`
	}
	json.Unmarshal(resp.Data, &meData)
	memberUserID := meData.Me.ID

	// Invite and accept
	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "projmember@test.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) { token }
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var inviteData struct {
		InviteMember struct {
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)

	acceptQuery := fmt.Sprintf(`mutation { acceptInvitation(token: "%s") { id } }`, inviteData.InviteMember.Token)
	resp, _ = ts.executeGraphQL(t, acceptQuery, memberCookies)
	require.Empty(t, resp.Errors)

	// Assign project-specific role
	assignQuery := fmt.Sprintf(`mutation {
		assignProjectRole(input: {
			projectId: "%s"
			userId: "%s"
			roleId: "00000000-0000-0000-0000-000000000002"
		}) {
			id
			user {
				username
			}
			role {
				name
			}
			project {
				name
			}
		}
	}`, projectID, memberUserID)

	resp, _ = ts.executeGraphQL(t, assignQuery, ownerCookies)
	assert.Empty(t, resp.Errors, "Expected no errors, got: %v", resp.Errors)

	var assignData struct {
		AssignProjectRole struct {
			ID   string `json:"id"`
			User struct {
				Username string `json:"username"`
			} `json:"user"`
			Role struct {
				Name string `json:"name"`
			} `json:"role"`
			Project struct {
				Name string `json:"name"`
			} `json:"project"`
		} `json:"assignProjectRole"`
	}
	json.Unmarshal(resp.Data, &assignData)

	assert.Equal(t, "projmember", assignData.AssignProjectRole.User.Username)
	assert.Equal(t, "Admin", assignData.AssignProjectRole.Role.Name)
	assert.Equal(t, "Assign Project", assignData.AssignProjectRole.Project.Name)
}

func TestRBAC_RemoveProjectMember_Success(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "removeprojowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "RemoveProj Org")
	projectID := ts.createProject(t, ownerCookies, orgID, "Remove Project", "REM")

	// Add member to org and assign project role
	memberCookies := ts.registerUser(t, "removeprojmember", "password123")

	meQuery := `query { me { id } }`
	resp, _ := ts.executeGraphQL(t, meQuery, memberCookies)
	var meData struct {
		Me struct {
			ID string `json:"id"`
		} `json:"me"`
	}
	json.Unmarshal(resp.Data, &meData)
	memberUserID := meData.Me.ID

	// Invite and accept
	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "removeprojmember@test.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) { token }
	}`, orgID)
	resp, _ = ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var inviteData struct {
		InviteMember struct {
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)

	acceptQuery := fmt.Sprintf(`mutation { acceptInvitation(token: "%s") { id } }`, inviteData.InviteMember.Token)
	resp, _ = ts.executeGraphQL(t, acceptQuery, memberCookies)
	require.Empty(t, resp.Errors)

	// Assign project role
	assignQuery := fmt.Sprintf(`mutation {
		assignProjectRole(input: {
			projectId: "%s"
			userId: "%s"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) { id }
	}`, projectID, memberUserID)
	resp, _ = ts.executeGraphQL(t, assignQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	// Verify member is in project
	membersQuery := fmt.Sprintf(`query {
		projectMembers(projectId: "%s") { id }
	}`, projectID)
	resp, _ = ts.executeGraphQL(t, membersQuery, ownerCookies)
	var beforeData struct {
		ProjectMembers []struct {
			ID string `json:"id"`
		} `json:"projectMembers"`
	}
	json.Unmarshal(resp.Data, &beforeData)
	assert.Len(t, beforeData.ProjectMembers, 1)

	// Remove from project
	removeQuery := fmt.Sprintf(`mutation {
		removeProjectMember(projectId: "%s", userId: "%s")
	}`, projectID, memberUserID)

	resp, _ = ts.executeGraphQL(t, removeQuery, ownerCookies)
	assert.Empty(t, resp.Errors)

	var removeData struct {
		RemoveProjectMember bool `json:"removeProjectMember"`
	}
	json.Unmarshal(resp.Data, &removeData)
	assert.True(t, removeData.RemoveProjectMember)

	// Verify member is removed from project
	resp, _ = ts.executeGraphQL(t, membersQuery, ownerCookies)
	var afterData struct {
		ProjectMembers []struct {
			ID string `json:"id"`
		} `json:"projectMembers"`
	}
	json.Unmarshal(resp.Data, &afterData)
	assert.Len(t, afterData.ProjectMembers, 0)
}

// =============================================================================
// Permission Enforcement Tests
// =============================================================================

func TestRBAC_ViewerCannotInvite(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Create owner and org
	ownerCookies := ts.registerUser(t, "enforceowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "Enforce Org")

	// Create and add viewer
	viewerCookies := ts.registerUser(t, "enforceviewer", "password123")

	// Invite viewer with Viewer role
	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "enforceviewer@test.com"
			roleId: "00000000-0000-0000-0000-000000000004"
		}) { token }
	}`, orgID)
	resp, _ := ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var inviteData struct {
		InviteMember struct {
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)

	acceptQuery := fmt.Sprintf(`mutation { acceptInvitation(token: "%s") { id } }`, inviteData.InviteMember.Token)
	resp, _ = ts.executeGraphQL(t, acceptQuery, viewerCookies)
	require.Empty(t, resp.Errors)

	// Viewer tries to invite someone
	viewerInviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "hacker@test.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) { id }
	}`, orgID)

	resp, _ = ts.executeGraphQL(t, viewerInviteQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to invite")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCannotManageRoles(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "roleenfowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "RoleEnf Org")

	// Add member
	memberCookies := ts.registerUser(t, "roleenfmember", "password123")

	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "roleenfmember@test.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) { token }
	}`, orgID)
	resp, _ := ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var inviteData struct {
		InviteMember struct {
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)

	acceptQuery := fmt.Sprintf(`mutation { acceptInvitation(token: "%s") { id } }`, inviteData.InviteMember.Token)
	resp, _ = ts.executeGraphQL(t, acceptQuery, memberCookies)
	require.Empty(t, resp.Errors)

	// Member tries to create a role
	createRoleQuery := fmt.Sprintf(`mutation {
		createRole(input: {
			organizationId: "%s"
			name: "Hacker Role"
			permissionCodes: ["org:view"]
		}) { id }
	}`, orgID)

	resp, _ = ts.executeGraphQL(t, createRoleQuery, memberCookies)
	assert.NotEmpty(t, resp.Errors, "Member should not be able to create roles")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_AdminCanInviteButNotManageRoles(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	ownerCookies := ts.registerUser(t, "adminenfowner", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "AdminEnf Org")

	// Add admin
	adminCookies := ts.registerUser(t, "adminenfadmin", "password123")

	inviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "adminenfadmin@test.com"
			roleId: "00000000-0000-0000-0000-000000000002"
		}) { token }
	}`, orgID)
	resp, _ := ts.executeGraphQL(t, inviteQuery, ownerCookies)
	require.Empty(t, resp.Errors)

	var inviteData struct {
		InviteMember struct {
			Token string `json:"token"`
		} `json:"inviteMember"`
	}
	json.Unmarshal(resp.Data, &inviteData)

	acceptQuery := fmt.Sprintf(`mutation { acceptInvitation(token: "%s") { id } }`, inviteData.InviteMember.Token)
	resp, _ = ts.executeGraphQL(t, acceptQuery, adminCookies)
	require.Empty(t, resp.Errors)

	// Admin CAN invite (has org:invite)
	adminInviteQuery := fmt.Sprintf(`mutation {
		inviteMember(input: {
			organizationId: "%s"
			email: "newuser@test.com"
			roleId: "00000000-0000-0000-0000-000000000003"
		}) { id }
	}`, orgID)

	resp, _ = ts.executeGraphQL(t, adminInviteQuery, adminCookies)
	assert.Empty(t, resp.Errors, "Admin should be able to invite")

	// Admin CANNOT create roles (no org:manage_roles)
	createRoleQuery := fmt.Sprintf(`mutation {
		createRole(input: {
			organizationId: "%s"
			name: "Admin Created Role"
			permissionCodes: ["org:view"]
		}) { id }
	}`, orgID)

	resp, _ = ts.executeGraphQL(t, createRoleQuery, adminCookies)
	assert.NotEmpty(t, resp.Errors, "Admin should not be able to create roles")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

// =============================================================================
// Card Permission Enforcement Tests
// =============================================================================

func TestRBAC_ViewerCannotCreateCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "cardowner1", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org1")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project", "CTA")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "cardviewer1", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "cardviewer1@test.com", "00000000-0000-0000-0000-000000000004") // Viewer role

	// Viewer tries to create a card - should fail
	createCardQuery := fmt.Sprintf(`mutation {
		createCard(input: {columnId: "%s", title: "Viewer Card"}) { id }
	}`, columnID)

	resp, _ := ts.executeGraphQL(t, createCardQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to create cards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCanCreateCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "cardowner2", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org2")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project2", "CTB")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "cardmember2", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "cardmember2@test.com", "00000000-0000-0000-0000-000000000003") // Member role

	// Member can create a card
	createCardQuery := fmt.Sprintf(`mutation {
		createCard(input: {columnId: "%s", title: "Member Card"}) { id }
	}`, columnID)

	resp, _ := ts.executeGraphQL(t, createCardQuery, memberCookies)
	assert.Empty(t, resp.Errors, "Member should be able to create cards")
}

func TestRBAC_ViewerCannotUpdateCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org, project and card
	ownerCookies := ts.registerUser(t, "cardowner3", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org3")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project3", "CTC")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)
	cardID := ts.createCard(t, ownerCookies, columnID, "Owner Card")

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "cardviewer3", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "cardviewer3@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer tries to update the card - should fail
	updateCardQuery := fmt.Sprintf(`mutation {
		updateCard(input: {id: "%s", title: "Hacked Title"}) { id }
	}`, cardID)

	resp, _ := ts.executeGraphQL(t, updateCardQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to update cards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCanUpdateCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org, project and card
	ownerCookies := ts.registerUser(t, "cardowner4", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org4")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project4", "CTD")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)
	cardID := ts.createCard(t, ownerCookies, columnID, "Owner Card")

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "cardmember4", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "cardmember4@test.com", "00000000-0000-0000-0000-000000000003")

	// Member can update the card
	updateCardQuery := fmt.Sprintf(`mutation {
		updateCard(input: {id: "%s", title: "Updated Title"}) { id title }
	}`, cardID)

	resp, _ := ts.executeGraphQL(t, updateCardQuery, memberCookies)
	assert.Empty(t, resp.Errors, "Member should be able to update cards")

	var data struct {
		UpdateCard struct {
			Title string `json:"title"`
		} `json:"updateCard"`
	}
	json.Unmarshal(resp.Data, &data)
	assert.Equal(t, "Updated Title", data.UpdateCard.Title)
}

func TestRBAC_ViewerCannotMoveCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org, project and card
	ownerCookies := ts.registerUser(t, "cardowner5", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org5")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project5", "CTE")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)
	cardID := ts.createCard(t, ownerCookies, columnID, "Owner Card")

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "cardviewer5", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "cardviewer5@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer tries to move the card - should fail (moving to same column for simplicity)
	moveCardQuery := fmt.Sprintf(`mutation {
		moveCard(input: {cardId: "%s", targetColumnId: "%s"}) { id }
	}`, cardID, columnID)

	resp, _ := ts.executeGraphQL(t, moveCardQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to move cards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCanMoveCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org, project and card
	ownerCookies := ts.registerUser(t, "cardowner6", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org6")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project6", "CTF")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)
	cardID := ts.createCard(t, ownerCookies, columnID, "Owner Card")

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "cardmember6", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "cardmember6@test.com", "00000000-0000-0000-0000-000000000003")

	// Member can move the card (moving to same column for simplicity)
	moveCardQuery := fmt.Sprintf(`mutation {
		moveCard(input: {cardId: "%s", targetColumnId: "%s"}) { id }
	}`, cardID, columnID)

	resp, _ := ts.executeGraphQL(t, moveCardQuery, memberCookies)
	assert.Empty(t, resp.Errors, "Member should be able to move cards")
}

func TestRBAC_ViewerCannotDeleteCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org, project and card
	ownerCookies := ts.registerUser(t, "cardowner7", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org7")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project7", "CTG")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)
	cardID := ts.createCard(t, ownerCookies, columnID, "Owner Card")

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "cardviewer7", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "cardviewer7@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer tries to delete the card - should fail
	deleteCardQuery := fmt.Sprintf(`mutation {
		deleteCard(id: "%s")
	}`, cardID)

	resp, _ := ts.executeGraphQL(t, deleteCardQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to delete cards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCannotDeleteCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org, project and card
	ownerCookies := ts.registerUser(t, "cardowner8", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org8")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project8", "CTH")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)
	cardID := ts.createCard(t, ownerCookies, columnID, "Owner Card")

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "cardmember8", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "cardmember8@test.com", "00000000-0000-0000-0000-000000000003")

	// Member tries to delete the card - should fail (members don't have card:delete)
	deleteCardQuery := fmt.Sprintf(`mutation {
		deleteCard(id: "%s")
	}`, cardID)

	resp, _ := ts.executeGraphQL(t, deleteCardQuery, memberCookies)
	assert.NotEmpty(t, resp.Errors, "Member should not be able to delete cards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_AdminCanDeleteCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org, project and card
	ownerCookies := ts.registerUser(t, "cardowner9", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org9")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project9", "CTI")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)
	cardID := ts.createCard(t, ownerCookies, columnID, "Owner Card")

	// Create an admin and add them to the org
	adminCookies := ts.registerUser(t, "cardadmin9", "password123")
	ts.inviteAndAccept(t, ownerCookies, adminCookies, orgID, "cardadmin9@test.com", "00000000-0000-0000-0000-000000000002") // Admin role

	// Admin can delete the card
	deleteCardQuery := fmt.Sprintf(`mutation {
		deleteCard(id: "%s")
	}`, cardID)

	resp, _ := ts.executeGraphQL(t, deleteCardQuery, adminCookies)
	assert.Empty(t, resp.Errors, "Admin should be able to delete cards")
}

func TestRBAC_ViewerCanViewCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org, project and card
	ownerCookies := ts.registerUser(t, "cardowner10", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org10")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project10", "CTJ")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)
	cardID := ts.createCard(t, ownerCookies, columnID, "Owner Card")

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "cardviewer10", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "cardviewer10@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer can view the card
	viewCardQuery := fmt.Sprintf(`query {
		card(id: "%s") { id title }
	}`, cardID)

	resp, _ := ts.executeGraphQL(t, viewCardQuery, viewerCookies)
	assert.Empty(t, resp.Errors, "Viewer should be able to view cards")

	var data struct {
		Card struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"card"`
	}
	json.Unmarshal(resp.Data, &data)
	assert.Equal(t, "Owner Card", data.Card.Title)
}

func TestRBAC_NonMemberCannotViewCard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org, project and card
	ownerCookies := ts.registerUser(t, "cardowner11", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "CardTest Org11")
	projectID := ts.createProject(t, ownerCookies, orgID, "CardTest Project11", "CTK")
	_, columnID := ts.getBoard(t, ownerCookies, projectID)
	cardID := ts.createCard(t, ownerCookies, columnID, "Secret Card")

	// Non-member user (not in the org)
	nonMemberCookies := ts.registerUser(t, "nonmember11", "password123")

	// Non-member tries to view the card - should fail
	viewCardQuery := fmt.Sprintf(`query {
		card(id: "%s") { id title }
	}`, cardID)

	resp, _ := ts.executeGraphQL(t, viewCardQuery, nonMemberCookies)
	assert.NotEmpty(t, resp.Errors, "Non-member should not be able to view cards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

// =============================================================================
// Board Permission Enforcement Tests
// =============================================================================

func TestRBAC_ViewerCannotCreateBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "boardowner1", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org1")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project", "BTA")

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "boardviewer1", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "boardviewer1@test.com", "00000000-0000-0000-0000-000000000004") // Viewer role

	// Viewer tries to create a board - should fail
	createBoardQuery := fmt.Sprintf(`mutation {
		createBoard(input: {projectId: "%s", name: "Viewer Board"}) { id }
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, createBoardQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to create boards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCanCreateBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "boardowner2", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org2")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project2", "BTB")

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "boardmember2", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "boardmember2@test.com", "00000000-0000-0000-0000-000000000003") // Member role

	// Member can create a board
	createBoardQuery := fmt.Sprintf(`mutation {
		createBoard(input: {projectId: "%s", name: "Member Board"}) { id name }
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, createBoardQuery, memberCookies)
	assert.Empty(t, resp.Errors, "Member should be able to create boards")

	var data struct {
		CreateBoard struct {
			Name string `json:"name"`
		} `json:"createBoard"`
	}
	json.Unmarshal(resp.Data, &data)
	assert.Equal(t, "Member Board", data.CreateBoard.Name)
}

func TestRBAC_ViewerCannotUpdateBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "boardowner3", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org3")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project3", "BTC")
	boardID, _ := ts.getBoard(t, ownerCookies, projectID)

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "boardviewer3", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "boardviewer3@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer tries to update the board - should fail
	updateBoardQuery := fmt.Sprintf(`mutation {
		updateBoard(input: {id: "%s", name: "Hacked Board"}) { id }
	}`, boardID)

	resp, _ := ts.executeGraphQL(t, updateBoardQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to update boards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCannotUpdateBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "boardowner4", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org4")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project4", "BTD")
	boardID, _ := ts.getBoard(t, ownerCookies, projectID)

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "boardmember4", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "boardmember4@test.com", "00000000-0000-0000-0000-000000000003")

	// Member tries to update the board - should fail (members don't have board:manage)
	updateBoardQuery := fmt.Sprintf(`mutation {
		updateBoard(input: {id: "%s", name: "Updated Board"}) { id }
	}`, boardID)

	resp, _ := ts.executeGraphQL(t, updateBoardQuery, memberCookies)
	assert.NotEmpty(t, resp.Errors, "Member should not be able to update boards (no board:manage)")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_AdminCanUpdateBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "boardowner5", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org5")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project5", "BTE")
	boardID, _ := ts.getBoard(t, ownerCookies, projectID)

	// Create an admin and add them to the org
	adminCookies := ts.registerUser(t, "boardadmin5", "password123")
	ts.inviteAndAccept(t, ownerCookies, adminCookies, orgID, "boardadmin5@test.com", "00000000-0000-0000-0000-000000000002")

	// Admin can update the board
	updateBoardQuery := fmt.Sprintf(`mutation {
		updateBoard(input: {id: "%s", name: "Admin Updated Board"}) { id name }
	}`, boardID)

	resp, _ := ts.executeGraphQL(t, updateBoardQuery, adminCookies)
	assert.Empty(t, resp.Errors, "Admin should be able to update boards")

	var data struct {
		UpdateBoard struct {
			Name string `json:"name"`
		} `json:"updateBoard"`
	}
	json.Unmarshal(resp.Data, &data)
	assert.Equal(t, "Admin Updated Board", data.UpdateBoard.Name)
}

func TestRBAC_ViewerCannotDeleteBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project with extra board
	ownerCookies := ts.registerUser(t, "boardowner6", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org6")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project6", "BTF")

	// Create an extra board
	createBoardQuery := fmt.Sprintf(`mutation {
		createBoard(input: {projectId: "%s", name: "Extra Board"}) { id }
	}`, projectID)
	resp, _ := ts.executeGraphQL(t, createBoardQuery, ownerCookies)
	require.Empty(t, resp.Errors)
	var createData struct {
		CreateBoard struct {
			ID string `json:"id"`
		} `json:"createBoard"`
	}
	json.Unmarshal(resp.Data, &createData)
	extraBoardID := createData.CreateBoard.ID

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "boardviewer6", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "boardviewer6@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer tries to delete the board - should fail
	deleteBoardQuery := fmt.Sprintf(`mutation {
		deleteBoard(id: "%s")
	}`, extraBoardID)

	resp, _ = ts.executeGraphQL(t, deleteBoardQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to delete boards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCannotDeleteBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project with extra board
	ownerCookies := ts.registerUser(t, "boardowner7", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org7")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project7", "BTG")

	// Create an extra board
	createBoardQuery := fmt.Sprintf(`mutation {
		createBoard(input: {projectId: "%s", name: "Extra Board"}) { id }
	}`, projectID)
	resp, _ := ts.executeGraphQL(t, createBoardQuery, ownerCookies)
	require.Empty(t, resp.Errors)
	var createData struct {
		CreateBoard struct {
			ID string `json:"id"`
		} `json:"createBoard"`
	}
	json.Unmarshal(resp.Data, &createData)
	extraBoardID := createData.CreateBoard.ID

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "boardmember7", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "boardmember7@test.com", "00000000-0000-0000-0000-000000000003")

	// Member tries to delete the board - should fail (members don't have board:delete)
	deleteBoardQuery := fmt.Sprintf(`mutation {
		deleteBoard(id: "%s")
	}`, extraBoardID)

	resp, _ = ts.executeGraphQL(t, deleteBoardQuery, memberCookies)
	assert.NotEmpty(t, resp.Errors, "Member should not be able to delete boards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_AdminCanDeleteBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project with extra board
	ownerCookies := ts.registerUser(t, "boardowner8", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org8")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project8", "BTH")

	// Create an extra board
	createBoardQuery := fmt.Sprintf(`mutation {
		createBoard(input: {projectId: "%s", name: "Extra Board"}) { id }
	}`, projectID)
	resp, _ := ts.executeGraphQL(t, createBoardQuery, ownerCookies)
	require.Empty(t, resp.Errors)
	var createData struct {
		CreateBoard struct {
			ID string `json:"id"`
		} `json:"createBoard"`
	}
	json.Unmarshal(resp.Data, &createData)
	extraBoardID := createData.CreateBoard.ID

	// Create an admin and add them to the org
	adminCookies := ts.registerUser(t, "boardadmin8", "password123")
	ts.inviteAndAccept(t, ownerCookies, adminCookies, orgID, "boardadmin8@test.com", "00000000-0000-0000-0000-000000000002")

	// Admin can delete the board
	deleteBoardQuery := fmt.Sprintf(`mutation {
		deleteBoard(id: "%s")
	}`, extraBoardID)

	resp, _ = ts.executeGraphQL(t, deleteBoardQuery, adminCookies)
	assert.Empty(t, resp.Errors, "Admin should be able to delete boards")
}

func TestRBAC_ViewerCanViewBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "boardowner9", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org9")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project9", "BTI")
	boardID, _ := ts.getBoard(t, ownerCookies, projectID)

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "boardviewer9", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "boardviewer9@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer can view the board
	viewBoardQuery := fmt.Sprintf(`query {
		board(id: "%s") { id name }
	}`, boardID)

	resp, _ := ts.executeGraphQL(t, viewBoardQuery, viewerCookies)
	assert.Empty(t, resp.Errors, "Viewer should be able to view boards")
}

func TestRBAC_NonMemberCannotViewBoard(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "boardowner10", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "BoardTest Org10")
	projectID := ts.createProject(t, ownerCookies, orgID, "BoardTest Project10", "BTJ")
	boardID, _ := ts.getBoard(t, ownerCookies, projectID)

	// Non-member user (not in the org)
	nonMemberCookies := ts.registerUser(t, "boardnonmember10", "password123")

	// Non-member tries to view the board - should fail
	viewBoardQuery := fmt.Sprintf(`query {
		board(id: "%s") { id name }
	}`, boardID)

	resp, _ := ts.executeGraphQL(t, viewBoardQuery, nonMemberCookies)
	assert.NotEmpty(t, resp.Errors, "Non-member should not be able to view boards")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

// =============================================================================
// Project Permission Enforcement Tests
// =============================================================================

func TestRBAC_ViewerCannotCreateProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org
	ownerCookies := ts.registerUser(t, "projowner1", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org1")

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "projviewer1", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "projviewer1@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer tries to create a project - should fail
	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: {organizationId: "%s", name: "Viewer Project", key: "PTA"}) { id }
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, createProjectQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to create projects")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCanCreateProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org
	ownerCookies := ts.registerUser(t, "projowner2", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org2")

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "projmember2", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "projmember2@test.com", "00000000-0000-0000-0000-000000000003")

	// Member can create a project
	createProjectQuery := fmt.Sprintf(`mutation {
		createProject(input: {organizationId: "%s", name: "Member Project", key: "PTB"}) { id name }
	}`, orgID)

	resp, _ := ts.executeGraphQL(t, createProjectQuery, memberCookies)
	assert.Empty(t, resp.Errors, "Member should be able to create projects")

	var data struct {
		CreateProject struct {
			Name string `json:"name"`
		} `json:"createProject"`
	}
	json.Unmarshal(resp.Data, &data)
	assert.Equal(t, "Member Project", data.CreateProject.Name)
}

func TestRBAC_ViewerCannotUpdateProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "projowner3", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org3")
	projectID := ts.createProject(t, ownerCookies, orgID, "ProjectTest Project3", "PTC")

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "projviewer3", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "projviewer3@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer tries to update the project - should fail
	updateProjectQuery := fmt.Sprintf(`mutation {
		updateProject(input: {id: "%s", name: "Hacked Project"}) { id }
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, updateProjectQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to update projects")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCannotUpdateProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "projowner4", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org4")
	projectID := ts.createProject(t, ownerCookies, orgID, "ProjectTest Project4", "PTD")

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "projmember4", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "projmember4@test.com", "00000000-0000-0000-0000-000000000003")

	// Member tries to update the project - should fail (members don't have project:manage)
	updateProjectQuery := fmt.Sprintf(`mutation {
		updateProject(input: {id: "%s", name: "Updated Project"}) { id }
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, updateProjectQuery, memberCookies)
	assert.NotEmpty(t, resp.Errors, "Member should not be able to update projects (no project:manage)")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_AdminCanUpdateProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "projowner5", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org5")
	projectID := ts.createProject(t, ownerCookies, orgID, "ProjectTest Project5", "PTE")

	// Create an admin and add them to the org
	adminCookies := ts.registerUser(t, "projadmin5", "password123")
	ts.inviteAndAccept(t, ownerCookies, adminCookies, orgID, "projadmin5@test.com", "00000000-0000-0000-0000-000000000002")

	// Admin can update the project
	updateProjectQuery := fmt.Sprintf(`mutation {
		updateProject(input: {id: "%s", name: "Admin Updated Project"}) { id name }
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, updateProjectQuery, adminCookies)
	assert.Empty(t, resp.Errors, "Admin should be able to update projects")

	var data struct {
		UpdateProject struct {
			Name string `json:"name"`
		} `json:"updateProject"`
	}
	json.Unmarshal(resp.Data, &data)
	assert.Equal(t, "Admin Updated Project", data.UpdateProject.Name)
}

func TestRBAC_ViewerCannotDeleteProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and extra project
	ownerCookies := ts.registerUser(t, "projowner6", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org6")
	projectID := ts.createProject(t, ownerCookies, orgID, "ProjectTest Project6", "PTF")

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "projviewer6", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "projviewer6@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer tries to delete the project - should fail
	deleteProjectQuery := fmt.Sprintf(`mutation {
		deleteProject(id: "%s")
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, deleteProjectQuery, viewerCookies)
	assert.NotEmpty(t, resp.Errors, "Viewer should not be able to delete projects")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_MemberCannotDeleteProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "projowner7", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org7")
	projectID := ts.createProject(t, ownerCookies, orgID, "ProjectTest Project7", "PTG")

	// Create a member and add them to the org
	memberCookies := ts.registerUser(t, "projmember7", "password123")
	ts.inviteAndAccept(t, ownerCookies, memberCookies, orgID, "projmember7@test.com", "00000000-0000-0000-0000-000000000003")

	// Member tries to delete the project - should fail (members don't have project:delete)
	deleteProjectQuery := fmt.Sprintf(`mutation {
		deleteProject(id: "%s")
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, deleteProjectQuery, memberCookies)
	assert.NotEmpty(t, resp.Errors, "Member should not be able to delete projects")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}

func TestRBAC_AdminCanDeleteProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "projowner8", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org8")
	projectID := ts.createProject(t, ownerCookies, orgID, "ProjectTest Project8", "PTH")

	// Create an admin and add them to the org
	adminCookies := ts.registerUser(t, "projadmin8", "password123")
	ts.inviteAndAccept(t, ownerCookies, adminCookies, orgID, "projadmin8@test.com", "00000000-0000-0000-0000-000000000002")

	// Admin can delete the project
	deleteProjectQuery := fmt.Sprintf(`mutation {
		deleteProject(id: "%s")
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, deleteProjectQuery, adminCookies)
	assert.Empty(t, resp.Errors, "Admin should be able to delete projects")
}

func TestRBAC_ViewerCanViewProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "projowner9", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org9")
	projectID := ts.createProject(t, ownerCookies, orgID, "ProjectTest Project9", "PTI")

	// Create a viewer and add them to the org
	viewerCookies := ts.registerUser(t, "projviewer9", "password123")
	ts.inviteAndAccept(t, ownerCookies, viewerCookies, orgID, "projviewer9@test.com", "00000000-0000-0000-0000-000000000004")

	// Viewer can view the project
	viewProjectQuery := fmt.Sprintf(`query {
		project(id: "%s") { id name }
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, viewProjectQuery, viewerCookies)
	assert.Empty(t, resp.Errors, "Viewer should be able to view projects")

	var data struct {
		Project struct {
			Name string `json:"name"`
		} `json:"project"`
	}
	json.Unmarshal(resp.Data, &data)
	assert.Equal(t, "ProjectTest Project9", data.Project.Name)
}

func TestRBAC_NonMemberCannotViewProject(t *testing.T) {
	ts := setupRBACTestServer(t)
	defer ts.cleanup(t)

	// Owner creates org and project
	ownerCookies := ts.registerUser(t, "projowner10", "password123")
	orgID := ts.createOrganization(t, ownerCookies, "ProjectTest Org10")
	projectID := ts.createProject(t, ownerCookies, orgID, "Secret Project", "PTJ")

	// Non-member user (not in the org)
	nonMemberCookies := ts.registerUser(t, "projnonmember10", "password123")

	// Non-member tries to view the project - should fail
	viewProjectQuery := fmt.Sprintf(`query {
		project(id: "%s") { id name }
	}`, projectID)

	resp, _ := ts.executeGraphQL(t, viewProjectQuery, nonMemberCookies)
	assert.NotEmpty(t, resp.Errors, "Non-member should not be able to view projects")
	assert.Contains(t, resp.Errors[0].Message, "unauthorized")
}
