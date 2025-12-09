package organization

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/organization"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/organization_member"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/user"
	"github.com/thatcatdev/pulse-backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrOrgNotFound     = errors.New("organization not found")
	ErrSlugTaken       = errors.New("organization slug already taken")
	ErrNotMember       = errors.New("user is not a member of this organization")
	ErrNotOwner        = errors.New("user is not the owner of this organization")
	ErrAlreadyMember   = errors.New("user is already a member of this organization")
	ErrCannotRemoveSelf = errors.New("cannot remove yourself from organization")
)

type Service interface {
	CreateOrganization(ctx context.Context, userID uuid.UUID, name, description string) (*organization.Organization, error)
	GetOrganization(ctx context.Context, id uuid.UUID) (*organization.Organization, error)
	GetOrganizationBySlug(ctx context.Context, slug string) (*organization.Organization, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*organization.Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	AddMember(ctx context.Context, orgID, userID uuid.UUID, role string) (*organization_member.OrganizationMember, error)
	RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error
	IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error)
	GetMembers(ctx context.Context, orgID uuid.UUID) ([]*organization_member.OrganizationMember, error)
	GetOwner(ctx context.Context, orgID uuid.UUID) (*user.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error)
}

type service struct {
	orgRepo    organization.Repository
	memberRepo organization_member.Repository
	userRepo   user.Repository
}

func NewService(
	orgRepo organization.Repository,
	memberRepo organization_member.Repository,
	userRepo user.Repository,
) Service {
	return &service{
		orgRepo:    orgRepo,
		memberRepo: memberRepo,
		userRepo:   userRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "organization.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "organization"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

// generateSlug creates a URL-friendly slug from a name
func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove any non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	// Trim leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

func (s *service) CreateOrganization(ctx context.Context, userID uuid.UUID, name, description string) (*organization.Organization, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateOrganization")
	span.SetAttributes(
		attribute.String("org.name", name),
		attribute.String("org.owner_id", userID.String()),
	)
	defer span.End()

	// Generate slug from name
	slug := generateSlug(name)
	if slug == "" {
		slug = uuid.New().String()[:8]
	}

	// Check if slug is taken
	existing, err := s.orgRepo.GetBySlug(ctx, slug)
	if err == nil && existing != nil {
		// Append random suffix to make unique
		slug = slug + "-" + uuid.New().String()[:4]
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	org := &organization.Organization{
		Name:        name,
		Slug:        slug,
		Description: description,
		OwnerID:     userID,
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	// Add owner as a member with "owner" role
	member := &organization_member.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         userID,
		Role:           "owner",
	}
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *service) GetOrganization(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	ctx, span := s.startServiceSpan(ctx, "GetOrganization")
	span.SetAttributes(attribute.String("org.id", id.String()))
	defer span.End()

	org, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return org, nil
}

func (s *service) GetOrganizationBySlug(ctx context.Context, slug string) (*organization.Organization, error) {
	ctx, span := s.startServiceSpan(ctx, "GetOrganizationBySlug")
	span.SetAttributes(attribute.String("org.slug", slug))
	defer span.End()

	org, err := s.orgRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return org, nil
}

func (s *service) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*organization.Organization, error) {
	ctx, span := s.startServiceSpan(ctx, "GetUserOrganizations")
	span.SetAttributes(attribute.String("user.id", userID.String()))
	defer span.End()

	return s.orgRepo.GetByUserID(ctx, userID)
}

func (s *service) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteOrganization")
	span.SetAttributes(attribute.String("org.id", id.String()))
	defer span.End()

	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrgNotFound
		}
		return err
	}

	return s.orgRepo.Delete(ctx, id)
}

func (s *service) AddMember(ctx context.Context, orgID, userID uuid.UUID, role string) (*organization_member.OrganizationMember, error) {
	ctx, span := s.startServiceSpan(ctx, "AddMember")
	span.SetAttributes(
		attribute.String("org.id", orgID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("member.role", role),
	)
	defer span.End()

	// Check if already a member
	existing, err := s.memberRepo.GetByOrgAndUser(ctx, orgID, userID)
	if err == nil && existing != nil {
		return nil, ErrAlreadyMember
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	member := &organization_member.OrganizationMember{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
	}

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *service) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "RemoveMember")
	span.SetAttributes(
		attribute.String("org.id", orgID.String()),
		attribute.String("user.id", userID.String()),
	)
	defer span.End()

	return s.memberRepo.Delete(ctx, orgID, userID)
}

func (s *service) IsMember(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	ctx, span := s.startServiceSpan(ctx, "IsMember")
	span.SetAttributes(
		attribute.String("org.id", orgID.String()),
		attribute.String("user.id", userID.String()),
	)
	defer span.End()

	member, err := s.memberRepo.GetByOrgAndUser(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return member != nil, nil
}

func (s *service) GetMembers(ctx context.Context, orgID uuid.UUID) ([]*organization_member.OrganizationMember, error) {
	ctx, span := s.startServiceSpan(ctx, "GetMembers")
	span.SetAttributes(attribute.String("org.id", orgID.String()))
	defer span.End()

	return s.memberRepo.GetByOrgID(ctx, orgID)
}

func (s *service) GetOwner(ctx context.Context, orgID uuid.UUID) (*user.User, error) {
	ctx, span := s.startServiceSpan(ctx, "GetOwner")
	span.SetAttributes(attribute.String("org.id", orgID.String()))
	defer span.End()

	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	return s.userRepo.GetByID(ctx, org.OwnerID)
}

func (s *service) GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
	ctx, span := s.startServiceSpan(ctx, "GetUserByID")
	span.SetAttributes(attribute.String("user.id", userID.String()))
	defer span.End()

	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return u, nil
}
