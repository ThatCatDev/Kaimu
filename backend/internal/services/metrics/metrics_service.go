package metrics

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/audit"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/board_column"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/card"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/metrics_history"
	"github.com/thatcatdev/kaimu/backend/internal/db/repositories/sprint"
	"github.com/thatcatdev/kaimu/backend/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	ErrSprintNotFound = errors.New("sprint not found")
	ErrBoardNotFound  = errors.New("board not found")
)

// MetricMode represents whether to use card count or story points
type MetricMode string

const (
	MetricModeCardCount   MetricMode = "CARD_COUNT"
	MetricModeStoryPoints MetricMode = "STORY_POINTS"
)

// DataPoint represents a single point on a chart
type DataPoint struct {
	Date  time.Time
	Value float64
}

// BurnDownData contains data for a burn down chart
type BurnDownData struct {
	SprintID   uuid.UUID
	SprintName string
	StartDate  time.Time
	EndDate    time.Time
	IdealLine  []DataPoint
	ActualLine []DataPoint
}

// BurnUpData contains data for a burn up chart
type BurnUpData struct {
	SprintID   uuid.UUID
	SprintName string
	StartDate  time.Time
	EndDate    time.Time
	ScopeLine  []DataPoint
	DoneLine   []DataPoint
}

// SprintVelocity represents velocity data for a single sprint
type SprintVelocity struct {
	SprintID        uuid.UUID
	SprintName      string
	CompletedCards  int
	CompletedPoints int
}

// VelocityData contains velocity data for multiple sprints
type VelocityData struct {
	Sprints []SprintVelocity
}

// ColumnFlowData represents flow data for a single column
type ColumnFlowData struct {
	ColumnID   uuid.UUID
	ColumnName string
	Color      string
	Values     []int
}

// CumulativeFlowData contains cumulative flow diagram data
type CumulativeFlowData struct {
	SprintID   uuid.UUID
	SprintName string
	Columns    []ColumnFlowData
	Dates      []time.Time
}

// SprintStats contains current statistics for a sprint
type SprintStats struct {
	TotalCards           int
	CompletedCards       int
	TotalStoryPoints     int
	CompletedStoryPoints int
	DaysRemaining        int
	DaysElapsed          int
}

type Service interface {
	// Snapshot operations
	RecordDailySnapshot(ctx context.Context, sprintID uuid.UUID) (*metrics_history.MetricsHistory, error)

	// Chart data queries
	GetBurnDownData(ctx context.Context, sprintID uuid.UUID, mode MetricMode) (*BurnDownData, error)
	GetBurnUpData(ctx context.Context, sprintID uuid.UUID, mode MetricMode) (*BurnUpData, error)
	GetVelocityData(ctx context.Context, boardID uuid.UUID, sprintCount int, mode MetricMode) (*VelocityData, error)
	GetCumulativeFlowData(ctx context.Context, sprintID uuid.UUID, mode MetricMode) (*CumulativeFlowData, error)

	// Current sprint stats
	GetSprintStats(ctx context.Context, sprintID uuid.UUID) (*SprintStats, error)
}

type service struct {
	sprintRepo      sprint.Repository
	cardRepo        card.Repository
	columnRepo      board_column.Repository
	metricsHistRepo metrics_history.Repository
	auditRepo       audit.Repository
}

func NewService(
	sprintRepo sprint.Repository,
	cardRepo card.Repository,
	columnRepo board_column.Repository,
	metricsHistRepo metrics_history.Repository,
	auditRepo audit.Repository,
) Service {
	return &service{
		sprintRepo:      sprintRepo,
		cardRepo:        cardRepo,
		columnRepo:      columnRepo,
		metricsHistRepo: metricsHistRepo,
		auditRepo:       auditRepo,
	}
}

