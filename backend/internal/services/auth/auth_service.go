package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/refreshtoken"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrUserExists            = errors.New("username already exists")
	ErrInvalidToken          = errors.New("invalid or expired token")
	ErrInvalidRefreshToken   = errors.New("invalid or expired refresh token")
	ErrRefreshTokenRevoked   = errors.New("refresh token has been revoked")
	ErrUserNotFound          = errors.New("user not found")
	ErrPasswordLoginDisabled = errors.New("password login is disabled for this user")
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenPair contains both access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // Access token expiry in seconds
}

type Service interface {
	Register(ctx context.Context, username, email, password string, userAgent, ipAddress string) (*user.User, *TokenPair, error)
	Login(ctx context.Context, username, password string, userAgent, ipAddress string) (*user.User, *TokenPair, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshTokens(ctx context.Context, refreshToken string, userAgent, ipAddress string) (*TokenPair, error)
	RevokeRefreshToken(ctx context.Context, refreshToken string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	// GenerateTokenPair generates tokens for a user (used by OIDC flow)
	GenerateTokenPair(ctx context.Context, userID uuid.UUID, userAgent, ipAddress string) (*TokenPair, error)
}

type service struct {
	userRepository         user.Repository
	refreshTokenRepository refreshtoken.Repository
	jwtSecret              []byte
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration
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

func NewService(userRepo user.Repository, refreshTokenRepo refreshtoken.Repository, jwtSecret string, accessTokenExpirationMinutes, refreshTokenExpirationDays int) Service {
	return &service{
		userRepository:         userRepo,
		refreshTokenRepository: refreshTokenRepo,
		jwtSecret:              []byte(jwtSecret),
		accessTokenExpiration:  time.Duration(accessTokenExpirationMinutes) * time.Minute,
		refreshTokenExpiration: time.Duration(refreshTokenExpirationDays) * 24 * time.Hour,
	}
}

func (s *service) Register(ctx context.Context, username, email, password string, userAgent, ipAddress string) (*user.User, *TokenPair, error) {
	ctx, span := s.startServiceSpan(ctx, "Register")
	span.SetAttributes(attribute.String("auth.username", username))
	defer span.End()

	// Check if user exists
	existing, err := s.userRepository.GetByUsername(ctx, username)
	if err == nil && existing != nil {
		return nil, nil, ErrUserExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	hashedPasswordStr := string(hashedPassword)

	// Create user with email (unverified)
	newUser := &user.User{
		Username:      username,
		Email:         &email,
		EmailVerified: false,
		PasswordHash:  &hashedPasswordStr,
	}

	if err := s.userRepository.Create(ctx, newUser); err != nil {
		return nil, nil, err
	}

	// Generate token pair
	tokenPair, err := s.GenerateTokenPair(ctx, newUser.ID, userAgent, ipAddress)
	if err != nil {
		return nil, nil, err
	}

	return newUser, tokenPair, nil
}

func (s *service) Login(ctx context.Context, username, password string, userAgent, ipAddress string) (*user.User, *TokenPair, error) {
	ctx, span := s.startServiceSpan(ctx, "Login")
	span.SetAttributes(attribute.String("auth.username", username))
	defer span.End()

	// Find user
	u, err := s.userRepository.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Check if user has a password set (OIDC-only users don't)
	if u.PasswordHash == nil || *u.PasswordHash == "" {
		return nil, nil, ErrPasswordLoginDisabled
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Generate token pair
	tokenPair, err := s.GenerateTokenPair(ctx, u.ID, userAgent, ipAddress)
	if err != nil {
		return nil, nil, err
	}

	return u, tokenPair, nil
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

func (s *service) RefreshTokens(ctx context.Context, refreshTokenStr string, userAgent, ipAddress string) (*TokenPair, error) {
	ctx, span := s.startServiceSpan(ctx, "RefreshTokens")
	defer span.End()

	// Hash the refresh token to look it up
	tokenHash := hashToken(refreshTokenStr)

	// Find the refresh token
	storedToken, err := s.refreshTokenRepository.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	// Check if token is valid (not revoked and not expired)
	if !storedToken.IsValid() {
		// Token reuse detected - revoke all tokens for this user (security measure)
		if storedToken.RevokedAt != nil {
			_ = s.refreshTokenRepository.RevokeAllForUser(ctx, storedToken.UserID)
		}
		return nil, ErrRefreshTokenRevoked
	}

	// Generate new token pair
	newTokenPair, err := s.generateTokenPairInternal(ctx, storedToken.UserID, userAgent, ipAddress)
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token (rotation)
	newTokenHash := hashToken(newTokenPair.RefreshToken)
	newStoredToken, _ := s.refreshTokenRepository.GetByTokenHash(ctx, newTokenHash)
	var replacedByID *uuid.UUID
	if newStoredToken != nil {
		replacedByID = &newStoredToken.ID
	}
	_ = s.refreshTokenRepository.Revoke(ctx, storedToken.ID, replacedByID)

	return newTokenPair, nil
}

func (s *service) RevokeRefreshToken(ctx context.Context, refreshTokenStr string) error {
	ctx, span := s.startServiceSpan(ctx, "RevokeRefreshToken")
	defer span.End()

	tokenHash := hashToken(refreshTokenStr)
	storedToken, err := s.refreshTokenRepository.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Already revoked or doesn't exist
		}
		return err
	}

	return s.refreshTokenRepository.Revoke(ctx, storedToken.ID, nil)
}

func (s *service) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "RevokeAllUserTokens")
	span.SetAttributes(attribute.String("auth.user_id", userID.String()))
	defer span.End()

	return s.refreshTokenRepository.RevokeAllForUser(ctx, userID)
}

func (s *service) GetUserByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	ctx, span := s.startServiceSpan(ctx, "GetUserByID")
	span.SetAttributes(attribute.String("auth.user_id", id.String()))
	defer span.End()

	u, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (s *service) GenerateTokenPair(ctx context.Context, userID uuid.UUID, userAgent, ipAddress string) (*TokenPair, error) {
	ctx, span := s.startServiceSpan(ctx, "GenerateTokenPair")
	span.SetAttributes(attribute.String("auth.user_id", userID.String()))
	defer span.End()

	return s.generateTokenPairInternal(ctx, userID, userAgent, ipAddress)
}

func (s *service) generateTokenPairInternal(ctx context.Context, userID uuid.UUID, userAgent, ipAddress string) (*TokenPair, error) {
	// Generate access token (short-lived JWT)
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (random string, stored in DB)
	refreshTokenStr, err := generateRandomToken(32)
	if err != nil {
		return nil, err
	}

	// Store refresh token hash in database
	tokenHash := hashToken(refreshTokenStr)
	var ua, ip *string
	if userAgent != "" {
		ua = &userAgent
	}
	if ipAddress != "" {
		ip = &ipAddress
	}

	refreshTokenEntity := &refreshtoken.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.refreshTokenExpiration),
		UserAgent: ua,
		IPAddress: ip,
	}

	if err := s.refreshTokenRepository.Create(ctx, refreshTokenEntity); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		ExpiresIn:    int64(s.accessTokenExpiration.Seconds()),
	}, nil
}

func (s *service) generateAccessToken(userID uuid.UUID) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "kaimu",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// generateRandomToken generates a cryptographically secure random token
func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashToken creates a SHA-256 hash of the token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
