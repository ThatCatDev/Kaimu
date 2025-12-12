-- Migrate sprints from project-level to board-level

-- Add board_id column (nullable initially for migration)
ALTER TABLE sprints ADD COLUMN board_id UUID REFERENCES boards(id) ON DELETE CASCADE;

-- Migrate existing sprints: assign to the first board of their project
-- This is a best-effort migration - existing sprints will be assigned to a board
UPDATE sprints s
SET board_id = (
    SELECT b.id FROM boards b
    WHERE b.project_id = s.project_id
    ORDER BY b.created_at ASC
    LIMIT 1
)
WHERE s.board_id IS NULL;

-- Delete any sprints that couldn't be assigned (project has no boards)
DELETE FROM sprints WHERE board_id IS NULL;

-- Make board_id NOT NULL now that migration is complete
ALTER TABLE sprints ALTER COLUMN board_id SET NOT NULL;

-- Drop the old project_id column and its index
DROP INDEX IF EXISTS idx_sprints_project_id;
DROP INDEX IF EXISTS idx_sprints_position;
ALTER TABLE sprints DROP COLUMN project_id;

-- Create new indexes for board-level sprints
CREATE INDEX idx_sprints_board_id ON sprints(board_id);
CREATE INDEX idx_sprints_board_position ON sprints(board_id, position);

-- Update sprint permissions to be board-level instead of project-level
UPDATE permissions SET resource_type = 'board' WHERE code IN ('sprint:manage', 'sprint:view');
