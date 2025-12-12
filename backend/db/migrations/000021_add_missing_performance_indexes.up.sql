-- Add additional performance indexes for common query patterns

-- Composite index for sprint queries filtered by board AND status
-- Improves: GetActiveByBoardID, GetFutureByBoardID, GetClosedByBoardID
-- The existing individual indexes (idx_sprints_board_id, idx_sprints_status) are less
-- efficient when both columns are used in WHERE clause
CREATE INDEX IF NOT EXISTS idx_sprints_board_status
    ON sprints(board_id, status);

-- Index for closed sprints with ordering by end_date (pagination)
-- Improves: GetClosedByBoardIDPaginated which orders by end_date DESC
CREATE INDEX IF NOT EXISTS idx_sprints_board_status_enddate
    ON sprints(board_id, status, end_date DESC)
    WHERE status = 'closed';

-- Composite index for cards by board with position ordering
-- Improves: GetByBoardID which uses WHERE board_id = ? ORDER BY position
-- Current idx_cards_board_id doesn't include position
CREATE INDEX IF NOT EXISTS idx_cards_board_position
    ON cards(board_id, position);

-- Index for card_sprints ordered by added_at
-- Improves: GetSprintIDsForCard which orders by added_at ASC
CREATE INDEX IF NOT EXISTS idx_card_sprints_card_added
    ON card_sprints(card_id, added_at ASC);

-- GIN index for JSONB metadata queries in audit_events (for sprint_id lookups)
-- Improves: GetSprintCardEvents which queries metadata->>'sprint_id'
-- Note: Only add if JSONB queries are frequent; may increase write overhead
CREATE INDEX IF NOT EXISTS idx_audit_events_metadata_gin
    ON audit_events USING GIN (metadata jsonb_path_ops);

-- Index for users by email_verified status (for email verification flows)
-- Improves: queries that filter by verified/unverified status
CREATE INDEX IF NOT EXISTS idx_users_email_verified
    ON users(email_verified)
    WHERE email_verified = FALSE;

-- Composite index for invitations lookup by org and expiry
-- Improves: queries for pending/expired invitations
CREATE INDEX IF NOT EXISTS idx_invitations_org_expires
    ON invitations(organization_id, expires_at DESC);
