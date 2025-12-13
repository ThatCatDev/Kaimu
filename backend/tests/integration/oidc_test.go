package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/thatcatdev/kaimu/backend/http/handlers"
	oidcIdentityRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/oidc_identity"
	refreshTokenRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/refreshtoken"
	userRepo "github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/services/auth"
	"github.com/thatcatdev/kaimu/backend/internal/services/oidc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type OIDCTestServer struct {
	handler      *handlers.OIDCHandler
	router       *mux.Router
	db           *gorm.DB
	stateManager oidc.StateManager
}

func setupOIDCTestServer(t *testing.T) *OIDCTestServer {
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

	// Run migrations for users and oidc_identities
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
		CREATE TABLE IF NOT EXISTS oidc_identities (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			issuer VARCHAR(512) NOT NULL,
			subject VARCHAR(512) NOT NULL,
			email VARCHAR(255),
			email_verified BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(issuer, subject)
		);
		CREATE INDEX IF NOT EXISTS idx_oidc_identities_user_id ON oidc_identities(user_id);
		CREATE INDEX IF NOT EXISTS idx_oidc_identities_issuer_subject ON oidc_identities(issuer, subject);

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
	`).Error
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clean up tables before test
	testDB.Exec("DELETE FROM oidc_identities")
	testDB.Exec("DELETE FROM refresh_tokens")
	testDB.Exec("DELETE FROM users")

	// Create test providers
	providers := []config.OIDCProvider{
		{
			Name:         "Test Provider",
			Slug:         "test",
			IssuerURL:    "https://test.example.com",
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Scopes:       "openid email profile",
		},
		{
			Name:         "Another Provider",
			Slug:         "another",
			IssuerURL:    "https://another.example.com",
			ClientID:     "another-client-id",
			ClientSecret: "another-client-secret",
			Scopes:       "openid email",
		},
	}

	// Create repositories
	userRepository := userRepo.NewRepository(testDB)
	identityRepository := oidcIdentityRepo.NewRepository(testDB)

	// Create state manager
	stateManager := oidc.NewStateManager(10)

	// Create refresh token repository
	refreshRepository := refreshTokenRepo.NewRepository(testDB)

	// Create auth service
	authService := auth.NewService(userRepository, refreshRepository, "test-jwt-secret", 15, 7)

	// Create OIDC service
	oidcService := oidc.NewService(
		providers,
		identityRepository,
		userRepository,
		stateManager,
		"http://localhost:3000",
		"http://localhost:4321",
	)

	// Create handler
	oidcHandler := handlers.NewOIDCHandler(oidcService, authService, "http://localhost:4321", false)

	// Create router
	router := mux.NewRouter()
	router.HandleFunc("/auth/oidc/providers", oidcHandler.ListProviders).Methods("GET")
	router.HandleFunc("/auth/oidc/{provider}/authorize", oidcHandler.Authorize).Methods("GET")
	router.HandleFunc("/auth/oidc/{provider}/callback", oidcHandler.Callback).Methods("GET")

	return &OIDCTestServer{
		handler:      oidcHandler,
		router:       router,
		db:           testDB,
		stateManager: stateManager,
	}
}

func (ts *OIDCTestServer) cleanup(t *testing.T) {
	ts.db.Exec("DELETE FROM oidc_identities")
	ts.db.Exec("DELETE FROM users")
}

func TestOIDC_ListProviders(t *testing.T) {
	ts := setupOIDCTestServer(t)
	defer ts.cleanup(t)

	req := httptest.NewRequest("GET", "/auth/oidc/providers", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var providers []oidc.ProviderInfo
	err := json.NewDecoder(w.Body).Decode(&providers)
	require.NoError(t, err)

	assert.Len(t, providers, 2)
	assert.Equal(t, "test", providers[0].Slug)
	assert.Equal(t, "Test Provider", providers[0].Name)
	assert.Equal(t, "another", providers[1].Slug)
	assert.Equal(t, "Another Provider", providers[1].Name)
}

func TestOIDC_Authorize_ProviderNotFound(t *testing.T) {
	ts := setupOIDCTestServer(t)
	defer ts.cleanup(t)

	req := httptest.NewRequest("GET", "/auth/oidc/nonexistent/authorize", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestOIDC_Authorize_MissingProvider(t *testing.T) {
	ts := setupOIDCTestServer(t)
	defer ts.cleanup(t)

	req := httptest.NewRequest("GET", "/auth/oidc//authorize", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	// mux treats empty provider differently
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestOIDC_Callback_MissingCode(t *testing.T) {
	ts := setupOIDCTestServer(t)
	defer ts.cleanup(t)

	req := httptest.NewRequest("GET", "/auth/oidc/test/callback?state=abc123", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
}

func TestOIDC_Callback_MissingState(t *testing.T) {
	ts := setupOIDCTestServer(t)
	defer ts.cleanup(t)

	req := httptest.NewRequest("GET", "/auth/oidc/test/callback?code=authcode", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
}

func TestOIDC_Callback_InvalidState(t *testing.T) {
	ts := setupOIDCTestServer(t)
	defer ts.cleanup(t)

	req := httptest.NewRequest("GET", "/auth/oidc/test/callback?code=authcode&state=invalid-state", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
	assert.Contains(t, w.Header().Get("Location"), "expired")
}

func TestOIDC_Callback_ProviderError(t *testing.T) {
	ts := setupOIDCTestServer(t)
	defer ts.cleanup(t)

	req := httptest.NewRequest("GET", "/auth/oidc/test/callback?error=access_denied&error_description=User+denied+access", nil)
	w := httptest.NewRecorder()

	ts.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/login?error=")
}

func TestOIDC_Integration_StateManagerRoundtrip(t *testing.T) {
	// Test state manager functionality used in OIDC flow
	stateManager := oidc.NewStateManager(10)

	// Create state
	state, data, err := stateManager.CreateState("test-provider", "http://redirect.com")
	require.NoError(t, err)
	assert.NotEmpty(t, state)
	assert.Equal(t, "test-provider", data.ProviderSlug)

	// Get state
	retrieved, err := stateManager.GetState(state)
	require.NoError(t, err)
	assert.Equal(t, data.ProviderSlug, retrieved.ProviderSlug)
	assert.Equal(t, data.CodeVerifier, retrieved.CodeVerifier)
	assert.Equal(t, data.Nonce, retrieved.Nonce)

	// Delete state
	stateManager.DeleteState(state)

	// State should no longer exist
	_, err = stateManager.GetState(state)
	assert.Error(t, err)
	assert.Equal(t, oidc.ErrInvalidState, err)
}

func TestOIDC_Integration_PKCECodeChallenge(t *testing.T) {
	// Test PKCE code challenge generation
	verifier := "test-code-verifier-with-sufficient-length-for-pkce"
	challenge := oidc.GenerateCodeChallenge(verifier)

	assert.NotEmpty(t, challenge)
	// Same verifier should always produce same challenge
	assert.Equal(t, challenge, oidc.GenerateCodeChallenge(verifier))
	// Different verifiers should produce different challenges
	assert.NotEqual(t, challenge, oidc.GenerateCodeChallenge("different-verifier"))
}
