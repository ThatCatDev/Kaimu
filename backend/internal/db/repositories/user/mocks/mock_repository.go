package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}
