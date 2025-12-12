-- Remove sprint permissions from roles
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN ('sprint:manage', 'sprint:view')
);

-- Remove sprint permissions
DELETE FROM permissions WHERE code IN ('sprint:manage', 'sprint:view');

-- Remove sprint_id from cards
DROP INDEX IF EXISTS idx_cards_sprint_id;
ALTER TABLE cards DROP COLUMN IF EXISTS sprint_id;

-- Drop sprints table and indexes
DROP INDEX IF EXISTS idx_sprints_position;
DROP INDEX IF EXISTS idx_sprints_status;
DROP INDEX IF EXISTS idx_sprints_project_id;
DROP TABLE IF EXISTS sprints;

-- Drop sprint status enum
DROP TYPE IF EXISTS sprint_status;
