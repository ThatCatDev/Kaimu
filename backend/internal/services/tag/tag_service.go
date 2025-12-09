package tag

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/tag"
	"github.com/thatcatdev/pulse-backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrTagNotFound     = errors.New("tag not found")
	ErrProjectNotFound = errors.New("project not found")
	ErrTagNameTaken    = errors.New("tag name already exists in this project")
)

type Service interface {
	CreateTag(ctx context.Context, projectID uuid.UUID, name, color, description string) (*tag.Tag, error)
	GetTag(ctx context.Context, id uuid.UUID) (*tag.Tag, error)
	GetTagsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*tag.Tag, error)
	GetTagsByIDs(ctx context.Context, ids []uuid.UUID) ([]*tag.Tag, error)
	UpdateTag(ctx context.Context, t *tag.Tag) (*tag.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
	GetProject(ctx context.Context, tagID uuid.UUID) (*project.Project, error)
}

type service struct {
	tagRepo     tag.Repository
	projectRepo project.Repository
}

func NewService(tagRepo tag.Repository, projectRepo project.Repository) Service {
	return &service{
		tagRepo:     tagRepo,
		projectRepo: projectRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "tag.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "tag"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

func (s *service) CreateTag(ctx context.Context, projectID uuid.UUID, name, color, description string) (*tag.Tag, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateTag")
	span.SetAttributes(
		attribute.String("tag.project_id", projectID.String()),
		attribute.String("tag.name", name),
	)
	defer span.End()

	// Verify project exists
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	// Check if tag name is already taken
	existing, err := s.tagRepo.GetByName(ctx, projectID, name)
	if err == nil && existing != nil {
		return nil, ErrTagNameTaken
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	t := &tag.Tag{
		ProjectID:   projectID,
		Name:        name,
		Color:       color,
		Description: description,
	}

	if t.Color == "" {
		t.Color = "#6B7280"
	}

	if err := s.tagRepo.Create(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *service) GetTag(ctx context.Context, id uuid.UUID) (*tag.Tag, error) {
	ctx, span := s.startServiceSpan(ctx, "GetTag")
	span.SetAttributes(attribute.String("tag.id", id.String()))
	defer span.End()

	t, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}
	return t, nil
}

func (s *service) GetTagsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*tag.Tag, error) {
	ctx, span := s.startServiceSpan(ctx, "GetTagsByProjectID")
	span.SetAttributes(attribute.String("tag.project_id", projectID.String()))
	defer span.End()

	return s.tagRepo.GetByProjectID(ctx, projectID)
}

func (s *service) GetTagsByIDs(ctx context.Context, ids []uuid.UUID) ([]*tag.Tag, error) {
	ctx, span := s.startServiceSpan(ctx, "GetTagsByIDs")
	defer span.End()

	return s.tagRepo.GetByIDs(ctx, ids)
}

func (s *service) UpdateTag(ctx context.Context, t *tag.Tag) (*tag.Tag, error) {
	ctx, span := s.startServiceSpan(ctx, "UpdateTag")
	span.SetAttributes(attribute.String("tag.id", t.ID.String()))
	defer span.End()

	// Check if new name conflicts with existing tag
	existing, err := s.tagRepo.GetByName(ctx, t.ProjectID, t.Name)
	if err == nil && existing != nil && existing.ID != t.ID {
		return nil, ErrTagNameTaken
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.tagRepo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) DeleteTag(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteTag")
	span.SetAttributes(attribute.String("tag.id", id.String()))
	defer span.End()

	return s.tagRepo.Delete(ctx, id)
}

func (s *service) GetProject(ctx context.Context, tagID uuid.UUID) (*project.Project, error) {
	ctx, span := s.startServiceSpan(ctx, "GetProject")
	span.SetAttributes(attribute.String("tag.id", tagID.String()))
	defer span.End()

	t, err := s.tagRepo.GetByID(ctx, tagID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	proj, err := s.projectRepo.GetByID(ctx, t.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return proj, nil
}
