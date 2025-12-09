package project

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
	`)

	// Clean up
	db.Exec("DELETE FROM projects")
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

func createTestOrganization(t *testing.T, db *gorm.DB, name, slug string, ownerID uuid.UUID) uuid.UUID {
	orgID := uuid.New()
	err := db.Exec("INSERT INTO organizations (id, name, slug, owner_id) VALUES (?, ?, ?, ?)",
		orgID, name, slug, ownerID).Error
	require.NoError(t, err)
	return orgID
}

func TestProjectRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	orgID := createTestOrganization(t, db, "Test Org", "test-org", userID)

	project := &Project{
		OrganizationID: orgID,
		Name:           "Test Project",
		Key:            "TEST",
		Description:    "A test project",
	}

	err := repo.Create(ctx, project)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, project.ID)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, "TEST", project.Key)
	assert.False(t, project.CreatedAt.IsZero())
}

func TestProjectRepository_Create_DuplicateKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	orgID := createTestOrganization(t, db, "Test Org", "test-org", userID)

	project1 := &Project{
		OrganizationID: orgID,
		Name:           "Project 1",
		Key:            "DUPKEY",
	}
	project2 := &Project{
		OrganizationID: orgID,
		Name:           "Project 2",
		Key:            "DUPKEY",
	}

	err := repo.Create(ctx, project1)
	require.NoError(t, err)

	err = repo.Create(ctx, project2)
	assert.Error(t, err) // Should fail due to unique constraint
}

func TestProjectRepository_Create_SameKeyDifferentOrg(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	orgID1 := createTestOrganization(t, db, "Org 1", "org-1", userID)
	orgID2 := createTestOrganization(t, db, "Org 2", "org-2", userID)

	project1 := &Project{
		OrganizationID: orgID1,
		Name:           "Project 1",
		Key:            "SAMEKEY",
	}
	project2 := &Project{
		OrganizationID: orgID2,
		Name:           "Project 2",
		Key:            "SAMEKEY",
	}

	err := repo.Create(ctx, project1)
	require.NoError(t, err)

	err = repo.Create(ctx, project2)
	require.NoError(t, err) // Should succeed - different orgs
}

func TestProjectRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	orgID := createTestOrganization(t, db, "Test Org", "test-org", userID)

	project := &Project{
		OrganizationID: orgID,
		Name:           "Find Me",
		Key:            "FIND",
	}
	repo.Create(ctx, project)

	found, err := repo.GetByID(ctx, project.ID)

	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Find Me", found.Name)
	assert.Equal(t, project.ID, found.ID)
}

func TestProjectRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
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

func TestProjectRepository_GetByOrgID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	orgID1 := createTestOrganization(t, db, "Org 1", "org-1", userID)
	orgID2 := createTestOrganization(t, db, "Org 2", "org-2", userID)

	// Create 2 projects for orgID1
	repo.Create(ctx, &Project{OrganizationID: orgID1, Name: "Project 1", Key: "PRJ1"})
	repo.Create(ctx, &Project{OrganizationID: orgID1, Name: "Project 2", Key: "PRJ2"})
	// Create 1 project for orgID2
	repo.Create(ctx, &Project{OrganizationID: orgID2, Name: "Project 3", Key: "PRJ3"})

	projects, err := repo.GetByOrgID(ctx, orgID1)

	require.NoError(t, err)
	assert.Len(t, projects, 2)
}

func TestProjectRepository_GetByKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	orgID := createTestOrganization(t, db, "Test Org", "test-org", userID)

	project := &Project{
		OrganizationID: orgID,
		Name:           "Key Project",
		Key:            "KEYPRJ",
	}
	repo.Create(ctx, project)

	found, err := repo.GetByKey(ctx, orgID, "KEYPRJ")

	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Key Project", found.Name)
}

func TestProjectRepository_GetByKey_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	orgID := createTestOrganization(t, db, "Test Org", "test-org", userID)

	found, err := repo.GetByKey(ctx, orgID, "NONEXIST")

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestProjectRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	orgID := createTestOrganization(t, db, "Test Org", "test-org", userID)

	project := &Project{
		OrganizationID: orgID,
		Name:           "Original Name",
		Key:            "ORIG",
	}
	repo.Create(ctx, project)

	project.Name = "Updated Name"
	project.Description = "New description"
	err := repo.Update(ctx, project)

	require.NoError(t, err)

	found, _ := repo.GetByID(ctx, project.ID)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "New description", found.Description)
}

func TestProjectRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM projects")
	defer db.Exec("DELETE FROM organizations")
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	userID := createTestUser(t, db, "testowner")
	orgID := createTestOrganization(t, db, "Test Org", "test-org", userID)

	project := &Project{
		OrganizationID: orgID,
		Name:           "Delete Me",
		Key:            "DEL",
	}
	repo.Create(ctx, project)

	err := repo.Delete(ctx, project.ID)
	require.NoError(t, err)

	found, err := repo.GetByID(ctx, project.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
}
