package label

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/label"
	"github.com/thatcatdev/pulse-backend/internal/db/repositories/project"
	"github.com/thatcatdev/pulse-backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrLabelNotFound   = errors.New("label not found")
	ErrProjectNotFound = errors.New("project not found")
	ErrLabelNameTaken  = errors.New("label name already exists in this project")
)

type Service interface {
	CreateLabel(ctx context.Context, projectID uuid.UUID, name, color, description string) (*label.Label, error)
	GetLabel(ctx context.Context, id uuid.UUID) (*label.Label, error)
	GetLabelsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*label.Label, error)
	GetLabelsByIDs(ctx context.Context, ids []uuid.UUID) ([]*label.Label, error)
	UpdateLabel(ctx context.Context, l *label.Label) (*label.Label, error)
	DeleteLabel(ctx context.Context, id uuid.UUID) error
	GetProject(ctx context.Context, labelID uuid.UUID) (*project.Project, error)
}

type service struct {
	labelRepo   label.Repository
	projectRepo project.Repository
}

func NewService(labelRepo label.Repository, projectRepo project.Repository) Service {
	return &service{
		labelRepo:   labelRepo,
		projectRepo: projectRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "label.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "label"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

func (s *service) CreateLabel(ctx context.Context, projectID uuid.UUID, name, color, description string) (*label.Label, error) {
	ctx, span := s.startServiceSpan(ctx, "CreateLabel")
	span.SetAttributes(
		attribute.String("label.project_id", projectID.String()),
		attribute.String("label.name", name),
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

	// Check if label name is already taken
	existing, err := s.labelRepo.GetByName(ctx, projectID, name)
	if err == nil && existing != nil {
		return nil, ErrLabelNameTaken
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	l := &label.Label{
		ProjectID:   projectID,
		Name:        name,
		Color:       color,
		Description: description,
	}

	if l.Color == "" {
		l.Color = "#6B7280"
	}

	if err := s.labelRepo.Create(ctx, l); err != nil {
		return nil, err
	}

	return l, nil
}

func (s *service) GetLabel(ctx context.Context, id uuid.UUID) (*label.Label, error) {
	ctx, span := s.startServiceSpan(ctx, "GetLabel")
	span.SetAttributes(attribute.String("label.id", id.String()))
	defer span.End()

	l, err := s.labelRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLabelNotFound
		}
		return nil, err
	}
	return l, nil
}

func (s *service) GetLabelsByProjectID(ctx context.Context, projectID uuid.UUID) ([]*label.Label, error) {
	ctx, span := s.startServiceSpan(ctx, "GetLabelsByProjectID")
	span.SetAttributes(attribute.String("label.project_id", projectID.String()))
	defer span.End()

	return s.labelRepo.GetByProjectID(ctx, projectID)
}

func (s *service) GetLabelsByIDs(ctx context.Context, ids []uuid.UUID) ([]*label.Label, error) {
	ctx, span := s.startServiceSpan(ctx, "GetLabelsByIDs")
	defer span.End()

	return s.labelRepo.GetByIDs(ctx, ids)
}

func (s *service) UpdateLabel(ctx context.Context, l *label.Label) (*label.Label, error) {
	ctx, span := s.startServiceSpan(ctx, "UpdateLabel")
	span.SetAttributes(attribute.String("label.id", l.ID.String()))
	defer span.End()

	// Check if new name conflicts with existing label
	existing, err := s.labelRepo.GetByName(ctx, l.ProjectID, l.Name)
	if err == nil && existing != nil && existing.ID != l.ID {
		return nil, ErrLabelNameTaken
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.labelRepo.Update(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (s *service) DeleteLabel(ctx context.Context, id uuid.UUID) error {
	ctx, span := s.startServiceSpan(ctx, "DeleteLabel")
	span.SetAttributes(attribute.String("label.id", id.String()))
	defer span.End()

	return s.labelRepo.Delete(ctx, id)
}

func (s *service) GetProject(ctx context.Context, labelID uuid.UUID) (*project.Project, error) {
	ctx, span := s.startServiceSpan(ctx, "GetProject")
	span.SetAttributes(attribute.String("label.id", labelID.String()))
	defer span.End()

	l, err := s.labelRepo.GetByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLabelNotFound
		}
		return nil, err
	}

	proj, err := s.projectRepo.GetByID(ctx, l.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	return proj, nil
}
