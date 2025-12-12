-- Revert sprints from board-level back to project-level

-- Add project_id column back
ALTER TABLE sprints ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE CASCADE;

-- Migrate: get project_id from board
UPDATE sprints s
SET project_id = (
    SELECT b.project_id FROM boards b WHERE b.id = s.board_id
);

-- Make project_id NOT NULL
ALTER TABLE sprints ALTER COLUMN project_id SET NOT NULL;

-- Drop board_id column and its indexes
DROP INDEX IF EXISTS idx_sprints_board_id;
DROP INDEX IF EXISTS idx_sprints_board_position;
ALTER TABLE sprints DROP COLUMN board_id;

-- Recreate old indexes
CREATE INDEX idx_sprints_project_id ON sprints(project_id);
CREATE INDEX idx_sprints_position ON sprints(project_id, position);

-- Revert sprint permissions back to project-level
UPDATE permissions SET resource_type = 'project' WHERE code IN ('sprint:manage', 'sprint:view');
