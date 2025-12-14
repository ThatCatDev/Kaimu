package invitation

//go:generate mockgen -source=invitation_service.go -destination=mocks/invitation_service_mock.go -package=mocks

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/config"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/invitation"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization_member"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/role"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/user"
	"github.com/thatcatdev/kaimu/backend/internal/services/mail"
	"github.com/thatcatdev/kaimu/backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const (
	// InvitationExpiry is the default expiration time for invitations
	InvitationExpiry = 7 * 24 * time.Hour // 7 days
	// TokenLength is the length of the invitation token in bytes (before base64 encoding)
	TokenLength = 32
)

var (
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInvitationExpired  = errors.New("invitation has expired")
	ErrInvitationAccepted = errors.New("invitation has already been accepted")
	ErrAlreadyMember      = errors.New("user is already a member of this organization")
	ErrPendingInvitation  = errors.New("there is already a pending invitation for this email")
	ErrEmailMismatch      = errors.New("your email does not match the invitation")
	ErrOrgNotFound        = errors.New("organization not found")
)

type Service interface {
	// Create a new invitation
	CreateInvitation(ctx context.Context, orgID uuid.UUID, email string, roleID uuid.UUID, invitedBy uuid.UUID) (*invitation.Invitation, error)

	// Get invitation by ID
	GetInvitation(ctx context.Context, id uuid.UUID) (*invitation.Invitation, error)

	// Get invitation by token
	GetInvitationByToken(ctx context.Context, token string) (*invitation.Invitation, error)

	// Get pending invitations for an organization
	GetPendingInvitations(ctx context.Context, orgID uuid.UUID) ([]*invitation.Invitation, error)

	// Cancel (delete) an invitation
	CancelInvitation(ctx context.Context, id uuid.UUID) error

	// Resend invitation (generates new token and extends expiration)
	ResendInvitation(ctx context.Context, id uuid.UUID) (*invitation.Invitation, error)

	// Accept an invitation (creates membership)
	AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*organization.Organization, error)

	// Get organization for invitation
	GetInvitationOrganization(ctx context.Context, invID uuid.UUID) (*organization.Organization, error)

	// Get role for invitation
	GetInvitationRole(ctx context.Context, invID uuid.UUID) (*role.Role, error)

	// Get inviter for invitation
	GetInviter(ctx context.Context, invID uuid.UUID) (*user.User, error)
}

type service struct {
	invitationRepo invitation.Repository
	orgRepo        organization.Repository
	orgMemberRepo  organization_member.Repository
	userRepo       user.Repository
	roleRepo       role.Repository
	mailService    mail.MailService
	emailConfig    config.EmailConfig
}

func NewService(
	invitationRepo invitation.Repository,
	orgRepo organization.Repository,
	orgMemberRepo organization_member.Repository,
	userRepo user.Repository,
	roleRepo role.Repository,
	mailService mail.MailService,
	emailConfig config.EmailConfig,
) Service {
	return &service{
		invitationRepo: invitationRepo,
		orgRepo:        orgRepo,
		orgMemberRepo:  orgMemberRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		mailService:    mailService,
		emailConfig:    emailConfig,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "invitation.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "invitation"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

// generateToken creates a secure random token for invitations
func generateToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *service) CreateInvitation(ctx context.Context, orgID uuid.UUID, email string, roleID uuid.UUID, invitedBy uuid.UUID) (*invitation.Invitation, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateInvitation")
	span.SetAttributes(
		attribute.String("org.id", orgID.String()),
		attribute.String("email", email),
		attribute.String("role.id", roleID.String()),
	)
	defer span.End()

	// Check if organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	// Check if user with this email is already a member
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		member, err := s.orgMemberRepo.GetByOrgAndUser(ctx, orgID, existingUser.ID)
		if err == nil && member != nil {
			return nil, ErrAlreadyMember
		}
	}

	// Check for existing pending invitation
	existing, err := s.invitationRepo.GetByOrgAndEmail(ctx, orgID, email)
	if err == nil && existing != nil && existing.IsPending() {
		return nil, ErrPendingInvitation
	}

	// Delete any expired/accepted invitation for this email
	if existing != nil {
		_ = s.invitationRepo.Delete(ctx, existing.ID)
	}

	// Generate token
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	inv := &invitation.Invitation{
		OrganizationID: orgID,
		Email:          email,
		RoleID:         &roleID,
		InvitedBy:      invitedBy,
		Token:          token,
		ExpiresAt:      time.Now().Add(InvitationExpiry),
	}

	if err := s.invitationRepo.Create(ctx, inv); err != nil {
		return nil, err
	}

	// Send invitation email asynchronously (use background context since request context will be canceled)
	go s.sendInvitationEmail(context.Background(), inv, invitedBy)

	return inv, nil
}

func (s *service) GetInvitation(ctx context.Context, id uuid.UUID) (*invitation.Invitation, error) {
	ctx, span := s.startServiceSpan(ctx, "GetInvitation")
	span.SetAttributes(attribute.String("invitation.id", id.String()))
	defer span.End()

	inv, err := s.invitationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}
	return inv, nil
}

func (s *service) GetInvitationByToken(ctx context.Context, token string) (*invitation.Invitation, error) {
	ctx, span := s.startServiceSpan(ctx, "GetInvitationByToken")
	defer span.End()

	inv, err := s.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}
	return inv, nil
}

func (s *service) GetPendingInvitations(ctx context.Context, orgID uuid.UUID) ([]*invitation.Invitation, error) {
	ctx, span := s.startServiceSpan(ctx, "GetPendingInvitations")
	span.SetAttributes(attribute.String("org.id", orgID.String()))
	defer span.End()

	return s.invitationRepo.GetPendingByOrgID(ctx, orgID)
}

