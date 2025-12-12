package metrics_history

//go:generate mockgen -source=metrics_history_repository.go -destination=mocks/metrics_history_repository_mock.go -package=mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Create(ctx context.Context, history *MetricsHistory) error
	Upsert(ctx context.Context, history *MetricsHistory) error
	GetBySprintID(ctx context.Context, sprintID uuid.UUID) ([]*MetricsHistory, error)
	GetBySprintIDAndDateRange(ctx context.Context, sprintID uuid.UUID, startDate, endDate time.Time) ([]*MetricsHistory, error)
	GetLatestBySprintID(ctx context.Context, sprintID uuid.UUID) (*MetricsHistory, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, history *MetricsHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// Upsert inserts or updates a metrics history record based on sprint_id and recorded_date
func (r *repository) Upsert(ctx context.Context, history *MetricsHistory) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sprint_id"}, {Name: "recorded_date"}},
		UpdateAll: true,
	}).Create(history).Error
}

func (r *repository) GetBySprintID(ctx context.Context, sprintID uuid.UUID) ([]*MetricsHistory, error) {
	var histories []*MetricsHistory
	err := r.db.WithContext(ctx).
		Where("sprint_id = ?", sprintID).
		Order("recorded_date ASC").
		Find(&histories).Error
	if err != nil {
		return nil, err
	}
	return histories, nil
}

func (r *repository) GetBySprintIDAndDateRange(ctx context.Context, sprintID uuid.UUID, startDate, endDate time.Time) ([]*MetricsHistory, error) {
	var histories []*MetricsHistory
	err := r.db.WithContext(ctx).
		Where("sprint_id = ? AND recorded_date >= ? AND recorded_date <= ?", sprintID, startDate, endDate).
		Order("recorded_date ASC").
		Find(&histories).Error
	if err != nil {
		return nil, err
	}
	return histories, nil
}

func (r *repository) GetLatestBySprintID(ctx context.Context, sprintID uuid.UUID) (*MetricsHistory, error) {
	var history MetricsHistory
	err := r.db.WithContext(ctx).
		Where("sprint_id = ?", sprintID).
		Order("recorded_date DESC").
		First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}
