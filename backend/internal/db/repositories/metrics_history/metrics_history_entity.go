package metrics_history

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ColumnSnapshotData represents the data for a single column in a snapshot
type ColumnSnapshotData struct {
	Name        string `json:"name"`
	CardCount   int    `json:"card_count"`
	StoryPoints int    `json:"story_points"`
}

// MetricsHistory stores daily snapshots of sprint metrics for burn charts
type MetricsHistory struct {
	ID                   uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SprintID             uuid.UUID       `gorm:"type:uuid;not null"`
	RecordedDate         time.Time       `gorm:"type:date;not null"`
	TotalCards           int             `gorm:"type:integer;not null;default:0"`
	CompletedCards       int             `gorm:"type:integer;not null;default:0"`
	TotalStoryPoints     int             `gorm:"type:integer;not null;default:0"`
	CompletedStoryPoints int             `gorm:"type:integer;not null;default:0"`
	ColumnSnapshot       json.RawMessage `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt            time.Time       `gorm:"autoCreateTime"`
}

func (MetricsHistory) TableName() string {
	return "metrics_history"
}

// GetColumnSnapshot parses the JSONB column snapshot into a map
func (m *MetricsHistory) GetColumnSnapshot() (map[string]ColumnSnapshotData, error) {
	var snapshot map[string]ColumnSnapshotData
	if err := json.Unmarshal(m.ColumnSnapshot, &snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

// SetColumnSnapshot serializes a map into JSONB for storage
func (m *MetricsHistory) SetColumnSnapshot(snapshot map[string]ColumnSnapshotData) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	m.ColumnSnapshot = data
	return nil
}
