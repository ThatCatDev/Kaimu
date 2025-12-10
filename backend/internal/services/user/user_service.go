package user

//go:generate mockgen -source=user_service.go -destination=mocks/user_service_mock.go -package=mocks

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"github.com/thatcatdev/pulse-backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Service interface {
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	Update(ctx context.Context, id uuid.UUID, displayName, email *string) (*user.User, error)
}

type service struct {
	repository user.Repository
}

func NewService(userRepo user.Repository) Service {
	return &service{
		repository: userRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "user.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "user"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	ctx, span := s.startServiceSpan(ctx, "GetByID")
	span.SetAttributes(attribute.String("user.id", id.String()))
	defer span.End()

	u, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, displayName, email *string) (*user.User, error) {
	ctx, span := s.startServiceSpan(ctx, "Update")
	span.SetAttributes(attribute.String("user.id", id.String()))
	defer span.End()

	u, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if displayName != nil {
		u.DisplayName = displayName
	}
	if email != nil {
		u.Email = email
	}

	if err := s.repository.Update(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}