func (s *service) startServiceSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	tracer := tracing.GetTracer(ctx)
	return tracer.Start(ctx, "metrics.service."+operationName,
		trace.WithAttributes(
			attribute.String("service", "metrics"),
			attribute.String("type", "service"),
			attribute.String("method", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
		tracing.GetEnvironmentAttribute(),
	)
}

// RecordDailySnapshot creates a snapshot of current sprint metrics
func (s *service) RecordDailySnapshot(ctx context.Context, sprintID uuid.UUID) (*metrics_history.MetricsHistory, error) {
	ctx, span := s.startServiceSpan(ctx, "RecordDailySnapshot")
	span.SetAttributes(attribute.String("sprint.id", sprintID.String()))
	defer span.End()

	// Get sprint
	sp, err := s.sprintRepo.GetByID(ctx, sprintID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	// Get all cards in the sprint
	cards, err := s.cardRepo.GetBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// Get all columns for the board to identify "done" columns
	columns, err := s.columnRepo.GetByBoardID(ctx, sp.BoardID)
	if err != nil {
		return nil, err
	}

	// Build a set of "done" column IDs
	doneColumnIDs := make(map[uuid.UUID]bool)
	columnMap := make(map[uuid.UUID]*board_column.BoardColumn)
	for _, col := range columns {
		columnMap[col.ID] = col
		if col.IsDone {
			doneColumnIDs[col.ID] = true
		}
	}

	// Calculate metrics
	var totalCards, completedCards int
	var totalStoryPoints, completedStoryPoints int
	columnSnapshot := make(map[string]metrics_history.ColumnSnapshotData)

	for _, c := range cards {
		totalCards++
		if c.StoryPoints != nil {
			totalStoryPoints += *c.StoryPoints
		}

		// Check if card is in a "done" column
		if doneColumnIDs[c.ColumnID] {
			completedCards++
			if c.StoryPoints != nil {
				completedStoryPoints += *c.StoryPoints
			}
		}

		// Update column snapshot
		colID := c.ColumnID.String()
		snap := columnSnapshot[colID]
		if col, ok := columnMap[c.ColumnID]; ok {
			snap.Name = col.Name
		}
		snap.CardCount++
		if c.StoryPoints != nil {
			snap.StoryPoints += *c.StoryPoints
		}
		columnSnapshot[colID] = snap
	}

	// Create metrics history record
	history := &metrics_history.MetricsHistory{
		SprintID:             sprintID,
		RecordedDate:         time.Now().Truncate(24 * time.Hour),
		TotalCards:           totalCards,
		CompletedCards:       completedCards,
		TotalStoryPoints:     totalStoryPoints,
		CompletedStoryPoints: completedStoryPoints,
	}
	if err := history.SetColumnSnapshot(columnSnapshot); err != nil {
		return nil, err
	}

	// Upsert (in case we already have a record for today)
	if err := s.metricsHistRepo.Upsert(ctx, history); err != nil {
		return nil, err
	}

	return history, nil
}

// cardState tracks a card's column and story points for burn chart calculation
type cardState struct {
	columnID    uuid.UUID
	storyPoints int
	inSprint    bool
}

// cardMovedMetadata represents the metadata stored in card_moved audit events
type cardMovedMetadata struct {
	FromColumnID string `json:"from_column_id"`
	ToColumnID   string `json:"to_column_id"`
}

// GetBurnDownData returns burn down chart data for a sprint using audit events
func (s *service) GetBurnDownData(ctx context.Context, sprintID uuid.UUID, mode MetricMode) (*BurnDownData, error) {
	ctx, span := s.startServiceSpan(ctx, "GetBurnDownData")
	span.SetAttributes(
		attribute.String("sprint.id", sprintID.String()),
		attribute.String("mode", string(mode)),
	)
	defer span.End()

	// Get sprint
	sp, err := s.sprintRepo.GetByID(ctx, sprintID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	// Determine date range
	startDate := sp.StartDate
	endDate := sp.EndDate
	if startDate == nil {
		startDate = &sp.CreatedAt
	}
	if endDate == nil {
		end := startDate.Add(14 * 24 * time.Hour)
		endDate = &end
	}

	// Get all columns for the board to identify "done" columns
	columns, err := s.columnRepo.GetByBoardID(ctx, sp.BoardID)
	if err != nil {
		return nil, err
	}

	doneColumnIDs := make(map[uuid.UUID]bool)
	for _, col := range columns {
		if col.IsDone {
			doneColumnIDs[col.ID] = true
		}
	}

	// Get current cards in sprint - this is our "end state"
	currentCards, err := s.cardRepo.GetBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// Build current state map
	currentState := make(map[uuid.UUID]*cardState)
	for _, c := range currentCards {
		sp := 0
		if c.StoryPoints != nil {
			sp = *c.StoryPoints
		}
		currentState[c.ID] = &cardState{
			columnID:    c.ColumnID,
			storyPoints: sp,
			inSprint:    true,
		}
	}

	// Get audit events for this board in the date range
	auditEvents, err := s.auditRepo.GetCardMovementsByBoardAndDateRange(ctx, sp.BoardID, *startDate, endDate.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	// Calculate total work from current state for ideal line
	var totalWork float64
	for _, cs := range currentState {
		if mode == MetricModeStoryPoints {
			totalWork += float64(cs.storyPoints)
		} else {
			totalWork++
		}
	}

	// Generate dates from start to end
	dates := generateDateRange(*startDate, *endDate)
	idealLine := make([]DataPoint, len(dates))
	for i, date := range dates {
		progress := float64(i) / float64(len(dates)-1)
		idealLine[i] = DataPoint{
			Date:  date,
			Value: totalWork * (1 - progress),
		}
	}

	// Build actual line by replaying events to calculate state at each day
	actualLine := s.calculateBurnFromAuditEvents(currentState, auditEvents, dates, doneColumnIDs, mode, sprintID)

	return &BurnDownData{
		SprintID:   sprintID,
		SprintName: sp.Name,
		StartDate:  *startDate,
		EndDate:    *endDate,
		IdealLine:  idealLine,
		ActualLine: actualLine,
	}, nil
}

// calculateBurnFromAuditEvents replays audit events backwards to reconstruct state at each date
func (s *service) calculateBurnFromAuditEvents(
	currentState map[uuid.UUID]*cardState,
	auditEvents []*audit.AuditEvent,
	dates []time.Time,
	doneColumnIDs map[uuid.UUID]bool,
	mode MetricMode,
	sprintID uuid.UUID,
) []DataPoint {
	// Sort events by time descending (most recent first) for backward replay
	sortedEvents := make([]*audit.AuditEvent, len(auditEvents))
	copy(sortedEvents, auditEvents)
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].OccurredAt.After(sortedEvents[j].OccurredAt)
	})

	// Create a deep copy of current state that we'll modify as we go backwards
	stateAtDate := make(map[uuid.UUID]*cardState)
	for id, cs := range currentState {
		stateAtDate[id] = &cardState{
			columnID:    cs.columnID,
			storyPoints: cs.storyPoints,
			inSprint:    cs.inSprint,
		}
	}

	// Calculate remaining work at current state (end of timeline)
	calculateRemaining := func(state map[uuid.UUID]*cardState) float64 {
		var remaining float64
		for _, cs := range state {
			if !cs.inSprint {
				continue
			}
			// Remaining = not in done columns
			if !doneColumnIDs[cs.columnID] {
				if mode == MetricModeStoryPoints {
					remaining += float64(cs.storyPoints)
				} else {
					remaining++
				}
			}
		}
		return remaining
	}

	// Build results from end to start
	results := make([]DataPoint, len(dates))
	eventIdx := 0

	// Process dates from end to start
	for i := len(dates) - 1; i >= 0; i-- {
		date := dates[i]

		// Apply events that happened after this date (in reverse)
		for eventIdx < len(sortedEvents) {
			evt := sortedEvents[eventIdx]
			evtDate := evt.OccurredAt.Truncate(24 * time.Hour)

			// If event is on or before this date, stop
			if !evtDate.After(date) {
				break
			}

			// Reverse the event to get prior state
			s.reverseAuditEvent(stateAtDate, evt, sprintID)
			eventIdx++
		}

		// Calculate remaining work at this date
		remaining := calculateRemaining(stateAtDate)
		results[i] = DataPoint{
			Date:  date,
			Value: remaining,
		}
	}

	return results
}

// reverseAuditEvent reverses an audit event to get the previous state
func (s *service) reverseAuditEvent(state map[uuid.UUID]*cardState, evt *audit.AuditEvent, sprintID uuid.UUID) {
	cardID := evt.EntityID

	switch evt.Action {
	case audit.ActionCardMoved:
		// Reverse a move: put card back in the "from" column
		if evt.Metadata != nil {
			var meta cardMovedMetadata
			if err := json.Unmarshal(evt.Metadata, &meta); err == nil {
				fromColID, err := uuid.Parse(meta.FromColumnID)
				if err == nil {
					if cs, ok := state[cardID]; ok {
						cs.columnID = fromColID
					}
				}
			}
		}

	case audit.ActionCreated:
		// Reverse a create: card didn't exist before
		delete(state, cardID)

	case audit.ActionDeleted:
		// Reverse a delete: card existed before
		// Try to get state from stateBefore
		if evt.StateBefore != nil {
			var cardData struct {
				ColumnID    string `json:"column_id"`
				StoryPoints *int   `json:"story_points"`
			}
			if err := json.Unmarshal(evt.StateBefore, &cardData); err == nil {
				colID, _ := uuid.Parse(cardData.ColumnID)
				sp := 0
				if cardData.StoryPoints != nil {
					sp = *cardData.StoryPoints
				}
				state[cardID] = &cardState{
					columnID:    colID,
					storyPoints: sp,
					inSprint:    true,
				}
			}
		}

	case audit.ActionCardAddedToSprint:
		// Reverse: card was not in this sprint before
		if cs, ok := state[cardID]; ok {
			cs.inSprint = false
		}

	case audit.ActionCardRemovedFromSprint:
		// Reverse: card was in this sprint before
		if cs, ok := state[cardID]; ok {
			cs.inSprint = true
		} else {
			// Card doesn't exist in state, need to reconstruct from event
			if evt.StateBefore != nil {
				var cardData struct {
					ColumnID    string `json:"column_id"`
					StoryPoints *int   `json:"story_points"`
				}
				if err := json.Unmarshal(evt.StateBefore, &cardData); err == nil {
					colID, _ := uuid.Parse(cardData.ColumnID)
					sp := 0
					if cardData.StoryPoints != nil {
						sp = *cardData.StoryPoints
					}
					state[cardID] = &cardState{
						columnID:    colID,
						storyPoints: sp,
						inSprint:    true,
					}
				}
			}
		}
	}
}

// GetBurnUpData returns burn up chart data for a sprint using audit events
func (s *service) GetBurnUpData(ctx context.Context, sprintID uuid.UUID, mode MetricMode) (*BurnUpData, error) {
	ctx, span := s.startServiceSpan(ctx, "GetBurnUpData")
	span.SetAttributes(
		attribute.String("sprint.id", sprintID.String()),
		attribute.String("mode", string(mode)),
	)
	defer span.End()

	// Get sprint
	sp, err := s.sprintRepo.GetByID(ctx, sprintID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	// Determine date range
	startDate := sp.StartDate
	endDate := sp.EndDate
	if startDate == nil {
		startDate = &sp.CreatedAt
	}
	if endDate == nil {
		end := startDate.Add(14 * 24 * time.Hour)
		endDate = &end
	}

	// Get all columns for the board to identify "done" columns
	columns, err := s.columnRepo.GetByBoardID(ctx, sp.BoardID)
	if err != nil {
		return nil, err
	}

	doneColumnIDs := make(map[uuid.UUID]bool)
	for _, col := range columns {
		if col.IsDone {
			doneColumnIDs[col.ID] = true
		}
	}

	// Get current cards in sprint - this is our "end state"
	currentCards, err := s.cardRepo.GetBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// Build current state map
	currentState := make(map[uuid.UUID]*cardState)
	for _, c := range currentCards {
		sp := 0
		if c.StoryPoints != nil {
			sp = *c.StoryPoints
		}
		currentState[c.ID] = &cardState{
			columnID:    c.ColumnID,
			storyPoints: sp,
			inSprint:    true,
		}
	}

	// Get audit events for this board in the date range
	auditEvents, err := s.auditRepo.GetCardMovementsByBoardAndDateRange(ctx, sp.BoardID, *startDate, endDate.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}

	// Generate dates from start to end
	dates := generateDateRange(*startDate, *endDate)

	// Build scope and done lines by replaying events
	scopeLine, doneLine := s.calculateBurnUpFromAuditEvents(currentState, auditEvents, dates, doneColumnIDs, mode, sprintID)

	return &BurnUpData{
		SprintID:   sprintID,
		SprintName: sp.Name,
		StartDate:  *startDate,
		EndDate:    *endDate,
		ScopeLine:  scopeLine,
		DoneLine:   doneLine,
	}, nil
}

// calculateBurnUpFromAuditEvents replays audit events backwards to reconstruct state at each date
func (s *service) calculateBurnUpFromAuditEvents(
	currentState map[uuid.UUID]*cardState,
	auditEvents []*audit.AuditEvent,
	dates []time.Time,
	doneColumnIDs map[uuid.UUID]bool,
	mode MetricMode,
	sprintID uuid.UUID,
) ([]DataPoint, []DataPoint) {
	// Sort events by time descending (most recent first) for backward replay
	sortedEvents := make([]*audit.AuditEvent, len(auditEvents))
	copy(sortedEvents, auditEvents)
	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].OccurredAt.After(sortedEvents[j].OccurredAt)
	})

	// Create a deep copy of current state that we'll modify as we go backwards
	stateAtDate := make(map[uuid.UUID]*cardState)
	for id, cs := range currentState {
		stateAtDate[id] = &cardState{
			columnID:    cs.columnID,
			storyPoints: cs.storyPoints,
			inSprint:    cs.inSprint,
		}
	}

	// Calculate scope (total in sprint) and done (completed) at a state
	calculateScopeAndDone := func(state map[uuid.UUID]*cardState) (float64, float64) {
		var scope, done float64
		for _, cs := range state {
			if !cs.inSprint {
				continue
			}
			// Scope = all cards in sprint
			if mode == MetricModeStoryPoints {
				scope += float64(cs.storyPoints)
			} else {
				scope++
			}
			// Done = cards in done columns
			if doneColumnIDs[cs.columnID] {
				if mode == MetricModeStoryPoints {
					done += float64(cs.storyPoints)
				} else {
					done++
				}
			}
		}
		return scope, done
	}

	// Build results from end to start
	scopeLine := make([]DataPoint, len(dates))
	doneLine := make([]DataPoint, len(dates))
	eventIdx := 0

	// Process dates from end to start
	for i := len(dates) - 1; i >= 0; i-- {
		date := dates[i]

		// Apply events that happened after this date (in reverse)
		for eventIdx < len(sortedEvents) {
			evt := sortedEvents[eventIdx]
			evtDate := evt.OccurredAt.Truncate(24 * time.Hour)

			// If event is on or before this date, stop
			if !evtDate.After(date) {
				break
			}

			// Reverse the event to get prior state
			s.reverseAuditEvent(stateAtDate, evt, sprintID)
			eventIdx++
		}

		// Calculate scope and done at this date
		scope, done := calculateScopeAndDone(stateAtDate)
		scopeLine[i] = DataPoint{
			Date:  date,
			Value: scope,
		}
		doneLine[i] = DataPoint{
			Date:  date,
			Value: done,
		}
	}

	return scopeLine, doneLine
}

