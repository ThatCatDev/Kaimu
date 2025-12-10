package project

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/organization"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/project"
	"github.com/thatcatdev/kaimu/backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrKeyTaken        = errors.New("project key already taken in this organization")
	ErrInvalidKey      = errors.New("project key must be 2-10 uppercase letters")
	ErrOrgNotFound     = errors.New("organization not found")
)

type Service interface {
	CreateProject(ctx context.Context, orgID uuid.UUID, name, key, description string) (*project.Project, error)
	GetProject(ctx context.Context, id uuid.UUID) (*project.Project, error)
	GetProjectByKey(ctx context.Context, orgID uuid.UUID, key string) (*project.Project, error)
	GetOrgProjects(ctx context.Context, orgID uuid.UUID) ([]*project.Project, error)
	UpdateProject(ctx context.Context, proj *project.Project) (*project.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
	GetOrganization(ctx context.Context, projectID uuid.UUID) (*organization.Organization, error)
}

type service struct {
	projectRepo project.Repository
	orgRepo     organization.Repository
}

func NewService(projectRepo project.Repository, orgRepo organization.Repository) Service {
	return &service{
		projectRepo: projectRepo,
		orgRepo:     orgRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "project.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "project"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

// validateKey checks that the key is 2-10 uppercase letters
func validateKey(key string) error {
	if len(key) < 2 || len(key) > 10 {
		return ErrInvalidKey
	}
	for _, c := range key {
		if c < 'A' || c > 'Z' {
			return ErrInvalidKey
		}
	}
	return nil
}

func (s *service) CreateProject(ctx context.Context, orgID uuid.UUID, name, key, description string) (*project.Project, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateProject")
	span.SetAttributes(
		attribute.String("project.name", name),
		attribute.String("project.key", key),
		attribute.String("project.org_id", orgID.String()),
	)
	defer span.End()

	// Normalize key to uppercase
	key = strings.ToUpper(key)

	// Validate key format
	if err := validateKey(key); err != nil {
		return nil, err
	}

	// Verify organization exists
	_, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	// Check if key is already taken in this org
	existing, err := s.projectRepo.GetByKey(ctx, orgID, key)
	if err == nil && existing != nil {
		return nil, ErrKeyTaken
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	proj := &project.Project{
		OrganizationID: orgID,
		Name:           name,
		Key:            key,
		Description:    description,
	}

	if err := s.projectRepo.Create(ctx, proj); err != nil {
		return nil, err
	}

	return proj, nil
}

func (s *service) GetProject(ctx context.Context, id uuid.UUID) (*project.Project, error) {
	ctx, span := s.startServiceSpan(ctx, "GetProject")
	span.SetAttributes(attribute.String("project.id", id.String()))
	defer span.End()

	proj, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return proj, nil
}

func (s *service) GetProjectByKey(ctx context.Context, orgID uuid.UUID, key string) (*project.Project, error) {
	ctx, span := s.startServiceSpan(ctx, "GetProjectByKey")
	span.SetAttributes(
		attribute.String("project.key", key),
		attribute.String("project.org_id", orgID.String()),
	)
	defer span.End()

	proj, err := s.projectRepo.GetByKey(ctx, orgID, strings.ToUpper(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return proj, nil
}

func (s *service) GetOrgProjects(ctx context.Context, orgID uuid.UUID) ([]*project.Project, error) {
	ctx, span := s.startServiceSpan(ctx, "GetOrgProjects")
	span.SetAttributes(attribute.String("project.org_id", orgID.String()))
	defer span.End()

	return s.projectRepo.GetByOrgID(ctx, orgID)
}

func (s *service) UpdateProject(ctx context.Context, proj *project.Project) (*project.Project, error) {
	ctx, span := s.startServiceSpan(ctx, "UpdateProject")
	span.SetAttributes(attribute.String("project.id", proj.ID.String()))
	defer span.End()

	if err := s.projectRepo.Update(ctx, proj); err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *service) DeleteProject(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteProject")
	span.SetAttributes(attribute.String("project.id", id.String()))
	defer span.End()

	return s.projectRepo.Delete(ctx, id)
}

func (s *service) GetOrganization(ctx context.Context, projectID uuid.UUID) (*organization.Organization, error) {
	ctx, span := s.startServiceSpan(ctx, "GetOrganization")
	span.SetAttributes(attribute.String("project.id", projectID.String()))
	defer span.End()

	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	org, err := s.orgRepo.GetByID(ctx, proj.OrganizationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	return org, nil
}
