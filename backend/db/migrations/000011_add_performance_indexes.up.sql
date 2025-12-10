-- Add performance indexes for common query patterns

-- Index for cards queries ordered by due_date when filtering by assignee
-- Used by: GetByAssigneeID which orders by "due_date ASC NULLS LAST, created_at DESC"
CREATE INDEX idx_cards_assignee_due_date ON cards (assignee_id, due_date ASC NULLS LAST, created_at DESC);

-- Index for board_columns filtered by is_hidden
-- Used by: GetVisibleByBoardID which filters WHERE is_hidden = FALSE
CREATE INDEX idx_board_columns_visible ON board_columns (board_id, is_hidden, position) WHERE is_hidden = FALSE;

-- Index for cards by card_id for the card_tags join table
-- Improves delete operations: WHERE card_id = ?
CREATE INDEX idx_card_tags_card_id ON card_tags (card_id);

-- Index for cards ordered by created_at (useful for recent cards queries)
CREATE INDEX idx_cards_created_at ON cards (created_at DESC);

-- Index for boards ordered by created_at (used in GetByProjectID)
CREATE INDEX idx_boards_created_at ON boards (project_id, created_at ASC);

-- Index for tags ordered by name (used in GetByProjectID which orders by name)
CREATE INDEX idx_tags_name ON tags (project_id, name ASC);
