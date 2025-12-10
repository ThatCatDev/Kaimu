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
	userRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/directives"
	"github.com/thatcatdev/kaimu/backend/internal/services/auth"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

type TestServer struct {
	handler     http.Handler
	db          *gorm.DB
	authService auth.Service
}

func setupTestServer(t *testing.T) *TestServer {
	// Use test database config from environment or defaults
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
			password_hash VARCHAR(255),
			email VARCHAR(255) UNIQUE,
			display_name VARCHAR(255),
			avatar_url TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`).Error
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clean up users table before test
	testDB.Exec("DELETE FROM users")

	// Create services
	userRepository := userRepo.NewRepository(testDB)
	authService := auth.NewService(userRepository, "test-jwt-secret", 24)

	// Create resolver
	cfg := config.Config{
		AppConfig: config.AppConfig{
			Env: "test",
		},
	}
	resolver := &graph.Resolver{
		Config:      cfg,
		AuthService: authService,
	}

	// Create GraphQL handler
	gqlConfig := generated.Config{
		Resolvers:  resolver,
		Directives: directives.GetDirectives(),
	}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(gqlConfig))

	// Wrap with auth middleware
	wrappedHandler := middleware.AuthMiddleware(authService)(srv)

	return &TestServer{
		handler:     wrappedHandler,
		db:          testDB,
		authService: authService,
	}
}

func (ts *TestServer) cleanup(t *testing.T) {
	ts.db.Exec("DELETE FROM users")
}

func (ts *TestServer) executeGraphQL(t *testing.T, query string, cookies []*http.Cookie) (*GraphQLResponse, []*http.Cookie) {
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

func TestIntegration_Register(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	query := `mutation {
		register(input: {username: "testuser", password: "password123"}) {
			user {
				id
				username
			}
		}
	}`

	resp, cookies := ts.executeGraphQL(t, query, nil)

	assert.Empty(t, resp.Errors, "Expected no errors")

	var data struct {
		Register struct {
			User struct {
				ID       string `json:"id"`
				Username string `json:"username"`
			} `json:"user"`
		} `json:"register"`
	}
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)

	assert.NotEmpty(t, data.Register.User.ID)
	assert.Equal(t, "testuser", data.Register.User.Username)

	// Check that cookie was set
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "pulse_token" {
			tokenCookie = c
			break
		}
	}
	assert.NotNil(t, tokenCookie, "Expected pulse_token cookie to be set")
	assert.NotEmpty(t, tokenCookie.Value)
}

func TestIntegration_Register_DuplicateUsername(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	// Register first user
	query := `mutation {
		register(input: {username: "duplicate", password: "password123"}) {
			user { id }
		}
	}`
	ts.executeGraphQL(t, query, nil)

	// Try to register with same username
	resp, _ := ts.executeGraphQL(t, query, nil)

	assert.NotEmpty(t, resp.Errors, "Expected error for duplicate username")
	assert.Contains(t, resp.Errors[0].Message, "already taken")
}

func TestIntegration_Login_Success(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	// First register a user
	registerQuery := `mutation {
		register(input: {username: "loginuser", password: "mypassword"}) {
			user { id }
		}
	}`
	ts.executeGraphQL(t, registerQuery, nil)

	// Now login
	loginQuery := `mutation {
		login(input: {username: "loginuser", password: "mypassword"}) {
			user {
				id
				username
			}
		}
	}`

	resp, cookies := ts.executeGraphQL(t, loginQuery, nil)

	assert.Empty(t, resp.Errors, "Expected no errors")

	var data struct {
		Login struct {
			User struct {
				ID       string `json:"id"`
				Username string `json:"username"`
			} `json:"user"`
		} `json:"login"`
	}
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)

	assert.Equal(t, "loginuser", data.Login.User.Username)

	// Check cookie
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "pulse_token" {
			tokenCookie = c
			break
		}
	}
	assert.NotNil(t, tokenCookie)
}

func TestIntegration_Login_InvalidPassword(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	// Register a user
	registerQuery := `mutation {
		register(input: {username: "wrongpass", password: "correctpassword"}) {
			user { id }
		}
	}`
	ts.executeGraphQL(t, registerQuery, nil)

	// Try to login with wrong password
	loginQuery := `mutation {
		login(input: {username: "wrongpass", password: "incorrectpassword"}) {
			user { id }
		}
	}`

	resp, _ := ts.executeGraphQL(t, loginQuery, nil)

	assert.NotEmpty(t, resp.Errors)
	assert.Contains(t, resp.Errors[0].Message, "invalid username or password")
}

func TestIntegration_Login_NonexistentUser(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	loginQuery := `mutation {
		login(input: {username: "nonexistent", password: "password"}) {
			user { id }
		}
	}`

	resp, _ := ts.executeGraphQL(t, loginQuery, nil)

	assert.NotEmpty(t, resp.Errors)
	assert.Contains(t, resp.Errors[0].Message, "invalid username or password")
}

func TestIntegration_Me_Authenticated(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	// Register and get cookie
	registerQuery := `mutation {
		register(input: {username: "meuser", password: "password123"}) {
			user { id username }
		}
	}`
	_, cookies := ts.executeGraphQL(t, registerQuery, nil)

	// Query me with the cookie
	meQuery := `query { me { id username createdAt } }`
	resp, _ := ts.executeGraphQL(t, meQuery, cookies)

	assert.Empty(t, resp.Errors)

	var data struct {
		Me struct {
			ID        string    `json:"id"`
			Username  string    `json:"username"`
			CreatedAt time.Time `json:"createdAt"`
		} `json:"me"`
	}
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)

	assert.Equal(t, "meuser", data.Me.Username)
	assert.NotEmpty(t, data.Me.ID)
	assert.False(t, data.Me.CreatedAt.IsZero())
}

func TestIntegration_Me_Unauthenticated(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	meQuery := `query { me { id username } }`
	resp, _ := ts.executeGraphQL(t, meQuery, nil)

	assert.Empty(t, resp.Errors)

	var data struct {
		Me *struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"me"`
	}
	err := json.Unmarshal(resp.Data, &data)
	require.NoError(t, err)

	assert.Nil(t, data.Me, "Expected me to be null when unauthenticated")
}

