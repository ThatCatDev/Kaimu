-- Add story_points to cards (nullable integer for optional estimation)
ALTER TABLE cards ADD COLUMN story_points INTEGER;

-- Add is_done flag to board_columns to identify "done" columns for metrics
ALTER TABLE board_columns ADD COLUMN is_done BOOLEAN NOT NULL DEFAULT FALSE;

-- Index for efficient querying of done columns
CREATE INDEX idx_board_columns_is_done ON board_columns(board_id, is_done) WHERE is_done = TRUE;
