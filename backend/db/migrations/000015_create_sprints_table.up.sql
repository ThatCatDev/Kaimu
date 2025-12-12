-- Create sprint status enum
CREATE TYPE sprint_status AS ENUM ('future', 'active', 'closed');

-- Create sprints table
CREATE TABLE sprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    goal TEXT,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    status sprint_status NOT NULL DEFAULT 'future',
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Index for fast project lookups
CREATE INDEX idx_sprints_project_id ON sprints(project_id);

-- Index for filtering by status
CREATE INDEX idx_sprints_status ON sprints(status);

-- Index for ordering sprints
CREATE INDEX idx_sprints_position ON sprints(project_id, position);

-- Add sprint_id to cards (nullable - cards in backlog have no sprint)
ALTER TABLE cards ADD COLUMN sprint_id UUID REFERENCES sprints(id) ON DELETE SET NULL;

-- Index for finding cards in a sprint
CREATE INDEX idx_cards_sprint_id ON cards(sprint_id);

-- Add sprint permissions to permissions table
INSERT INTO permissions (code, name, description, resource_type) VALUES
    ('sprint:manage', 'Manage Sprints', 'Create, update, delete, start, and complete sprints', 'project'),
    ('sprint:view', 'View Sprints', 'View sprints and sprint details', 'project')
ON CONFLICT (code) DO NOTHING;

-- Grant sprint permissions to existing roles
-- Owner role gets all sprint permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'Owner' AND p.code IN ('sprint:manage', 'sprint:view')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Admin role gets manage_sprints
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'Admin' AND p.code = 'sprint:manage'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Admin role gets view_sprints
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'Admin' AND p.code = 'sprint:view'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Member role gets view_sprints
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'Member' AND p.code = 'sprint:view'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Viewer role gets view_sprints
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'Viewer' AND p.code = 'sprint:view'
ON CONFLICT (role_id, permission_id) DO NOTHING;