// GetVelocityData returns velocity data for closed sprints on a board
func (s *service) GetVelocityData(ctx context.Context, boardID uuid.UUID, sprintCount int, mode MetricMode) (*VelocityData, error) {
	ctx, span := s.startServiceSpan(ctx, "GetVelocityData")
	span.SetAttributes(
		attribute.String("board.id", boardID.String()),
		attribute.Int("sprint_count", sprintCount),
		attribute.String("mode", string(mode)),
	)
	defer span.End()

	// Get closed sprints (most recent first)
	closedSprints, _, err := s.sprintRepo.GetClosedByBoardIDPaginated(ctx, boardID, sprintCount, 0)
	if err != nil {
		return nil, err
	}

	// Calculate velocity for each sprint
	velocities := make([]SprintVelocity, 0, len(closedSprints))
	for _, sp := range closedSprints {
		// Get the final snapshot for this sprint
		history, err := s.metricsHistRepo.GetLatestBySprintID(ctx, sp.ID)
		if err != nil {
			// If no history, calculate from current state
			history = &metrics_history.MetricsHistory{}
			cards, cardErr := s.cardRepo.GetBySprintID(ctx, sp.ID)
			if cardErr == nil {
				columns, _ := s.columnRepo.GetByBoardID(ctx, sp.BoardID)
				doneColumnIDs := make(map[uuid.UUID]bool)
				for _, col := range columns {
					if col.IsDone {
						doneColumnIDs[col.ID] = true
					}
				}
				for _, c := range cards {
					if doneColumnIDs[c.ColumnID] {
						history.CompletedCards++
						if c.StoryPoints != nil {
							history.CompletedStoryPoints += *c.StoryPoints
						}
					}
				}
			}
		}

		velocities = append(velocities, SprintVelocity{
			SprintID:        sp.ID,
			SprintName:      sp.Name,
			CompletedCards:  history.CompletedCards,
			CompletedPoints: history.CompletedStoryPoints,
		})
	}

	// Reverse to show oldest first (chronological order)
	for i, j := 0, len(velocities)-1; i < j; i, j = i+1, j-1 {
		velocities[i], velocities[j] = velocities[j], velocities[i]
	}

	return &VelocityData{Sprints: velocities}, nil
}