func TestIntegration_Logout(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	// Register and get cookie
	registerQuery := `mutation {
		register(input: {username: "logoutuser", password: "password123"}) {
			user { id }
		}
	}`
	_, cookies := ts.executeGraphQL(t, registerQuery, nil)

	// Verify we're logged in
	meQuery := `query { me { username } }`
	resp, _ := ts.executeGraphQL(t, meQuery, cookies)
	var meData struct {
		Me *struct{ Username string } `json:"me"`
	}
	json.Unmarshal(resp.Data, &meData)
	assert.NotNil(t, meData.Me)

	// Logout
	logoutQuery := `mutation { logout }`
	resp, logoutCookies := ts.executeGraphQL(t, logoutQuery, cookies)

	assert.Empty(t, resp.Errors)

	var logoutData struct {
		Logout bool `json:"logout"`
	}
	json.Unmarshal(resp.Data, &logoutData)
	assert.True(t, logoutData.Logout)

	// Check that cookie was cleared
	var clearedCookie *http.Cookie
	for _, c := range logoutCookies {
		if c.Name == "pulse_token" {
			clearedCookie = c
			break
		}
	}
	assert.NotNil(t, clearedCookie)
	assert.Equal(t, "", clearedCookie.Value)
	assert.Equal(t, -1, clearedCookie.MaxAge)
}

func TestIntegration_FullAuthFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup(t)

	// 1. Register
	registerQuery := `mutation {
		register(input: {username: "flowuser", password: "flowpass"}) {
			user { id username }
		}
	}`
	resp, cookies := ts.executeGraphQL(t, registerQuery, nil)
	assert.Empty(t, resp.Errors)

	// 2. Check me is authenticated
	meQuery := `query { me { username } }`
	resp, _ = ts.executeGraphQL(t, meQuery, cookies)
	var meData struct {
		Me *struct{ Username string } `json:"me"`
	}
	json.Unmarshal(resp.Data, &meData)
	assert.Equal(t, "flowuser", meData.Me.Username)

	// 3. Logout
	logoutQuery := `mutation { logout }`
	resp, cookies = ts.executeGraphQL(t, logoutQuery, cookies)
	assert.Empty(t, resp.Errors)

	// 4. Check me is now null (use cleared cookies)
	resp, _ = ts.executeGraphQL(t, meQuery, cookies)
	json.Unmarshal(resp.Data, &meData)
	assert.Nil(t, meData.Me)

	// 5. Login again
	loginQuery := `mutation {
		login(input: {username: "flowuser", password: "flowpass"}) {
			user { username }
		}
	}`
	resp, cookies = ts.executeGraphQL(t, loginQuery, nil)
	assert.Empty(t, resp.Errors)

	// 6. Check me is authenticated again
	resp, _ = ts.executeGraphQL(t, meQuery, cookies)
	json.Unmarshal(resp.Data, &meData)
	assert.Equal(t, "flowuser", meData.Me.Username)
}
