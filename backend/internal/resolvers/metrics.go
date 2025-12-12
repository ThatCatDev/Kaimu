package resolvers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/graph/model"
	"github.com/thatcatdev/kaimu/backend/internal/services/metrics"
)

// MetricsResolver handles metrics-related GraphQL queries
type MetricsResolver struct {
	metricsService metrics.Service
}

// NewMetricsResolver creates a new metrics resolver
func NewMetricsResolver(metricsService metrics.Service) *MetricsResolver {
	return &MetricsResolver{
		metricsService: metricsService,
	}
}

// BurnDownData returns burn down chart data for a sprint
func (r *MetricsResolver) BurnDownData(ctx context.Context, sprintID string, mode model.MetricMode) (*model.BurnDownData, error) {
	id, err := uuid.Parse(sprintID)
	if err != nil {
		return nil, err
	}

	metricsMode := metrics.MetricModeCardCount
	if mode == model.MetricModeStoryPoints {
		metricsMode = metrics.MetricModeStoryPoints
	}

	data, err := r.metricsService.GetBurnDownData(ctx, id, metricsMode)
	if err != nil {
		return nil, err
	}

	// Convert to GraphQL model
	idealLine := make([]*model.DataPoint, len(data.IdealLine))
	for i, p := range data.IdealLine {
		idealLine[i] = &model.DataPoint{
			Date:  p.Date,
			Value: p.Value,
		}
	}

	actualLine := make([]*model.DataPoint, len(data.ActualLine))
	for i, p := range data.ActualLine {
		actualLine[i] = &model.DataPoint{
			Date:  p.Date,
			Value: p.Value,
		}
	}

	return &model.BurnDownData{
		SprintID:   data.SprintID.String(),
		SprintName: data.SprintName,
		StartDate:  data.StartDate,
		EndDate:    data.EndDate,
		IdealLine:  idealLine,
		ActualLine: actualLine,
	}, nil
}

// BurnUpData returns burn up chart data for a sprint
func (r *MetricsResolver) BurnUpData(ctx context.Context, sprintID string, mode model.MetricMode) (*model.BurnUpData, error) {
	id, err := uuid.Parse(sprintID)
	if err != nil {
		return nil, err
	}

	metricsMode := metrics.MetricModeCardCount
	if mode == model.MetricModeStoryPoints {
		metricsMode = metrics.MetricModeStoryPoints
	}

	data, err := r.metricsService.GetBurnUpData(ctx, id, metricsMode)
	if err != nil {
		return nil, err
	}

	// Convert to GraphQL model
	scopeLine := make([]*model.DataPoint, len(data.ScopeLine))
	for i, p := range data.ScopeLine {
		scopeLine[i] = &model.DataPoint{
			Date:  p.Date,
			Value: p.Value,
		}
	}

	doneLine := make([]*model.DataPoint, len(data.DoneLine))
	for i, p := range data.DoneLine {
		doneLine[i] = &model.DataPoint{
			Date:  p.Date,
			Value: p.Value,
		}
	}

	return &model.BurnUpData{
		SprintID:   data.SprintID.String(),
		SprintName: data.SprintName,
		StartDate:  data.StartDate,
		EndDate:    data.EndDate,
		ScopeLine:  scopeLine,
		DoneLine:   doneLine,
	}, nil
}

// VelocityData returns velocity data for closed sprints on a board
func (r *MetricsResolver) VelocityData(ctx context.Context, boardID string, sprintCount *int, mode model.MetricMode) (*model.VelocityData, error) {
	id, err := uuid.Parse(boardID)
	if err != nil {
		return nil, err
	}

	count := 10
	if sprintCount != nil {
		count = *sprintCount
	}

	metricsMode := metrics.MetricModeCardCount
	if mode == model.MetricModeStoryPoints {
		metricsMode = metrics.MetricModeStoryPoints
	}

	data, err := r.metricsService.GetVelocityData(ctx, id, count, metricsMode)
	if err != nil {
		return nil, err
	}

	// Convert to GraphQL model
	sprints := make([]*model.SprintVelocity, len(data.Sprints))
	for i, sv := range data.Sprints {
		sprints[i] = &model.SprintVelocity{
			SprintID:        sv.SprintID.String(),
			SprintName:      sv.SprintName,
			CompletedCards:  sv.CompletedCards,
			CompletedPoints: sv.CompletedPoints,
		}
	}

	return &model.VelocityData{
		Sprints: sprints,
	}, nil
}

// CumulativeFlowData returns cumulative flow diagram data for a sprint
func (r *MetricsResolver) CumulativeFlowData(ctx context.Context, sprintID string, mode model.MetricMode) (*model.CumulativeFlowData, error) {
	id, err := uuid.Parse(sprintID)
	if err != nil {
		return nil, err
	}

	metricsMode := metrics.MetricModeCardCount
	if mode == model.MetricModeStoryPoints {
		metricsMode = metrics.MetricModeStoryPoints
	}

	data, err := r.metricsService.GetCumulativeFlowData(ctx, id, metricsMode)
	if err != nil {
		return nil, err
	}

	// Convert to GraphQL model
	columns := make([]*model.ColumnFlowData, len(data.Columns))
	for i, col := range data.Columns {
		columns[i] = &model.ColumnFlowData{
			ColumnID:   col.ColumnID.String(),
			ColumnName: col.ColumnName,
			Color:      col.Color,
			Values:     col.Values,
		}
	}

	// Convert dates to time.Time pointers
	dates := make([]*time.Time, len(data.Dates))
	for i := range data.Dates {
		d := data.Dates[i]
		dates[i] = &d
	}

	return &model.CumulativeFlowData{
		SprintID:   data.SprintID.String(),
		SprintName: data.SprintName,
		Columns:    columns,
		Dates:      dates,
	}, nil
}

// SprintStats returns current statistics for a sprint
func (r *MetricsResolver) SprintStats(ctx context.Context, sprintID string) (*model.SprintStats, error) {
	id, err := uuid.Parse(sprintID)
	if err != nil {
		return nil, err
	}

	stats, err := r.metricsService.GetSprintStats(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.SprintStats{
		TotalCards:           stats.TotalCards,
		CompletedCards:       stats.CompletedCards,
		TotalStoryPoints:     stats.TotalStoryPoints,
		CompletedStoryPoints: stats.CompletedStoryPoints,
		DaysRemaining:        stats.DaysRemaining,
		DaysElapsed:          stats.DaysElapsed,
	}, nil
}
