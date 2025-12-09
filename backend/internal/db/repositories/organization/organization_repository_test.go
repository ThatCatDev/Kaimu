package organization

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: could not connect to test database: %v", err)
	}

	// Setup schema
	db.Exec(`
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
	`)

	// Clean up
	db.Exec("DELETE FROM organization_members")
	db.Exec("DELETE FROM organizations")
	db.Exec("DELETE FROM users")

	return db
}

func createTestUser(t *testing.T, db *gorm.DB, username string) uuid.UUID {
	userID := uuid.New()
	err := db.Exec("INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)",
		userID, username, "hashedpassword").Error
	require.NoError(t, err)
	return userID
}

func TestOrganizationRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM organization_members")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")

	org := &Organization{
		Name:        "Test Org",
		Slug:        "test-org",
		Description: "A test organization",
		OwnerID:     userID,
	}

	err := repo.Create(ctx, org)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, org.ID)
	assert.Equal(t, "Test Org", org.Name)
	assert.Equal(t, "test-org", org.Slug)
	assert.False(t, org.CreatedAt.IsZero())
}

func TestOrganizationRepository_Create_DuplicateSlug(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM organization_members")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")

	org1 := &Organization{
		Name:    "Org 1",
		Slug:    "duplicate-slug",
		OwnerID: userID,
	}
	org2 := &Organization{
		Name:    "Org 2",
		Slug:    "duplicate-slug",
		OwnerID: userID,
	}

	err := repo.Create(ctx, org1)
	require.NoError(t, err)

	err = repo.Create(ctx, org2)
	assert.Error(t, err)
}

func TestOrganizationRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM organization_members")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")

	org := &Organization{
		Name:    "Find Me",
		Slug:    "find-me",
		OwnerID: userID,
	}
	repo.Create(ctx, org)

	found, err := repo.GetByID(ctx, org.ID)

	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Find Me", found.Name)
	assert.Equal(t, org.ID, found.ID)
}

func TestOrganizationRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM organization_members")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	randomID := uuid.New()
	found, err := repo.GetByID(ctx, randomID)

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestOrganizationRepository_GetBySlug(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM organization_members")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")

	org := &Organization{
		Name:    "Slug Org",
		Slug:    "slug-org",
		OwnerID: userID,
	}
	repo.Create(ctx, org)

	found, err := repo.GetBySlug(ctx, "slug-org")

	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Slug Org", found.Name)
}

func TestOrganizationRepository_GetByOwnerID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM organization_members")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	user2ID := createTestUser(t, db, "testowner2")

	// Create 2 orgs for userID
	repo.Create(ctx, &Organization{Name: "Org 1", Slug: "org-1", OwnerID: userID})
	repo.Create(ctx, &Organization{Name: "Org 2", Slug: "org-2", OwnerID: userID})
	// Create 1 org for user2ID
	repo.Create(ctx, &Organization{Name: "Org 3", Slug: "org-3", OwnerID: user2ID})

	orgs, err := repo.GetByOwnerID(ctx, userID)

	require.NoError(t, err)
	assert.Len(t, orgs, 2)
}

func TestOrganizationRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM organization_members")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")

	org := &Organization{
		Name:    "Original Name",
		Slug:    "original-slug",
		OwnerID: userID,
	}
	repo.Create(ctx, org)

	org.Name = "Updated Name"
	org.Description = "New description"
	err := repo.Update(ctx, org)

	require.NoError(t, err)

	found, _ := repo.GetByID(ctx, org.ID)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "New description", found.Description)
}

func TestOrganizationRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM organization_members")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")

	org := &Organization{
		Name:    "Delete Me",
		Slug:    "delete-me",
		OwnerID: userID,
	}
	repo.Create(ctx, org)

	err := repo.Delete(ctx, org.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, org.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
}
