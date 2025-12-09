package user

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
	`)

	// Clean up before test
	db.Exec("DELETE FROM users")

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	user := &User{
		Username:     "testcreate",
		PasswordHash: "hashedpassword",
	}

	err := repo.Create(ctx, user)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, "testcreate", user.Username)
	assert.False(t, user.CreatedAt.IsZero())
}

func TestUserRepository_Create_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	user1 := &User{
		Username:     "duplicate",
		PasswordHash: "hash1",
	}
	user2 := &User{
		Username:     "duplicate",
		PasswordHash: "hash2",
	}

	err := repo.Create(ctx, user1)
	require.NoError(t, err)

	err = repo.Create(ctx, user2)
	assert.Error(t, err)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	// Create a user first
	user := &User{
		Username:     "findme",
		PasswordHash: "password",
	}
	repo.Create(ctx, user)

	// Find the user
	found, err := repo.GetByUsername(ctx, "findme")

	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "findme", found.Username)
	assert.Equal(t, user.ID, found.ID)
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	found, err := repo.GetByUsername(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	// Create a user first
	user := &User{
		Username:     "findbyid",
		PasswordHash: "password",
	}
	repo.Create(ctx, user)

	// Find the user by ID
	found, err := repo.GetByID(ctx, user.ID)

	require.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "findbyid", found.Username)
	assert.Equal(t, user.ID, found.ID)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Exec("DELETE FROM users")

	repo := NewRepository(db)
	ctx := context.Background()

	randomID := uuid.New()
	found, err := repo.GetByID(ctx, randomID)

	assert.Error(t, err)
	assert.Nil(t, found)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