// GetCumulativeFlowData returns cumulative flow diagram data for a sprint
func (s *service) GetCumulativeFlowData(ctx context.Context, sprintID uuid.UUID, mode MetricMode) (*CumulativeFlowData, error) {
	ctx, span := s.startServiceSpan(ctx, "GetCumulativeFlowData")
	span.SetAttributes(
		attribute.String("sprint.id", sprintID.String()),
		attribute.String("mode", string(mode)),
	)
	defer span.End()

	// Get sprint
	sp, err := s.sprintRepo.GetByID(ctx, sprintID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	// Get all columns for the board
	columns, err := s.columnRepo.GetByBoardID(ctx, sp.BoardID)
	if err != nil {
		return nil, err
	}

	// Get metrics history
	histories, err := s.metricsHistRepo.GetBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// If no history, record current snapshot
	if len(histories) == 0 {
		_, err := s.RecordDailySnapshot(ctx, sprintID)
		if err != nil {
			return nil, err
		}
		histories, err = s.metricsHistRepo.GetBySprintID(ctx, sprintID)
		if err != nil {
			return nil, err
		}
	}

	// Build dates list
	dates := make([]time.Time, len(histories))
	for i, h := range histories {
		dates[i] = h.RecordedDate
	}

	// Build column flow data
	columnFlows := make([]ColumnFlowData, 0, len(columns))
	for _, col := range columns {
		if col.IsHidden {
			continue
		}
		flow := ColumnFlowData{
			ColumnID:   col.ID,
			ColumnName: col.Name,
			Color:      col.Color,
			Values:     make([]int, len(histories)),
		}

		// Fill values from history snapshots
		for i, h := range histories {
			snapshot, _ := h.GetColumnSnapshot()
			if data, ok := snapshot[col.ID.String()]; ok {
				if mode == MetricModeStoryPoints {
					flow.Values[i] = data.StoryPoints
				} else {
					flow.Values[i] = data.CardCount
				}
			}
		}

		columnFlows = append(columnFlows, flow)
	}

	return &CumulativeFlowData{
		SprintID:   sprintID,
		SprintName: sp.Name,
		Columns:    columnFlows,
		Dates:      dates,
	}, nil
}

// GetSprintStats returns current statistics for a sprint
func (s *service) GetSprintStats(ctx context.Context, sprintID uuid.UUID) (*SprintStats, error) {
	ctx, span := s.startServiceSpan(ctx, "GetSprintStats")
	span.SetAttributes(attribute.String("sprint.id", sprintID.String()))
	defer span.End()

	// Get sprint
	sp, err := s.sprintRepo.GetByID(ctx, sprintID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSprintNotFound
		}
		return nil, err
	}

	// Get all cards in the sprint
	cards, err := s.cardRepo.GetBySprintID(ctx, sprintID)
	if err != nil {
		return nil, err
	}

	// Get all columns for the board to identify "done" columns
	columns, err := s.columnRepo.GetByBoardID(ctx, sp.BoardID)
	if err != nil {
		return nil, err
	}

	// Build a set of "done" column IDs
	doneColumnIDs := make(map[uuid.UUID]bool)
	for _, col := range columns {
		if col.IsDone {
			doneColumnIDs[col.ID] = true
		}
	}

	// Calculate stats
	stats := &SprintStats{}
	for _, c := range cards {
		stats.TotalCards++
		if c.StoryPoints != nil {
			stats.TotalStoryPoints += *c.StoryPoints
		}

		if doneColumnIDs[c.ColumnID] {
			stats.CompletedCards++
			if c.StoryPoints != nil {
				stats.CompletedStoryPoints += *c.StoryPoints
			}
		}
	}

	// Calculate days elapsed and remaining
	now := time.Now()
	if sp.StartDate != nil {
		stats.DaysElapsed = int(now.Sub(*sp.StartDate).Hours() / 24)
		if stats.DaysElapsed < 0 {
			stats.DaysElapsed = 0
		}
	}
	if sp.EndDate != nil {
		stats.DaysRemaining = int(sp.EndDate.Sub(now).Hours() / 24)
		if stats.DaysRemaining < 0 {
			stats.DaysRemaining = 0
		}
	}

	return stats, nil
}

// Helper function to generate date range
func generateDateRange(start, end time.Time) []time.Time {
	start = start.Truncate(24 * time.Hour)
	end = end.Truncate(24 * time.Hour)

	var dates []time.Time
	for d := start; !d.After(end); d = d.Add(24 * time.Hour) {
		dates = append(dates, d)
	}
	return dates
}
