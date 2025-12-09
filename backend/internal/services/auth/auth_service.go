package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"github.com/thatcatdev/pulse-backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserExists         = errors.New("username already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUserNotFound       = errors.New("user not found")
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type Service interface {
	Register(ctx context.Context, username, password string) (*user.User, string, error)
	Login(ctx context.Context, username, password string) (*user.User, string, error)
	ValidateToken(tokenString string) (*Claims, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}

type service struct {
	repository    user.Repository
	jwtSecret     []byte
	jwtExpiration time.Duration
}

// startServiceSpan starts a new OpenTelemetry span for service operations
func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "auth.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "auth"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

func NewService(userRepo user.Repository, jwtSecret string, jwtExpirationHours int) Service {
	return &service{
		repository:    userRepo,
		jwtSecret:     []byte(jwtSecret),
		jwtExpiration: time.Duration(jwtExpirationHours) * time.Hour,
	}
}

func (s *service) Register(ctx context.Context, username, password string) (*user.User, string, error) {
	ctx, span := s.startServiceSpan(ctx, "Register")
	span.SetAttributes(attribute.String("auth.username", username))
	defer span.End()

	// Check if user exists
	existing, err := s.repository.GetByUsername(ctx, username)
	if err == nil && existing != nil {
		return nil, "", ErrUserExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// Create user
	newUser := &user.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repository.Create(ctx, newUser); err != nil {
		return nil, "", err
	}

	// Generate token
	token, err := s.generateToken(newUser.ID)
	if err != nil {
		return nil, "", err
	}

	return newUser, token, nil
}

func (s *service) Login(ctx context.Context, username, password string) (*user.User, string, error) {
	ctx, span := s.startServiceSpan(ctx, "Login")
	span.SetAttributes(attribute.String("auth.username", username))
	defer span.End()

	// Find user
	u, err := s.repository.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// Generate token
	token, err := s.generateToken(u.ID)
	if err != nil {
		return nil, "", err
	}

	return u, token, nil
}

func (s *service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	ctx, span := s.startServiceSpan(ctx, "GetUserByID")
	span.SetAttributes(attribute.String("auth.user_id", id.String()))
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

func (s *service) generateToken(userID uuid.UUID) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "pulse",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
