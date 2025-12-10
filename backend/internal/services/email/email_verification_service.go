package email

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/email_verification_token"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/services/mail"
)

var (
	ErrTokenNotFound = errors.New("verification token not found")
	ErrTokenExpired  = errors.New("verification token has expired")
	ErrTokenUsed     = errors.New("verification token has already been used")
	ErrEmailMismatch = errors.New("email does not match token")
)

const (
	TokenExpirationDuration = 24 * time.Hour // Tokens expire after 24 hours
)

type EmailVerificationService interface {
	SendVerificationEmail(ctx context.Context, userID uuid.UUID, email, name string) error
	VerifyEmail(ctx context.Context, token string) (*user.User, error)
	ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error
}

type emailVerificationService struct {
	tokenRepo   email_verification_token.EmailVerificationTokenRepository
	userRepo    user.Repository
	mailService mail.MailService
	config      config.EmailConfig
}

func NewEmailVerificationService(
	tokenRepo email_verification_token.EmailVerificationTokenRepository,
	userRepo user.Repository,
	mailService mail.MailService,
	cfg config.EmailConfig,
) EmailVerificationService {
	return &emailVerificationService{
		tokenRepo:   tokenRepo,
		userRepo:    userRepo,
		mailService: mailService,
		config:      cfg,
	}
}

func (s *emailVerificationService) SendVerificationEmail(ctx context.Context, userID uuid.UUID, email, name string) error {
	// Delete any existing tokens for this user
	_ = s.tokenRepo.DeleteByUserID(ctx, userID)

	// Create new verification token
	token, err := s.tokenRepo.Create(ctx, userID, email, TokenExpirationDuration)
	if err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Build verification URL
	verificationURL := fmt.Sprintf("%s?token=%s", s.config.VerificationURL, token.Token)

	// Send email
	err = s.mailService.SendMail(ctx, []string{email}, "Verify your Kaimu account", "verification.mjml", map[string]string{
		"name":      name,
		"token_url": verificationURL,
	})
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

func (s *emailVerificationService) VerifyEmail(ctx context.Context, tokenStr string) (*user.User, error) {
	// Find token
	token, err := s.tokenRepo.FindByToken(ctx, tokenStr)
	if err != nil {
		return nil, ErrTokenNotFound
	}

	// Check if token has been used
	if token.UsedAt != nil {
		return nil, ErrTokenUsed
	}

	// Check if token has expired
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Get user
	u, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update user email and mark as verified
	u.Email = &token.Email
	u.EmailVerified = true
	if err := s.userRepo.Update(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Mark token as used
	if err := s.tokenRepo.MarkAsUsed(ctx, tokenStr); err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	return u, nil
}

func (s *emailVerificationService) ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// Get user
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check if already verified
	if u.EmailVerified {
		return errors.New("email is already verified")
	}

	// Get email from user or pending token
	var email string
	if u.Email != nil && *u.Email != "" {
		email = *u.Email
	} else {
		return errors.New("no email address found for user")
	}

	// Use username or display name
	name := u.Username
	if u.DisplayName != nil && *u.DisplayName != "" {
		name = *u.DisplayName
	}

	return s.SendVerificationEmail(ctx, userID, email, name)
}
