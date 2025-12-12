-- Create metrics_history table for daily sprint snapshots
-- Used for burn down/up charts and cumulative flow diagrams
CREATE TABLE metrics_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    recorded_date DATE NOT NULL,

    -- Card count metrics
    total_cards INTEGER NOT NULL DEFAULT 0,
    completed_cards INTEGER NOT NULL DEFAULT 0,

    -- Story points metrics
    total_story_points INTEGER NOT NULL DEFAULT 0,
    completed_story_points INTEGER NOT NULL DEFAULT 0,

    -- Column flow snapshot (JSONB for cumulative flow diagram)
    -- Format: {"column_id": {"name": "Todo", "card_count": 5, "story_points": 13}, ...}
    column_snapshot JSONB NOT NULL DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Ensure one snapshot per sprint per day
    UNIQUE(sprint_id, recorded_date)
);

-- Index for efficient date range queries
CREATE INDEX idx_metrics_history_sprint_date ON metrics_history(sprint_id, recorded_date);