func (s *service) CancelInvitation(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "CancelInvitation")
	span.SetAttributes(attribute.String("invitation.id", id.String()))
	defer span.End()

	// Verify invitation exists
	_, err := s.invitationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvitationNotFound
		}
		return err
	}

	return s.invitationRepo.Delete(ctx, id)
}

func (s *service) ResendInvitation(ctx context.Context, id uuid.UUID) (*invitation.Invitation, error) {
	ctx, span := s.startServiceSpan(ctx, "ResendInvitation")
	span.SetAttributes(attribute.String("invitation.id", id.String()))
	defer span.End()

	inv, err := s.invitationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	if inv.IsAccepted() {
		return nil, ErrInvitationAccepted
	}

	// Generate new token and extend expiration
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	inv.Token = token
	inv.ExpiresAt = time.Now().Add(InvitationExpiry)

	if err := s.invitationRepo.Update(ctx, inv); err != nil {
		return nil, err
	}

	// Send invitation email asynchronously (use background context since request context will be canceled)
	go s.sendInvitationEmail(context.Background(), inv, inv.InvitedBy)

	return inv, nil
}

func (s *service) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*organization.Organization, error) {
	ctx, span := s.startServiceSpan(ctx, "AcceptInvitation")
	span.SetAttributes(attribute.String("user.id", userID.String()))
	defer span.End()

	// Get invitation
	inv, err := s.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	// Check if already accepted
	if inv.IsAccepted() {
		return nil, ErrInvitationAccepted
	}

	// Check if expired
	if inv.IsExpired() {
		return nil, ErrInvitationExpired
	}

	// Optionally verify email matches (if user has email)
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Email != nil && *user.Email != "" && *user.Email != inv.Email {
		return nil, ErrEmailMismatch
	}

	// Check if already a member
	existingMember, err := s.orgMemberRepo.GetByOrgAndUser(ctx, inv.OrganizationID, userID)
	if err == nil && existingMember != nil {
		return nil, ErrAlreadyMember
	}

	// Create membership
	member := &organization_member.OrganizationMember{
		OrganizationID: inv.OrganizationID,
		UserID:         userID,
		RoleID:         inv.RoleID,
		Role:           "member", // Legacy field
	}

	if err := s.orgMemberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	// Mark invitation as accepted
	now := time.Now()
	inv.AcceptedAt = &now
	if err := s.invitationRepo.Update(ctx, inv); err != nil {
		return nil, err
	}

	// Return the organization
	return s.orgRepo.GetByID(ctx, inv.OrganizationID)
}

func (s *service) GetInvitationOrganization(ctx context.Context, invID uuid.UUID) (*organization.Organization, error) {
	ctx, span := s.startServiceSpan(ctx, "GetInvitationOrganization")
	span.SetAttributes(attribute.String("invitation.id", invID.String()))
	defer span.End()

	inv, err := s.invitationRepo.GetByID(ctx, invID)
	if err != nil {
		return nil, err
	}

	return s.orgRepo.GetByID(ctx, inv.OrganizationID)
}

func (s *service) GetInvitationRole(ctx context.Context, invID uuid.UUID) (*role.Role, error) {
	ctx, span := s.startServiceSpan(ctx, "GetInvitationRole")
	span.SetAttributes(attribute.String("invitation.id", invID.String()))
	defer span.End()

	inv, err := s.invitationRepo.GetByID(ctx, invID)
	if err != nil {
		return nil, err
	}

	if inv.RoleID == nil {
		return nil, nil
	}

	return s.roleRepo.GetByID(ctx, *inv.RoleID)
}

func (s *service) GetInviter(ctx context.Context, invID uuid.UUID) (*user.User, error) {
	ctx, span := s.startServiceSpan(ctx, "GetInviter")
	span.SetAttributes(attribute.String("invitation.id", invID.String()))
	defer span.End()

	inv, err := s.invitationRepo.GetByID(ctx, invID)
	if err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, inv.InvitedBy)
}

// sendInvitationEmail sends an invitation email to the invitee
func (s *service) sendInvitationEmail(ctx context.Context, inv *invitation.Invitation, invitedByID uuid.UUID) {
	// Get organization name
	org, err := s.orgRepo.GetByID(ctx, inv.OrganizationID)
	if err != nil {
		return // Silently fail - email is not critical
	}

	// Get inviter name
	inviter, err := s.userRepo.GetByID(ctx, invitedByID)
	if err != nil {
		return
	}
	inviterName := inviter.Username
	if inviter.DisplayName != nil && *inviter.DisplayName != "" {
		inviterName = *inviter.DisplayName
	}

	// Get role name
	roleName := "Member"
	if inv.RoleID != nil {
		role, err := s.roleRepo.GetByID(ctx, *inv.RoleID)
		if err == nil && role != nil {
			roleName = role.Name
		}
	}

	// Build invitation URL
	inviteURL := fmt.Sprintf("%s/%s", s.emailConfig.InvitationURL, inv.Token)

	// Send the email if mail service is configured
	if s.mailService == nil {
		return
	}
	err = s.mailService.SendMail(ctx, []string{inv.Email}, fmt.Sprintf("You've been invited to join %s", org.Name), "invitation.mjml", map[string]string{
		"organization_name": org.Name,
		"inviter_name":      inviterName,
		"role_name":         roleName,
		"invite_url":        inviteURL,
	})
	if err != nil {
		// Log error but don't fail - email is not critical
		return
	}
}
